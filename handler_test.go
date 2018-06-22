package graphqlmultipart_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	graphqlmultipart "github.com/lucassabreu/graphql-multipart-middleware"
	"github.com/lucassabreu/graphql-multipart-middleware/testutil"

	"github.com/stretchr/testify/require"
)

func TestMultipartMiddleware_ForwardsRequestsOtherThanMultipart(t *testing.T) {
	newRequest := func(contentType string) *http.Request {
		r, _ := http.NewRequest("GET", "/graphql", strings.NewReader(""))
		r.Header.Set("Content-type", contentType)
		return r
	}

	cases := map[string]*http.Request{
		"application/json":                  newRequest("application/json"),
		"application/graphql":               newRequest("application/graphql"),
		"application/x-www-form-urlencoded": newRequest("application/x-www-form-urlencoded"),
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("reached me"))
	})

	for n, r := range cases {
		t.Run(n, func(t *testing.T) {
			resp := httptest.NewRecorder()

			mh := graphqlmultipart.NewHandler(
				&testutil.Schema,
				1*1024,
				h,
			)

			mh.ServeHTTP(resp, r)

			body, _ := ioutil.ReadAll(resp.Result().Body)

			require.Equal(t, string(body), "reached me")
		})
	}
}

func newFileUploadRequest(params map[string]string, files map[string]string) *http.Request {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for paramName, path := range files {
		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		fileContents, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		fi, err := file.Stat()
		if err != nil {
			panic(err)
		}
		file.Close()

		part, err := writer.CreateFormFile(paramName, fi.Name())
		if err != nil {
			panic(err)
		}
		part.Write(fileContents)
	}

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	if err := writer.Close(); err != nil {
		panic(err)
	}

	r, _ := http.NewRequest("POST", "/graphql", body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	return r
}

func getJSONError(m string, v ...interface{}) string {
	if len(v) > 0 {
		m = fmt.Sprintf(m, v...)
	}

	return fmt.Sprintf(
		"{\"data\":null,\"errors\": [{\"message\":%s, \"locations\":[]}]}",
		strconv.Quote(m),
	)
}

func TestHandlerShouldValidateRequest(t *testing.T) {
	type test struct {
		req   *http.Request
		respo string
	}
	cases := map[string]test{
		"missing_operation_field": test{
			req:   newFileUploadRequest(make(map[string]string), make(map[string]string)),
			respo: getJSONError(graphqlmultipart.OperationsFieldMissingMessage),
		},
		"missing_map_field": test{
			req: newFileUploadRequest(
				map[string]string{"operations": "{}"},
				make(map[string]string),
			),
			respo: getJSONError(graphqlmultipart.MapFieldMissingMessage),
		},
		"invalid_operaction_field": test{
			req: newFileUploadRequest(
				map[string]string{
					"operations": "{}",
					"map":        "{}",
				},
				make(map[string]string),
			),
			respo: getJSONError(graphqlmultipart.InvalidOperationsFieldMessage),
		},
		"invalid_operaction_field_empty_array": test{
			req: newFileUploadRequest(
				map[string]string{
					"operations": "[]",
					"map":        "{}",
				},
				make(map[string]string),
			),
			respo: getJSONError(graphqlmultipart.InvalidOperationsFieldMessage),
		},
		"invalid_map_field": test{
			req: newFileUploadRequest(
				map[string]string{
					"operations": `{"query":"hero{id}","variables":{}}`,
					"map":        "[]",
				},
				make(map[string]string),
			),
			respo: getJSONError(graphqlmultipart.InvalidMapFieldMessage),
		},
		"missing_file": test{
			req: newFileUploadRequest(
				map[string]string{
					"operations": "[{\"query\":\"query hero{id}\",\"variables\":{}}]",
					"map":        "{\"file\":[]}",
				},
				make(map[string]string),
			),
			respo: "[" + getJSONError(graphqlmultipart.MissingFileMessage, "file") + "]",
		},
		"invalid_map_path": test{
			req: newFileUploadRequest(
				map[string]string{
					"operations": "{\"query\":\"query hero{id}\",\"variables\":{}}",
					"map":        "{\"file\":[\"variables.file\"]}",
				},
				map[string]string{"file": "handler.go"},
			),
			respo: getJSONError(graphqlmultipart.InvalidMapPathMessage, "variables.file", "file"),
		},
		"invalid_map_path_batching": test{
			req: newFileUploadRequest(
				map[string]string{
					"operations": "[{\"query\":\"query hero{id}\",\"variables\":{}}]",
					"map":        "{\"file\":[\"0.variables.file\"]}",
				},
				map[string]string{"file": "handler.go"},
			),
			respo: "[" + getJSONError(graphqlmultipart.InvalidMapPathMessage, "0.variables.file", "file") + "]",
		},
	}

	mh := graphqlmultipart.NewHandler(
		&testutil.Schema,
		1*1024,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Should not have forwarded the request"))
		}),
	)

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {

			resp := httptest.NewRecorder()

			mh.ServeHTTP(resp, test.req)

			body, _ := ioutil.ReadAll(resp.Result().Body)

			require.JSONEq(t, string(body), test.respo)
		})
	}
}

