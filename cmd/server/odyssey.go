// オデュッセイア編だけに存在するアイテムの特殊な副作用(ロトスの実・
// ヘリオスの牛・風の革袋)と、クルーリソースの初期化・消費処理。
// これらは世界データ(汎用の myth_requirement 等)では表現しきれない、
// アイテムID決め打ちの特殊ケースなのでここにまとめている(memo.md 7.6, 7.9)。
package main

const (
	itemLotusFruit    = "item.lotus_fruit"
	itemSacredCattle  = "item.sacred_cattle"
	itemBagOfWinds    = "item.bag_of_winds"
	ithacaShoreRoomID = "loc.ody_ithaca_shore"

	windsCrewLoss = 3
)

// initializeCrewLocked は、プレイヤーが運命の間(ハブ)からオデュッセイア編の
// 開始地点(odysseyStartRoomID)に入った瞬間に、クルーを startingCrew に戻す。
// 航海をやり直すたびに仲間が満員に戻るので、クルーを減らしすぎても
// 永久に詰むことはない(以前は初回の1回だけ付与していた)。開始地点へ
// 途中の部屋から戻ってきただけでは戻らない(fromRoomIDで判定)ため、
// 行き来してクルーを増やすことはできない。fromRoomID は移動前の部屋。
func (s *Server) initializeCrewLocked(player *Player, fromRoomID string) {
	if player.RoomID != odysseyStartRoomID || s.world == nil || fromRoomID != s.world.StartRoomID {
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
// アイテムなら何もせずnilを返す。
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

// applyDropConsequencesLocked は、風の革袋をイタケの浜以外でDROPした
// (=早まって開けてしまった)場合にクルーを消費させる。それ以外の
// アイテム/場所では何もせずnilを返す。
func (s *Server) applyDropConsequencesLocked(player *Player, name, itemID string) *flavor {
	if itemID != itemBagOfWinds || player.RoomID == ithacaShoreRoomID {
		return nil
	}
	lost := spendCrewLocked(player, windsCrewLoss)
	return &flavor{key: "winds_opened", player: name, n: lost}
}
