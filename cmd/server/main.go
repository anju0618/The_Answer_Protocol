// The Answer Protocol (TAP) のTCPサーバー本体。エントリーポイント(main)は
// TCPポート4242で待ち受けを開始し、接続ごとに1 goroutineを立てるだけの
// 薄いラッパーで、実際のコマンド処理は Server.handleClient 以降(他ファイル)に
// 委譲する。
package main

import (
	"fmt"
	"log"
	"net"
)

// main はサーバーを起動し、TCP接続を待ち受け続ける。
// 接続ごとにgoroutineを1つ立てる「接続ごとgoroutine」方式を採用している
// (README「Architecture」参照)。
func main() {
	server := NewServer()

	listener, err := net.Listen("tcp", ":4242")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Println("Server started on port 4242")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		fmt.Println("New connection:", conn.RemoteAddr())
		go server.handleClient(conn)
	}
}
