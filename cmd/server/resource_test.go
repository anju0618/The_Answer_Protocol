package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestInventoryDropAndTalk(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = takeTestWorld()
	server.world.Rooms["loc.start"].Exits = map[string]string{"east": "loc.other"}
	server.world.NPCs = map[string]*NPC{
		"npc.a":         {Name: "Old Guide", RoomID: "loc.start", Dialogue: []string{"Welcome back.", "Second line."}},
		"npc.z":         {Name: "Old Guide", RoomID: "loc.start", Dialogue: []string{"bye"}},
		"npc.remote":    {Name: "Far Guide", RoomID: "loc.other", Dialogue: []string{"You found me."}},
		"npc.silent":    {Name: "Silent", RoomID: "loc.start"},
		"npc.multiline": {Name: "Multiline", RoomID: "loc.start", Dialogue: []string{"Hi\nthere"}},
	}
	alice := startTestClient(t, server)
	alice.command(t, "INVENTORY", "ERR 400 BAD_REQUEST")
	alice.command(t, "DROP item.sword", "ERR 400 BAD_REQUEST")
	alice.command(t, "TALK Old Guide", "ERR 400 BAD_REQUEST")
	alice.connect(t, "alice")
	alice.command(t, "INVENTORY extra", "ERR 400 BAD_REQUEST")
	alice.command(t, "INVENTORY", "OK []")
	alice.command(t, "DROP", "ERR 400 BAD_REQUEST")
	alice.command(t, "DROP item.sword", "ERR 404 ITEM_NOT_IN_INVENTORY")
	alice.command(t, "TALK", "ERR 400 BAD_REQUEST")
	alice.command(t, "TALK npc.remote", "ERR 404 NPC_NOT_FOUND")
	alice.command(t, "TALK missing", "ERR 404 NPC_NOT_FOUND")
	alice.command(t, "TALK Old Guide", "OK Welcome back.")
	alice.command(t, "TALK npc.z", "OK bye")
	alice.command(t, "TALK npc.silent", "ERR 500 STATE_ERROR")
	alice.command(t, "TALK npc.multiline", "ERR 500 STATE_ERROR")
	alice.command(t, "WHO", "OK players=1")

	alice.command(t, "TAKE rUsTy SwOrD", "OK taken=item.sword")
	alice.command(t, "TAKE item.shield", "OK taken=item.shield")
	alice.command(t, "INVENTORY", `OK ["item.shield","item.sword"]`)
	alice.command(t, "DROP SILVER SHIELD", "OK dropped=item.shield")
	alice.command(t, "INVENTORY", `OK ["item.sword"]`)
	alice.command(t, "DROP item.shield", "ERR 404 ITEM_NOT_IN_INVENTORY")
	bob := startTestClient(t, server)
	bob.connect(t, "bob", alice)
	bob.command(t, "TAKE item.shield", "OK taken=item.shield")
	bob.command(t, "INVENTORY", `OK ["item.shield"]`)
	alice.command(t, "MOVE east", "OK room=loc.other")
	alice.expect(t, "EVT ROOM PRESENCE ENTER alice")
	bob.expect(t, "EVT ROOM PRESENCE LEAVE alice")
	alice.command(t, "TALK Far Guide", "OK You found me.")
	alice.command(t, "TALK npc.a", "ERR 404 NPC_NOT_FOUND")
}

