package main

import (
	"decoy/config"
	"decoy/listeners"
	"decoy/logger"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	fmt.Print("==========================\n")
	fmt.Print("| Starting Decoy Service |\n")
	fmt.Println("|    A lightweight \U0001F36F    |")
	fmt.Print("==========================\n")

	// load the whole config
	cfg := config.Load()

	// Validate config version
	if cfg.Version != "1.2" {
		log.Fatalf("unsupported config version: %s", cfg.Version)
	}

	// Validate HTTPS config upfront
	for _, l := range cfg.Listeners {
		if l.Type == "http" && l.Ssl {
			if cfg.Https.CertFile == "" || cfg.Https.KeyFile == "" {
				log.Fatal("https listener configured but https.serverCertificate or https.serverCertificateKey is missing")
			}
		}
		if l.Type == "http" && l.Path == "" {
			log.Fatal("http listener path is missing")
		}
	}

	appLog := logger.New(logger.SyslogConfig(cfg.Syslog))
	appLog.Log("decoy_started", map[string]any{"listener_count": len(cfg.Listeners)})

	for _, l := range cfg.Listeners {
		switch l.Type {
		case "tcp":
			go listeners.StartTCP(l.Port, appLog)
		case "http":
			go listeners.StartHTTP(l.Port, l.Ssl, l.Path, appLog, cfg.Https)
		case "ssh":
			go listeners.StartSSH(l.Port, appLog, cfg.Ssh)
		default:
			appLog.Log("unknown_listener_type", map[string]any{
				"type": l.Type,
				"port": l.Port,
			})
		}
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	appLog.Log("decoy_stopped", nil)
}
