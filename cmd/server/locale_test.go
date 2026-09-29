package main

import "testing"

func localeTestWorld() *World {
	return &World{
		StartRoomID: "loc.start",
		Rooms: map[string]*Room{
			"loc.start": {
				ID:          "loc.start",
				Name:        LocalizedText{"en": "Starting Room", "ja": "始まりの部屋"},
				Description: LocalizedText{"en": "A quiet room.", "ja": "静かな部屋だ。"},
			},
			"loc.untranslated": {
				ID:          "loc.untranslated",
				Name:        LocalizedText{"en": "English Only"},
				Description: LocalizedText{"en": "No Japanese text yet."},
			},
		},
		NPCs: map[string]*NPC{
			"npc.guide": {
				Name:     LocalizedText{"en": "Guide", "ja": "案内人"},
				RoomID:   "loc.start",
				Dialogue: []LocalizedText{{"en": "Welcome, traveler.", "ja": "ようこそ、旅人よ。"}},
			},
		},
	}
}

func TestLangSelectsStoryLanguage(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = localeTestWorld()

	alice := startTestClient(t, server)
	alice.connect(t, "alice")
	look := alice.cmdJSON(t, "LOOK")
	if look["room"].(map[string]any)["name"] != "Starting Room" {
		t.Fatalf("default-locale room name = %v, want English", look["room"])
	}
	alice.cmd(t, "TALK npc.guide", "OK Welcome, traveler.")

	bob := startTestClient(t, server)
	bob.cmd(t, "LANG ja", "OK lang=ja")
	bob.connect(t, "bob", alice)
	look = bob.cmdJSON(t, "LOOK")
	room := look["room"].(map[string]any)
	if room["name"] != "始まりの部屋" || room["description"] != "静かな部屋だ。" {
		t.Fatalf("ja-locale room = %v, want Japanese text", room)
	}
	bob.cmd(t, "TALK npc.guide", "OK ようこそ、旅人よ。")

	bob.cmd(t, "LANG en", "ERR 400 BAD_REQUEST")

	carol := startTestClient(t, server)
	carol.cmd(t, "LANG fr", "ERR 400 BAD_REQUEST")

	server.mu.Lock()
	server.players["bob"].RoomID = "loc.untranslated"
	server.mu.Unlock()
	look = bob.cmdJSON(t, "LOOK")
	room = look["room"].(map[string]any)
	if room["name"] != "English Only" {
		t.Fatalf("untranslated room name = %v, want English fallback", room["name"])
	}
}
