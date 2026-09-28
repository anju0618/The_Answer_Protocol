package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func loadRealWorld(t *testing.T) *World {
	t.Helper()
	world, err := loadWorld(filepath.Join("..", "..", "data", "world.json"))
	if err != nil {
		t.Fatalf("load real world: %v", err)
	}
	return world
}

func TestRealWorldGuideAndQuestHintsExist(t *testing.T) {
	world := loadRealWorld(t)
	server := newServer(t.TempDir())
	server.world = world

	guide := server.guideNPC()
	if guide == nil {
		t.Fatal("no guide NPC in the start room")
	}
	for _, locale := range []string{"en", "ja"} {
		for i, line := range guide.Dialogue {
			text := line[locale]
			if strings.TrimSpace(text) == "" {
				t.Errorf("guide line %d has no %s text", i, locale)
			}
			if len("EVT PLAYER GUIDE ")+len(text) > maxProtocolLineBytes {
				t.Errorf("guide line %d (%s) is too long for one protocol line", i, locale)
			}
		}
	}

	for id, quest := range world.Quests {
		// クエスト文は状況の説明だけ。やり方・場所・答えを教えるヒントは書かない。
		for _, locale := range []string{"en", "ja"} {
			text := quest.Description[locale]
			if strings.TrimSpace(text) == "" {
				t.Errorf("quest %s has no %s description", id, locale)
			}
			for _, spoiler := range []string{"Hint", "ヒント", "TAKE", "ATTACK", "FLEE", "QUEST ", "TALK"} {
				if strings.Contains(text, spoiler) {
					t.Errorf("quest %s (%s) gives away the answer with %q: %s", id, locale, spoiler, text)
				}
			}
		}
		if quest.Name["ja"] == "" {
			t.Errorf("quest %s has no Japanese name", id)
		}
	}
}

func TestRealWorldJapaneseIntroAndDeathMessage(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = loadRealWorld(t)
	alice := startTestClient(t, server)
	alice.command(t, "LANG ja", "OK lang=ja")
	alice.connect(t, "alice")

	first := alice.waitEvent(t, "EVT PLAYER GUIDE ")
	if !strings.Contains(first, "運命の間") {
		t.Fatalf("first guide line = %q, want a Japanese welcome mentioning 運命の間", first)
	}

	server.mu.Lock()
	server.players["alice"].RoomID = "loc.ody_cyclops"
	server.mu.Unlock()
	if _, err := fmt.Fprintln(alice.conn, "ATTACK ポリュペモス"); err != nil {
		t.Fatal(err)
	}
	death := alice.waitEvent(t, "EVT PLAYER DEATH ")
	for _, want := range []string{"ポリュペモス", "運命の間"} {
		if !strings.Contains(death, want) {
			t.Errorf("death message %q does not mention %q", death, want)
		}
	}
	if strings.Contains(death, "杭") {
		t.Errorf("death message gives away the answer: %q", death)
	}
}

func TestRealWorldQuestGiverAnnouncedOnEntry(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = loadRealWorld(t)
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	if _, err := fmt.Fprintln(alice.conn, "MOVE west"); err != nil {
		t.Fatal(err)
	}
	line := alice.waitEvent(t, "EVT PLAYER QUEST ")
	if !strings.Contains(line, "Tiphys") || !strings.Contains(line, "QUEST Tiphys") {
		t.Fatalf("announcement = %q, want Tiphys and the QUEST command", line)
	}
	data := alice.cmdJSON(t, "QUEST Tiphys")
	if data["quest_id"] != "quest.amycus_bout" {
		t.Fatalf("QUEST Tiphys = %v", data)
	}
}
