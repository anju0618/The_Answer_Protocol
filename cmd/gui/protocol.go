package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

type messageKind uint8

const (
	messageGreeting messageKind = iota
	messageResponse
	messageEvent
	messageNotice
	messageDisconnect
)

type serverMessage struct {
	kind    messageKind
	command string
	request string
	line    string
	err     error
}

type protocolClient struct {
	conn     net.Conn
	outgoing chan string
	incoming chan serverMessage
	done     chan struct{}
	closeOne sync.Once
	mu       sync.Mutex
	pending  []string
}

func newProtocolClient(conn net.Conn) *protocolClient {
	client := &protocolClient{
		conn:     conn,
		outgoing: make(chan string, 64),
		incoming: make(chan serverMessage, 128),
		done:     make(chan struct{}),
	}
	go client.writeLoop()
	go client.readLoop()
	return client
}

func (client *protocolClient) Send(line string) error {
	if strings.TrimSpace(line) == "" || !utf8.ValidString(line) || strings.ContainsAny(line, "\r\n") {
		return errors.New("invalid command line")
	}
	select {
	case <-client.done:
		return net.ErrClosed
	default:
	}
	select {
	case client.outgoing <- line:
		return nil
	case <-client.done:
		return net.ErrClosed
	default:
		return errors.New("command queue is full")
	}
}

func (client *protocolClient) Close() {
	client.closeOne.Do(func() {
		close(client.done)
		client.conn.Close()
	})
}

func (client *protocolClient) canPoll() bool {
	client.mu.Lock()
	pending := len(client.pending)
	client.mu.Unlock()
	return pending+len(client.outgoing) < 4
}

func (client *protocolClient) writeLoop() {
	for {
		select {
		case line := <-client.outgoing:
			client.mu.Lock()
			client.pending = append(client.pending, line)
			client.mu.Unlock()
			if err := client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				client.conn.Close()
				return
			}
			if _, err := fmt.Fprintln(client.conn, line); err != nil {
				client.conn.Close()
				return
			}
		case <-client.done:
			return
		}
	}
}

func (client *protocolClient) readLoop() {
	defer close(client.incoming)
	defer client.Close()
	scanner := bufio.NewScanner(client.conn)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	firstLine := true
	for scanner.Scan() {
		line := scanner.Text()
		message := serverMessage{line: line}
		switch {
		case firstLine && line == "OK hello proto=1":
			message.kind = messageGreeting
		case strings.HasPrefix(line, "EVT "):
			message.kind = messageEvent
		case line == "OK" || strings.HasPrefix(line, "OK ") || strings.HasPrefix(line, "ERR "):
			message.kind = messageResponse
			client.mu.Lock()
			if len(client.pending) > 0 {
				message.request = client.pending[0]
				message.command = strings.ToUpper(strings.Fields(message.request)[0])
				client.pending = client.pending[1:]
			}
			client.mu.Unlock()
		default:
			message.kind = messageNotice
		}
		firstLine = false
		select {
		case client.incoming <- message:
		case <-client.done:
			return
		}
	}
	err := scanner.Err()
	if err == nil {
		err = io.EOF
	}
	select {
	case client.incoming <- serverMessage{kind: messageDisconnect, err: err}:
	case <-client.done:
	}
}
