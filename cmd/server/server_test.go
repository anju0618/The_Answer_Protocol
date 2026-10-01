package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	server *Server
	name   string
}

type failEventConn struct {
	net.Conn
	prefix string
}

func (conn failEventConn) Write(data []byte) (int, error) {
	if strings.HasPrefix(string(data), conn.prefix) {
		return 0, io.ErrClosedPipe
	}
	return conn.Conn.Write(data)
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
	client := &testClient{conn: clientConn, reader: bufio.NewReader(clientConn), done: done, server: server}
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

func (client *testClient) connect(t *testing.T, name string, peers ...*testClient) {
	t.Helper()
	client.command(t, "CONNECT "+name, "OK connected")
	client.name = name
	client.server.mu.Lock()
	count := len(client.server.players)
	roomID := client.server.players[name].RoomID
	sameRoom := make([]bool, len(peers))
	for i, peer := range peers {
		sameRoom[i] = client.server.players[peer.name].RoomID == roomID
	}
	client.server.mu.Unlock()
	stats := fmt.Sprintf("EVT STATS players=%d", count)
	client.expect(t, stats)
	for i, peer := range peers {
		if sameRoom[i] {
			peer.expect(t, "EVT ROOM PRESENCE ENTER "+name)
		}
		peer.expect(t, stats)
	}
}

func takeTestWorld() *World {
	return &World{
		StartRoomID: "loc.start",
		Rooms: map[string]*Room{
			"loc.start": {ID: "loc.start"},
			"loc.other": {ID: "loc.other"},
		},
		Items: map[string]*Item{
			"item.sword":  {Name: en("Rusty Sword"), RoomID: "loc.start", Obtainable: true},
			"item.shield": {Name: en("Silver Shield"), RoomID: "loc.start", Obtainable: true},
			"item.fixed":  {Name: en("Stone Statue"), RoomID: "loc.start"},
			"item.remote": {Name: en("Distant Key"), RoomID: "loc.other", Obtainable: true},
		},
	}
}

func TestLookShowsCurrentRoomState(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = takeTestWorld()
	server.world.Rooms["loc.start"].Name = en("Starting Room")
	server.world.Rooms["loc.start"].Description = en("A quiet room.")
	server.world.Rooms["loc.start"].Exits = map[string]string{"east": "loc.other"}
	server.world.NPCs = map[string]*NPC{
		"npc.guide":  {RoomID: "loc.start"},
		"npc.vendor": {RoomID: "loc.start"},
	}

	type lookResponse struct {
		Room    roomView `json:"room"`
		Players []string `json:"players"`
		Items   []string `json:"items"`
		NPCs    []string `json:"npcs"`
	}
	look := func(client *testClient) lookResponse {
		t.Helper()
		if _, err := fmt.Fprintln(client.conn, "LOOK"); err != nil {
			t.Fatal(err)
		}
		line, err := client.reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(line, "OK ") {
			t.Fatalf("LOOK response = %q", line)
		}
		var response lookResponse
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "OK ")), &response); err != nil {
			t.Fatalf("decode LOOK response %q: %v", line, err)
		}
		return response
	}

	alice := startTestClient(t, server)
	alice.command(t, "LOOK", "ERR 400 BAD_REQUEST")
	alice.connect(t, "alice")
	alice.command(t, "LOOK extra", "ERR 400 BAD_REQUEST")
	bob := startTestClient(t, server)
	bob.connect(t, "bob", alice)

	initial := look(alice)
	if initial.Room.ID != "loc.start" || initial.Room.Name != "Starting Room" || initial.Room.Description != "A quiet room." || !reflect.DeepEqual(initial.Room.Exits, map[string]string{"east": "loc.other"}) {
		t.Fatalf("initial room = %+v", initial.Room)
	}
	if !reflect.DeepEqual(initial.Players, []string{"alice", "bob"}) ||
		!reflect.DeepEqual(initial.Items, []string{"item.fixed", "item.shield", "item.sword"}) ||
		!reflect.DeepEqual(initial.NPCs, []string{"npc.guide", "npc.vendor"}) {
		t.Fatalf("initial LOOK = %+v", initial)
	}

	alice.command(t, "TAKE item.sword", "OK taken=item.sword")
	afterTake := look(bob)
	if !reflect.DeepEqual(afterTake.Items, []string{"item.fixed", "item.shield"}) {
		t.Fatalf("items after TAKE = %v", afterTake.Items)
	}

	alice.command(t, "MOVE east", "OK room=loc.other")
	alice.expect(t, "EVT ROOM PRESENCE ENTER alice")
	bob.expect(t, "EVT ROOM PRESENCE LEAVE alice")
	afterMove := look(alice)
	if afterMove.Room.ID != "loc.other" || !reflect.DeepEqual(afterMove.Room.Exits, map[string]string{}) ||
		!reflect.DeepEqual(afterMove.Players, []string{"alice"}) ||
		!reflect.DeepEqual(afterMove.Items, []string{"item.remote"}) ||
		!reflect.DeepEqual(afterMove.NPCs, []string{}) {
		t.Fatalf("LOOK after MOVE = %+v", afterMove)
	}
	alice.command(t, "TAKE item.remote", "OK taken=item.remote")
	if got := look(alice).Items; !reflect.DeepEqual(got, []string{}) {
		t.Fatalf("items after taking last item = %v", got)
	}
	if got := look(bob).Players; !reflect.DeepEqual(got, []string{"bob"}) {
		t.Fatalf("players after MOVE = %v", got)
	}
}

