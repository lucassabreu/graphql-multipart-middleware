// Package graphqlmultipart provide a implementation of the graphql multipart request spec created by [@jaydenseric](https://github.com/jaydenseric) to provide support for handling file uploads in a GraphQL server, [click here to see the spec](https://github.com/jaydenseric/graphql-multipart-request-spec).
//
// Using the methods `graphqlmultipart.NewHandler` or `graphqlmultipart.NewMiddlewareWrapper` you will be abble to wrap your GraphQL handler and so every request made with the `Content-Type`: `multipart/form-data` will be handled by this package (using a provided GraphQL schema), and other `Content-Types` will be directed to your handler.
//
// The package also provide a scalar for the uploaded content called `graphqlmultipart.Upload`, when used it will populate your `InputObjects` or arguments with a `*multipart.FileHeader` for the uploaded file that can be used inside your queries/mutations.
package graphqlmultipart

import (
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
)

const specURL = "https://github.com/jaydenseric/graphql-multipart-request-spec/tree/v2.0.0"

var (
	// FailedToParseFormMessage is shown when it is a multipart/form-data, but its invalid
	FailedToParseFormMessage = "Failed to parse multipart form"

	// OperationsFieldMissingMessage is shown when the operations field is missing
	OperationsFieldMissingMessage = fmt.Sprintf("Field \"operations\" was not found in the form (%s)", specURL)

	// MapFieldMissingMessage is shown when the map field is missing
	MapFieldMissingMessage = fmt.Sprintf("Field \"map\" was not found in the form (%s)", specURL)

	// InvalidMapFieldMessage is shown when the map field format is invalid
	InvalidMapFieldMessage = fmt.Sprintf("Field \"map\" format is not valid (%s)", specURL)

	// InvalidOperationsFieldMessage is shown when operations field format is not valid
	InvalidOperationsFieldMessage = fmt.Sprintf("Field \"operations\" format is not valid (%s)", specURL)

	// MissingFileMessage is shown when a file is mapped, but not sent
	MissingFileMessage = fmt.Sprintf("File \"%%[1]s\" is missing, but exists in the map association (%s)", specURL)

	// InvalidMapPathMessage is shown when is not possible to find or populate the variable path
	InvalidMapPathMessage = fmt.Sprintf("Invalid mapping path \"%%[1]s\" for file %%[2]s (%s)", specURL)
)

// MultipartHandler implements the specification for handling multipart/form-data
// see more at: https://github.com/jaydenseric/graphql-multipart-request-spec/tree/v2.0.0
type MultipartHandler struct {
	Schema    *graphql.Schema
	next      http.Handler
	maxMemory int64
}

// NewHandler wraps the default GraphQL handler within a MultipartHandler, if it
// receives a request that is not "multipart/form-data", it will be forwarded to
// the wrapped handler
func NewHandler(s *graphql.Schema, maxMemory int64, next http.Handler) http.Handler {
	return MultipartHandler{
		Schema:    s,
		maxMemory: maxMemory,
		next:      next,
	}
}

// NewMiddlewareWrapper retrieves a func to help wrap multiple GraphQL handler with
// the MultipartHandler
func NewMiddlewareWrapper(s *graphql.Schema, maxMemory int64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return NewHandler(s, maxMemory, next)
	}
}

type operationField struct {
	Query         string                  `json:"query"`
	Variables     *map[string]interface{} `json:"variables"`
	OperationName string                  `json:"operationName"`
	mapPrefix     string
}

