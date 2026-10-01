package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/snahmik/golang-http-server/internal/auth"
	"github.com/snahmik/golang-http-server/internal/database"
	"github.com/snahmik/golang-http-server/internal/response"
)

func MiddlewareIncrementHit(config *ApiConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config.serverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (config *ApiConfig) MiddlewareAuthJWT(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			response.UnauthorizedError(w, fmt.Errorf("getting jwt token from header in auth jwt middleware: %w", err))
			return
		}

		userID, err := auth.ValidateJWT(token, config.jwtSecret)
		if err != nil {
			response.UnauthorizedError(w, fmt.Errorf("validating jwt token in auth jwt middleware: %w", err))
			return
		}

		userSessionContext := auth.SessionContext{
			Token:  token,
			UserID: userID,
		}
		context := auth.CreateSessionContext(r.Context(), userSessionContext)

		next.ServeHTTP(w, r.WithContext(context))
	})
}

func (config *ApiConfig) MiddlewareAuthRefreshToken(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			response.UnauthorizedError(w, fmt.Errorf("getting jwt token from header in auth refresh token middleware: %w", err))
			return
		}

		params := database.FetchRefreshTokenParams{
			Token:     token,
			ExpiresAt: time.Now(),
		}
		refreshTokenRow, err := config.db.FetchRefreshToken(r.Context(), params)
		if err != nil {
			response.UnauthorizedError(w, fmt.Errorf("fetching refresh token %s from db: %v", token, err))
			return
		}

		userSessionContext := auth.SessionContext{
			Token:  token,
			UserID: refreshTokenRow.UserID,
		}
		context := auth.CreateSessionContext(r.Context(), userSessionContext)

		next.ServeHTTP(w, r.WithContext(context))
	})
}
