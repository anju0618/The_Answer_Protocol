package main

type RoomHazard struct {
	Type           string `json:"type"`
	RequiredItemID string `json:"required_item_id,omitempty"`
	CrewLoss       int    `json:"crew_loss,omitempty"`
	MinPartyTotal  int    `json:"min_party_total,omitempty"`
}

type Room struct {
	ID          string                `json:"id"`
	Name        LocalizedText         `json:"name"`
	Description LocalizedText         `json:"description"`
	Exits       map[string]string     `json:"exits"`
	SecretExits map[string]SecretExit `json:"secret_exits,omitempty"`
	Hazard      *RoomHazard           `json:"hazard,omitempty"`
}

type SecretExit struct {
	Room            string   `json:"room"`
	RequiresEndings []string `json:"requires_endings"`
}

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