func TestHandler_SchemaCanAccessUploads(t *testing.T) {
	type test struct {
		req   *http.Request
		respo string
	}

	cases := map[string]test{
		"simple": test{
			req: newFileUploadRequest(
				map[string]string{
					"operations": `{
						"query":"query($file:Upload) { upload(file: $file){ filename } }",
						"variables":{"file":null}
					}`,
					"map": `{"file":["variables.file"]}`,
				},
				map[string]string{"file": "handler.go"},
			),
			respo: `{"data":{"upload":{"filename":"handler.go"}}}`,
		},
		"batching": test{
			req: newFileUploadRequest(
				map[string]string{
					"operations": `[
						{
							"query":"query ($file: Upload){ upload(file:$file){filename} }",
							"variables":{"file":null}
						},
						{
							"query":"query ($other: Upload, $file: Upload) { other: upload(file:$other) {filename}, file:upload(file:$file) {filename} }",
							"variables":{"other":null,"file":null}
						}
					]`,
					"map": `{"file":["0.variables.file","1.variables.file"],"other_file":["1.variables.other"]}`,
				},
				map[string]string{"file": "handler.go", "other_file": "handler.go"},
			),
			respo: `[
				{"data":{"upload":{"filename":"handler.go"}}},
				{
					"data":{
						"other":{"filename":"handler.go"},
						"file":{"filename":"handler.go"}
					}
				}
			]`,
		},
		"simple_deeper": test{
			req: newFileUploadRequest(
				map[string]string{
					"operations": `{
						"query":"query($files: [Upload]) { uploads(files: $files){ filename } }",
						"variables":{"files":[null,null]}
					}`,
					"map": `{"file":["variables.files.0","variables.files.1"]}`,
				},
				map[string]string{"file": "handler.go"},
			),
			respo: `{
				"data":{
					"uploads":[
						{"filename":"handler.go"},
						{"filename":"handler.go"}
					]
				}
			}`,
		},
		"complex_deeper": test{
			req: newFileUploadRequest(
				map[string]string{
					"operations": `{
						"query":"query($input: SpecialUploadInput) { u:specialUploads(input: $input){ filename } }",
						"variables":{
							"input":{
								"name":"test",
								"files":[null,null]
							}
						}
					}`,
					"map": `{"file":["variables.input.files.0","variables.input.files.1"]}`,
				},
				map[string]string{"file": "handler.go"},
			),
			respo: `{
				"data":{
					"u":[
						{"filename":"handler.go"},
						{"filename":"handler.go"}
					]
				}
			}`,
		},
	}

	mh := graphqlmultipart.NewHandler(
		&testutil.Schema,
		1*1024,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("should not have forwarded the request"))
		}),
	)

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			resp := httptest.NewRecorder()
			mh.ServeHTTP(resp, test.req)
			body, _ := ioutil.ReadAll(resp.Result().Body)
			require.JSONEq(t, string(body), test.respo)
		})
	}
}
