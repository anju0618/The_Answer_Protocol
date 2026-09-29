// ATTACK/FLEEコマンドと、戦闘まわりの共通処理(ダメージ計算・死亡と
// リスポーン)。神話ゲート(必要アイテム/クエスト)・FLEEの成否・
// Unwinnableな敵の扱いなど、戦闘システムの設計判断はREADME
// 「Combat System」とmemo.md 7章にまとめてある。
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"strings"
)

// randDamage は [min, max] の範囲(両端含む)でランダムなダメージ量を返す。
func randDamage(min, max int) int {
	return min + rand.IntN(max-min+1)
}

// respawnPlayerLocked はプレイヤーを死亡させ、ワールドの安全な開始地点
// (通常は運命の間)にHP respawnHP で送り返す。ATTACK・FLEE失敗・
// 神話ゲート・部屋ハザードなど、あらゆる「死」の共通処理としてここに
// まとめている。呼び出し側はs.muを保持していること。
func (s *Server) respawnPlayerLocked(player *Player, name string) {
	oldRoomID := player.RoomID
	destination := defaultStartRoomID
	if s.world != nil {
		destination = s.world.StartRoomID
	}
	player.HP = respawnHP
	player.CombatTargetID = ""
	player.RoomID = destination
	if oldRoomID == destination {
		return
	}
	for playerName, current := range s.players {
		recipient := s.clients[playerName]
		if recipient == nil {
			continue
		}
		switch current.RoomID {
		case oldRoomID:
			recipient.enqueueEvent("EVT ROOM PRESENCE LEAVE " + name)
		case destination:
			recipient.enqueueEvent("EVT ROOM PRESENCE ENTER " + name)
		}
	}
}

// combatResult はATTACKのJSONレスポンス本体(RFC 5.4.5 の形式)。
type combatResult struct {
	AttackerHP int    `json:"attacker_hp"`
	TargetHP   int    `json:"target_hp"`
	Damage     int    `json:"damage"`
	Status     string `json:"status"`
}

