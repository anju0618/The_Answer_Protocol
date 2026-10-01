package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net"
	"strings"
)

var randDamage = func(min, max int) int {
	return min + rand.IntN(max-min+1)
}

// guardReductionPercent is how much a DEFEND stance cuts the next counter-attack.
const guardReductionPercent = 50

// takeGuard returns the extra reduction from a DEFEND stance and ends the stance: it covers one counter only.
func takeGuard(player *Player) int {
	if !player.guarding {
		return 0
	}
	player.guarding = false
	return guardReductionPercent
}

// reduceCounter lowers a counter-attack by percent (capped), never below 1 damage.
func reduceCounter(counter, percent int) int {
	percent = min(percent, maxCounterReduction)
	if percent <= 0 {
		return counter
	}
	return max(1, counter*(100-percent)/100)
}

// respawnPlayerLocked kills the player and sends them back to the hub.
// subject is the ID of the NPC, room or item that caused the death; the Moirai can later turn it into a hint.
func (s *Server) respawnPlayerLocked(player *Player, name, cause, subject string, args ...any) {
	player.LastDeathSubject = subject
	outcome := s.applyDeathPenaltyLocked(player, name)
	s.notifyDeathLocked(name, cause, outcome, args...)
	oldRoomID := player.RoomID
	logger.Info("player_died", "player", name, "cause", cause, "room", oldRoomID, "belongings", string(outcome))
	destination := defaultStartRoomID
	if s.world != nil {
		destination = s.world.StartRoomID
	}
	player.HP = respawnHP
	player.guarding = false
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

type combatResult struct {
	AttackerHP int    `json:"attacker_hp"`
	TargetHP   int    `json:"target_hp"`
	Damage     int    `json:"damage"`
	Status     string `json:"status"`
}

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
	enemyHP := player.enemyHP(npcID, npc)
	if npc.Role != "enemy" || enemyHP <= 0 {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 405 NPC_NOT_HOSTILE")
		return false
	}
	encounterRoomID := player.RoomID

	var event flavor
	var result combatResult

	switch {
	case npc.Unwinnable:
		lost := spendCrewLocked(player, npc.CrewLossOnAttack)
		player.CombatTargetID = ""
		event = flavor{key: "attack_unwinnable", player: *name, npc: npc, n: lost}
		result = combatResult{player.HP, enemyHP, 0, "overwhelmed"}

	case npc.hasMythRequirement() && !player.meetsMythRequirement(npc):
		s.respawnPlayerLocked(player, *name, "attack_unprepared", npcID, npc.Name.Get(locale))
		event = flavor{key: "attack_unprepared", player: *name, npc: npc}
		result = combatResult{0, enemyHP, 0, "dead"}

	default:

		allies := s.alliesInRoomLocked(*name)
		bonus := allyBonusCount(allies)
		damage := randDamage(combatMinDamage, combatMaxDamage) + bonus*allyDamageBonus + s.blessingTotalLocked(player, blessingDamageBonus)
		enemyHP -= damage
		if enemyHP < 0 {
			enemyHP = 0
		}
		player.setEnemyHP(npcID, enemyHP)
		if enemyHP == 0 {
			player.CombatTargetID = ""
			s.checkQuestObjectiveLocked(player, "defeat_npc", npcID)
			s.shareVictoryLocked(*name, allies, npcID, npc)
			event = flavor{key: "attack_defeat", player: *name, npc: npc}
			result = combatResult{player.HP, 0, damage, "victory"}
		} else {
			player.CombatTargetID = npcID
			counter := randDamage(counterMinDamage, counterMaxDamage)
			counter = reduceCounter(counter, bonus*allyCounterReductionPercent+s.blessingTotalLocked(player, blessingCounterReduction)+takeGuard(player))
			player.HP -= counter
			if player.HP <= 0 {
				s.respawnPlayerLocked(player, *name, "attack_counter", npcID, npc.Name.Get(locale))
				event = flavor{key: "attack_struck_down", player: *name, npc: npc}
				result = combatResult{0, enemyHP, damage, "dead"}
			} else {
				event = flavor{key: "attack_hit", player: *name, npc: npc, n: damage, m: counter}
				result = combatResult{player.HP, enemyHP, damage, "combat"}
			}
		}
	}

	logger.Info("combat_attack", "player", *name, "npc", npcID, "status", result.Status, "damage", result.Damage, "attacker_hp", result.AttackerHP, "target_hp", result.TargetHP)
	data, err := json.Marshal(result)
	if err != nil {
		s.mu.Unlock()
		logger.Error("encode_response_failed", "command", "ATTACK", "error", err.Error())
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	client := conn.(*serverClient)
	response, err := client.enqueueResponse("OK " + string(data))
	if err == nil {
		s.broadcastFlavorLocked(encounterRoomID, event)
	}
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}

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
	targetID := player.CombatTargetID
	npc := s.world.NPCs[targetID]
	if targetID == "" || npc == nil {

		player.CombatTargetID = ""
		targetID, npc = s.blockingEnemyLocked(player, player.RoomID)
		if npc == nil {
			s.mu.Unlock()
			fmt.Fprintln(conn, "ERR 407 NOT_IN_COMBAT")
			return false
		}
	}
	encounterRoomID := player.RoomID

	fleeSucceeds := npc.FleeAccurate || (npc.FleeSucceedsOnce && !player.FledFrom[targetID])

	var result string
	var event flavor
	if fleeSucceeds {
		result = "success"
		event = flavor{key: "flee_success", player: *name, npc: npc}
		player.CombatTargetID = ""
		player.guarding = false
		if player.FledFrom == nil {
			player.FledFrom = make(map[string]bool)
		}
		player.FledFrom[targetID] = true
	} else {
		counter := reduceCounter(randDamage(counterMinDamage, counterMaxDamage), s.blessingTotalLocked(player, blessingCounterReduction)+takeGuard(player))
		player.HP -= counter
		if player.HP <= 0 {
			s.respawnPlayerLocked(player, *name, "flee_failed", targetID, npc.Name.Get(locale))
			result = "failure_dead"
			event = flavor{key: "flee_dead", player: *name, npc: npc}
		} else {
			result = "failure"
			event = flavor{key: "flee_hit", player: *name, npc: npc, n: counter}
		}
	}

	logger.Info("combat_flee", "player", *name, "npc", targetID, "result", result, "hp", player.HP)
	data, err := json.Marshal(struct {
		HP     int    `json:"hp"`
		Result string `json:"result"`
	}{player.HP, result})
	if err != nil {
		s.mu.Unlock()
		logger.Error("encode_response_failed", "command", "FLEE", "error", err.Error())
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	client := conn.(*serverClient)
	response, err := client.enqueueResponse("OK " + string(data))
	if err == nil {
		s.broadcastFlavorLocked(encounterRoomID, event)
	}
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}

// handleDefend is a custom, additive command (like FLEE): brace instead of attacking.
// You deal no damage, but the next counter-attack against you is halved. Only valid during a fight.
func handleDefend(s *Server, conn net.Conn, name *string, parts []string) bool {
	if !requireExactArgs(conn, parts, 1) {
		return false
	}
	if *name == "" {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
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
	player.guarding = true
	encounterRoomID := player.RoomID
	logger.Info("combat_defend", "player", *name, "npc", player.CombatTargetID, "hp", player.HP)
	data, err := json.Marshal(struct {
		HP     int    `json:"hp"`
		Result string `json:"result"`
	}{player.HP, "braced"})
	if err != nil {
		s.mu.Unlock()
		logger.Error("encode_response_failed", "command", "DEFEND", "error", err.Error())
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	client := conn.(*serverClient)
	response, err := client.enqueueResponse("OK " + string(data))
	if err == nil {
		s.broadcastFlavorLocked(encounterRoomID, flavor{key: "defend", player: *name, npc: npc})
	}
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}
