package service

import (
	"image"
	"image/color"
	"testing"
)

// makeImage creates a simple RGBA image with the given dimensions filled with
// the specified color.
func makeImage(w, h int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func TestParseOperation_Valid(t *testing.T) {
	tests := []struct {
		input    string
		wantType string
		angle    int
		width    int
		height   int
	}{
		{"rotate-90", "rotate", 90, 0, 0},
		{"rotate-180", "rotate", 180, 0, 0},
		{"rotate-270", "rotate", 270, 0, 0},
		{"resize-800x600", "resize", 0, 800, 600},
		{"resize-1x1", "resize", 0, 1, 1},
		{"resize-1400x1400", "resize", 0, 1400, 1400},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			op, err := ParseOperation(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if op.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", op.Type, tt.wantType)
			}
			if op.Angle != tt.angle {
				t.Errorf("Angle = %d, want %d", op.Angle, tt.angle)
			}
			if op.Width != tt.width {
				t.Errorf("Width = %d, want %d", op.Width, tt.width)
			}
			if op.Height != tt.height {
				t.Errorf("Height = %d, want %d", op.Height, tt.height)
			}
		})
	}
}

func TestParseOperation_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"unknown", "flip-horizontal"},
		{"rotate bad angle", "rotate-45"},
		{"resize no dimensions", "resize-"},
		{"resize missing height", "resize-800"},
		{"resize zero width", "resize-0x600"},
		{"resize negative", "resize--1x600"},
		{"resize non-numeric", "resize-abcxdef"},
		{"resize exceeds max width", "resize-1500x600"},
		{"resize exceeds max height", "resize-600x1500"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseOperation(tt.input)
			if err == nil {
				t.Fatalf("expected error for %q, got nil", tt.input)
			}
		})
	}
}

func TestParseOperations_CommaSeparated(t *testing.T) {
	ops, err := ParseOperations("rotate-90,resize-200x100,rotate-180")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ops) != 3 {
		t.Fatalf("expected 3 operations, got %d", len(ops))
	}

	if ops[0].Type != "rotate" || ops[0].Angle != 90 {
		t.Errorf("ops[0] = %+v, want rotate-90", ops[0])
	}
	if ops[1].Type != "resize" || ops[1].Width != 200 || ops[1].Height != 100 {
		t.Errorf("ops[1] = %+v, want resize-200x100", ops[1])
	}
	if ops[2].Type != "rotate" || ops[2].Angle != 180 {
		t.Errorf("ops[2] = %+v, want rotate-180", ops[2])
	}
}

func TestParseOperations_Empty(t *testing.T) {
	_, err := ParseOperations("")
	if err == nil {
		t.Fatal("expected error for empty string, got nil")
	}
}

func TestRotate90_Dimensions(t *testing.T) {
	// 100x50 should become 50x100 after 90-degree rotation.
	src := makeImage(100, 50, color.RGBA{R: 255, A: 255})
	op := Operation{Type: "rotate", Angle: 90}

	result, err := Apply(src, op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b := result.Bounds()
	if b.Dx() != 50 || b.Dy() != 100 {
		t.Errorf("expected 50x100, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestRotate180_PreservesDimensions(t *testing.T) {
	src := makeImage(100, 50, color.RGBA{G: 255, A: 255})
	op := Operation{Type: "rotate", Angle: 180}

	result, err := Apply(src, op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b := result.Bounds()
	if b.Dx() != 100 || b.Dy() != 50 {
		t.Errorf("expected 100x50, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestResize_ExactDimensions(t *testing.T) {
	src := makeImage(200, 100, color.RGBA{B: 255, A: 255})
	op := Operation{Type: "resize", Width: 50, Height: 50}

	result, err := Apply(src, op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b := result.Bounds()
	if b.Dx() != 50 || b.Dy() != 50 {
		t.Errorf("expected 50x50, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestResize_NonSquareSource(t *testing.T) {
	// A wide source (300x100) resized to 100x100 should cover/crop correctly.
	src := makeImage(300, 100, color.RGBA{R: 128, G: 128, A: 255})
	op := Operation{Type: "resize", Width: 100, Height: 100}

	result, err := Apply(src, op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b := result.Bounds()
	if b.Dx() != 100 || b.Dy() != 100 {
		t.Errorf("expected 100x100, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestResize_TallSource(t *testing.T) {
	// A tall source (100x400) resized to 80x80 should cover/crop correctly.
	src := makeImage(100, 400, color.RGBA{R: 50, G: 200, B: 100, A: 255})
	op := Operation{Type: "resize", Width: 80, Height: 80}

	result, err := Apply(src, op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b := result.Bounds()
	if b.Dx() != 80 || b.Dy() != 80 {
		t.Errorf("expected 80x80, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestMaxDimensionValidation(t *testing.T) {
	_, err := ParseOperation("resize-1401x800")
	if err == nil {
		t.Fatal("expected error for width exceeding max, got nil")
	}

	_, err = ParseOperation("resize-800x1401")
	if err == nil {
		t.Fatal("expected error for height exceeding max, got nil")
	}
}

func TestApplyAll_ChainsOperations(t *testing.T) {
	// Start with 200x100, rotate 90 -> 100x200, then resize to 50x50.
	src := makeImage(200, 100, color.RGBA{R: 255, G: 255, A: 255})
	ops := []Operation{
		{Type: "rotate", Angle: 90},
		{Type: "resize", Width: 50, Height: 50},
	}

	result, err := ApplyAll(src, ops)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b := result.Bounds()
	if b.Dx() != 50 || b.Dy() != 50 {
		t.Errorf("expected 50x50, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestApply_UnknownType(t *testing.T) {
	src := makeImage(10, 10, color.RGBA{A: 255})
	_, err := Apply(src, Operation{Type: "blur"})
	if err == nil {
		t.Fatal("expected error for unknown operation type, got nil")
	}
}

func TestRotate270_Dimensions(t *testing.T) {
	src := makeImage(100, 50, color.RGBA{B: 255, A: 255})
	op := Operation{Type: "rotate", Angle: 270}

	result, err := Apply(src, op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b := result.Bounds()
	if b.Dx() != 50 || b.Dy() != 100 {
		t.Errorf("expected 50x100, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestApplyRotate_UnsupportedAngle(t *testing.T) {
	src := makeImage(10, 10, color.RGBA{A: 255})
	op := Operation{Type: "rotate", Angle: 45}

	_, err := Apply(src, op)
	if err == nil {
		t.Fatal("expected error for unsupported rotation angle")
	}
}
