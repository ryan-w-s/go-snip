# Implementation plan: go-snip (screenshot app)

This document is a step-by-step plan to build the screenshot app described in `README.MD`, following the target architecture:

- `cmd/main.go`: entry point; hotkeys; orchestration
- `internal/capture/capture.go`: screenshot + crop primitives
- `internal/overlay/selection.go`: fullscreen Fyne overlay; mouse drag; visual selection rect
- `internal/utils/file_save.go`: output directory + image encoding + naming

## Goals (what “done” means)

- **Full-screen capture**: Hotkey captures the full primary screen and saves a PNG.
- **Area capture**: Hotkey opens a fullscreen overlay, lets the user drag to select an area, captures that region, and saves a PNG.
- **Configurable output directory**: Defaults to `./screenshots/`, can be overridden via setting in ui.
- **Minimal tests**: At least one unit test per file, focusing on deterministic logic (naming, path creation, crop math).
- **Developer workflow**: `go fmt`, `go vet`, `go test`, `go build` all pass.

## Non-goals (for the first iteration)

- Global tray/menu UI, multi-monitor selection UX polish, image annotation tools, history browser, auto-upload/sharing, or OCR.

## Milestones

### Milestone 0 — Project skeleton & conventions

1. **Create directories and stubs**
   - Add `cmd/`, `internal/capture/`, `internal/overlay/`, `internal/utils/`, and `screenshots/` (created at runtime).
   - Add stub files matching the README architecture.

2. **Decide public APIs between packages**
   - `capture`: functions return `image.Image` (or `*image.RGBA`) plus error.
   - `overlay`: returns a selection rectangle in screen coordinates or image coordinates, plus “cancelled” state.
   - `utils`: provides `EnsureDir`, `DefaultOutputDir`, `GenerateFilename`, and `SavePNG`.

3. **Basic logging & error strategy**
   - Prefer returning errors up the stack; only `main` prints/logs and decides “what to do next”.

Acceptance criteria:
- `go run cmd/main.go` builds and starts (even if it only logs “not implemented”).

---

### Milestone 1 — Saving images to disk (`internal/utils/file_save.go`)

1. **Implement directory handling**
   - `EnsureDir(path string) error`: creates output directory if missing.
   - Default output directory: `./screenshots` (relative to working directory).

2. **Implement deterministic naming**
   - File naming scheme: `YYYYMMDD_HHMMSS_mmm.png` (or include a counter on collisions).
   - Keep naming logic in a pure function for easy testing.

3. **Implement PNG writing**
   - `SavePNG(img image.Image, destPath string) error`
   - Use `os.Create`, `png.Encode`, and close errors correctly.

4. **Write tests**
   - Test filename format and collision logic.
   - Test `EnsureDir` creates directories.

Acceptance criteria:
- A small generated image can be saved to the output dir with a predictable name.

---

### Milestone 2 — Screen capture primitives (`internal/capture/capture.go`)

1. **Define capture API**
   - `CaptureDisplay(displayIndex int) (image.Image, error)` for full display capture.
   - `Crop(img image.Image, r image.Rectangle) (image.Image, error)` for cropping (bounds-checked).

2. **Implement full-screen capture using `github.com/kbinani/screenshot`**
   - Use `screenshot.NumActiveDisplays()` and `screenshot.GetDisplayBounds(i)`.
   - Use `screenshot.CaptureRect(bounds)` for full display.

3. **Implement crop with bounds checks**
   - Ensure selection rectangle is normalized (min/max).
   - Clamp to `img.Bounds()`; return error if empty area after clamp.

4. **Write tests**
   - Crop math tests (pure, deterministic):
     - Normalization of rectangles
     - Clamping behavior
     - Detect empty rectangles

Acceptance criteria:
- A display image can be captured and cropped correctly (crop tests cover edge cases).

---

### Milestone 3 — Selection overlay (`internal/overlay/selection.go`)

This is the interactive piece. Keep it minimal for the first version, then iterate.

1. **Define selection API**
   - `SelectArea() (rect image.Rectangle, cancelled bool, err error)`
   - The returned rectangle should be in **screen coordinates** matching the screenshot bounds used in `capture`.

2. **Implement a fullscreen Fyne window**
   - Create a transparent/overlay-like window that:
     - Covers the screen capture target (initially primary display).
     - Renders a dimmed background and a highlighted selection rectangle.

3. **Mouse interactions**
   - On mouse down: record start point.
   - On drag: update current point; repaint selection rectangle.
   - On mouse up: finalize rectangle and close window.

4. **Keyboard interactions**
   - Escape cancels selection and closes window.
   - Enter (optional) finalizes selection if you want to support click-start then enter.

5. **Coordinate management**
   - Convert from Fyne canvas coordinates to screen coordinates.
   - Ensure returned rect aligns with the bounds used by `screenshot.CaptureRect`.

6. **Testing approach**
   - UI is hard to unit-test; keep most logic testable:
     - Put rectangle normalization/conversion functions in helpers with unit tests.
     - For the UI itself, rely on manual smoke testing.

Acceptance criteria:
- Running area capture shows an overlay; dragging returns a non-empty rectangle; ESC cancels.

---

### Milestone 4 — Hotkeys and orchestration (`cmd/main.go`)

1. **Define hotkeys**
   - Example defaults (adjustable later):
     - Full screen: `Ctrl+Shift+1`
     - Area selection: `Ctrl+Shift+2`
   - Use `golang.design/x/hotkey`.

2. **Implement main loop**
   - Start hotkey listeners.
   - When “full screen” triggers:
     - capture full display
     - save to output dir
     - print path to stdout (and/or show a lightweight notification later)
   - When “area” triggers:
     - open overlay selection window
     - if cancelled: do nothing
     - else: capture full display, crop to selection, save

3. **Config handling**
   - Add a `-out` flag (and optionally `GO_SNIP_OUT`) to override output directory.
   - Validate and `EnsureDir` on startup.

4. **Shutdown behavior**
   - Handle Ctrl+C (SIGINT) for clean unregistration of hotkeys.
   - Ensure no goroutine leaks (hotkey channel listeners stopped).

Tests:
- Keep `main` thin; test business logic via package functions rather than trying to test `main`.

Acceptance criteria:
- Hotkeys work; PNG files appear in the configured directory.

---

### Milestone 5 — Polish and cross-cutting improvements

1. **Multi-monitor support (incremental)**
   - Decide: capture primary display only (v1) vs. detect which display the cursor is on (v2).
   - If adding full support: overlay should open on the same display being captured.

2. **User feedback**
   - Print saved path.
   - Optional: small on-screen “Saved!” Fyne toast, or OS notification (later).

3. **Robustness**
   - Handle failures gracefully (permission issues, invalid display bounds, cancelled selection).
   - Add basic retry on filename collision.

4. **Developer commands in README**
   - Keep README instructions accurate for Windows usage.

Acceptance criteria:
- Smooth manual UX, fewer edge-case crashes, predictable filenames/paths.

## Suggested implementation order (smallest-to-largest risk)

1. `internal/utils/file_save.go` (pure, testable)
2. `internal/capture/capture.go` crop helpers + capture wrapper
3. `cmd/main.go` full-screen hotkey → capture → save
4. `internal/overlay/selection.go` area selection
5. Integrate area selection into `main`
6. Multi-monitor + UX polish

## Manual test checklist

- **Full screen capture**: triggers via hotkey; produces PNG; opens correctly.
- **Area selection**: drag small/large/near-edge; cancel via ESC; zero-area selection handled.
- **Output directory**: default created automatically; `-out` works; errors are readable.


