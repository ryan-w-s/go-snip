package overlay

import (
	"image"
	"testing"
)

func TestOverlayPackageBuilds(t *testing.T) {
	t.Parallel()
}

func TestNormalizeRect(t *testing.T) {
	t.Parallel()

	r := image.Rect(10, 20, 5, 15) // inverted
	got := normalizeRect(r)
	want := image.Rect(5, 15, 10, 20)
	if got != want {
		t.Fatalf("normalizeRect: got=%v want=%v", got, want)
	}
}

func TestCanvasRectToScreenRect_Scale1(t *testing.T) {
	t.Parallel()

	display := image.Rect(100, 200, 1100, 700) // 1000x500
	canvasSize := CanvasSize{W: 1000, H: 500}

	start := CanvasPos{X: 10, Y: 20}
	end := CanvasPos{X: 110, Y: 120}

	got := canvasRectToScreenRect(start, end, canvasSize, display)
	want := image.Rect(110, 220, 210, 320)
	if got != want {
		t.Fatalf("canvasRectToScreenRect(scale1): got=%v want=%v", got, want)
	}
}

func TestCanvasRectToScreenRect_ScaledCanvas(t *testing.T) {
	t.Parallel()

	display := image.Rect(100, 200, 1100, 700) // 1000x500
	canvasSize := CanvasSize{W: 500, H: 250}   // 2x scaling to pixels

	start := CanvasPos{X: 10, Y: 20}
	end := CanvasPos{X: 110, Y: 120}

	got := canvasRectToScreenRect(start, end, canvasSize, display)
	want := image.Rect(120, 240, 320, 440)
	if got != want {
		t.Fatalf("canvasRectToScreenRect(scaled): got=%v want=%v", got, want)
	}
}

func TestCanvasRectToScreenRect_ClampsToDisplay(t *testing.T) {
	t.Parallel()

	display := image.Rect(100, 200, 1100, 700) // 1000x500
	canvasSize := CanvasSize{W: 500, H: 250}

	start := CanvasPos{X: -10, Y: -10}
	end := CanvasPos{X: 600, Y: 300}

	got := canvasRectToScreenRect(start, end, canvasSize, display)
	// Before clamp: x=[80..1300], y=[180..800] => after clamp => display bounds.
	want := display
	if got != want {
		t.Fatalf("canvasRectToScreenRect(clamp): got=%v want=%v", got, want)
	}
}

func TestCanvasRectToScreenRect_ZeroSize(t *testing.T) {
	t.Parallel()

	display := image.Rect(0, 0, 100, 100)
	got := canvasRectToScreenRect(CanvasPos{X: 0, Y: 0}, CanvasPos{X: 10, Y: 10}, CanvasSize{W: 0, H: 100}, display)
	if got.Dx() != 0 || got.Dy() != 0 {
		t.Fatalf("expected empty rect, got=%v", got)
	}
}
