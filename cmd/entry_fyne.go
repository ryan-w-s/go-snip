//go:build fyne
// +build fyne

package main

import (
	"context"
	"io"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"

	"go-snip/internal/config"
)

func runEntry(ctx context.Context, outDir string, cfgPath string, cfg config.Config, now func() time.Time, out io.Writer) error {
	a := app.NewWithID("go-snip")

	errCh := make(chan error, 1)
	// Important: Fyne's underlying driver (GLFW on desktop) is initialized when the app is started.
	// If we allow hotkeys to call UI APIs before that, fyne.Do[AndWait] can panic with
	// "GLFW library is not initialized".
	var startOnce sync.Once
	var keepAlive fyne.Window
	a.Lifecycle().SetOnStarted(func() {
		startOnce.Do(func() {
			// Keep the UI loop alive even when transient windows (settings/selection) are closed.
			//
			// Without a persistent window, closing the last window can cause a.Run() to return,
			// which shuts down the driver (GLFW). Hotkeys keep running and the *next* UI open
			// will crash with "GLFW library is not initialized".
			fyne.DoAndWait(func() {
				keepAlive = a.NewWindow("go-snip")
				keepAlive.SetPadded(false)
				keepAlive.Resize(fyne.NewSize(1, 1))
				keepAlive.SetContent(widget.NewLabel(""))
				keepAlive.SetCloseIntercept(func() { keepAlive.Hide() })

				// Show once to ensure it's registered with the driver, then hide it.
				keepAlive.Show()
				keepAlive.Hide()
			})

			go func() {
				// Hotkeys run in the background; UI runs on the main thread via a.Run().
				errCh <- runHotkeys(ctx, outDir, cfgPath, cfg, now, out)
				// When the hotkey loop stops (Ctrl+C), exit the UI loop too.
				a.Quit()
			}()
		})
	})

	a.Run()
	return <-errCh
}
