package main

import "fmt"

func (s *Server) applyRoomHazardLocked(player *Player, name string, room *Room, locale string) string {
	if room == nil || room.Hazard == nil {
		return ""
	}
	hazard := room.Hazard
	roomName := room.Name.Get(locale)

	switch hazard.Type {
	case "lethal":
		s.respawnPlayerLocked(player, name)
		return fmt.Sprintf("%s is lost to %s.", name, roomName)

	case "item_gate":
		if player.hasItem(hazard.RequiredItemID) {
			return ""
		}
		s.respawnPlayerLocked(player, name)
		return fmt.Sprintf("%s, unprepared, does not survive %s.", name, roomName)

	case "crew_gate":
		if player.Crew+1 < hazard.MinPartyTotal {
			s.respawnPlayerLocked(player, name)
			return fmt.Sprintf("%s and the remaining crew are lost passing %s.", name, roomName)
		}
		lost := spendCrewLocked(player, hazard.CrewLoss)
		return fmt.Sprintf("%s passes %s, losing %d crew.", name, roomName, lost)
	}
	return ""
}
