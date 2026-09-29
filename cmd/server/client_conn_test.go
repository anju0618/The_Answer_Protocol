package main

import (
	"net"
	"testing"
	"time"
)

func TestServerClientWriteTimesOut(t *testing.T) {
	serverConn, peerConn := net.Pipe()
	defer peerConn.Close()
	client := newServerClientWithWriteTimeout(serverConn, 50*time.Millisecond)
	defer client.Close()

	result := make(chan error, 1)
	go func() {
		_, err := client.Write([]byte("peer does not read"))
		result <- err
	}()
	select {
	case err := <-result:
		if err == nil {
			t.Fatal("write succeeded without a reader")
		}
	case <-time.After(time.Second):
		t.Fatal("write did not time out")
	}
	select {
	case <-client.done:
	case <-time.After(time.Second):
		t.Fatal("timed-out connection stayed open")
	}
}
