package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

type testClient struct {
	conn   net.Conn
	reader *bufio.Reader
	done   <-chan struct{}
}

func startTestClient(t *testing.T, server *Server) *testClient {
	t.Helper()
	serverConn, clientConn := net.Pipe()
	done := make(chan struct{})
	go func() {
		server.handleClient(serverConn)
		close(done)
	}()
	t.Cleanup(func() {
		clientConn.Close()
		<-done
	})
	if err := clientConn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	client := &testClient{conn: clientConn, reader: bufio.NewReader(clientConn), done: done}
	client.expect(t, "OK hello proto=1")
	return client
}

func (client *testClient) expect(t *testing.T, want string) {
	t.Helper()
	line, err := client.reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if got := strings.TrimSuffix(line, "\n"); got != want {
		t.Fatalf("response = %q, want %q", got, want)
	}
}

func (client *testClient) command(t *testing.T, command, want string) {
	t.Helper()
	if _, err := fmt.Fprintln(client.conn, command); err != nil {
		t.Fatalf("send %q: %v", command, err)
	}
	client.expect(t, want)
}

func TestQuitSavesAndConnectRestoresPlayer(t *testing.T) {
	saveDir := t.TempDir()
	server := newServer(saveDir)
	first := startTestClient(t, server)
	first.command(t, "CONNECT アリス", "OK connected")

	want := Player{
		Name:      "アリス",
		HP:        67,
		RoomID:    "loc.bakery",
		Inventory: []string{"item.herbs", "item.key"},
	}
	server.mu.Lock()
	*server.players[want.Name] = want
	server.mu.Unlock()
	first.command(t, "QUIT", "OK bye")
	<-first.done

	data, err := os.ReadFile(server.playersSavePath())
	if err != nil {
		t.Fatal(err)
	}
	var saved map[string]Player
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if len(saved) != 1 || !reflect.DeepEqual(saved[want.Name], want) {
		t.Fatalf("saved players = %+v, want %+v", saved, want)
	}

	second := startTestClient(t, server)
	second.command(t, "CONNECT アリス", "OK connected")
	server.mu.Lock()
	restored := *server.players[want.Name]
	server.mu.Unlock()
	if !reflect.DeepEqual(restored, want) {
		t.Fatalf("restored player = %+v, want %+v", restored, want)
	}

	duplicate := startTestClient(t, server)
	duplicate.command(t, "CONNECT アリス", "ERR 201 NAME_IN_USE")
	duplicate.command(t, "WHO", "OK players=1")
	second.command(t, "QUIT", "OK bye")
	<-second.done

	restarted := newServer(saveDir)
	third := startTestClient(t, restarted)
	third.command(t, "CONNECT アリス", "OK connected")
	restarted.mu.Lock()
	restored = *restarted.players[want.Name]
	restarted.mu.Unlock()
	if !reflect.DeepEqual(restored, want) {
		t.Fatalf("restored after restart = %+v, want %+v", restored, want)
	}
	third.command(t, "QUIT", "OK bye")
	<-third.done
	entries, err := os.ReadDir(saveDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != playersSaveFile {
		t.Fatalf("save files = %v, want one file for %q", entries, want.Name)
	}
}

func TestQuitPreservesOtherPlayersInOneFile(t *testing.T) {
	server := newServer(t.TempDir())
	players := []Player{
		{Name: "alice", HP: 45, RoomID: "loc.bakery"},
		{Name: "bob", HP: 80, RoomID: "loc.square", Inventory: []string{"item.key"}},
	}
	for _, want := range players {
		client := startTestClient(t, server)
		client.command(t, "CONNECT "+want.Name, "OK connected")
		server.mu.Lock()
		*server.players[want.Name] = want
		server.mu.Unlock()
		client.command(t, "QUIT", "OK bye")
		<-client.done
	}

	data, err := os.ReadFile(server.playersSavePath())
	if err != nil {
		t.Fatal(err)
	}
	var saved map[string]Player
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if len(saved) != len(players) {
		t.Fatalf("saved players = %d, want %d", len(saved), len(players))
	}
	for _, want := range players {
		if !reflect.DeepEqual(saved[want.Name], want) {
			t.Fatalf("saved %q = %+v, want %+v", want.Name, saved[want.Name], want)
		}
	}
	entries, err := os.ReadDir(server.saveDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != playersSaveFile {
		t.Fatalf("save files = %v, want %s only", entries, playersSaveFile)
	}
}

func TestQuitSaveFailureKeepsPlayerOnline(t *testing.T) {
	saveDir := t.TempDir()
	blocked := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(blocked, []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}
	server := newServer(saveDir)
	client := startTestClient(t, server)
	client.command(t, "CONNECT alice", "OK connected")

	server.mu.Lock()
	server.saveDir = blocked
	server.mu.Unlock()
	client.command(t, "QUIT", "ERR 500 STATE_ERROR")
	client.command(t, "WHO", "OK players=1")

	server.mu.Lock()
	server.saveDir = saveDir
	server.mu.Unlock()
	client.command(t, "QUIT", "OK bye")
	if _, err := os.Stat(server.playersSavePath()); err != nil {
		t.Fatal(err)
	}
}

func TestCorruptSaveDoesNotCreateNewPlayer(t *testing.T) {
	server := newServer(t.TempDir())
	if err := os.WriteFile(server.playersSavePath(), []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	client := startTestClient(t, server)
	client.command(t, "CONNECT alice", "ERR 500 STATE_ERROR")
	client.command(t, "WHO", "OK players=0")
	data, err := os.ReadFile(server.playersSavePath())
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "{" {
		t.Fatalf("corrupt save was overwritten: %q", data)
	}
}

func TestConcurrentConnectAllowsOnlyOnePlayerName(t *testing.T) {
	server := newServer(t.TempDir())
	results := make(chan error, 2)
	var group sync.WaitGroup
	for range 2 {
		group.Go(func() {
			results <- server.connectPlayer("alice")
		})
	}
	group.Wait()
	close(results)

	successes := 0
	duplicates := 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, errNameInUse):
			duplicates++
		default:
			t.Fatalf("unexpected connect error: %v", err)
		}
	}
	if successes != 1 || duplicates != 1 || len(server.players) != 1 {
		t.Fatalf("successes=%d duplicates=%d players=%d", successes, duplicates, len(server.players))
	}
}
