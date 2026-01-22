package kyugo

import (
	"fmt"
	"net/http"
	"time"

	cfg "kyugo.dev/kyugo/v1/config"
	logger "kyugo.dev/kyugo/v1/logger"
)

// CORS returns a middleware that applies simple CORS headers based on config.
func CORS(c cfg.CorsConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(c.AllowedOrigins) > 0 {
				w.Header().Set("Access-Control-Allow-Origin", c.AllowedOrigins[0])
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
			if len(c.AllowedMethods) > 0 {
				// join not required for minimal implementation
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// responseRecorder captures status and size written by the handler.
type responseRecorder struct {
	http.ResponseWriter
	status int
	size   int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.size += n
	return n, err
}

// Logger logs each HTTP request in a single, console-friendly line.
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rr := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(rr, r)
		dur := time.Since(start)

		// Build a single-line, console-friendly message with sorted fields
		// Order: duration_ms, method, path, remote_addr, size, status, status_text
		statusText := http.StatusText(rr.status)
		// colorize keys for easier scanning
		dKey := logger.Colorize("duration_ms", "36")
		mKey := logger.Colorize("method", "36")
		pKey := logger.Colorize("path", "36")
		raKey := logger.Colorize("remote_addr", "36")
		sizeKey := logger.Colorize("size", "36")
		statusKey := logger.Colorize("status", "36")
		stKey := logger.Colorize("status_text", "36")

		msg := fmt.Sprintf("HTTP.Request %s=%d %s=%s %s=%s %s=%s %s=%d %s=%d %s=\"%s\"",
			dKey, dur.Milliseconds(),
			mKey, r.Method,
			pKey, r.URL.Path,
			raKey, r.RemoteAddr,
			sizeKey, rr.size,
			statusKey, rr.status,
			stKey, statusText,
		)

		logger.Info(msg, nil)
	})
}
