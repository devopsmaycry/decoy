package listeners

import (
	"decoy/logger"
	"decoy/services"
	"net"
	"time"
)

func StartTCP(port string, service string, log *logger.Logger) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Log("tcp_listen_error", map[string]any{"port": port, "error": err.Error()})
		return
	}
	log.Log("tcp_listening", map[string]any{"port": port, "service": service})

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Log("tcp_accept_error", map[string]any{"port": port, "error": err.Error()})
			continue
		}
		go func(c net.Conn) {
			remoteAddr := c.RemoteAddr().String()
			if !connLimiter.allowConn(remoteAddr) {
				log.Log("tcp_rate_limited", map[string]any{"port": port, "remote_ip": remoteAddr, "service": service})
				c.Close()
				return
			}
			defer connLimiter.releaseConn()
			handleTCP(c, port, service, log)
		}(conn)
	}
}

func handleTCP(conn net.Conn, port string, service string, log *logger.Logger) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	if svc := services.Get(service); svc != nil {
		if _, err := conn.Write(svc.Banner()); err != nil {
			log.Log("tcp_write_error", map[string]any{"port": port, "remote_ip": conn.RemoteAddr().String(), "error": err.Error()})
			return
		}
	}

	buf := make([]byte, 4096)
	n, _ := conn.Read(buf)

	fields := map[string]any{
		"port":      port,
		"remote_ip": conn.RemoteAddr().String(),
		"service":   service,
	}
	if n > 0 {
		fields["data"] = string(buf[:n])
	}
	log.Log("tcp_connection", fields)
}
