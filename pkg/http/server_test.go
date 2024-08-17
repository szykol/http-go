package http

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type serverTestF struct {
	server *Server
	ctx    context.Context
}

func noOpHandler(ResponseWriter, *Request) {}

func setupServerTest(t *testing.T) serverTestF {
	ctx, cancel := context.WithCancel(context.Background())

	t.Cleanup(func() {
		cancel()
	})

	return serverTestF{
		server: NewServer(),
		ctx:    ctx,
	}
}

func TestAddHandler(t *testing.T) {
	f := setupServerTest(t)

	f.server.AddHandler("POST", "/test", noOpHandler)
	f.server.AddHandler("GET", "/test", noOpHandler)

	assert.Panics(t, func() {
		f.server.AddHandler("POST", "/test", noOpHandler)
	})
}

func TestHandleNewConnection(t *testing.T) {
	f := setupServerTest(t)

	f.server.AddHandler("POST", "/test", func(w ResponseWriter, r *Request) {
		w.Write([]byte("unit test"))
	})

	request := &Request{
		path:   "/test",
		Method: "POST",
	}

	rd := &bytes.Buffer{}

	f.server.handleRequest(f.ctx, request, rd)

	expectedResponse := []byte("HTTP/1.1 200 OK\r\nContent-Length: 9\r\nContent-Type: application/x-www-form-urlencoded\r\nConnection: Keep-Alive\r\nServer: go-simple-server\r\n\r\nunit test")

	assert.Equal(t, expectedResponse, rd.Bytes())
}
