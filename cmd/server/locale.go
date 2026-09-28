package main

import (
	"fmt"
	"net"
	"strings"
)

const defaultLocale = "en"

type LocalizedText map[string]string

func (t LocalizedText) Get(locale string) string {
	if s, ok := t[locale]; ok && s != "" {
		return s
	}
	return t[defaultLocale]
}

var supportedLocales = map[string]bool{"en": true, "ja": true}

func clientLocale(conn net.Conn) string {
	if client, ok := conn.(*serverClient); ok && client.locale != "" {
		return client.locale
	}
	return defaultLocale
}

type roomView struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       map[string]string `json:"exits"`
}

func newRoomView(room *Room, locale string) roomView {
	exits := make(map[string]string, len(room.Exits))
	for dir, dest := range room.Exits {
		exits[dir] = dest
	}
	return roomView{ID: room.ID, Name: room.Name.Get(locale), Description: room.Description.Get(locale), Exits: exits}
}

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
