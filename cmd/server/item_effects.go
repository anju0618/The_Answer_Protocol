package main

import "fmt"

const (
	effectMaxHP = "max_hp"

	questMaxHPDivisor = 5
	minPlayerMaxHP    = 20
)

type ItemEffect struct {
	Effect string `json:"effect"`
	Value  int    `json:"value"`
}

func validItemEffect(effect string) bool {
	switch effect {
	case effectMaxHP, blessingCounterReduction, blessingRegenBonus, blessingDamageBonus:
		return true
	}
	return false
}

func (w *World) validateItemEffects() error {
	for id, item := range w.Items {
		if item == nil {
			continue
		}
		for _, e := range item.Effects {
			if !validItemEffect(e.Effect) {
				return fmt.Errorf("item %q has unknown effect %q", id, e.Effect)
			}
			if e.Value == 0 {
				return fmt.Errorf("item %q effect %q has value 0", id, e.Effect)
			}
		}
	}
	return nil
}

func (s *Server) itemEffectTotalLocked(player *Player, effect string) int {
	if s.world == nil {
		return 0
	}
	total := 0
	for _, itemID := range player.Inventory {
		if item := s.world.Items[itemID]; item != nil {
			for _, e := range item.Effects {
				if e.Effect == effect {
					total += e.Value
				}
			}
		}
	}
	return total
}

func (s *Server) effectTotalLocked(player *Player, effect string) int {
	return s.blessingTotalLocked(player, effect) + s.itemEffectTotalLocked(player, effect)
}

func (s *Server) maxHPLocked(player *Player) int {
	return max(minPlayerMaxHP, maxPlayerHP+player.MaxHPBonus+s.itemEffectTotalLocked(player, effectMaxHP))
}

func questMaxHPGain(quest *Quest) int {
	return max(1, quest.Reward.HP/questMaxHPDivisor)
}
