//go:build fyne
// +build fyne

package ui

import (
	"errors"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"go-snip/internal/config"
)

type settingsResult struct {
	cfg   config.Config
	saved bool
	err   error
}

// ShowSettings opens a settings window.
// Closing the window returns saved=false unless the user clicks Save.
func ShowSettings(initialOutDir string, initialPostCapturePrompt bool) (newCfg config.Config, saved bool, err error) {
	a := fyne.CurrentApp()
	if a == nil {
		return config.Config{}, false, ErrSettingsUnavailable
	}
	if a.Driver() == nil {
		return config.Config{}, false, errors.New("ui: fyne driver unavailable (app not running?)")
	}

	done := make(chan settingsResult, 1)
	var once sync.Once
	send := func(res settingsResult) {
		once.Do(func() {
			select {
			case done <- res:
			default:
			}
		})
	}

	// Important: Fyne UI must be mutated on the main/UI goroutine. Using fyne.DoAndWait
	// keeps us compatible with Fyne's thread-safety checks (and avoids window lifecycle hangs).
	fyne.DoAndWait(func() {
		w := a.NewWindow("go-snip: settings")
		w.Resize(fyne.NewSize(560, 240))

		outEntry := widget.NewEntry()
		outEntry.SetText(initialOutDir)
		outEntry.SetPlaceHolder("Output directory (e.g. C:\\screenshots)")

		postPrompt := widget.NewCheck("Ask for a name after capture (preview + Save/Delete)", func(bool) {})
		postPrompt.SetChecked(initialPostCapturePrompt)

		browseBtn := widget.NewButton("Browse…", func() {
			fd := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				if uri == nil {
					// user cancelled
					return
				}
				outEntry.SetText(uri.Path())
			}, w)
			fd.Show()
		})

		saveBtn := widget.NewButton("Save", func() {
			send(settingsResult{
				cfg: config.Config{
					OutputDir:         strings.TrimSpace(outEntry.Text),
					PostCapturePrompt: postPrompt.Checked,
				},
				saved: true,
			})
			w.Close()
		})

		closeBtn := widget.NewButton("Close", func() {
			w.Close()
		})

		w.SetOnClosed(func() {
			send(settingsResult{saved: false})
		})

		form := container.NewVBox(
			widget.NewLabel("Output directory"),
			container.NewBorder(nil, nil, nil, browseBtn, outEntry),
			widget.NewSeparator(),
			postPrompt,
			container.NewHBox(layout.NewSpacer(), closeBtn, saveBtn),
		)
		w.SetContent(container.NewPadded(form))
		w.Show()
	})

	res := <-done
	return res.cfg, res.saved, res.err
}
