package main

import "testing"

func combatRoomRegressionWorld() *World {
	return &World{
		StartRoomID: "loc.enemy",
		Rooms: map[string]*Room{
			"loc.enemy": {ID: "loc.enemy", Exits: map[string]string{"east": "loc.safe", "wait": "loc.enemy"}},
			"loc.safe":  {ID: "loc.safe", Exits: map[string]string{"west": "loc.enemy"}},
		},
		NPCs: map[string]*NPC{
			"npc.hero": {Name: en("Hero"), Role: "enemy", RoomID: "loc.enemy", HP: 1000, FleeSucceedsOnce: true},
		},
	}
}

func TestMoveEndsCombatAfterFleeAndReattack(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = combatRoomRegressionWorld()
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	alice.cmdJSON(t, "ATTACK npc.hero")
	if result := alice.cmdJSON(t, "FLEE"); result["result"] != "success" {
		t.Fatalf("first FLEE = %v, want success", result)
	}
	attack := alice.cmdJSON(t, "ATTACK npc.hero")
	alice.cmdJSON(t, "DEFEND")
	alice.cmd(t, "MOVE missing", "ERR 301 NO_EXIT")
	alice.cmd(t, "MOVE wait", "OK room=loc.enemy")

	server.mu.Lock()
	player := server.players["alice"]
	if player.CombatTargetID != "npc.hero" || !player.guarding {
		t.Errorf("same-room or rejected MOVE cleared combat state: target=%q guarding=%v", player.CombatTargetID, player.guarding)
	}
	server.mu.Unlock()

	alice.cmd(t, "MOVE east", "OK room=loc.safe")
	server.mu.Lock()
	player = server.players["alice"]
	if player.RoomID != "loc.safe" || player.CombatTargetID != "" || player.guarding {
		t.Errorf("MOVE state: room=%q target=%q guarding=%v, want safe room without combat or guard", player.RoomID, player.CombatTargetID, player.guarding)
	}
	if !player.FledFrom["npc.hero"] {
		t.Error("MOVE lost the remembered first flee")
	}
	server.mu.Unlock()

	status := alice.cmdJSON(t, "STATUS")
	if status["status"] != "healthy" || status["hp"] != attack["attacker_hp"] {
		t.Errorf("STATUS after MOVE = %v, want healthy with hp %v", status, attack["attacker_hp"])
	}
	alice.cmd(t, "FLEE", "ERR 407 NOT_IN_COMBAT")
	alice.cmd(t, "DEFEND", "ERR 407 NOT_IN_COMBAT")
	if status := alice.cmdJSON(t, "STATUS"); status["hp"] != attack["attacker_hp"] {
		t.Errorf("out-of-room FLEE changed hp: %v, want %v", status, attack["attacker_hp"])
	}

	alice.cmd(t, "MOVE west", "OK room=loc.enemy")
	alice.cmdJSON(t, "ATTACK npc.hero")
	if result := alice.cmdJSON(t, "FLEE"); result["result"] != "failure" {
		t.Fatalf("second FLEE after returning = %v, want failure", result)
	}
}

func TestFleeRejectsStaleSavedCombatTarget(t *testing.T) {
	for _, scenario := range []string{"other room", "missing NPC", "non-hostile NPC", "defeated NPC"} {
		t.Run(scenario, func(t *testing.T) {
			server := newServer(t.TempDir())
			server.world = combatRoomRegressionWorld()
			saved := &Player{Name: "alice", HP: 50, RoomID: "loc.safe", CombatTargetID: "npc.hero"}
			switch scenario {
			case "missing NPC":
				server.world.NPCs = nil
			case "non-hostile NPC":
				server.world.NPCs["npc.hero"].RoomID = "loc.safe"
				server.world.NPCs["npc.hero"].Role = "quest_giver"
			case "defeated NPC":
				server.world.NPCs["npc.hero"].RoomID = "loc.safe"
				saved.EnemyHP = map[string]int{"npc.hero": 0}
			}
			if err := server.savePlayer(saved); err != nil {
				t.Fatal(err)
			}
			alice := startTestClient(t, server)
			alice.connect(t, "alice")
			server.mu.Lock()
			server.players["alice"].guarding = true
			server.mu.Unlock()

			alice.cmd(t, "FLEE", "ERR 407 NOT_IN_COMBAT")
			server.mu.Lock()
			defer server.mu.Unlock()
			player := server.players["alice"]
			if player.HP != 50 || player.CombatTargetID != "" || player.guarding {
				t.Errorf("FLEE state: hp=%d target=%q guarding=%v, want 50 with cleared combat and guard", player.HP, player.CombatTargetID, player.guarding)
			}
		})
	}
}

func TestFleeUsesCurrentRoomEnemyAfterStaleSavedTarget(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = combatRoomRegressionWorld()
	server.world.NPCs["npc.local"] = &NPC{Name: en("Local Guard"), Role: "enemy", RoomID: "loc.safe", HP: 1000, FleeAccurate: true}
	if err := server.savePlayer(&Player{
		Name: "alice", HP: 50, RoomID: "loc.safe", CombatTargetID: "npc.hero", FledFrom: map[string]bool{"npc.hero": true},
	}); err != nil {
		t.Fatal(err)
	}
	alice := startTestClient(t, server)
	alice.connect(t, "alice")
	if result := alice.cmdJSON(t, "FLEE"); result["result"] != "success" || result["hp"] != float64(50) {
		t.Fatalf("FLEE = %v, want success from the current room enemy with unchanged hp", result)
	}
	server.mu.Lock()
	defer server.mu.Unlock()
	player := server.players["alice"]
	if player.CombatTargetID != "" || !player.FledFrom["npc.local"] || !player.FledFrom["npc.hero"] {
		t.Errorf("FLEE state: target=%q fled=%v, want cleared target and remembered flee from both NPCs", player.CombatTargetID, player.FledFrom)
	}
}
