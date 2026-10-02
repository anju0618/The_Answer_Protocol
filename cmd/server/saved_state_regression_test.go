package main

import (
	"bufio"
	"bytes"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoadPlayersRejectsNullQuestState(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = takeTestWorld()
	data := []byte(`{"alice":{"name":"alice","hp":20,"room_id":"loc.start","inventory":["item.sword"],"quests":{"quest.phineus_harpies":null}}}`)
	if err := os.WriteFile(server.playersSavePath(), data, 0600); err != nil {
		t.Fatal(err)
	}
	players, err := server.loadPlayers()
	if err == nil || players != nil {
		t.Fatalf("loaded malformed quest state: players=%v error=%v", players, err)
	}
	if !strings.Contains(err.Error(), "alice") || !strings.Contains(err.Error(), "quest.phineus_harpies") {
		t.Fatalf("validation error does not identify the player and quest: %v", err)
	}
	if err := server.restoreItemOwnership(); err == nil {
		t.Fatal("startup accepted malformed quest state")
	}
	if got := server.world.Items["item.sword"].RoomID; got != "loc.start" {
		t.Fatalf("failed restore changed item ownership: room=%q", got)
	}
	after, err := os.ReadFile(server.playersSavePath())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, data) {
		t.Fatal("validation changed the saved player data")
	}
}

func TestLoadPlayersAllowsEmptyAndValidQuests(t *testing.T) {
	for _, test := range []struct {
		name   string
		quests string
	}{
		{"omitted", ""},
		{"null_collection", `,"quests":null`},
		{"empty_collection", `,"quests":{}`},
		{"valid_entries", `,"quests":{"quest.active":{"status":"active","progress":0},"quest.completed":{"status":"completed","progress":1}}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := newServer(t.TempDir())
			data := []byte(`{"alice":{"name":"alice","hp":20,"room_id":"loc.start"` + test.quests + `}}`)
			if err := os.WriteFile(server.playersSavePath(), data, 0600); err != nil {
				t.Fatal(err)
			}
			players, err := server.loadPlayers()
			if err != nil {
				t.Fatalf("load valid player state: %v", err)
			}
			if test.name == "valid_entries" {
				active := players["alice"].Quests["quest.active"]
				completed := players["alice"].Quests["quest.completed"]
				if active == nil || active.Status != "active" || active.Progress != 0 || completed == nil || completed.Status != "completed" || completed.Progress != 1 {
					t.Fatalf("quest states changed during validation: %+v", players["alice"].Quests)
				}
			}
		})
	}
}

func TestConnectRejectsNullSavedQuestWithoutBlockingPeers(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = takeTestWorld()
	bob := startTestClient(t, server)
	bob.connect(t, "bob")
	data := []byte(`{"alice":{"name":"alice","hp":20,"room_id":"loc.start","quests":{"quest.phineus_harpies":null}},"bob":{"name":"bob","hp":100,"room_id":"loc.start"}}`)
	if err := os.WriteFile(server.playersSavePath(), data, 0600); err != nil {
		t.Fatal(err)
	}
	alice := startTestClient(t, server)
	alice.command(t, "CONNECT alice", "ERR 500 STATE_ERROR")
	alice.command(t, "WHO", "OK players=1")
	bob.command(t, "WHO", "OK players=1")
	server.mu.Lock()
	installed := server.players["alice"] != nil
	server.mu.Unlock()
	if installed {
		t.Fatal("malformed saved player was installed")
	}
	after, err := os.ReadFile(server.playersSavePath())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, data) {
		t.Fatal("failed login overwrote malformed saved player data")
	}
}

func TestQuestsInvalidStateDoesNotBlockOtherPlayers(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = takeTestWorld()
	server.players["alice"] = &Player{Name: "alice", Quests: map[string]*PlayerQuest{"quest.phineus_harpies": nil}}
	server.players["bob"] = &Player{Name: "bob"}
	serverConn, peer := net.Pipe()
	client := newServerClient(serverConn)
	defer client.Close()
	defer peer.Close()
	otherServerConn, otherPeer := net.Pipe()
	otherClient := newServerClient(otherServerConn)
	defer otherClient.Close()
	defer otherPeer.Close()
	if err := peer.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := otherPeer.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	done := make(chan any, 1)
	go func() {
		// A regression can panic while holding the lock. Isolate this handler
		// from handleClient's saving defer so the test reports it without hanging.
		defer func() {
			done <- recover()
			client.Close()
		}()
		name := "alice"
		handleQuests(server, client, &name, []string{"QUESTS"})
		name = "bob"
		handleWho(server, otherClient, &name, []string{"WHO"})
	}()
	for _, response := range []struct {
		conn net.Conn
		want string
	}{{peer, "ERR 500 STATE_ERROR\n"}, {otherPeer, "OK players=2\n"}} {
		line, err := bufio.NewReader(response.conn).ReadString('\n')
		if err != nil {
			select {
			case recovered := <-done:
				t.Fatalf("handler failed: panic=%v read=%v", recovered, err)
			default:
				t.Fatalf("handler did not respond: %v", err)
			}
		}
		if line != response.want {
			t.Fatalf("response = %q, want %q", line, response.want)
		}
	}
	if recovered := <-done; recovered != nil {
		t.Fatalf("handler panicked: %v", recovered)
	}
}

func TestConnectStartsRegenForSavedPlayer(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = takeTestWorld()
	if err := server.writePlayers(map[string]*Player{"alice": {Name: "alice", HP: 20, RoomID: "loc.start"}}); err != nil {
		t.Fatal(err)
	}
	for connection := range 2 {
		before := time.Now()
		if err := server.connectPlayer("alice"); err != nil {
			t.Fatal(err)
		}
		after := time.Now()
		server.mu.Lock()
		player := server.players["alice"]
		started := player.lastRegen
		if started.Before(before) || started.After(after) {
			server.mu.Unlock()
			t.Fatalf("connection %d: regen start = %v, want within [%v, %v]", connection, started, before, after)
		}
		initialHP := player.HP
		// Advance the clock from the connection's real baseline without sleeping
		// or issuing an update that could initialize a missing baseline.
		player.regenLocked(started.Add(regenInterval), 0, maxPlayerHP)
		gotHP := player.HP
		server.mu.Unlock()
		if want := initialHP + regenAmount; gotHP != want {
			t.Fatalf("connection %d: first update HP = %d, want %d", connection, gotHP, want)
		}
		if err := server.saveAndRemovePlayer("alice"); err != nil {
			t.Fatalf("disconnect %d: %v", connection, err)
		}
	}
}