// handleAttack はATTACKコマンドを処理する。対象NPCの状態に応じて
// 3パターンに分岐する:
//  1. Unwinnable(倒せない敵): ダメージは発生させず、代わりにクルーを消費する。
//  2. 神話ゲート未達成: 即死させる。
//  3. それ以外: 通常の数値戦闘(固定範囲のランダムダメージ+反撃)を1ラウンド進める。
func handleAttack(s *Server, conn net.Conn, name *string, parts []string) bool {
	if !requireArgs(conn, parts, 2) {
		return false
	}
	if *name == "" {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	query := strings.Join(parts[1:], " ")
	locale := clientLocale(conn)

	s.mu.Lock()
	player := s.playerForUpdateLocked(*name)
	if player == nil || s.world == nil || s.world.Rooms[player.RoomID] == nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	npcID := s.world.resolveNPCInRoom(player.RoomID, query, locale)
	if npcID == "" {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 404 NPC_NOT_FOUND")
		return false
	}
	npc := s.world.NPCs[npcID]
	if npc.Role != "enemy" || npc.HP <= 0 {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 405 NPC_NOT_HOSTILE")
		return false
	}
	encounterRoomID := player.RoomID

	var resultText string
	var result combatResult

	switch {
	case npc.Unwinnable:
		lost := spendCrewLocked(player, npc.CrewLossOnAttack)
		player.CombatTargetID = ""
		resultText = fmt.Sprintf("%s attacks %s and is driven back, losing %d crew.", *name, npc.Name.Get(locale), lost)
		result = combatResult{player.HP, npc.HP, 0, "overwhelmed"}

	case npc.hasMythRequirement() && !player.meetsMythRequirement(npc):
		s.respawnPlayerLocked(player, *name)
		resultText = fmt.Sprintf("%s attacks %s unprepared and is killed.", *name, npc.Name.Get(locale))
		result = combatResult{0, npc.HP, 0, "dead"}

	default:
		damage := randDamage(combatMinDamage, combatMaxDamage)
		npc.HP -= damage
		if npc.HP < 0 {
			npc.HP = 0
		}
		if npc.HP == 0 {
			player.CombatTargetID = ""
			s.checkQuestObjectiveLocked(player, "defeat_npc", npcID)
			resultText = fmt.Sprintf("%s defeats %s.", *name, npc.Name.Get(locale))
			result = combatResult{player.HP, 0, damage, "victory"}
		} else {
			player.CombatTargetID = npcID
			counter := randDamage(counterMinDamage, counterMaxDamage)
			player.HP -= counter
			if player.HP <= 0 {
				s.respawnPlayerLocked(player, *name)
				resultText = fmt.Sprintf("%s is struck down by %s.", *name, npc.Name.Get(locale))
				result = combatResult{0, npc.HP, damage, "dead"}
			} else {
				resultText = fmt.Sprintf("%s attacks %s for %d damage and takes %d in return.", *name, npc.Name.Get(locale), damage, counter)
				result = combatResult{player.HP, npc.HP, damage, "combat"}
			}
		}
	}

	data, err := json.Marshal(result)
	if err != nil {
		s.mu.Unlock()
		log.Printf("encode ATTACK response: %v", err)
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	client := conn.(*serverClient)
	response, err := client.enqueueResponse("OK " + string(data))
	if err == nil {
		s.broadcastRoomEventLocked(encounterRoomID, "EVT ROOM COMBAT "+resultText)
	}
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}

// handleFlee はFLEEコマンド(独自拡張、引数なし)を処理する。現在戦闘中の
// NPCから逃げようとする。成否はNPCごとの FleeAccurate(常に成功)/
// FleeSucceedsOnce(そのNPCから初めて逃げる時だけ成功)で決まり、成功時は
// Player.FledFrom に記録される(部屋封鎖の解除判定にも使われる、
// hazard.go の blockingEnemyLocked を参照)。失敗すると反撃を受ける。
func handleFlee(s *Server, conn net.Conn, name *string, parts []string) bool {
	if !requireExactArgs(conn, parts, 1) {
		return false
	}
	if *name == "" {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	locale := clientLocale(conn)

	s.mu.Lock()
	player := s.playerForUpdateLocked(*name)
	if player == nil || s.world == nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	npc := s.world.NPCs[player.CombatTargetID]
	if player.CombatTargetID == "" || npc == nil {
		player.CombatTargetID = ""
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 407 NOT_IN_COMBAT")
		return false
	}
	encounterRoomID := player.RoomID

	targetID := player.CombatTargetID
	fleeSucceeds := npc.FleeAccurate || (npc.FleeSucceedsOnce && !player.FledFrom[targetID])

	var resultText, result string
	if fleeSucceeds {
		result = "success"
		resultText = fmt.Sprintf("%s flees from %s.", *name, npc.Name.Get(locale))
		player.CombatTargetID = ""
		if player.FledFrom == nil {
			player.FledFrom = make(map[string]bool)
		}
		player.FledFrom[targetID] = true
	} else {
		counter := randDamage(counterMinDamage, counterMaxDamage)
		player.HP -= counter
		if player.HP <= 0 {
			s.respawnPlayerLocked(player, *name)
			result = "failure_dead"
			resultText = fmt.Sprintf("%s tries to flee %s and is cut down.", *name, npc.Name.Get(locale))
		} else {
			result = "failure"
			resultText = fmt.Sprintf("%s tries to flee %s and is struck for %d.", *name, npc.Name.Get(locale), counter)
		}
	}

	data, err := json.Marshal(struct {
		HP     int    `json:"hp"`
		Result string `json:"result"`
	}{player.HP, result})
	if err != nil {
		s.mu.Unlock()
		log.Printf("encode FLEE response: %v", err)
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	client := conn.(*serverClient)
	response, err := client.enqueueResponse("OK " + string(data))
	if err == nil {
		s.broadcastRoomEventLocked(encounterRoomID, "EVT ROOM COMBAT "+resultText)
	}
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}
