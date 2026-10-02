package main

import (
	"fmt"
	"strings"
	"testing"
)

func defeatTestWorld() *World {
	return &World{
		StartRoomID: "loc.start",
		Rooms:       map[string]*Room{"loc.start": {ID: "loc.start"}},
		NPCs: map[string]*NPC{
			"npc.grunt": {Name: en("Grunt"), Role: "enemy", RoomID: "loc.start", HP: 8},
			"npc.ghost": {Name: en("Ghost"), Role: "enemy", RoomID: "loc.start", HP: 8, Unwinnable: true},
			"npc.giver": {Name: en("Giver"), Role: "dialogue", RoomID: "loc.start",
				Dialogue: ens("Help us!"), DialogueCleared: ens("Thank you!")},
		},
	}
}

func TestDialogueAndLookChangeOnceTheRoomIsCleared(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = defeatTestWorld()
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	alice.cmd(t, "TALK npc.giver", "OK Help us!")
	if look := alice.cmdRaw(t, "LOOK"); strings.Contains(look, "defeated") {
		t.Fatalf("LOOK before the fight mentions defeated: %s", look)
	}

	data := alice.cmdJSON(t, "ATTACK npc.grunt")
	if data["status"] != "victory" {
		t.Fatalf("attack = %v, want victory", data)
	}
	alice.cmd(t, "TALK npc.giver", "OK Thank you!")
	if look := alice.cmdRaw(t, "LOOK"); !strings.Contains(look, `"defeated":["npc.grunt"]`) {
		t.Fatalf("LOOK after the fight = %s, want the defeated grunt (and not the unwinnable ghost)", look)
	}
}

func TestDefeatIsForgottenOnDeath(t *testing.T) {
	server := effectTestServer(nil)
	server.world.NPCs = defeatTestWorld().NPCs
	player := &Player{Name: "alice", RoomID: "loc.start", EnemyHP: map[string]int{"npc.grunt": 0}}
	if !server.roomClearedLocked(player, "loc.start") {
		t.Fatal("room should be cleared once the only beatable enemy is down")
	}
	server.applyDeathPenaltyLocked(player, "alice")
	if server.roomClearedLocked(player, "loc.start") {
		t.Fatal("the enemy should be back after the player died")
	}
}

// cmdRaw sends a command and returns the first non-event response line.
func (client *testClient) cmdRaw(t *testing.T, command string) string {
	t.Helper()
	if _, err := fmt.Fprintln(client.conn, command); err != nil {
		t.Fatalf("send %q: %v", command, err)
	}
	for {
		line, err := client.reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read response: %v", err)
		}
		if !strings.HasPrefix(line, "EVT ") {
			return strings.TrimSuffix(line, "\n")
		}
	}
}

