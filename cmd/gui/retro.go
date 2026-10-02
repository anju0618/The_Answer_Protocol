package main

import (
	_ "embed"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

//go:embed assets/fonts/DroidSansFallbackFull.ttf
var droidFont []byte

var droidResource = fyne.NewStaticResource("DroidSansFallbackFull.ttf", droidFont)

var (
	ink     = color.NRGBA{R: 5, G: 12, B: 30, A: 255}
	navy    = color.NRGBA{R: 12, G: 27, B: 56, A: 255}
	ivory   = color.NRGBA{R: 244, G: 232, B: 191, A: 255}
	gold    = color.NRGBA{R: 204, G: 168, B: 90, A: 255}
	dimGold = color.NRGBA{R: 112, G: 96, B: 62, A: 255}
)

type retroTheme struct{ base fyne.Theme }

func (t retroTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if c, ok := storyColor(name); ok {
		return c
	}
	switch name {
	case theme.ColorNameBackground, theme.ColorNameMenuBackground:
		return ink
	case theme.ColorNameButton, theme.ColorNameInputBackground:
		return navy
	case theme.ColorNameForeground, theme.ColorNameForegroundOnPrimary:
		return ivory
	case theme.ColorNamePrimary, theme.ColorNameFocus, theme.ColorNameSelection:
		return gold
	case theme.ColorNameInputBorder, theme.ColorNameSeparator:
		return dimGold
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 147, G: 153, B: 166, A: 255}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 28, G: 38, B: 60, A: 255}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 150, G: 156, B: 172, A: 255}
	}
	return t.base.Color(name, theme.VariantDark)
}

func (t retroTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Symbol {
		return t.base.Font(style)
	}
	return droidResource
}

func (t retroTheme) Icon(name fyne.ThemeIconName) fyne.Resource { return t.base.Icon(name) }

func (t retroTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 15
	case theme.SizeNameHeadingText:
		return 19
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameButtonRadius, theme.SizeNameInputRadius,
		theme.SizeNameDialogRadius, theme.SizeNamePopupRadius:
		return 0
	}
	return t.base.Size(name)
}

func framed(title string, content fyne.CanvasObject) fyne.CanvasObject {
	frame := canvas.NewRectangle(ink)
	frame.StrokeColor = dimGold
	frame.StrokeWidth = 1
	caption := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	inside := container.NewBorder(caption, nil, nil, nil, content)
	return container.NewStack(frame, container.NewPadded(inside))
}

func sceneImage() *canvas.Image {
	image := canvas.NewImageFromImage(loadArt("rooms", "unknown"))
	image.FillMode = canvas.ImageFillContain
	image.ScaleMode = canvas.ImageScaleSmooth
	return image
}
