//go:build !fyne

package ui

import "go-snip/internal/config"

// ShowSettings is unavailable unless built with the `fyne` build tag.
func ShowSettings(initialOutDir string, initialPostCapturePrompt bool) (newCfg config.Config, saved bool, err error) {
	return config.Config{}, false, ErrSettingsUnavailable
}
