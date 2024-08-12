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
	server.Run(ctx, "0.0.0.0:1337")
}
