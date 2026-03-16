package service

import (
	"fmt"
	"image"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
)

const (
	// MaxOutputWidth is the maximum allowed output width in pixels.
	MaxOutputWidth = 1400
	// MaxOutputHeight is the maximum allowed output height in pixels.
	MaxOutputHeight = 1400
)

// Operation represents a single image transformation.
type Operation struct {
	Type   string // "rotate" or "resize"
	Angle  int    // rotation angle: 90, 180, or 270
	Width  int    // target width for resize
	Height int    // target height for resize
}

// ParseOperation parses a single operation string such as "rotate-90" or
// "resize-800x600" into an Operation.
func ParseOperation(op string) (Operation, error) {
	op = strings.TrimSpace(op)
	if op == "" {
		return Operation{}, fmt.Errorf("empty operation")
	}

	switch op {
	case "rotate-90":
		return Operation{Type: "rotate", Angle: 90}, nil
	case "rotate-180":
		return Operation{Type: "rotate", Angle: 180}, nil
	case "rotate-270":
		return Operation{Type: "rotate", Angle: 270}, nil
	default:
		if strings.HasPrefix(op, "resize-") {
			return parseResize(op)
		}
		return Operation{}, fmt.Errorf("unknown operation: %q", op)
	}
}

// parseResize handles "resize-WxH" strings.
func parseResize(op string) (Operation, error) {
	dims := strings.TrimPrefix(op, "resize-")
	parts := strings.SplitN(dims, "x", 2)
	if len(parts) != 2 {
		return Operation{}, fmt.Errorf("invalid resize format %q: expected resize-WxH", op)
	}

	w, err := strconv.Atoi(parts[0])
	if err != nil || w <= 0 {
		return Operation{}, fmt.Errorf("invalid resize width in %q: must be a positive integer", op)
	}

	h, err := strconv.Atoi(parts[1])
	if err != nil || h <= 0 {
		return Operation{}, fmt.Errorf("invalid resize height in %q: must be a positive integer", op)
	}

	if w > MaxOutputWidth || h > MaxOutputHeight {
		return Operation{}, fmt.Errorf("resize dimensions %dx%d exceed maximum allowed %dx%d", w, h, MaxOutputWidth, MaxOutputHeight)
	}

	return Operation{Type: "resize", Width: w, Height: h}, nil
}

// ParseOperations parses a comma-separated list of operation strings.
func ParseOperations(ops string) ([]Operation, error) {
	ops = strings.TrimSpace(ops)
	if ops == "" {
		return nil, fmt.Errorf("empty operations string")
	}

	parts := strings.Split(ops, ",")
	result := make([]Operation, 0, len(parts))
	for _, p := range parts {
		op, err := ParseOperation(p)
		if err != nil {
			return nil, err
		}
		result = append(result, op)
	}
	return result, nil
}

// Apply executes a single Operation on the provided image and returns the
// transformed result.
func Apply(img image.Image, op Operation) (image.Image, error) {
	switch op.Type {
	case "rotate":
		return applyRotate(img, op.Angle)
	case "resize":
		return applyResize(img, op.Width, op.Height)
	default:
		return nil, fmt.Errorf("unsupported operation type: %q", op.Type)
	}
}

// applyRotate rotates the image by the given angle.
func applyRotate(img image.Image, angle int) (image.Image, error) {
	switch angle {
	case 90:
		return imaging.Rotate90(img), nil
	case 180:
		return imaging.Rotate180(img), nil
	case 270:
		return imaging.Rotate270(img), nil
	default:
		return nil, fmt.Errorf("unsupported rotation angle: %d", angle)
	}
}

// applyResize uses cover/crop (Fill) to resize the image to exactly WxH,
// preserving aspect ratio by cropping from center.
func applyResize(img image.Image, width, height int) (image.Image, error) {
	return imaging.Fill(img, width, height, imaging.Center, imaging.Lanczos), nil
}

// ApplyAll applies a sequence of operations to an image, returning the final result.
func ApplyAll(img image.Image, ops []Operation) (image.Image, error) {
	var err error
	for i, op := range ops {
		img, err = Apply(img, op)
		if err != nil {
			return nil, fmt.Errorf("operation %d (%s): %w", i, op.Type, err)
		}
	}
	return img, nil
}
