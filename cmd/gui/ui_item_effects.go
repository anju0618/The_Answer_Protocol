package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type itemEffect struct {
	Effect string `json:"effect"`
	Value  int    `json:"value"`
}

func (ui *gui) effectLine(e itemEffect) string {
	switch e.Effect {
	case "damage_bonus":
		return fmt.Sprintf(ui.tr("Damage dealt %+d", "与ダメージ %+d"), e.Value)
	case "counter_reduction":
		return fmt.Sprintf(ui.tr("Damage taken %+d%%", "被ダメージ %+d%%"), -e.Value)
	case "regen_bonus":
		return fmt.Sprintf(ui.tr("HP recovery %+d per 2 sec", "HP回復量 %+d／2秒"), e.Value)
	case "max_hp":
		return fmt.Sprintf(ui.tr("Max HP %+d", "最大HP %+d"), e.Value)
	}
	return ""
}

func (ui *gui) itemEffectLines(id string) []fyne.CanvasObject {
	if ui.catalog == nil {
		return nil
	}
	var lines []fyne.CanvasObject
	for _, e := range ui.catalog.Items[id].Effects {
		text := ui.effectLine(e)
		if text == "" {
			continue
		}
		colour := storyColorFor(colorStoryTeam)
		if e.Value < 0 {
			colour = storyColorFor(colorStoryDeath)
		}
		line := canvas.NewText(text, colour)
		line.TextSize = 12
		lines = append(lines, line)
	}
	return lines
}

func storyColorFor(name fyne.ThemeColorName) color.Color {
	c, _ := storyColor(name)
	return c
}
