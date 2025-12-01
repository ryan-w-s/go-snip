//go:build fyne
// +build fyne

package ui

import (
	"errors"
	"image"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type promptResult struct {
	name string
	save bool
	err  error
}

// PromptSave shows a post-capture dialog with a small preview and a name field.
// If the user clicks Save, save=true. If the user clicks Delete or closes the window, save=false.
func PromptSave(img image.Image) (name string, save bool, err error) {
	a := fyne.CurrentApp()
	if a == nil {
		return "", false, ErrPromptUnavailable
	}
	if a.Driver() == nil {
		return "", false, errors.New("ui: fyne driver unavailable (app not running?)")
	}

	done := make(chan promptResult, 1)
	var once sync.Once
	send := func(res promptResult) {
		once.Do(func() {
			select {
			case done <- res:
			default:
			}
		})
	}

	fyne.DoAndWait(func() {
		w := a.NewWindow("go-snip: screenshot")
		w.Resize(fyne.NewSize(560, 420))

		preview := canvas.NewImageFromImage(img)
		preview.FillMode = canvas.ImageFillContain
		preview.SetMinSize(fyne.NewSize(520, 280))

		nameEntry := widget.NewEntry()
		nameEntry.SetPlaceHolder("Optional name (appended to filename)")

		doSave := func() {
			send(promptResult{name: strings.TrimSpace(nameEntry.Text), save: true})
			w.Close()
		}

		saveBtn := widget.NewButton("Save", func() {
			doSave()
		})

		deleteBtn := widget.NewButton("Delete", func() {
			send(promptResult{save: false})
			w.Close()
		})

		// Hitting Enter in the name field saves.
		nameEntry.OnSubmitted = func(string) {
			doSave()
		}

		w.SetOnClosed(func() {
			send(promptResult{save: false})
		})

		content := container.NewVBox(
			widget.NewLabel("Preview"),
			container.NewPadded(preview),
			widget.NewSeparator(),
			widget.NewLabel("Name (optional)"),
			nameEntry,
			container.NewHBox(layout.NewSpacer(), deleteBtn, saveBtn),
		)
		w.SetContent(container.NewPadded(content))
		w.Show()
	})

	res := <-done
	return res.name, res.save, res.err
}
