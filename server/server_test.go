package server_test

import (
	"github.com/abates/bms/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http/httptest"
)

var _ = Describe("Server", func() {
	BeforeEach(func() {
		testServer = httptest.NewServer(server.Handlers())
	})

	It("Should handle GET requests to /foo", func() {
		Expect(get("/foo").StatusCode).To(Equal(200))
	})
})
