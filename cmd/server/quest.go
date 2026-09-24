// QUEST/QUESTSコマンドと、クエスト進行の自動判定。手動の完了報告コマンドは
// 用意せず、TAKE/ATTACKが成功した瞬間にサーバー側で該当する進行中クエストを
// 自動チェック・自動達成・自動報酬付与する(README「Quest System」参照)。
package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strings"
)

// checkQuestObjectiveLocked は、objType("collect_item"または"defeat_npc")と
// targetID(取得したアイテムID/倒したNPC ID)に一致する、player が
// 受注済み(active)の全クエストを1つ進める。目標数に達したクエストは
// completedにし、報酬HPを即座に回復させる(上限maxPlayerHP)。
// TAKE成功時・ATTACK勝利時から呼ばれる。
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
		s.advanceQuestLocked(player, quest, state)
	}
}

// advanceQuestLocked は player の進行中クエスト1件を1つ進める。目標数に
// 達したら達成にして報酬HPを与える(上限maxPlayerHP)。進捗も達成も、
// 本人にEVT PLAYER QUESTで知らせる(以前は何も通知されず、達成したかどうかが
// QUESTSを打つまで分からなかった)。
func (s *Server) advanceQuestLocked(player *Player, quest *Quest, state *PlayerQuest) {
	locale := s.localeOfLocked(player.Name)
	state.Progress++
	if state.Progress < quest.Objective.Count {
		s.sendPlayerEventLocked(player.Name, "QUEST", LocalizedText{
			"en": "Quest \"%s\" progress: %d/%d.",
			"ja": "クエスト「%s」の進捗: %d/%d。",
		}.Format(locale, quest.Name.Get(locale), state.Progress, quest.Objective.Count))
		return
	}
	state.Status = "completed"
	player.HP += quest.Reward.HP
	if player.HP > maxPlayerHP {
		player.HP = maxPlayerHP
	}
	s.sendPlayerEventLocked(player.Name, "QUEST", LocalizedText{
		"en": "Quest complete: \"%s\"! Reward: +%d HP (HP is now %d). Use QUESTS to see your quests, and look for the next NPC with a request.",
		"ja": "クエスト達成: 「%s」! 報酬: HP+%d(現在HP %d)。QUESTSで一覧を確認し、次の依頼を探そう。",
	}.Format(locale, quest.Name.Get(locale), quest.Reward.HP, player.HP))
}

// questGiverNPCIDsLocked は roomID にいる、クエストを持つNPCのIDをID順で返す。
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

// announceQuestGiversLocked は player が今いる部屋に、まだ受けていない
// クエストを持つNPCがいれば「依頼がある。QUEST <名前>で受けられる」と
// 本人にEVT PLAYER QUESTで知らせる。LOOKのレスポンスにはNPCのIDしか
// 載らず(RFCの形式は変えられない)、誰がクエストを持っているかが
// 分からなかったため、部屋に入った瞬間に案内する。MOVE後・CONNECT後に呼ぶ。
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

// sendQuestHintLocked は quest_giver に TALK した後、そのNPCのクエストの
// 状況(未受注なら受け方、進行中なら進捗)を本人へ知らせる。
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
			"en": "%s has a quest for you: \"%s\" (reward: +%d HP). Type QUEST %s to accept it.",
			"ja": "%sはあなたに頼みたいことがある: 「%s」(報酬: HP+%d)。QUEST %s で受注できる。",
		}.Format(locale, npcName, quest.Name.Get(locale), quest.Reward.HP, npcName))
	case state.Status == "active":
		s.sendPlayerEventLocked(player.Name, "QUEST", LocalizedText{
			"en": "Quest \"%s\" is in progress (%d/%d). Type QUEST %s to hear the details again.",
			"ja": "クエスト「%s」は進行中(%d/%d)。QUEST %s で内容をもう一度聞ける。",
		}.Format(locale, quest.Name.Get(locale), state.Progress, quest.Objective.Count, npcName))
	}
}

// handleQuest はQUEST <npc> コマンドを処理する。指定NPCが持つクエストの
// 情報(未達成なら)を返し、初回呼び出し時はそのクエストをプレイヤーの
// アクティブなクエストとして登録する。達成済みクエストを再度尋ねたり、
// クエストを持たないNPCに尋ねたりするとERR 406を返す。
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
	}

	data, err := json.Marshal(struct {
		QuestID     string `json:"quest_id"`
		Description string `json:"description"`
		Reward      int    `json:"reward"`
		Status      string `json:"status"`
	}{questID, quest.Description.Get(locale), quest.Reward.HP, "available"})
	if err != nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	client := conn.(*serverClient)
	response, err := client.enqueueResponse("OK " + string(data))
	if err == nil && newlyAccepted && s.objectiveAlreadyMetLocked(player, quest) {
		// 受注前に済ませていた分も数える(進捗はTAKE/ATTACKの瞬間にしか
		// 判定しないため、先に拾ったり倒したりしていると永遠に達成できなく
		// なってしまう)。
		s.advanceQuestLocked(player, quest, player.Quests[questID])
	}
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}

// handleQuests はQUESTS(引数なし)コマンドを処理する。プレイヤーがこれ
// までに受注した(進行中・達成済み問わず)全クエストを、進行状況
// "現在数/目標数" 付きでID順に一覧表示する。
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

// objectiveAlreadyMetLocked は、クエストを受ける前の時点で player が既に
// 目標を果たしているか(対象アイテムを持っている/対象の敵を自分の世界で
// 倒し済み)を返す。
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
