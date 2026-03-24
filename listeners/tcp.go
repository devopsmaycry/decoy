package listeners

import (
	"decoy/logger"
	"net"
	"time"
)

func StartTCP(port string, log *logger.Logger) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Log("tcp_listen_error", map[string]any{"port": port, "error": err.Error()})
		return
	}
	log.Log("tcp_listening", map[string]any{"port": port})

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Log("tcp_accept_error", map[string]any{"port": port, "error": err.Error()})
			continue
		}
		go handleTCP(conn, port, log)
	}
}

func handleTCP(conn net.Conn, port string, log *logger.Logger) {
	defer conn.Close()

	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, _ = conn.Read(buf)

	log.Log("tcp_connection", map[string]any{
		"port":      port,
		"remote_ip": conn.RemoteAddr().String(),
	})
}
