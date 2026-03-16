// Package main is the entry point for the image-processing-api service.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/genekuo/image-processing-api/internal/config"
	"github.com/genekuo/image-processing-api/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	srv := server.New(cfg)

	// Start server in background goroutine.
	go func() {
		if err := srv.Start(); err != nil {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("server started", "port", cfg.Port)

	// Wait for interrupt signal.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Graceful shutdown with 10s timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "error", err)
		os.Exit(1)
	}
}