func TestTakeMovesItemToInventoryAndRejectsUnavailableItems(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = takeTestWorld()
	alice := startTestClient(t, server)
	alice.command(t, "TAKE Rusty Sword", "ERR 400 BAD_REQUEST")
	alice.connect(t, "alice")
	alice.command(t, "TAKE", "ERR 400 BAD_REQUEST")
	alice.command(t, "TAKE missing", "ERR 404 ITEM_NOT_FOUND")
	alice.command(t, "TAKE item.fixed", "ERR 404 ITEM_NOT_FOUND")
	alice.command(t, "TAKE item.remote", "ERR 404 ITEM_NOT_FOUND")
	alice.command(t, "TAKE   rUsTy   SwOrD", "OK taken=item.sword")

	server.mu.Lock()
	inventory := append([]string(nil), server.players["alice"].Inventory...)
	roomID := server.world.Items["item.sword"].RoomID
	server.mu.Unlock()
	if !reflect.DeepEqual(inventory, []string{"item.sword"}) || roomID != "" {
		t.Fatalf("inventory=%v item room=%q, want [item.sword] and no room", inventory, roomID)
	}

	bob := startTestClient(t, server)
	bob.connect(t, "bob", alice)
	bob.command(t, "TAKE item.sword", "ERR 404 ITEM_NOT_FOUND")
	alice.command(t, "TAKE item.sword", "ERR 404 ITEM_NOT_FOUND")
	bob.command(t, "TAKE item.shield", "OK taken=item.shield")
	server.mu.Lock()
	bobInventory := append([]string(nil), server.players["bob"].Inventory...)
	server.mu.Unlock()
	if !reflect.DeepEqual(bobInventory, []string{"item.shield"}) {
		t.Fatalf("bob inventory=%v, want [item.shield]", bobInventory)
	}
}

func TestTakeRestoresItemOwnershipAfterRestart(t *testing.T) {
	saveDir := t.TempDir()
	server := newServer(saveDir)
	server.world = takeTestWorld()
	alice := startTestClient(t, server)
	alice.connect(t, "alice")
	alice.command(t, "TAKE item.sword", "OK taken=item.sword")
	alice.command(t, "QUIT", "OK bye")
	<-alice.done

	restarted := newServer(saveDir)
	restarted.world = takeTestWorld()
	if err := restarted.restoreItemOwnership(); err != nil {
		t.Fatal(err)
	}
	if got := restarted.world.Items["item.sword"].RoomID; got != "" {
		t.Fatalf("restored item room=%q, want no room", got)
	}
	bob := startTestClient(t, restarted)
	bob.connect(t, "bob")
	bob.command(t, "TAKE item.sword", "ERR 404 ITEM_NOT_FOUND")
	reconnected := startTestClient(t, restarted)
	reconnected.connect(t, "alice", bob)
	restarted.mu.Lock()
	inventory := append([]string(nil), restarted.players["alice"].Inventory...)
	restarted.mu.Unlock()
	if !reflect.DeepEqual(inventory, []string{"item.sword"}) {
		t.Fatalf("restored inventory=%v, want [item.sword]", inventory)
	}
}

