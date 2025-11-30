package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultPath(t *testing.T) {
	t.Parallel()

	p, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() error: %v", err)
	}
	if p == "" {
		t.Fatalf("DefaultPath() returned empty path")
	}
	if !strings.HasSuffix(filepath.ToSlash(p), "go-snip/config.json") {
		t.Fatalf("DefaultPath()=%q, expected suffix go-snip/config.json", p)
	}
}

func TestLoad_MissingFileReturnsDefault(t *testing.T) {
	t.Parallel()

	p := filepath.Join(t.TempDir(), "missing.json")
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.OutputDir != "" {
		t.Fatalf("expected empty config, got=%+v", cfg)
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	t.Parallel()

	p := filepath.Join(t.TempDir(), "config.json")
	orig := Config{OutputDir: `C:\some\dir`}

	if err := Save(p, orig); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("stat saved file error: %v", err)
	}

	got, err := Load(p)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got != orig {
		t.Fatalf("roundtrip mismatch got=%+v want=%+v", got, orig)
	}
}
