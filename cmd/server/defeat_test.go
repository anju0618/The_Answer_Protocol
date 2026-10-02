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

func attackPeopleWorld() *World {
	return &World{
		StartRoomID: "loc.start",
		Rooms:       map[string]*Room{"loc.start": {ID: "loc.start"}},
		NPCs: map[string]*NPC{
			"npc.villager": {Name: en("Villager"), Role: "dialogue", RoomID: "loc.start", HP: 20, Dialogue: ens("Hello.")},
			"npc.witch":    {Name: en("Witch"), Role: "quest_giver", RoomID: "loc.start", HP: 45, Mighty: true, Dialogue: ens("Hm.")},
		},
	}
}

func TestAttackingAPersonWoundsThenKillingThemEndsYourRun(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = attackPeopleWorld()
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	first := alice.cmdJSON(t, "ATTACK npc.villager")
	if first["status"] != "wounded" || first["attacker_hp"] != float64(100) {
		t.Fatalf("first hit = %v, want wounded with no counter-attack", first)
	}
	killed := false
	for hit := 2; hit <= 3; hit++ {
		result := alice.cmdJSON(t, "ATTACK npc.villager")
		if result["status"] == "murder" {
			killed = true
			break
		}
		if result["status"] != "wounded" || result["attacker_hp"] != float64(100) {
			t.Fatalf("hit %d = %v, want wounded with no counter-attack", hit, result)
		}
	}
	if !killed {
		t.Fatal("villager survived three hits of at least 8 damage despite having 20 HP")
	}
	server.mu.Lock()
	hp, bonus := server.players["alice"].HP, server.players["alice"].EnemyHP
	server.mu.Unlock()
	if hp != respawnHP || len(bonus) != 0 {
		t.Fatalf("after the murder HP = %d, wounds = %v; want respawn at %d HP with the villager restored", hp, bonus, respawnHP)
	}
	alice.cmd(t, "TALK npc.villager", "OK Hello.")
}

func TestAttackingAMightyNPCIsInstantDeath(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = attackPeopleWorld()
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	data := alice.cmdJSON(t, "ATTACK npc.witch")
	if data["status"] != "smitten" || data["damage"] != float64(0) {
		t.Fatalf("attack on the witch = %v, want smitten", data)
	}
	server.mu.Lock()
	hp := server.players["alice"].HP
	server.mu.Unlock()
	if hp != respawnHP {
		t.Fatalf("HP after being smitten = %d, want %d", hp, respawnHP)
	}
}

func TestAttackingADefeatedEnemyIsStillRefused(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = defeatTestWorld()
	alice := startTestClient(t, server)
	alice.connect(t, "alice")
	alice.cmdJSON(t, "ATTACK npc.grunt")
	alice.cmd(t, "ATTACK npc.grunt", "ERR 405 NPC_NOT_HOSTILE")
}
