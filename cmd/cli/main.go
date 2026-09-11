// The Answer Protocol (TAP) の生プロトコル中継CLIクライアント。
// 標準入力に打った行をそのままTCP接続に送り、サーバーからの応答行をそのまま
// 標準出力に表示するだけの薄いクライアント。プロトコルをユーザーフレンドリーな
// コマンドに翻訳する機能は持たない(README「Instructions」参照)。
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

// main はサーバーに接続し、runClient に処理を委譲する。
// コマンドライン引数で接続先の host:port を指定できる(省略時は127.0.0.1:4242)。
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

// clientConn は net.Conn に半クローズ(送信方向だけ閉じる)ができる
// CloseWrite を足したインターフェース。標準入力がEOFになった時に、
// 受信はまだ続けたいまま送信だけを閉じるために使う。
type clientConn interface {
	net.Conn
	CloseWrite() error
}

// runClient は「サーバーからの受信を読み続けるgoroutine」「標準入力を
// 読み続けるgoroutine」「両者からのメッセージをselectでさばくメインループ」の
// 3つで構成される。QUITを送った後は、サーバーからの応答(OK bye / ERR)を
// 待ってから終了するようにしている(送信直後に打ち切ると応答を取りこぼすため)。
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
