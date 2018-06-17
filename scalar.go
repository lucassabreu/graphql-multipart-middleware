package graphqlmultipart

import (
	"fmt"
	"mime/multipart"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

// Upload is a scalar represents a uploaded file using \"multipart/form-data\" as described in the graphql multipart spec
var Upload = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "Upload",
	Description: fmt.Sprintf("The `Upload` scalar represents a uploaded file using \"multipart/form-data\" as described in the spec: (%s)", specURL),
	Serialize: func(v interface{}) interface{} {
		panic("cannot serialize upload type")
	},
	ParseValue: func(v interface{}) interface{} {
		switch v.(type) {
		case *multipart.FileHeader:
			return v
		case multipart.FileHeader:
			return v
		default:
			return nil
		}
	},
	ParseLiteral: func(valueAST ast.Value) interface{} {
		panic("cannot parse a upload literal")
	},
})
