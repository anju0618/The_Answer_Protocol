// オデュッセイア編だけに存在するアイテムの特殊な副作用(ロトスの実・
// ヘリオスの牛・風の革袋)と、クルーリソースの初期化・消費処理。
// これらは世界データ(汎用の myth_requirement 等)では表現しきれない、
// アイテムID決め打ちの特殊ケースなのでここにまとめている(memo.md 7.6, 7.9)。
package main

import "fmt"

const (
	itemLotusFruit    = "item.lotus_fruit"
	itemSacredCattle  = "item.sacred_cattle"
	itemBagOfWinds    = "item.bag_of_winds"
	ithacaShoreRoomID = "loc.ody_ithaca_shore"

	windsCrewLoss = 3
)

// initializeCrewLocked は、プレイヤーが初めてオデュッセイア編の開始地点
// (odysseyStartRoomID)に足を踏み入れた瞬間にクルーの初期値を設定する。
// 一度きりの処理で、CrewInitialized フラグで二重付与を防いでいる。
func (s *Server) initializeCrewLocked(player *Player) {
	if player.CrewInitialized || player.RoomID != odysseyStartRoomID {
		return
	}
	player.Crew = startingCrew
	player.CrewInitialized = true
}

// spendCrewLocked は player のクルーを amount 消費し、実際に消費できた
// 人数を返す(残クルーが amount 未満なら残り全員を消費して打ち止め、
// マイナスにはならない)。
func spendCrewLocked(player *Player, amount int) int {
	if amount > player.Crew {
		amount = player.Crew
	}
	player.Crew -= amount
	return amount
}

// applyTakeConsequencesLocked は、特定アイテムをTAKEした際の追加の
// 結末(ロトスの実=即死、ヘリオスの牛=即死)を適用する。該当しない
// アイテムなら何もせず空文字列を返す。
func (s *Server) applyTakeConsequencesLocked(player *Player, name, itemID string) string {
	switch itemID {
	case itemLotusFruit:
		s.respawnPlayerLocked(player, name)
		return fmt.Sprintf("%s tastes the lotus, and forgets there ever was a home to return to.", name)
	case itemSacredCattle:
		s.respawnPlayerLocked(player, name)
		return fmt.Sprintf("%s lays a hand on the cattle of Helios, and the sky answers.", name)
	}
	return ""
}

// applyDropConsequencesLocked は、風の革袋をイタケの浜以外でDROPした
// (=早まって開けてしまった)場合にクルーを消費させる。それ以外の
// アイテム/場所では何もしない。
func (s *Server) applyDropConsequencesLocked(player *Player, name, itemID string) string {
	if itemID != itemBagOfWinds || player.RoomID == ithacaShoreRoomID {
		return ""
	}
	lost := spendCrewLocked(player, windsCrewLoss)
	return fmt.Sprintf("%s opens the bag of winds too soon, and a storm drives the ship back, costing %d crew.", name, lost)
}
