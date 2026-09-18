package main

import (
	"embed"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

//go:embed assets/items/*.png
var itemPhotoAssets embed.FS

func itemPhotoCard(id string) fyne.CanvasObject {
	data, err := itemPhotoAssets.ReadFile("assets/items/" + id + ".png")
	if err != nil {
		return nil
	}
	photo := canvas.NewImageFromResource(fyne.NewStaticResource(id+".png", data))
	photo.FillMode = canvas.ImageFillContain
	photo.ScaleMode = canvas.ImageScaleSmooth
	frame := canvas.NewRectangle(color.NRGBA{R: 12, G: 27, B: 56, A: 255})
	frame.StrokeColor = gold
	frame.StrokeWidth = 2
	return container.NewGridWrap(fyne.NewSize(104, 104), container.NewStack(
		frame, container.NewPadded(photo),
	))
}
