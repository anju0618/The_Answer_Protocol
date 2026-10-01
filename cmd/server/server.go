package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

var errNameInUse = errors.New("player name in use")

const defaultStartRoomID = "loc.hall_of_fates"
const maxProtocolLineBytes = bufio.MaxScanTokenSize - 1

const (
	maxPlayerHP      = 100
	respawnHP        = 20
	combatMinDamage  = 8
	combatMaxDamage  = 14
	counterMinDamage = 7
	counterMaxDamage = 14

	odysseyStartRoomID = "loc.ody_troy_shore"
	startingCrew       = 12
)

type Server struct {
	mu            sync.Mutex
	ioMu          sync.Mutex
	players       map[string]*Player
	clients       map[string]*serverClient
	groups        map[string]*Group
	groupByPlayer map[string]string
	unsavedTakes  map[string]map[string]string
	nextGroupID   uint64
	abuse         *abuseMonitor
	saveDir       string
	world         *World
}

func NewServer() *Server {
	s := newServer("saves")
	world, err := loadWorld("data/world.json")
	if err != nil {
		fatal("load_world_failed", err)
	}
	s.world = world
	if err := s.restoreItemLocations(); err != nil {
		fatal("restore_item_locations_failed", err)
	}
	if err := s.restoreItemOwnership(); err != nil {
		fatal("restore_item_ownership_failed", err)
	}
	return s
}

func newServer(saveDir string) *Server {
	return &Server{
		players:       make(map[string]*Player),
		clients:       make(map[string]*serverClient),
		groups:        make(map[string]*Group),
		groupByPlayer: make(map[string]string),
		unsavedTakes:  make(map[string]map[string]string),
		abuse:         newAbuseMonitor(),
		saveDir:       saveDir,
	}
}

func (s *Server) connectPlayer(name string) error {
	s.mu.Lock()
	if _, exists := s.players[name]; exists {
		s.mu.Unlock()
		return errNameInUse
	}
	s.mu.Unlock()

	s.ioMu.Lock()
	players, err := s.loadPlayers()
	var player *Player
	if err == nil {
		player = players[name]
		switch {
		case player == nil:
			startRoom := defaultStartRoomID
			if s.world != nil {
				startRoom = s.world.StartRoomID
			}
			player = &Player{Name: name, HP: 100, RoomID: startRoom}
			players[name] = player
			err = s.writePlayers(players)
		case s.world != nil && s.world.Rooms[player.RoomID] == nil:
			err = fmt.Errorf("saved player %q has unknown room %q", name, player.RoomID)
		}
	}
	s.ioMu.Unlock()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.players[name]; exists {
		return errNameInUse
	}
	s.players[name] = player
	return nil
}

func (s *Server) saveAndRemovePlayer(name string) error {
	s.mu.Lock()
	player := s.players[name]
	if player == nil {
		s.mu.Unlock()
		return fmt.Errorf("player %q is not connected", name)
	}
	if player.exiting {
		s.mu.Unlock()
		return fmt.Errorf("player %q is already exiting", name)
	}
	player.exiting = true
	snapshot := *player
	snapshot.Inventory = append([]string(nil), player.Inventory...)
	s.mu.Unlock()

	s.ioMu.Lock()
	err := s.savePlayer(&snapshot)
	s.ioMu.Unlock()
	s.mu.Lock()
	if err != nil {
		player.exiting = false
		s.mu.Unlock()
		return err
	}

	s.removePlayerLocked(name, false)
	s.mu.Unlock()
	return nil
}

func (s *Server) removePlayerLocked(name string, restoreUnsavedTakes bool) {
	player := s.players[name]
	if player == nil {
		return
	}
	if restoreUnsavedTakes && s.world != nil {
		inventory := make(map[string]struct{}, len(player.Inventory))
		for _, itemID := range player.Inventory {
			inventory[itemID] = struct{}{}
		}
		for itemID, roomID := range s.unsavedTakes[name] {
			if _, held := inventory[itemID]; !held || s.world.Rooms[roomID] == nil {
				continue
			}
			if item := s.world.Items[itemID]; item != nil && item.RoomID == "" {
				item.RoomID = roomID
			}
		}
	}
	roomID := player.RoomID
	delete(s.players, name)
	delete(s.clients, name)
	delete(s.unsavedTakes, name)
	s.removeGroupMemberLocked(name)
	s.clearGroupInvitesLocked(name)
	for otherName, other := range s.players {
		if other.RoomID == roomID {
			if client := s.clients[otherName]; client != nil {
				client.enqueueEvent("EVT ROOM PRESENCE LEAVE " + name)
			}
		}
	}
	s.broadcastPlayerCountLocked()
}

