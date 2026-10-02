package main

import (
	"fmt"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

var endingOrder = []string{"ending.argo", "ending.troy", "ending.odyssey", "ending.final"}

type endingSlot struct {
	id, rewardItem string
	name           localizedName
	blessing       string
}

func (catalog *worldCatalog) endingSlots(locale string) []endingSlot {
	var slots []endingSlot
	if catalog == nil {
		return slots
	}
	for _, npc := range catalog.NPCs {
		if npc.Ending != nil {
			slot := endingSlot{id: npc.Ending.ID, rewardItem: npc.Ending.RewardItem, name: npc.Ending.Name}
			if b := npc.Ending.Blessing; b != nil {
				slot.blessing = b.God.get(locale) + ": " + b.Name.get(locale) + " — " + b.Description.get(locale)
			}
			slots = append(slots, slot)
		}
	}
	rank := func(id string) int {
		if i := slices.Index(endingOrder, id); i >= 0 {
			return i
		}
		return len(endingOrder)
	}
	slices.SortFunc(slots, func(a, b endingSlot) int { return rank(a.id) - rank(b.id) })
	return slots
}

func (catalog *worldCatalog) lethalRoomCount() int {
	count := 0
	if catalog == nil {
		return count
	}
	for id := range catalog.Rooms {
		if _, ok := catalog.gameOverRoom(id); ok {
			count++
		}
	}
	return count
}

func (ui *gui) fatalKey() string { return "fatal." + strings.TrimSpace(ui.nameEntry.Text) }

func (ui *gui) fatalRooms() []string {
	if app := fyne.CurrentApp(); app != nil {
		return app.Preferences().StringList(ui.fatalKey())
	}
	return nil
}

func (ui *gui) recordFatalRoom(roomID string) {
	app := fyne.CurrentApp()
	if app == nil || slices.Contains(ui.fatalRooms(), roomID) {
		return
	}
	app.Preferences().SetStringList(ui.fatalKey(), append(ui.fatalRooms(), roomID))
	ui.showEndings()
}

func (ui *gui) showEndings() {
	if ui.endingBox == nil {
		return
	}
	var rows []fyne.CanvasObject
	reached := 0
	slots := ui.catalog.endingSlots(ui.locale)
	for _, slot := range slots {
		card := textVBox(journalName("？？？"))
		if slices.Contains(ui.inventory, slot.rewardItem) {
			card = textVBox(journalName("★ " + slot.name.get(ui.locale)))
			if slot.blessing != "" {
				blessing := widget.NewLabel(slot.blessing)
				blessing.Wrapping = fyne.TextWrapWord
				card.Add(blessing)
			}
			reached++
		}
		rows = append(rows, journalCard(card))
	}
	if len(slots) == 0 {
		ui.endingBox.Objects = nil
		ui.endingBox.Refresh()
		return
	}
	header := journalName(fmt.Sprintf(ui.tr("Endings %d/%d", "エンディング %d/%d"), reached, len(slots)))
	fatal := widget.NewLabel(fmt.Sprintf(ui.tr("Fatal choices found: %d/%d", "見つけた即死の選択: %d/%d"), len(ui.fatalRooms()), ui.catalog.lethalRoomCount()))
	fatal.Wrapping = fyne.TextWrapWord
	ui.endingBox.Objects = append(append([]fyne.CanvasObject{header}, rows...), fatal)
	ui.endingBox.Refresh()
}

func newEndingBox() *fyne.Container { return textVBox() }
