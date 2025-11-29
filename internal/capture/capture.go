package capture

import (
	"errors"
	"image"
	"image/draw"

	"github.com/kbinani/screenshot"
)

var (
	ErrNoActiveDisplays = errors.New("capture: no active displays")
	ErrNilImage         = errors.New("capture: nil image")
	ErrEmptyCrop        = errors.New("capture: empty crop after clamping")
)

// CaptureDisplay captures a full screenshot of the specified display.
// If displayIndex is out of range, it is clamped to the nearest valid index.
func CaptureDisplay(displayIndex int) (image.Image, error) {
	n := screenshot.NumActiveDisplays()
	if n <= 0 {
		return nil, ErrNoActiveDisplays
	}
	i := clampDisplayIndex(displayIndex, n)
	bounds := screenshot.GetDisplayBounds(i)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// Crop returns a bounds-checked crop of img as a new *image.RGBA (always a copy).
//
// The input rectangle is normalized (Min/Max fixed up), then clamped to img.Bounds().
// If the resulting area is empty, Crop returns ErrEmptyCrop.
//
// The returned image's bounds are normalized to start at (0,0) with size (Dx,Dy).
func Crop(img image.Image, r image.Rectangle) (image.Image, error) {
	if img == nil {
		return nil, ErrNilImage
	}

	norm := normalizeRect(r)
	clamped := clampRect(norm, img.Bounds())
	if clamped.Dx() <= 0 || clamped.Dy() <= 0 {
		return nil, ErrEmptyCrop
	}

	dst := image.NewRGBA(image.Rect(0, 0, clamped.Dx(), clamped.Dy()))
	// Copy from source starting at clamped.Min into dst at (0,0).
	draw.Draw(dst, dst.Bounds(), img, clamped.Min, draw.Src)
	return dst, nil
}

func clampDisplayIndex(displayIndex, numDisplays int) int {
	if numDisplays <= 0 {
		return 0
	}
	if displayIndex < 0 {
		return 0
	}
	if displayIndex >= numDisplays {
		return numDisplays - 1
	}
	return displayIndex
}

func normalizeRect(r image.Rectangle) image.Rectangle {
	minX, maxX := r.Min.X, r.Max.X
	if minX > maxX {
		minX, maxX = maxX, minX
	}
	minY, maxY := r.Min.Y, r.Max.Y
	if minY > maxY {
		minY, maxY = maxY, minY
	}
	return image.Rect(minX, minY, maxX, maxY)
}

func clampRect(r, bounds image.Rectangle) image.Rectangle {
	return r.Intersect(bounds)
}
