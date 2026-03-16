package handler

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/genekuo/image-processing-api/internal/cache"
	"github.com/genekuo/image-processing-api/internal/config"
	"github.com/genekuo/image-processing-api/internal/service"
)

// testPNG creates a minimal valid PNG byte slice with the given dimensions.
func testPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encoding test PNG: %v", err)
	}
	return buf.Bytes()
}

// newTestHandler creates an ImageHandler wired to a test image server that
// serves the given PNG data. It returns the handler and the test server URL.
func newTestHandler(t *testing.T, pngData []byte) (*ImageHandler, string) {
	t.Helper()

	imgSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(pngData)
	}))
	t.Cleanup(imgSrv.Close)

	cfg := &config.Config{
		Port:            0,
		MaxSourceSize:   50 * 1024 * 1024,
		MaxOutputWidth:  1400,
		MaxOutputHeight: 1400,
		CacheTTL:        5 * time.Minute,
	}
	c := cache.New(cfg.CacheTTL)
	t.Cleanup(c.Stop)
	dl := service.NewDownloader(cfg.MaxSourceSize)
	h := NewImageHandler(dl, c, cfg)

	return h, imgSrv.URL
}

func TestServeHTTP_SuccessfulProcessing(t *testing.T) {
	pngData := testPNG(t, 100, 100)
	h, imgURL := newTestHandler(t, pngData)

	req := httptest.NewRequest(http.MethodGet, "/image?url="+imgURL+"/test.png&op=rotate-90", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("expected Content-Type image/png, got %q", ct)
	}

	// Verify the response is a valid PNG.
	_, err := png.Decode(bytes.NewReader(rec.Body.Bytes()))
	if err != nil {
		t.Fatalf("response is not valid PNG: %v", err)
	}
}

func TestServeHTTP_MissingURL(t *testing.T) {
	pngData := testPNG(t, 10, 10)
	h, _ := newTestHandler(t, pngData)

	req := httptest.NewRequest(http.MethodGet, "/image?op=rotate-90", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("expected Content-Type image/png, got %q", ct)
	}
}

func TestServeHTTP_MissingOp(t *testing.T) {
	pngData := testPNG(t, 10, 10)
	h, imgURL := newTestHandler(t, pngData)

	req := httptest.NewRequest(http.MethodGet, "/image?url="+imgURL+"/test.png", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("expected Content-Type image/png, got %q", ct)
	}
}

func TestServeHTTP_InvalidOperation(t *testing.T) {
	pngData := testPNG(t, 10, 10)
	h, imgURL := newTestHandler(t, pngData)

	req := httptest.NewRequest(http.MethodGet, "/image?url="+imgURL+"/test.png&op=flip-horizontal", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("expected Content-Type image/png, got %q", ct)
	}
}

func TestServeHTTP_CacheHit(t *testing.T) {
	pngData := testPNG(t, 100, 100)
	h, imgURL := newTestHandler(t, pngData)

	target := "/image?url=" + imgURL + "/test.png&op=rotate-90"

	// First request: populates cache.
	req1 := httptest.NewRequest(http.MethodGet, target, nil)
	rec1 := httptest.NewRecorder()
	h.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("first request: expected 200, got %d", rec1.Code)
	}

	// Second request: should be a cache hit.
	req2 := httptest.NewRequest(http.MethodGet, target, nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("second request: expected 200, got %d", rec2.Code)
	}
	if ct := rec2.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("expected Content-Type image/png, got %q", ct)
	}

	// Verify the two responses have identical bodies (same cached data).
	if !bytes.Equal(rec1.Body.Bytes(), rec2.Body.Bytes()) {
		t.Fatal("expected identical responses from cache hit")
	}
}

func TestServeHTTP_ChainedOperations(t *testing.T) {
	// Start with 200x100, rotate-90 -> 100x200, resize to 50x50.
	pngData := testPNG(t, 200, 100)
	h, imgURL := newTestHandler(t, pngData)

	req := httptest.NewRequest(http.MethodGet, "/image?url="+imgURL+"/test.png&op=rotate-90,resize-50x50", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("expected Content-Type image/png, got %q", ct)
	}

	// Decode and verify dimensions.
	img, err := png.Decode(bytes.NewReader(rec.Body.Bytes()))
	if err != nil {
		t.Fatalf("response is not valid PNG: %v", err)
	}
	b := img.Bounds()
	if b.Dx() != 50 || b.Dy() != 50 {
		t.Fatalf("expected 50x50, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestServeHTTP_ContentTypeOnError(t *testing.T) {
	// Verify that error responses also have Content-Type: image/png.
	pngData := testPNG(t, 10, 10)
	h, _ := newTestHandler(t, pngData)

	tests := []struct {
		name   string
		target string
	}{
		{"missing_url", "/image?op=rotate-90"},
		{"missing_op", "/image?url=http://example.com/img.png"},
		{"invalid_op", "/image?url=http://example.com/img.png&op=bad"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
				t.Errorf("expected Content-Type image/png, got %q", ct)
			}
		})
	}
}

func TestServeHTTP_DownloadFailure(t *testing.T) {
	// Server that returns an error
	failSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(failSrv.Close)

	cfg := &config.Config{
		Port:            0,
		MaxSourceSize:   50 * 1024 * 1024,
		MaxOutputWidth:  1400,
		MaxOutputHeight: 1400,
		CacheTTL:        5 * time.Minute,
	}
	c := cache.New(cfg.CacheTTL)
	t.Cleanup(c.Stop)
	dl := service.NewDownloader(cfg.MaxSourceSize)
	h := NewImageHandler(dl, c, cfg)

	req := httptest.NewRequest(http.MethodGet, "/image?url="+failSrv.URL+"/bad.png&op=resize-100x100", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d", rec.Code)
	}
}

func TestExtractDimensions(t *testing.T) {
	tests := []struct {
		name       string
		ops        []service.Operation
		wantWidth  int
		wantHeight int
	}{
		{
			name:       "no resize",
			ops:        []service.Operation{{Type: "rotate", Angle: 90}},
			wantWidth:  0,
			wantHeight: 0,
		},
		{
			name:       "with resize",
			ops:        []service.Operation{{Type: "rotate", Angle: 90}, {Type: "resize", Width: 200, Height: 100}},
			wantWidth:  200,
			wantHeight: 100,
		},
		{
			name:       "multiple resize uses last",
			ops:        []service.Operation{{Type: "resize", Width: 100, Height: 50}, {Type: "resize", Width: 300, Height: 200}},
			wantWidth:  300,
			wantHeight: 200,
		},
		{
			name:       "empty ops",
			ops:        nil,
			wantWidth:  0,
			wantHeight: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h := extractDimensions(tt.ops)
			if w != tt.wantWidth || h != tt.wantHeight {
				t.Errorf("extractDimensions() = (%d, %d), want (%d, %d)", w, h, tt.wantWidth, tt.wantHeight)
			}
		})
	}
}

func TestCacheKey(t *testing.T) {
	k1 := cacheKey("http://example.com/img.png", "rotate-90")
	k2 := cacheKey("http://example.com/img.png", "rotate-90")
	k3 := cacheKey("http://example.com/img.png", "rotate-180")

	if k1 != k2 {
		t.Error("same inputs should produce same key")
	}
	if k1 == k3 {
		t.Error("different inputs should produce different keys")
	}
}
