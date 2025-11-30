//go:build !fyne

package main

import (
	"context"
	"io"
	"time"

	"go-snip/internal/config"
)

func runEntry(ctx context.Context, outDir string, cfgPath string, cfg config.Config, now func() time.Time, out io.Writer) error {
	return runHotkeys(ctx, outDir, cfgPath, cfg, now, out)
}
