package main

import (
	"image"
	"testing"
)

func TestMainPackageBuilds(t *testing.T) {
	t.Parallel()
}

func TestResolveOutputDir_PrefersFlag(t *testing.T) {
	t.Parallel()

	got := resolveOutputDir(`C:\custom\out`, func(string) (string, bool) {
		return `C:\env\out`, true
	}, `C:\config\out`)
	if got != `C:\custom\out` {
		t.Fatalf("got=%q want=%q", got, `C:\custom\out`)
	}
}

func TestResolveOutputDir_PrefersEnvOverDefault(t *testing.T) {
	t.Parallel()

	got := resolveOutputDir("", func(key string) (string, bool) {
		if key != outputDirEnv {
			t.Fatalf("unexpected env key: %q", key)
		}
		return `C:\env\out`, true
	}, `C:\config\out`)
	if got != `C:\env\out` {
		t.Fatalf("got=%q want=%q", got, `C:\env\out`)
	}
}

func TestResolveOutputDir_PrefersConfigOverDefault(t *testing.T) {
	t.Parallel()

	got := resolveOutputDir("", func(string) (string, bool) { return "", false }, `C:\config\out`)
	if got != `C:\config\out` {
		t.Fatalf("got=%q want=%q", got, `C:\config\out`)
	}
}

func TestResolveOutputDir_Default(t *testing.T) {
	t.Parallel()

	got := resolveOutputDir("", func(string) (string, bool) { return "", false }, "")
	if got == "" {
		t.Fatalf("expected non-empty default output dir")
	}
}

func TestCropRectFor_ImageBoundsMatchDisplay(t *testing.T) {
	t.Parallel()

	display := image.Rect(100, 200, 1100, 700)
	imgBounds := display
	selection := image.Rect(150, 250, 300, 400)

	got := cropRectFor(imgBounds, display, selection)
	if got != selection {
		t.Fatalf("got=%v want=%v", got, selection)
	}
}

func TestCropRectFor_ImageBoundsAtOrigin(t *testing.T) {
	t.Parallel()

	display := image.Rect(100, 200, 1100, 700) // 1000x500
	imgBounds := image.Rect(0, 0, 1000, 500)
	selection := image.Rect(150, 250, 300, 400)

	got := cropRectFor(imgBounds, display, selection)
	want := image.Rect(50, 50, 200, 200)
	if got != want {
		t.Fatalf("got=%v want=%v", got, want)
	}
}

func TestCropRectFor_ImageBoundsHaveOffset(t *testing.T) {
	t.Parallel()

	display := image.Rect(100, 200, 1100, 700) // 1000x500
	imgBounds := image.Rect(10, 20, 1010, 520)
	selection := image.Rect(150, 250, 300, 400)

	got := cropRectFor(imgBounds, display, selection)
	want := image.Rect(60, 70, 210, 220)
	if got != want {
		t.Fatalf("got=%v want=%v", got, want)
	}
}
