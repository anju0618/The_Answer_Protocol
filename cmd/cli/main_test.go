package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

type pipeClientConn struct {
	net.Conn
}

func (pipeClientConn) CloseWrite() error {
	return nil
}

func TestQuitErrorAllowsAnotherCommand(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()
	if err := serverConn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatal(err)
	}

	stdin, input := io.Pipe()
	defer stdin.Close()
	defer input.Close()

	clientDone := make(chan struct{})
	go func() {
		runClient(pipeClientConn{clientConn}, stdin)
		close(clientDone)
	}()

	reader := bufio.NewReader(serverConn)

	command := func(want string) {
		t.Helper()
		got, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(got) != want {
			t.Fatalf("command = %q, want %q", strings.TrimSpace(got), want)
		}
	}

	fmt.Fprintln(serverConn, "OK hello proto=1")
	fmt.Fprintln(input, "QUIT")
	command("QUIT")
	fmt.Fprintln(serverConn, "ERR 500 STATE_ERROR")
	fmt.Fprintln(input, "WHO")
	command("WHO")
	fmt.Fprintln(serverConn, "OK players=1")
	fmt.Fprintln(input, "QUIT")
	command("QUIT")
	input.Close()
	fmt.Fprintln(serverConn, "OK bye")

	select {
	case <-clientDone:
	case <-time.After(3 * time.Second):
		t.Fatal("client did not exit after OK bye")
	}
}

func TestQuitWaitsForItsOwnResponse(t *testing.T) {
	for _, tt := range []struct {
		name     string
		command  string
		response string
	}{
		{"previous error", "BAD", "ERR 400 BAD_REQUEST"},
		{"dialogue bye", "TALK Baker", "OK bye"},
		{"dialogue greeting", "TALK Baker", "OK hello proto=1"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			testQuitWaitsForItsOwnResponse(t, tt.command, tt.response)
		})
	}
}

func testQuitWaitsForItsOwnResponse(t *testing.T, firstCommand, firstResponse string) {
	t.Helper()
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()
	if err := serverConn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	clientDone := make(chan struct{})
	go func() {
		runClient(pipeClientConn{clientConn}, strings.NewReader(firstCommand+"\nQUIT\nWHO\n"))
		close(clientDone)
	}()
	reader := bufio.NewReader(serverConn)
	if _, err := fmt.Fprintln(serverConn, "OK hello proto=1"); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{firstCommand, "QUIT"} {
		line, err := reader.ReadString('\n')
		if err != nil || strings.TrimSpace(line) != want {
			t.Fatalf("command = %q, err = %v, want %q", line, err, want)
		}
	}
	if _, err := fmt.Fprintln(serverConn, firstResponse); err != nil {
		t.Fatal(err)
	}
	if err := serverConn.SetReadDeadline(time.Now().Add(200 * time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	line, err := reader.ReadString('\n')
	if err == nil {
		t.Fatalf("sent %q before QUIT response", strings.TrimSpace(line))
	}
	if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
		t.Fatalf("read before QUIT response: %v", err)
	}
	select {
	case <-clientDone:
		t.Fatal("client exited before receiving the QUIT response")
	default:
	}
	if err := serverConn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := fmt.Fprintln(serverConn, "ERR 500 STATE_ERROR"); err != nil {
		t.Fatal(err)
	}
	line, err = reader.ReadString('\n')
	if err != nil || strings.TrimSpace(line) != "WHO" {
		t.Fatalf("command after QUIT rejection = %q, err = %v", line, err)
	}
	if _, err := fmt.Fprintln(serverConn, "OK players=1"); err != nil {
		t.Fatal(err)
	}
	serverConn.Close()
	select {
	case <-clientDone:
	case <-time.After(3 * time.Second):
		t.Fatal("client did not stop after disconnection")
	}
}
