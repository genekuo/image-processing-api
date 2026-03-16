// Package main is the entry point for the image-processing-api service.
package main

import (
	"log/slog"
	"os"

	"github.com/genekuo/image-processing-api/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	slog.Info("server starting", "port", cfg.Port)
}