func TestConcurrentTakeOnlyOnePlayerGetsItem(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = takeTestWorld()
	alice := startTestClient(t, server)
	alice.connect(t, "alice")
	bob := startTestClient(t, server)
	bob.connect(t, "bob", alice)

	results := make(chan string, 2)
	for _, client := range []*testClient{alice, bob} {
		go func() {
			if _, err := fmt.Fprintln(client.conn, "TAKE item.sword"); err != nil {
				results <- err.Error()
				return
			}
			line, err := client.reader.ReadString('\n')
			if err != nil {
				results <- err.Error()
				return
			}
			results <- strings.TrimSuffix(line, "\n")
		}()
	}
	counts := map[string]int{}
	for range 2 {
		counts[<-results]++
	}
	if counts["OK taken=item.sword"] != 1 || counts["ERR 404 ITEM_NOT_FOUND"] != 1 {
		t.Fatalf("concurrent TAKE responses=%v", counts)
	}
	server.mu.Lock()
	total := len(server.players["alice"].Inventory) + len(server.players["bob"].Inventory)
	server.mu.Unlock()
	if total != 1 {
		t.Fatalf("inventory instances=%d, want 1", total)
	}
}

func TestMoveUpdatesRoomAndRejectsInvalidDirections(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = &World{
		StartRoomID: "loc.start",
		Rooms: map[string]*Room{
			"loc.start": {ID: "loc.start", Exits: map[string]string{"north": "loc.next"}},
			"loc.next":  {ID: "loc.next", Exits: map[string]string{"south": "loc.start"}},
		},
	}
	client := startTestClient(t, server)
	client.command(t, "MOVE north", "ERR 400 BAD_REQUEST")
	client.connect(t, "alice")
	client.command(t, "MOVE", "ERR 400 BAD_REQUEST")
	client.command(t, "MOVE north extra", "ERR 400 BAD_REQUEST")
	client.command(t, "MOVE west", "ERR 301 NO_EXIT")

	server.mu.Lock()
	roomID := server.players["alice"].RoomID
	server.mu.Unlock()
	if roomID != "loc.start" {
		t.Fatalf("room after invalid moves = %q, want loc.start", roomID)
	}

	client.command(t, "MOVE NORTH", "OK room=loc.next")
	client.expect(t, "EVT ROOM PRESENCE ENTER alice")
	server.mu.Lock()
	roomID = server.players["alice"].RoomID
	server.mu.Unlock()
	if roomID != "loc.next" {
		t.Fatalf("room after moving north = %q, want loc.next", roomID)
	}
	client.command(t, "QUIT", "OK bye")
	<-client.done

	reconnected := startTestClient(t, server)
	reconnected.connect(t, "alice")
	server.mu.Lock()
	roomID = server.players["alice"].RoomID
	server.mu.Unlock()
	if roomID != "loc.next" {
		t.Fatalf("restored room = %q, want loc.next", roomID)
	}
	reconnected.command(t, "MOVE south", "OK room=loc.start")
	reconnected.expect(t, "EVT ROOM PRESENCE ENTER alice")
	reconnected.command(t, "QUIT", "OK bye")
}

