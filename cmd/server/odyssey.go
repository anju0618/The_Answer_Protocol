package main

import "fmt"

// Odyssey-arc-specific item behavior (memo.md 7.6, 7.9). These are hardcoded
// by item ID rather than a generic data-driven field, since only three items
// need this treatment.
const (
	itemLotusFruit    = "item.lotus_fruit"
	itemSacredCattle  = "item.sacred_cattle"
	itemBagOfWinds    = "item.bag_of_winds"
	ithacaShoreRoomID = "loc.ody_ithaca_shore"

	lotusCrewLoss = 2
	windsCrewLoss = 3
)

// initializeCrewLocked grants the player their starting crew the first time
// they set foot in the Odyssey arc. Callers must hold s.mu.
func (s *Server) initializeCrewLocked(player *Player) {
	if player.CrewInitialized || player.RoomID != odysseyStartRoomID {
		return
	}
	player.Crew = startingCrew
	player.CrewInitialized = true
}

func spendCrewLocked(player *Player, amount int) int {
	if amount > player.Crew {
		amount = player.Crew
	}
	player.Crew -= amount
	return amount
}

// applyTakeConsequencesLocked applies fallout for picking up a "forbidden" or
// costly item. Callers must hold s.mu.
func (s *Server) applyTakeConsequencesLocked(player *Player, name, itemID string) string {
	switch itemID {
	case itemLotusFruit:
		lost := spendCrewLocked(player, lotusCrewLoss)
		return fmt.Sprintf("%s tastes the lotus, and %d of the crew sent after them stay behind too.", name, lost)
	case itemSacredCattle:
		s.respawnPlayerLocked(player, name)
		return fmt.Sprintf("%s lays a hand on the cattle of Helios, and the sky answers.", name)
	}
	return ""
}

// applyDropConsequencesLocked applies fallout for dropping an item in the
// wrong place. Callers must hold s.mu.
func (s *Server) applyDropConsequencesLocked(player *Player, name, itemID string) string {
	if itemID != itemBagOfWinds || player.RoomID == ithacaShoreRoomID {
		return ""
	}
	lost := spendCrewLocked(player, windsCrewLoss)
	return fmt.Sprintf("%s opens the bag of winds too soon, and a storm drives the ship back, costing %d crew.", name, lost)
}
