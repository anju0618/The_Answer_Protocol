package main

import (
	"fmt"
	"net"
	"strings"
)

type Group struct {
	ID      string
	Leader  string
	Members map[string]struct{}
	Invited map[string]struct{}
}

func newGroup(id, creator string) *Group {
	return &Group{
		ID:      id,
		Leader:  creator,
		Members: map[string]struct{}{creator: {}},
		Invited: make(map[string]struct{}),
	}
}

func handleGroup(s *Server, conn net.Conn, name *string, parts []string) bool {
	if !requireArgs(conn, parts, 2) {
		return false
	}
	action := strings.ToUpper(parts[1])
	switch action {
	case "CREATE", "LEAVE":
		if !requireExactArgs(conn, parts, 2) {
			return false
		}
	case "INVITE", "JOIN":
		if !requireExactArgs(conn, parts, 3) {
			return false
		}
	default:
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	if *name == "" {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}

	switch action {
	case "CREATE":
		return handleGroupCreate(s, conn, *name)
	case "INVITE":
		return handleGroupInvite(s, conn, *name, parts[2])
	case "JOIN":
		return handleGroupJoin(s, conn, *name, parts[2])
	default:
		return handleGroupLeave(s, conn, *name)
	}
}

func handleGroupCreate(s *Server, conn net.Conn, name string) bool {
	client := conn.(*serverClient)
	s.mu.Lock()
	if s.players[name] == nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	if _, exists := s.groupByPlayer[name]; exists {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 402 ALREADY_IN_GROUP")
		return false
	}
	id := fmt.Sprintf("group.%d", s.nextGroupID+1)
	response, err := client.enqueueResponse("OK group=" + id)
	if err == nil {
		s.nextGroupID++
		s.groups[id] = newGroup(id, name)
		s.groupByPlayer[name] = id
		s.clearGroupInvitesLocked(name)
	}
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}

func handleGroupInvite(s *Server, conn net.Conn, name, target string) bool {
	client := conn.(*serverClient)
	s.mu.Lock()
	id, member := s.groupByPlayer[name]
	group := s.groups[id]
	if !member {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 401 NOT_IN_GROUP")
		return false
	}
	if group == nil || s.players[name] == nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	if _, exists := s.groupByPlayer[target]; exists {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 402 ALREADY_IN_GROUP")
		return false
	}
	invitee := s.clients[target]
	if s.players[target] == nil || invitee == nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	response, err := client.enqueueResponse("OK")
	if err == nil {
		group.Invited[target] = struct{}{}
		invitee.enqueueEvent("EVT GROUP INVITE " + group.Leader)
	}
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}

func handleGroupJoin(s *Server, conn net.Conn, name, leader string) bool {
	client := conn.(*serverClient)
	s.mu.Lock()
	if s.players[name] == nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	if _, exists := s.groupByPlayer[name]; exists {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 402 ALREADY_IN_GROUP")
		return false
	}
	id := s.groupByPlayer[leader]
	group := s.groups[id]
	if group == nil || group.Leader != leader {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	if _, invited := group.Invited[name]; !invited {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	response, err := client.enqueueResponse("OK group=" + id)
	if err == nil {
		group.Members[name] = struct{}{}
		s.groupByPlayer[name] = id
		s.clearGroupInvitesLocked(name)
		s.broadcastGroupLocked(group, "EVT GROUP JOIN "+name)
	}
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}

func handleGroupLeave(s *Server, conn net.Conn, name string) bool {
	client := conn.(*serverClient)
	s.mu.Lock()
	if _, member := s.groupByPlayer[name]; !member {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 401 NOT_IN_GROUP")
		return false
	}
	response, err := client.enqueueResponse("OK")
	if err == nil {
		s.removeGroupMemberLocked(name)
	}
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}

func (s *Server) broadcastGroupLocked(group *Group, event string) {
	for member := range group.Members {
		if client := s.clients[member]; client != nil {
			client.enqueueEvent(event)
		}
	}
}

func (s *Server) clearGroupInvitesLocked(name string) {
	for _, group := range s.groups {
		delete(group.Invited, name)
	}
}

func (s *Server) removeGroupMemberLocked(name string) {
	id, member := s.groupByPlayer[name]
	if !member {
		return
	}
	delete(s.groupByPlayer, name)
	group := s.groups[id]
	if group == nil {
		return
	}
	delete(group.Members, name)
	if len(group.Members) == 0 {
		delete(s.groups, id)
		return
	}
	if group.Leader == name {
		group.Leader = ""
		for member := range group.Members {
			if group.Leader == "" || member < group.Leader {
				group.Leader = member
			}
		}
		clear(group.Invited)
	}
	s.broadcastGroupLocked(group, "EVT GROUP LEAVE "+name)
}
