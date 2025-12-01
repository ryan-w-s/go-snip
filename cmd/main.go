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
	"sync/atomic"
	"time"

	"github.com/kbinani/screenshot"
	"golang.design/x/hotkey"

	"go-snip/internal/capture"
	"go-snip/internal/config"
	"go-snip/internal/overlay"
	"go-snip/internal/ui"
	"go-snip/internal/utils"
)

const outputDirEnv = "GO_SNIP_OUT"

func main() {
	var outFlag string
	flag.StringVar(&outFlag, "out", "", "Output directory for screenshots (overrides GO_SNIP_OUT)")
	flag.Parse()

	cfgPath, cfg := loadConfig()
	outDir := resolveOutputDir(outFlag, os.LookupEnv, cfg.OutputDir)
	if err := utils.EnsureDir(outDir); err != nil {
		log.Fatalf("failed to create output dir %q: %v", outDir, err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := runEntry(ctx, outDir, cfgPath, cfg, time.Now, os.Stdout); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("go-snip: %v", err)
	}
}

func resolveOutputDir(flagOut string, lookupEnv func(string) (string, bool), configOut string) string {
	if strings.TrimSpace(flagOut) != "" {
		return flagOut
	}
	if v, ok := lookupEnv(outputDirEnv); ok && strings.TrimSpace(v) != "" {
		return v
	}
	if v := strings.TrimSpace(configOut); v != "" {
		return v
	}
	return utils.DefaultOutputDir()
}

func loadConfig() (path string, cfg config.Config) {
	p, err := config.DefaultPath()
	if err != nil {
		log.Printf("config path unavailable: %v", err)
		return "", config.Config{}
	}
	c, err := config.Load(p)
	if err != nil {
		log.Printf("failed to load config %q: %v", p, err)
		return p, config.Config{}
	}
	return p, c
}

func runHotkeys(ctx context.Context, initialOutDir string, cfgPath string, cfg config.Config, now func() time.Time, out io.Writer) error {
	if out == nil {
		out = io.Discard
	}

	fullHK := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.Key1)
	areaHK := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.Key2)
	settingsHK := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyS)

	if err := fullHK.Register(); err != nil {
		return fmt.Errorf("register fullscreen hotkey: %w", err)
	}
	defer fullHK.Unregister()

	if err := areaHK.Register(); err != nil {
		return fmt.Errorf("register area hotkey: %w", err)
	}
	defer areaHK.Unregister()

	if err := settingsHK.Register(); err != nil {
		return fmt.Errorf("register settings hotkey: %w", err)
	}
	defer settingsHK.Unregister()

	var outDir atomic.Value
	outDir.Store(initialOutDir)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-fullHK.Keydown():
			path, cancelled, err := handleFull(outDir.Load().(string), cfg.PostCapturePrompt, now)
			if cancelled {
				continue
			}
			if err != nil {
				log.Printf("fullscreen capture failed: %v", err)
				continue
			}
			if path != "" {
				fmt.Fprintln(out, path)
			}
		case <-areaHK.Keydown():
			path, cancelled, err := handleArea(outDir.Load().(string), cfg.PostCapturePrompt, now)
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
			if path != "" {
				fmt.Fprintln(out, path)
			}
		case <-settingsHK.Keydown():
			current := outDir.Load().(string)
			newCfg, saved, err := ui.ShowSettings(current, cfg.PostCapturePrompt)
			if err != nil {
				if errors.Is(err, ui.ErrSettingsUnavailable) {
					log.Printf("settings unavailable (build with -tags=fyne): %v", err)
				} else {
					log.Printf("settings failed: %v", err)
				}
				continue
			}
			if !saved {
				continue
			}

			raw := strings.TrimSpace(newCfg.OutputDir)
			effective := raw
			if effective == "" {
				effective = utils.DefaultOutputDir()
			}
			if err := utils.EnsureDir(effective); err != nil {
				log.Printf("failed to create output dir %q: %v", effective, err)
				continue
			}

			outDir.Store(effective)

			// Persist (best-effort).
			cfg = newCfg
			cfg.OutputDir = raw
			if strings.TrimSpace(cfgPath) != "" {
				if err := config.Save(cfgPath, cfg); err != nil {
					log.Printf("failed to save config %q: %v", cfgPath, err)
				}
			}
		}
	}
}

func handleFull(outDir string, postCapturePrompt bool, now func() time.Time) (savedPath string, cancelled bool, err error) {
	img, err := capture.CaptureDisplay(0)
	if err != nil {
		return "", false, err
	}

	t := now()
	if postCapturePrompt {
		name, save, err := ui.PromptSave(img)
		if err != nil {
			// Don't lose the capture just because the prompt UI failed.
			log.Printf("post-capture prompt failed (saving anyway): %v", err)
		} else if !save {
			return "", true, nil
		} else {
			if utils.SanitizeFilenameComponent(name) != "" {
				base := utils.BaseNameForTimeAndName(t, name)
				dest := utils.UniquePathWithBase(outDir, base, func(p string) bool {
					_, statErr := os.Stat(p)
					return statErr == nil
				})
				if err := utils.SavePNG(img, dest); err != nil {
					return "", false, err
				}
				return dest, false, nil
			}
			// Empty (or fully-sanitized-to-empty) name: keep the existing timestamp-only scheme.
			dest := utils.UniquePath(outDir, t, func(p string) bool {
				_, statErr := os.Stat(p)
				return statErr == nil
			})
			if err := utils.SavePNG(img, dest); err != nil {
				return "", false, err
			}
			return dest, false, nil
		}
	}

	dest := utils.UniquePath(outDir, t, func(p string) bool {
		_, statErr := os.Stat(p)
		return statErr == nil
	})
	if err := utils.SavePNG(img, dest); err != nil {
		return "", false, err
	}
	return dest, false, nil
}

func handleArea(outDir string, postCapturePrompt bool, now func() time.Time) (savedPath string, cancelled bool, err error) {
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

	t := now()
	if postCapturePrompt {
		name, save, err := ui.PromptSave(cropped)
		if err != nil {
			log.Printf("post-capture prompt failed (saving anyway): %v", err)
		} else if !save {
			return "", true, nil
		} else {
			if utils.SanitizeFilenameComponent(name) != "" {
				base := utils.BaseNameForTimeAndName(t, name)
				dest := utils.UniquePathWithBase(outDir, base, func(p string) bool {
					_, statErr := os.Stat(p)
					return statErr == nil
				})
				if err := utils.SavePNG(cropped, dest); err != nil {
					return "", false, err
				}
				return dest, false, nil
			}
			dest := utils.UniquePath(outDir, t, func(p string) bool {
				_, statErr := os.Stat(p)
				return statErr == nil
			})
			if err := utils.SavePNG(cropped, dest); err != nil {
				return "", false, err
			}
			return dest, false, nil
		}
	}

	dest := utils.UniquePath(outDir, t, func(p string) bool {
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
