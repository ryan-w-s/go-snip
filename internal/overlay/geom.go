package overlay

import (
	"errors"
	"image"
	"math"
)

var (
	ErrNoActiveDisplays = errors.New("overlay: no active displays")
)

type CanvasPos struct {
	X float32
	Y float32
}

type CanvasSize struct {
	W float32
	H float32
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

// canvasRectToScreenRect converts a start/end drag in canvas coordinates into a screen-space
// rectangle compatible with screenshot.CaptureRect on the provided display.
//
// The conversion scales the drag positions by the ratio between the display pixel bounds
// and the canvas logical size, and offsets by displayBounds.Min.
func canvasRectToScreenRect(start, end CanvasPos, canvasSize CanvasSize, displayBounds image.Rectangle) image.Rectangle {
	if canvasSize.W <= 0 || canvasSize.H <= 0 || displayBounds.Dx() <= 0 || displayBounds.Dy() <= 0 {
		return image.Rectangle{}
	}

	sx := float64(displayBounds.Dx()) / float64(canvasSize.W)
	sy := float64(displayBounds.Dy()) / float64(canvasSize.H)

	toScreen := func(p CanvasPos) image.Point {
		// Round to nearest pixel to reduce off-by-one drift under scaling.
		x := int(math.Round(float64(p.X) * sx))
		y := int(math.Round(float64(p.Y) * sy))
		return image.Pt(displayBounds.Min.X+x, displayBounds.Min.Y+y)
	}

	p1 := toScreen(start)
	p2 := toScreen(end)

	r := image.Rectangle{Min: p1, Max: p2}
	r = normalizeRect(r)
	return clampRect(r, displayBounds)
}
