package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// cmd behaves like testClient.command, but first discards any pending
// "EVT ..." lines. ATTACK/FLEE/MOVE/TAKE/TALK now broadcast a combat/presence
// event to the acting player's own connection (same convention as MOVE's
// self "EVT ROOM PRESENCE ENTER" notification), queued right after that
// command's own response. The next command's reply is therefore not always
// the very next line on the wire.
func (client *testClient) cmd(t *testing.T, command, want string) {
	t.Helper()
	if _, err := fmt.Fprintln(client.conn, command); err != nil {
		t.Fatalf("send %q: %v", command, err)
	}
	for {
		line, err := client.reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read response: %v", err)
		}
		line = strings.TrimSuffix(line, "\n")
		if strings.HasPrefix(line, "EVT ") {
			continue
		}
		if line != want {
			t.Fatalf("response = %q, want %q", line, want)
		}
		return
	}
}

// cmdJSON is cmd's counterpart for responses whose exact bytes aren't
// predictable (they contain a random damage roll).
func (client *testClient) cmdJSON(t *testing.T, command string) map[string]any {
	t.Helper()
	if _, err := fmt.Fprintln(client.conn, command); err != nil {
		t.Fatalf("send %q: %v", command, err)
	}
	for {
		line, err := client.reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read response: %v", err)
		}
		line = strings.TrimSuffix(line, "\n")
		if strings.HasPrefix(line, "EVT ") {
			continue
		}
		if !strings.HasPrefix(line, "OK ") {
			t.Fatalf("response = %q, want OK-prefixed JSON", line)
		}
		var data map[string]any
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "OK ")), &data); err != nil {
			t.Fatalf("decode %q: %v", line, err)
		}
		return data
	}
}

func TestAttackMythGateKillsWithoutItem(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = &World{
		StartRoomID: "loc.start",
		Rooms:       map[string]*Room{"loc.start": {ID: "loc.start"}},
		Items:       map[string]*Item{"item.stake": {Name: "Stake", RoomID: "loc.start", Obtainable: true}},
		NPCs: map[string]*NPC{
			"npc.beast": {Name: "Beast", Role: "enemy", RoomID: "loc.start", HP: 100, MythRequirementItem: "item.stake"},
		},
	}
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	alice.cmd(t, "ATTACK npc.beast", `OK {"attacker_hp":0,"target_hp":100,"damage":0,"status":"dead"}`)
	alice.cmd(t, "STATUS", `OK {"hp":20,"max_hp":100,"status":"healthy"}`)

	alice.cmd(t, "TAKE item.stake", "OK taken=item.stake")
	data := alice.cmdJSON(t, "ATTACK npc.beast")
	if data["status"] == "dead" {
		t.Fatalf("attack with required item should not be instant death: %v", data)
	}
}

func TestAttackUnwinnableCostsCrewNotHP(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = &World{
		StartRoomID: "loc.start",
		Rooms:       map[string]*Room{"loc.start": {ID: "loc.start"}},
		NPCs: map[string]*NPC{
			"npc.giant": {Name: "Giant", Role: "enemy", RoomID: "loc.start", HP: 65, Unwinnable: true, CrewLossOnAttack: 8, FleeAccurate: true},
		},
	}
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	server.mu.Lock()
	server.players["alice"].Crew = 20
	server.mu.Unlock()

	alice.cmd(t, "ATTACK npc.giant", `OK {"attacker_hp":100,"target_hp":65,"damage":0,"status":"overwhelmed"}`)
	alice.cmd(t, "STATUS", `OK {"hp":100,"max_hp":100,"status":"healthy"}`)

	server.mu.Lock()
	crew := server.players["alice"].Crew
	server.mu.Unlock()
	if crew != 12 {
		t.Fatalf("crew = %d, want 12 (20-8)", crew)
	}
}

func TestAttackDefeatCompletesQuestAndHeals(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = &World{
		StartRoomID: "loc.start",
		Rooms:       map[string]*Room{"loc.start": {ID: "loc.start"}},
		NPCs: map[string]*NPC{
			"npc.grunt": {Name: "Grunt", Role: "enemy", RoomID: "loc.start", HP: 8}, // <= min damage: dies in one hit
			"npc.giver": {Name: "Giver", Role: "quest_giver", RoomID: "loc.start", Dialogue: []string{"Kill it."}},
		},
		Quests: map[string]*Quest{
			"quest.defeat": {
				Name: "Defeat the Grunt", Description: "Defeat the grunt.", GiverNPCID: "npc.giver",
				Objective: QuestObjective{Type: "defeat_npc", TargetID: "npc.grunt", Count: 1},
				Reward:    QuestReward{HP: 30},
			},
		},
	}
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	alice.cmd(t, "QUEST npc.giver", `OK {"quest_id":"quest.defeat","description":"Defeat the grunt.","reward":30,"status":"available"}`)

	server.mu.Lock()
	server.players["alice"].HP = 50
	server.mu.Unlock()

	data := alice.cmdJSON(t, "ATTACK npc.grunt")
	if data["status"] != "victory" || data["target_hp"] != float64(0) {
		t.Fatalf("attack result = %v, want a victory with target_hp 0", data)
	}
	alice.cmd(t, "STATUS", `OK {"hp":80,"max_hp":100,"status":"healthy"}`)
	alice.cmd(t, "QUESTS", `OK [{"quest_id":"quest.defeat","status":"completed","progress":"1/1"}]`)
	alice.cmd(t, "QUEST npc.giver", "ERR 406 NO_QUEST_AVAILABLE")
}

