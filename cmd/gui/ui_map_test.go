package main

import (
	"fmt"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

func TestMapLayoutPlacesEveryOdysseyRoomOnItsOwnCell(t *testing.T) {
	catalog, err := loadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	positions := computeMapLayout(catalog, "loc.ody")
	seen := map[gridPos]string{}
	for id, pos := range positions {
		if other, dup := seen[pos]; dup {
			t.Errorf("%s and %s share cell %v", id, other, pos)
		}
		seen[pos] = id
	}
	for id := range catalog.Rooms {
		if arcOf(id) == "loc.ody" {
			if _, ok := positions[id]; !ok {
				t.Errorf("Odyssey room %s was not placed on the map", id)
			}
		}
	}
	if positions[hubRoomID] != (gridPos{0, 0}) || positions["loc.ody_troy_shore"] != (gridPos{1, 0}) {
		t.Fatalf("hub/east placement wrong: %v %v", positions[hubRoomID], positions["loc.ody_troy_shore"])
	}
}

func TestWideSceneShowsMapInLeftMargin(t *testing.T) {
	for _, locale := range []string{"en", "ja"} {
		for _, textSize := range []float32{15, 22} {
			t.Run(fmt.Sprintf("%s-text-%g", locale, textSize), func(t *testing.T) { checkWideSceneMap(t, locale, textSize) })
		}
	}
}

func checkWideSceneMap(t *testing.T, locale string, textSize float32) {
	t.Helper()
	application := test.NewApp()
	application.Settings().SetTheme(layoutTestTheme{Theme: retroTheme{base: theme.DarkTheme()}, textSize: textSize})
	defer application.Quit()
	window := application.NewWindow("left map layout")
	defer window.Close()
	catalog, err := loadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	ui := &gui{window: window, locale: locale, catalog: catalog}
	ui.build()
	mapTab := ui.journal.Items[3]
	ui.chatEntry.SetText("unsent message")
	ui.showRoom(lookView{Room: roomView{ID: hubRoomID, Name: catalog.label("room", hubRoomID, locale)}, NPCs: []string{"npc.moirai"}})
	window.Show()
	for step, size := range []fyne.Size{fyne.NewSize(1280, 900), fyne.NewSize(1920, 1080), fyne.NewSize(640, 900), fyne.NewSize(1920, 1080), fyne.NewSize(1280, 900), fyne.NewSize(1920, 1080)} {
		window.SetFullScreen(size.Width == 1920)
		window.Resize(size)
		capture := window.Canvas().Capture()
		if size.Width == 1920 {
			if !ui.sceneMapPanel.Visible() {
				t.Fatal("wide scene did not show the map")
			}
			if len(ui.journal.Items) != 3 {
				t.Fatal("wide scene still has a duplicate map tab")
			}
			if step == 1 && ui.journal.SelectedIndex() != 0 {
				t.Fatal("removing the selected map tab should return to Around")
			}
			checkJournalBounds(t, ui.sceneMapPanel)
			renderedWidth := ui.scene.Size().Height * float32(artWidth) / float32(artHeight)
			leftSpace := (ui.scene.Size().Width - renderedWidth) / 2
			if ui.sceneMapPanel.Position().X+ui.sceneMapPanel.Size().Width > ui.scene.Position().X+leftSpace {
				t.Fatal("map covers the room artwork")
			}
		} else {
			if ui.sceneMapPanel.Visible() {
				t.Fatal("map should hide when the artwork has no large left margin")
			}
			if len(ui.journal.Items) != 4 || ui.journal.Items[3] != mapTab {
				t.Fatal("narrow scene should restore the same map tab once")
			}
		}
		if step >= 2 {
			wantTab := 1
			if step >= 4 {
				wantTab = 2
			}
			if ui.journal.SelectedIndex() != wantTab {
				t.Fatal("resizing should keep the inventory or quests tab selected")
			}
		}
		if ui.chatEntry.Text != "unsent message" {
			t.Fatal("resizing should preserve the chat draft")
		}
		saveOverlapPreview(t, fmt.Sprintf("%s-map-text-%g-%gx%g-step%d", locale, textSize, size.Width, size.Height, step), capture)
		if step == 0 {
			ui.journal.Select(mapTab)
		} else if step == 1 {
			ui.journal.SelectIndex(1)
		} else if step == 3 {
			ui.journal.SelectIndex(2)
		}
	}
	otherLocale := "ja"
	if locale == "ja" {
		otherLocale = "en"
	}
	ui.switchLocale(otherLocale)
	window.Canvas().Capture()
	if len(ui.journal.Items) != 3 || ui.journal.SelectedIndex() != 2 || ui.chatEntry.Text != "unsent message" {
		t.Fatal("switching language in a wide scene should keep tabs and the chat draft without restoring the map tab")
	}
	if ui.sceneMapBox == ui.mapBox || len(ui.sceneMapBox.Objects) != len(ui.mapBox.Objects) {
		t.Fatal("left map must render independently from the map tab")
	}
	for index, object := range ui.sceneMapBox.Objects {
		if object == ui.mapBox.Objects[index] {
			t.Fatal("the two maps share mutable drawing objects")
		}
	}
	before := len(ui.sceneMapBox.Objects)
	ui.markVisited("loc.ody_troy_shore")
	if len(ui.sceneMapBox.Objects) <= before || len(ui.sceneMapBox.Objects) != len(ui.mapBox.Objects) {
		t.Fatal("left map did not follow the visited rooms")
	}
}

func TestMapShowsVisitedRoomsAndTheirNeighbours(t *testing.T) {
	catalog, err := loadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	application := test.NewApp()
	defer application.Quit()
	window := application.NewWindow("test")
	defer window.Close()
	ui := &gui{window: window, locale: "en", catalog: catalog}
	ui.build()

	ui.markVisited(hubRoomID)
	hubOnly := len(ui.mapBox.Objects)
	if hubOnly != 3 {
		t.Fatalf("objects with only the hub visited = %d, want 3 (2 cells + 1 line)", hubOnly)
	}
	ui.markVisited("loc.ody_troy_shore")
	if len(ui.mapBox.Objects) <= hubOnly {
		t.Fatal("visiting another room must add cells to the map")
	}
	ui.disconnect()
	if len(ui.visited) != 0 {
		t.Fatal("disconnecting must forget visited rooms")
	}
}
