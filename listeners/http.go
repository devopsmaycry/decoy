package listeners

import (
	"decoy/config"
	"decoy/logger"
	"decoy/web"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// realIP extracts the client IP, preferring X-Forwarded-For when present (proxy/LB deployments).
func realIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if host, _, err := net.SplitHostPort(strings.SplitN(xff, ",", 2)[0]); err == nil {
			return host
		}
		return strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func setDecoyHeaders(w http.ResponseWriter, opts config.HttpServerConfig) {
	w.Header().Set("Server", opts.Server)
	w.Header().Set("X-Powered-By", opts.XPoweredBy)
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
}

func renderLoginPage(errorMsg string) []byte {
	return []byte(strings.ReplaceAll(string(web.LoginPage), "{{ERROR}}", errorMsg))
}

func StartHTTP(cfg config.HttpServerConfig, log *logger.Logger) {
	mux := http.NewServeMux()

	// Catch-all: log probes to any path not matching the configured login path.
	if cfg.Path != "/" {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if !connLimiter.allowRequest(realIP(r)) {
				log.Log("http_rate_limited", map[string]any{"port": cfg.Port, "remote_ip": realIP(r)})
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			log.Log("http_probe", map[string]any{
				"port":      cfg.Port,
				"remote_ip": realIP(r),
				"method":    r.Method,
				"uri":       r.RequestURI,
				"ssl":       cfg.SslEnabled,
			})
			setDecoyHeaders(w, cfg)
			http.NotFound(w, r)
		})
	}

	mux.HandleFunc(cfg.Path, func(w http.ResponseWriter, r *http.Request) {
		if !connLimiter.allowRequest(realIP(r)) {
			log.Log("http_rate_limited", map[string]any{"port": cfg.Port, "remote_ip": realIP(r)})
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		body, _ := io.ReadAll(io.LimitReader(r.Body, 4096))

		headers := map[string]string{}
		for k, v := range r.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}

		fields := map[string]any{
			"port":      cfg.Port,
			"remote_ip": realIP(r),
			"method":    r.Method,
			"uri":       r.RequestURI,
			"query":     r.URL.RawQuery,
			"headers":   headers,
			"body":      string(body),
			"ssl":       cfg.SslEnabled,
		}
		if len(body) == 4096 {
			fields["body_truncated"] = true
		}

		if r.Method == http.MethodPost {
			if params, err := url.ParseQuery(string(body)); err == nil {
				fields["username"] = params.Get("username")
				fields["password"] = params.Get("password")
			}
		}

		log.Log("http_request", fields)
		setDecoyHeaders(w, cfg)

		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			w.Write(renderLoginPage(`<div class="error-msg">&#x26A0; Invalid username or password. Please try again.</div>`))
			return
		}

		if cfg.RedirectUrl != "" {
			http.Redirect(w, r, cfg.RedirectUrl, http.StatusFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		if cfg.WebEnabled {
			w.Write(renderLoginPage(""))
		}
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Log("http_listening", map[string]any{"port": cfg.Port, "ssl": cfg.SslEnabled})

	var err error
	if cfg.SslEnabled {
		err = srv.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
	} else {
		err = srv.ListenAndServe()
	}
	if err != nil {
		log.Log("http_listen_error", map[string]any{"port": cfg.Port, "path": cfg.Path, "ssl": cfg.SslEnabled, "error": err.Error()})
	}
}
