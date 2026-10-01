package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

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
		Items:       map[string]*Item{"item.stake": {Name: en("Stake"), RoomID: "loc.start", Obtainable: true}},
		NPCs: map[string]*NPC{
			"npc.beast": {Name: en("Beast"), Role: "enemy", RoomID: "loc.start", HP: 100, MythRequirementItem: "item.stake"},
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
			"npc.giant": {Name: en("Giant"), Role: "enemy", RoomID: "loc.start", HP: 65, Unwinnable: true, CrewLossOnAttack: 8, FleeAccurate: true},
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
			"npc.grunt": {Name: en("Grunt"), Role: "enemy", RoomID: "loc.start", HP: 8},
			"npc.giver": {Name: en("Giver"), Role: "quest_giver", RoomID: "loc.start", Dialogue: ens("Kill it.")},
		},
		Quests: map[string]*Quest{
			"quest.defeat": {
				Name: en("Defeat the Grunt"), Description: en("Defeat the grunt."), GiverNPCID: "npc.giver",
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
		Rooms:       map[string]*Room{"loc.start": {ID: "loc.start"}, "loc.empty": {ID: "loc.empty"}},
		NPCs: map[string]*NPC{
			"npc.brave":    {Name: en("Brave"), Role: "enemy", RoomID: "loc.start", HP: 1000, FleeAccurate: true},
			"npc.stubborn": {Name: en("Stubborn"), Role: "enemy", RoomID: "loc.start", HP: 1000, FleeAccurate: false},
		},
	}
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	server.mu.Lock()
	server.players["alice"].RoomID = "loc.empty"
	server.mu.Unlock()
	alice.cmd(t, "FLEE", "ERR 407 NOT_IN_COMBAT")
	server.mu.Lock()
	server.players["alice"].RoomID = "loc.start"
	server.mu.Unlock()

	atk := alice.cmdJSON(t, "ATTACK npc.brave")
	fleeData := alice.cmdJSON(t, "FLEE")
	if fleeData["result"] != "success" || fleeData["hp"] != atk["attacker_hp"] {
		t.Fatalf("flee from a myth-accurate retreat should succeed with unchanged hp, attack=%v flee=%v", atk, fleeData)
	}

	server.mu.Lock()
	if server.players["alice"].CombatTargetID != "" {
		t.Errorf("combat target after fleeing = %q, want none", server.players["alice"].CombatTargetID)
	}
	server.mu.Unlock()

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
				"east": "loc.gate", "south": "loc.pit", "west": "loc.strait", "north": "loc.toll",
			}},
			"loc.gate":   {ID: "loc.gate", Name: en("Gate"), Exits: map[string]string{"west": "loc.start"}, Hazard: &RoomHazard{Type: "item_gate", RequiredItemID: "item.key"}},
			"loc.pit":    {ID: "loc.pit", Name: en("Pit"), Hazard: &RoomHazard{Type: "lethal"}},
			"loc.strait": {ID: "loc.strait", Name: en("Strait"), Hazard: &RoomHazard{Type: "crew_gate", CrewLoss: 6, MinPartyTotal: 7}},
			"loc.toll":   {ID: "loc.toll", Name: en("Toll Road"), Hazard: &RoomHazard{Type: "crew_cost", CrewLoss: 2}},
		},
		Items: map[string]*Item{
			"item.key": {Name: en("Brass Key"), RoomID: "loc.start", Obtainable: true},
		},
	}
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	alice.cmd(t, "MOVE south", "OK room=loc.pit")
	alice.cmd(t, "STATUS", `OK {"hp":20,"max_hp":100,"status":"healthy"}`)

	alice.cmd(t, "MOVE east", "OK room=loc.gate")
	alice.cmd(t, "STATUS", `OK {"hp":20,"max_hp":100,"status":"healthy"}`)
	alice.cmd(t, "TAKE item.key", "OK taken=item.key")
	alice.cmd(t, "MOVE east", "OK room=loc.gate")
	alice.cmd(t, "STATUS", `OK {"hp":20,"max_hp":100,"status":"healthy"}`)
	alice.cmd(t, "MOVE west", "OK room=loc.start")

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

	server.mu.Lock()
	server.players["alice"].RoomID = "loc.start"
	server.players["alice"].Crew = 5
	server.mu.Unlock()
	alice.cmd(t, "MOVE north", "OK room=loc.toll")
	server.mu.Lock()
	crew = server.players["alice"].Crew
	server.mu.Unlock()
	if crew != 3 {
		t.Fatalf("crew after a flat crew_cost hazard = %d, want 3 (5-2)", crew)
	}
}

