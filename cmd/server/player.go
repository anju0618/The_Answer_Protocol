package main

type PlayerQuest struct {
	Status   string `json:"status"`
	Progress int    `json:"progress"`
}

type Player struct {
	Name            string                  `json:"name"`
	HP              int                     `json:"hp"`
	RoomID          string                  `json:"room_id"`
	Inventory       []string                `json:"inventory"`
	Crew            int                     `json:"crew,omitempty"`
	CrewInitialized bool                    `json:"crew_initialized,omitempty"`
	CombatTargetID  string                  `json:"combat_target_id,omitempty"`
	FledFrom        map[string]bool         `json:"fled_from,omitempty"`
	Quests          map[string]*PlayerQuest `json:"quests,omitempty"`
	exiting         bool
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
