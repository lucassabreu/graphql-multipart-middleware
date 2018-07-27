package testutil

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

// NewGraphQLFileUploadRequest creates a simple send operations and map through the `fields` param
func NewGraphQLFileUploadRequest(url string, fields map[string]string, files map[string]string) *http.Request {
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

	for key, val := range fields {
		_ = writer.WriteField(key, val)
	}

	if err := writer.Close(); err != nil {
		panic(err)
	}

	r, _ := http.NewRequest("POST", url, body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	return r
}
