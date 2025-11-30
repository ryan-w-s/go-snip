//go:build !fyne

package ui

// ShowSettings is unavailable unless built with the `fyne` build tag.
func ShowSettings(initialOutDir string) (newOutDir string, saved bool, err error) {
	return "", false, ErrSettingsUnavailable
}
