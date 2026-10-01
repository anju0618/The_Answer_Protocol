package main

import "testing"

func TestDefendNeedsAFightAndHalvesTheNextCounter(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = &World{
		StartRoomID: "loc.start",
		Rooms:       map[string]*Room{"loc.start": {ID: "loc.start"}},
		NPCs: map[string]*NPC{
			"npc.brute": {Name: en("Brute"), Role: "enemy", RoomID: "loc.start", HP: 1000},
		},
	}
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	alice.cmd(t, "DEFEND", "ERR 407 NOT_IN_COMBAT")

	old := randDamage
	randDamage = func(min, max int) int { return 10 }
	defer func() { randDamage = old }()

	alice.cmdJSON(t, "ATTACK npc.brute") // 10 damage back at us
	server.mu.Lock()
	afterPlain := server.players["alice"].HP
	server.mu.Unlock()

	braced := alice.cmdJSON(t, "DEFEND")
	if braced["result"] != "braced" || braced["hp"] != float64(afterPlain) {
		t.Fatalf("DEFEND = %v, want braced with unchanged hp %d", braced, afterPlain)
	}
	alice.cmdJSON(t, "ATTACK npc.brute") // the counter is halved: 10 -> 5
	server.mu.Lock()
	afterGuarded := server.players["alice"].HP
	guarding := server.players["alice"].guarding
	server.mu.Unlock()
	if afterGuarded != afterPlain-5 {
		t.Fatalf("hp after guarded counter = %d, want %d", afterGuarded, afterPlain-5)
	}
	if guarding {
		t.Fatal("the stance must end after one counter-attack")
	}
	alice.cmdJSON(t, "ATTACK npc.brute") // back to the full 10
	server.mu.Lock()
	defer server.mu.Unlock()
	if got := server.players["alice"].HP; got != afterGuarded-10 {
		t.Fatalf("hp after unguarded counter = %d, want %d", got, afterGuarded-10)
	}
}
