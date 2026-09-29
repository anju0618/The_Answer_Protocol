package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGroupCreateInviteJoinLeave(t *testing.T) {
	server := newServer(t.TempDir())
	unconnected := startTestClient(t, server)
	unconnected.command(t, "GROUP CREATE", "ERR 400 BAD_REQUEST")

	alice := startTestClient(t, server)
	alice.connect(t, "alice")
	bob := startTestClient(t, server)
	bob.connect(t, "bob", alice)
	carol := startTestClient(t, server)
	carol.connect(t, "carol", alice, bob)

	alice.command(t, "GROUP", "ERR 400 BAD_REQUEST")
	alice.command(t, "GROUP INVITE bob", "ERR 401 NOT_IN_GROUP")
	bob.command(t, "GROUP LEAVE", "ERR 401 NOT_IN_GROUP")
	bob.command(t, "GROUP JOIN alice", "ERR 400 BAD_REQUEST")
	alice.command(t, "GROUP CREATE extra", "ERR 400 BAD_REQUEST")
	alice.command(t, "GROUP CREATE", "OK group=group.1")
	alice.command(t, "GROUP CREATE", "ERR 402 ALREADY_IN_GROUP")
	alice.command(t, "GROUP INVITE missing", "ERR 400 BAD_REQUEST")
	alice.command(t, "GROUP INVITE bob", "OK")
	bob.expect(t, "EVT GROUP INVITE alice")
	bob.command(t, "GROUP JOIN group.1", "ERR 400 BAD_REQUEST")
	bob.command(t, "GROUP JOIN alice", "OK group=group.1")
	bob.expect(t, "EVT GROUP JOIN bob")
	alice.expect(t, "EVT GROUP JOIN bob")
	alice.command(t, "GROUP INVITE bob", "ERR 402 ALREADY_IN_GROUP")
	bob.command(t, "GROUP INVITE carol", "OK")
	carol.expect(t, "EVT GROUP INVITE alice")
	carol.command(t, "GROUP JOIN alice", "OK group=group.1")
	carol.expect(t, "EVT GROUP JOIN carol")
	alice.expect(t, "EVT GROUP JOIN carol")
	bob.expect(t, "EVT GROUP JOIN carol")

	carol.command(t, "GROUP LEAVE", "OK")
	alice.expect(t, "EVT GROUP LEAVE carol")
	bob.expect(t, "EVT GROUP LEAVE carol")
	carol.command(t, "GROUP LEAVE", "ERR 401 NOT_IN_GROUP")
	carol.command(t, "GROUP CREATE", "OK group=group.2")
	alice.command(t, "GROUP INVITE carol", "ERR 402 ALREADY_IN_GROUP")

	bob.command(t, "GROUP LEAVE", "OK")
	alice.expect(t, "EVT GROUP LEAVE bob")
	alice.command(t, "GROUP LEAVE", "OK")
	server.mu.Lock()
	remaining := len(server.groups)
	server.mu.Unlock()
	if remaining != 1 {
		t.Fatalf("groups after last member left = %d, want 1", remaining)
	}
}

func TestGroupLeaderDisconnectTransfersLeadership(t *testing.T) {
	server := newServer(t.TempDir())
	alice := startTestClient(t, server)
	alice.connect(t, "alice")
	bob := startTestClient(t, server)
	bob.connect(t, "bob", alice)
	carol := startTestClient(t, server)
	carol.connect(t, "carol", alice, bob)

	alice.command(t, "GROUP CREATE", "OK group=group.1")
	alice.command(t, "GROUP INVITE bob", "OK")
	bob.expect(t, "EVT GROUP INVITE alice")
	bob.command(t, "GROUP JOIN alice", "OK group=group.1")
	bob.expect(t, "EVT GROUP JOIN bob")
	alice.expect(t, "EVT GROUP JOIN bob")
	alice.command(t, "GROUP INVITE carol", "OK")
	carol.expect(t, "EVT GROUP INVITE alice")
	if err := alice.conn.Close(); err != nil {
		t.Fatal(err)
	}
	<-alice.done
	bob.expect(t, "EVT GROUP LEAVE alice")
	bob.expect(t, "EVT ROOM PRESENCE LEAVE alice")
	bob.expect(t, "EVT STATS players=2")
	carol.expect(t, "EVT ROOM PRESENCE LEAVE alice")
	carol.expect(t, "EVT STATS players=2")
	carol.command(t, "GROUP JOIN alice", "ERR 400 BAD_REQUEST")
	bob.command(t, "GROUP INVITE carol", "OK")
	carol.expect(t, "EVT GROUP INVITE bob")
	carol.command(t, "GROUP JOIN bob", "OK group=group.1")
	carol.expect(t, "EVT GROUP JOIN carol")
	bob.expect(t, "EVT GROUP JOIN carol")

	server.mu.Lock()
	leader := server.groups["group.1"].Leader
	server.mu.Unlock()
	if leader != "bob" {
		t.Fatalf("new group leader = %q, want bob", leader)
	}
}

