package main

import (
	"net"
	"os"
)

func main() {
	closeLog := setupLogging()
	defer closeLog()

	server := NewServer()

	listener, err := net.Listen("tcp", ":4242")
	if err != nil {
		closeLog()
		fatal("listen_failed", err)
	}
	defer listener.Close()

	logger.Info("server_started", "port", 4242, "pid", os.Getpid())

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Warn("accept_failed", "error", err.Error())
			continue
		}

		go server.handleClient(conn)
	}
}
