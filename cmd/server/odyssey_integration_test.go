package main

import (
	"path/filepath"
	"testing"
)

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
	atStartRoom := func() bool {
		server.mu.Lock()
		defer server.mu.Unlock()
		return server.players["alice"].RoomID == world.StartRoomID
	}

	alice.cmd(t, "MOVE east", "OK room=loc.ody_troy_shore")
	server.mu.Lock()
	crew := server.players["alice"].Crew
	server.mu.Unlock()
	if crew != startingCrew {
		t.Fatalf("crew after entering the Odyssey arc = %d, want %d", crew, startingCrew)
	}

	teleport("loc.ody_lotus")
	alice.cmd(t, "TAKE item.lotus_fruit", "OK taken=item.lotus_fruit")
	server.mu.Lock()
	crew = server.players["alice"].Crew
	server.mu.Unlock()
	if crew != startingCrew-2 {
		t.Fatalf("crew after eating the lotus = %d, want %d", crew, startingCrew-2)
	}

	teleport("loc.ody_aeolus")
	alice.cmd(t, "TAKE item.bag_of_winds", "OK taken=item.bag_of_winds")
	alice.cmd(t, "DROP item.bag_of_winds", "OK dropped=item.bag_of_winds")
	server.mu.Lock()
	crew = server.players["alice"].Crew
	server.mu.Unlock()
	if crew != startingCrew-2-3 {
		t.Fatalf("crew after opening the bag of winds early = %d, want %d", crew, startingCrew-2-3)
	}

	teleport("loc.ody_cyclops")
	beast := alice.cmdJSON(t, "ATTACK npc.polyphemus")
	if beast["status"] != "dead" {
		t.Fatalf("attacking Polyphemus unarmed should kill alice, got %v", beast)
	}
	teleport("loc.ody_cyclops")
	server.mu.Lock()
	server.players["alice"].HP = maxPlayerHP
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

	teleport("loc.ody_circe")
	alice.cmd(t, "TALK npc.circe", "OK dead")
	teleport("loc.ody_circe")
	alice.cmd(t, "TAKE item.moly", "OK taken=item.moly")
	alice.cmd(t, "TALK npc.circe", "OK Eat the moly root first, and my cup will only make you stronger, never smaller.")

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

	alice.cmd(t, "MOVE south", "OK room=loc.ody_charybdis")
	if !atStartRoom() {
		t.Fatalf("charybdis should always be lethal")
	}

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
	crew = server.players["alice"].Crew
	server.mu.Unlock()
	if crew != 4 {
		t.Fatalf("crew after scylla = %d, want 4 (10-6)", crew)
	}

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
