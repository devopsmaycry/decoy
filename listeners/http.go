package listeners

import (
	"decoy/config"
	"decoy/logger"
	"decoy/web"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

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

func StartHTTP(port string, ssl bool, path string,
	log *logger.Logger, opts config.HttpServerConfig) {
	mux := http.NewServeMux()

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(io.LimitReader(r.Body, 4096))

		headers := map[string]string{}
		for k, v := range r.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}

		fields := map[string]any{
			"port":      port,
			"remote_ip": r.RemoteAddr,
			"method":    r.Method,
			"uri":       r.RequestURI,
			"query":     r.URL.RawQuery,
			"headers":   headers,
			"body":      string(body),
			"ssl":       ssl,
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
		setDecoyHeaders(w, opts)

		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			w.Write(renderLoginPage(`<div class="error-msg">&#x26A0; Invalid username or password. Please try again.</div>`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(renderLoginPage(""))
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Log("http_listening", map[string]any{"port": port, "ssl": ssl})

	var err error
	if ssl {
		err = srv.ListenAndServeTLS(opts.CertFile, opts.KeyFile)
	} else {
		err = srv.ListenAndServe()
	}
	if err != nil {
		log.Log("http_listen_error", map[string]any{"port": port, "path": path, "ssl": ssl, "error": err.Error()})
	}
}
