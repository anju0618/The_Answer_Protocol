// プレイヤーの状態(playerdata.json)の読み書きと、サーバー起動時の
// アイテム所持状態の復元。CONNECT/QUIT/切断のたびにこのファイルが
// 読み書きされ、プレイヤーはHP・所持品・居場所などを保ったまま
// 再接続できる。
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const playersSaveFile = "playerdata.json"

// playersSavePath はプレイヤー状態の保存先ファイルパスを返す。
func (s *Server) playersSavePath() string {
	return filepath.Join(s.saveDir, playersSaveFile)
}

// loadPlayers は保存済みの全プレイヤー状態(名前→Player)を読み込む。
// 保存ファイルが無い/空なら空のマップを返す。
func (s *Server) loadPlayers() (map[string]*Player, error) {
	data, err := os.ReadFile(s.playersSavePath())
	if errors.Is(err, os.ErrNotExist) {
		return make(map[string]*Player), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read player state: %w", err)
	}
	if len(data) == 0 {
		return make(map[string]*Player), nil
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

// restoreItemOwnership はサーバー起動時に一度だけ呼ばれ、保存済み
// プレイヤーの所持品に含まれるアイテムを「まだ部屋には無い」状態
// (RoomID=="")にする(そうしないと、誰かの持ち物であるはずのアイテムが
// world.jsonの初期配置どおり部屋にも存在する、という矛盾が起きる)。
// 同じアイテムが2人の所持品に重複して入っていれば異常として弾く。
func (s *Server) restoreItemOwnership() error {
	players, err := s.loadPlayers()
	if err != nil {
		return err
	}
	owners := make(map[string]string)
	for name, player := range players {
		for _, itemID := range player.Inventory {
			if s.world.Items[itemID] == nil {
				continue
			}
			if owner, exists := owners[itemID]; exists {
				return fmt.Errorf("item %q appears in inventories of %q and %q", itemID, owner, name)
			}
			owners[itemID] = name
		}
	}
	for itemID := range owners {
		s.world.Items[itemID].RoomID = ""
	}
	return nil
}

// savePlayer は player 1人ぶんの状態を、既存の保存データにマージして
// 書き込む。
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

// writePlayers は players 全員ぶんの状態をplayerdata.jsonへ書き込む。
// item_store.goのwriteItemLocationsと同じく、一時ファイル+renameの
// アトミックな保存パターン。
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
