package service

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestImage creates a simple RGBA image with the given dimensions.
func newTestImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	return img
}

// encodeJPEG encodes the image as JPEG bytes.
func encodeJPEG(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatalf("encoding JPEG: %v", err)
	}
	return buf.Bytes()
}

// encodePNG encodes the image as PNG bytes.
func encodePNG(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encoding PNG: %v", err)
	}
	return buf.Bytes()
}

func TestDownload_JPEG(t *testing.T) {
	data := encodeJPEG(t, newTestImage(100, 50))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write(data)
	}))
	defer srv.Close()

	d := NewDownloader(50 * 1024 * 1024)
	img, err := d.Download(context.Background(), srv.URL+"/test.jpg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 50 {
		t.Errorf("expected 100x50, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestDownload_PNG(t *testing.T) {
	data := encodePNG(t, newTestImage(80, 60))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(data)
	}))
	defer srv.Close()

	d := NewDownloader(50 * 1024 * 1024)
	img, err := d.Download(context.Background(), srv.URL+"/test.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 80 || bounds.Dy() != 60 {
		t.Errorf("expected 80x60, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestDownload_MaxSizeExceeded(t *testing.T) {
	// Create a PNG that is larger than our tiny limit.
	data := encodePNG(t, newTestImage(200, 200))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(data)
	}))
	defer srv.Close()

	// Set maxSize to something very small so the image body gets truncated.
	d := NewDownloader(10)
	_, err := d.Download(context.Background(), srv.URL+"/big.png")
	if err == nil {
		t.Fatal("expected error for oversized image, got nil")
	}
	if !strings.Contains(err.Error(), "decoding image") {
		t.Errorf("expected decode error, got: %v", err)
	}
}

func TestDownload_InvalidScheme(t *testing.T) {
	d := NewDownloader(50 * 1024 * 1024)

	tests := []struct {
		name string
		url  string
	}{
		{"ftp", "ftp://example.com/image.jpg"},
		{"no scheme", "example.com/image.jpg"},
		{"data URI", "data:image/png;base64,abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := d.Download(context.Background(), tt.url)
			if err == nil {
				t.Fatal("expected error for invalid scheme, got nil")
			}
			if !strings.Contains(err.Error(), "invalid URL scheme") {
				t.Errorf("expected invalid URL scheme error, got: %v", err)
			}
		})
	}
}

func TestDownload_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	d := NewDownloader(50 * 1024 * 1024)
	_, err := d.Download(context.Background(), srv.URL+"/missing.jpg")
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("expected 404 in error, got: %v", err)
	}
}

func TestDownload_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block until request context is cancelled.
		<-r.Context().Done()
	}))
	defer srv.Close()

	d := NewDownloader(50 * 1024 * 1024)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := d.Download(ctx, srv.URL+"/slow.jpg")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "downloading image") {
		t.Errorf("expected download error, got: %v", err)
	}
}
