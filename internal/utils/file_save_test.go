package utils

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEnsureDir_CreatesNested(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "a", "b", "c")

	if err := EnsureDir(target); err != nil {
		t.Fatalf("EnsureDir() error = %v", err)
	}

	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected directory, got mode=%v", info.Mode())
	}
}

func TestEnsureDir_Empty(t *testing.T) {
	t.Parallel()

	if err := EnsureDir(""); err == nil {
		t.Fatalf("expected error for empty path")
	}
}

func TestFilenameForTime_Format(t *testing.T) {
	t.Parallel()

	tt := time.Date(2025, 1, 2, 3, 4, 5, 678*int(time.Millisecond), time.Local)
	got := FilenameForTime(tt)
	want := "20250102_030405_678.png"

	if got != want {
		t.Fatalf("FilenameForTime() = %q, want %q", got, want)
	}
}

func TestUniquePath_NoCollision(t *testing.T) {
	t.Parallel()

	dir := filepath.Join("some", "dir")
	tt := time.Date(2025, 1, 2, 3, 4, 5, 0, time.Local)

	got := UniquePath(dir, tt, func(string) bool { return false })
	want := filepath.Join(dir, FilenameForTime(tt))

	if got != want {
		t.Fatalf("UniquePath() = %q, want %q", got, want)
	}
}

func TestUniquePath_WithCollisions(t *testing.T) {
	t.Parallel()

	dir := filepath.Join("some", "dir")
	tt := time.Date(2025, 1, 2, 3, 4, 5, 111*int(time.Millisecond), time.Local)

	baseName := FilenameForTime(tt)
	baseNameNoExt := baseName[:len(baseName)-len(".png")]

	base := filepath.Join(dir, baseName)
	collide1 := filepath.Join(dir, baseNameNoExt+"_001.png")

	existsSet := map[string]bool{
		base:     true,
		collide1: true,
	}

	got := UniquePath(dir, tt, func(p string) bool { return existsSet[p] })
	want := filepath.Join(dir, baseNameNoExt+"_002.png")

	if got != want {
		t.Fatalf("UniquePath() = %q, want %q", got, want)
	}
}

func TestSavePNG_WritesValidPNG(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 3, 2))
	img.Set(1, 1, color.RGBA{R: 10, G: 20, B: 30, A: 255})

	dest := filepath.Join(t.TempDir(), "out.png")
	if err := SavePNG(img, dest); err != nil {
		t.Fatalf("SavePNG() error = %v", err)
	}

	f, err := os.Open(dest)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer f.Close()

	decoded, err := png.Decode(f)
	if err != nil {
		t.Fatalf("png.Decode() error = %v", err)
	}

	if decoded.Bounds().Dx() != 3 || decoded.Bounds().Dy() != 2 {
		t.Fatalf("decoded bounds = %v, want 3x2", decoded.Bounds())
	}
}
