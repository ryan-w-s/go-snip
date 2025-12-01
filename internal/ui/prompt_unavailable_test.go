//go:build !fyne

package ui

import (
	"errors"
	"image"
	"testing"
)

func TestPromptSave_UnavailableWithoutFyne(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	_, _, err := PromptSave(img)
	if !errors.Is(err, ErrPromptUnavailable) {
		t.Fatalf("expected ErrPromptUnavailable, got=%v", err)
	}
}
