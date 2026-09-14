package main

import (
	"path/filepath"
	"testing"
)

func TestNoEnemyIsBothUnwinnableAndUnfleeable(t *testing.T) {
	world, err := loadWorld(filepath.Join("..", "..", "data", "world.json"))
	if err != nil {
		t.Fatalf("load real world: %v", err)
	}
	for id, npc := range world.NPCs {
		if npc.Role != "enemy" {
			continue
		}
		if npc.Unwinnable && !npc.FleeAccurate && !npc.FleeSucceedsOnce {
			t.Fatalf("NPC %q is unwinnable but can never be fled from either, so its room can never be left (blockingEnemyLocked in hazard.go)", id)
		}
	}
}
