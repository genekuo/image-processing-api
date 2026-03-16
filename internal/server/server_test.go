package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
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

func TestStartAndShutdown(t *testing.T) {
	cfg := &config.Config{
		Port:            0, // OS assigns a free port
		MaxSourceSize:   50 * 1024 * 1024,
		MaxOutputWidth:  1400,
		MaxOutputHeight: 1400,
		CacheTTL:        5 * time.Minute,
	}

	srv := New(cfg)

	// Pick a free port by opening a listener, reading its port, and closing it.
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("finding free port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()

	srv.httpServer.Addr = fmt.Sprintf(":%d", port)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	// Give the server a moment to start listening.
	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}

	// Start should return nil because ListenAndServe returns ErrServerClosed.
	if err := <-errCh; err != nil {
		t.Fatalf("expected nil from Start after graceful shutdown, got: %v", err)
	}
}

func TestStartInvalidPort(t *testing.T) {
	cfg := &config.Config{
		Port:            0,
		MaxSourceSize:   50 * 1024 * 1024,
		MaxOutputWidth:  1400,
		MaxOutputHeight: 1400,
		CacheTTL:        5 * time.Minute,
	}

	srv := New(cfg)
	// Set an invalid address so ListenAndServe fails immediately.
	srv.httpServer.Addr = ":-1"

	err := srv.Start()
	if err == nil {
		t.Fatal("expected error from Start with invalid port, got nil")
	}

	t.Cleanup(func() {
		srv.cache.Stop()
	})
}
