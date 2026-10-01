package main

import (
	"bufio"
	"net"
	"testing"
	"time"
)

func TestStateSnapshotIncludesCrewRosterAndPrivateInvitations(t *testing.T) {
	server := newServer(t.TempDir())
	alice := startTestClient(t, server)
	alice.command(t, "STATE", "ERR 400 BAD_REQUEST")
	alice.connect(t, "alice")
	bob := startTestClient(t, server)
	bob.connect(t, "bob", alice)
	carol := startTestClient(t, server)
	carol.connect(t, "carol", alice, bob)

	alice.command(t, "STATE extra", "ERR 400 BAD_REQUEST")
	alice.command(t, "GROUP CREATE", "OK group=group.1")
	alice.command(t, "GROUP INVITE bob", "OK")
	bob.expect(t, "EVT GROUP INVITE alice")
	server.mu.Lock()
	server.players["alice"].Crew = 8
	server.players["alice"].CrewInitialized = true
	server.players["carol"].exiting = true
	server.mu.Unlock()

	alice.command(t, "STATE", `OK {"crew":8,"crew_initialized":true,"players":["alice","bob"],"group":"group.1","invitations":[]}`)
	bob.command(t, "STATE", `OK {"crew":0,"crew_initialized":false,"players":["alice","bob"],"group":"","invitations":["alice"]}`)
	carol.command(t, "STATE", "ERR 500 STATE_ERROR")
	alice.command(t, "STATUS", `OK {"hp":100,"max_hp":100,"status":"healthy"}`)
	bob.command(t, "WHO", "OK players=3")

	bob.command(t, "GROUP JOIN alice", "OK group=group.1")
	bob.expect(t, "EVT GROUP JOIN bob")
	alice.expect(t, "EVT GROUP JOIN bob")
	bob.command(t, "STATE", `OK {"crew":0,"crew_initialized":false,"players":["alice","bob"],"group":"group.1","invitations":[]}`)

	server.mu.Lock()
	server.players["alice"].Crew = 0
	server.mu.Unlock()
	alice.command(t, "STATE", `OK {"crew":0,"crew_initialized":true,"players":["alice","bob"],"group":"group.1","invitations":[]}`)
	server.mu.Lock()
	server.players["carol"].exiting = false
	server.mu.Unlock()
}

func TestStateOverTCPTracksCrewAfterMoving(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = &World{
		StartRoomID: "loc.hall",
		Rooms: map[string]*Room{
			"loc.hall":         {ID: "loc.hall", Exits: map[string]string{"east": odysseyStartRoomID}},
			odysseyStartRoomID: {ID: odysseyStartRoomID, Exits: map[string]string{"north": "loc.danger"}},
			"loc.danger":       {ID: "loc.danger", Hazard: &RoomHazard{Type: "crew_cost", CrewLoss: 2}},
		},
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err == nil {
			server.handleClient(conn)
		}
	}()
	conn, err := net.DialTimeout("tcp", listener.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close(); listener.Close(); <-done })
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	client := &testClient{conn: conn, reader: bufio.NewReader(conn), server: server, done: done}
	client.expect(t, "OK hello proto=1")
	client.connect(t, "alice")
	client.cmd(t, "MOVE east", "OK room="+odysseyStartRoomID)
	client.cmd(t, "STATE", `OK {"crew":12,"crew_initialized":true,"players":["alice"],"group":"","invitations":[]}`)
	client.cmd(t, "MOVE north", "OK room=loc.danger")
	client.cmd(t, "STATE", `OK {"crew":10,"crew_initialized":true,"players":["alice"],"group":"","invitations":[]}`)
	client.cmd(t, "STATUS", `OK {"hp":100,"max_hp":100,"status":"healthy"}`)
}
