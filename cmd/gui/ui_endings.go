package main

import (
	"fmt"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// endingOrder is the order the gallery lists endings in (the final one last).
var endingOrder = []string{"ending.argo", "ending.troy", "ending.odyssey", "ending.final"}

type endingSlot struct {
	id, rewardItem string
	name           localizedName
}

// endingSlots lists every ending of the world, ordered by endingOrder.
func (catalog *worldCatalog) endingSlots() []endingSlot {
	var slots []endingSlot
	if catalog == nil {
		return slots
	}
	for _, npc := range catalog.NPCs {
		if npc.Ending != nil {
			slots = append(slots, endingSlot{id: npc.Ending.ID, rewardItem: npc.Ending.RewardItem, name: npc.Ending.Name})
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

// lethalRoomCount is how many game-over rooms the world has.
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

// fatalRooms are the game-over rooms this player has already found (kept between sessions).
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

// showEndings redraws the gallery: endings reached (we own their trophy) and game-over rooms found.
func (ui *gui) showEndings() {
	if ui.endingBox == nil {
		return
	}
	var rows []fyne.CanvasObject
	reached := 0
	slots := ui.catalog.endingSlots()
	for _, slot := range slots {
		text := "？？？"
		if slices.Contains(ui.inventory, slot.rewardItem) {
			text = "★ " + slot.name.get(ui.locale)
			reached++
		}
		rows = append(rows, journalCard(journalName(text)))
	}
	if len(slots) == 0 {
		ui.endingBox.Objects = nil
		ui.endingBox.Refresh()
		return
	}
	header := widget.NewLabelWithStyle(fmt.Sprintf(ui.tr("Endings %d/%d", "エンディング %d/%d"), reached, len(slots)), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	fatal := widget.NewLabel(fmt.Sprintf(ui.tr("Fatal choices found: %d/%d", "見つけた即死の選択: %d/%d"), len(ui.fatalRooms()), ui.catalog.lethalRoomCount()))
	ui.endingBox.Objects = append(append([]fyne.CanvasObject{header}, rows...), fatal)
	ui.endingBox.Refresh()
}

func newEndingBox() *fyne.Container { return container.NewVBox() }
