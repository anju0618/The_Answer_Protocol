package main

import "sort"

// hasDefeated reports whether this player has beaten the enemy. Defeats are per player and are forgotten on death.
// An unwinnable enemy can never be defeated, only escaped.
func (p *Player) hasDefeated(npcID string, npc *NPC) bool {
	return npc != nil && npc.Role == "enemy" && !npc.Unwinnable && p.enemyHP(npcID, npc) <= 0
}

// defeatedNPCsLocked lists (sorted) the NPCs in the room that this player has defeated.
func (s *Server) defeatedNPCsLocked(player *Player, roomID string) []string {
	var ids []string
	for id, npc := range s.world.NPCs {
		if npc != nil && npc.RoomID == roomID && player.hasDefeated(id, npc) {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}

// roomClearedLocked is true when the room has enemies that can be beaten and this player has beaten every one.
func (s *Server) roomClearedLocked(player *Player, roomID string) bool {
	found := false
	for id, npc := range s.world.NPCs {
		if npc == nil || npc.RoomID != roomID || npc.Role != "enemy" || npc.Unwinnable {
			continue
		}
		if !player.hasDefeated(id, npc) {
			return false
		}
		found = true
	}
	return found
}

// dialogueFor is what the NPC says to this player: the "cleared" lines once the room's enemies are beaten.
func (s *Server) dialogueFor(player *Player, npcID string, npc *NPC) []LocalizedText {
	if len(npc.DialogueCleared) > 0 && npc.Role != "enemy" && s.roomClearedLocked(player, npc.RoomID) {
		return npc.DialogueCleared
	}
	return npc.Dialogue
}
