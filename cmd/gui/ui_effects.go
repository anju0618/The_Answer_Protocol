package main

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

var (
	flashClear  = color.NRGBA{}
	flashHurt   = color.NRGBA{R: 200, G: 30, B: 30, A: 110}
	flashDeath  = color.NRGBA{R: 120, G: 0, B: 0, A: 200}
	flashTravel = color.NRGBA{A: 230}
)

func newFlashLayer() *canvas.Rectangle {
	return canvas.NewRectangle(flashClear)
}

func (ui *gui) flashScene(from color.NRGBA, duration time.Duration) {
	if ui.flash == nil {
		return
	}
	if ui.flashAnim != nil {
		ui.flashAnim.Stop()
	}
	ui.flashAnim = canvas.NewColorRGBAAnimation(from, flashClear, duration, func(c color.Color) {
		ui.flash.FillColor = c
		ui.flash.Refresh()
	})
	ui.flashAnim.Curve = fyne.AnimationEaseOut
	ui.flashAnim.Start()
}
