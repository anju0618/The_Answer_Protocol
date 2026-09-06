package main

import "fmt"

// applyRoomHazardLocked evaluates the destination room's entry hazard, if
// any, right after a successful MOVE. It may kill and respawn the player or
// spend crew. It returns a flavor line to broadcast to the room the hazard
// happened in, or "" if nothing happened. Callers must hold s.mu.
func (s *Server) applyRoomHazardLocked(player *Player, name string, room *Room) string {
	if room == nil || room.Hazard == nil {
		return ""
	}
	hazard := room.Hazard

	switch hazard.Type {
	case "lethal":
		s.respawnPlayerLocked(player, name)
		return fmt.Sprintf("%s is lost to %s.", name, room.Name)

	case "item_gate":
		if player.hasItem(hazard.RequiredItemID) {
			return ""
		}
		s.respawnPlayerLocked(player, name)
		return fmt.Sprintf("%s, unprepared, does not survive %s.", name, room.Name)

	case "crew_gate":
		if player.Crew+1 < hazard.MinPartyTotal {
			s.respawnPlayerLocked(player, name)
			return fmt.Sprintf("%s and the remaining crew are lost passing %s.", name, room.Name)
		}
		lost := spendCrewLocked(player, hazard.CrewLoss)
		return fmt.Sprintf("%s passes %s, losing %d crew.", name, room.Name, lost)
	}
	return ""
}
