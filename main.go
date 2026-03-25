package main

import (
	"decoy/listeners"
	"decoy/logger"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/yaml.v3"
)

type ListenerConfig struct {
	Port string `yaml:"port"`
	Type string `yaml:"type"`
	Ssl  bool   `yaml:"ssl"`
}

type Config struct {
	Listeners    []ListenerConfig       `yaml:"listeners"`
	SSHoptions   listeners.SSHOptions   `yaml:"ssh"`
	Syslog       logger.SyslogConfig    `yaml:"syslog"`
	HttpsOptions listeners.HttpsOptions `yaml:"https"`
}

func main() {
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	fmt.Print("==========================\n")
	fmt.Print("| Starting Decoy Service |\n")
	fmt.Println("|    A lightweight \U0001F36F    |")
	fmt.Print("==========================\n")

	data, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("cannot read config: %v", err)
	}

	cfg := Config{
		SSHoptions: listeners.SSHOptions{
			LogUsername: false,
			LogPassword: false,
		},
		Syslog: logger.SyslogConfig{
			Enabled:    false,
			CliEnabled: true,
		},
		HttpsOptions: listeners.HttpsOptions{
			ServerCert: "",
			ServerKey:  "",
		},
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("cannot parse config: %v", err)
	}

	// Validate HTTPS config upfront
	for _, l := range cfg.Listeners {
		if l.Type == "http" && l.Ssl {
			if cfg.HttpsOptions.ServerCert == "" || cfg.HttpsOptions.ServerKey == "" {
				log.Fatal("https listener configured but https.serverCertificate or https.serverCertificateKey is missing")
			}
		}
	}

	appLog := logger.New(cfg.Syslog)
	appLog.Log("decoy_started", map[string]any{"listener_count": len(cfg.Listeners)})

	for _, l := range cfg.Listeners {
		switch l.Type {
		case "tcp":
			go listeners.StartTCP(l.Port, appLog)
		case "http":
			go listeners.StartHTTP(l.Port, appLog, cfg.HttpsOptions, l.Ssl)
		case "ssh":
			go listeners.StartSSH(l.Port, appLog, cfg.SSHoptions)
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
