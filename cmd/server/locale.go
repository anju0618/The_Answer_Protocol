// 多言語対応(LANGコマンド)の仕組み。RFC規定のJSONレスポンスの形は
// 言語によらず一切変えず、ワイヤーに乗せる直前にサーバー内部の
// LocalizedText を該当言語のプレーン文字列へ解決する、という方針
// (memo.md 8章、README「Protocol Implementation」参照)。
package main

import (
	"fmt"
	"net"
	"strings"
)

const defaultLocale = "en"

// LocalizedText は同じテキストの多言語版を言語コード("en"/"ja")で
// 引けるようにしたもの。data/world.json の name/description/dialogue が
// この形で格納されている。
type LocalizedText map[string]string

// Get は locale のテキストを返す。未翻訳(または locale が空)なら
// 英語("en")にフォールバックする。
func (t LocalizedText) Get(locale string) string {
	if s, ok := t[locale]; ok && s != "" {
		return s
	}
	return t[defaultLocale]
}

var supportedLocales = map[string]bool{"en": true, "ja": true}

// clientLocale は conn(実体は *serverClient)に設定された言語を返す。
// LANGコマンドを一度も送っていない接続では常に defaultLocale("en")。
func clientLocale(conn net.Conn) string {
	if client, ok := conn.(*serverClient); ok && client.locale != "" {
		return client.locale
	}
	return defaultLocale
}

// roomView はLOOKが実際にワイヤーへ乗せる形。Name/Descriptionはプレーン
// 文字列で、Room構造体(LocalizedTextを持つ)を直接JSONにすると言語コード
// →テキストのマップになってしまうため、別の型として用意している。
type roomView struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       map[string]string `json:"exits"`
}

// newRoomView は room を locale で解決した roomView に変換する。
func newRoomView(room *Room, locale string) roomView {
	exits := make(map[string]string, len(room.Exits))
	for dir, dest := range room.Exits {
		exits[dir] = dest
	}
	return roomView{ID: room.ID, Name: room.Name.Get(locale), Description: room.Description.Get(locale), Exits: exits}
}

// handleLang はLANGコマンドを処理する。CONNECT前(*name==="")にしか
// 送れない独自コマンドで、以降そのコネクションのストーリーテキストが
// 指定言語で返るようになる。RFC定義のコマンドではないため、対応言語コード
// (en/ja)以外や、CONNECT後の呼び出しはERR 400 BAD_REQUESTを返す。
func handleLang(s *Server, conn net.Conn, name *string, parts []string) bool {
	if !requireExactArgs(conn, parts, 2) {
		return false
	}
	if *name != "" {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	code := strings.ToLower(parts[1])
	if !supportedLocales[code] {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	if client, ok := conn.(*serverClient); ok {
		client.locale = code
	}
	fmt.Fprintln(conn, "OK lang="+code)
	return false
}
