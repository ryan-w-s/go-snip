//go:build !fyne

package ui

import "image"

// PromptSave is unavailable unless built with the `fyne` build tag.
func PromptSave(img image.Image) (name string, save bool, err error) {
	return "", false, ErrPromptUnavailable
}
