package middleware

import (
	"net/http"

	"github.com/snahmik/golang-http-server/internal/handlers"
)

func IncrementHit(config *handlers.ApiConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config.IncrementServerHit()
		next.ServeHTTP(w, r)
	})
}
