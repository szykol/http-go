package http

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/szykol/http/pkg/log"
)

type serverTestF struct {
	server *Server
	ctx    context.Context
}

func noOpHandler(ResponseWriter, *Request) {}

func setupServerTest(t *testing.T) serverTestF {
	ctx, cancel := context.WithCancel(context.Background())
	logger := log.NewLogger(log.NewZapCfg())
	ctx = log.WithContext(ctx, logger)

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

func TestHandleRequest(t *testing.T) {
	f := setupServerTest(t)

	f.server.AddHandler("POST", "/test", func(w ResponseWriter, r *Request) {
		_, err := w.Write([]byte("unit test"))
		assert.Nil(t, err, "Handler should not return error")
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

func TestHandleRequestHandlerNotFound(t *testing.T) {
	f := setupServerTest(t)

	f.server.AddHandler("POST", "/test", func(w ResponseWriter, r *Request) {
		_, err := w.Write([]byte("unit test"))
		assert.Nil(t, err, "Handler should not return error")
	})

	request := &Request{
		path:   "/nonexistent",
		Method: "POST",
	}

	rd := &bytes.Buffer{}

	f.server.handleRequest(f.ctx, request, rd)

	assert.Contains(t, rd.String(), "HTTP/1.1 404 Not Found")
}

func TestHandleRequestInternalServerError(t *testing.T) {
	f := setupServerTest(t)

	f.server.AddHandler("POST", "/test", func(w ResponseWriter, r *Request) {
		panic("unit test")
	})

	request := &Request{
		path:   "/test",
		Method: "POST",
	}

	rd := &bytes.Buffer{}

	f.server.handleRequest(f.ctx, request, rd)

	assert.Contains(t, rd.String(), "HTTP/1.1 500 Internal Server Error")
}
