package main

import "testing"

func TestSecretExitOpensAfterFinalEnding(t *testing.T) {
	room := &Room{
		Exits:       map[string]string{"north": "loc.a"},
		SecretExits: map[string]SecretExit{"south": {Room: "loc.b", RequiresEndings: []string{"ending.final"}}},
	}
	player := &Player{Endings: map[string]bool{}}

	if _, ok := room.exitsFor(player)["south"]; ok {
		t.Fatal("secret exit is visible before the ending")
	}
	player.Endings["ending.final"] = true
	exits := room.exitsFor(player)
	if exits["south"] != "loc.b" || exits["north"] != "loc.a" {
		t.Fatalf("exits after the ending = %v", exits)
	}
}
