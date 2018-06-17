package testutil

import (
	"mime/multipart"

	"github.com/graphql-go/graphql"
	graphqlmultipart "github.com/lucassabreu/graphql-multipart-middleware"
)

var (
	nonNullStringField = &graphql.Field{
		Type: graphql.NewNonNull(graphql.String),
	}

	headerType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Header",
		Fields: graphql.Fields{
			"name": nonNullStringField,
			"values": &graphql.Field{
				Type: graphql.NewList(graphql.NewNonNull(graphql.String)),
			},
		},
	})

	uploadedType = graphql.NewObject(graphql.ObjectConfig{
		Name: "UploadedFile",
		Fields: graphql.Fields{
			"filename": nonNullStringField,
			"headers": &graphql.Field{
				Type: graphql.NewList(headerType),
			},
			"size": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
	})

	// Schema is a simple schema that receives files and return its metadata
	Schema graphql.Schema
)

type uploadedHeader struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type uploadedFile struct {
	Filename string           `json:"filename"`
	Headers  []uploadedHeader `json:"headers"`
	Size     int64            `json:"size"`
}

func init() {
	var err error
	Schema, err = graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "RootQuery",
			Fields: graphql.Fields{
				"upload": &graphql.Field{
					Name:        "UploadQuery",
					Description: "Receives a Uploaded file and returns its metadata",
					Args: graphql.FieldConfigArgument{
						"file": &graphql.ArgumentConfig{
							Type: graphqlmultipart.Upload,
						},
					},
					Type: uploadedType,
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						file := p.Args["file"].(*multipart.FileHeader)
						r := uploadedFile{
							Filename: file.Filename,
							Size:     file.Size,
							Headers:  make([]uploadedHeader, len(file.Header)),
						}

						i := 0
						for n, vs := range file.Header {
							r.Headers[i] = uploadedHeader{
								Name:   n,
								Values: vs,
							}
							i++
						}
						return r, nil
					},
				},
			},
		}),
	})
	if err != nil {
		panic(err)
	}
}