func TestChatScopesAndMessageText(t *testing.T) {
	server := newServer(t.TempDir())
	unconnected := startTestClient(t, server)
	unconnected.command(t, "CHAT GLOBAL hi", "ERR 400 BAD_REQUEST")
	alice := startTestClient(t, server)
	alice.connect(t, "alice")
	bob := startTestClient(t, server)
	bob.connect(t, "bob", alice)
	carol := startTestClient(t, server)
	carol.connect(t, "carol", alice, bob)
	server.mu.Lock()
	server.players["carol"].RoomID = "loc.other"
	server.mu.Unlock()
	alice.command(t, "GROUP CREATE", "OK group=group.1")
	alice.command(t, "GROUP INVITE carol", "OK")
	carol.expect(t, "EVT GROUP INVITE alice")
	carol.command(t, "GROUP JOIN alice", "OK group=group.1")
	carol.expect(t, "EVT GROUP JOIN carol")
	alice.expect(t, "EVT GROUP JOIN carol")

	alice.command(t, "CHAT", "ERR 400 BAD_REQUEST")
	alice.command(t, "CHAT GLOBAL", "ERR 400 BAD_REQUEST")
	alice.command(t, "CHAT UNKNOWN hi", "ERR 400 BAD_REQUEST")
	alice.command(t, "CHAT GLOBAL bad\x1btext", "ERR 400 BAD_REQUEST")
	alice.command(t, "CHAT GLOBAL bad\xfftext", "ERR 400 BAD_REQUEST")
	bob.command(t, "CHAT GROUP secret", "ERR 401 NOT_IN_GROUP")

	alice.command(t, "CHAT global Hello   world", "OK")
	alice.expect(t, "EVT GLOBAL CHAT alice Hello   world")
	bob.expect(t, "EVT GLOBAL CHAT alice Hello   world")
	carol.expect(t, "EVT GLOBAL CHAT alice Hello   world")

	alice.command(t, "CHAT ROOM nearby", "OK")
	alice.expect(t, "EVT ROOM CHAT alice nearby")
	bob.expect(t, "EVT ROOM CHAT alice nearby")
	carol.command(t, "WHO", "OK players=3")

	alice.command(t, "CHAT GROUP distant", "OK")
	alice.expect(t, "EVT GROUP CHAT alice distant")
	carol.expect(t, "EVT GROUP CHAT alice distant")
	bob.command(t, "WHO", "OK players=3")

	carol.command(t, "CHAT ROOM elsewhere", "OK")
	carol.expect(t, "EVT ROOM CHAT carol elsewhere")
	alice.command(t, "WHO", "OK players=3")
	bob.command(t, "WHO", "OK players=3")

	if _, err := fmt.Fprint(alice.conn, "CHAT GLOBAL Hello  again\r\n"); err != nil {
		t.Fatal(err)
	}
	alice.expect(t, "OK")
	alice.expect(t, "EVT GLOBAL CHAT alice Hello  again")
	bob.expect(t, "EVT GLOBAL CHAT alice Hello  again")
	carol.expect(t, "EVT GLOBAL CHAT alice Hello  again")
}

