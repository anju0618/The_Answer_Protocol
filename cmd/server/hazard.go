// 部屋そのものに紐づく危険を扱う。NPCとの戦闘によるものではなく、
// MOVEで入室した瞬間/その部屋に生きた敵がいる間に発生する処理
// (memo.md 7.5〜7.6、7章末尾の「部屋封鎖」を参照)。
package main

// applyRoomHazardLocked は、MOVEで room に入室した直後にその部屋自身の
// Hazard(4種: lethal即死・item_gate要アイテム・crew_gate閾値付きクルー
// 消費・crew_cost無条件クルー消費)を評価する。プレイヤーを死なせたり
// クルーを消費させたりした場合、部屋全体に流す実況(flavor)を返す
// (何も起きなければnil)。呼び出し側はs.muを保持していること。
func (s *Server) applyRoomHazardLocked(player *Player, name string, room *Room, locale string) *flavor {
	if room == nil || room.Hazard == nil {
		return nil
	}
	hazard := room.Hazard
	roomName := room.Name.Get(locale)

	switch hazard.Type {
	case "lethal":
		s.respawnPlayerLocked(player, name, "hazard_lethal", roomName)
		return &flavor{key: "hazard_lethal", player: name, room: room}

	case "item_gate":
		if player.hasItem(hazard.RequiredItemID) {
			return nil
		}
		s.respawnPlayerLocked(player, name, "hazard_item", roomName, s.world.Items[hazard.RequiredItemID].Name.Get(locale))
		return &flavor{key: "hazard_item", player: name, room: room}

	case "crew_gate":
		if player.Crew+1 < hazard.MinPartyTotal {
			s.respawnPlayerLocked(player, name, "hazard_crew", roomName, hazard.MinPartyTotal, player.Crew+1)
			return &flavor{key: "hazard_crew_dead", player: name, room: room}
		}
		lost := spendCrewLocked(player, hazard.CrewLoss)
		return &flavor{key: "hazard_crew_loss", player: name, room: room, n: lost}

	case "crew_cost":
		lost := spendCrewLocked(player, hazard.CrewLoss)
		return &flavor{key: "hazard_crew_loss", player: name, room: room, n: lost}
	}
	return nil
}

// blockingEnemyLocked は roomID 内に「player から見て生きていて(HP>0)、
// player がまだ倒しても振り切ってもいない」role:"enemy"のNPCがいるかを
// 調べる。いればそのNPCのID・実体を返し(複数いればID辞書順で最小のもの)、
// いなければ("", nil) を返す。MOVEでこの部屋を出ようとした際、ここで
// 見つかったNPCがあればその移動は即死になる(server.go の handleMove を
// 参照)。敵のHPはプレイヤーごとなので(Player.enemyHP)、他のプレイヤーが
// 倒したかどうかは関係ない。「振り切った」はPlayer.FledFrom(FLEE成功時に
// 記録)で判定するため、Unwinnableな敵(例: ライストリュゴネス族)は
// FLEEでしか部屋を出られない。
func (s *Server) blockingEnemyLocked(player *Player, roomID string) (string, *NPC) {
	blockID := ""
	for id, npc := range s.world.NPCs {
		if npc == nil || npc.RoomID != roomID || npc.Role != "enemy" || player.enemyHP(id, npc) <= 0 {
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
