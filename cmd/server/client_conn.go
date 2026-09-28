package main

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

var errSendQueueFull = errors.New("client send queue full")

const defaultWriteTimeout = 10 * time.Second

type outboundMessage struct {
	data []byte
	done chan writeResult
}

type writeResult struct {
	n   int
	err error
}

type serverClient struct {
	net.Conn
	out          chan outboundMessage
	done         chan struct{}
	writeTimeout time.Duration
	closeOnce    sync.Once
	locale       string
}

func newServerClient(conn net.Conn) *serverClient {
	return newServerClientWithWriteTimeout(conn, defaultWriteTimeout)
}

func newServerClientWithWriteTimeout(conn net.Conn, timeout time.Duration) *serverClient {
	client := &serverClient{
		Conn:         conn,
		out:          make(chan outboundMessage, 64),
		done:         make(chan struct{}),
		writeTimeout: timeout,
	}
	go client.writeLoop()
	return client
}

func (client *serverClient) writeLoop() {
	for {
		select {
		case message := <-client.out:
			err := client.Conn.SetWriteDeadline(time.Now().Add(client.writeTimeout))
			n := 0
			if err == nil {
				n, err = client.Conn.Write(message.data)
			}
			if err == nil && n != len(message.data) {
				err = io.ErrShortWrite
			}
			if message.done != nil {
				message.done <- writeResult{n: n, err: err}
			}
			if err != nil {
				client.Close()
				return
			}
		case <-client.done:
			return
		}
	}
}

func (client *serverClient) Write(data []byte) (int, error) {
	message := outboundMessage{
		data: append([]byte(nil), data...),
		done: make(chan writeResult, 1),
	}
	select {
	case client.out <- message:
	case <-client.done:
		return 0, net.ErrClosed
	}
	select {
	case result := <-message.done:
		return result.n, result.err
	case <-client.done:
		return 0, net.ErrClosed
	}
}

func (client *serverClient) enqueueResponse(line string) (<-chan writeResult, error) {
	message := outboundMessage{
		data: []byte(line + "\n"),
		done: make(chan writeResult, 1),
	}
	select {
	case client.out <- message:
		return message.done, nil
	case <-client.done:
		return nil, net.ErrClosed
	default:
		return nil, errSendQueueFull
	}
}

func (client *serverClient) waitResponse(done <-chan writeResult) error {
	select {
	case result := <-done:
		return result.err
	case <-client.done:
		return net.ErrClosed
	}
}

func (client *serverClient) enqueueEvent(line string) {
	select {
	case <-client.done:
		return
	default:
	}
	select {
	case client.out <- outboundMessage{data: []byte(line + "\n")}:
	case <-client.done:
	default:
		client.Close()
	}
}

func (client *serverClient) Close() error {
	var err error
	client.closeOnce.Do(func() {
		close(client.done)
		err = client.Conn.Close()
	})
	return err
}
