package main

import (
	"testing"

	"fyne.io/fyne/v2/test"
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
