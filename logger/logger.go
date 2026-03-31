package logger

import (
	"encoding/json"
	"fmt"
	"log/syslog"
	"os"
	"sync"
	"time"
)

type Logger struct {
	mu           sync.Mutex
	cliEnabled   bool
	syslogWriter *syslog.Writer
}

type SyslogConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Server     string `yaml:"server"`
	Port       string `yaml:"port"`
	CliEnabled bool   `yaml:"cliEnabled"`
}

func New(cfg SyslogConfig) *Logger {
	l := &Logger{cliEnabled: cfg.CliEnabled}

	if cfg.Enabled {
		w, err := syslog.Dial("udp", fmt.Sprintf("%s:%s", cfg.Server, cfg.Port), syslog.LOG_INFO|syslog.LOG_DAEMON, "decoy")
		if err != nil {
			fmt.Fprintf(os.Stderr, "syslog connect error: %v\n", err)
		} else {
			l.syslogWriter = w
		}
	}

	return l
}

func (l *Logger) Log(event string, fields map[string]any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := map[string]any{
		"time":  time.Now().UTC().Format(time.RFC3339),
		"event": event,
	}
	for k, v := range fields {
		entry[k] = v
	}
	b, err := json.Marshal(entry)
	if err != nil {
		b = fmt.Appendf(nil, `{"time":%q,"event":%q,"logger_error":%q}`,
			entry["time"], entry["event"], err.Error())
	}

	if l.cliEnabled {
		os.Stdout.Write(append(b, '\n'))
	}
	if l.syslogWriter != nil {
		l.syslogWriter.Info(string(b))
	}
}
