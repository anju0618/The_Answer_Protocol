package main

import (
	"log"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":4242")
	if err != nil {
		log.Fatalf("listen failed: %v", err)
	}
	log.Println("listening on :4242")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}
		log.Printf("connection from %s", conn.RemoteAddr())
		conn.Close()
	}
}
