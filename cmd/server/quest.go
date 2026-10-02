package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strings"
)

func (s *Server) checkQuestObjectiveLocked(player *Player, objType, targetID string) {
	if s.world == nil {
		return
	}
	for questID, quest := range s.world.Quests {
		if quest == nil || quest.Objective.Type != objType || quest.Objective.TargetID != targetID {
			continue
		}
		state := player.Quests[questID]
		if state == nil || state.Status != "active" {
			continue
		}
		s.advanceQuestLocked(player, questID, quest, state)
	}
}

func (s *Server) advanceQuestLocked(player *Player, questID string, quest *Quest, state *PlayerQuest) {
	locale := s.localeOfLocked(player.Name)
	state.Progress++
	if state.Progress < quest.Objective.Count {
		logger.Info("quest_progress", "player", player.Name, "quest", questID, "progress", state.Progress, "target", quest.Objective.Count)
	}
	if state.Progress < quest.Objective.Count {
		s.sendPlayerEventLocked(player.Name, "QUEST", LocalizedText{
			"en": "Quest \"%s\" progress: %d/%d.",
			"ja": "クエスト「%s」の進捗: %d/%d。",
		}.Format(locale, quest.Name.Get(locale), state.Progress, quest.Objective.Count))
		return
	}
	state.Status = "completed"
	logger.Info("quest_completed", "player", player.Name, "quest", questID, "max_hp_gain", questMaxHPGain(quest))
	gain := questMaxHPGain(quest)
	player.MaxHPBonus += gain
	player.HP = s.maxHPLocked(player)
	s.sendPlayerEventLocked(player.Name, "QUEST", LocalizedText{
		"en": "Quest complete: \"%s\"! Reward: max HP +%d (now %d), and your HP is fully restored. Use QUESTS to see your quests, and look for the next NPC with a request.",
		"ja": "クエスト達成: 「%s」! 報酬: 最大HP+%d(現在の最大HP %d)、HPも全回復。QUESTSで一覧を確認し、次の依頼を探そう。",
	}.Format(locale, quest.Name.Get(locale), gain, s.maxHPLocked(player)))
}