func TestDropLocationAndOwnershipSurviveRestart(t *testing.T) {
	saveDir := t.TempDir()
	server := newServer(saveDir)
	server.world = takeTestWorld()
	server.world.Rooms["loc.start"].Exits = map[string]string{"east": "loc.other"}
	alice := startTestClient(t, server)
	alice.connect(t, "alice")
	alice.command(t, "TAKE item.sword", "OK taken=item.sword")
	alice.command(t, "MOVE east", "OK room=loc.other")
	alice.expect(t, "EVT ROOM PRESENCE ENTER alice")
	alice.command(t, "DROP item.sword", "OK dropped=item.sword")
	alice.command(t, "INVENTORY", "OK []")
	alice.command(t, "QUIT", "OK bye")
	<-alice.done

	restarted := newServer(saveDir)
	restarted.world = takeTestWorld()
	restarted.world.Rooms["loc.start"].Exits = map[string]string{"east": "loc.other"}
	if err := restarted.restoreItemLocations(); err != nil {
		t.Fatal(err)
	}
	if err := restarted.restoreItemOwnership(); err != nil {
		t.Fatal(err)
	}
	if got := restarted.world.Items["item.sword"].RoomID; got != "loc.other" {
		t.Fatalf("sword room after restart = %q, want loc.other", got)
	}
	bob := startTestClient(t, restarted)
	bob.connect(t, "bob")
	bob.command(t, "TAKE item.sword", "ERR 404 ITEM_NOT_FOUND")
	bob.command(t, "MOVE east", "OK room=loc.other")
	bob.expect(t, "EVT ROOM PRESENCE ENTER bob")
	bob.command(t, "TAKE item.sword", "OK taken=item.sword")
	bob.command(t, "QUIT", "OK bye")
	<-bob.done

	again := newServer(saveDir)
	again.world = takeTestWorld()
	if err := again.restoreItemLocations(); err != nil {
		t.Fatal(err)
	}
	if err := again.restoreItemOwnership(); err != nil {
		t.Fatal(err)
	}
	if got := again.world.Items["item.sword"].RoomID; got != "" {
		t.Fatalf("owned sword appeared in room %q after restart", got)
	}
	players, err := again.loadPlayers()
	if err != nil {
		t.Fatal(err)
	}
	if len(players["alice"].Inventory) != 0 || !reflect.DeepEqual(players["bob"].Inventory, []string{"item.sword"}) {
		t.Fatalf("inventories after transfer: alice=%v bob=%v", players["alice"].Inventory, players["bob"].Inventory)
	}
}

func TestDropSaveFailureKeepsItemHeld(t *testing.T) {
	saveDir := t.TempDir()
	server := newServer(saveDir)
	server.world = takeTestWorld()
	alice := startTestClient(t, server)
	alice.connect(t, "alice")
	alice.command(t, "TAKE item.sword", "OK taken=item.sword")
	bob := startTestClient(t, server)
	bob.connect(t, "bob", alice)
	blocked := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(blocked, []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}
	server.mu.Lock()
	server.saveDir = blocked
	server.mu.Unlock()
	alice.command(t, "DROP item.sword", "ERR 500 STATE_ERROR")
	alice.command(t, "INVENTORY", `OK ["item.sword"]`)
	bob.command(t, "TAKE item.sword", "ERR 404 ITEM_NOT_FOUND")
	server.mu.Lock()
	server.saveDir = saveDir
	server.mu.Unlock()
	alice.command(t, "DROP item.sword", "OK dropped=item.sword")
	bob.command(t, "TAKE item.sword", "OK taken=item.sword")
}