func TestChatBroadcastContinuesAfterRecipientFailure(t *testing.T) {
	server := newServer(t.TempDir())
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	serverConn, clientConn := net.Pipe()
	bobDone := make(chan struct{})
	go func() {
		server.handleClient(failEventConn{serverConn, "EVT GLOBAL CHAT "})
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

	alice.command(t, "CHAT GLOBAL hi", "OK")
	alice.expect(t, "EVT GLOBAL CHAT alice hi")
	carol.expect(t, "EVT GLOBAL CHAT alice hi")
	select {
	case <-bobDone:
	case <-time.After(5 * time.Second):
		t.Fatal("failed event recipient stayed connected")
	}
	alice.expect(t, "EVT ROOM PRESENCE LEAVE bob")
	alice.expect(t, "EVT STATS players=2")
	carol.expect(t, "EVT ROOM PRESENCE LEAVE bob")
	carol.expect(t, "EVT STATS players=2")
	alice.command(t, "WHO", "OK players=2")
}

func TestGroupQuitSaveFailureKeepsMembership(t *testing.T) {
	saveDir := t.TempDir()
	blocked := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(blocked, []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}
	server := newServer(saveDir)
	alice := startTestClient(t, server)
	alice.connect(t, "alice")
	bob := startTestClient(t, server)
	bob.connect(t, "bob", alice)
	alice.command(t, "GROUP CREATE", "OK group=group.1")
	alice.command(t, "GROUP INVITE bob", "OK")
	bob.expect(t, "EVT GROUP INVITE alice")
	bob.command(t, "GROUP JOIN alice", "OK group=group.1")
	bob.expect(t, "EVT GROUP JOIN bob")
	alice.expect(t, "EVT GROUP JOIN bob")

	server.mu.Lock()
	server.saveDir = blocked
	server.mu.Unlock()
	alice.command(t, "QUIT", "ERR 500 STATE_ERROR")
	bob.command(t, "CHAT GROUP still here", "OK")
	bob.expect(t, "EVT GROUP CHAT bob still here")
	alice.expect(t, "EVT GROUP CHAT bob still here")
	server.mu.Lock()
	server.saveDir = saveDir
	server.mu.Unlock()
	alice.command(t, "QUIT", "OK bye")
	<-alice.done
	bob.expect(t, "EVT GROUP LEAVE alice")
	bob.expect(t, "EVT ROOM PRESENCE LEAVE alice")
	bob.expect(t, "EVT STATS players=1")
}

func TestChatChecksEncodedEventLength(t *testing.T) {
	for _, scope := range []string{"GLOBAL", "ROOM", "GROUP"} {
		t.Run(scope, func(t *testing.T) {
			server := newServer(t.TempDir())
			alice := startTestClient(t, server)
			alice.connect(t, "アリス")
			bob := startTestClient(t, server)
			bob.connect(t, "bob", alice)
			if scope == "GROUP" {
				alice.command(t, "GROUP CREATE", "OK group=group.1")
				alice.command(t, "GROUP INVITE bob", "OK")
				bob.expect(t, "EVT GROUP INVITE アリス")
				bob.command(t, "GROUP JOIN アリス", "OK group=group.1")
				alice.expect(t, "EVT GROUP JOIN bob")
				bob.expect(t, "EVT GROUP JOIN bob")
			}
			prefix := "EVT " + scope + " CHAT アリス "
			bodyBytes := maxProtocolLineBytes - len(prefix)
			body := strings.Repeat("あ", bodyBytes/len("あ")) + strings.Repeat("x", bodyBytes%len("あ"))
			alice.command(t, "CHAT "+scope+" "+body+"x", "ERR 400 BAD_REQUEST")
			alice.command(t, "WHO", "OK players=2")
			bob.command(t, "WHO", "OK players=2")
			alice.command(t, "CHAT "+scope+" "+body, "OK")
			for _, client := range []*testClient{alice, bob} {
				scanner := bufio.NewScanner(client.reader)
				if !scanner.Scan() {
					t.Fatalf("read maximum-length event: %v", scanner.Err())
				}
				if got := scanner.Text(); got != prefix+body {
					t.Fatalf("event differs from sent message: bytes=%d", len(got))
				}
				client.command(t, "WHO", "OK players=2")
			}
		})
	}
}