func (s *Server) broadcastPlayerCountLocked() {
	event := fmt.Sprintf("EVT STATS players=%d", len(s.players))
	for _, client := range s.clients {
		client.enqueueEvent(event)
	}
}

func (s *Server) playerForUpdateLocked(name string) *Player {
	player := s.players[name]
	if player == nil || player.exiting {
		return nil
	}
	player.regenLocked(time.Now())
	return player
}

func requireArgs(conn net.Conn, parts []string, min int) bool {
	if len(parts) < min {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	return true
}

func requireExactArgs(conn net.Conn, parts []string, n int) bool {
	if len(parts) != n {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	return true
}

type commandHandler func(s *Server, conn net.Conn, name *string, parts []string) (stop bool)

var commandHandlers = map[string]commandHandler{
	"LANG":      handleLang,
	"CONNECT":   handleConnect,
	"LOOK":      handleLook,
	"MOVE":      handleMove,
	"WHO":       handleWho,
	"QUIT":      handleQuit,
	"CHAT":      handleChat,
	"GROUP":     handleGroup,
	"TAKE":      handleTake,
	"DROP":      handleDrop,
	"INVENTORY": handleInventory,
	"TALK":      handleTalk,
	"ATTACK":    handleAttack,
	"FLEE":      handleFlee,
	"STATUS":    handleStatus,
	"STATE":     handleState,
	"QUEST":     handleQuest,
	"QUESTS":    handleQuests,
}

func handleConnect(s *Server, conn net.Conn, name *string, parts []string) bool {
	client := conn.(*serverClient)
	if len(parts) != 2 || *name != "" {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}

	requestedName := parts[1]
	if len("EVT ROOM PRESENCE ENTER ")+len(requestedName) > maxProtocolLineBytes ||
		!utf8.ValidString(requestedName) ||
		strings.IndexFunc(requestedName, unicode.IsControl) >= 0 {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	if err := s.connectPlayer(requestedName); errors.Is(err, errNameInUse) {
		fmt.Fprintln(conn, "ERR 201 NAME_IN_USE")
		return false
	} else if err != nil {
		logger.Error("connect_player_failed", "player", requestedName, "error", err.Error())
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	*name = requestedName
	client.setContext(requestedName, "CONNECT")
	s.mu.Lock()
	response, err := client.enqueueResponse("OK connected")
	if err == nil {
		s.clients[requestedName] = client
		roomID := s.players[requestedName].RoomID
		for otherName, other := range s.players {
			if otherName != requestedName && other.RoomID == roomID {
				if recipient := s.clients[otherName]; recipient != nil {
					recipient.enqueueEvent("EVT ROOM PRESENCE ENTER " + requestedName)
				}
			}
		}
		s.broadcastPlayerCountLocked()
		if player := s.players[requestedName]; player != nil {

			if !player.IntroSeen {
				player.IntroSeen = true
				s.sendGuideLocked(requestedName, 0)
			}
			s.announceQuestGiversLocked(player)
		}
	}
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}

func handleLook(s *Server, conn net.Conn, name *string, parts []string) bool {
	if !requireExactArgs(conn, parts, 1) {
		return false
	}
	if *name == "" {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}

	s.mu.Lock()
	player := s.players[*name]
	if player == nil || s.world == nil || s.world.Rooms[player.RoomID] == nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	roomID := player.RoomID
	room := newRoomView(s.world.Rooms[roomID], clientLocale(conn))

	players := make([]string, 0)
	for playerName, other := range s.players {
		if other.RoomID == roomID {
			players = append(players, playerName)
		}
	}
	items := make([]string, 0)
	for itemID, item := range s.world.Items {
		if item != nil && item.visibleTo(player, itemID, roomID) {
			items = append(items, itemID)
		}
	}
	npcs := make([]string, 0)
	for npcID, npc := range s.world.NPCs {
		if npc != nil && npc.RoomID == roomID {
			npcs = append(npcs, npcID)
		}
	}
	sort.Strings(players)
	sort.Strings(items)
	sort.Strings(npcs)
	data, err := json.Marshal(struct {
		Room    roomView `json:"room"`
		Players []string `json:"players"`
		Items   []string `json:"items"`
		NPCs    []string `json:"npcs"`
	}{room, players, items, npcs})
	if err != nil {
		s.mu.Unlock()
		logger.Error("encode_response_failed", "command", "LOOK", "error", err.Error())
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	if len("OK ")+len(data) > maxProtocolLineBytes {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	client := conn.(*serverClient)
	response, err := client.enqueueResponse("OK " + string(data))
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}

func handleMove(s *Server, conn net.Conn, name *string, parts []string) bool {
	if !requireExactArgs(conn, parts, 2) {
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
	room := s.world.Rooms[player.RoomID]
	if room == nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	destination, ok := room.Exits[strings.ToLower(parts[1])]
	if !ok {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 301 NO_EXIT")
		return false
	}
	if s.world.Rooms[destination] == nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	if blockerID, blocker := s.blockingEnemyLocked(player, player.RoomID); blocker != nil {
		encounterRoomID := player.RoomID
		locale := clientLocale(conn)
		client := conn.(*serverClient)
		response, err := client.enqueueResponse("OK room=" + destination)
		if err == nil {
			s.respawnPlayerLocked(player, *name, "slip_past", blockerID, blocker.Name.Get(locale))
			s.broadcastFlavorLocked(encounterRoomID, flavor{key: "slip_past", player: *name, npc: blocker})
		}
		s.mu.Unlock()
		if err != nil {
			return true
		}
		return client.waitResponse(response) != nil
	}
	client := conn.(*serverClient)
	response, err := client.enqueueResponse("OK room=" + destination)
	if err == nil {
		oldRoomID := player.RoomID
		player.RoomID = destination
		logger.Info("player_moved", "player", *name, "from", oldRoomID, "to", destination)
		if oldRoomID != destination {
			for playerName, current := range s.players {
				recipient := s.clients[playerName]
				if recipient == nil {
					continue
				}
				switch current.RoomID {
				case oldRoomID:
					recipient.enqueueEvent("EVT ROOM PRESENCE LEAVE " + *name)
				case destination:
					recipient.enqueueEvent("EVT ROOM PRESENCE ENTER " + *name)
				}
			}
		}
		s.initializeCrewLocked(player, oldRoomID)
		if event := s.applyRoomHazardLocked(player, *name, s.world.Rooms[destination], clientLocale(conn)); event != nil {
			s.broadcastFlavorLocked(destination, *event)
		}
		s.announceQuestGiversLocked(player)
	}
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}

func handleWho(s *Server, conn net.Conn, name *string, parts []string) bool {
	if !requireExactArgs(conn, parts, 1) {
		return false
	}
	client := conn.(*serverClient)
	s.mu.Lock()
	response, err := client.enqueueResponse(fmt.Sprintf("OK players=%d", len(s.players)))
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}

func handleQuit(s *Server, conn net.Conn, name *string, parts []string) bool {
	if len(parts) != 1 {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	if *name != "" {
		if err := s.saveAndRemovePlayer(*name); err != nil {
			logger.Error("save_player_failed", "player", *name, "when", "quit", "error", err.Error())
			fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
			return false
		}
		logger.Info("player_quit", "player", *name)
		*name = ""
	}
	fmt.Fprintln(conn, "OK bye")
	return true
}

func handleTake(s *Server, conn net.Conn, name *string, parts []string) bool {
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

	locale := clientLocale(conn)
	itemID := ""
	if item := s.world.Items[query]; item != nil && item.availableTo(player, query) {
		itemID = query
	} else {
		for id, item := range s.world.Items {
			if item != nil && item.availableTo(player, id) && strings.EqualFold(item.Name.Get(locale), query) && (itemID == "" || id < itemID) {
				itemID = id
			}
		}
	}
	if itemID == "" {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 404 ITEM_NOT_FOUND")
		return false
	}

	if !s.world.Items[itemID].Renewable {

		if s.unsavedTakes[*name] == nil {
			s.unsavedTakes[*name] = make(map[string]string)
		}
		s.unsavedTakes[*name][itemID] = player.RoomID
		s.world.Items[itemID].RoomID = ""
	}
	player.Inventory = append(player.Inventory, itemID)
	logger.Info("item_taken", "player", *name, "item", itemID, "room", player.RoomID, "renewable", s.world.Items[itemID].Renewable)
	s.checkQuestObjectiveLocked(player, "collect_item", itemID)
	takenInRoomID := player.RoomID

	client := conn.(*serverClient)
	response, err := client.enqueueResponse("OK taken=" + itemID)
	if err == nil {
		if event := s.applyTakeConsequencesLocked(player, *name, itemID); event != nil {
			s.broadcastFlavorLocked(takenInRoomID, *event)
		}
	}
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}

func handleDrop(s *Server, conn net.Conn, name *string, parts []string) bool {
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
	index := -1
	for i, itemID := range player.Inventory {
		if itemID == query {
			index = i
			break
		}
	}
	if index == -1 {
		locale := clientLocale(conn)
		for i, itemID := range player.Inventory {
			if item := s.world.Items[itemID]; item != nil && strings.EqualFold(item.Name.Get(locale), query) && (index == -1 || itemID < player.Inventory[index]) {
				index = i
			}
		}
	}
	if index == -1 {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 404 ITEM_NOT_IN_INVENTORY")
		return false
	}
	itemID := player.Inventory[index]
	item := s.world.Items[itemID]
	if item == nil || (!item.Renewable && item.RoomID != "") {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	if len("OK dropped=")+len(itemID) > maxProtocolLineBytes {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}

	snapshot := *player
	snapshot.Inventory = append(make([]string, 0, len(player.Inventory)-1), player.Inventory[:index]...)
	snapshot.Inventory = append(snapshot.Inventory, player.Inventory[index+1:]...)
	s.ioMu.Lock()
	var err error
	if item.Renewable {

		err = s.savePlayer(&snapshot)
	} else {
		var locations map[string]string
		locations, err = s.loadItemLocations()
		if err == nil {
			previousRoom, hadPreviousRoom := locations[itemID]
			locations[itemID] = player.RoomID
			err = s.writeItemLocations(locations)
			if err == nil {
				err = s.savePlayer(&snapshot)
				if err != nil {
					if hadPreviousRoom {
						locations[itemID] = previousRoom
					} else {
						delete(locations, itemID)
					}
					if rollbackErr := s.writeItemLocations(locations); rollbackErr != nil {
						logger.Error("restore_item_location_failed", "item", itemID, "error", rollbackErr.Error())
					}
				}
			}
		}
	}
	s.ioMu.Unlock()
	if err != nil {
		s.mu.Unlock()
		logger.Error("drop_item_failed", "player", *name, "item", itemID, "error", err.Error())
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}

	player.Inventory = snapshot.Inventory
	if !item.Renewable {
		item.RoomID = player.RoomID
	}
	logger.Info("item_dropped", "player", *name, "item", itemID, "room", player.RoomID)
	delete(s.unsavedTakes, *name)
	client := conn.(*serverClient)
	response, err := client.enqueueResponse("OK dropped=" + itemID)
	if err == nil {
		if event := s.applyDropConsequencesLocked(player, *name, itemID); event != nil {
			s.broadcastFlavorLocked(player.RoomID, *event)
		}
	}
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}

func handleInventory(s *Server, conn net.Conn, name *string, parts []string) bool {
	if !requireExactArgs(conn, parts, 1) {
		return false
	}
	if *name == "" {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	s.mu.Lock()
	player := s.players[*name]
	if player == nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	items := append(make([]string, 0, len(player.Inventory)), player.Inventory...)
	sort.Strings(items)
	data, err := json.Marshal(items)
	if err != nil || len("OK ")+len(data) > maxProtocolLineBytes {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	client := conn.(*serverClient)
	response, err := client.enqueueResponse("OK " + string(data))
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}

func handleTalk(s *Server, conn net.Conn, name *string, parts []string) bool {
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
	locale := clientLocale(conn)
	npcID := s.world.resolveNPCInRoom(player.RoomID, query, locale)
	if npcID == "" {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 404 NPC_NOT_FOUND")
		return false
	}
	npc := s.world.NPCs[npcID]

	if npc.Role != "enemy" && npc.hasMythRequirement() && !player.meetsMythRequirement(npc) {
		encounterRoomID := player.RoomID
		client := conn.(*serverClient)
		response, err := client.enqueueResponse("OK dead")
		if err == nil {
			s.respawnPlayerLocked(player, *name, "talk_unprepared", npcID, npc.Name.Get(locale))
			s.broadcastFlavorLocked(encounterRoomID, flavor{key: "talk_unprepared", player: *name, npc: npc})
		}
		s.mu.Unlock()
		if err != nil {
			return true
		}
		return client.waitResponse(response) != nil
	}

	dialogue := ""
	if len(npc.Dialogue) > 0 {
		dialogue = npc.Dialogue[0].Get(locale)
	}
	if dialogue == "" || strings.TrimSpace(dialogue) == "" || !utf8.ValidString(dialogue) ||
		strings.IndexFunc(dialogue, unicode.IsControl) >= 0 || len("OK ")+len(dialogue) > maxProtocolLineBytes {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	client := conn.(*serverClient)
	response, err := client.enqueueResponse("OK " + dialogue)
	if err == nil {
		logger.Info("npc_interaction", "player", *name, "npc", npcID, "room", player.RoomID)
		if npc.Guide {

			s.sendGuideLocked(*name, 1)
			s.sendDeathHintLocked(player, *name)
		}
		s.sendQuestHintLocked(player, npcID)
		s.talkEndingLocked(player, npc)
	}
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}

func handleStatus(s *Server, conn net.Conn, name *string, parts []string) bool {
	if !requireExactArgs(conn, parts, 1) {
		return false
	}
	if *name == "" {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}

	s.mu.Lock()
	player := s.players[*name]
	if player == nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	player.regenLocked(time.Now())
	status := "healthy"
	if player.CombatTargetID != "" {
		status = "combat"
	}
	data, err := json.Marshal(struct {
		HP     int    `json:"hp"`
		MaxHP  int    `json:"max_hp"`
		Status string `json:"status"`
	}{player.HP, maxPlayerHP, status})
	if err != nil {
		s.mu.Unlock()
		logger.Error("encode_response_failed", "command", "STATUS", "error", err.Error())
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	client := conn.(*serverClient)
	response, err := client.enqueueResponse("OK " + string(data))
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}

func (s *Server) handleClient(rawConn net.Conn) {
	conn := newServerClient(rawConn)
	var name string
	lastPlayer := ""
	openedAt := time.Now()

	logger.Info("connection_open", "remote", conn.remote)
	s.abuse.noteConnection(hostOf(conn.remote), openedAt)

	defer func() {
		if name != "" {
			if err := s.saveAndRemovePlayer(name); err != nil {
				logger.Error("save_player_failed", "player", name, "when", "disconnect", "error", err.Error())
				s.mu.Lock()
				s.removePlayerLocked(name, true)
				s.mu.Unlock()
			}
		}
		logger.Info("connection_close", "remote", conn.remote, "player", lastPlayer, "duration_ms", time.Since(openedAt).Milliseconds())
		conn.Close()
	}()

	if _, err := fmt.Fprintln(conn, "OK hello proto=1"); err != nil {
		return
	}

	scanner := bufio.NewScanner(conn)
	var flood floodTracker

	for scanner.Scan() {
		parts := parseCommandParts(scanner.Text())
		if len(parts) == 0 {
			continue
		}

		command := strings.ToUpper(parts[0])
		conn.setContext(name, command)
		logger.Info("command", "remote", conn.remote, "player", name, "command", command, "args", clip(strings.Join(parts[1:], " ")))
		if count, warn := flood.record(time.Now()); warn {
			logger.Warn("abuse_command_flood", "remote", conn.remote, "player", name, "commands", count, "window_ms", commandWindow.Milliseconds())
		}

		handler, ok := commandHandlers[command]
		if !ok {
			fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
			continue
		}
		stop := handler(s, conn, &name, parts)
		conn.setContext(name, command)
		if name != "" {
			lastPlayer = name
		}
		if stop {
			return
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Warn("connection_read_error", "remote", conn.remote, "player", name, "error", err.Error())
	}
}
