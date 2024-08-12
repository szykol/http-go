package log

import (
	"context"

	"go.uber.org/zap"
)

type key struct{}

func WithContext(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, key{}, logger)
}

func FromContext(ctx context.Context) *zap.SugaredLogger {
	ctxValue := ctx.Value(key{})
	if logger, ok := ctxValue.(*zap.SugaredLogger); ok {
		return logger
	}
	panic("Could not find logger in ctx")
}

func NewLogger(cfg zap.Config) *zap.SugaredLogger {
	logger := zap.Must(cfg.Build())
	sugar := logger.Sugar()
	return sugar
}

func NewZapCfg() zap.Config {
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "console"
	cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	return cfg
}