func TestLiveEnemyBlocksMove(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = &World{
		StartRoomID: "loc.safe",
		Rooms: map[string]*Room{
			"loc.safe":  {ID: "loc.safe", Exits: map[string]string{"east": "loc.start"}},
			"loc.start": {ID: "loc.start", Exits: map[string]string{"west": "loc.safe", "east": "loc.den"}},
			"loc.den":   {ID: "loc.den", Exits: map[string]string{"west": "loc.start"}},
		},
		NPCs: map[string]*NPC{
			"npc.guard":  {Name: en("Guard"), Role: "enemy", RoomID: "loc.den", HP: 8},
			"npc.runner": {Name: en("Runner"), Role: "enemy", RoomID: "loc.den", HP: 1000, FleeAccurate: true},
		},
	}
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	alice.cmd(t, "MOVE east", "OK room=loc.start")
	alice.cmd(t, "MOVE east", "OK room=loc.den")

	alice.cmd(t, "MOVE west", "OK room=loc.start")
	server.mu.Lock()
	respawned := server.players["alice"].RoomID == "loc.safe"
	server.mu.Unlock()
	if !respawned {
		t.Fatalf("moving out of a room with two live, unresolved enemies should kill and respawn alice at the safe room")
	}

	server.mu.Lock()
	server.players["alice"].RoomID = "loc.den"
	server.mu.Unlock()
	atk := alice.cmdJSON(t, "ATTACK npc.runner")
	if atk["status"] != "combat" {
		t.Fatalf("runner should survive one hit (HP 1000), got %v", atk)
	}
	flee := alice.cmdJSON(t, "FLEE")
	if flee["result"] != "success" {
		t.Fatalf("fleeing an flee-accurate NPC should succeed, got %v", flee)
	}

	alice.cmd(t, "MOVE west", "OK room=loc.start")
	server.mu.Lock()
	respawned = server.players["alice"].RoomID == "loc.safe"
	server.mu.Unlock()
	if !respawned {
		t.Fatalf("the guard is still alive and unresolved, so leaving should still be blocked and kill alice")
	}

	server.mu.Lock()
	server.players["alice"].RoomID = "loc.den"
	server.mu.Unlock()

	alice.cmdJSON(t, "ATTACK npc.runner")
	alice.cmdJSON(t, "FLEE")
	won := false
	for range 5 {
		result := alice.cmdJSON(t, "ATTACK npc.guard")
		if result["status"] == "victory" {
			won = true
			break
		}
	}
	if !won {
		t.Fatalf("did not defeat the guard (HP 8) within 5 rounds")
	}

	alice.cmd(t, "MOVE west", "OK room=loc.start")
	server.mu.Lock()
	leftDen := server.players["alice"].RoomID == "loc.start"
	server.mu.Unlock()
	if !leftDen {
		t.Fatalf("both enemies are now resolved (fled/defeated), leaving should succeed")
	}
}

