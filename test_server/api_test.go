package test_server

import (
	"net/http"
	"testing"
)

func TestAPi(t *testing.T) {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("hello world"))
	})

	if err := http.ListenAndServe(":8182", nil); err != nil {
		panic(err)
	}
}

// http://127.0.0.1:8182/
