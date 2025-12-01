package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds user-configurable settings for go-snip.
type Config struct {
	// OutputDir is the directory where screenshots are saved.
	// If empty, callers should fall back to other sources (env/flags/default).
	OutputDir string `json:"outputDir"`

	// PostCapturePrompt enables showing a post-capture dialog that lets the user
	// preview, name, and choose Save/Delete before writing the file.
	PostCapturePrompt bool `json:"postCapturePrompt"`
}

// DefaultPath returns the per-user config file path:
// <UserConfigDir>/go-snip/config.json
func DefaultPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "go-snip", "config.json"), nil
}

// Load loads the config from path. If the file does not exist, it returns a zero Config and nil error.
func Load(path string) (Config, error) {
	if path == "" {
		return Config{}, errors.New("config path is empty")
	}

	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse %q: %w", path, err)
	}
	return cfg, nil
}

// Save writes cfg to path as JSON, creating parent directories as needed.
// It writes atomically via a temp file + rename.
func Save(path string, cfg Config) error {
	if path == "" {
		return errors.New("config path is empty")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')

	tmp, err := os.CreateTemp(dir, "config-*.tmp")
	if err != nil {
		return err
	}

	tmpName := tmp.Name()
	writeErr := func() error {
		if _, err := tmp.Write(b); err != nil {
			return err
		}
		return tmp.Close()
	}()

	if writeErr != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return writeErr
	}

	if err := atomicReplace(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return nil
}

func atomicReplace(srcTmp, dest string) error {
	// Best-effort: os.Rename won't overwrite on Windows; remove dest then rename.
	if err := os.Rename(srcTmp, dest); err == nil {
		return nil
	} else if errors.Is(err, os.ErrExist) || errors.Is(err, os.ErrPermission) {
		_ = os.Remove(dest)
		return os.Rename(srcTmp, dest)
	} else {
		// Try again after a remove; some platforms return different errors.
		_ = os.Remove(dest)
		if err2 := os.Rename(srcTmp, dest); err2 == nil {
			return nil
		}
		return err
	}
}
