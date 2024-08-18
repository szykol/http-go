package main

import (
	"context"
	"net"

	"github.com/szykol/http/pkg/http"
	"github.com/szykol/http/pkg/log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := log.NewLogger(log.NewZapCfg())
	ctx = log.WithContext(ctx, logger)

	listener, err := net.Listen("tcp", "0.0.0.0:1337")
	if err != nil {
		panic(err)
	}

	server := http.NewServer()
	server.AddHandler("POST", "/echo", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(r.Payload); err != nil {
			logger.Errorw("error handling request", "error", err)
			return
		}
		logger.Debugw("Successfully written data")
	})
	server.AddHandler("GET", "/test", func(w http.ResponseWriter, r *http.Request) {
		_ = w.SetStatus(200)
	})

	server.Run(ctx, listener)
}
