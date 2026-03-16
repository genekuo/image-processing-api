package server

import (
	"encoding/json"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RegisterRoutes binds the application endpoints to the provided ServeMux.
// It registers the image handler, health/readiness probes, and the Prometheus
// metrics endpoint.
func RegisterRoutes(mux *http.ServeMux, imgHandler http.Handler) {
	mux.Handle("GET /image", imgHandler)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	})

	mux.Handle("GET /metrics", promhttp.Handler())
}
