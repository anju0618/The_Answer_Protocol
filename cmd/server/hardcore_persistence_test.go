package main

import (
	"os"
	"reflect"
	"testing"
)

func TestDeathReturnedItemOwnershipSurvivesRestart(t *testing.T) {
	saveDir := t.TempDir()
	server := newServer(saveDir)
	server.world = loadRealWorld(t)
	alice := newWalker(t, server, "alice")
	alice.teleport("loc.argo_lemnos")
	alice.client.cmd(t, "TAKE item.hospitality_gift", "OK taken=item.hospitality_gift")
	server.mu.Lock()
	server.players["alice"].Inventory = append(server.players["alice"].Inventory, "item.thread_of_fate")
	server.players["alice"].Quests = map[string]*PlayerQuest{"quest.saved": {Status: "active", Progress: 2}}
	server.mu.Unlock()
	alice.client.cmd(t, "QUIT", "OK bye")
	<-alice.client.done

	alice = newWalker(t, server, "alice")
	bob := newWalker(t, server, "bob")
	alice.teleport("loc.ody_cyclops")
	if result := alice.client.cmdJSON(t, "ATTACK npc.polyphemus"); result["status"] != "dead" {
		t.Fatalf("unprepared attack = %v, want death", result)
	}
	if alice.has("item.hospitality_gift") || !alice.has("item.thread_of_fate") {
		t.Fatal("death should lose the unique item and retain the trophy")
	}
	bob.teleport("loc.argo_lemnos")
	bob.client.cmd(t, "TAKE item.hospitality_gift", "OK taken=item.hospitality_gift")
	bob.client.cmd(t, "QUIT", "OK bye")
	<-bob.client.done
	server.mu.Lock()
	aliceOnline := server.players["alice"] != nil
	server.mu.Unlock()
	if !aliceOnline {
		t.Fatal("regression requires Alice to remain online without saving again")
	}

	restarted := newServer(saveDir)
	restarted.world = loadRealWorld(t)
	if err := restarted.restoreItemLocations(); err != nil {
		t.Fatal(err)
	}
	if err := restarted.restoreItemOwnership(); err != nil {
		t.Fatalf("restart after death and transfer: %v", err)
	}
	players, err := restarted.loadPlayers()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(players["alice"].Inventory, []string{"item.thread_of_fate"}) || !reflect.DeepEqual(players["bob"].Inventory, []string{"item.hospitality_gift"}) {
		t.Fatalf("saved inventories after death and transfer: alice=%v bob=%v", players["alice"].Inventory, players["bob"].Inventory)
	}
	if state := players["alice"].Quests["quest.saved"]; state == nil || state.Status != "active" || state.Progress != 2 {
		t.Fatalf("death changed saved quest progress: %+v", state)
	}
	if players["alice"].RoomID != "loc.argo_lemnos" || players["alice"].HP != maxPlayerHP {
		t.Fatalf("death saved transient pre-respawn state: %+v", players["alice"])
	}
	if room := restarted.world.Items["item.hospitality_gift"].RoomID; room != "" {
		t.Fatalf("Bob's saved item appeared in room %q after restart", room)
	}
}

func TestDeathPersistenceFailureKeepsBelongings(t *testing.T) {
	for _, failure := range []string{"item save write", "player save read"} {
		t.Run(failure, func(t *testing.T) {
			server := newServer(t.TempDir())
			server.world = takeTestWorld()
			server.world.Items["item.sword"].HomeRoomID = "loc.start"
			server.world.Items["item.sword"].RoomID = ""
			server.world.Items["item.shield"].Renewable = true
			server.world.Items["item.trophy"] = &Item{RewardOnly: true}
			player := &Player{
				Name: "alice", HP: 60, RoomID: "loc.other",
				Inventory: []string{"item.sword", "item.shield", "item.trophy"},
				Quests:    map[string]*PlayerQuest{"quest.saved": {Status: "active", Progress: 2}},
			}
			server.players["alice"] = player
			if err := server.savePlayer(player); err != nil {
				t.Fatal(err)
			}
			previousLocations := map[string]string{"item.sword": "loc.other"}
			if err := server.writeItemLocations(previousLocations); err != nil {
				t.Fatal(err)
			}
			player.Inventory = append(player.Inventory, "item.remote")
			player.Quests["quest.saved"].Progress = 3
			server.world.Items["item.remote"].RoomID = ""
			server.unsavedTakes["alice"] = map[string]string{"item.remote": "loc.other"}
			wantInventory := append([]string(nil), player.Inventory...)
			previousPlayers, err := os.ReadFile(server.playersSavePath())
			if err != nil {
				t.Fatal(err)
			}

			switch failure {
			case "item save write":
				if err := os.Chmod(server.saveDir, 0500); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { os.Chmod(server.saveDir, 0700) })
				if probe, err := os.CreateTemp(server.saveDir, ".permission-check-*"); err == nil {
					probe.Close()
					t.Skip("directory permissions do not prevent writes for this user")
				}
			case "player save read":
				if err := os.WriteFile(server.playersSavePath(), []byte("{"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			server.mu.Lock()
			outcome := server.applyDeathPenaltyLocked(player, "alice")
			server.mu.Unlock()
			if outcome != outcomeNothingLost || !reflect.DeepEqual(player.Inventory, wantInventory) {
				t.Fatalf("failed death persistence lost belongings: outcome=%q inventory=%v", outcome, player.Inventory)
			}
			for _, itemID := range []string{"item.sword", "item.remote"} {
				if room := server.world.Items[itemID].RoomID; room != "" {
					t.Fatalf("failed death persistence released %s to %q", itemID, room)
				}
			}
			if room := server.unsavedTakes["alice"]["item.remote"]; room != "loc.other" {
				t.Fatalf("failed death persistence discarded unsaved take: %q", room)
			}
			if progress := player.Quests["quest.saved"].Progress; progress != 3 {
				t.Fatalf("failed death persistence changed current quest progress to %d", progress)
			}
			if err := os.Chmod(server.saveDir, 0700); err != nil {
				t.Fatal(err)
			}
			if failure == "player save read" {
				locations, err := server.loadItemLocations()
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(locations, previousLocations) {
					t.Fatalf("failed player save did not restore previous item locations: %v", locations)
				}
				if err := os.WriteFile(server.playersSavePath(), previousPlayers, 0600); err != nil {
					t.Fatal(err)
				}
			}
			players, err := server.loadPlayers()
			if err != nil {
				t.Fatal(err)
			}
			if !players["alice"].hasItem("item.sword") || players["alice"].Quests["quest.saved"].Progress != 2 {
				t.Fatalf("failed death persistence changed saved inventory or progress: %+v", players["alice"])
			}

			if err := server.writeItemLocations(map[string]string{"item.sword": "loc.start"}); err != nil {
				t.Fatal(err)
			}
			restarted := newServer(server.saveDir)
			restarted.world = takeTestWorld()
			if err := restarted.restoreItemLocations(); err != nil {
				t.Fatal(err)
			}
			if err := restarted.restoreItemOwnership(); err != nil {
				t.Fatal(err)
			}
			if room := restarted.world.Items["item.sword"].RoomID; room != "" {
				t.Fatalf("failed death persistence exposed a saved owner's item after restart: %q", room)
			}
		})
	}
}
