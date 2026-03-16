// Package placeholder generates PNG placeholder images for error responses.
// Images display the HTTP status code centered on a color-coded background:
// orange for 4xx client errors, red for 5xx server errors, and gray otherwise.
package placeholder

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	"github.com/disintegration/imaging"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	// defaultWidth is used when width is 0.
	defaultWidth = 400
	// defaultHeight is used when height is 0.
	defaultHeight = 300
	// maxDimension is the upper bound for width and height.
	maxDimension = 1400
)

var (
	// colorOrange is the background for 4xx errors.
	colorOrange = color.RGBA{R: 0xFF, G: 0x8C, B: 0x00, A: 0xFF}
	// colorRed is the background for 5xx errors.
	colorRed = color.RGBA{R: 0xDC, G: 0x14, B: 0x3C, A: 0xFF}
	// colorGray is the background for all other status codes.
	colorGray = color.RGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xFF}
)

// Generate creates a PNG-encoded placeholder image displaying the given HTTP
// status code. Width and height default to 400x300 when zero and are clamped
// to a maximum of 1400x1400.
func Generate(statusCode int, width, height int) ([]byte, error) {
	if width <= 0 {
		width = defaultWidth
	}
	if height <= 0 {
		height = defaultHeight
	}
	if width > maxDimension {
		width = maxDimension
	}
	if height > maxDimension {
		height = maxDimension
	}

	bg := backgroundColor(statusCode)

	// Create the output image and fill with the background color.
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)

	// Render the status code text centered on the image.
	drawCenteredText(img, fmt.Sprintf("%d", statusCode), width, height)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("placeholder: failed to encode PNG: %w", err)
	}
	return buf.Bytes(), nil
}

// backgroundColor returns the fill color based on the HTTP status code range.
func backgroundColor(code int) color.RGBA {
	switch {
	case code >= 400 && code < 500:
		return colorOrange
	case code >= 500 && code < 600:
		return colorRed
	default:
		return colorGray
	}
}

// drawCenteredText renders the status code string onto img, scaled up for
// visibility. It first draws the text at basicfont size onto a small temporary
// image, scales it up using imaging.Resize, then pastes it centered.
func drawCenteredText(img *image.RGBA, text string, width, height int) {
	face := basicfont.Face7x13

	// Measure text bounds using the basic font (7 pixels wide, 13 pixels tall).
	charWidth := 7
	charHeight := 13
	textWidth := len(text) * charWidth
	textHeight := charHeight

	// Add a small margin around the text in the temporary image.
	margin := 2
	tmpW := textWidth + margin*2
	tmpH := textHeight + margin*2

	// Draw white text on a transparent background.
	tmp := image.NewRGBA(image.Rect(0, 0, tmpW, tmpH))

	d := &font.Drawer{
		Dst:  tmp,
		Src:  image.NewUniform(color.White),
		Face: face,
		// Dot is the baseline origin; ascent is ~11px for Face7x13.
		Dot: fixed.P(margin, margin+charHeight-2),
	}
	d.DrawString(text)

	// Determine the scale factor so the text fills roughly 60% of the
	// smaller image dimension, but never exceeds the image bounds.
	targetH := int(float64(height) * 0.4)
	if targetH < 1 {
		targetH = 1
	}
	scaleFactor := float64(targetH) / float64(tmpH)
	targetW := int(float64(tmpW) * scaleFactor)
	if targetW > int(float64(width)*0.8) {
		targetW = int(float64(width) * 0.8)
		scaleFactor = float64(targetW) / float64(tmpW)
		targetH = int(float64(tmpH) * scaleFactor)
	}
	if targetW < 1 {
		targetW = 1
	}
	if targetH < 1 {
		targetH = 1
	}

	scaled := imaging.Resize(tmp, targetW, targetH, imaging.NearestNeighbor)

	// Paste scaled text centered on the output image.
	offsetX := (width - scaled.Bounds().Dx()) / 2
	offsetY := (height - scaled.Bounds().Dy()) / 2
	draw.Draw(
		img,
		image.Rect(offsetX, offsetY, offsetX+scaled.Bounds().Dx(), offsetY+scaled.Bounds().Dy()),
		scaled,
		scaled.Bounds().Min,
		draw.Over,
	)
}
