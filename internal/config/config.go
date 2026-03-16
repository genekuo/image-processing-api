// Package config provides application configuration loaded from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration values.
type Config struct {
	Port            int
	MaxSourceSize   int64
	MaxOutputWidth  int
	MaxOutputHeight int
	CacheTTL        time.Duration
}

// Load reads configuration from environment variables, falling back to sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		Port:            8080,
		MaxSourceSize:   52428800, // 50 MB
		MaxOutputWidth:  1400,
		MaxOutputHeight: 1400,
		CacheTTL:        5 * time.Minute,
	}

	if v := os.Getenv("PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT %q: %w", v, err)
		}
		cfg.Port = p
	}

	if v := os.Getenv("MAX_SOURCE_SIZE"); v != "" {
		s, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid MAX_SOURCE_SIZE %q: %w", v, err)
		}
		cfg.MaxSourceSize = s
	}

	if v := os.Getenv("MAX_OUTPUT_WIDTH"); v != "" {
		w, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid MAX_OUTPUT_WIDTH %q: %w", v, err)
		}
		cfg.MaxOutputWidth = w
	}

	if v := os.Getenv("MAX_OUTPUT_HEIGHT"); v != "" {
		h, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid MAX_OUTPUT_HEIGHT %q: %w", v, err)
		}
		cfg.MaxOutputHeight = h
	}

	if v := os.Getenv("CACHE_TTL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid CACHE_TTL %q: %w", v, err)
		}
		cfg.CacheTTL = d
	}

	return cfg, nil
}
