package logger

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"sync"
)

type ctxKey struct{}

var once sync.Once
var logger *zap.Logger

// Get initializes a zap.Logger instance if it has not been initialized
// already and returns the same instance for subsequent calls.
func Get(logLevelValue string) *zap.Logger {
	once.Do(func() {
		stdout := zapcore.AddSync(os.Stdout)

		level := zap.InfoLevel
		if logLevelValue != "" {
			levelFromEnv, err := zapcore.ParseLevel(logLevelValue)
			if err != nil {
				log.Println(
					fmt.Errorf("invalid level, defaulting to INFO: %w", err),
				)
			}

			level = levelFromEnv
		}

		logLevel := zap.NewAtomicLevelAt(level)

		developmentCfg := zap.NewDevelopmentEncoderConfig()
		developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

		consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)

		core := zapcore.NewTee(zapcore.NewCore(consoleEncoder, stdout, logLevel))

		logger = zap.New(core)
	})

	return logger
}

// Logger returns the Logger associated with the context. If no logger
// is associated, the default logger is returned, unless it is nil
// in which case a disabled logger is returned.
func Logger(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
		return l
	} else if l := logger; l != nil {
		return l
	}

	return zap.NewNop()
}

// Attach returns a copy of the context with the Logger attached.
func Attach(ctx context.Context, l *zap.Logger) context.Context {
	if lp, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
		if lp == l {
			// Do not store same logger.
			return ctx
		}
	}

	return context.WithValue(ctx, ctxKey{}, l)
}
