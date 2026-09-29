package main

import (
	"bytes"
	"encoding/json"
	"image"
	_ "image/png"
	"io/fs"
	"os"
	"strings"
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestItemPhotosCoverWorld(t *testing.T) {
	data, err := os.ReadFile("../../data/world.json")
	if err != nil {
		t.Fatal(err)
	}
	var world struct {
		Items map[string]json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(data, &world); err != nil {
		t.Fatal(err)
	}
	paths, err := fs.Glob(itemPhotoAssets, "assets/items/*.png")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != len(world.Items) {
		t.Fatalf("photo count = %d, item count = %d", len(paths), len(world.Items))
	}
	credits, err := os.ReadFile("assets/items/SOURCES.md")
	if err != nil {
		t.Fatal(err)
	}
	for id := range world.Items {
		path := "assets/items/" + id + ".png"
		data, err := itemPhotoAssets.ReadFile(path)
		if err != nil {
			t.Errorf("photo for %s: %v", id, err)
			continue
		}
		if _, format, err := image.DecodeConfig(bytes.NewReader(data)); err != nil || format != "png" {
			t.Errorf("invalid PNG for %s: %v, %q", id, err, format)
		}
		if !strings.Contains(string(credits), "`"+id+"`") {
			t.Errorf("no source credit for %s", id)
		}
	}
}

func TestGUIItemPhotosFollowLook(t *testing.T) {
	application := test.NewApp()
	defer application.Quit()
	window := application.NewWindow("test")
	defer window.Close()
	ui := &gui{window: window, locale: "en"}
	ui.build()
	ui.showRoom(lookView{Room: roomView{ID: "loc.argo_grove"}, Items: []string{"item.golden_fleece", "item.medeas_draught"}})
	if len(ui.itemPhotoBox.Objects) != 2 {
		t.Fatalf("two present items have %d photos", len(ui.itemPhotoBox.Objects))
	}
	if !ui.itemPhotoScroll.Visible() {
		t.Fatal("item illustrations are hidden when items are present")
	}
	ui.showRoom(lookView{Room: roomView{ID: "loc.argo_grove"}, Items: []string{"item.medeas_draught"}})
	if len(ui.itemPhotoBox.Objects) != 1 {
		t.Fatalf("one present item has %d photos", len(ui.itemPhotoBox.Objects))
	}
	ui.showRoom(lookView{Room: roomView{ID: "loc.argo_grove"}})
	if len(ui.itemPhotoBox.Objects) != 0 {
		t.Fatalf("empty room has %d photos", len(ui.itemPhotoBox.Objects))
	}
	if ui.itemPhotoScroll.Visible() {
		t.Fatal("item illustrations remain visible in an empty room")
	}
}
