// Package handler provides HTTP handlers for the image-processing-api service.
package handler

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image/png"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/genekuo/image-processing-api/internal/cache"
	"github.com/genekuo/image-processing-api/internal/config"
	"github.com/genekuo/image-processing-api/internal/placeholder"
	"github.com/genekuo/image-processing-api/internal/service"
)

var (
	imageCacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "image_cache_hits_total",
		Help: "Total number of image cache hits.",
	})

	imageCacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "image_cache_misses_total",
		Help: "Total number of image cache misses.",
	})

	imageProcessingDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "image_processing_duration_seconds",
		Help:    "Duration of image processing pipeline in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation"})

	imageCacheEntries = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "image_cache_entries",
		Help: "Current number of entries in the image cache.",
	})
)

// ImageHandler serves processed images based on URL query parameters.
type ImageHandler struct {
	downloader *service.Downloader
	cache      *cache.Cache
	cfg        *config.Config
}

// NewImageHandler creates an ImageHandler with the provided dependencies.
func NewImageHandler(dl *service.Downloader, c *cache.Cache, cfg *config.Config) *ImageHandler {
	return &ImageHandler{
		downloader: dl,
		cache:      c,
		cfg:        cfg,
	}
}

// ServeHTTP handles image processing requests. It expects query parameters
// "url" (the source image URL) and "op" (comma-separated operations such as
// "rotate-90" or "resize-800x600"). On any error, a colour-coded placeholder
// image is returned with the appropriate HTTP status code.
func (h *ImageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Update cache entries gauge on every request.
	defer func() {
		imageCacheEntries.Set(float64(h.cache.Len()))
	}()

	rawURL := r.URL.Query().Get("url")
	opStr := r.URL.Query().Get("op")

	if rawURL == "" {
		h.writeError(w, http.StatusBadRequest, "missing required parameter: url", 0, 0)
		return
	}
	if opStr == "" {
		h.writeError(w, http.StatusBadRequest, "missing required parameter: op", 0, 0)
		return
	}

	ops, err := service.ParseOperations(opStr)
	if err != nil {
		ew, eh := extractDimensions(ops)
		h.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid operation: %v", err), ew, eh)
		return
	}

	// Generate cache key from URL + operations string.
	key := cacheKey(rawURL, opStr)

	// Check cache.
	if data, ok := h.cache.Get(key); ok {
		imageCacheHits.Inc()
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}
	imageCacheMisses.Inc()

	// Download and process.
	start := time.Now()

	img, err := h.downloader.Download(r.Context(), rawURL)
	if err != nil {
		slog.Error("download failed", "url", rawURL, "error", err)
		ew, eh := extractDimensions(ops)
		h.writeError(w, http.StatusBadGateway, fmt.Sprintf("download failed: %v", err), ew, eh)
		return
	}

	result, err := service.ApplyAll(img, ops)
	if err != nil {
		slog.Error("processing failed", "error", err)
		ew, eh := extractDimensions(ops)
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("processing failed: %v", err), ew, eh)
		return
	}

	duration := time.Since(start).Seconds()
	imageProcessingDuration.WithLabelValues(opStr).Observe(duration)

	// Encode result to PNG.
	var buf bytes.Buffer
	if err := png.Encode(&buf, result); err != nil {
		slog.Error("encode failed", "error", err)
		ew, eh := extractDimensions(ops)
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("encode failed: %v", err), ew, eh)
		return
	}

	data := buf.Bytes()
	h.cache.Set(key, data)

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// writeError generates and writes a placeholder image response with the given
// HTTP status code. Width and height of zero use the placeholder defaults.
func (h *ImageHandler) writeError(w http.ResponseWriter, statusCode int, msg string, width, height int) {
	slog.Warn("returning error placeholder", "status", statusCode, "message", msg)
	data, err := placeholder.Generate(statusCode, width, height)
	if err != nil {
		slog.Error("failed to generate placeholder", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(statusCode)
	w.Write(data)
}

// cacheKey computes a SHA-256 hex digest of the URL and operations string.
func cacheKey(rawURL, ops string) string {
	h := sha256.Sum256([]byte(rawURL + "|" + ops))
	return fmt.Sprintf("%x", h)
}

// extractDimensions returns the width and height from the last resize operation
// in the list, or (0, 0) if no resize is present.
func extractDimensions(ops []service.Operation) (int, int) {
	for i := len(ops) - 1; i >= 0; i-- {
		if ops[i].Type == "resize" {
			return ops[i].Width, ops[i].Height
		}
	}
	return 0, 0
}
