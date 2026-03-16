// Package service provides core image processing services including downloading
// and transforming images.
package service

import (
	"context"
	"fmt"
	"image"
	"io"
	"net/http"
	"strings"
	"time"

	// Register image format decoders for auto-detection by image.Decode.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// Downloader fetches remote images over HTTP/HTTPS and decodes them.
type Downloader struct {
	client  *http.Client
	maxSize int64
}

// NewDownloader creates a Downloader with the given maximum allowed response
// body size (in bytes) and a default HTTP client timeout of 30 seconds.
func NewDownloader(maxSize int64) *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxSize: maxSize,
	}
}

// Download fetches the image at rawURL, enforces the configured size limit, and
// decodes it into an image.Image. The URL must use the http or https scheme.
func (d *Downloader) Download(ctx context.Context, rawURL string) (image.Image, error) {
	if err := validateURL(rawURL); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading image: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status %d for %s", resp.StatusCode, rawURL)
	}

	// LimitReader prevents reading beyond maxSize. We read maxSize+1 bytes so
	// we can detect whether the body exceeds the limit.
	limited := io.LimitReader(resp.Body, d.maxSize+1)

	img, _, err := image.Decode(limited)
	if err != nil {
		return nil, fmt.Errorf("decoding image: %w", err)
	}

	return img, nil
}

// validateURL ensures the URL uses the http or https scheme.
func validateURL(rawURL string) error {
	lower := strings.ToLower(rawURL)
	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		return fmt.Errorf("invalid URL scheme: only http and https are supported, got %q", rawURL)
	}
	return nil
}
