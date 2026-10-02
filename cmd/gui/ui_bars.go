package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

type statBar struct {
	box    *fyne.Container
	fill   *canvas.Rectangle
	layout *barLayout
}

type barLayout struct{ ratio float32 }

func (*barLayout) MinSize([]fyne.CanvasObject) fyne.Size { return fyne.NewSize(0, 0) }

func (l *barLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(size)
	objects[1].Move(fyne.NewPos(0, 0))
	objects[1].Resize(fyne.NewSize(size.Width*l.ratio, size.Height))
}

func newStatBar() *statBar {
	back := canvas.NewRectangle(color.NRGBA{R: 28, G: 38, B: 60, A: 255})
	fill := canvas.NewRectangle(barColor(1))
	layout := &barLayout{}
	return &statBar{box: container.New(layout, back, fill), fill: fill, layout: layout}
}

func (b *statBar) Set(current, max int) {
	b.layout.ratio = barRatio(current, max)
	b.fill.FillColor = barColor(b.layout.ratio)
	b.box.Refresh()
}

func barRatio(current, max int) float32 {
	if max <= 0 || current <= 0 {
		return 0
	}
	if current >= max {
		return 1
	}
	return float32(current) / float32(max)
}

func barColor(ratio float32) color.NRGBA {
	switch {
	case ratio > 0.5:
		return color.NRGBA{R: 46, G: 125, B: 80, A: 255}
	case ratio > 0.25:
		return color.NRGBA{R: 176, G: 128, B: 32, A: 255}
	}
	return color.NRGBA{R: 170, G: 48, B: 48, A: 255}
}
