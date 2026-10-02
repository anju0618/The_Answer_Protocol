package main

import "fmt"

// Items change the player while they are carried (not while they lie on the floor). Unlike a god's blessing,
// an item effect can be bad, so Value may be negative. It is a separate system from Ending.Blessing.
const (
	effectMaxHP = "max_hp" // Value more (or fewer) max HP while carried

	// A quest's reward HP is turned into a permanent max HP gain: reward / questMaxHPDivisor (at least 1).
	questMaxHPDivisor = 5
	// minPlayerMaxHP keeps a bad item from pushing max HP under the HP a player respawns with.
	minPlayerMaxHP = 20
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

// itemEffectTotalLocked adds up one effect over everything the player is carrying.
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

// effectTotalLocked is the player's total for one effect: god blessings plus carried items.
func (s *Server) effectTotalLocked(player *Player, effect string) int {
	return s.blessingTotalLocked(player, effect) + s.itemEffectTotalLocked(player, effect)
}

// maxHPLocked is the player's current max HP: the base, quest gains (kept after death) and carried items.
func (s *Server) maxHPLocked(player *Player) int {
	return max(minPlayerMaxHP, maxPlayerHP+player.MaxHPBonus+s.itemEffectTotalLocked(player, effectMaxHP))
}

// questMaxHPGain is how much max HP finishing the quest gives.
func questMaxHPGain(quest *Quest) int {
	return max(1, quest.Reward.HP/questMaxHPDivisor)
}
