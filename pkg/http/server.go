package http

import (
	"context"
	"io"
	"net"

	"github.com/szykol/http/pkg/log"
)

type Server struct {
	handlers map[handlerIdentifier]RequestHandler
}

func NewServer() *Server {
	return &Server{
		handlers: make(map[handlerIdentifier]RequestHandler),
	}
}

func (s *Server) Run(ctx context.Context, l net.Listener) {
	logger := log.FromContext(ctx)

	logger.Debugw("Starting Server on", "listenAddr", l.Addr())

	newConnections := make(chan net.Conn)

	go listener(ctx, l, newConnections)

	for {
		select {
		case connection, ok := <-newConnections:
			if !ok {
				return
			}
			ctx := log.WithContext(ctx, logger.With("remote", connection.RemoteAddr().String()))
			go s.handleNewConnection(ctx, connection)
		case <-ctx.Done():
			logger.Debugw("Server.Run context done")
			return
		}
	}
}

func (s *Server) AddHandler(method, path string, requestHandler RequestHandler) {
	ident := handlerIdentifier{
		method: method,
		path:   path,
	}

	if _, ok := s.getHandler(ident); ok {
		panic("Handler for this method and path already registered")
	}

	s.handlers[ident] = requestHandler
}

func listener(ctx context.Context, listener net.Listener, connChan chan<- net.Conn) {
	logger := log.FromContext(ctx)

	for ctx.Err() == nil {
		conn, err := listener.Accept()
		if err != nil {
			logger.Errorw("Error accepting connection", "error", err)
			continue
		}

		connChan <- conn
	}

	close(connChan)
}

func (s *Server) handleNewConnection(ctx context.Context, rd io.ReadWriteCloser) {
	defer rd.Close()

	logger := log.FromContext(ctx)
	logger.Debug("Handling new connection")

	request, err := parseRequest(rd)
	if err != nil {
		logger.Errorw("Error parsing request", "request", request, "err", err)
		return
	}

	s.handleRequest(ctx, &request, rd)
}

func (s *Server) handleRequest(ctx context.Context, request *Request, rd io.ReadWriter) {
	ident := handlerIdentifier{
		path:   request.path,
		method: request.Method,
	}

	handler, ok := s.getHandler(ident)
	if !ok {
		handler = NotFoundHandler
	}

	requestWriter := newResponseWriter(rd)

	handle(ctx, handler, requestWriter, request)
}

func handle(ctx context.Context, h RequestHandler, w ResponseWriter, req *Request) {
	defer func() {
		if r := recover(); r != nil {
			logger := log.FromContext(ctx)
			logger.Errorw("Error when handling request")
			InternalServerErrorHandler(w, req)
		}
	}()

	h(w, req)
}

func (s *Server) getHandler(handlerIdentifier handlerIdentifier) (RequestHandler, bool) {
	// NOTE: assuming this will get more complicated over time

	handler, ok := s.handlers[handlerIdentifier]
	return handler, ok
}
