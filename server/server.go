package server

import (
	"fmt"
	"net/http"
)

func handle(response http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(response, "OK")
}

func Handlers() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/foo", handle)

	return mux
}
