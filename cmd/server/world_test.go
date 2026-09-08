package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func validTestWorld() *World {
	return &World{
		StartRoomID: "loc.start",
		Rooms: map[string]*Room{
			"loc.start": {ID: "loc.start", Exits: map[string]string{"east": "loc.next"}},
			"loc.next":  {ID: "loc.next"},
			"loc.extra": {ID: "loc.extra"},
		},
		Items: map[string]*Item{
			"item.key": {Name: en("Key"), RoomID: "loc.start", Obtainable: true},
		},
		NPCs: map[string]*NPC{
			"npc.guide": {Name: en("Guide"), RoomID: "loc.start", Role: "quest_giver"},
			"npc.enemy": {Name: en("Enemy"), RoomID: "loc.next", Role: "enemy"},
		},
		Quests: map[string]*Quest{
			"quest.key": {
				Name:       en("Find the Key"),
				GiverNPCID: "npc.guide",
				Objective:  QuestObjective{Type: "collect_item", TargetID: "item.key", Count: 1},
			},
		},
	}
}

func TestLoadWorldData(t *testing.T) {
	if _, err := loadWorld(filepath.Join("..", "..", "data", "world.json")); err != nil {
		t.Fatalf("load world data: %v", err)
	}
}

func TestWorldValidateRejectsInvalidReferences(t *testing.T) {
	var nilWorld *World
	if err := nilWorld.validate(); err == nil {
		t.Fatal("nil world passed validation")
	}
	if err := validTestWorld().validate(); err != nil {
		t.Fatalf("valid world failed validation: %v", err)
	}

	tests := []struct {
		name   string
		change func(*World)
		want   string
	}{
		{"missing start room", func(w *World) { w.StartRoomID = "loc.missing" }, "start room"},
		{"null start room", func(w *World) { w.Rooms["loc.start"] = nil }, "start room"},
		{"null room", func(w *World) { w.Rooms["loc.extra"] = nil }, "loc.extra"},
		{"room ID mismatch", func(w *World) { w.Rooms["loc.extra"].ID = "loc.other" }, "has ID"},
		{"missing exit destination", func(w *World) { w.Rooms["loc.start"].Exits["east"] = "loc.missing" }, "exit"},
		{"null exit destination", func(w *World) { w.Rooms["loc.start"].Exits["east"] = "loc.extra"; w.Rooms["loc.extra"] = nil }, "loc.extra"},
		{"null item", func(w *World) { w.Items["item.key"] = nil }, "item \"item.key\" is null"},
		{"missing item room", func(w *World) { w.Items["item.key"].RoomID = "loc.missing" }, "item \"item.key\" points to unknown room"},
		{"null NPC", func(w *World) { w.NPCs["npc.guide"] = nil }, "NPC \"npc.guide\" is null"},
		{"missing NPC room", func(w *World) { w.NPCs["npc.guide"].RoomID = "loc.missing" }, "NPC \"npc.guide\" points to unknown room"},
		{"null quest", func(w *World) { w.Quests["quest.key"] = nil }, "quest \"quest.key\" is null"},
		{"missing giver", func(w *World) { w.Quests["quest.key"].GiverNPCID = "npc.missing" }, "unknown giver NPC"},
		{"missing item target", func(w *World) { w.Quests["quest.key"].Objective.TargetID = "item.missing" }, "unknown item"},
		{"missing NPC target", func(w *World) {
			w.Quests["quest.key"].Objective = QuestObjective{Type: "defeat_npc", TargetID: "npc.missing", Count: 1}
		}, "unknown NPC"},
		{"unknown objective type", func(w *World) { w.Quests["quest.key"].Objective.Type = "unknown" }, "unknown objective type"},
		{"zero objective count", func(w *World) { w.Quests["quest.key"].Objective.Count = 0 }, "invalid objective count"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := validTestWorld()
			tt.change(world)
			if err := world.validate(); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("validate() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}
