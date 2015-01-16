package main

import (
	"github.com/abates/bms/server"
	"net/http"
)

func main() {
	http.ListenAndServe(":15115", server.Handlers())
}
