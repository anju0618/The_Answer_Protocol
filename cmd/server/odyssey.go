package main

const (
	itemLotusFruit    = "item.lotus_fruit"
	itemSacredCattle  = "item.sacred_cattle"
	itemBagOfWinds    = "item.bag_of_winds"
	ithacaShoreRoomID = "loc.ody_ithaca_shore"

	windsCrewLoss = 3
)

func (s *Server) initializeCrewLocked(player *Player, fromRoomID string) {
	if player.RoomID != odysseyStartRoomID || s.world == nil || fromRoomID != s.world.StartRoomID {
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

func (s *Server) applyTakeConsequencesLocked(player *Player, name, itemID string) *flavor {
	switch itemID {
	case itemLotusFruit:
		s.respawnPlayerLocked(player, name, "lotus")
		return &flavor{key: "lotus", player: name}
	case itemSacredCattle:
		s.respawnPlayerLocked(player, name, "cattle")
		return &flavor{key: "cattle", player: name}
	}
	return nil
}

func (s *Server) applyDropConsequencesLocked(player *Player, name, itemID string) *flavor {
	if itemID != itemBagOfWinds || player.RoomID == ithacaShoreRoomID {
		return nil
	}
	lost := spendCrewLocked(player, windsCrewLoss)
	return &flavor{key: "winds_opened", player: name, n: lost}
}
