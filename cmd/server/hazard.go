// 部屋そのものに紐づく危険を扱う。NPCとの戦闘によるものではなく、
// MOVEで入室した瞬間/その部屋に生きた敵がいる間に発生する処理
// (memo.md 7.5〜7.6、7章末尾の「部屋封鎖」を参照)。
package main

import "fmt"

// applyRoomHazardLocked は、MOVEで room に入室した直後にその部屋自身の
// Hazard(4種: lethal即死・item_gate要アイテム・crew_gate閾値付きクルー
// 消費・crew_cost無条件クルー消費)を評価する。プレイヤーを死なせたり
// クルーを消費させたりした場合、ブロードキャスト用の説明文を返す
// (何も起きなければ空文字列)。呼び出し側はs.muを保持していること。
func (s *Server) applyRoomHazardLocked(player *Player, name string, room *Room, locale string) string {
	if room == nil || room.Hazard == nil {
		return ""
	}
	hazard := room.Hazard
	roomName := room.Name.Get(locale)

	switch hazard.Type {
	case "lethal":
		s.respawnPlayerLocked(player, name)
		return fmt.Sprintf("%s is lost to %s.", name, roomName)

	case "item_gate":
		if player.hasItem(hazard.RequiredItemID) {
			return ""
		}
		s.respawnPlayerLocked(player, name)
		return fmt.Sprintf("%s, unprepared, does not survive %s.", name, roomName)

	case "crew_gate":
		if player.Crew+1 < hazard.MinPartyTotal {
			s.respawnPlayerLocked(player, name)
			return fmt.Sprintf("%s and the remaining crew are lost passing %s.", name, roomName)
		}
		lost := spendCrewLocked(player, hazard.CrewLoss)
		return fmt.Sprintf("%s passes %s, losing %d crew.", name, roomName, lost)

	case "crew_cost":
		lost := spendCrewLocked(player, hazard.CrewLoss)
		return fmt.Sprintf("%s passes %s, losing %d crew.", name, roomName, lost)
	}
	return ""
}

// blockingEnemyLocked は roomID 内に「生きていて(HP>0)、player がまだ
// 倒しても振り切ってもいない」role:"enemy"のNPCがいるかを調べる。いれば
// そのNPCのID・実体を返し(複数いればID辞書順で最小のもの)、いなければ
// ("", nil) を返す。MOVEでこの部屋を出ようとした際、ここで見つかったNPCが
// あればその移動は即死になる(server.go の handleMove を参照)。
// 「振り切った」はPlayer.FledFrom(FLEE成功時に記録)で判定するため、
// Unwinnableな敵(例: ライストリュゴネス族)はFLEEでしか部屋を出られない。
func (s *Server) blockingEnemyLocked(player *Player, roomID string) (string, *NPC) {
	blockID := ""
	for id, npc := range s.world.NPCs {
		if npc == nil || npc.RoomID != roomID || npc.Role != "enemy" || npc.HP <= 0 {
			continue
		}
		if player.FledFrom[id] {
			continue
		}
		if blockID == "" || id < blockID {
			blockID = id
		}
	}
	if blockID == "" {
		return "", nil
	}
	return blockID, s.world.NPCs[blockID]
}
