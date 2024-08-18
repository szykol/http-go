package http

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	mock_net "github.com/szykol/http/mocks"
	"github.com/szykol/http/pkg/log"
	"go.uber.org/mock/gomock"
)

type serverTestF struct {
	server *Server
	ctx    context.Context
	ctrl   *gomock.Controller
}

func noOpHandler(ResponseWriter, *Request) {}

func setupServerTest(t *testing.T) serverTestF {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	logger := log.NewLogger(log.NewZapCfg())
	ctx = log.WithContext(ctx, logger)
	ctrl := gomock.NewController(t)

	t.Cleanup(func() {
		cancel()
		ctrl.Finish()
	})

	return serverTestF{
		server: NewServer(),
		ctx:    ctx,
		ctrl:   ctrl,
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

// Wait for conn on channel or handle test timeout, whichever happens first
func getNextConn(t *testing.T, f serverTestF, c chan net.Conn) (net.Conn, bool) {
	select {
	case conn, ok := <-c:
		return conn, ok
	case <-f.ctx.Done():
		t.Fatal("Test timeout before channel receive")
	}

	return nil, false
}

func TestListener(t *testing.T) {
	f := setupServerTest(t)

	ctx, cancel := context.WithCancel(f.ctx)
	defer cancel()

	listenerMock := mock_net.NewMockListener(f.ctrl)
	connMock := mock_net.NewMockConn(f.ctrl)

	connChan := make(chan net.Conn)

	listenerMock.EXPECT().Accept().Return(connMock, nil)
	listenerMock.EXPECT().Accept().DoAndReturn(func() (net.Conn, error) {
		cancel()
		return nil, fmt.Errorf("unit test")
	})

	go listener(ctx, listenerMock, connChan)

	conn, _ := getNextConn(t, f, connChan)
	assert.Equal(t, connMock, conn)

	_, ok := getNextConn(t, f, connChan)
	assert.False(t, ok)
}
