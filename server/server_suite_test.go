package server_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	testServer *httptest.Server
)

var request = func(verb string, path string) *http.Response {
	url := fmt.Sprintf("%s%s", testServer.URL, path)
	req, err := http.NewRequest(verb, url, nil)
	if err != nil {
		Fail(fmt.Sprintf("Could not create %s request to %s: %v", verb, path, err))
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		Fail(fmt.Sprintf("Could not execute request to %s: %v", path, err))
	}
	return res
}

var get = func(path string) *http.Response {
	return request("GET", path)
}

func TestServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Server Suite")
}
