package main

type Player struct {
	Name      string   `json:"name"`
	HP        int      `json:"hp"`
	RoomID    string   `json:"room_id"`
	Inventory []string `json:"inventory"`
}
