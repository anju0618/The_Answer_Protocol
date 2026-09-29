package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const itemsSaveFile = "itemdata.json"

func (s *Server) itemsSavePath() string {
	return filepath.Join(s.saveDir, itemsSaveFile)
}

func (s *Server) loadItemLocations() (map[string]string, error) {
	data, err := os.ReadFile(s.itemsSavePath())
	if errors.Is(err, os.ErrNotExist) {
		return make(map[string]string), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read item locations: %w", err)
	}
	var locations map[string]string
	if err := json.Unmarshal(data, &locations); err != nil {
		return nil, fmt.Errorf("decode item locations: %w", err)
	}
	if locations == nil {
		return nil, errors.New("invalid item locations: expected object")
	}
	return locations, nil
}

func (s *Server) restoreItemLocations() error {
	locations, err := s.loadItemLocations()
	if err != nil {
		return err
	}
	for itemID, roomID := range locations {
		if s.world.Items[itemID] == nil {
			return fmt.Errorf("saved item %q does not exist", itemID)
		}
		if s.world.Rooms[roomID] == nil {
			return fmt.Errorf("saved item %q has unknown room %q", itemID, roomID)
		}
	}
	for itemID, roomID := range locations {
		if s.world.Items[itemID].Renewable {

			continue
		}
		s.world.Items[itemID].RoomID = roomID
	}
	return nil
}

func (s *Server) writeItemLocations(locations map[string]string) error {
	data, err := json.MarshalIndent(locations, "", "  ")
	if err != nil {
		return fmt.Errorf("encode item locations: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(s.saveDir, 0700); err != nil {
		return fmt.Errorf("create save directory: %w", err)
	}
	temp, err := os.CreateTemp(s.saveDir, ".items-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary item save: %w", err)
	}
	defer os.Remove(temp.Name())
	defer temp.Close()
	if _, err := temp.Write(data); err != nil {
		return fmt.Errorf("write item locations: %w", err)
	}
	if err := temp.Sync(); err != nil {
		return fmt.Errorf("sync item locations: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close item locations: %w", err)
	}
	if err := os.Rename(temp.Name(), s.itemsSavePath()); err != nil {
		return fmt.Errorf("replace item locations: %w", err)
	}
	return nil
}
