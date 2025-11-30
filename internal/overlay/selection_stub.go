//go:build !fyne

package overlay

import (
	"errors"
	"image"
)

var ErrSelectionUnavailable = errors.New("overlay: selection UI unavailable (build with -tags=fyne)")

// SelectArea is unavailable unless built with the `fyne` build tag.
func SelectArea() (rect image.Rectangle, cancelled bool, err error) {
	return image.Rectangle{}, false, ErrSelectionUnavailable
}
