package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/graphql-go/handler"
	graphqlmultipart "github.com/lucassabreu/graphql-multipart-middleware"
	"github.com/lucassabreu/graphql-multipart-middleware/testutil"
)

func main() {
	h := http.NewServeMux()
	s := &testutil.Schema
	h.Handle("/graphql", graphqlmultipart.NewHandler(
		s,
		4*1024*1024,
		handler.New(&handler.Config{Schema: s, GraphiQL: true}),
	))

	port := os.Getenv("PORT")

	fmt.Println("Now server is running on port " + port)
	fmt.Println(`Test with: curl -g 'http://localhost:8000/graphql'
	-F 'operations={"query":"query($file:Upload) { upload(file: $file){ filename, size, headers {name, values} } }","variables":{"file":null}}'
	-F 'map={"file":["variables.file"]}'
	-F file=@main.go`)

	http.ListenAndServe(":"+port, h)
}
