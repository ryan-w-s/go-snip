//go:build fyne
// +build fyne

package overlay

import (
	"errors"
	"image"
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/kbinani/screenshot"
)

type selectionResult struct {
	rect      image.Rectangle
	cancelled bool
	err       error
}

// SelectArea displays a fullscreen overlay (primary display only) and lets the user
// drag to select an area.
//
// The returned rectangle is in screen coordinates compatible with screenshot.CaptureRect.
// If the user cancels (Esc or closing the window), cancelled is true.
func SelectArea() (rect image.Rectangle, cancelled bool, err error) {
	a := fyne.CurrentApp()
	if a == nil {
		return image.Rectangle{}, false, ErrSelectionUnavailable
	}
	if a.Driver() == nil {
		return image.Rectangle{}, false, errors.New("overlay: fyne driver unavailable (app not running?)")
	}

	n := screenshot.NumActiveDisplays()
	if n <= 0 {
		return image.Rectangle{}, false, ErrNoActiveDisplays
	}

	displayBounds := screenshot.GetDisplayBounds(0) // primary-only for v1
	bgImg, err := screenshot.CaptureRect(displayBounds)
	if err != nil {
		return image.Rectangle{}, false, err
	}

	done := make(chan selectionResult, 1)
	var once sync.Once
	send := func(res selectionResult) {
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
		w := a.NewWindow("go-snip: select area")
		w.SetPadded(false)
		w.SetFullScreen(true)

		finish := func(r image.Rectangle, cancelled bool) {
			send(selectionResult{rect: r, cancelled: cancelled})
			w.Close()
		}

		selector := newSelectionWidget(bgImg, displayBounds, finish)
		w.SetContent(selector)

		// Escape cancels selection.
		w.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
			if ev == nil {
				return
			}
			if ev.Name == fyne.KeyEscape {
				finish(image.Rectangle{}, true)
			}
		})

		w.SetOnClosed(func() {
			send(selectionResult{cancelled: true})
		})

		w.Show()
	})

	res := <-done
	return res.rect, res.cancelled, res.err
}

type selectionWidget struct {
	widget.BaseWidget

	bgImg         image.Image
	displayBounds image.Rectangle

	start    fyne.Position
	current  fyne.Position
	hasStart bool

	finish func(rect image.Rectangle, cancelled bool)
}

func newSelectionWidget(bgImg image.Image, displayBounds image.Rectangle, finish func(rect image.Rectangle, cancelled bool)) *selectionWidget {
	w := &selectionWidget{
		bgImg:         bgImg,
		displayBounds: displayBounds,
		current:       fyne.NewPos(0, 0),
		finish:        finish,
	}
	w.ExtendBaseWidget(w)
	return w
}

// MouseDown starts a selection.
func (w *selectionWidget) MouseDown(ev *desktop.MouseEvent) {
	if ev == nil {
		return
	}
	w.start = ev.Position
	w.current = ev.Position
	w.hasStart = true
	w.Refresh()
}

// MouseUp finalizes a selection if one is active.
func (w *selectionWidget) MouseUp(ev *desktop.MouseEvent) {
	if !w.hasStart {
		return
	}
	if ev != nil {
		w.current = ev.Position
	}
	w.finalize()
}

func (w *selectionWidget) Dragged(ev *fyne.DragEvent) {
	if !w.hasStart || ev == nil {
		return
	}
	w.current = ev.Position
	w.Refresh()
}

func (w *selectionWidget) DragEnd() {
	if !w.hasStart {
		return
	}
	w.finalize()
}

func (w *selectionWidget) finalize() {
	defer func() {
		w.hasStart = false
	}()

	sz := w.Size()
	r := canvasRectToScreenRect(
		CanvasPos{X: w.start.X, Y: w.start.Y},
		CanvasPos{X: w.current.X, Y: w.current.Y},
		CanvasSize{W: sz.Width, H: sz.Height},
		w.displayBounds,
	)
	if r.Dx() <= 0 || r.Dy() <= 0 {
		w.finish(image.Rectangle{}, true)
		return
	}
	w.finish(r, false)
}

func (w *selectionWidget) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewImageFromImage(w.bgImg)
	bg.FillMode = canvas.ImageFillStretch

	dim := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 120})

	sel := canvas.NewRectangle(color.NRGBA{R: 0, G: 120, B: 255, A: 60})
	sel.StrokeColor = color.NRGBA{R: 0, G: 120, B: 255, A: 200}
	sel.StrokeWidth = 2
	sel.Hide()

	return &selectionRenderer{
		w:   w,
		bg:  bg,
		dim: dim,
		sel: sel,
		objects: []fyne.CanvasObject{
			bg,
			dim,
			sel,
		},
	}
}

type selectionRenderer struct {
	w *selectionWidget

	bg  *canvas.Image
	dim *canvas.Rectangle
	sel *canvas.Rectangle

	objects []fyne.CanvasObject
}

func (r *selectionRenderer) Layout(size fyne.Size) {
	r.bg.Move(fyne.NewPos(0, 0))
	r.bg.Resize(size)

	r.dim.Move(fyne.NewPos(0, 0))
	r.dim.Resize(size)

	if !r.w.hasStart {
		r.sel.Hide()
		return
	}

	minX := min32(r.w.start.X, r.w.current.X)
	minY := min32(r.w.start.Y, r.w.current.Y)
	maxX := max32(r.w.start.X, r.w.current.X)
	maxY := max32(r.w.start.Y, r.w.current.Y)

	r.sel.Move(fyne.NewPos(minX, minY))
	r.sel.Resize(fyne.NewSize(maxX-minX, maxY-minY))
	r.sel.Show()
}

func (r *selectionRenderer) MinSize() fyne.Size {
	return fyne.NewSize(10, 10)
}

func (r *selectionRenderer) Refresh() {
	r.Layout(r.w.Size())
	r.bg.Refresh()
	r.dim.Refresh()
	r.sel.Refresh()
}

func (r *selectionRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *selectionRenderer) Destroy() {}

func min32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func max32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
