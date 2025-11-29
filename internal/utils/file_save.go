package utils

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefaultOutputDir returns the default directory (relative to the working directory)
// where screenshots are saved.
func DefaultOutputDir() string {
	return filepath.Join(".", "screenshots")
}

// EnsureDir creates path (and any parents) if it doesn't already exist.
func EnsureDir(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("path is empty")
	}
	return os.MkdirAll(path, 0o755)
}

// FilenameForTime returns a deterministic PNG filename in local time:
// YYYYMMDD_HHMMSS_mmm.png
func FilenameForTime(t time.Time) string {
	t = t.Local()
	base := t.Format("20060102_150405")
	ms := t.Nanosecond() / int(time.Millisecond)
	return fmt.Sprintf("%s_%03d.png", base, ms)
}

// UniquePath returns a destination path inside dir for the provided time.
// If the base filename already exists, it appends a counter suffix:
// ..._001.png, ..._002.png, ...
//
// The exists function is injected for testability.
func UniquePath(dir string, t time.Time, exists func(path string) bool) string {
	filename := FilenameForTime(t)
	base := strings.TrimSuffix(filename, ".png")

	candidate := filepath.Join(dir, filename)
	if !exists(candidate) {
		return candidate
	}

	for i := 1; ; i++ {
		name := fmt.Sprintf("%s_%03d.png", base, i)
		candidate = filepath.Join(dir, name)
		if !exists(candidate) {
			return candidate
		}
	}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// SavePNG writes img as a PNG to destPath, creating the parent directory if needed.
func SavePNG(img image.Image, destPath string) error {
	if strings.TrimSpace(destPath) == "" {
		return errors.New("destPath is empty")
	}

	if err := EnsureDir(filepath.Dir(destPath)); err != nil {
		return err
	}

	f, err := os.Create(destPath)
	if err != nil {
		return err
	}

	encodeErr := png.Encode(f, img)
	closeErr := f.Close()
	return errors.Join(encodeErr, closeErr)
}
