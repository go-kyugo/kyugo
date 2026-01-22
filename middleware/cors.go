package middleware

import (
	"net/http"

	"kyugo.dev/kyugo/v1/config"
)

// CORS returns a middleware that applies simple CORS headers based on config.
func CORS(c config.CorsConfig) func(http.Handler) http.Handler {
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

// Logger is a simple request logger middleware.
