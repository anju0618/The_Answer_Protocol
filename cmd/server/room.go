// data/world.json の「部屋(room)」に対応するデータモデル。
package main

// RoomHazard は部屋に入室した瞬間に発動する危険(即死/アイテム所持/クルー消費
// など)を表す。実際の判定処理は hazard.go の applyRoomHazardLocked が行う。
type RoomHazard struct {
	Type           string `json:"type"`
	RequiredItemID string `json:"required_item_id,omitempty"`
	CrewLoss       int    `json:"crew_loss,omitempty"`
	MinPartyTotal  int    `json:"min_party_total,omitempty"`
}

// Room はワールド内の1部屋。Name/Description は LocalizedText なので、
// LOOK応答としてそのまま返すことはできない(locale.go の roomView を介して
// 該当言語のプレーン文字列に解決してから返す)。
type Room struct {
	ID          string            `json:"id"`
	Name        LocalizedText     `json:"name"`
	Description LocalizedText     `json:"description"`
	Exits       map[string]string `json:"exits"`
	Hazard      *RoomHazard       `json:"hazard,omitempty"`
}
