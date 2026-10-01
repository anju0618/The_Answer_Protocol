package main

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestEndingSlotsFollowStoryOrderWithFinalLast(t *testing.T) {
	catalog, err := loadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	slots := catalog.endingSlots()
	if len(slots) != 4 || slots[0].id != "ending.argo" || slots[3].id != "ending.final" {
		t.Fatalf("ending slots = %+v", slots)
	}
	if catalog.lethalRoomCount() == 0 {
		t.Fatal("the Odyssey has game-over rooms")
	}
}

func TestEndingGalleryShowsTrophiesAndRemembersFatalRooms(t *testing.T) {
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
	ui.nameEntry.SetText("gallery-tester")

	ui.showInventory([]string{"item.olive_branch_of_ithaca"})
	// header + 4 slots + fatal counter
	if len(ui.endingBox.Objects) != 6 {
		t.Fatalf("gallery objects = %d, want 6", len(ui.endingBox.Objects))
	}

	room := "loc.ody_lotus_garden"
	ui.recordFatalRoom(room)
	ui.recordFatalRoom(room) // the same room twice counts once
	if got := ui.fatalRooms(); len(got) != 1 || got[0] != room {
		t.Fatalf("fatal rooms = %v", got)
	}
}
