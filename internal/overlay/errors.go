package overlay

import "errors"

// ErrSelectionUnavailable indicates the selection overlay UI is not available in the current build.
//
// In this repo, the selection UI is enabled by building with the `fyne` build tag.
// This error must exist in both `fyne` and `!fyne` builds so callers can reference it unconditionally.
var ErrSelectionUnavailable = errors.New("overlay: selection UI unavailable (build with -tags=fyne)")
