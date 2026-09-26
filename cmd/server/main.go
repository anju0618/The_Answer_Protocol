package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	server := NewServer()

	listener, err := net.Listen("tcp", "127.0.0.1:4242")
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
