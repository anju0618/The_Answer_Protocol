package main

import (
	"fmt"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

func TestJapaneseInitialAroundLayoutWithoutTabSwitch(t *testing.T) {
	for _, textSize := range []float32{15, 22} {
		for _, size := range []fyne.Size{fyne.NewSize(1280, 900), fyne.NewSize(640, 900)} {
			t.Run(fmt.Sprintf("text-%g-%gx%g", textSize, size.Width, size.Height), func(t *testing.T) {
				checkInitialAroundLayout(t, textSize, size)
			})
		}
	}
}

func checkInitialAroundLayout(t *testing.T, textSize float32, size fyne.Size) {
	t.Helper()
	application := test.NewApp()
	application.Settings().SetTheme(layoutTestTheme{Theme: retroTheme{base: theme.DarkTheme()}, textSize: textSize})
	defer application.Quit()
	window := application.NewWindow("initial room layout")
	defer window.Close()
	window.Resize(size)
	catalog, err := loadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	ui := &gui{window: window, locale: "en", catalog: catalog}
	ui.build()
	window.Show()
	window.Canvas().Capture()
	ui.languageSelect.SetSelected(japaneseLanguageOption)
	window.Canvas().Capture()
	ui.nameEntry.SetText("alice")
	room := catalog.Rooms[hubRoomID]
	ui.showRoom(lookView{
		Room:    roomView{ID: hubRoomID, Name: room.Name.get("ja"), Description: room.Description.get("ja"), Exits: room.Exits},
		Players: []string{"alice"}, NPCs: []string{"npc.moirai"},
	})
	saveOverlapPreview(t, fmt.Sprintf("ja-initial-around-text-%g-%gx%g", textSize, size.Width, size.Height), window.Canvas().Capture())
	checkJournalBounds(t, ui.journal.Items[0].Content.(*container.Scroll).Content)
	boxes := []*fyne.Container{ui.exitBox, ui.npcBox, ui.itemBox, ui.playerBox}
	initialSizes := make([]fyne.Size, len(boxes))
	for index, box := range boxes {
		initialSizes[index] = box.Size()
	}
	ui.journal.SelectIndex(3)
	ui.journal.SelectIndex(0)
	window.Canvas().Capture()
	checkJournalBounds(t, ui.journal.Items[0].Content.(*container.Scroll).Content)
	for index, box := range boxes {
		if box.Size() != initialSizes[index] {
			t.Errorf("tab switch changed a row height: initial=%v after=%v", initialSizes[index], box.Size())
		}
	}
}
