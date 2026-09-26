package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

var errNameInUse = errors.New("player name in use")

const defaultStartRoomID = "loc.square"

type Server struct {
	mu      sync.Mutex
	ioMu    sync.Mutex
	players map[string]*Player
	saveDir string
	world   *World
}

func NewServer() *Server {
	s := newServer("saves")
	world, err := loadWorld("data/world.json")
	if err != nil {
		log.Fatalf("load world: %v", err)
	}
	s.world = world
	return s
}

func newServer(saveDir string) *Server {
	return &Server{
		players: make(map[string]*Player),
		saveDir: saveDir,
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
	player, err := s.loadPlayer(name)
	s.ioMu.Unlock()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.players[name]; exists {
		return errNameInUse
	}
	if player == nil {
		startRoom := defaultStartRoomID
		if s.world != nil {
			startRoom = s.world.StartRoomID
		}
		player = &Player{Name: name, HP: 100, RoomID: startRoom}
	}
	s.players[name] = player
	return nil
}

func (s *Server) saveAndRemovePlayer(name string) error {
	s.mu.Lock()
	player := s.players[name]
	s.mu.Unlock()

	s.ioMu.Lock()
	err := s.savePlayer(player)
	s.ioMu.Unlock()
	if err != nil {
		return err
	}

	s.mu.Lock()
	delete(s.players, name)
	s.mu.Unlock()
	return nil
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
	"STATUS":    handleStatus,
	"QUEST":     handleQuest,
	"QUESTS":    handleQuests,
}

func handleConnect(s *Server, conn net.Conn, name *string, parts []string) bool {
	if len(parts) != 2 || *name != "" {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}

	requestedName := parts[1]
	if !utf8.ValidString(requestedName) ||
		strings.IndexFunc(requestedName, unicode.IsControl) >= 0 {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	if err := s.connectPlayer(requestedName); errors.Is(err, errNameInUse) {
		fmt.Fprintln(conn, "ERR 201 NAME_IN_USE")
		return false
	} else if err != nil {
		log.Printf("load player %q: %v", requestedName, err)
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	*name = requestedName
	fmt.Fprintln(conn, "OK connected")
	return false
}

func handleLook(s *Server, conn net.Conn, name *string, parts []string) bool {
	requireExactArgs(conn, parts, 1)
	return false
}

func handleMove(s *Server, conn net.Conn, name *string, parts []string) bool {
	requireExactArgs(conn, parts, 2)
	return false
}

func handleWho(s *Server, conn net.Conn, name *string, parts []string) bool {
	if !requireExactArgs(conn, parts, 1) {
		return false
	}
	s.mu.Lock()
	count := len(s.players)
	s.mu.Unlock()
	fmt.Fprintf(conn, "OK players=%d\n", count)
	return false
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

func handleChat(s *Server, conn net.Conn, name *string, parts []string) bool {
	requireArgs(conn, parts, 3)
	return false
}

func handleGroup(s *Server, conn net.Conn, name *string, parts []string) bool {
	if !requireArgs(conn, parts, 2) {
		return false
	}
	switch strings.ToUpper(parts[1]) {
	case "CREATE":
		requireExactArgs(conn, parts, 2)
	case "INVITE":
		requireExactArgs(conn, parts, 3)
	case "JOIN":
		requireExactArgs(conn, parts, 3)
	case "LEAVE":
		requireExactArgs(conn, parts, 2)
	default:
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
	}
	return false
}

func handleTake(s *Server, conn net.Conn, name *string, parts []string) bool {
	requireArgs(conn, parts, 2)
	return false
}

func handleDrop(s *Server, conn net.Conn, name *string, parts []string) bool {
	requireArgs(conn, parts, 2)
	return false
}

func handleInventory(s *Server, conn net.Conn, name *string, parts []string) bool {
	requireExactArgs(conn, parts, 1)
	return false
}

func handleTalk(s *Server, conn net.Conn, name *string, parts []string) bool {
	requireArgs(conn, parts, 2)
	return false
}

func handleAttack(s *Server, conn net.Conn, name *string, parts []string) bool {
	requireArgs(conn, parts, 2)
	return false
}

func handleStatus(s *Server, conn net.Conn, name *string, parts []string) bool {
	requireExactArgs(conn, parts, 1)
	return false
}

func handleQuest(s *Server, conn net.Conn, name *string, parts []string) bool {
	requireArgs(conn, parts, 2)
	return false
}

func handleQuests(s *Server, conn net.Conn, name *string, parts []string) bool {
	requireExactArgs(conn, parts, 1)
	return false
}

func (s *Server) handleClient(conn net.Conn) {
	var name string

	defer func() {
		if name != "" {
			if err := s.saveAndRemovePlayer(name); err != nil {
				log.Printf("save player %q on disconnect: %v", name, err)
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
		parts := strings.Fields(scanner.Text())
		if len(parts) == 0 {
			continue
		}

		handler, ok := commandHandlers[strings.ToUpper(parts[0])]
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
