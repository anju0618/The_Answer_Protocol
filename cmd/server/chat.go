package main

import (
	"fmt"
	"net"
	"strings"
	"unicode"
	"unicode/utf8"
)

func parseCommandParts(line string) []string {
	parts := strings.Fields(line)
	if len(parts) < 3 || !strings.EqualFold(parts[0], "CHAT") {
		return parts
	}

	rest := strings.TrimLeftFunc(line, unicode.IsSpace)
	for range 2 {
		separator := strings.IndexFunc(rest, unicode.IsSpace)
		if separator < 0 {
			return parts
		}
		rest = strings.TrimLeftFunc(rest[separator:], unicode.IsSpace)
	}
	return []string{parts[0], parts[1], strings.TrimSuffix(rest, "\r")}
}

func handleChat(s *Server, conn net.Conn, name *string, parts []string) bool {
	if !requireArgs(conn, parts, 3) {
		return false
	}
	if *name == "" {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	scope := strings.ToUpper(parts[1])
	if scope != "GLOBAL" && scope != "ROOM" && scope != "GROUP" {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	message := strings.Join(parts[2:], " ")
	if strings.TrimSpace(message) == "" || !utf8.ValidString(message) || strings.IndexFunc(message, unicode.IsControl) >= 0 {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	event := "EVT " + scope + " CHAT " + *name + " " + message
	if len(event) > maxProtocolLineBytes {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}

	client := conn.(*serverClient)
	s.mu.Lock()
	player := s.players[*name]
	if player == nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	var group *Group
	if scope == "GROUP" {
		id, member := s.groupByPlayer[*name]
		if !member {
			s.mu.Unlock()
			fmt.Fprintln(conn, "ERR 401 NOT_IN_GROUP")
			return false
		}
		group = s.groups[id]
		if group == nil {
			s.mu.Unlock()
			fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
			return false
		}
	}
	response, err := client.enqueueResponse("OK")
	if err == nil {
		switch scope {
		case "GLOBAL":
			for _, recipient := range s.clients {
				recipient.enqueueEvent(event)
			}
		case "ROOM":
			for name, other := range s.players {
				if other.RoomID == player.RoomID {
					if recipient := s.clients[name]; recipient != nil {
						recipient.enqueueEvent(event)
					}
				}
			}
		case "GROUP":
			s.broadcastGroupLocked(group, event)
		}
	}
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}
