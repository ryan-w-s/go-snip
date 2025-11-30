package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/kbinani/screenshot"
	"golang.design/x/hotkey"

	"go-snip/internal/capture"
	"go-snip/internal/overlay"
	"go-snip/internal/utils"
)

const outputDirEnv = "GO_SNIP_OUT"

func main() {
	var outFlag string
	flag.StringVar(&outFlag, "out", "", "Output directory for screenshots (overrides GO_SNIP_OUT)")
	flag.Parse()

	outDir := resolveOutputDir(outFlag, os.LookupEnv)
	if err := utils.EnsureDir(outDir); err != nil {
		log.Fatalf("failed to create output dir %q: %v", outDir, err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, outDir, time.Now, os.Stdout); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("go-snip: %v", err)
	}
}

func resolveOutputDir(flagOut string, lookupEnv func(string) (string, bool)) string {
	if strings.TrimSpace(flagOut) != "" {
		return flagOut
	}
	if v, ok := lookupEnv(outputDirEnv); ok && strings.TrimSpace(v) != "" {
		return v
	}
	return utils.DefaultOutputDir()
}

func run(ctx context.Context, outDir string, now func() time.Time, out io.Writer) error {
	if out == nil {
		out = io.Discard
	}

	fullHK := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.Key1)
	areaHK := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.Key2)

	if err := fullHK.Register(); err != nil {
		return fmt.Errorf("register fullscreen hotkey: %w", err)
	}
	defer fullHK.Unregister()

	if err := areaHK.Register(); err != nil {
		return fmt.Errorf("register area hotkey: %w", err)
	}
	defer areaHK.Unregister()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-fullHK.Keydown():
			path, err := handleFull(outDir, now)
			if err != nil {
				log.Printf("fullscreen capture failed: %v", err)
				continue
			}
			fmt.Fprintln(out, path)
		case <-areaHK.Keydown():
			path, cancelled, err := handleArea(outDir, now)
			if cancelled {
				continue
			}
			if err != nil {
				if errors.Is(err, overlay.ErrSelectionUnavailable) {
					log.Printf("area selection unavailable (build with -tags=fyne): %v", err)
				} else {
					log.Printf("area capture failed: %v", err)
				}
				continue
			}
			fmt.Fprintln(out, path)
		}
	}
}

func handleFull(outDir string, now func() time.Time) (savedPath string, err error) {
	img, err := capture.CaptureDisplay(0)
	if err != nil {
		return "", err
	}

	dest := utils.UniquePath(outDir, now(), func(p string) bool {
		_, statErr := os.Stat(p)
		return statErr == nil
	})

	if err := utils.SavePNG(img, dest); err != nil {
		return "", err
	}
	return dest, nil
}

func handleArea(outDir string, now func() time.Time) (savedPath string, cancelled bool, err error) {
	rect, cancelled, err := overlay.SelectArea()
	if err != nil {
		return "", false, err
	}
	if cancelled {
		return "", true, nil
	}

	img, err := capture.CaptureDisplay(0)
	if err != nil {
		return "", false, err
	}

	displayBounds := screenshot.GetDisplayBounds(0)
	cropRect := cropRectFor(img.Bounds(), displayBounds, rect)
	cropped, err := capture.Crop(img, cropRect)
	if err != nil {
		return "", false, err
	}

	dest := utils.UniquePath(outDir, now(), func(p string) bool {
		_, statErr := os.Stat(p)
		return statErr == nil
	})
	if err := utils.SavePNG(cropped, dest); err != nil {
		return "", false, err
	}
	return dest, false, nil
}

// cropRectFor maps a screen-space selection rectangle (relative to displayBounds) into the
// coordinate space of the captured image bounds.
func cropRectFor(imgBounds, displayBounds, selectionRect image.Rectangle) image.Rectangle {
	// Screen coordinate (s) -> image coordinate (i):
	// i = s - displayBounds.Min + imgBounds.Min
	min := selectionRect.Min.Sub(displayBounds.Min).Add(imgBounds.Min)
	max := selectionRect.Max.Sub(displayBounds.Min).Add(imgBounds.Min)
	return image.Rectangle{Min: min, Max: max}
}
