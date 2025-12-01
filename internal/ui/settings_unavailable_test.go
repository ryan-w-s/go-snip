//go:build !fyne

package ui

import (
	"errors"
	"testing"
)

func TestShowSettings_UnavailableWithoutFyne(t *testing.T) {
	t.Parallel()

	_, _, err := ShowSettings("C:\\test", false)
	if !errors.Is(err, ErrSettingsUnavailable) {
		t.Fatalf("expected ErrSettingsUnavailable, got=%v", err)
	}
}
