package main

import (
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	maxLoggedText = 300

	connectWindow        = 10 * time.Second
	maxConnectsPerWindow = 8

	commandWindow        = time.Second
	maxCommandsPerWindow = 20

	abuseWarnInterval = 5 * time.Second
)

var logger = newLogger(os.Stderr, slog.LevelInfo)

func newLogger(w io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level}))
}

func setupLogging() (closeFn func()) {
	level := slog.LevelInfo
	switch strings.ToLower(os.Getenv("TAP_LOG_LEVEL")) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	var out io.Writer = os.Stderr
	closeFn = func() {}
	if path := os.Getenv("TAP_LOG_FILE"); path != "" {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			newLogger(os.Stderr, slog.LevelInfo).Error("open_log_file", "path", path, "error", err.Error())
		} else {
			out = io.MultiWriter(os.Stderr, file)
			closeFn = func() { file.Close() }
		}
	}
	logger = newLogger(out, level)
	slog.SetDefault(logger)
	return closeFn
}

func fatal(msg string, err error) {
	logger.Error(msg, "error", err.Error())
	os.Exit(1)
}

func clip(s string) string {
	runes := []rune(s)
	if len(runes) <= maxLoggedText {
		return s
	}
	return string(runes[:maxLoggedText]) + "…"
}

func hostOf(remote string) string {
	if host, _, err := net.SplitHostPort(remote); err == nil {
		return host
	}
	return remote
}

func logOutbound(client *serverClient, data []byte) {
	player, command := client.context()
	for _, line := range strings.Split(strings.TrimRight(string(data), "\r\n"), "\n") {
		switch {
		case line == "OK" || strings.HasPrefix(line, "OK "):
			logger.Info("response", "remote", client.remote, "player", player, "command", command, "line", clip(line))
		case strings.HasPrefix(line, "ERR "):
			code, _, _ := strings.Cut(strings.TrimPrefix(line, "ERR "), " ")
			logger.Warn("error_response", "remote", client.remote, "player", player, "command", command, "code", code, "line", clip(line))
		case strings.HasPrefix(line, "EVT "):
			logger.Debug("event", "remote", client.remote, "player", player, "line", clip(line))
		}
	}
}

type abuseMonitor struct {
	mu       sync.Mutex
	connects map[string][]time.Time
	lastWarn map[string]time.Time
}

func newAbuseMonitor() *abuseMonitor {
	return &abuseMonitor{connects: make(map[string][]time.Time), lastWarn: make(map[string]time.Time)}
}

func (m *abuseMonitor) noteConnection(ip string, now time.Time) {
	m.mu.Lock()
	for knownIP, times := range m.connects {
		recent := times[:0]
		for _, t := range times {
			if now.Sub(t) < connectWindow {
				recent = append(recent, t)
			}
		}
		if len(recent) == 0 {
			delete(m.connects, knownIP)
			delete(m.lastWarn, knownIP)
		} else {
			m.connects[knownIP] = recent
		}
	}
	recent := append(m.connects[ip], now)
	m.connects[ip] = recent
	count := len(recent)
	warn := count > maxConnectsPerWindow && now.Sub(m.lastWarn[ip]) >= abuseWarnInterval
	if warn {
		m.lastWarn[ip] = now
	}
	m.mu.Unlock()
	if warn {
		logger.Warn("abuse_rapid_connections", "ip", ip, "connections", count, "window_ms", connectWindow.Milliseconds())
	}
}

type floodTracker struct {
	times    []time.Time
	lastWarn time.Time
}

func (f *floodTracker) record(now time.Time) (count int, warn bool) {
	recent := f.times[:0]
	for _, t := range f.times {
		if now.Sub(t) < commandWindow {
			recent = append(recent, t)
		}
	}
	f.times = append(recent, now)
	count = len(f.times)
	if count > maxCommandsPerWindow && now.Sub(f.lastWarn) >= abuseWarnInterval {
		f.lastWarn = now
		return count, true
	}
	return count, false
}
