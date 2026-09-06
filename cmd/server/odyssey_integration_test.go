package main

import (
	"path/filepath"
	"testing"
)

// TestOdysseyArcAgainstRealWorldData exercises the myth-gate mechanics
// (memo.md 7.3-7.6) against the actual data/world.json content, not a
// synthetic test world — so it catches ID typos or wiring mistakes between
// the data and the engine that combat_test.go's isolated worlds can't see.
func TestOdysseyArcAgainstRealWorldData(t *testing.T) {
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
	// respawnHP is not a reliable "did they die" signal once HP is already
	// sitting at respawnHP from an earlier death, so check where they ended
	// up instead: death always sends them back to the world's start room.
	atStartRoom := func() bool {
		server.mu.Lock()
		defer server.mu.Unlock()
		return server.players["alice"].RoomID == world.StartRoomID
	}

	// Polyphemus: dies without the stake, defeatable with it.
	teleport("loc.ody_cyclops")
	beast := alice.cmdJSON(t, "ATTACK npc.polyphemus")
	if beast["status"] != "dead" {
		t.Fatalf("attacking Polyphemus unarmed should kill alice, got %v", beast)
	}
	teleport("loc.ody_cyclops")
	server.mu.Lock()
	server.players["alice"].HP = maxPlayerHP // the earlier instant death left HP at respawnHP
	server.mu.Unlock()
	alice.cmd(t, "TAKE item.olive_stake", "OK taken=item.olive_stake")
	won := false
	for range 20 {
		result := alice.cmdJSON(t, "ATTACK npc.polyphemus")
		if result["status"] == "dead" {
			t.Fatalf("attacking Polyphemus with the stake should never be instant death, got %v", result)
		}
		if result["status"] == "victory" {
			won = true
			break
		}
	}
	if !won {
		t.Fatalf("did not defeat Polyphemus within 20 rounds")
	}

	// Circe: TALK kills without moly, is safe with it.
	teleport("loc.ody_circe")
	alice.cmd(t, "TALK npc.circe", "OK dead")
	teleport("loc.ody_circe")
	alice.cmd(t, "TAKE item.moly", "OK taken=item.moly")
	alice.cmd(t, "TALK npc.circe", "OK Eat the moly root first, and my cup will only make you stronger, never smaller.")

	// Sirens: entering the room kills without beeswax, is safe with it.
	teleport("loc.ody_circe")
	alice.cmd(t, "MOVE south", "OK room=loc.ody_sirens")
	if !atStartRoom() {
		t.Fatalf("entering the sirens without beeswax should kill alice")
	}
	teleport("loc.ody_circe")
	alice.cmd(t, "TAKE item.beeswax", "OK taken=item.beeswax")
	alice.cmd(t, "MOVE south", "OK room=loc.ody_sirens")
	if atStartRoom() {
		t.Fatalf("entering the sirens with beeswax should not kill alice")
	}

	// Charybdis: always lethal.
	alice.cmd(t, "MOVE south", "OK room=loc.ody_charybdis")
	if !atStartRoom() {
		t.Fatalf("charybdis should always be lethal")
	}

	// Scylla: needs crew+1 >= 7 to survive, loses 6 crew when it does.
	teleport("loc.ody_sirens")
	server.mu.Lock()
	server.players["alice"].Crew = 2
	server.mu.Unlock()
	alice.cmd(t, "MOVE east", "OK room=loc.ody_scylla")
	if !atStartRoom() {
		t.Fatalf("scylla with too little crew should be lethal")
	}
	teleport("loc.ody_sirens")
	server.mu.Lock()
	server.players["alice"].Crew = 10
	server.mu.Unlock()
	alice.cmd(t, "MOVE east", "OK room=loc.ody_scylla")
	server.mu.Lock()
	crew := server.players["alice"].Crew
	server.mu.Unlock()
	if crew != 4 {
		t.Fatalf("crew after scylla = %d, want 4 (10-6)", crew)
	}

	// Laestrygonians: attacking costs crew, not HP.
	teleport("loc.ody_laestrygonians")
	server.mu.Lock()
	server.players["alice"].Crew = 12
	server.mu.Unlock()
	giant := alice.cmdJSON(t, "ATTACK npc.laestrygonian")
	if giant["status"] != "overwhelmed" {
		t.Fatalf("attacking the Laestrygonians should never be won outright, got %v", giant)
	}
	server.mu.Lock()
	crew = server.players["alice"].Crew
	server.mu.Unlock()
	if crew != 4 {
		t.Fatalf("crew after laestrygonian attack = %d, want 4 (12-8)", crew)
	}

	// Suitors: dies without the strung bow, defeatable with it.
	teleport("loc.ody_palace")
	suitor := alice.cmdJSON(t, "ATTACK npc.antinous")
	if suitor["status"] != "dead" {
		t.Fatalf("attacking the suitors bare-handed should kill alice, got %v", suitor)
	}
	teleport("loc.ody_palace")
	alice.cmd(t, "TAKE item.odysseus_bow", "OK taken=item.odysseus_bow")
	suitor = alice.cmdJSON(t, "ATTACK npc.antinous")
	if suitor["status"] == "dead" {
		t.Fatalf("attacking the suitors with the bow should not be instant death")
	}
}
