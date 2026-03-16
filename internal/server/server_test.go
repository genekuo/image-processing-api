package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/genekuo/image-processing-api/internal/config"
)

// newTestServer creates a fully-wired Server for testing purposes and returns
// both the Server and the underlying http.Handler (with middleware applied).
func newTestServer(t *testing.T) http.Handler {
	t.Helper()

	cfg := &config.Config{
		Port:            0,
		MaxSourceSize:   50 * 1024 * 1024,
		MaxOutputWidth:  1400,
		MaxOutputHeight: 1400,
		CacheTTL:        5 * time.Minute,
	}

	srv := New(cfg)
	t.Cleanup(func() {
		srv.cache.Stop()
	})

	return srv.httpServer.Handler
}

func TestHealthEndpoint(t *testing.T) {
	h := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", body["status"])
	}
}

func TestReadyEndpoint(t *testing.T) {
	h := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body["status"] != "ready" {
		t.Fatalf("expected status ready, got %q", body["status"])
	}
}

func TestMetricsEndpoint(t *testing.T) {
	h := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "http_requests_total") && !strings.Contains(body, "go_") {
		t.Fatal("expected prometheus metrics in response body")
	}
}

func TestCORSHeaders(t *testing.T) {
	h := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	origin := rec.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin *, got %q", origin)
	}

	methods := rec.Header().Get("Access-Control-Allow-Methods")
	if methods != "GET, OPTIONS" {
		t.Fatalf("expected Access-Control-Allow-Methods GET, OPTIONS, got %q", methods)
	}
}

func TestCORSPreflight(t *testing.T) {
	h := newTestServer(t)

	req := httptest.NewRequest(http.MethodOptions, "/image", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204 for OPTIONS preflight, got %d", rec.Code)
	}

	origin := rec.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin *, got %q", origin)
	}
}
