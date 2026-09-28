// 3本の物語(アルゴ船・トロイア戦争・オデュッセイア)それぞれのエンディングと、
// 3本すべてを終えた人だけが見られる最終エンディングの仕組み。
//
// エンディングは新しいコマンドを増やさず、「最終地点にいるNPCにTALKする」
// ことで発生する。NPCの Ending(data/world.json)に、必要なアイテム・クエスト・
// (最終エンディングでは)他のエンディングが書かれていて、揃った状態で
// TALKすると物語の結末が EVT PLAYER ENDING で流れ、報酬アイテムが所持品に
// 入る。足りない場合は、何が足りないかを同じ EVT PLAYER ENDING で教える。
// 必要アイテムは「見せる」だけで消費しない(鍵アイテムはRenewableな各自の
// コピーなので、取り合いにならない。world.go の Item のコメント参照)。
package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

// Ending は NPC が持つエンディング1つぶんの定義。
type Ending struct {
	ID              string          `json:"id"`
	Name            LocalizedText   `json:"name"`
	RequiresItems   []string        `json:"requires_items,omitempty"`
	RequiresQuests  []string        `json:"requires_quests,omitempty"`
	RequiresEndings []string        `json:"requires_endings,omitempty"`
	RewardItem      string          `json:"reward_item,omitempty"`
	Hint            LocalizedText   `json:"hint"`
	Text            []LocalizedText `json:"text"`
}

// validateEndings は全NPCのEndingの参照整合性(必要アイテム・クエスト・
// 報酬アイテム・他のエンディングのIDが実在するか、IDが重複していないか)を
// 検証する。World.validate から呼ばれる。
func (w *World) validateEndings() error {
	ids := make(map[string]string)
	for npcID, npc := range w.NPCs {
		if npc == nil || npc.Ending == nil {
			continue
		}
		e := npc.Ending
		if e.ID == "" {
			return fmt.Errorf("NPC %q has an ending with no id", npcID)
		}
		if other, dup := ids[e.ID]; dup {
			return fmt.Errorf("ending %q is defined by both %q and %q", e.ID, other, npcID)
		}
		ids[e.ID] = npcID
		if len(e.Text) == 0 {
			return fmt.Errorf("ending %q has no text", e.ID)
		}
		for _, itemID := range e.RequiresItems {
			if w.Items[itemID] == nil {
				return fmt.Errorf("ending %q requires unknown item %q", e.ID, itemID)
			}
		}
		for _, questID := range e.RequiresQuests {
			if w.Quests[questID] == nil {
				return fmt.Errorf("ending %q requires unknown quest %q", e.ID, questID)
			}
		}
		if e.RewardItem != "" {
			item := w.Items[e.RewardItem]
			if item == nil || !item.RewardOnly {
				return fmt.Errorf("ending %q reward %q must be an existing reward_only item", e.ID, e.RewardItem)
			}
		}
	}
	for _, npcID := range ids {
		for _, required := range w.NPCs[npcID].Ending.RequiresEndings {
			if _, ok := ids[required]; !ok {
				return fmt.Errorf("ending %q requires unknown ending %q", w.NPCs[npcID].Ending.ID, required)
			}
		}
	}
	return nil
}

