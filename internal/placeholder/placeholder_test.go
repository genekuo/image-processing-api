package placeholder

import (
	"bytes"
	"image/color"
	"image/png"
	"testing"
)

func TestGenerate_4xxReturnsOrangeBackground(t *testing.T) {
	codes := []int{400, 401, 403, 404, 422, 429, 499}
	for _, code := range codes {
		data, err := Generate(code, 100, 100)
		if err != nil {
			t.Fatalf("Generate(%d) returned error: %v", code, err)
		}

		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("failed to decode PNG for %d: %v", code, err)
		}

		// Sample the top-left corner which should be the background color.
		r, g, b, _ := img.At(0, 0).RGBA()
		// colorOrange is #FF8C00 -> RGBA 0xFFFF, 0x8C8C, 0x0000
		if r>>8 != 0xFF || g>>8 != 0x8C || b>>8 != 0x00 {
			t.Errorf("Generate(%d): expected orange background, got (%02X, %02X, %02X)",
				code, r>>8, g>>8, b>>8)
		}
	}
}

func TestGenerate_5xxReturnsRedBackground(t *testing.T) {
	codes := []int{500, 502, 503, 504}
	for _, code := range codes {
		data, err := Generate(code, 100, 100)
		if err != nil {
			t.Fatalf("Generate(%d) returned error: %v", code, err)
		}

		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("failed to decode PNG for %d: %v", code, err)
		}

		r, g, b, _ := img.At(0, 0).RGBA()
		// colorRed is #DC143C -> RGBA 0xDCDC, 0x1414, 0x3C3C
		if r>>8 != 0xDC || g>>8 != 0x14 || b>>8 != 0x3C {
			t.Errorf("Generate(%d): expected red background, got (%02X, %02X, %02X)",
				code, r>>8, g>>8, b>>8)
		}
	}
}

func TestGenerate_OtherCodesReturnGrayBackground(t *testing.T) {
	codes := []int{200, 301, 302, 600, 0}
	for _, code := range codes {
		data, err := Generate(code, 100, 100)
		if err != nil {
			t.Fatalf("Generate(%d) returned error: %v", code, err)
		}

		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("failed to decode PNG for %d: %v", code, err)
		}

		r, g, b, _ := img.At(0, 0).RGBA()
		if r>>8 != 0x80 || g>>8 != 0x80 || b>>8 != 0x80 {
			t.Errorf("Generate(%d): expected gray background, got (%02X, %02X, %02X)",
				code, r>>8, g>>8, b>>8)
		}
	}
}

func TestGenerate_DefaultDimensions(t *testing.T) {
	data, err := Generate(404, 0, 0)
	if err != nil {
		t.Fatalf("Generate with zero dims returned error: %v", err)
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to decode PNG: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != defaultWidth || bounds.Dy() != defaultHeight {
		t.Errorf("expected default dimensions %dx%d, got %dx%d",
			defaultWidth, defaultHeight, bounds.Dx(), bounds.Dy())
	}
}

func TestGenerate_MaxDimensionClamping(t *testing.T) {
	data, err := Generate(500, 5000, 3000)
	if err != nil {
		t.Fatalf("Generate with oversized dims returned error: %v", err)
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to decode PNG: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != maxDimension || bounds.Dy() != maxDimension {
		t.Errorf("expected clamped dimensions %dx%d, got %dx%d",
			maxDimension, maxDimension, bounds.Dx(), bounds.Dy())
	}
}

func TestGenerate_OutputIsValidPNG(t *testing.T) {
	tests := []struct {
		name   string
		code   int
		width  int
		height int
	}{
		{"404 default size", 404, 0, 0},
		{"500 custom size", 500, 200, 150},
		{"200 small", 200, 10, 10},
		{"503 max size", 503, 1400, 1400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := Generate(tt.code, tt.width, tt.height)
			if err != nil {
				t.Fatalf("Generate returned error: %v", err)
			}

			if len(data) == 0 {
				t.Fatal("Generate returned empty byte slice")
			}

			// Verify PNG magic bytes.
			pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
			if !bytes.HasPrefix(data, pngHeader) {
				t.Error("output does not start with PNG magic bytes")
			}

			// Verify full decode round-trip.
			img, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("failed to decode PNG: %v", err)
			}

			if img.Bounds().Dx() == 0 || img.Bounds().Dy() == 0 {
				t.Error("decoded image has zero dimensions")
			}
		})
	}
}

func TestBackgroundColor(t *testing.T) {
	tests := []struct {
		code     int
		expected color.RGBA
	}{
		{400, colorOrange},
		{404, colorOrange},
		{499, colorOrange},
		{500, colorRed},
		{503, colorRed},
		{599, colorRed},
		{200, colorGray},
		{301, colorGray},
		{600, colorGray},
	}

	for _, tt := range tests {
		got := backgroundColor(tt.code)
		if got != tt.expected {
			t.Errorf("backgroundColor(%d) = %v, want %v", tt.code, got, tt.expected)
		}
	}
}
