package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	logger = newLogger(io.Discard, slog.LevelDebug)
	os.Exit(m.Run())
}

type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) records(t *testing.T) []map[string]any {
	t.Helper()
	b.mu.Lock()
	defer b.mu.Unlock()
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(b.buf.String()), "\n") {
		if line == "" {
			continue
		}
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("log line is not JSON: %q: %v", line, err)
		}
		out = append(out, record)
	}
	return out
}

func captureLogs(t *testing.T) *lockedBuffer {
	t.Helper()
	buffer := &lockedBuffer{}
	original := logger
	logger = newLogger(buffer, slog.LevelDebug)
	t.Cleanup(func() { logger = original })
	return buffer
}

func find(records []map[string]any, msg string, attrs ...any) map[string]any {
	for _, record := range records {
		if record["msg"] != msg {
			continue
		}
		match := true
		for i := 0; i+1 < len(attrs); i += 2 {
			if fmt.Sprint(record[attrs[i].(string)]) != fmt.Sprint(attrs[i+1]) {
				match = false
			}
		}
		if match {
			return record
		}
	}
	return nil
}

func TestLogsConnectionsCommandsResponsesAndErrors(t *testing.T) {
	logs := captureLogs(t)
	server := newServer(t.TempDir())
	server.world = questTestWorld()
	alice := startTestClient(t, server)
	alice.connect(t, "alice")
	alice.cmd(t, "TAKE item.herb", "OK taken=item.herb")
	alice.cmd(t, "FOO", "ERR 400 BAD_REQUEST")
	alice.cmd(t, "TAKE nothing here", "ERR 404 ITEM_NOT_FOUND")
	alice.command(t, "QUIT", "OK bye")
	<-alice.done

	records := logs.records(t)
	for _, r := range records {
		if r["time"] == nil || r["level"] == nil || r["msg"] == nil {
			t.Fatalf("record lacks time/level/msg: %v", r)
		}
		if _, err := time.Parse(time.RFC3339Nano, r["time"].(string)); err != nil {
			t.Fatalf("bad timestamp %v: %v", r["time"], err)
		}
	}
	if find(records, "connection_open") == nil || find(records, "connection_open")["remote"] == nil {
		t.Error("no connection_open with the remote address")
	}
	if find(records, "connection_close", "player", "alice") == nil {
		t.Error("no connection_close for alice")
	}
	if find(records, "command", "player", "alice", "command", "TAKE", "args", "item.herb") == nil {
		t.Error("TAKE was not logged with the player name and parameters")
	}
	if find(records, "command", "player", "", "command", "CONNECT", "args", "alice") == nil {
		t.Error("CONNECT was not logged")
	}
	if r := find(records, "response", "player", "alice", "command", "TAKE"); r == nil || r["level"] != "INFO" {
		t.Errorf("TAKE response not logged at INFO: %v", r)
	}
	if r := find(records, "error_response", "code", "404", "command", "TAKE"); r == nil || r["level"] != "WARN" {
		t.Errorf("404 not logged at WARN: %v", r)
	}
	if r := find(records, "error_response", "code", "400", "command", "FOO"); r == nil || r["level"] != "WARN" {
		t.Errorf("unknown command error not logged at WARN: %v", r)
	}
	if find(records, "item_taken", "player", "alice", "item", "item.herb") == nil {
		t.Error("item movement was not logged")
	}
	if find(records, "event") == nil {
		t.Error("EVT lines should be logged at DEBUG when enabled")
	}
}

func TestLogsWorldChangesAndQuestProgress(t *testing.T) {
	useBestCaseDice(t)
	logs := captureLogs(t)
	server := newServer(t.TempDir())
	server.world = loadRealWorld(t)
	w := newWalker(t, server, "alice")
	w.teleport("loc.argo_salmydessus")
	w.do("QUEST npc.phineus")
	w.do("kill npc.harpy")
	w.do("TALK npc.phineus")
	w.teleport("loc.ody_cyclops")
	w.client.run(t, "ATTACK npc.polyphemus")
	w.teleport(server.world.StartRoomID)
	w.do("MOVE east")

	records := logs.records(t)
	for _, want := range []struct {
		msg   string
		attrs []any
	}{
		{"quest_accepted", []any{"player", "alice", "quest", "quest.phineus_harpies"}},
		{"combat_attack", []any{"player", "alice", "npc", "npc.harpy", "status", "victory"}},
		{"quest_completed", []any{"player", "alice", "quest", "quest.phineus_harpies"}},
		{"npc_interaction", []any{"player", "alice", "npc", "npc.phineus"}},
		{"player_died", []any{"player", "alice", "cause", "attack_unprepared"}},
		{"player_moved", []any{"player", "alice", "to", "loc.ody_troy_shore"}},
	} {
		if find(records, want.msg, want.attrs...) == nil {
			t.Errorf("no %s log with %v", want.msg, want.attrs)
		}
	}
}

func TestAbuseIsDetectedAndLogged(t *testing.T) {
	logs := captureLogs(t)

	var flood floodTracker
	now := time.Now()
	warned := 0
	for i := 0; i < maxCommandsPerWindow+10; i++ {
		if _, warn := flood.record(now.Add(time.Duration(i) * time.Millisecond)); warn {
			warned++
		}
	}
	if warned != 1 {
		t.Errorf("flood warned %d times, want exactly 1 (throttled)", warned)
	}
	if _, warn := flood.record(now.Add(2 * commandWindow)); warn {
		t.Error("a slow command after the burst should not warn")
	}

	monitor := newAbuseMonitor()
	for i := 0; i < maxConnectsPerWindow+3; i++ {
		monitor.noteConnection("203.0.113.9", now.Add(time.Duration(i)*time.Millisecond))
	}
	monitor.noteConnection("198.51.100.1", now)

	records := logs.records(t)
	r := find(records, "abuse_rapid_connections", "ip", "203.0.113.9")
	if r == nil || r["level"] != "WARN" {
		t.Fatalf("rapid reconnects not warned at WARN: %v", r)
	}
	if find(records, "abuse_rapid_connections", "ip", "198.51.100.1") != nil {
		t.Error("a single connection must not be reported")
	}
}

func TestCommandFloodOnALiveConnectionIsLogged(t *testing.T) {
	logs := captureLogs(t)
	server := newServer(t.TempDir())
	client := startTestClient(t, server)
	for i := 0; i < maxCommandsPerWindow+5; i++ {
		client.command(t, "WHO", "OK players=0")
	}
	if find(logs.records(t), "abuse_command_flood") == nil {
		t.Error("command flood on a live connection was not logged")
	}
}

func TestClipAndLogFile(t *testing.T) {
	long := strings.Repeat("あ", maxLoggedText+50)
	if got := []rune(clip(long)); len(got) != maxLoggedText+1 {
		t.Errorf("clip kept %d runes, want %d", len(got), maxLoggedText+1)
	}
	if clip("short") != "short" {
		t.Error("short text must not change")
	}

	path := t.TempDir() + "/tap.log"
	t.Setenv("TAP_LOG_FILE", path)
	t.Setenv("TAP_LOG_LEVEL", "warn")
	original := logger
	closeLog := setupLogging()
	logger.Info("hidden")
	logger.Warn("visible", "k", "v")
	closeLog()
	logger = original
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"msg":"visible"`) || strings.Contains(string(data), "hidden") {
		t.Errorf("log file = %q, want only the WARN record", data)
	}
}