// ServeHTTP will process requests of the type "multipart/form-data", if other
// content-type was sent, it will be forwarded to the wrapped handler
func (m MultipartHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	ct = strings.Split(ct, ";")[0]
	if ct != "multipart/form-data" {
		m.next.ServeHTTP(w, r)
		return
	}

	if err := r.ParseMultipartForm(m.maxMemory); err != nil {
		log.Printf("[MultipartHandler] Fail do parse multipart form: %s", err.Error())
		writeError(w, FailedToParseFormMessage)
		return
	}

	form := r.MultipartForm

	var vs []string
	var ok bool

	if vs, ok = form.Value["operations"]; !ok {
		writeError(w, OperationsFieldMissingMessage)
		return
	}
	opsStr := vs[0]

	if vs, ok = form.Value["map"]; !ok {
		writeError(w, MapFieldMissingMessage)
		return
	}
	fileMapStr := vs[0]

	fileMap := make(map[string][]string)
	if err := json.Unmarshal([]byte(fileMapStr), &fileMap); err != nil {
		writeError(w, InvalidMapFieldMessage)
		return
	}

	batching := true
	ops := make([]operationField, 0)
	if err := json.Unmarshal([]byte(opsStr), &ops); err != nil {
		batching = false
		op := operationField{}
		err = json.Unmarshal([]byte(opsStr), &op)
		if err != nil || len(op.Query) == 0 || op.Variables == nil {
			writeError(w, InvalidOperationsFieldMessage)
			return
		}

		ops = append(ops, op)
	}

	if len(ops) == 0 {
		writeError(w, InvalidOperationsFieldMessage)
		return
	}

	results := make([]*graphql.Result, len(ops))

	for i, op := range ops {
		if batching {
			op.mapPrefix = fmt.Sprintf("%d.variables.", i)
		} else {
			op.mapPrefix = "variables."
		}
		results[i] = m.execute(op, fileMap, r)
	}

	w.WriteHeader(http.StatusOK)
	var buff []byte
	if batching {
		buff, _ = json.Marshal(results)
	} else {
		buff, _ = json.Marshal(results[0])
	}
	w.Write(buff)
}

func (m MultipartHandler) execute(op operationField, fMap map[string][]string, r *http.Request) *graphql.Result {

	errs := make([]error, 0)

	for f, ps := range fMap {

		if _, ok := r.MultipartForm.File[f]; !ok {
			errs = append(errs, fmt.Errorf(fmt.Sprintf(MissingFileMessage, f)))
			continue
		}

		for _, p := range ps {
			if !strings.HasPrefix(p, op.mapPrefix) {
				continue
			}

			vars, ok := injectFile(
				r.MultipartForm.File[f][0],
				op.Variables,
				p[len(op.mapPrefix):],
			)

			if !ok {
				errs = append(errs, fmt.Errorf(InvalidMapPathMessage, p, f))
				continue
			}
			*op.Variables = vars.(map[string]interface{})
		}
	}

	if len(errs) > 0 {
		return &graphql.Result{
			Errors: gqlerrors.FormatErrors(errs...),
		}
	}

	return graphql.Do(graphql.Params{
		Schema:         *m.Schema,
		RequestString:  op.Query,
		VariableValues: *op.Variables,
		OperationName:  op.OperationName,
		Context:        r.Context(),
	})
}

func injectFile(f *multipart.FileHeader, vars interface{}, path string) (interface{}, bool) {
	var field, next string

	field = path
	if i := strings.Index(".", path); i != -1 {
		field = path[0:i]
		next = path[i+1:]
	}

	switch v := vars.(type) {
	case map[string]interface{}:
		_, ok := v[field]
		if !ok {
			return v, false
		}

		if len(next) == 0 {
			v[field] = f
			return v, true
		}

		t, ok := injectFile(f, v[field], next)
		if !ok {
			return v, false
		}
		v[field] = t
		return v, true
	case []interface{}:
		index, err := strconv.Atoi(field)
		if err != nil {
			return v, false
		}

		if len(v) < (index + 1) {
			return v, false
		}

		if len(next) == 0 {
			v[index] = f
			return v, true
		}

		t, ok := injectFile(f, v[index], next)
		if !ok {
			return v, false
		}
		v[index] = t
		return v, true

	default:
		return v, false
	}
}

func writeError(w http.ResponseWriter, errs ...string) {
	fErrs := make([]gqlerrors.FormattedError, len(errs))
	for i, err := range errs {
		fErrs[i] = gqlerrors.NewFormattedError(err)
	}

	w.WriteHeader(http.StatusOK)
	buff, _ := json.Marshal(graphql.Result{Errors: fErrs})
	w.Write(buff)
}
