package main

import (
	"path/filepath"
	"strings"
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

	alice.cmd(t, "MOVE east", "OK room=loc.ody_cicones")
	server.mu.Lock()
	crew = server.players["alice"].Crew
	server.mu.Unlock()
	if crew != startingCrew-2 {
		t.Fatalf("crew after the Cicones raid = %d, want %d", crew, startingCrew-2)
	}

	alice.cmd(t, "MOVE east", "OK room=loc.ody_lotus")
	alice.cmd(t, "TAKE item.lotus_fruit", "OK taken=item.lotus_fruit")
	if !atStartRoom() {
		t.Fatalf("eating the lotus fruit should kill alice")
	}

	teleport("loc.ody_lotus")
	alice.cmd(t, "MOVE east", "OK room=loc.ody_cyclops")
	server.mu.Lock()
	crew = server.players["alice"].Crew
	server.mu.Unlock()
	if crew != startingCrew-2-2 {
		t.Fatalf("crew after Polyphemus's cave = %d, want %d", crew, startingCrew-2-2)
	}

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

	teleport("loc.ody_aeolus")
	alice.cmd(t, "TAKE item.bag_of_winds", "OK taken=item.bag_of_winds")
	alice.cmd(t, "DROP item.bag_of_winds", "OK dropped=item.bag_of_winds")
	server.mu.Lock()
	crew = server.players["alice"].Crew
	server.mu.Unlock()
	if crew != startingCrew-2-2-3 {
		t.Fatalf("crew after opening the bag of winds early = %d, want %d", crew, startingCrew-2-2-3)
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

	alice.cmd(t, "MOVE east", "OK room=loc.ody_thrinacia")
	if !atStartRoom() {
		t.Fatalf("Scylla is still alive and unresolved, leaving toward Thrinacia without fleeing her should kill alice")
	}
	teleport("loc.ody_scylla")
	alice.cmdJSON(t, "ATTACK npc.scylla")
	fleeResult := alice.cmdJSON(t, "FLEE")
	if fleeResult["result"] != "success" {
		t.Fatalf("fleeing Scylla should succeed (she's myth-accurate to flee, not fight), got %v", fleeResult)
	}
	alice.cmd(t, "MOVE east", "OK room=loc.ody_thrinacia")
	if atStartRoom() {
		t.Fatalf("Scylla is resolved (fled from), leaving should now succeed")
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

func TestOdysseyWrongTurnsAreFatal(t *testing.T) {
	world, err := loadWorld(filepath.Join("..", "..", "data", "world.json"))
	if err != nil {
		t.Fatalf("load real world: %v", err)
	}
	server := newServer(t.TempDir())
	server.world = world
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	wrongTurns := []struct{ from, direction, to string }{
		{"loc.ody_cicones", "north", "loc.ody_ismarus_feast"},
		{"loc.ody_lotus", "north", "loc.ody_lotus_garden"},
		{"loc.ody_cyclops", "south", "loc.ody_sealed_cave"},
		{"loc.ody_laestrygonians", "north", "loc.ody_laestrygonian_depths"},
		{"loc.ody_circe", "north", "loc.ody_pigsty"},
		{"loc.ody_calypso", "south", "loc.ody_eternal_ogygia"},
	}
	for _, turn := range wrongTurns {
		if world.Rooms[turn.to].Hazard == nil || world.Rooms[turn.to].Hazard.Type != "lethal" {
			t.Fatalf("%s should be a lethal room", turn.to)
		}
		server.mu.Lock()
		player := server.players["alice"]
		player.RoomID = turn.from
		player.HP = maxPlayerHP
		player.FledFrom = map[string]bool{"npc.polyphemus": true, "npc.laestrygonian": true}
		server.mu.Unlock()

		alice.cmd(t, "MOVE "+turn.direction, "OK room="+turn.to)
		server.mu.Lock()
		room, hp := server.players["alice"].RoomID, server.players["alice"].HP
		server.mu.Unlock()
		if room != world.StartRoomID || hp != respawnHP {
			t.Errorf("%s %s: alice ended in %s with %d HP, want respawn at %s with %d HP",
				turn.from, turn.direction, room, hp, world.StartRoomID, respawnHP)
		}
		story := alice.waitEvent(t, "EVT PLAYER DEATH ")
		if len(story) < len("EVT PLAYER DEATH ")+20 || strings.Contains(story, "did not survive") {
			t.Errorf("%s: first death message should be the room's story, got %q", turn.to, story)
		}
		alice.waitEvent(t, "EVT PLAYER DEATH ")
	}
}
