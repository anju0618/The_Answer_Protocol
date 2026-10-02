package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/fs"
	"os"
	"sort"
	"testing"
)

func TestArtCoversWorld(t *testing.T) {
	data, err := os.ReadFile("../../data/world.json")
	if err != nil {
		t.Fatal(err)
	}
	var world struct {
		Rooms map[string]json.RawMessage `json:"rooms"`
		NPCs  map[string]json.RawMessage `json:"npcs"`
	}
	if err := json.Unmarshal(data, &world); err != nil {
		t.Fatal(err)
	}
	roomFiles, err := fs.Glob(artAssets, "assets/rooms/*.png")
	if err != nil {
		t.Fatal(err)
	}
	npcFiles, err := fs.Glob(artAssets, "assets/npcs/*.png")
	if err != nil {
		t.Fatal(err)
	}
	if len(roomFiles) != len(world.Rooms)+1 || len(npcFiles) != len(world.NPCs)+1 {
		t.Fatalf("art coverage: %d/%d rooms, %d/%d NPCs (including unknown fallbacks)", len(roomFiles), len(world.Rooms)+1, len(npcFiles), len(world.NPCs)+1)
	}
	roomIDs := make([]string, 0, len(world.Rooms))
	npcIDs := make([]string, 0, len(world.NPCs))
	roomHashes := make(map[[32]byte]string)
	spriteHashes := make(map[[32]byte]string)
	for id := range world.Rooms {
		data, err := artAssets.ReadFile("assets/rooms/" + id + ".png")
		if err != nil {
			t.Errorf("room %s has no art: %v", id, err)
			continue
		}
		config, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil || config.Width != artWidth || config.Height != artHeight {
			t.Errorf("room %s has invalid PNG: %v, %dx%d", id, err, config.Width, config.Height)
			continue
		}
		roomIDs = append(roomIDs, id)
		hash := sha256.Sum256(data)
		if previous := roomHashes[hash]; previous != "" {
			t.Errorf("rooms %s and %s have identical art", previous, id)
		}
		roomHashes[hash] = id
	}
	for id := range world.NPCs {
		data, err := artAssets.ReadFile("assets/npcs/" + id + ".png")
		if err != nil {
			t.Errorf("NPC %s has no portrait: %v", id, err)
			continue
		}
		config, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil || config.Width != spriteWidth || config.Height != spriteHeight {
			t.Errorf("NPC %s has invalid PNG: %v, %dx%d", id, err, config.Width, config.Height)
			continue
		}
		npcIDs = append(npcIDs, id)
		hash := sha256.Sum256(data)
		if previous := spriteHashes[hash]; previous != "" {
			t.Errorf("NPCs %s and %s have identical portraits", previous, id)
		}
		spriteHashes[hash] = id
	}
	if len(roomIDs) == len(world.Rooms) && len(npcIDs) == len(world.NPCs) {
		sort.Strings(roomIDs)
		sort.Strings(npcIDs)
		previewArt(t, roomIDs, npcIDs)
	}
}

func TestSceneChangesWithRoomAndNPCs(t *testing.T) {
	empty := composeScene("loc.argo_grove", nil, nil)
	withMedea := composeScene("loc.argo_grove", []string{"npc.medea"}, nil)
	withDragon := composeScene("loc.argo_grove", []string{"npc.colchis_dragon"}, nil)
	otherRoom := composeScene("loc.argo_bull_field", []string{"npc.medea"}, nil)
	for name, img := range map[string]*image.RGBA{"Medea": withMedea, "dragon": withDragon, "other room": otherRoom} {
		if sha256.Sum256(img.Pix) == sha256.Sum256(empty.Pix) {
			t.Errorf("scene did not change for %s", name)
		}
	}
	if sha256.Sum256(withMedea.Pix) == sha256.Sum256(withDragon.Pix) {
		t.Fatal("different NPCs produced the same scene")
	}
	if sha256.Sum256(withMedea.Pix) == sha256.Sum256(otherRoom.Pix) {
		t.Fatal("different rooms produced the same scene")
	}
}

func TestUnknownRoomAndNPCUsePNGAssets(t *testing.T) {
	for _, path := range []string{"assets/rooms/unknown.png", "assets/npcs/unknown.png"} {
		data, err := artAssets.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := png.Decode(bytes.NewReader(data)); err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
	}
	actual := composeScene("loc.other_team", []string{"npc.other_team"}, nil)
	expected := composeScene("unknown", []string{"unknown"}, nil)
	if sha256.Sum256(actual.Pix) != sha256.Sum256(expected.Pix) {
		t.Fatal("unknown IDs did not use the PNG fallbacks")
	}
}

func previewArt(t *testing.T, rooms, npcs []string) {
	if path := os.Getenv("TAP_ROOMS_PREVIEW"); path != "" {
		rows := (len(rooms) + 4) / 5
		contact := image.NewRGBA(image.Rect(0, 0, 5*artWidth, rows*artHeight))
		for i, id := range rooms {
			var occupants []string
			if len(npcs) > 0 {
				occupants = []string{npcs[i%len(npcs)]}
			}
			x, y := i%5*artWidth, i/5*artHeight
			draw.Draw(contact, image.Rect(x, y, x+artWidth, y+artHeight), composeScene(id, occupants, nil), image.Point{}, draw.Src)
		}
		writePreview(t, path, contact)
	}
	if path := os.Getenv("TAP_NPCS_PREVIEW"); path != "" {
		rows := (len(npcs) + 7) / 8
		contact := image.NewRGBA(image.Rect(0, 0, 8*80, rows*120))
		draw.Draw(contact, contact.Bounds(), &image.Uniform{C: color.RGBA{R: 45, G: 61, B: 79, A: 255}}, image.Point{}, draw.Src)
		for i, id := range npcs {
			x, y := i%8*80+8, i/8*120+8
			draw.Draw(contact, image.Rect(x, y, x+spriteWidth, y+spriteHeight), loadArt("npcs", id), image.Point{}, draw.Over)
		}
		writePreview(t, path, contact)
	}
}

func writePreview(t *testing.T, path string, img image.Image) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		t.Fatal(err)
	}
}

func TestDefeatedEnemyIsDrawnDifferently(t *testing.T) {
	alive := composeScene("loc.argo_bebrycia", []string{"npc.amycus"}, nil)
	beaten := composeScene("loc.argo_bebrycia", []string{"npc.amycus"}, []string{"npc.amycus"})
	if bytes.Equal(alive.Pix, beaten.Pix) {
		t.Fatal("a defeated enemy looks the same as a living one")
	}
	sprite := defeatedSprite("npc.amycus")
	if sprite == nil || sprite.Bounds().Dx() <= sprite.Bounds().Dy() {
		t.Fatalf("defeated sprite should be lying down (wider than tall), got %v", sprite)
	}
}
