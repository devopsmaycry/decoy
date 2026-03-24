package listeners

import (
	"decoy/logger"
	"io"
	"net/http"
)

func StartHTTP(port string, log *logger.Logger) {
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
		}
		if len(body) == 4096 {
			fields["body_truncated"] = true
		}

		log.Log("http_request", fields)
		w.WriteHeader(http.StatusOK)
	})

	log.Log("http_listening", map[string]any{"port": port})
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Log("http_listen_error", map[string]any{"port": port, "error": err.Error()})
	}
}
