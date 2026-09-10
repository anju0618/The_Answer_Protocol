package main

import (
	"path/filepath"
	"testing"
)

func TestArgonautsAndTroyMythGatesAgainstRealWorldData(t *testing.T) {
	world, err := loadWorld(filepath.Join("..", "..", "data", "world.json"))
	if err != nil {
		t.Fatalf("load real world: %v", err)
	}

	server := newServer(t.TempDir())
	server.world = world
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	teleport := func(roomID string) {
		server.mu.Lock()
		server.players["alice"].RoomID = roomID
		server.mu.Unlock()
	}
	fullHeal := func() {
		server.mu.Lock()
		server.players["alice"].HP = maxPlayerHP
		server.mu.Unlock()
	}

	teleport("loc.argo_bull_field")
	bulls := alice.cmdJSON(t, "ATTACK npc.khalkotauroi")
	if bulls["status"] != "dead" {
		t.Fatalf("attacking the bronze bulls unprotected should kill alice, got %v", bulls)
	}
	teleport("loc.argo_bull_field")
	fullHeal()
	alice.cmd(t, "TAKE item.medeas_ointment", "ERR 404 ITEM_NOT_FOUND")
	teleport("loc.argo_court_aeetes")
	alice.cmd(t, "TAKE item.medeas_ointment", "OK taken=item.medeas_ointment")
	teleport("loc.argo_bull_field")
	won := false
	for range 30 {
		result := alice.cmdJSON(t, "ATTACK npc.khalkotauroi")
		if result["status"] == "dead" {
			t.Fatalf("attacking the bulls with the ointment should never be instant death, got %v", result)
		}
		if result["status"] == "victory" {
			won = true
			break
		}
	}
	if !won {
		t.Fatalf("did not defeat the bronze bulls within 30 rounds")
	}

	teleport("loc.argo_grove")
	dragon := alice.cmdJSON(t, "ATTACK npc.colchis_dragon")
	if dragon["status"] != "dead" {
		t.Fatalf("attacking the dragon without the draught should kill alice, got %v", dragon)
	}
	teleport("loc.argo_grove")
	fullHeal()
	alice.cmd(t, "TAKE item.medeas_draught", "OK taken=item.medeas_draught")
	won = false
	for range 30 {
		result := alice.cmdJSON(t, "ATTACK npc.colchis_dragon")
		if result["status"] == "dead" {
			t.Fatalf("attacking the dragon with the draught should never be instant death, got %v", result)
		}
		if result["status"] == "victory" {
			won = true
			break
		}
	}
	if !won {
		t.Fatalf("did not defeat the Colchis dragon within 30 rounds")
	}

	teleport("loc.argo_crete")
	talos := alice.cmdJSON(t, "ATTACK npc.talos")
	if talos["status"] != "dead" {
		t.Fatalf("attacking Talos without Medea's aid should kill alice, got %v", talos)
	}
	teleport("loc.argo_crete")
	fullHeal()
	server.mu.Lock()
	server.players["alice"].Quests = map[string]*PlayerQuest{"quest.golden_fleece": {Status: "completed", Progress: 1}}
	server.mu.Unlock()
	won = false
	for range 30 {
		result := alice.cmdJSON(t, "ATTACK npc.talos")
		if result["status"] == "dead" {
			t.Fatalf("attacking Talos with the golden fleece quest completed should never be instant death, got %v", result)
		}
		if result["status"] == "victory" {
			won = true
			break
		}
	}
	if !won {
		t.Fatalf("did not defeat Talos within 30 rounds")
	}

	teleport("loc.troy_gate")
	hector := alice.cmdJSON(t, "ATTACK npc.hector")
	if hector["status"] != "dead" {
		t.Fatalf("attacking Hector unshielded should kill alice, got %v", hector)
	}
	teleport("loc.troy_gate")
	fullHeal()
	alice.cmd(t, "TAKE item.shield_of_achilles", "ERR 404 ITEM_NOT_FOUND")
	teleport("loc.troy_achilles_tent")
	alice.cmd(t, "TAKE item.shield_of_achilles", "OK taken=item.shield_of_achilles")
	teleport("loc.troy_gate")

	alice.cmdJSON(t, "ATTACK npc.hector")
	firstFlee := alice.cmdJSON(t, "FLEE")
	if firstFlee["result"] != "success" {
		t.Fatalf("the first ever flee from Hector should succeed, got %v", firstFlee)
	}
	alice.cmdJSON(t, "ATTACK npc.hector")
	secondFlee := alice.cmdJSON(t, "FLEE")
	if secondFlee["result"] == "success" {
		t.Fatalf("a second flee from Hector should fail, got %v", secondFlee)
	}
}
