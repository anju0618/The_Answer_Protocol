package main

import (
	"fmt"
	"strings"
	"testing"
)

// readLinesUntilResponse は response(OK/ERRで始まる行)が届くまでの全行を返す。
// 返り値の最後の要素が応答行で、それより前はその間に届いたEVT行。
func (client *testClient) readLinesUntilResponse(t *testing.T) []string {
	t.Helper()
	var lines []string
	for {
		line, err := client.reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read line: %v (got so far %q)", err, lines)
		}
		line = strings.TrimSuffix(line, "\n")
		lines = append(lines, line)
		if line == "OK" || strings.HasPrefix(line, "OK ") || strings.HasPrefix(line, "ERR ") {
			return lines
		}
	}
}

// waitEvent は prefix で始まるEVT行が届くまで読み進めて、その行を返す。
func (client *testClient) waitEvent(t *testing.T, prefix string) string {
	t.Helper()
	for i := 0; i < 50; i++ {
		line, err := client.reader.ReadString('\n')
		if err != nil {
			t.Fatalf("waiting for %q: %v", prefix, err)
		}
		line = strings.TrimSuffix(line, "\n")
		if strings.HasPrefix(line, prefix) {
			return line
		}
	}
	t.Fatalf("event %q never arrived", prefix)
	return ""
}

func guideTestWorld() *World {
	return &World{
		StartRoomID: "loc.start",
		Rooms:       map[string]*Room{"loc.start": {ID: "loc.start", Name: en("Hall")}},
		Items:       map[string]*Item{"item.stake": {Name: en("Stake"), RoomID: "loc.start", Obtainable: true}},
		NPCs: map[string]*NPC{
			"npc.guide": {Name: en("Guide"), Role: "dialogue", RoomID: "loc.start", HP: 10, Guide: true,
				Dialogue: []LocalizedText{en("line one"), en("line two"), en("line three")}},
			"npc.beast": {Name: en("Beast"), Role: "enemy", RoomID: "loc.start", HP: 100, MythRequirementItem: "item.stake"},
		},
	}
}

// 死んだ本人にだけ、死因・必要だったもの・復活先が届く。
func TestDeathMessageReachesTheDyingPlayer(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = guideTestWorld()
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	if _, err := fmt.Fprintln(alice.conn, "ATTACK npc.beast"); err != nil {
		t.Fatal(err)
	}
	death := alice.waitEvent(t, "EVT PLAYER DEATH ")
	for _, want := range []string{"Beast", "Hall", "20 HP"} {
		if !strings.Contains(death, want) {
			t.Errorf("death message %q does not mention %q", death, want)
		}
	}
	// 何が足りなかったか(答え)は教えない。
	if strings.Contains(death, "Stake") {
		t.Errorf("death message gives away the answer: %q", death)
	}
}

// 死亡文言は全キー・全言語で、渡す引数の数と書式が食い違っていない。
func TestDeathTextsFormatCleanly(t *testing.T) {
	args := map[string][]any{
		"attack_counter":    {"Foe"},
		"attack_unprepared": {"Foe"},
		"flee_failed":       {"Foe"},
		"slip_past":         {"Foe"},
		"talk_unprepared":   {"Foe"},
		"hazard_lethal":     {"Place"},
		"hazard_item":       {"Place"},
		"hazard_crew":       {"Place"},
		"lotus":             {},
		"cattle":            {},
	}
	if len(args) != len(deathTexts) {
		t.Fatalf("test covers %d causes, deathTexts has %d", len(args), len(deathTexts))
	}
	for cause, text := range deathTexts {
		for _, locale := range []string{"en", "ja"} {
			if text[locale] == "" {
				t.Errorf("%s has no %s text", cause, locale)
			}
			got := text.Format(locale, args[cause]...)
			if strings.Contains(got, "%!") {
				t.Errorf("%s/%s formats badly: %q", cause, locale, got)
			}
		}
	}
}

// 初回接続でだけガイドが自動再生され、TALKでは1行目が応答・残りがEVTで届く。
func TestGuidePlaysOnFirstConnectAndReplaysOnTalk(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = guideTestWorld()
	first := startTestClient(t, server)
	first.connect(t, "alice")
	for _, want := range []string{"line one", "line two", "line three"} {
		first.expect(t, "EVT PLAYER GUIDE "+want)
	}

	if _, err := fmt.Fprintln(first.conn, "TALK Guide"); err != nil {
		t.Fatal(err)
	}
	first.expect(t, "OK line one")
	first.expect(t, "EVT PLAYER GUIDE line two")
	first.expect(t, "EVT PLAYER GUIDE line three")

	first.command(t, "QUIT", "OK bye")
	<-first.done

	second := startTestClient(t, server)
	second.connect(t, "alice")
	if _, err := fmt.Fprintln(second.conn, "WHO"); err != nil {
		t.Fatal(err)
	}
	for _, line := range second.readLinesUntilResponse(t) {
		if strings.HasPrefix(line, "EVT PLAYER GUIDE") {
			t.Fatalf("intro replayed on reconnect: %q", line)
		}
	}
}

func questTestWorld() *World {
	return &World{
		StartRoomID: "loc.start",
		Rooms:       map[string]*Room{"loc.start": {ID: "loc.start", Name: en("Start")}},
		Items:       map[string]*Item{"item.herb": {Name: en("Herb"), RoomID: "loc.start", Obtainable: true}},
		NPCs: map[string]*NPC{
			"npc.healer": {Name: en("Healer"), Role: "quest_giver", RoomID: "loc.start", HP: 10, Dialogue: []LocalizedText{en("Hello.")}},
		},
		Quests: map[string]*Quest{
			"quest.herb": {
				Name: en("Fetch Herb"), Description: en("Bring a herb."), GiverNPCID: "npc.healer",
				Objective: QuestObjective{Type: "collect_item", TargetID: "item.herb", Count: 1},
				Reward:    QuestReward{HP: 10},
			},
		},
	}
}

// クエストを持つNPCがいる部屋に着くと通知が来て、達成すると通知が来る。
// 受注前に拾ったアイテムも、受注した時点で数えられる。
func TestQuestAnnouncedAndCompletedEvenIfItemTakenFirst(t *testing.T) {
	server := newServer(t.TempDir())
	server.world = questTestWorld()
	alice := startTestClient(t, server)
	alice.connect(t, "alice")

	line := alice.waitEvent(t, "EVT PLAYER QUEST ")
	if !strings.Contains(line, "Healer") || !strings.Contains(line, "QUEST Healer") {
		t.Fatalf("announcement = %q, want it to name the giver and the command", line)
	}

	alice.cmd(t, "TAKE item.herb", "OK taken=item.herb")
	if _, err := fmt.Fprintln(alice.conn, "QUEST Healer"); err != nil {
		t.Fatal(err)
	}
	done := alice.waitEvent(t, "EVT PLAYER QUEST Quest complete")
	if !strings.Contains(done, "Fetch Herb") {
		t.Fatalf("completion = %q, want quest name", done)
	}
	alice.cmd(t, "STATUS", `OK {"hp":100,"max_hp":100,"status":"healthy"}`)
}
