package main

import (
	"bufio"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestLongPlayerCombatEventsRemainReadable(t *testing.T) {
	for _, tc := range []struct {
		name       string
		playerName string
		locale     string
		short      bool
	}{
		{"ordinary name", "alice", "en", true},
		{"long ASCII name, English", strings.Repeat("n", 65511), "en", false},
		{"long ASCII name, Japanese", strings.Repeat("n", 65511), "ja", false},
		{"long UTF-8 name, Japanese", strings.Repeat("名", 21837), "ja", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := newServer(t.TempDir())
			server.world = takeTestWorld()
			server.world.NPCs = map[string]*NPC{
				"npc.guard": {
					Name: LocalizedText{"en": "Guard", "ja": "🦉番人"},
					Role: "enemy", RoomID: "loc.start", HP: 1,
				},
			}
			attacker := startTestClient(t, server)
			attacker.connect(t, tc.playerName)
			observer := startTestClient(t, server)
			if tc.locale == "ja" {
				observer.command(t, "LANG ja", "OK lang=ja")
			}
			observer.connect(t, "bob", attacker)
			result := attacker.cmdJSON(t, "ATTACK npc.guard")
			if result["status"] != "victory" {
				t.Fatalf("attack status = %v, want victory", result["status"])
			}
			for _, recipient := range []*testClient{attacker, observer} {
				line, err := recipient.reader.ReadString('\n')
				if err != nil {
					t.Fatal(err)
				}
				event := strings.TrimSuffix(line, "\n")
				if !strings.HasPrefix(event, "EVT ROOM COMBAT ") || !utf8.ValidString(event) {
					t.Fatal("combat event lost its prefix or contains invalid UTF-8")
				}
				if len(event) > maxProtocolLineBytes {
					t.Fatalf("combat event is %d bytes, limit is %d", len(event), maxProtocolLineBytes)
				}
				scanner := bufio.NewScanner(strings.NewReader(line))
				if !scanner.Scan() || scanner.Text() != event || scanner.Scan() || scanner.Err() != nil {
					t.Fatalf("default CLI scanner cannot read the combat event: %v", scanner.Err())
				}
				if tc.short && event != "EVT ROOM COMBAT alice defeats Guard." {
					t.Fatalf("ordinary event changed: %q", event)
				}
				if !tc.short && !strings.HasSuffix(event, "...") {
					t.Fatal("truncated combat event has no indication of omitted text")
				}
				recipient.command(t, "WHO", "OK players=2")
			}
		})
	}
}
