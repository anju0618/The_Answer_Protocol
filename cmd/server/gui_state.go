package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sort"
)

func handleState(s *Server, conn net.Conn, name *string, parts []string) bool {
	if !requireExactArgs(conn, parts, 1) {
		return false
	}
	if *name == "" {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	s.mu.Lock()
	player := s.players[*name]
	if player == nil || player.exiting {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	state := struct {
		Crew            int      `json:"crew"`
		CrewInitialized bool     `json:"crew_initialized"`
		Players         []string `json:"players"`
		Group           string   `json:"group"`
		Invitations     []string `json:"invitations"`
	}{
		Crew: player.Crew, CrewInitialized: player.CrewInitialized,
		Group: s.groupByPlayer[*name], Players: []string{}, Invitations: []string{},
	}
	for otherName, other := range s.players {
		if !other.exiting {
			state.Players = append(state.Players, otherName)
		}
	}
	for _, group := range s.groups {
		if _, invited := group.Invited[*name]; invited {
			state.Invitations = append(state.Invitations, group.Leader)
		}
	}
	sort.Strings(state.Players)
	sort.Strings(state.Invitations)
	data, err := json.Marshal(state)
	if err != nil || len(data)+len("OK ") > maxProtocolLineBytes {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	client := conn.(*serverClient)
	response, err := client.enqueueResponse("OK " + string(data))
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}
