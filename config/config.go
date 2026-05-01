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

type LogConfig struct {
	SyslogEnabled      bool   `yaml:"syslogEnabled"`
	SyslogServer       string `yaml:"syslogServer"`
	SyslogPort         string `yaml:"syslogPort"`
	CliEnabled         bool   `yaml:"cliEnabled"`
	FileLoggingEnabled bool   `yaml:"fileLoggingEnabled"`
	FilePath           string `yaml:"filePath"`
	LogLevel           string `yaml:"logLevel"`
}

type Config struct {
	Version       string               `yaml:"version"`
	Listeners     []listenerConfig     `yaml:"listeners"`
	HttpListeners []HttpListenerConfig `yaml:"httpListeners"`
	Ssh           SshConfig            `yaml:"ssh"`
	Service       ServiceConfig        `yaml:"service"`
	Log           LogConfig            `yaml:"log"`
}

func Load() Config {
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	data, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("cannot read config: %v", err)
	}

	cfg := Config{
		Version:   "1.3",
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
		Log: LogConfig{
			SyslogEnabled:      false,
			CliEnabled:         true,
			FileLoggingEnabled: true,
			FilePath:           "/var/log/decoy.log",
			LogLevel:           "info",
		},
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("cannot parse config: %v", err)
	}

	return cfg
}