func TestMovePresenceEventsReachOnlyRelevantRooms(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = &World{
		StartRoomID: "loc.old",
		Rooms: map[string]*Room{
			"loc.old":   {ID: "loc.old", Exits: map[string]string{"north": "loc.new"}},
			"loc.new":   {ID: "loc.new", Exits: map[string]string{"south": "loc.old"}},
			"loc.other": {ID: "loc.other"},
		},
	}
	alice := startTestClient(t, server)
	alice.connect(t, "alice")
	bob := startTestClient(t, server)
	bob.connect(t, "bob", alice)
	carol := startTestClient(t, server)
	carol.connect(t, "carol", alice, bob)
	dave := startTestClient(t, server)
	dave.connect(t, "dave", alice, bob, carol)

	server.mu.Lock()
	server.players["carol"].RoomID = "loc.new"
	server.players["dave"].RoomID = "loc.other"
	server.mu.Unlock()

	alice.command(t, "MOVE north", "OK room=loc.new")
	alice.expect(t, "EVT ROOM PRESENCE ENTER alice")
	bob.expect(t, "EVT ROOM PRESENCE LEAVE alice")
	carol.expect(t, "EVT ROOM PRESENCE ENTER alice")
	dave.command(t, "WHO", "OK players=4")
	bob.command(t, "WHO", "OK players=4")
	carol.command(t, "WHO", "OK players=4")

	alice.command(t, "MOVE west", "ERR 301 NO_EXIT")
	alice.command(t, "WHO", "OK players=4")
	bob.command(t, "WHO", "OK players=4")
	carol.command(t, "WHO", "OK players=4")
	dave.command(t, "WHO", "OK players=4")
}

func TestMoveBroadcastContinuesAfterRecipientWriteFailure(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = &World{
		StartRoomID: "loc.old",
		Rooms: map[string]*Room{
			"loc.old": {ID: "loc.old", Exits: map[string]string{"north": "loc.new"}},
			"loc.new": {ID: "loc.new", Exits: map[string]string{"south": "loc.old"}},
		},
	}
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	serverConn, clientConn := net.Pipe()
	bobDone := make(chan struct{})
	go func() {
		server.handleClient(failEventConn{serverConn, "EVT ROOM PRESENCE LEAVE alice"})
		close(bobDone)
	}()
	t.Cleanup(func() {
		clientConn.Close()
		<-bobDone
	})
	if err := clientConn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	bob := &testClient{conn: clientConn, reader: bufio.NewReader(clientConn), done: bobDone, server: server}
	bob.expect(t, "OK hello proto=1")
	bob.connect(t, "bob", alice)

	carol := startTestClient(t, server)
	carol.connect(t, "carol", alice, bob)
	server.mu.Lock()
	server.players["carol"].RoomID = "loc.new"
	server.mu.Unlock()

	alice.command(t, "MOVE north", "OK room=loc.new")
	alice.expect(t, "EVT ROOM PRESENCE ENTER alice")
	carol.expect(t, "EVT ROOM PRESENCE ENTER alice")
	select {
	case <-bobDone:
	case <-time.After(5 * time.Second):
		t.Fatal("failed event recipient stayed connected")
	}
	alice.expect(t, "EVT STATS players=2")
	carol.expect(t, "EVT STATS players=2")
	alice.command(t, "WHO", "OK players=2")
}

func TestMoveQueuesPresenceBeforeWaitingForResponse(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = takeTestWorld()
	server.world.Rooms["loc.start"].Exits = map[string]string{"east": "loc.other"}
	alice := startTestClient(t, server)
	alice.connect(t, "alice")
	bob := startTestClient(t, server)
	bob.connect(t, "bob", alice)
	if _, err := fmt.Fprintln(alice.conn, "MOVE east"); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for {
		server.mu.Lock()
		moved := server.players["alice"].RoomID == "loc.other"
		server.mu.Unlock()
		if moved {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("alice did not move")
		}
		time.Sleep(time.Millisecond)
	}
	if _, err := fmt.Fprintln(bob.conn, "MOVE east"); err != nil {
		t.Fatal(err)
	}
	bob.expect(t, "EVT ROOM PRESENCE LEAVE alice")
	bob.expect(t, "OK room=loc.other")
	bob.expect(t, "EVT ROOM PRESENCE ENTER bob")
	if _, err := fmt.Fprintln(bob.conn, "LOOK"); err != nil {
		t.Fatal(err)
	}
	line, err := bob.reader.ReadString('\n')
	if err != nil || !strings.Contains(line, `"players":["alice","bob"]`) {
		t.Fatalf("LOOK = %q, err = %v", line, err)
	}
	alice.expect(t, "OK room=loc.other")
	alice.expect(t, "EVT ROOM PRESENCE ENTER alice")
	alice.expect(t, "EVT ROOM PRESENCE ENTER bob")
	bob.command(t, "WHO", "OK players=2")
}

