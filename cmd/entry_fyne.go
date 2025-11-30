//go:build fyne
// +build fyne

package main

import (
	"context"
	"io"
	"time"

	"fyne.io/fyne/v2/app"

	"go-snip/internal/config"
)

func runEntry(ctx context.Context, outDir string, cfgPath string, cfg config.Config, now func() time.Time, out io.Writer) error {
	a := app.NewWithID("go-snip")

	errCh := make(chan error, 1)
	go func() {
		// Hotkeys run in the background; UI runs on the main thread via a.Run().
		errCh <- runHotkeys(ctx, outDir, cfgPath, cfg, now, out)
		// When the hotkey loop stops (Ctrl+C), exit the UI loop too.
		a.Quit()
	}()

	a.Run()
	return <-errCh
}