func (s *Server) questGiverNPCIDsLocked(roomID string) []string {
	var ids []string
	for id, npc := range s.world.NPCs {
		if npc == nil || npc.RoomID != roomID {
			continue
		}
		if _, quest := s.world.questByGiver(id); quest != nil {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}

func (s *Server) announceQuestGiversLocked(player *Player) {
	if s.world == nil {
		return
	}
	locale := s.localeOfLocked(player.Name)
	for _, npcID := range s.questGiverNPCIDsLocked(player.RoomID) {
		questID, quest := s.world.questByGiver(npcID)
		if player.Quests[questID] != nil {
			continue
		}
		npcName := s.world.NPCs[npcID].Name.Get(locale)
		s.sendPlayerEventLocked(player.Name, "QUEST", LocalizedText{
			"en": "%s has a request for you: \"%s\". Type QUEST %s to hear it and accept.",
			"ja": "%sから依頼がある: 「%s」。QUEST %s と入力すると内容を聞いて受注できる。",
		}.Format(locale, npcName, quest.Name.Get(locale), npcName))
	}
}

func (s *Server) sendQuestHintLocked(player *Player, npcID string) {
	questID, quest := s.world.questByGiver(npcID)
	if quest == nil {
		return
	}
	locale := s.localeOfLocked(player.Name)
	npcName := s.world.NPCs[npcID].Name.Get(locale)
	state := player.Quests[questID]
	switch {
	case state == nil:
		s.sendPlayerEventLocked(player.Name, "QUEST", LocalizedText{
			"en": "%s has a quest for you: \"%s\" (reward: max HP +%d). Type QUEST %s to accept it.",
			"ja": "%sはあなたに頼みたいことがある: 「%s」(報酬: 最大HP+%d)。QUEST %s で受注できる。",
		}.Format(locale, npcName, quest.Name.Get(locale), questMaxHPGain(quest), npcName))
	case state.Status == "active":
		s.sendPlayerEventLocked(player.Name, "QUEST", LocalizedText{
			"en": "Quest \"%s\" is in progress (%d/%d). Type QUEST %s to hear the details again.",
			"ja": "クエスト「%s」は進行中(%d/%d)。QUEST %s で内容をもう一度聞ける。",
		}.Format(locale, quest.Name.Get(locale), state.Progress, quest.Objective.Count, npcName))
	}
}

func handleQuest(s *Server, conn net.Conn, name *string, parts []string) bool {
	if !requireArgs(conn, parts, 2) {
		return false
	}
	if *name == "" {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	query := strings.Join(parts[1:], " ")
	locale := clientLocale(conn)

	s.mu.Lock()
	player := s.playerForUpdateLocked(*name)
	if player == nil || s.world == nil || s.world.Rooms[player.RoomID] == nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	npcID := s.world.resolveNPCInRoom(player.RoomID, query, locale)
	if npcID == "" {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 404 NPC_NOT_FOUND")
		return false
	}
	questID, quest := s.world.questByGiver(npcID)
	if quest == nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 406 NO_QUEST_AVAILABLE")
		return false
	}
	if state := player.Quests[questID]; state != nil && state.Status == "completed" {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 406 NO_QUEST_AVAILABLE")
		return false
	}
	if player.Quests == nil {
		player.Quests = make(map[string]*PlayerQuest)
	}
	newlyAccepted := player.Quests[questID] == nil
	if newlyAccepted {
		player.Quests[questID] = &PlayerQuest{Status: "active"}
		logger.Info("quest_accepted", "player", *name, "quest", questID, "giver", quest.GiverNPCID)
	}

	data, err := json.Marshal(struct {
		QuestID     string `json:"quest_id"`
		Description string `json:"description"`
		Reward      int    `json:"reward"`
		Status      string `json:"status"`
	}{questID, quest.Description.Get(locale), questMaxHPGain(quest), "available"})
	if err != nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	client := conn.(*serverClient)
	response, err := client.enqueueResponse("OK " + string(data))
	if err == nil && newlyAccepted && s.objectiveAlreadyMetLocked(player, quest) {

		s.advanceQuestLocked(player, questID, quest, player.Quests[questID])
	}
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}

func handleQuests(s *Server, conn net.Conn, name *string, parts []string) bool {
	if !requireExactArgs(conn, parts, 1) {
		return false
	}
	if *name == "" {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}

	s.mu.Lock()
	player := s.players[*name]
	if player == nil || s.world == nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	ids := make([]string, 0, len(player.Quests))
	for id := range player.Quests {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	type questEntry struct {
		QuestID  string `json:"quest_id"`
		Status   string `json:"status"`
		Progress string `json:"progress"`
	}
	list := make([]questEntry, 0, len(ids))
	for _, id := range ids {
		state := player.Quests[id]
		if state == nil {
			s.mu.Unlock()
			fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
			return false
		}
		count := 1
		if quest := s.world.Quests[id]; quest != nil {
			count = quest.Objective.Count
		}
		list = append(list, questEntry{id, state.Status, fmt.Sprintf("%d/%d", state.Progress, count)})
	}

	data, err := json.Marshal(list)
	if err != nil || len("OK ")+len(data) > maxProtocolLineBytes {
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

func (s *Server) objectiveAlreadyMetLocked(player *Player, quest *Quest) bool {
	switch quest.Objective.Type {
	case "collect_item":
		return player.hasItem(quest.Objective.TargetID)
	case "defeat_npc":
		npc := s.world.NPCs[quest.Objective.TargetID]
		return npc != nil && player.enemyHP(quest.Objective.TargetID, npc) <= 0
	}
	return false
}
