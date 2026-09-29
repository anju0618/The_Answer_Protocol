package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strings"
)

// checkQuestObjectiveLocked advances any of the player's active quests whose
// objective matches (objType, targetID) — called after a successful TAKE or
// a successful ATTACK. Completing a quest heals the player by its reward,
// capped at maxPlayerHP. Callers must hold s.mu.
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
		state.Progress++
		if state.Progress >= quest.Objective.Count {
			state.Status = "completed"
			player.HP += quest.Reward.HP
			if player.HP > maxPlayerHP {
				player.HP = maxPlayerHP
			}
		}
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

	s.mu.Lock()
	player := s.playerForUpdateLocked(*name)
	if player == nil || s.world == nil || s.world.Rooms[player.RoomID] == nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	npcID := s.world.resolveNPCInRoom(player.RoomID, query)
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
	if player.Quests[questID] == nil {
		player.Quests[questID] = &PlayerQuest{Status: "active"}
	}

	data, err := json.Marshal(struct {
		QuestID     string `json:"quest_id"`
		Description string `json:"description"`
		Reward      int    `json:"reward"`
		Status      string `json:"status"`
	}{questID, quest.Description, quest.Reward.HP, "available"})
	if err != nil {
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
