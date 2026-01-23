package middleware

import (
	"net/http"

	logger "github.com/go-kyugo/kyugo/logger"
)

// Example is a sample middleware used by the example controller chain.
func Example(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("example middleware", nil)
		next.ServeHTTP(w, r)
	})
}
