package config

import (
	"flag"
	"log"
	"os"

	"decoy/logger"

	"gopkg.in/yaml.v3"
)

type listenerConfig struct {
	Port string `yaml:"port"`
	Type string `yaml:"type"`
	Ssl  bool   `yaml:"ssl"`
	Path string `yaml:"path"`
}

type SshConfig struct {
	LogUsername      bool   `yaml:"logUsername"`
	LogPassword      bool   `yaml:"logPassword"`
	SshServerVersion string `yaml:"sshShowedVersion"`
}

type HttpsConfig struct {
	CertFile string `yaml:"serverCertificate"`
	KeyFile  string `yaml:"serverCertificateKey"`
}

type SyslogConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Server     string `yaml:"server"`
	Port       string `yaml:"port"`
	CliEnabled bool   `yaml:"cliEnabled"`
}

type Config struct {
	Version   string           `yaml:"version"`
	Listeners []listenerConfig `yaml:"listeners"`
	Ssh       SshConfig        `yaml:"ssh"`
	Https     HttpsConfig      `yaml:"https"`
	Syslog    SyslogConfig     `yaml:"syslog"`
	LogLevel  string           `yaml:"logLevel"`
}

func Load(logger *logger.Logger) Config {
	logger.Log("Loading config file", nil)
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	data, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("cannot read config: %v", err)
	}

	cfg := Config{
		Version: "1.2",
		Listeners: []listenerConfig{
			{Port: "2222", Type: "ssh", Ssl: false},
			{Port: "80808", Type: "http", Ssl: false},
		},
		Ssh: SshConfig{
			LogUsername:      false,
			LogPassword:      false,
			SshServerVersion: "SSH-2.0-OpenSSH_8.9p1 Debian-3",
		},
		Https: HttpsConfig{
			CertFile: "",
			KeyFile:  "",
		},
		Syslog: SyslogConfig{
			Enabled:    false,
			CliEnabled: true,
		},
		LogLevel: "info",
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("cannot parse config: %v", err)
	}

	return cfg
}
