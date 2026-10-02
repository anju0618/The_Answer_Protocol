package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestRealWorldHintsAreWellFormed(t *testing.T) {
	world := loadRealWorld(t)
	if len(world.Hints) == 0 {
		t.Fatal("the world defines no hints")
	}
	for subject, hint := range world.Hints {
		for _, locale := range []string{"en", "ja"} {
			text := hint[locale]
			if strings.TrimSpace(text) == "" {
				t.Errorf("hint %s has no %s text", subject, locale)
			}
			if len("EVT PLAYER HINT ")+len(text) > maxProtocolLineBytes {
				t.Errorf("hint %s (%s) is too long for one protocol line", subject, locale)
			}
		}
	}
	for id, room := range world.Rooms {
		if strings.HasPrefix(id, "loc.ody_") && room.Hazard != nil && room.Hazard.Type != "crew_cost" {
			if _, ok := world.Hints[id]; !ok {
				t.Errorf("room %s is deadly but has no hint", id)
			}
		}
	}
}

func TestUnknownHintSubjectIsRejected(t *testing.T) {
	world := loadRealWorld(t)
	world.Hints["npc.does_not_exist"] = LocalizedText{"en": "x"}
	if err := world.validate(); err == nil {
		t.Fatal("a hint about an unknown subject must fail validation")
	}
}

func TestMoiraiGivesOneHintAboutTheLastDeath(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = loadRealWorld(t)
	alice := startTestClient(t, server)
	alice.connect(t, "alice")
	alice.waitEvent(t, "EVT PLAYER GUIDE ")

	server.mu.Lock()
	server.players["alice"].RoomID = "loc.ody_cyclops"
	server.mu.Unlock()
	if _, err := fmt.Fprintln(alice.conn, "ATTACK Polyphemus"); err != nil {
		t.Fatal(err)
	}
	death := alice.waitEvent(t, "EVT PLAYER DEATH ")
	if strings.Contains(death, "olive") {
		t.Fatalf("the death message itself must not hint: %q", death)
	}
	alice.waitEvent(t, "OK ")

	if _, err := fmt.Fprintln(alice.conn, "TALK npc.moirai"); err != nil {
		t.Fatal(err)
	}
	hint := alice.waitEvent(t, "EVT PLAYER HINT ")
	if !strings.Contains(hint, "olive") {
		t.Fatalf("hint = %q, want it to point at the olive stake", hint)
	}

	server.mu.Lock()
	remaining := server.players["alice"].LastDeathSubject
	server.mu.Unlock()
	if remaining != "" {
		t.Fatalf("LastDeathSubject = %q after the hint, want it cleared", remaining)
	}
}
