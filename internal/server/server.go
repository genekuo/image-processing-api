package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/genekuo/image-processing-api/internal/cache"
	"github.com/genekuo/image-processing-api/internal/config"
	"github.com/genekuo/image-processing-api/internal/handler"
	"github.com/genekuo/image-processing-api/internal/service"
)

// Server wraps the HTTP server and its dependencies including the image cache.
type Server struct {
	httpServer *http.Server
	cache      *cache.Cache
}

// New creates a fully-wired Server from the provided configuration. It
// initialises the cache, downloader, image handler, registers routes, and
// applies the middleware chain: CORS -> Metrics -> Logging -> routes.
func New(cfg *config.Config) *Server {
	c := cache.New(cfg.CacheTTL)
	dl := service.NewDownloader(cfg.MaxSourceSize)
	imgHandler := handler.NewImageHandler(dl, c, cfg)

	mux := http.NewServeMux()
	RegisterRoutes(mux, imgHandler)

	// Middleware chain: outermost runs first.
	// CORS -> Metrics -> Logging -> routes
	var h http.Handler = mux
	h = LoggingMiddleware(h)
	h = MetricsMiddleware(h)
	h = CORSMiddleware(h)

	return &Server{
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Port),
			Handler: h,
		},
		cache: c,
	}
}

// Start begins listening for HTTP requests in a background goroutine.
// It returns immediately. If the listener fails to start, the error is logged.
func (s *Server) Start() error {
	slog.Info("server listening", "addr", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		return fmt.Errorf("listen and serve: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the HTTP server within the given context deadline
// and stops the cache eviction goroutine.
func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("shutting down server")
	s.cache.Stop()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}
	slog.Info("server stopped")
	return nil
}
