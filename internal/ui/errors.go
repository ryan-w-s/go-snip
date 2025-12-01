package ui

import "errors"

// ErrSettingsUnavailable indicates the settings UI is not available in the current build.
//
// In this repo, the settings UI is enabled by building with the `fyne` build tag.
var ErrSettingsUnavailable = errors.New("ui: settings UI unavailable (build with -tags=fyne)")

// ErrPromptUnavailable indicates the post-capture prompt UI is not available in the current build.
//
// In this repo, the prompt UI is enabled by building with the `fyne` build tag.
var ErrPromptUnavailable = errors.New("ui: prompt UI unavailable (build with -tags=fyne)")