func TestConnectAndDisconnectPublishPresenceAndCounts(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = takeTestWorld()
	server.world.Rooms["loc.start"].Exits = map[string]string{"east": "loc.other"}
	alice := startTestClient(t, server)
	alice.connect(t, "alice")
	carol := startTestClient(t, server)
	carol.connect(t, "carol", alice)
	carol.command(t, "MOVE east", "OK room=loc.other")
	carol.expect(t, "EVT ROOM PRESENCE ENTER carol")
	alice.expect(t, "EVT ROOM PRESENCE LEAVE carol")
	unconnected := startTestClient(t, server)
	bob := startTestClient(t, server)
	bob.command(t, "CONNECT bob", "OK connected")
	bob.expect(t, "EVT STATS players=3")
	alice.expect(t, "EVT ROOM PRESENCE ENTER bob")
	alice.expect(t, "EVT STATS players=3")
	carol.expect(t, "EVT STATS players=3")
	carol.command(t, "WHO", "OK players=3")
	unconnected.command(t, "CONNECT bob", "ERR 201 NAME_IN_USE")
	unconnected.command(t, "WHO", "OK players=3")
	alice.command(t, "WHO", "OK players=3")
	if err := bob.conn.Close(); err != nil {
		t.Fatal(err)
	}
	<-bob.done
	alice.expect(t, "EVT ROOM PRESENCE LEAVE bob")
	alice.expect(t, "EVT STATS players=2")
	carol.expect(t, "EVT STATS players=2")
	carol.command(t, "WHO", "OK players=2")
	reconnected := startTestClient(t, server)
	reconnected.connect(t, "bob", alice, carol)
	reconnected.command(t, "QUIT", "OK bye")
	<-reconnected.done
	alice.expect(t, "EVT ROOM PRESENCE LEAVE bob")
	alice.expect(t, "EVT STATS players=2")
	carol.expect(t, "EVT STATS players=2")
	carol.command(t, "WHO", "OK players=2")
}

func TestConnectRejectsNameThatWouldOverflowPresence(t *testing.T) {
	server := newServer(t.TempDir())
	alice := startTestClient(t, server)
	alice.connect(t, "alice")
	other := startTestClient(t, server)
	name := strings.Repeat("x", maxProtocolLineBytes-len("EVT ROOM PRESENCE ENTER ")+1)
	other.command(t, "CONNECT "+name, "ERR 400 BAD_REQUEST")
	alice.command(t, "WHO", "OK players=1")
	other.connect(t, "bob", alice)
}

func TestLookRejectsOversizedResponseWithoutDisconnecting(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = takeTestWorld()
	server.world.Rooms["loc.start"].Description = en(strings.Repeat("x", maxProtocolLineBytes))
	client := startTestClient(t, server)
	client.connect(t, "alice")
	client.command(t, "LOOK", "ERR 500 STATE_ERROR")
	client.command(t, "WHO", "OK players=1")
}

func TestQuitBroadcastsRoomLeave(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = takeTestWorld()
	alice := startTestClient(t, server)
	alice.connect(t, "alice")
	bob := startTestClient(t, server)
	bob.connect(t, "bob", alice)
	carol := startTestClient(t, server)
	carol.connect(t, "carol", alice, bob)
	server.mu.Lock()
	server.players["carol"].RoomID = "loc.other"
	server.mu.Unlock()

	alice.command(t, "QUIT", "OK bye")
	<-alice.done
	bob.expect(t, "EVT ROOM PRESENCE LEAVE alice")
	bob.expect(t, "EVT STATS players=2")
	carol.expect(t, "EVT STATS players=2")
	bob.command(t, "WHO", "OK players=2")
	carol.command(t, "WHO", "OK players=2")
}