func TestFleeSuccessAndFailureAndNotInCombat(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = &World{
		StartRoomID: "loc.start",
		Rooms:       map[string]*Room{"loc.start": {ID: "loc.start"}},
		NPCs: map[string]*NPC{
			"npc.brave":    {Name: "Brave", Role: "enemy", RoomID: "loc.start", HP: 1000, FleeAccurate: true},
			"npc.stubborn": {Name: "Stubborn", Role: "enemy", RoomID: "loc.start", HP: 1000, FleeAccurate: false},
		},
	}
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	alice.cmd(t, "FLEE", "ERR 407 NOT_IN_COMBAT")

	atk := alice.cmdJSON(t, "ATTACK npc.brave") // survives (HP 1000), enters combat
	fleeData := alice.cmdJSON(t, "FLEE")
	if fleeData["result"] != "success" || fleeData["hp"] != atk["attacker_hp"] {
		t.Fatalf("flee from a myth-accurate retreat should succeed with unchanged hp, attack=%v flee=%v", atk, fleeData)
	}
	alice.cmd(t, "FLEE", "ERR 407 NOT_IN_COMBAT")

	alice.cmdJSON(t, "ATTACK npc.stubborn")
	fleeResp := alice.cmdJSON(t, "FLEE")
	if fleeResp["result"] != "failure" {
		t.Fatalf("flee from a myth-inaccurate retreat should fail, got %v", fleeResp)
	}
}

func TestRoomHazards(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = &World{
		StartRoomID: "loc.start",
		Rooms: map[string]*Room{
			"loc.start": {ID: "loc.start", Exits: map[string]string{
				"east": "loc.gate", "south": "loc.pit", "west": "loc.strait",
			}},
			"loc.gate":   {ID: "loc.gate", Name: "Gate", Exits: map[string]string{"west": "loc.start"}, Hazard: &RoomHazard{Type: "item_gate", RequiredItemID: "item.key"}},
			"loc.pit":    {ID: "loc.pit", Name: "Pit", Hazard: &RoomHazard{Type: "lethal"}},
			"loc.strait": {ID: "loc.strait", Name: "Strait", Hazard: &RoomHazard{Type: "crew_gate", CrewLoss: 6, MinPartyTotal: 7}},
		},
		Items: map[string]*Item{
			"item.key": {Name: "Brass Key", RoomID: "loc.start", Obtainable: true},
		},
	}
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	// lethal: always dies and respawns at the start room.
	alice.cmd(t, "MOVE south", "OK room=loc.pit")
	alice.cmd(t, "STATUS", `OK {"hp":20,"max_hp":100,"status":"healthy"}`)

	// item_gate: dies without the key, survives with it.
	alice.cmd(t, "MOVE east", "OK room=loc.gate")
	alice.cmd(t, "STATUS", `OK {"hp":20,"max_hp":100,"status":"healthy"}`)
	alice.cmd(t, "TAKE item.key", "OK taken=item.key")
	alice.cmd(t, "MOVE east", "OK room=loc.gate")
	alice.cmd(t, "STATUS", `OK {"hp":20,"max_hp":100,"status":"healthy"}`) // unchanged: no death this time
	alice.cmd(t, "MOVE west", "OK room=loc.start")

	// crew_gate: not enough crew (Crew+1 < 7) dies; enough crew survives and pays the toll.
	server.mu.Lock()
	server.players["alice"].Crew = 3
	server.mu.Unlock()
	alice.cmd(t, "MOVE west", "OK room=loc.strait")
	server.mu.Lock()
	died := server.players["alice"].RoomID == "loc.start" && server.players["alice"].Crew == 3
	server.players["alice"].Crew = 10
	server.players["alice"].RoomID = "loc.start"
	server.mu.Unlock()
	if !died {
		t.Fatalf("expected death and respawn when crew+1 < min_party_total")
	}
	alice.cmd(t, "MOVE west", "OK room=loc.strait")
	server.mu.Lock()
	crew := server.players["alice"].Crew
	server.mu.Unlock()
	if crew != 4 {
		t.Fatalf("crew after passing scylla-like gate = %d, want 4 (10-6)", crew)
	}
}

func TestTalkMythGateOnNonHostileNPC(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = &World{
		StartRoomID: "loc.start",
		Rooms:       map[string]*Room{"loc.start": {ID: "loc.start"}},
		Items:       map[string]*Item{"item.moly": {Name: "Moly Root", RoomID: "loc.start", Obtainable: true}},
		NPCs: map[string]*NPC{
			"npc.host": {Name: "Host", Role: "quest_giver", RoomID: "loc.start", MythRequirementItem: "item.moly", Dialogue: []string{"Welcome."}},
		},
	}
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	alice.cmd(t, "TALK npc.host", "OK dead")
	alice.cmd(t, "STATUS", `OK {"hp":20,"max_hp":100,"status":"healthy"}`)

	alice.cmd(t, "TAKE item.moly", "OK taken=item.moly")
	alice.cmd(t, "TALK npc.host", "OK Welcome.")
}
