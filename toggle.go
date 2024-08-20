package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type toggleRenderer struct {
	toggle     *Toggle
	background *canvas.Rectangle
	swtch      *canvas.Circle
	objects    []fyne.CanvasObject
}

func (t *toggleRenderer) MinSize() fyne.Size {
	return fyne.NewSize(45, 21)
}

func (t *toggleRenderer) Layout(size fyne.Size) {
	t.background.Resize(size)
	swtchSize := fyne.Min(size.Height/.5, size.Width/2.5)
	t.swtch.Resize(fyne.NewSize(swtchSize, swtchSize))

	if t.toggle.On {
		t.swtch.Move(fyne.NewPos(size.Width-swtchSize-1, 1))
		t.swtch.FillColor = theme.Color(theme.ColorNamePrimary)
	} else {
		t.swtch.Move(fyne.NewPos(1, 1))
		t.swtch.FillColor = theme.Color(theme.ColorNameBackground)
	}
}

func (t *toggleRenderer) Refresh() {
	t.Layout(t.toggle.Size())
	canvas.Refresh(t.toggle)
}

func (t *toggleRenderer) Objects() []fyne.CanvasObject {
	return t.objects
}

func (t toggleRenderer) Destroy() {}

type Toggle struct {
	widget.BaseWidget
	On        bool
	onChanged func(bool)
}

func (t *Toggle) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	swtch := canvas.NewCircle(theme.Color(theme.ColorNameBackground))

	bg.CornerRadius = 8.0
	bg.StrokeWidth = 2.0
	bg.StrokeColor = theme.Color(theme.ColorNameInputBorder)

	swtch.StrokeWidth = 2.0
	swtch.StrokeColor = theme.Color(theme.ColorNamePressed)

	return &toggleRenderer{
		toggle:     t,
		background: bg,
		swtch:      swtch,
		objects:    []fyne.CanvasObject{bg, swtch},
	}
}

func (t *Toggle) OnChanged(f func(bool)) {
	t.onChanged = f
}

func (t *Toggle) Tapped(_ *fyne.PointEvent) {
	t.On = !t.On
	t.Refresh()
	if t.onChanged != nil {
		t.onChanged(t.On)
	}
}

func NewToggle() *Toggle {
	t := &Toggle{}
	t.ExtendBaseWidget(t)
	return t
}
