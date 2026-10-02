package main

type RoomHazard struct {
	Type           string `json:"type"`
	RequiredItemID string `json:"required_item_id,omitempty"`
	CrewLoss       int    `json:"crew_loss,omitempty"`
	MinPartyTotal  int    `json:"min_party_total,omitempty"`
}

type Room struct {
	ID          string            `json:"id"`
	Name        LocalizedText     `json:"name"`
	Description LocalizedText     `json:"description"`
	Exits       map[string]string `json:"exits"`
	// SecretExits are exits that only exist for players who have reached all the listed endings.
	SecretExits map[string]SecretExit `json:"secret_exits,omitempty"`
	Hazard      *RoomHazard           `json:"hazard,omitempty"`
}

// SecretExit is an exit that opens once the player has reached every ending in RequiresEndings.
type SecretExit struct {
	Room            string   `json:"room"`
	RequiresEndings []string `json:"requires_endings"`
}

// exitsFor returns the exits this player can see and use: the normal ones plus any secret exit they have unlocked.
func (r *Room) exitsFor(player *Player) map[string]string {
	exits := make(map[string]string, len(r.Exits)+len(r.SecretExits))
	for dir, dest := range r.Exits {
		exits[dir] = dest
	}
	for dir, secret := range r.SecretExits {
		unlocked := true
		for _, endingID := range secret.RequiresEndings {
			if !player.Endings[endingID] {
				unlocked = false
			}
		}
		if unlocked {
			exits[dir] = secret.Room
		}
	}
	return exits
}
