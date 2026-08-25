package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const playersSaveFile = "playerdata.json"

func (s *Server) playersSavePath() string {
	return filepath.Join(s.saveDir, playersSaveFile)
}

func (s *Server) loadPlayer(name string) (*Player, error) {
	players, err := s.loadPlayers()
	if err != nil {
		return nil, err
	}
	return players[name], nil
}

func (s *Server) loadPlayers() (map[string]*Player, error) {
	data, err := os.ReadFile(s.playersSavePath())
	if errors.Is(err, os.ErrNotExist) {
		return make(map[string]*Player), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read player state: %w", err)
	}

	var players map[string]*Player
	if err := json.Unmarshal(data, &players); err != nil {
		return nil, fmt.Errorf("decode player state: %w", err)
	}
	if players == nil {
		return nil, errors.New("invalid player state: expected object")
	}
	for name, player := range players {
		if player == nil || player.Name != name {
			return nil, fmt.Errorf("player name mismatch for %q", name)
		}
	}
	return players, nil
}

func (s *Server) savePlayer(player *Player) error {
	if player == nil {
		return errors.New("missing player state")
	}
	players, err := s.loadPlayers()
	if err != nil {
		return err
	}
	players[player.Name] = player
	return s.writePlayers(players)
}

func (s *Server) writePlayers(players map[string]*Player) error {
	data, err := json.MarshalIndent(players, "", "  ")
	if err != nil {
		return fmt.Errorf("encode player state: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(s.saveDir, 0700); err != nil {
		return fmt.Errorf("create save directory: %w", err)
	}

	temp, err := os.CreateTemp(s.saveDir, ".players-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary save: %w", err)
	}
	defer os.Remove(temp.Name())
	defer temp.Close()
	if _, err := temp.Write(data); err != nil {
		return fmt.Errorf("write player state: %w", err)
	}
	if err := temp.Sync(); err != nil {
		return fmt.Errorf("sync player state: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close player state: %w", err)
	}
	if err := os.Rename(temp.Name(), s.playersSavePath()); err != nil {
		return fmt.Errorf("replace player state: %w", err)
	}
	return nil
}
