//go:build windows

package logger

import (
	"fmt"
	"os"

	"decoy/config"
)

func initSyslog(cfg config.LogConfig) syslogSink {
	if cfg.SyslogEnabled {
		fmt.Fprintf(os.Stderr, "syslog is not supported on Windows; ignoring syslog config (use fileLoggingEnabled + a Windows log forwarder instead)\n")
	}
	return nil
}
