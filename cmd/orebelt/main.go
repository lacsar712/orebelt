package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lacsar712/orebelt/internal/app"
	"github.com/lacsar712/orebelt/internal/config"
	"github.com/lacsar712/orebelt/internal/logx"
)

func main() {
	cfg, err := config.ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "orebelt: %v\n", err)
		os.Exit(2)
	}

	logger := logx.New().WithLevel(parseLogLevel(cfg.LogLevel))
	application, err := app.New(cfg, logger)
	if err != nil {
		logger.Error("startup failed: %v", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := application.Run(ctx); err != nil {
		logger.Error("runtime error: %v", err)
		os.Exit(1)
	}
}

func parseLogLevel(raw string) logx.Level {
	switch raw {
	case "debug":
		return logx.LevelDebug
	case "warn":
		return logx.LevelWarn
	case "error":
		return logx.LevelError
	default:
		return logx.LevelInfo
	}
}
