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
}

type Config struct {
	Listeners  []ListenerConfig     `yaml:"listeners"`
	SSHoptions listeners.SSHOptions `yaml:"ssh"`
	Syslog     logger.SyslogConfig  `yaml:"syslog"`
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
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("cannot parse config: %v", err)
	}

	appLog := logger.New(cfg.Syslog)
	appLog.Log("decoy_started", map[string]any{"listener_count": len(cfg.Listeners)})

	for _, l := range cfg.Listeners {
		switch l.Type {
		case "tcp":
			go listeners.StartTCP(l.Port, appLog)
		case "http":
			go listeners.StartHTTP(l.Port, appLog)
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
