package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEmptyPlayerSaveAllowsStartupAndRegistration(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = takeTestWorld()
	if err := os.WriteFile(server.playersSavePath(), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := server.restoreItemOwnership(); err != nil {
		t.Fatalf("restore empty player save: %v", err)
	}
	info, err := os.Stat(server.playersSavePath())
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 0 {
		t.Fatalf("startup changed player save size to %d", info.Size())
	}

	client := startTestClient(t, server)
	client.connect(t, "alice")
	players, err := server.loadPlayers()
	if err != nil {
		t.Fatalf("load registered player: %v", err)
	}
	if players["alice"] == nil || players["alice"].Name != "alice" {
		t.Fatal("registered player was not saved")
	}
}

func TestConnectWithoutPasswordRegistrationAndReconnect(t *testing.T) {
	saveDir := t.TempDir()
	server := newServer(saveDir)
	first := startTestClient(t, server)
	first.command(t, "PASSWORD secret", "ERR 400 BAD_REQUEST")
	first.command(t, "CONNECT alice extra", "ERR 400 BAD_REQUEST")
	first.connect(t, "alice")
	first.command(t, "CONNECT bob", "ERR 400 BAD_REQUEST")

	data, err := os.ReadFile(server.playersSavePath())
	if err != nil {
		t.Fatal(err)
	}
	var saved map[string]map[string]json.RawMessage
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if _, exists := saved["alice"]["password"]; exists {
		t.Fatal("player state contains a password")
	}
	info, err := os.Stat(server.playersSavePath())
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("player state permissions = %o, want 600", info.Mode().Perm())
	}

	server.mu.Lock()
	server.players["alice"].HP = 73
	server.mu.Unlock()
	first.command(t, "QUIT", "OK bye")
	<-first.done

	second := startTestClient(t, server)
	second.connect(t, "alice")
	server.mu.Lock()
	hp := server.players["alice"].HP
	server.mu.Unlock()
	if hp != 73 {
		t.Fatalf("restored HP = %d, want 73", hp)
	}
	second.command(t, "QUIT", "OK bye")
	<-second.done

	restarted := newServer(saveDir)
	third := startTestClient(t, restarted)
	third.connect(t, "alice")
	third.command(t, "QUIT", "OK bye")
	<-third.done
}

func TestConnectIgnoresLegacyPasswordAndRemovesItOnSave(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = takeTestWorld()
	legacy := []byte(`{"alice":{"name":"alice","password":"old-password","hp":73,"room_id":"loc.start","inventory":[]}}`)
	if err := os.WriteFile(server.playersSavePath(), legacy, 0600); err != nil {
		t.Fatal(err)
	}
	client := startTestClient(t, server)
	client.connect(t, "alice")
	server.mu.Lock()
	hp := server.players["alice"].HP
	server.mu.Unlock()
	if hp != 73 {
		t.Fatalf("restored HP = %d, want 73", hp)
	}
	client.command(t, "QUIT", "OK bye")
	<-client.done

	data, err := os.ReadFile(server.playersSavePath())
	if err != nil {
		t.Fatal(err)
	}
	var saved map[string]map[string]json.RawMessage
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if _, exists := saved["alice"]["password"]; exists {
		t.Fatal("legacy password was not removed")
	}
}

func TestConnectRejectsInaccessibleSaveDirectory(t *testing.T) {
	blocked := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(blocked, []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}
	server := newServer(blocked)
	client := startTestClient(t, server)
	client.command(t, "CONNECT alice", "ERR 500 STATE_ERROR")
	server.mu.Lock()
	count := len(server.players)
	server.mu.Unlock()
	if count != 0 {
		t.Fatalf("players after failed login = %d, want 0", count)
	}
}
