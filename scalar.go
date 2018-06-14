package graphqlmultipart

import (
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

var Upload = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "Upload",
	Description: fmt.Sprintf("The `Upload` scalar type represents a uploaded file using \"multipart/form-data\" as described in the spec: (%s)", specURL),
	Serialize: func(v interface{}) interface{} {
		panic("cannot serialize upload type")
	},
	ParseValue: func(v interface{}) interface{} {
		return v
	},
	ParseLiteral: func(valueAST ast.Value) interface{} {
		panic("cannot parse a upload literal")
	},
})
