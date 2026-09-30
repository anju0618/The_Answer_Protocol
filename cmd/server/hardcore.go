// 鬼畜難易度(死ぬと持ち物を失う)と、それを和らげる協力要素(GROUPの
// 仲間が同じ部屋にいると助け合える)をまとめたファイル。
//
//   - 死ぬと、報酬アイテム(エンディングのクリア記念品。Item.RewardOnly)以外の
//     持ち物をすべて失う。一意のアイテムは元の部屋に戻り、Renewableな鍵アイテムの
//     コピーは消える(また取りに行ける)。敵のHPと「逃げ切った」記録もリセット
//     され、敵は全快する。
//   - ただし同じGROUPの仲間が同じ部屋にいれば、仲間が持ち物を守ってくれて
//     何も失わない(協力して進む動機)。
//   - 仲間が同じ部屋にいるATTACKは、味方1人につきダメージが増え、反撃が
//     分散して弱まり、倒した敵の討伐(クエスト進捗を含む)は全員の手柄になる。
package main

import "log"

const (
	// allyDamageBonus は同じ部屋の味方1人ごとに加わるATTACKのダメージ。
	allyDamageBonus = 5
	// allyCounterReductionPercent は味方1人ごとに減る反撃ダメージの割合(%)。
	allyCounterReductionPercent = 20
	// maxAllyBonusCount は効果が加算される味方の人数の上限。
	maxAllyBonusCount = 3
)

// alliesInRoomLocked は name と同じGROUPで、同じ部屋にいる他のプレイヤーの
// 名前(名前順ではなく順不同)を返す。GROUPに入っていなければ空。
func (s *Server) alliesInRoomLocked(name string) []string {
	player := s.players[name]
	group := s.groups[s.groupByPlayer[name]]
	if player == nil || group == nil {
		return nil
	}
	var allies []string
	for memberName := range group.Members {
		member := s.players[memberName]
		if memberName != name && member != nil && !member.exiting && member.RoomID == player.RoomID {
			allies = append(allies, memberName)
		}
	}
	return allies
}

// allyBonusCount は加算対象になる味方の人数(上限あり)を返す。
func allyBonusCount(allies []string) int {
	if len(allies) > maxAllyBonusCount {
		return maxAllyBonusCount
	}
	return len(allies)
}

// deathOutcome は死亡時の持ち物の扱い。notifyDeathLocked が文言を選ぶのに使う。
type deathOutcome string

const (
	outcomeNothingLost deathOutcome = ""     // 失うものが元から無かった
	outcomeLost        deathOutcome = "lost" // 持ち物を失った
	outcomeKept        deathOutcome = "kept" // 仲間が守ってくれた
)

// applyDeathPenaltyLocked は死亡時のペナルティ(持ち物の喪失、敵のリセット)を
// 適用して結果を返す。仲間が同じ部屋にいれば持ち物は守られる(敵のリセットは
// 起こる)。復活先への移動より前、つまり死んだ部屋にいるうちに呼ぶこと。
// 呼び出し側はs.muを保持していること。
func (s *Server) applyDeathPenaltyLocked(player *Player, name string) deathOutcome {
	protected := len(s.alliesInRoomLocked(name)) > 0
	player.EnemyHP = nil
	player.FledFrom = nil
	if s.world == nil {
		return outcomeNothingLost
	}

	// 失う対象は「報酬アイテム以外の持ち物」。
	var kept, lost []string
	for _, itemID := range player.Inventory {
		if item := s.world.Items[itemID]; item == nil || item.RewardOnly {
			kept = append(kept, itemID)
		} else {
			lost = append(lost, itemID)
		}
	}
	if len(lost) == 0 {
		return outcomeNothingLost
	}
	if protected {
		return outcomeKept
	}

	player.Inventory = kept
	var returned []string
	for _, itemID := range lost {
		if !s.world.Items[itemID].Renewable {
			returned = append(returned, itemID)
		}
	}
	s.returnItemsHomeLocked(name, returned)
	return outcomeLost
}

// returnItemsHomeLocked は失われた一意のアイテムを元の部屋へ戻し、その
// 位置をディスクにも反映する(再起動後に取り残されないように)。
func (s *Server) returnItemsHomeLocked(name string, itemIDs []string) {
	if len(itemIDs) == 0 {
		return
	}
	for _, itemID := range itemIDs {
		item := s.world.Items[itemID]
		item.RoomID = item.HomeRoomID
		if item.RoomID == "" {
			item.RoomID = s.world.StartRoomID
		}
		delete(s.unsavedTakes[name], itemID)
	}
	s.ioMu.Lock()
	defer s.ioMu.Unlock()
	locations, err := s.loadItemLocations()
	if err != nil {
		log.Printf("return lost items of %q: %v", name, err)
		return
	}
	for _, itemID := range itemIDs {
		locations[itemID] = s.world.Items[itemID].RoomID
	}
	if err := s.writeItemLocations(locations); err != nil {
		log.Printf("save returned items of %q: %v", name, err)
	}
}

// shareVictoryLocked は attacker が倒した敵 npcID の討伐を、同じ部屋の
// 味方全員の手柄にする(その味方の世界でも敵を倒し済みにし、defeat_npc
// クエストも進める)。呼び出し側はs.muを保持していること。
func (s *Server) shareVictoryLocked(attacker string, allies []string, npcID string, npc *NPC) {
	for _, allyName := range allies {
		ally := s.players[allyName]
		if ally == nil || ally.exiting || ally.enemyHP(npcID, npc) <= 0 {
			continue
		}
		ally.setEnemyHP(npcID, 0)
		ally.CombatTargetID = ""
		s.checkQuestObjectiveLocked(ally, "defeat_npc", npcID)
		locale := s.localeOfLocked(allyName)
		s.sendPlayerEventLocked(allyName, "TEAM", LocalizedText{
			"en": "Your ally %s defeated %s, and you share the victory.",
			"ja": "仲間の%sが%sを打ち倒した。その手柄はあなたにも与えられる。",
		}.Format(locale, attacker, npc.Name.Get(locale)))
	}
}
