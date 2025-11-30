//go:build !fyne

package overlay

import (
	"image"
)

// SelectArea is unavailable unless built with the `fyne` build tag.
func SelectArea() (rect image.Rectangle, cancelled bool, err error) {
	return image.Rectangle{}, false, ErrSelectionUnavailable
}
