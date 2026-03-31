package main

import (
	"decoy/config"
	"decoy/listeners"
	"decoy/logger"
	"decoy/services"
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

	// Validate httpListener config upfront
	for _, hL := range cfg.HttpListeners {
		if hL.SslEnabled {
			if hL.CertFile == "" || hL.KeyFile == "" {
				log.Fatal("https listener configured but https.serverCertificate or https.serverCertificateKey is missing")
			}
		}
		if hL.Path == "" {
			log.Fatal("http listener path is missing")
		}
	}

	// validate service banners
	if cfg.Service.FtpBanner == "" {
		cfg.Service.FtpBanner = "220 Microsoft FTP Service"
	}
	if cfg.Service.RedisBanner == "" {
		cfg.Service.RedisBanner = "-NOAUTH Authentication required."
	}
	if cfg.Service.SmtpBanner == "" {
		cfg.Service.SmtpBanner = "220 mail.corp.local ESMTP Postfix (Debian/GNU)"
	}

	services.Init(cfg.Service.FtpBanner, cfg.Service.RedisBanner, cfg.Service.SmtpBanner)

	appLog := logger.New(logger.SyslogConfig(cfg.Syslog))
	appLog.Log("decoy_started", map[string]any{"listener_count": len(cfg.Listeners)})

	for _, l := range cfg.Listeners {
		switch l.Type {
		case "tcp":
			go listeners.StartTCP(l.Port, l.Service, appLog)
		case "ssh":
			go listeners.StartSSH(l.Port, appLog, cfg.Ssh)
		default:
			appLog.Log("unknown_listener_type", map[string]any{
				"type": l.Type,
			})
		}
	}

	for i := range cfg.HttpListeners {
		hL := &cfg.HttpListeners[i]
		if hL.Server == "" {
			hL.Server = "Apache/2.2.22 (Debian)"
		}
		if hL.XPoweredBy == "" {
			hL.XPoweredBy = "PHP/5.6.40"
		}
		go listeners.StartHTTP(hL.HttpServerConfig, appLog)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	appLog.Log("decoy_stopped", nil)
}
