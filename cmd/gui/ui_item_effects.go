package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

// itemEffect is one effect of a carried item, as written in data/world.json.
type itemEffect struct {
	Effect string `json:"effect"`
	Value  int    `json:"value"`
}

// effectLine words one effect, e.g. "与ダメージ +3" or "被ダメージ +10%". Positive Value is always the good direction,
// so the counter_reduction number is flipped when it is shown (reducing damage taken by 10% reads "-10%").
func (ui *gui) effectLine(e itemEffect) string {
	switch e.Effect {
	case "damage_bonus":
		return fmt.Sprintf(ui.tr("Damage dealt %+d", "与ダメージ %+d"), e.Value)
	case "counter_reduction":
		return fmt.Sprintf(ui.tr("Damage taken %+d%%", "被ダメージ %+d%%"), -e.Value)
	case "regen_bonus":
		return fmt.Sprintf(ui.tr("Regeneration %+d", "回復速度 %+d"), e.Value)
	case "max_hp":
		return fmt.Sprintf(ui.tr("Max HP %+d", "最大HP %+d"), e.Value)
	}
	return ""
}

// itemEffectLines shows an item's effects while it is carried: green when helpful, red when harmful.
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
