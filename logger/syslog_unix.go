//go:build !windows

package logger

import (
	"fmt"
	"log/syslog"
	"os"

	"decoy/config"
)

func initSyslog(cfg config.LogConfig) syslogSink {
	if !cfg.SyslogEnabled {
		return nil
	}

	w, err := syslog.Dial("udp", fmt.Sprintf("%s:%s", cfg.SyslogServer, cfg.SyslogPort), syslog.LOG_INFO|syslog.LOG_DAEMON, "decoy")
	if err != nil {
		fmt.Fprintf(os.Stderr, "syslog connect error: %v\n", err)			
		return nil
	}
	return w
}