func TestQuitSavesAndConnectRestoresPlayer(t *testing.T) {
	saveDir := t.TempDir()
	server := newServer(saveDir)
	first := startTestClient(t, server)
	first.connect(t, "アリス")

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
	second.connect(t, "アリス")
	server.mu.Lock()
	restored := *server.players[want.Name]
	server.mu.Unlock()
	restored.lastRegen = time.Time{}

	wantRestored := want
	wantRestored.IntroSeen = true
	if !reflect.DeepEqual(restored, wantRestored) {
		t.Fatalf("restored player = %+v, want %+v", restored, wantRestored)
	}

	duplicate := startTestClient(t, server)
	duplicate.command(t, "CONNECT アリス", "ERR 201 NAME_IN_USE")
	duplicate.command(t, "WHO", "OK players=1")
	second.command(t, "QUIT", "OK bye")
	<-second.done

	restarted := newServer(saveDir)
	third := startTestClient(t, restarted)
	third.connect(t, "アリス")
	restarted.mu.Lock()
	restored = *restarted.players[want.Name]
	restarted.mu.Unlock()
	restored.lastRegen = time.Time{}
	if !reflect.DeepEqual(restored, wantRestored) {
		t.Fatalf("restored after restart = %+v, want %+v", restored, wantRestored)
	}
	third.command(t, "QUIT", "OK bye")
	<-third.done
	entries, err := os.ReadDir(saveDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != playersSaveFile {
		t.Fatalf("save files = %v, want %s only", entries, playersSaveFile)
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
		client.connect(t, want.Name)
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

func TestSaveAndRemovePlayerUsesSnapshotAndRejectsUpdates(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = takeTestWorld()
	server.world.Rooms["loc.start"].Exits = map[string]string{"east": "loc.other"}
	if err := server.connectPlayer("alice"); err != nil {
		t.Fatal(err)
	}
	server.mu.Lock()
	player := server.players["alice"]
	player.Inventory = []string{"item.sword"}
	server.world.Items["item.sword"].RoomID = ""
	server.mu.Unlock()

	server.ioMu.Lock()
	ioLocked := true
	defer func() {
		if ioLocked {
			server.ioMu.Unlock()
		}
	}()
	saved := make(chan error, 1)
	go func() {
		saved <- server.saveAndRemovePlayer("alice")
	}()

	deadline := time.After(2 * time.Second)
	for {
		server.mu.Lock()
		exiting := player.exiting
		server.mu.Unlock()
		if exiting {
			break
		}
		select {
		case err := <-saved:
			t.Fatalf("save finished before I/O was released: %v", err)
		case <-deadline:
			t.Fatal("player never entered exiting state")
		default:
			time.Sleep(time.Millisecond)
		}
	}

	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()
	if err := clientConn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	commandsDone := make(chan struct{})
	go func() {
		name := "alice"
		handleMove(server, serverConn, &name, []string{"MOVE", "east"})
		handleTake(server, serverConn, &name, []string{"TAKE", "item.shield"})
		close(commandsDone)
	}()
	reader := bufio.NewReader(clientConn)
	for range 2 {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if line != "ERR 500 STATE_ERROR\n" {
			t.Fatalf("response during save = %q", line)
		}
	}
	<-commandsDone

	server.mu.Lock()
	if player.RoomID != "loc.start" || len(player.Inventory) != 1 {
		server.mu.Unlock()
		t.Fatal("player changed while exiting")
	}
	player.Inventory[0] = "item.changed"
	server.mu.Unlock()
	server.ioMu.Unlock()
	ioLocked = false
	if err := <-saved; err != nil {
		t.Fatal(err)
	}

	players, err := server.loadPlayers()
	if err != nil {
		t.Fatal(err)
	}
	if got := players["alice"].Inventory; !reflect.DeepEqual(got, []string{"item.sword"}) {
		t.Fatalf("saved inventory = %v, want snapshot", got)
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
	client.connect(t, "alice")

	server.mu.Lock()
	server.saveDir = blocked
	server.mu.Unlock()
	client.command(t, "QUIT", "ERR 500 STATE_ERROR")
	client.command(t, "WHO", "OK players=1")
	server.mu.Lock()
	exiting := server.players["alice"].exiting
	server.mu.Unlock()
	if exiting {
		t.Fatal("player stayed in exiting state after failed save")
	}

	server.mu.Lock()
	server.saveDir = saveDir
	server.mu.Unlock()
	client.command(t, "QUIT", "OK bye")
	if _, err := os.Stat(server.playersSavePath()); err != nil {
		t.Fatal(err)
	}
}

func TestDisconnectSaveFailureRemovesPlayer(t *testing.T) {
	saveDir := t.TempDir()
	blocked := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(blocked, []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}
	server := newServer(saveDir)
	client := startTestClient(t, server)
	client.connect(t, "alice")

	server.mu.Lock()
	server.saveDir = blocked
	server.mu.Unlock()
	if err := client.conn.Close(); err != nil {
		t.Fatal(err)
	}
	<-client.done

	server.mu.Lock()
	remaining := len(server.players)
	server.saveDir = saveDir
	server.mu.Unlock()
	if remaining != 0 {
		t.Fatalf("online players after failed disconnect save = %d, want 0", remaining)
	}

	reconnected := startTestClient(t, server)
	reconnected.connect(t, "alice")
	reconnected.command(t, "QUIT", "OK bye")
}

func TestDisconnectSaveFailureRestoresNewlyTakenItems(t *testing.T) {
	saveDir := t.TempDir()
	server := newServer(saveDir)
	server.world = takeTestWorld()
	if err := server.writePlayers(map[string]*Player{
		"alice": {Name: "alice", HP: 100, RoomID: "loc.start", Inventory: []string{"item.sword"}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := server.restoreItemOwnership(); err != nil {
		t.Fatal(err)
	}
	alice := startTestClient(t, server)
	alice.connect(t, "alice")
	bob := startTestClient(t, server)
	bob.connect(t, "bob", alice)
	alice.command(t, "TAKE item.shield", "OK taken=item.shield")

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

	server.mu.Lock()
	shieldRoom := server.world.Items["item.shield"].RoomID
	swordRoom := server.world.Items["item.sword"].RoomID
	server.saveDir = saveDir
	server.mu.Unlock()
	if shieldRoom != "loc.start" || swordRoom != "" {
		t.Fatalf("item rooms after failed save: shield=%q sword=%q", shieldRoom, swordRoom)
	}
	bob.command(t, "TAKE item.shield", "OK taken=item.shield")
	bob.command(t, "TAKE item.sword", "ERR 404 ITEM_NOT_FOUND")

	reconnected := startTestClient(t, server)
	reconnected.connect(t, "alice", bob)
	server.mu.Lock()
	inventory := append([]string(nil), server.players["alice"].Inventory...)
	server.mu.Unlock()
	if !reflect.DeepEqual(inventory, []string{"item.sword"}) {
		t.Fatalf("restored inventory = %v, want saved sword only", inventory)
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
	server.mu.Lock()
	count := len(server.players)
	server.mu.Unlock()
	if count != 0 {
		t.Fatalf("players after corrupt save = %d, want 0", count)
	}
	data, err := os.ReadFile(server.playersSavePath())
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "{" {
		t.Fatalf("corrupt save was overwritten: %q", data)
	}
}

func TestConnectRejectsSavedPlayerInRemovedRoom(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = takeTestWorld()
	if err := server.writePlayers(map[string]*Player{
		"alice": {Name: "alice", HP: 75, RoomID: "loc.removed"},
		"bob":   {Name: "bob", HP: 80, RoomID: "loc.other"},
	}); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(server.playersSavePath())
	if err != nil {
		t.Fatal(err)
	}

	stale := startTestClient(t, server)
	stale.command(t, "CONNECT alice", "ERR 500 STATE_ERROR")
	stale.command(t, "WHO", "OK players=0")
	server.mu.Lock()
	count := len(server.players)
	server.mu.Unlock()
	if count != 0 {
		t.Fatalf("players after invalid room = %d, want 0", count)
	}
	after, err := os.ReadFile(server.playersSavePath())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(after, before) {
		t.Fatal("invalid saved player was overwritten")
	}

	valid := startTestClient(t, server)
	valid.connect(t, "bob")
	server.mu.Lock()
	roomID := server.players["bob"].RoomID
	server.mu.Unlock()
	if roomID != "loc.other" {
		t.Fatalf("restored room = %q, want loc.other", roomID)
	}
	valid.command(t, "QUIT", "OK bye")
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
