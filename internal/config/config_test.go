package config

import (
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want 8080", cfg.Port)
	}
	if cfg.MaxSourceSize != 52428800 {
		t.Errorf("MaxSourceSize = %d, want 52428800", cfg.MaxSourceSize)
	}
	if cfg.MaxOutputWidth != 1400 {
		t.Errorf("MaxOutputWidth = %d, want 1400", cfg.MaxOutputWidth)
	}
	if cfg.MaxOutputHeight != 1400 {
		t.Errorf("MaxOutputHeight = %d, want 1400", cfg.MaxOutputHeight)
	}
	if cfg.CacheTTL != 5*time.Minute {
		t.Errorf("CacheTTL = %v, want 5m0s", cfg.CacheTTL)
	}
}

func TestLoad_ValidOverrides(t *testing.T) {
	tests := []struct {
		name   string
		envs   map[string]string
		check  func(t *testing.T, cfg *Config)
	}{
		{
			name: "PORT override",
			envs: map[string]string{"PORT": "9090"},
			check: func(t *testing.T, cfg *Config) {
				if cfg.Port != 9090 {
					t.Errorf("Port = %d, want 9090", cfg.Port)
				}
			},
		},
		{
			name: "MAX_SOURCE_SIZE override",
			envs: map[string]string{"MAX_SOURCE_SIZE": "1048576"},
			check: func(t *testing.T, cfg *Config) {
				if cfg.MaxSourceSize != 1048576 {
					t.Errorf("MaxSourceSize = %d, want 1048576", cfg.MaxSourceSize)
				}
			},
		},
		{
			name: "MAX_OUTPUT_WIDTH override",
			envs: map[string]string{"MAX_OUTPUT_WIDTH": "800"},
			check: func(t *testing.T, cfg *Config) {
				if cfg.MaxOutputWidth != 800 {
					t.Errorf("MaxOutputWidth = %d, want 800", cfg.MaxOutputWidth)
				}
			},
		},
		{
			name: "MAX_OUTPUT_HEIGHT override",
			envs: map[string]string{"MAX_OUTPUT_HEIGHT": "600"},
			check: func(t *testing.T, cfg *Config) {
				if cfg.MaxOutputHeight != 600 {
					t.Errorf("MaxOutputHeight = %d, want 600", cfg.MaxOutputHeight)
				}
			},
		},
		{
			name: "CACHE_TTL override",
			envs: map[string]string{"CACHE_TTL": "10m"},
			check: func(t *testing.T, cfg *Config) {
				if cfg.CacheTTL != 10*time.Minute {
					t.Errorf("CacheTTL = %v, want 10m0s", cfg.CacheTTL)
				}
			},
		},
		{
			name: "all overrides",
			envs: map[string]string{
				"PORT":              "3000",
				"MAX_SOURCE_SIZE":   "2097152",
				"MAX_OUTPUT_WIDTH":  "500",
				"MAX_OUTPUT_HEIGHT": "400",
				"CACHE_TTL":         "30s",
			},
			check: func(t *testing.T, cfg *Config) {
				if cfg.Port != 3000 {
					t.Errorf("Port = %d, want 3000", cfg.Port)
				}
				if cfg.MaxSourceSize != 2097152 {
					t.Errorf("MaxSourceSize = %d, want 2097152", cfg.MaxSourceSize)
				}
				if cfg.MaxOutputWidth != 500 {
					t.Errorf("MaxOutputWidth = %d, want 500", cfg.MaxOutputWidth)
				}
				if cfg.MaxOutputHeight != 400 {
					t.Errorf("MaxOutputHeight = %d, want 400", cfg.MaxOutputHeight)
				}
				if cfg.CacheTTL != 30*time.Second {
					t.Errorf("CacheTTL = %v, want 30s", cfg.CacheTTL)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envs {
				t.Setenv(k, v)
			}

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() unexpected error: %v", err)
			}

			tt.check(t, cfg)
		})
	}
}

func TestLoad_InvalidValues(t *testing.T) {
	tests := []struct {
		name    string
		envKey  string
		envVal  string
		wantErr string
	}{
		{
			name:    "invalid PORT",
			envKey:  "PORT",
			envVal:  "not-a-number",
			wantErr: "invalid PORT",
		},
		{
			name:    "invalid MAX_SOURCE_SIZE",
			envKey:  "MAX_SOURCE_SIZE",
			envVal:  "abc",
			wantErr: "invalid MAX_SOURCE_SIZE",
		},
		{
			name:    "invalid MAX_OUTPUT_WIDTH",
			envKey:  "MAX_OUTPUT_WIDTH",
			envVal:  "wide",
			wantErr: "invalid MAX_OUTPUT_WIDTH",
		},
		{
			name:    "invalid MAX_OUTPUT_HEIGHT",
			envKey:  "MAX_OUTPUT_HEIGHT",
			envVal:  "tall",
			wantErr: "invalid MAX_OUTPUT_HEIGHT",
		},
		{
			name:    "invalid CACHE_TTL",
			envKey:  "CACHE_TTL",
			envVal:  "not-a-duration",
			wantErr: "invalid CACHE_TTL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.envKey, tt.envVal)

			cfg, err := Load()
			if err == nil {
				t.Fatalf("Load() expected error containing %q, got nil (cfg=%+v)", tt.wantErr, cfg)
			}
			if got := err.Error(); !contains(got, tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", got, tt.wantErr)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
