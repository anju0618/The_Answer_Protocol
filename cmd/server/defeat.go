package main

import "sort"

func (p *Player) hasDefeated(npcID string, npc *NPC) bool {
	return npc != nil && npc.Role == "enemy" && !npc.Unwinnable && p.enemyHP(npcID, npc) <= 0
}

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

func (s *Server) dialogueFor(player *Player, npcID string, npc *NPC) []LocalizedText {
	if len(npc.DialogueCleared) > 0 && npc.Role != "enemy" && s.roomClearedLocked(player, npc.RoomID) {
		return npc.DialogueCleared
	}
	return npc.Dialogue
}
