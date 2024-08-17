package main

import (
	"context"

	"github.com/szykol/http/pkg/http"
	"github.com/szykol/http/pkg/http/log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := log.NewLogger(log.NewZapCfg())

	ctx = log.WithContext(ctx, logger)

	server := http.NewServer()

	server.AddHandler("POST", "/echo", func(w http.ResponseWriter, r *http.Request) {
		logger.Debugw("Received body", "payload", r.Payload)
		if _, err := w.Write(r.Payload); err != nil {
			logger.Errorw("error handling request", "error", err)
			return
		}
		logger.Debugw("Successfully written data")
	})

	server.AddHandler("GET", "/test", func(w http.ResponseWriter, r *http.Request) {
		w.SetStatus(200)
	})

	server.Run(ctx, "0.0.0.0:1337")
}
