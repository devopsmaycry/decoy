package config

import (
	"flag"
	"log"
	"os"

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

type HttpServerConfig struct {
	Server     string `yaml:"Server"`
	XPoweredBy string `yaml:"X-Powered-By"`
	CertFile   string `yaml:"serverCertificate"`
	KeyFile    string `yaml:"serverCertificateKey"`
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
	Https     HttpServerConfig `yaml:"httpServer"`
	Syslog    SyslogConfig     `yaml:"syslog"`
	LogLevel  string           `yaml:"logLevel"`
}

func Load() Config {
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	data, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("cannot read config: %v", err)
	}

	cfg := Config{
		Version: "1.2",
		Listeners: []listenerConfig{
			{Port: "2222", Type: "ssh", Ssl: false, Path: "/"},
			{Port: "80808", Type: "http", Ssl: false, Path: "/"},
		},
		Ssh: SshConfig{
			LogUsername:      false,
			LogPassword:      false,
			SshServerVersion: "SSH-2.0-OpenSSH_8.9p1 Debian-3",
		},
		Https: HttpServerConfig{
			CertFile:   "",
			KeyFile:    "",
			Server:     "Apache/2.2.22 (Debian)",
			XPoweredBy: "PHP/5.6.40",
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