func TestAttackMythGateByCompletedQuest(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = &World{
		StartRoomID: "loc.start",
		Rooms:       map[string]*Room{"loc.start": {ID: "loc.start"}},
		NPCs: map[string]*NPC{
			"npc.giant":  {Name: en("Bronze Giant"), Role: "enemy", RoomID: "loc.start", HP: 100, MythRequirementQuest: "quest.favor"},
			"npc.helper": {Name: en("Helper"), Role: "quest_giver", RoomID: "loc.start", Dialogue: ens("Help me first.")},
		},
		Quests: map[string]*Quest{
			"quest.favor": {
				Name: en("Earn Favor"), Description: en("Earn the sorceress's favor."), GiverNPCID: "npc.helper",
				Objective: QuestObjective{Type: "defeat_npc", TargetID: "npc.placeholder", Count: 1},
				Reward:    QuestReward{HP: 5},
			},
		},
	}
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	beast := alice.cmdJSON(t, "ATTACK npc.giant")
	if beast["status"] != "dead" {
		t.Fatalf("attacking a quest-gated NPC without the quest done should kill alice, got %v", beast)
	}

	server.mu.Lock()
	server.players["alice"].HP = maxPlayerHP
	server.players["alice"].Quests = map[string]*PlayerQuest{"quest.favor": {Status: "completed", Progress: 1}}
	server.mu.Unlock()

	result := alice.cmdJSON(t, "ATTACK npc.giant")
	if result["status"] == "dead" {
		t.Fatalf("attacking with the gating quest completed should not be instant death: %v", result)
	}
}

func TestFleeSucceedsOnceThenFails(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = &World{
		StartRoomID: "loc.start",
		Rooms:       map[string]*Room{"loc.start": {ID: "loc.start"}},
		NPCs: map[string]*NPC{
			"npc.hero": {Name: en("Proud Defender"), Role: "enemy", RoomID: "loc.start", HP: 1000, FleeSucceedsOnce: true},
		},
	}
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	alice.cmdJSON(t, "ATTACK npc.hero")
	first := alice.cmdJSON(t, "FLEE")
	if first["result"] != "success" {
		t.Fatalf("first flee ever from this NPC should succeed, got %v", first)
	}

	alice.cmdJSON(t, "ATTACK npc.hero")
	second := alice.cmdJSON(t, "FLEE")
	if second["result"] == "success" {
		t.Fatalf("a second flee from this NPC, even in a new encounter, should fail: %v", second)
	}
}

func TestTalkMythGateOnNonHostileNPC(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = &World{
		StartRoomID: "loc.start",
		Rooms:       map[string]*Room{"loc.start": {ID: "loc.start"}},
		Items:       map[string]*Item{"item.moly": {Name: en("Moly Root"), RoomID: "loc.start", Obtainable: true}},
		NPCs: map[string]*NPC{
			"npc.host": {Name: en("Host"), Role: "quest_giver", RoomID: "loc.start", MythRequirementItem: "item.moly", Dialogue: ens("Welcome.")},
		},
	}
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	alice.cmd(t, "TALK npc.host", "OK dead")
	alice.cmd(t, "STATUS", `OK {"hp":20,"max_hp":100,"status":"healthy"}`)

	alice.cmd(t, "TAKE item.moly", "OK taken=item.moly")
	alice.cmd(t, "TALK npc.host", "OK Welcome.")
}

func TestLethalRoomTellsWhatHappened(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = &World{
		StartRoomID: "loc.start",
		Rooms: map[string]*Room{
			"loc.start": {ID: "loc.start", Name: en("Hall"), Exits: map[string]string{"south": "loc.pit"}},
			"loc.pit": {ID: "loc.pit", Name: en("Pit"), Description: en("The floor gives way beneath you."),
				Hazard: &RoomHazard{Type: "lethal"}},
		},
	}
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	alice.cmd(t, "MOVE south", "OK room=loc.pit")
	story := alice.waitEvent(t, "EVT PLAYER DEATH ")
	if !strings.Contains(story, "The floor gives way beneath you.") {
		t.Errorf("first death message = %q, want the room's description", story)
	}
	verdict := alice.waitEvent(t, "EVT PLAYER DEATH ")
	if !strings.Contains(verdict, "Pit") || !strings.Contains(verdict, "Hall") {
		t.Errorf("second death message = %q, want the cause and the respawn room", verdict)
	}
}