func TestDropRestoresItemLocationWhenPlayerSaveFails(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = takeTestWorld()
	alice := startTestClient(t, server)
	alice.connect(t, "alice")
	alice.command(t, "TAKE item.sword", "OK taken=item.sword")
	if err := os.WriteFile(server.playersSavePath(), []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	alice.command(t, "DROP item.sword", "ERR 500 STATE_ERROR")
	locations, err := server.loadItemLocations()
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := locations["item.sword"]; exists {
		t.Fatal("failed DROP left a saved item location")
	}
	alice.command(t, "INVENTORY", `OK ["item.sword"]`)
	server.mu.Lock()
	roomID := server.world.Items["item.sword"].RoomID
	server.mu.Unlock()
	if roomID != "" {
		t.Fatalf("failed DROP released sword to %q", roomID)
	}
	if err := os.WriteFile(server.playersSavePath(), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	alice.command(t, "QUIT", "OK bye")
	<-alice.done
}

func TestDropSavedItemCanTransferDespiteLaterSaveFailure(t *testing.T) {
	saveDir := t.TempDir()
	server := newServer(saveDir)
	server.world = takeTestWorld()
	alice := startTestClient(t, server)
	alice.connect(t, "alice")
	alice.command(t, "TAKE item.sword", "OK taken=item.sword")
	alice.command(t, "QUIT", "OK bye")
	<-alice.done
	reconnected := startTestClient(t, server)
	reconnected.connect(t, "alice")
	bob := startTestClient(t, server)
	bob.connect(t, "bob", reconnected)
	reconnected.command(t, "DROP item.sword", "OK dropped=item.sword")
	bob.command(t, "TAKE item.sword", "OK taken=item.sword")
	bob.command(t, "QUIT", "OK bye")
	<-bob.done
	blocked := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(blocked, []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}
	server.mu.Lock()
	server.saveDir = blocked
	server.mu.Unlock()
	if err := reconnected.conn.Close(); err != nil {
		t.Fatal(err)
	}
	<-reconnected.done
	server.mu.Lock()
	server.saveDir = saveDir
	server.mu.Unlock()

	restarted := newServer(saveDir)
	restarted.world = takeTestWorld()
	if err := restarted.restoreItemLocations(); err != nil {
		t.Fatal(err)
	}
	if err := restarted.restoreItemOwnership(); err != nil {
		t.Fatal(err)
	}
	players, err := restarted.loadPlayers()
	if err != nil {
		t.Fatal(err)
	}
	if len(players["alice"].Inventory) != 0 || !reflect.DeepEqual(players["bob"].Inventory, []string{"item.sword"}) {
		t.Fatalf("duplicate ownership after failed save: alice=%v bob=%v", players["alice"].Inventory, players["bob"].Inventory)
	}
	if got := restarted.world.Items["item.sword"].RoomID; got != "" {
		t.Fatalf("transferred sword appeared in room %q", got)
	}
}

func TestDropDoesNotRestoreOtherSavedItemsAfterDisconnectFailure(t *testing.T) {
	saveDir := t.TempDir()
	server := newServer(saveDir)
	server.world = takeTestWorld()
	alice := startTestClient(t, server)
	alice.connect(t, "alice")
	bob := startTestClient(t, server)
	bob.connect(t, "bob", alice)
	alice.command(t, "TAKE item.sword", "OK taken=item.sword")
	alice.command(t, "TAKE item.shield", "OK taken=item.shield")
	alice.command(t, "DROP item.sword", "OK dropped=item.sword")
	blocked := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(blocked, []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}
	server.mu.Lock()
	server.saveDir = blocked
	server.mu.Unlock()
	if err := alice.conn.Close(); err != nil {
		t.Fatal(err)
	}
	<-alice.done
	bob.expect(t, "EVT ROOM PRESENCE LEAVE alice")
	bob.expect(t, "EVT STATS players=1")
	bob.command(t, "TAKE item.shield", "ERR 404 ITEM_NOT_FOUND")
	bob.command(t, "TAKE item.sword", "OK taken=item.sword")
	server.mu.Lock()
	server.saveDir = saveDir
	server.mu.Unlock()
	bob.command(t, "QUIT", "OK bye")
	<-bob.done

	restarted := newServer(saveDir)
	restarted.world = takeTestWorld()
	if err := restarted.restoreItemLocations(); err != nil {
		t.Fatal(err)
	}
	if err := restarted.restoreItemOwnership(); err != nil {
		t.Fatal(err)
	}
	players, err := restarted.loadPlayers()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(players["alice"].Inventory, []string{"item.shield"}) || !reflect.DeepEqual(players["bob"].Inventory, []string{"item.sword"}) {
		t.Fatalf("saved ownership: alice=%v bob=%v", players["alice"].Inventory, players["bob"].Inventory)
	}
	if restarted.world.Items["item.sword"].RoomID != "" || restarted.world.Items["item.shield"].RoomID != "" {
		t.Fatal("owned item appeared in a room after restart")
	}
}

func TestRestoreItemLocationsRejectsUnknownReferences(t *testing.T) {
	for _, tc := range []struct {
		name      string
		locations map[string]string
	}{
		{"missing item", map[string]string{"item.unknown": "loc.start"}},
		{"missing room", map[string]string{"item.sword": "loc.unknown"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := newServer(t.TempDir())
			server.world = takeTestWorld()
			if err := server.writeItemLocations(tc.locations); err != nil {
				t.Fatal(err)
			}
			before := server.world.Items["item.sword"].RoomID
			if err := server.restoreItemLocations(); err == nil || !strings.Contains(err.Error(), "saved item") {
				t.Fatalf("restore error = %v", err)
			}
			if got := server.world.Items["item.sword"].RoomID; got != before {
				t.Fatalf("item room changed to %q despite invalid save", got)
			}
		})
	}
}
