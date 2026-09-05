package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"strings"
)

func randDamage(min, max int) int {
	return min + rand.IntN(max-min+1)
}

// respawnPlayerLocked kills the player and sends them back to the world's
// start room with a reduced HP, per TASKS.md's "HP0で安全地帯にリスポーン"
// requirement. Callers must hold s.mu.
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

	s.mu.Lock()
	player := s.playerForUpdateLocked(*name)
	if player == nil || s.world == nil || s.world.Rooms[player.RoomID] == nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	npcID := s.world.resolveNPCInRoom(player.RoomID, query)
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
		// This fight can never be won outright: attacking it costs crew
		// instead of HP (see memo.md 7.3, Laestrygonians).
		lost := spendCrewLocked(player, npc.CrewLossOnAttack)
		player.CombatTargetID = ""
		resultText = fmt.Sprintf("%s attacks %s and is driven back, losing %d crew.", *name, npc.Name, lost)
		result = combatResult{player.HP, npc.HP, 0, "overwhelmed"}

	case npc.MythRequirementItem != "" && !player.hasItem(npc.MythRequirementItem):
		s.respawnPlayerLocked(player, *name)
		resultText = fmt.Sprintf("%s attacks %s unprepared and is killed.", *name, npc.Name)
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
			resultText = fmt.Sprintf("%s defeats %s.", *name, npc.Name)
			result = combatResult{player.HP, 0, damage, "victory"}
		} else {
			player.CombatTargetID = npcID
			counter := randDamage(counterMinDamage, counterMaxDamage)
			player.HP -= counter
			if player.HP <= 0 {
				s.respawnPlayerLocked(player, *name)
				resultText = fmt.Sprintf("%s is struck down by %s.", *name, npc.Name)
				result = combatResult{0, npc.HP, damage, "dead"}
			} else {
				resultText = fmt.Sprintf("%s attacks %s for %d damage and takes %d in return.", *name, npc.Name, damage, counter)
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

func handleFlee(s *Server, conn net.Conn, name *string, parts []string) bool {
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
	encounterRoomID := player.RoomID

	var resultText, result string
	if npc.FleeAccurate {
		result = "success"
		resultText = fmt.Sprintf("%s flees from %s.", *name, npc.Name)
		player.CombatTargetID = ""
	} else {
		counter := randDamage(counterMinDamage, counterMaxDamage)
		player.HP -= counter
		if player.HP <= 0 {
			s.respawnPlayerLocked(player, *name)
			result = "failure_dead"
			resultText = fmt.Sprintf("%s tries to flee %s and is cut down.", *name, npc.Name)
		} else {
			result = "failure"
			resultText = fmt.Sprintf("%s tries to flee %s and is struck for %d.", *name, npc.Name, counter)
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
