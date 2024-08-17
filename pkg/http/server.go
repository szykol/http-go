package http

import (
	"context"
	"net"

	"github.com/szykol/http/pkg/http/log"
)

type Server struct {
	handlers map[handlerIdentifier]RequestHandler
}

func NewServer() *Server {
	return &Server{
		handlers: make(map[handlerIdentifier]RequestHandler),
	}
}

func (s *Server) Run(ctx context.Context, listenAddr string) {
	logger := log.FromContext(ctx)

	logger.Debugw("Starting Server on", "listenAddr", listenAddr)

	newConnections := make(chan net.Conn)

	go listener(ctx, listenAddr, newConnections)

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

func listener(ctx context.Context, listenAddr string, connChan chan<- net.Conn) {
	logger := log.FromContext(ctx)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		logger.Errorw("Error creating listener", "error", err)
		return
	}

	for ctx.Err() == nil {
		conn, err := listener.Accept()
		if err != nil {
			logger.Errorw("Error accepting connection", "error", err)
			continue
		}

		connChan <- conn
	}
}

func (s *Server) handleNewConnection(ctx context.Context, conn net.Conn) {
	logger := log.FromContext(ctx)
	logger.Debug("Handling new connection")

	defer conn.Close()

	request, err := parseRequest(conn)
	if err != nil {
		logger.Errorw("Error parsing request", "request", request, "err", err)
		return
	}

	ident := handlerIdentifier{
		path:   request.path,
		method: request.Method,
	}

	handler, ok := s.getHandler(ident)
	if !ok {
		logger.Errorw("Could not get handler for request", "request", request.path, "err", err)
		return
	}

	requestWriter := newResponseWriter(conn)

	handler(requestWriter, &request)
}

func (s *Server) getHandler(handlerIdentifier handlerIdentifier) (RequestHandler, bool) {
	// NOTE: assuming this will get more complicated over time

	handler, ok := s.handlers[handlerIdentifier]
	return handler, ok
}
