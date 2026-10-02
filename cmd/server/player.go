package main

import "time"

type PlayerQuest struct {
	Status   string `json:"status"`
	Progress int    `json:"progress"`
}

type Player struct {
	Name             string                  `json:"name"`
	HP               int                     `json:"hp"`
	MaxHPBonus       int                     `json:"max_hp_bonus,omitempty"` // earned from quests; kept after death
	RoomID           string                  `json:"room_id"`
	Inventory        []string                `json:"inventory"`
	Crew             int                     `json:"crew,omitempty"`
	CrewInitialized  bool                    `json:"crew_initialized,omitempty"`
	IntroSeen        bool                    `json:"intro_seen,omitempty"`
	LastDeathSubject string                  `json:"last_death_subject,omitempty"`
	CombatTargetID   string                  `json:"combat_target_id,omitempty"`
	FledFrom         map[string]bool         `json:"fled_from,omitempty"`
	Quests           map[string]*PlayerQuest `json:"quests,omitempty"`
	Endings          map[string]bool         `json:"endings,omitempty"`
	EnemyHP          map[string]int          `json:"enemy_hp,omitempty"`
	lastRegen        time.Time
	guarding         bool // braced with DEFEND: the next counter-attack is halved (not saved)
	exiting          bool
}

func (p *Player) hasItem(itemID string) bool {
	for _, id := range p.Inventory {
		if id == itemID {
			return true
		}
	}
	return false
}

func (p *Player) meetsMythRequirement(npc *NPC) bool {
	if npc.MythRequirementItem != "" && !p.hasItem(npc.MythRequirementItem) {
		return false
	}
	if npc.MythRequirementQuest != "" {
		state := p.Quests[npc.MythRequirementQuest]
		if state == nil || state.Status != "completed" {
			return false
		}
	}
	return true
}

func (npc *NPC) hasMythRequirement() bool {
	return npc.MythRequirementItem != "" || npc.MythRequirementQuest != ""
}

func (p *Player) enemyHP(npcID string, npc *NPC) int {
	if hp, ok := p.EnemyHP[npcID]; ok {
		return hp
	}
	return npc.HP
}

func (p *Player) setEnemyHP(npcID string, hp int) {
	if p.EnemyHP == nil {
		p.EnemyHP = make(map[string]int)
	}
	p.EnemyHP[npcID] = hp
}

const (
	regenInterval = 2 * time.Second
	regenAmount   = 1
)

// regenLocked heals over time up to maxHP; bonus is extra HP per tick (Hera's blessing, items), possibly negative.
func (p *Player) regenLocked(now time.Time, bonus, maxHP int) {
	if p.HP > maxHP {
		p.HP = maxHP // an item that raised max HP was dropped
	}
	if p.HP >= maxHP || p.lastRegen.IsZero() {
		p.lastRegen = now
		return
	}
	ticks := int(now.Sub(p.lastRegen) / regenInterval)
	if ticks <= 0 {
		return
	}
	p.HP += ticks * max(0, regenAmount+bonus)
	if p.HP >= maxHP {
		p.HP = maxHP
		p.lastRegen = now
		return
	}
	p.lastRegen = p.lastRegen.Add(time.Duration(ticks) * regenInterval)
}
