package main

import (
	"bufio"
	"fmt"
	"net"
	"testing"
	"time"
)

func nextMessage(t *testing.T, client *protocolClient) serverMessage {
	t.Helper()
	select {
	case message, ok := <-client.incoming:
		if !ok {
			t.Fatal("message channel closed")
		}
		return message
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for message")
		return serverMessage{}
	}
}

func TestProtocolMatchesResponsesAroundEvents(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer serverConn.Close()
	serverConn.SetDeadline(time.Now().Add(2 * time.Second))
	client := newProtocolClient(clientConn)
	defer client.Close()
	reader := bufio.NewScanner(serverConn)
	if _, err := fmt.Fprintln(serverConn, "OK hello proto=1"); err != nil {
		t.Fatal(err)
	}
	if message := nextMessage(t, client); message.kind != messageGreeting {
		t.Fatalf("greeting = %+v", message)
	}
	if err := client.Send("LOOK"); err != nil {
		t.Fatal(err)
	}
	if err := client.Send("WHO"); err != nil {
		t.Fatal(err)
	}
	if !reader.Scan() || reader.Text() != "LOOK" {
		t.Fatalf("first command = %q, err = %v", reader.Text(), reader.Err())
	}
	if _, err := fmt.Fprintln(serverConn, "EVT STATS players=2"); err != nil {
		t.Fatal(err)
	}
	if message := nextMessage(t, client); message.kind != messageEvent || message.command != "" {
		t.Fatalf("event = %+v", message)
	}
	if _, err := fmt.Fprintln(serverConn, `OK {"room":{"id":"r"}}`); err != nil {
		t.Fatal(err)
	}
	if message := nextMessage(t, client); message.kind != messageResponse || message.command != "LOOK" || message.request != "LOOK" {
		t.Fatalf("LOOK response = %+v", message)
	}
	if !reader.Scan() || reader.Text() != "WHO" {
		t.Fatalf("second command = %q, err = %v", reader.Text(), reader.Err())
	}
	if _, err := fmt.Fprintln(serverConn, "OK players=2"); err != nil {
		t.Fatal(err)
	}
	if message := nextMessage(t, client); message.kind != messageResponse || message.command != "WHO" {
		t.Fatalf("WHO response = %+v", message)
	}
}

func TestProtocolRejectsBlankAndMultilineCommands(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer serverConn.Close()
	client := newProtocolClient(clientConn)
	defer client.Close()
	for _, line := range []string{"", " \t ", "LOOK\nWHO", "LOOK\rWHO"} {
		if err := client.Send(line); err == nil {
			t.Errorf("Send(%q) succeeded", line)
		}
	}
}
