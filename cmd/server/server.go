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

type Server struct {
	mu      sync.Mutex
	online  map[string]bool
	players map[string]*Player
	saveDir string
}

func NewServer() *Server {
	return newServer("saves")
}

func newServer(saveDir string) *Server {
	return &Server{
		online:  make(map[string]bool),
		players: make(map[string]*Player),
		saveDir: saveDir,
	}
}

func (s *Server) connectPlayer(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.online[name] {
		return errNameInUse
	}
	player, err := s.loadPlayer(name)
	if err != nil {
		return err
	}
	if player == nil {
		player = &Player{Name: name, HP: 100, RoomID: "loc.square"}
	}
	s.players[name] = player
	s.online[name] = true
	return nil
}

func (s *Server) saveAndRemovePlayer(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.savePlayer(s.players[name]); err != nil {
		return err
	}
	delete(s.players, name)
	delete(s.online, name)
	return nil
}

func (s *Server) removePlayer(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.players, name)
	delete(s.online, name)
}

func (s *Server) handleClient(conn net.Conn) {
	var name string

	defer func() {
		if name != "" {
			if err := s.saveAndRemovePlayer(name); err != nil {
				log.Printf("save player %q on disconnect: %v", name, err)
				s.removePlayer(name)
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

		command := strings.ToUpper(parts[0])

		switch command {
		case "CONNECT":
			if len(parts) != 2 || name != "" {
				fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
				continue
			}

			requestedName := parts[1]
			if !utf8.ValidString(requestedName) ||
				strings.IndexFunc(requestedName, unicode.IsControl) >= 0 {
				fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
				continue
			}
			if err := s.connectPlayer(requestedName); errors.Is(err, errNameInUse) {
				fmt.Fprintln(conn, "ERR 201 NAME_IN_USE")
				continue
			} else if err != nil {
				log.Printf("load player %q: %v", requestedName, err)
				fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
				continue
			}
			name = requestedName
			fmt.Fprintln(conn, "OK connected")

		case "LOOK":
			if len(parts) != 1 {
				fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
				continue
			}

		case "MOVE":
			if len(parts) != 2 {
				fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
				continue
			}

		case "WHO":
			if len(parts) != 1 {
				fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
				continue
			}
			s.mu.Lock()
			count := len(s.online)
			s.mu.Unlock()

			fmt.Fprintf(conn, "OK players=%d\n", count)

		case "QUIT":
			if len(parts) != 1 {
				fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
				continue
			}
			if name != "" {
				if err := s.saveAndRemovePlayer(name); err != nil {
					log.Printf("save player %q on quit: %v", name, err)
					fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
					continue
				}
				log.Println(name, "disconnected")
				name = ""
			}
			fmt.Fprintln(conn, "OK bye")
			return

		case "CHAT":
			if len(parts) < 3 {
				fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
				continue
			}

		case "GROUP":
			if len(parts) < 2 {
				fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
				continue
			}
			switch strings.ToUpper(parts[1]) {
			case "CREATE":
				if len(parts) != 2 {
					fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
					continue
				}
			case "INVITE":
				if len(parts) != 3 {
					fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
					continue
				}
			case "JOIN":
				if len(parts) != 3 {
					fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
					continue
				}
			case "LEAVE":
				if len(parts) != 2 {
					fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
					continue
				}
			default:
				fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
				continue
			}

		case "TAKE":
			if len(parts) < 2 {
				fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
				continue
			}

		case "DROP":
			if len(parts) < 2 {
				fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
				continue
			}

		case "INVENTORY":
			if len(parts) != 1 {
				fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
				continue
			}

		case "TALK":
			if len(parts) < 2 {
				fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
				continue
			}

		case "ATTACK":
			if len(parts) < 2 {
				fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
				continue
			}

		case "STATUS":
			if len(parts) != 1 {
				fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
				continue
			}

		case "QUEST":
			if len(parts) < 2 {
				fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
				continue
			}

		case "QUESTS":
			if len(parts) != 1 {
				fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
				continue
			}

		default:
			fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}
