package main

import (
	"bytes"
	"embed"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strings"

	xdraw "golang.org/x/image/draw"
)

const (
	artWidth     = 960
	artHeight    = 576
	spriteWidth  = 192
	spriteHeight = 312
)

//go:embed assets/rooms/*.png assets/npcs/*.png
var artAssets embed.FS

func loadArt(kind, id string) image.Image {
	if id == "" || strings.ContainsAny(id, `/\`) {
		id = "unknown"
	}
	data, err := artAssets.ReadFile("assets/" + kind + "/" + id + ".png")
	if err != nil {
		data, err = artAssets.ReadFile("assets/" + kind + "/unknown.png")
	}
	if err != nil {
		return nil
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	return img
}

func composeScene(roomID string, npcs []string) *image.RGBA {
	background := image.NewRGBA(image.Rect(0, 0, artWidth, artHeight))
	if room := loadArt("rooms", roomID); room != nil {
		if room.Bounds().Dx() == artWidth && room.Bounds().Dy() == artHeight {
			draw.Draw(background, background.Bounds(), room, room.Bounds().Min, draw.Src)
		} else {
			xdraw.CatmullRom.Scale(background, background.Bounds(), room, room.Bounds(), draw.Src, nil)
		}
	}
	count := len(npcs)
	if count == 0 {
		return background
	}
	width := spriteWidth
	if count > 4 {
		width = (artWidth - 60) / count
	}
	if width < 42 {
		width = 42
	}
	height := spriteHeight * width / spriteWidth
	for i, id := range npcs {
		x := 645
		if count > 1 {
			x = 36 + i*(artWidth-width-72)/(count-1)
		}
		if count == 2 {
			x = 495 + i*219
		}
		y := artHeight - height - 21
		shadow := &image.Uniform{C: color.NRGBA{R: 6, G: 15, B: 32, A: 125}}
		draw.Draw(background, image.Rect(x+21, artHeight-45, x+width-21, artHeight-30), shadow, image.Point{}, draw.Over)
		sprite := loadArt("npcs", id)
		if sprite == nil {
			continue
		}
		dest := image.Rect(x, y, x+width, y+height)
		if sprite.Bounds().Dx() == width && sprite.Bounds().Dy() == height {
			draw.Draw(background, dest, sprite, sprite.Bounds().Min, draw.Over)
		} else {
			xdraw.CatmullRom.Scale(background, dest, sprite, sprite.Bounds(), draw.Over, nil)
		}
	}
	return background
}
