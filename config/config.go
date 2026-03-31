package config

import (
	"flag"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type listenerConfig struct {
	Port    string `yaml:"port"`
	Type    string `yaml:"type"`
	Service string `yaml:"service"`
}

type HttpListenerConfig struct {
	HttpServerConfig `yaml:",inline"`
}

type SshConfig struct {
	LogUsername      bool   `yaml:"logUsername"`
	LogPassword      bool   `yaml:"logPassword"`
	SshServerVersion string `yaml:"sshShowedVersion"`
}

type ServiceConfig struct {
	FtpBanner   string `yaml:"ftpBanner"`
	RedisBanner string `yaml:"redisBanner"`
	SmtpBanner  string `yaml:"smtpBanner"`
}

type HttpServerConfig struct {
	Port        string `yaml:"port"`
	Server      string `yaml:"Server"`
	Path        string `yaml:"path"`
	WebEnabled  bool   `yaml:"websiteEnabled"`
	RedirectUrl string `yaml:"redirectUrl"`
	XPoweredBy  string `yaml:"X-Powered-By"`
	SslEnabled  bool   `yaml:"sslEnabled"`
	CertFile    string `yaml:"serverCertificate"`
	KeyFile     string `yaml:"serverCertificateKey"`
}

type SyslogConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Server     string `yaml:"server"`
	Port       string `yaml:"port"`
	CliEnabled bool   `yaml:"cliEnabled"`
}

type Config struct {
	Version       string               `yaml:"version"`
	Listeners     []listenerConfig     `yaml:"listeners"`
	HttpListeners []HttpListenerConfig `yaml:"httpListeners"`
	Ssh           SshConfig            `yaml:"ssh"`
	Service       ServiceConfig        `yaml:"service"`
	Syslog        SyslogConfig         `yaml:"syslog"`
	LogLevel      string               `yaml:"logLevel"`
}

func Load() Config {
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	data, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("cannot read config: %v", err)
	}

	cfg := Config{
		Version:   "1.2",
		Listeners: []listenerConfig{},
		Ssh: SshConfig{
			LogUsername:      false,
			LogPassword:      false,
			SshServerVersion: "SSH-2.0-OpenSSH_8.9p1 Debian-3",
		},
		Service: ServiceConfig{
			FtpBanner:   "220 Microsoft FTP Service",
			RedisBanner: "+PONG",
			SmtpBanner:  "220 mail.example.com ESMTP Postfix",
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
