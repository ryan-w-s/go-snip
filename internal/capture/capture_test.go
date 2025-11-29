package capture

import (
	"image"
	"image/color"
	"testing"
)

func TestCapturePackageBuilds(t *testing.T) {
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

func TestClampRect(t *testing.T) {
	t.Parallel()

	bounds := image.Rect(0, 0, 100, 100)

	// Partially outside.
	r1 := image.Rect(-10, -10, 10, 10)
	got1 := clampRect(normalizeRect(r1), bounds)
	want1 := image.Rect(0, 0, 10, 10)
	if got1 != want1 {
		t.Fatalf("clampRect(1): got=%v want=%v", got1, want1)
	}

	// Fully outside => empty intersection.
	r2 := image.Rect(-20, -20, -10, -10)
	got2 := clampRect(normalizeRect(r2), bounds)
	if got2.Dx() != 0 || got2.Dy() != 0 {
		t.Fatalf("clampRect(2): expected empty, got=%v", got2)
	}
}

func TestClampDisplayIndex(t *testing.T) {
	t.Parallel()

	n := 3
	cases := []struct {
		in   int
		want int
	}{
		{in: -1, want: 0},
		{in: 0, want: 0},
		{in: 2, want: 2},
		{in: 3, want: 2},
		{in: 100, want: 2},
	}

	for _, tc := range cases {
		got := clampDisplayIndex(tc.in, n)
		if got != tc.want {
			t.Fatalf("clampDisplayIndex(%d,%d): got=%d want=%d", tc.in, n, got, tc.want)
		}
	}
}

func TestCrop_NilImage(t *testing.T) {
	t.Parallel()

	_, err := Crop(nil, image.Rect(0, 0, 10, 10))
	if err == nil {
		t.Fatalf("expected error for nil image")
	}
}

func TestCrop_EmptyAfterClamp(t *testing.T) {
	t.Parallel()

	src := image.NewRGBA(image.Rect(0, 0, 10, 10))
	_, err := Crop(src, image.Rect(100, 100, 200, 200))
	if err == nil {
		t.Fatalf("expected error for empty crop")
	}
}

func TestCrop_CopiesPixelsAndNormalizesBounds(t *testing.T) {
	t.Parallel()

	src := image.NewRGBA(image.Rect(0, 0, 3, 3))
	// Color a few pixels for validation.
	src.Set(1, 1, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	src.Set(2, 1, color.RGBA{R: 40, G: 50, B: 60, A: 255})
	src.Set(1, 2, color.RGBA{R: 70, G: 80, B: 90, A: 255})
	src.Set(2, 2, color.RGBA{R: 100, G: 110, B: 120, A: 255})

	// Crop bottom-right 2x2 area.
	out, err := Crop(src, image.Rect(1, 1, 3, 3))
	if err != nil {
		t.Fatalf("Crop error: %v", err)
	}

	if out.Bounds() != image.Rect(0, 0, 2, 2) {
		t.Fatalf("unexpected out bounds: got=%v want=%v", out.Bounds(), image.Rect(0, 0, 2, 2))
	}

	// Validate mapping: out(0,0) == src(1,1), etc.
	check := func(xOut, yOut, xSrc, ySrc int) {
		got := out.At(xOut, yOut)
		want := src.At(xSrc, ySrc)
		if got != want {
			t.Fatalf("pixel mismatch out(%d,%d) got=%v want=%v (src(%d,%d))", xOut, yOut, got, want, xSrc, ySrc)
		}
	}
	check(0, 0, 1, 1)
	check(1, 0, 2, 1)
	check(0, 1, 1, 2)
	check(1, 1, 2, 2)
}
