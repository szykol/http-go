package http

import (
	"bufio"
	"context"
	"net"

	"github.com/szykol/http/pkg/http/log"
)

type Server struct{}

func NewServer() *Server {
	return &Server{}
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

	reader := bufio.NewReader(conn)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		logger.Errorw("Error reading from remote", "error", err)
	}

	_, err = conn.Write(line)
	if err != nil {
		logger.Errorw("Error writing to remote", "error", err)
	}
}
