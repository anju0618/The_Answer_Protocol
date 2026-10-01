package main

func (s *Server) applyRoomHazardLocked(player *Player, name string, room *Room, locale string) *flavor {
	if room == nil || room.Hazard == nil {
		return nil
	}
	hazard := room.Hazard
	roomName := room.Name.Get(locale)

	switch hazard.Type {
	case "lethal":
		if description := room.Description.Get(locale); description != "" {
			s.sendPlayerEventLocked(name, "DEATH", description)
		}
		s.respawnPlayerLocked(player, name, "hazard_lethal", roomName)
		return &flavor{key: "hazard_lethal", player: name, room: room}

	case "item_gate":
		if player.hasItem(hazard.RequiredItemID) {
			return nil
		}
		s.respawnPlayerLocked(player, name, "hazard_item", roomName)
		return &flavor{key: "hazard_item", player: name, room: room}

	case "crew_gate":
		if player.Crew+1 < hazard.MinPartyTotal {
			s.respawnPlayerLocked(player, name, "hazard_crew", roomName)
			return &flavor{key: "hazard_crew_dead", player: name, room: room}
		}
		lost := spendCrewLocked(player, hazard.CrewLoss)
		return &flavor{key: "hazard_crew_loss", player: name, room: room, n: lost}

	case "crew_cost":
		lost := spendCrewLocked(player, hazard.CrewLoss)
		return &flavor{key: "hazard_crew_loss", player: name, room: room, n: lost}
	}
	return nil
}

func (s *Server) blockingEnemyLocked(player *Player, roomID string) (string, *NPC) {
	blockID := ""
	for id, npc := range s.world.NPCs {
		if npc == nil || npc.RoomID != roomID || npc.Role != "enemy" || player.enemyHP(id, npc) <= 0 {
			continue
		}
		if player.FledFrom[id] {
			continue
		}
		if blockID == "" || id < blockID {
			blockID = id
		}
	}
	if blockID == "" {
		return "", nil
	}
	return blockID, s.world.NPCs[blockID]
}
