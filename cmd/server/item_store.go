// アイテムの現在位置(TAKE/DROPでworld.jsonの初期配置から動いた分)を
// ディスク(itemdata.json)へ永続化する処理。サーバー再起動後もアイテムの
// 場所が復元されるようにするためのもの。
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const itemsSaveFile = "itemdata.json"

// itemsSavePath はアイテム位置の保存先ファイルパスを返す。
func (s *Server) itemsSavePath() string {
	return filepath.Join(s.saveDir, itemsSaveFile)
}

// loadItemLocations は保存済みのアイテムID→部屋IDのマップを読み込む。
// 保存ファイルが無ければ(初回起動時)空のマップを返す。
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

// restoreItemLocations は保存済みのアイテム位置をロードし、
// s.world.Items の RoomID を上書きする。サーバー起動時に一度だけ
// 呼ばれる(NewServer)。保存データが指すアイテム/部屋がworld.jsonに
// 実在しなければエラーを返す。
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
			// 鍵アイテムは常に元の部屋に置かれている(過去の保存データで
			// 別の場所に動いていても無視する)。
			continue
		}
		s.world.Items[itemID].RoomID = roomID
	}
	return nil
}

// writeItemLocations は locations をitemdata.jsonへ書き込む。一時ファイルに
// 書いてから rename する(=書き込み途中でクラッシュしても既存の保存
// ファイルが壊れない)アトミックな保存パターン。
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
