package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/snahmik/golang-http-server/internal/auth"
	"github.com/snahmik/golang-http-server/internal/database"
	"github.com/snahmik/golang-http-server/internal/response"
)

func (config *ApiConfig) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	type res struct {
		Token string `json:"token"`
	}

	context, err := auth.GetSessionContext(r)
	if err != nil {
		response.UnauthorizedError(w, fmt.Errorf("getting session context for refresh request: %w", err))
		return
	}

	params := database.FetchRefreshTokenParams{
		Token:     context.Token,
		ExpiresAt: time.Now(),
	}

	refreshTokenRow, err := config.db.FetchRefreshToken(r.Context(), params)
	if err != nil {
		errMsg := fmt.Errorf("fetching refresh token %s from db: %w", context.Token, err)
		if errors.Is(err, sql.ErrNoRows) {
			response.UnauthorizedError(w, errMsg)
			return
		}

		response.InternalServerError(w, errMsg)
		return
	}

	jwtToken, err := auth.MakeJWT(refreshTokenRow.UserID, config.jwtSecret)
	if err != nil {
		response.InternalServerError(w, fmt.Errorf("making jwt for refresh request: %w", err))
		return
	}

	resData := res{
		Token: jwtToken,
	}

	response.JSON(w, http.StatusOK, resData)
}

func (config *ApiConfig) RevokeHandler(w http.ResponseWriter, r *http.Request) {
	ctx, err := auth.GetSessionContext(r)
	if err != nil {
		response.UnauthorizedError(w, fmt.Errorf("getting session context for revoke request: %w", err))
		return
	}

	params := database.RevokeRefreshTokenParams{
		Token: ctx.Token,
		RevokedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	}

	affectedRows, err := config.db.RevokeRefreshToken(r.Context(), params)
	if err != nil {
		response.InternalServerError(w, fmt.Errorf("revoking refresh token %s for revoke request: %w", ctx.Token, err))
		return
	}

	if affectedRows == 0 {
		response.UnauthorizedError(w, fmt.Errorf("revoke token %s invalid for revoke request", ctx.Token))
		return
	}

	response.JSON(w, http.StatusNoContent, struct {
	}{})
}
