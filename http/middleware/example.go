package middleware

import (
	"net/http"

	"kyugo.dev/kyugo/v1/logger"
)

// Example is a simple middleware that logs and sets a response header so
// you can see it applied both at group and route level in the example app.
func Example(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Example.Middleware called", nil)
		w.Header().Set("X-Example-Middleware", "true")
		next.ServeHTTP(w, r)
	})
}
