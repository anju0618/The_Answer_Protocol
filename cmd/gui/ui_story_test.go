package main

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestStoryEventsAreColouredByKind(t *testing.T) {
	application := test.NewApp()
	defer application.Quit()
	window := application.NewWindow("test")
	defer window.Close()
	ui := &gui{window: window, locale: "en"}
	ui.build()

	ui.handleEvent("EVT PLAYER DEATH You were crushed.")
	ui.handleEvent("EVT PLAYER QUEST A request awaits.")
	ui.handleEvent("EVT ROOM COMBAT The giant roars.")

	want := []storyKind{storyPlain, storyDeath, storyQuest, storyCombat}
	if len(ui.storyLines) != len(want) {
		t.Fatalf("story lines = %d, want %d", len(ui.storyLines), len(want))
	}
	for i, kind := range want {
		if ui.storyLines[i].kind != kind {
			t.Errorf("line %d kind = %v, want %v", i, ui.storyLines[i].kind, kind)
		}
	}
	if len(ui.storyText.Segments) != len(want) {
		t.Fatalf("rich text segments = %d, want %d", len(ui.storyText.Segments), len(want))
	}
}

func TestStoryColourNamesResolveInTheme(t *testing.T) {
	th := retroTheme{base: nil}
	for _, kind := range []storyKind{storyDeath, storyQuest, storyCombat, storyEnding, storyTeam, storyError} {
		name := storyColorName(kind)
		if _, ok := storyColor(name); !ok {
			t.Errorf("kind %v: colour name %q has no colour", kind, name)
		}
		if th.Color(name, 0) == nil {
			t.Errorf("theme returned nil for %q", name)
		}
	}
}
