package main

type RoomHazard struct {
	Type           string `json:"type"`
	RequiredItemID string `json:"required_item_id,omitempty"`
	CrewLoss       int    `json:"crew_loss,omitempty"`
	MinPartyTotal  int    `json:"min_party_total,omitempty"`
}

type Room struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       map[string]string `json:"exits"`
	Hazard      *RoomHazard       `json:"hazard,omitempty"`
}
