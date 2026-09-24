// The Answer Protocol サーバーの中核。Server構造体、接続ライフサイクル
// (接続・切断・セーブ/ロード)、コマンドディスパッチテーブル、および
// LOOK/MOVE/CONNECT/QUIT/TAKE/DROP/INVENTORY/TALK/STATUSの各ハンドラを
// まとめている(ATTACK/FLEEはcombat.go、QUEST/QUESTSはquest.go、
// CHAT/GROUPはchat.go/group.goに分離)。
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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

// Server はサーバー全体の状態。mu が接続中プレイヤー・ワールド・
// グループなどインメモリの共有状態全体を保護する単一ロックで、ioMu は
// それとは別に、ディスクへの保存処理(player_store.go / item_store.go)を
// 直列化する(ディスクI/Oの遅延でmuを長時間ロックしないため)。
type Server struct {
	mu            sync.Mutex
	ioMu          sync.Mutex
	players       map[string]*Player
	clients       map[string]*serverClient
	groups        map[string]*Group
	groupByPlayer map[string]string
	unsavedTakes  map[string]map[string]string
	nextGroupID   uint64
	saveDir       string
	world         *World
}

// NewServer は本番用のServerを作る。data/world.json からワールドを
// ロードし、保存済みのアイテム位置・所持状態を復元する。失敗時は
// log.Fatalfでプロセスごと終了する(起動時にしか呼ばれないため)。
func NewServer() *Server {
	s := newServer("saves")
	world, err := loadWorld("data/world.json")
	if err != nil {
		log.Fatalf("load world: %v", err)
	}
	s.world = world
	if err := s.restoreItemLocations(); err != nil {
		log.Fatalf("restore item locations: %v", err)
	}
	if err := s.restoreItemOwnership(); err != nil {
		log.Fatalf("restore item ownership: %v", err)
	}
	return s
}

// newServer は空の(ワールド未設定の)Serverを作る。テストでは実際の
// world.jsonを読まず、直接 server.world にテスト用のWorldを差し込む
// ために本体からNewServerとは別に切り出されている。
func newServer(saveDir string) *Server {
	return &Server{
		players:       make(map[string]*Player),
		clients:       make(map[string]*serverClient),
		groups:        make(map[string]*Group),
		groupByPlayer: make(map[string]string),
		unsavedTakes:  make(map[string]map[string]string),
		saveDir:       saveDir,
	}
}

// connectPlayer は name でのCONNECTを処理する。既にその名前で接続中なら
// errNameInUseを返す。保存済みのプレイヤーがいればその状態を復元し、
// いなければ新規プレイヤー(ワールドの開始地点、HP100)を作って保存する。
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

// saveAndRemovePlayer は name の現在の状態をディスクへ保存してから、
// インメモリの接続中プレイヤー一覧から除去する(QUIT・正常切断の共通処理)。
// 保存に失敗した場合は除去せず、次回の接続時に再試行できるようにする。
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

// removePlayerLocked は name をインメモリの状態(players/clients/groups等)
// から除去する。restoreUnsavedTakes が true の場合(=保存できずに切断した
// 異常系)、TAKEしたがまだディスクに保存されていないアイテムを元の部屋に
// 戻す。同じ部屋にいる他プレイヤーへEVT ROOM PRESENCE LEAVEを通知する。
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

// broadcastPlayerCountLocked は現在の接続人数を全クライアントへ
// EVT STATSとして送る(接続・切断のたびに呼ばれる)。
func (s *Server) broadcastPlayerCountLocked() {
	event := fmt.Sprintf("EVT STATS players=%d", len(s.players))
	for _, client := range s.clients {
		client.enqueueEvent(event)
	}
}

// playerForUpdateLocked は、状態を書き換えてよいプレイヤーを返す。
// 切断処理中(exiting==true)のプレイヤーはnilを返す(切断のちょうど
// 最中に別のコマンドが割り込んで状態を書き換えてしまうのを防ぐため)。
// 読み取り専用の用途(LOOKなど)ではこれを使わず s.players を直接見てよい。
func (s *Server) playerForUpdateLocked(name string) *Player {
	player := s.players[name]
	if player == nil || player.exiting {
		return nil
	}
	player.regenLocked(time.Now())
	return player
}

