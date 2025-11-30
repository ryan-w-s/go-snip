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
)

type settingsResult struct {
	outDir string
	saved  bool
	err    error
}

// ShowSettings opens a settings window (currently only output directory).
// Closing the window returns saved=false unless the user clicks Save.
func ShowSettings(initialOutDir string) (newOutDir string, saved bool, err error) {
	a := fyne.CurrentApp()
	if a == nil {
		return "", false, ErrSettingsUnavailable
	}
	d := a.Driver()
	if d == nil {
		return "", false, errors.New("ui: fyne driver unavailable (app not running?)")
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

	d.DoFromGoroutine(func() {
		w := a.NewWindow("go-snip: settings")
		w.Resize(fyne.NewSize(520, 180))

		outEntry := widget.NewEntry()
		outEntry.SetText(initialOutDir)
		outEntry.SetPlaceHolder("Output directory (e.g. C:\\screenshots)")

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
			send(settingsResult{outDir: strings.TrimSpace(outEntry.Text), saved: true})
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
			container.NewHBox(layout.NewSpacer(), closeBtn, saveBtn),
		)
		w.SetContent(container.NewPadded(form))
		w.Show()
	}, true)

	res := <-done
	return res.outDir, res.saved, res.err
}
