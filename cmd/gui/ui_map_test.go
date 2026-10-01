package main

import (
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
	application := test.NewApp()
	application.Settings().SetTheme(retroTheme{base: theme.DarkTheme()})
	defer application.Quit()
	window := application.NewWindow("left map layout")
	defer window.Close()
	catalog, err := loadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	ui := &gui{window: window, locale: "ja", catalog: catalog}
	ui.build()
	ui.showRoom(lookView{Room: roomView{ID: hubRoomID, Name: catalog.label("room", hubRoomID, "ja")}, NPCs: []string{"npc.moirai"}})
	window.Show()
	for _, size := range []fyne.Size{fyne.NewSize(1280, 900), fyne.NewSize(1920, 1080), fyne.NewSize(640, 900), fyne.NewSize(1920, 1080)} {
		window.SetFullScreen(size.Width == 1920)
		window.Resize(size)
		capture := window.Canvas().Capture()
		if size.Width == 1920 {
			if !ui.sceneMapPanel.Visible() {
				t.Fatal("wide scene did not show the map")
			}
			checkJournalBounds(t, ui.sceneMapPanel)
			renderedWidth := ui.scene.Size().Height * float32(artWidth) / float32(artHeight)
			leftSpace := (ui.scene.Size().Width - renderedWidth) / 2
			if ui.sceneMapPanel.Position().X+ui.sceneMapPanel.Size().Width > ui.scene.Position().X+leftSpace {
				t.Fatal("map covers the room artwork")
			}
			saveOverlapPreview(t, "ja-wide-left-map", capture)
		} else if ui.sceneMapPanel.Visible() {
			t.Fatal("map should hide when the artwork has no large left margin")
		}
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
	if hubOnly != 3 { // hub + the unknown room east of it (2 cells), joined by 1 line
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
