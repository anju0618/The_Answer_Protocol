package main

const (
	allyDamageBonus = 5

	allyCounterReductionPercent = 20

	maxAllyBonusCount = 3
)

func (s *Server) alliesInRoomLocked(name string) []string {
	player := s.players[name]
	group := s.groups[s.groupByPlayer[name]]
	if player == nil || group == nil {
		return nil
	}
	var allies []string
	for memberName := range group.Members {
		member := s.players[memberName]
		if memberName != name && member != nil && !member.exiting && member.RoomID == player.RoomID {
			allies = append(allies, memberName)
		}
	}
	return allies
}

func allyBonusCount(allies []string) int {
	if len(allies) > maxAllyBonusCount {
		return maxAllyBonusCount
	}
	return len(allies)
}

type deathOutcome string

const (
	outcomeNothingLost deathOutcome = ""
	outcomeLost        deathOutcome = "lost"
	outcomeKept        deathOutcome = "kept"
)

func (s *Server) applyDeathPenaltyLocked(player *Player, name string) deathOutcome {
	protected := len(s.alliesInRoomLocked(name)) > 0
	player.EnemyHP = nil
	player.FledFrom = nil
	if s.world == nil {
		return outcomeNothingLost
	}

	var kept, lost []string
	for _, itemID := range player.Inventory {
		if item := s.world.Items[itemID]; item == nil || item.RewardOnly {
			kept = append(kept, itemID)
		} else {
			lost = append(lost, itemID)
		}
	}
	if len(lost) == 0 {
		return outcomeNothingLost
	}
	if protected {
		return outcomeKept
	}

	if err := s.returnItemsHomeLocked(name, lost); err != nil {
		logger.Error("return_lost_items_failed", "player", name, "error", err.Error())
		return outcomeNothingLost
	}
	player.Inventory = kept
	return outcomeLost
}

func (s *Server) returnItemsHomeLocked(name string, itemIDs []string) error {
	if len(itemIDs) == 0 {
		return nil
	}
	lost := make(map[string]bool, len(itemIDs))
	returned := make(map[string]string)
	for _, itemID := range itemIDs {
		lost[itemID] = true
		item := s.world.Items[itemID]
		if item.Renewable {
			continue
		}
		roomID := item.HomeRoomID
		if roomID == "" {
			roomID = s.world.StartRoomID
		}
		returned[itemID] = roomID
	}

	s.ioMu.Lock()
	defer s.ioMu.Unlock()
	locations, err := s.loadItemLocations()
	if err != nil {
		return err
	}
	previousLocations := make(map[string]string, len(locations))
	for itemID, roomID := range locations {
		previousLocations[itemID] = roomID
	}
	for itemID, roomID := range returned {
		locations[itemID] = roomID
	}
	if len(returned) > 0 {
		if err := s.writeItemLocations(locations); err != nil {
			return err
		}
	}

	players, err := s.loadPlayers()
	if err == nil {
		if saved := players[name]; saved != nil {
			var inventory []string
			for _, itemID := range saved.Inventory {
				if !lost[itemID] {
					inventory = append(inventory, itemID)
				}
			}
			saved.Inventory = inventory
			err = s.writePlayers(players)
		}
	}
	if err != nil {
		if len(returned) > 0 {
			if rollbackErr := s.writeItemLocations(previousLocations); rollbackErr != nil {
				logger.Error("restore_item_location_failed", "player", name, "error", rollbackErr.Error())
			}
		}
		return err
	}

	for itemID, roomID := range returned {
		s.world.Items[itemID].RoomID = roomID
		delete(s.unsavedTakes[name], itemID)
	}
	return nil
}

func (s *Server) shareVictoryLocked(attacker string, allies []string, npcID string, npc *NPC) {
	for _, allyName := range allies {
		ally := s.players[allyName]
		if ally == nil || ally.exiting || ally.enemyHP(npcID, npc) <= 0 {
			continue
		}
		ally.setEnemyHP(npcID, 0)
		logger.Info("victory_shared", "player", allyName, "ally_of", attacker, "npc", npcID)
		ally.CombatTargetID = ""
		s.checkQuestObjectiveLocked(ally, "defeat_npc", npcID)
		locale := s.localeOfLocked(allyName)
		s.sendPlayerEventLocked(allyName, "TEAM", LocalizedText{
			"en": "Your ally %s defeated %s, and you share the victory.",
			"ja": "仲間の%sが%sを打ち倒した。その手柄はあなたにも与えられる。",
		}.Format(locale, attacker, npc.Name.Get(locale)))
	}
}
