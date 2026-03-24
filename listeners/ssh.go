package listeners

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"decoy/logger"
	"fmt"
	"net"

	"golang.org/x/crypto/ssh"
)

type SSHOptions struct {
	LogUsername bool `yaml:"logUsername"`
	LogPassword bool `yaml:"logPassword"`
}

func StartSSH(port string, log *logger.Logger, opts SSHOptions) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Log("ssh_keygen_error", map[string]any{"error": err.Error()})
		return
	}
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		log.Log("ssh_keygen_error", map[string]any{"error": err.Error()})
		return
	}

	config := &ssh.ServerConfig{
		ServerVersion: "SSH-2.0-OpenSSH_8.9p1 Ubuntu-3ubuntu0.6",
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			fields := map[string]any{
				"port":           port,
				"remote_ip":      conn.RemoteAddr().String(),
				"client_version": string(conn.ClientVersion()),
			}
			if opts.LogUsername {
				fields["username"] = conn.User()
			} else {
				fields["username"] = "********"
			}
			if opts.LogPassword {
				fields["password"] = string(password)
			} else {
				fields["password"] = "********"
			}
			log.Log("ssh_auth_attempt", fields)
			return nil, fmt.Errorf("denied")
		},
	}
	config.AddHostKey(signer)

	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Log("ssh_listen_error", map[string]any{"port": port, "error": err.Error()})
		return
	}
	log.Log("ssh_listening", map[string]any{"port": port})

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Log("ssh_accept_error", map[string]any{"port": port, "error": err.Error()})
			continue
		}
		go func(c net.Conn) {
			defer c.Close()
			ssh.NewServerConn(c, config)
		}(conn)
	}
}
