package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"decoy/config"
)

type syslogSink interface {
	Info(string) error
	Close() error
}

type Logger struct {
	mu                 sync.Mutex
	cliEnabled         bool
	fileLoggingEnabled bool
	filePath           string
	fileWriter         *os.File
	syslogWriter       syslogSink
}

func New(cfg config.LogConfig) *Logger {
	l := &Logger{cliEnabled: cfg.CliEnabled, fileLoggingEnabled: cfg.FileLoggingEnabled, filePath: cfg.FilePath}

	l.syslogWriter = initSyslog(cfg)

	if cfg.FileLoggingEnabled {
		f, err := os.OpenFile(cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "file log error: %v\n", err)
		} else {
			l.fileWriter = f
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

	if l.fileWriter != nil {
		if _, err := l.fileWriter.Write(append(b, '\n')); err != nil {
			fmt.Fprintf(os.Stderr, "file log error: %v\n", err)
		}
	}
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var firstError error
	if l.syslogWriter != nil {
		if err := l.syslogWriter.Close(); err != nil {
			firstError = err
		}
		l.syslogWriter = nil
	}
	if l.fileWriter != nil {
		if err := l.fileWriter.Close(); err != nil {
			firstError = err
		}
		l.fileWriter = nil
	}

	return firstError
}
