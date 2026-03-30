package listeners

import (
	"decoy/config"
	"decoy/logger"
	"io"
	"net/http"
	"time"
)

func StartHTTP(port string, log *logger.Logger, opts config.HttpsConfig, ssl bool, path string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

		log.Log("http_request", fields)
		w.WriteHeader(http.StatusOK)
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
		log.Log("http_listen_error", map[string]any{"port": port, "ssl": ssl, "error": err.Error()})
	}
}