// routeEndings は「物語1本ぶんのエンディング」(他のエンディングを必要と
// しないもの)をID順で返す。最終エンディングは含まない。
func (w *World) routeEndings() []*Ending {
	var list []*Ending
	for _, npc := range w.NPCs {
		if npc != nil && npc.Ending != nil && len(npc.Ending.RequiresEndings) == 0 {
			list = append(list, npc.Ending)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	return list
}

// endingByID は id のエンディングを返す(無ければnil)。
func (w *World) endingByID(id string) *Ending {
	for _, npc := range w.NPCs {
		if npc != nil && npc.Ending != nil && npc.Ending.ID == id {
			return npc.Ending
		}
	}
	return nil
}

// missingForEndingLocked は player が e の条件のうち、まだ満たしていない
// ものを locale の言語で説明した断片のリストで返す(全部満たしていれば空)。
func (s *Server) missingForEndingLocked(player *Player, e *Ending, locale string) []string {
	var missing []string
	for _, itemID := range e.RequiresItems {
		if !player.hasItem(itemID) {
			missing = append(missing, LocalizedText{
				"en": "\"%s\" in your inventory",
				"ja": "所持品の「%s」",
			}.Format(locale, s.world.Items[itemID].Name.Get(locale)))
		}
	}
	for _, questID := range e.RequiresQuests {
		if state := player.Quests[questID]; state == nil || state.Status != "completed" {
			missing = append(missing, LocalizedText{
				"en": "the completed quest \"%s\"",
				"ja": "クエスト「%s」の達成",
			}.Format(locale, s.world.Quests[questID].Name.Get(locale)))
		}
	}
	for _, endingID := range e.RequiresEndings {
		if !player.Endings[endingID] {
			missing = append(missing, LocalizedText{
				"en": "the ending \"%s\"",
				"ja": "エンディング「%s」への到達",
			}.Format(locale, s.world.endingByID(endingID).Name.Get(locale)))
		}
	}
	return missing
}

// sendEndingLocked は name にEVT PLAYER ENDINGを1行送る。
func (s *Server) sendEndingLocked(name, text string) {
	s.sendPlayerEventLocked(name, "ENDING", text)
}

// sendEndingProgressLocked は「到達したエンディング n/3」と各エンディングの
// 達成状況を1行で本人に伝える(モイライにTALKしたとき・エンディング到達時)。
func (s *Server) sendEndingProgressLocked(player *Player) {
	routes := s.world.routeEndings()
	if len(routes) == 0 {
		return
	}
	locale := s.localeOfLocked(player.Name)
	done := 0
	parts := make([]string, 0, len(routes))
	for _, e := range routes {
		state := LocalizedText{"en": "not yet", "ja": "未達成"}.Get(locale)
		if player.Endings[e.ID] {
			done++
			state = LocalizedText{"en": "done", "ja": "達成"}.Get(locale)
		}
		parts = append(parts, e.Name.Get(locale)+" ("+state+")")
	}
	separator := ", "
	if locale == "ja" {
		separator = "、"
	}
	s.sendEndingLocked(player.Name, LocalizedText{
		"en": "Endings reached: %d/%d. %s",
		"ja": "到達したエンディング: %d/%d。%s",
	}.Format(locale, done, len(routes), strings.Join(parts, separator)))
}

// grantEndingRewardLocked は報酬アイテムを、まだ持っていなければ所持品に加える。
func (s *Server) grantEndingRewardLocked(player *Player, e *Ending) {
	if e.RewardItem != "" && !player.hasItem(e.RewardItem) {
		player.Inventory = append(player.Inventory, e.RewardItem)
	}
}

// talkEndingLocked は npc に Ending があるときの TALK の追加処理。
//   - 既に到達済み: 結末をもう一度流す(報酬を落としていたら再度渡す)
//   - 条件を満たした: 到達を記録し、結末と報酬を渡す
//   - 足りない: 何が足りないかを教える
//
// 呼び出し側はs.muを保持していること。
func (s *Server) talkEndingLocked(player *Player, npc *NPC) {
	e := npc.Ending
	if e == nil || s.world == nil {
		return
	}
	locale := s.localeOfLocked(player.Name)
	play := func() {
		for _, line := range e.Text {
			s.sendEndingLocked(player.Name, line.Get(locale))
		}
	}

	if player.Endings[e.ID] {
		s.grantEndingRewardLocked(player, e)
		play()
		return
	}
	if missing := s.missingForEndingLocked(player, e, locale); len(missing) > 0 {
		// 何が足りないかは教えない。NPCの一言だけを返す。
		s.sendEndingLocked(player.Name, e.Hint.Get(locale))
		return
	}

	if player.Endings == nil {
		player.Endings = make(map[string]bool)
	}
	player.Endings[e.ID] = true
	s.grantEndingRewardLocked(player, e)
	log.Printf("ending reached: player=%q ending=%q", player.Name, e.ID)
	s.sendEndingLocked(player.Name, LocalizedText{
		"en": "=== ENDING: %s ===",
		"ja": "=== エンディング: %s ===",
	}.Format(locale, e.Name.Get(locale)))
	play()
	if reward := s.world.Items[e.RewardItem]; reward != nil {
		s.sendEndingLocked(player.Name, LocalizedText{
			"en": "You received: %s (see INVENTORY).",
			"ja": "報酬を受け取った: %s(INVENTORYで確認できる)。",
		}.Format(locale, reward.Name.Get(locale)))
	}
	s.sendEndingProgressLocked(player)
}
