package main

import (
	"fmt"
	"sort"
	"strings"
)

// Blessing is a permanent boon from the god of an arc, earned by reaching that arc's ending.
// It is derived from the endings a player has reached, so nothing extra is stored on the player.
type Blessing struct {
	God         LocalizedText `json:"god"`
	Name        LocalizedText `json:"name"`
	Description LocalizedText `json:"description"`
	Effect      string        `json:"effect"`
	Value       int           `json:"value"`
}

const (
	blessingCounterReduction = "counter_reduction" // enemy counter-attacks hurt Value% less
	blessingRegenBonus       = "regen_bonus"       // Value extra HP every regen tick
	blessingDamageBonus      = "damage_bonus"      // Value extra damage on every hit
	maxCounterIncrease       = 50                  // bad items never make counter-attacks hurt more than 50% extra
	maxCounterReduction      = 80                  // allies (3 x 20%) plus a blessing never reach 100%
)

type Ending struct {
	ID              string          `json:"id"`
	Name            LocalizedText   `json:"name"`
	RequiresItems   []string        `json:"requires_items,omitempty"`
	RequiresQuests  []string        `json:"requires_quests,omitempty"`
	RequiresEndings []string        `json:"requires_endings,omitempty"`
	RewardItem      string          `json:"reward_item,omitempty"`
	Blessing        *Blessing       `json:"blessing,omitempty"`
	Hint            LocalizedText   `json:"hint"`
	Text            []LocalizedText `json:"text"`
}

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
		if b := e.Blessing; b != nil {
			switch b.Effect {
			case blessingCounterReduction, blessingRegenBonus, blessingDamageBonus:
			default:
				return fmt.Errorf("ending %q blessing has unknown effect %q", e.ID, b.Effect)
			}
			if b.Value < 1 || b.God["en"] == "" || b.Name["en"] == "" || b.Name["ja"] == "" || b.God["ja"] == "" {
				return fmt.Errorf("ending %q blessing needs a positive value and en/ja god and name", e.ID)
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

func (w *World) endingByID(id string) *Ending {
	for _, npc := range w.NPCs {
		if npc != nil && npc.Ending != nil && npc.Ending.ID == id {
			return npc.Ending
		}
	}
	return nil
}

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

// blessingTotalLocked adds up one effect over every blessing the player has earned.
func (s *Server) blessingTotalLocked(player *Player, effect string) int {
	if s.world == nil {
		return 0
	}
	total := 0
	for _, npc := range s.world.NPCs {
		if npc == nil || npc.Ending == nil || npc.Ending.Blessing == nil {
			continue
		}
		if b := npc.Ending.Blessing; b.Effect == effect && player.Endings[npc.Ending.ID] {
			total += b.Value
		}
	}
	return total
}

func (s *Server) sendEndingLocked(name, text string) {
	s.sendPlayerEventLocked(name, "ENDING", text)
}

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

func (s *Server) grantEndingRewardLocked(player *Player, e *Ending) {
	if e.RewardItem != "" && !player.hasItem(e.RewardItem) {
		player.Inventory = append(player.Inventory, e.RewardItem)
	}
}

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

		s.sendEndingLocked(player.Name, e.Hint.Get(locale))
		return
	}

	if player.Endings == nil {
		player.Endings = make(map[string]bool)
	}
	player.Endings[e.ID] = true
	s.grantEndingRewardLocked(player, e)
	logger.Info("ending_reached", "player", player.Name, "ending", e.ID)
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
	if b := e.Blessing; b != nil {
		s.sendEndingLocked(player.Name, LocalizedText{
			"en": "%s grants you her blessing, for good: %s. %s",
			"ja": "%sの祝福を受けた(永続): %s。%s",
		}.Format(locale, b.God.Get(locale), b.Name.Get(locale), b.Description.Get(locale)))
	}
	s.sendEndingProgressLocked(player)
}
