package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

func main() {
	address := "127.0.0.1:4242"
	if len(os.Args) > 2 {
		log.Fatalf("usage: %s [host:port]", os.Args[0])
	}
	if len(os.Args) == 2 {
		address = os.Args[1]
	}

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("connection failed: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connected to", address)
	runClient(conn.(*net.TCPConn), os.Stdin)
}

type clientConn interface {
	net.Conn
	CloseWrite() error
}

func runClient(conn clientConn, stdin io.Reader) {
	serverDone := make(chan error, 1)
	quitAcknowledged := make(chan struct{}, 1)
	quitRejected := make(chan struct{}, 1)
	var pendingMu sync.Mutex
	var pending []bool

	go func() {
		scanner := bufio.NewScanner(conn)
		firstLine := true

		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
			if firstLine {
				firstLine = false
				if line == "OK hello proto=1" {
					continue
				}
			}
			if line != "OK" && !strings.HasPrefix(line, "OK ") && !strings.HasPrefix(line, "ERR ") {
				continue
			}
			pendingMu.Lock()
			isQuit := false
			if len(pending) > 0 {
				isQuit = pending[0]
				pending = pending[1:]
			}
			pendingMu.Unlock()
			if !isQuit {
				continue
			}
			if line == "OK bye" {
				select {
				case quitAcknowledged <- struct{}{}:
				default:
				}
			} else if strings.HasPrefix(line, "ERR ") {
				select {
				case quitRejected <- struct{}{}:
				default:
				}
			}
		}

		serverDone <- scanner.Err()
	}()

	input := make(chan string)

	go func() {
		defer close(input)

		scanner := bufio.NewScanner(stdin)

		for scanner.Scan() {
			input <- scanner.Text()
		}

		if err := scanner.Err(); err != nil {
			log.Println("input error:", err)
		}
	}()

	for {
		select {
		case line, ok := <-input:
			if !ok {
				if err := conn.CloseWrite(); err != nil {
					log.Println("close write failed:", err)
					return
				}
				input = nil
				continue
			}

			if strings.TrimSpace(line) == "" {
				continue
			}

			isQuit := strings.EqualFold(strings.TrimSpace(line), "QUIT")
			pendingMu.Lock()
			pending = append(pending, isQuit)
			pendingMu.Unlock()
			if _, err := fmt.Fprintln(conn, line); err != nil {
				log.Println("send failed:", err)
				return
			}

			if isQuit {
				select {
				case <-quitAcknowledged:
					return
				case <-quitRejected:
					continue
				case err := <-serverDone:
					if err != nil {
						log.Println("receive error:", err)
					}
				}
				return
			}

		case <-quitAcknowledged:
			return

		case err := <-serverDone:
			if err != nil {
				log.Println("receive error:", err)
			}
			fmt.Println("Disconnected from server")
			return
		}
	}
}
