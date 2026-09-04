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
