package main

import "time"

type PlayerQuest struct {
	Status   string `json:"status"`
	Progress int    `json:"progress"`
}

type Player struct {
	Name             string                  `json:"name"`
	HP               int                     `json:"hp"`
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

func (p *Player) regenLocked(now time.Time) {
	if p.HP >= maxPlayerHP || p.lastRegen.IsZero() {
		p.lastRegen = now
		return
	}
	ticks := int(now.Sub(p.lastRegen) / regenInterval)
	if ticks <= 0 {
		return
	}
	p.HP += ticks * regenAmount
	if p.HP >= maxPlayerHP {
		p.HP = maxPlayerHP
		p.lastRegen = now
		return
	}
	p.lastRegen = p.lastRegen.Add(time.Duration(ticks) * regenInterval)
}
