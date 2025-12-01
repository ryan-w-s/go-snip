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

// SanitizeFilenameComponent returns a string safe to use as a filename component on Windows.
//
// It trims whitespace, replaces reserved characters with '_' and removes ASCII control chars.
// It also removes trailing dots/spaces (not allowed by Windows) and prefixes '_' for reserved
// base names such as CON, PRN, AUX, NUL, COM1..COM9, LPT1..LPT9.
func SanitizeFilenameComponent(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		// Windows disallows ASCII control chars in filenames.
		if r >= 0 && r < 32 {
			continue
		}
		switch r {
		case '<', '>', ':', '"', '/', '\\', '|', '?', '*':
			b.WriteByte('_')
		default:
			b.WriteRune(r)
		}
	}

	out := strings.TrimSpace(b.String())
	out = strings.TrimRight(out, ". ")
	if out == "" {
		return ""
	}
	if isWindowsReservedBaseName(strings.ToUpper(out)) {
		out = "_" + out
	}
	return out
}

func isWindowsReservedBaseName(upper string) bool {
	switch upper {
	case "CON", "PRN", "AUX", "NUL":
		return true
	}
	if len(upper) == 4 && strings.HasPrefix(upper, "COM") {
		c := upper[3]
		return c >= '1' && c <= '9'
	}
	if len(upper) == 4 && strings.HasPrefix(upper, "LPT") {
		c := upper[3]
		return c >= '1' && c <= '9'
	}
	return false
}

// BaseNameForTimeAndName returns a deterministic base filename (no extension) in local time:
// YYYYMMDD_HHMMSS_mmm or YYYYMMDD_HHMMSS_mmm - <name>
func BaseNameForTimeAndName(t time.Time, rawName string) string {
	ts := strings.TrimSuffix(FilenameForTime(t), ".png")
	name := SanitizeFilenameComponent(rawName)
	if name == "" {
		return ts
	}
	return ts + " - " + name
}

// UniquePathWithBase returns a destination path inside dir for the provided base (no extension).
// If the base filename already exists, it appends a counter suffix:
// ... - 001.png, ... - 002.png, ...
//
// The exists function is injected for testability.
func UniquePathWithBase(dir string, base string, exists func(path string) bool) string {
	base = strings.TrimSpace(base)
	if base == "" {
		// Fall back to something sane (callers generally pass a timestamp base).
		base = "screenshot"
	}

	candidate := filepath.Join(dir, base+".png")
	if !exists(candidate) {
		return candidate
	}
	for i := 1; ; i++ {
		name := fmt.Sprintf("%s - %03d.png", base, i)
		candidate = filepath.Join(dir, name)
		if !exists(candidate) {
			return candidate
		}
	}
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