// requireArgs は引数の数が min 以上あるかを検証し、無ければERR 400を
// 返してfalseを返す(足りていればtrue)。各ハンドラの先頭で使う共通処理。
func requireArgs(conn net.Conn, parts []string, min int) bool {
	if len(parts) < min {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	return true
}

// requireExactArgs はrequireArgsの「ちょうどn個」版。
func requireExactArgs(conn net.Conn, parts []string, n int) bool {
	if len(parts) != n {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	return true
}

// commandHandler は各コマンドの処理関数の型。戻り値は「この接続を
// 切断すべきか」(true=切断)。
type commandHandler func(s *Server, conn net.Conn, name *string, parts []string) (stop bool)

// commandHandlers はコマンド名→処理関数のディスパッチテーブル。
// LANG/FLEEはRFCに無い独自拡張、それ以外はRFC 5章で定義された15コマンド。
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
	"QUEST":     handleQuest,
	"QUESTS":    handleQuests,
}

// handleConnect はCONNECT <name>コマンドを処理する。1接続につき1回だけ
// 呼べる(*nameが既に設定されていればERR 400)。名前の重複はERR 201。
// 成功すると同じ部屋にいる他プレイヤーへEVT ROOM PRESENCE ENTERを通知する。
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
		log.Printf("connect player %q: %v", requestedName, err)
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	*name = requestedName
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
			// 初回接続時だけ運命の間のチュートリアルを自動で流す(以降は
			// モイライへのTALKでいつでも聞き直せる)。
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

// handleLook はLOOKコマンドを処理する。現在の部屋の情報(名前・説明・
// 出口)と、同じ部屋にいるプレイヤー・アイテムID・NPC IDの一覧を返す。
// 部屋名・説明はclientLocaleで解決した言語のプレーン文字列になる
// (locale.goのroomView参照)。
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
		log.Printf("encode LOOK response: %v", err)
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

// handleMove はMOVE <direction>コマンドを処理する。処理順は:
//  1. 指定方向の出口が存在するか(無ければERR 301)
//  2. 現在の部屋に「まだ倒しても振り切ってもいない生きた敵」がいないか
//     (いれば部屋封鎖=即死。hazard.goのblockingEnemyLockedを参照)
//  3. 実際に移動し、他プレイヤーへ入退室イベントを通知
//  4. オデュッセイア編ならクルー初期化、移動先の部屋ハザードを判定
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
	if _, blocker := s.blockingEnemyLocked(player, player.RoomID); blocker != nil {
		encounterRoomID := player.RoomID
		locale := clientLocale(conn)
		client := conn.(*serverClient)
		response, err := client.enqueueResponse("OK room=" + destination)
		if err == nil {
			s.respawnPlayerLocked(player, *name, "slip_past", blocker.Name.Get(locale))
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

// handleWho はWHOコマンドを処理する。現在の全接続人数を返す(認証前でも呼べる)。
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
			log.Printf("save player %q on quit: %v", *name, err)
			fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
			return false
		}
		log.Println(*name, "disconnected")
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
		// 一意のアイテムは部屋から消える。Renewableな鍵アイテムは部屋に残り、
		// このプレイヤー専用のコピーが所持品に入るだけ。
		if s.unsavedTakes[*name] == nil {
			s.unsavedTakes[*name] = make(map[string]string)
		}
		s.unsavedTakes[*name][itemID] = player.RoomID
		s.world.Items[itemID].RoomID = ""
	}
	player.Inventory = append(player.Inventory, itemID)
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
		// 鍵アイテムのコピーは捨てるだけ(部屋には元から残っている)ので、
		// アイテム位置の保存は不要で、プレイヤーの所持品だけ保存する。
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
						log.Printf("restore item location %q after failed drop: %v", itemID, rollbackErr)
					}
				}
			}
		}
	}
	s.ioMu.Unlock()
	if err != nil {
		s.mu.Unlock()
		log.Printf("drop item %q for %q: %v", itemID, *name, err)
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}

	player.Inventory = snapshot.Inventory
	if !item.Renewable {
		item.RoomID = player.RoomID
	}
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
			s.respawnPlayerLocked(player, *name, "talk_unprepared", npc.Name.Get(locale), s.mythNeedText(npc, locale))
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
		if npc.Guide {
			// モイライのTALK: 1行目はレスポンス、残りはEVT PLAYER GUIDEで再生する。
			s.sendGuideLocked(*name, 1)
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
		log.Printf("encode STATUS response: %v", err)
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

	defer func() {
		if name != "" {
			if err := s.saveAndRemovePlayer(name); err != nil {
				log.Printf("save player %q on disconnect: %v", name, err)
				s.mu.Lock()
				s.removePlayerLocked(name, true)
				s.mu.Unlock()
			}
			log.Println(name, "disconnected")
		}
		conn.Close()
	}()

	if _, err := fmt.Fprintln(conn, "OK hello proto=1"); err != nil {
		return
	}

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		parts := parseCommandParts(scanner.Text())
		if len(parts) == 0 {
			continue
		}

		command := strings.ToUpper(parts[0])
		handler, ok := commandHandlers[command]
		if !ok {
			fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
			continue
		}
		if handler(s, conn, &name, parts) {
			return
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}
