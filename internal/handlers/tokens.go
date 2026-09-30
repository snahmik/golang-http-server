package handlers

import (
	"database/sql"
	"errors"
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

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		response.UnauthorizedError(w, err)
		return
	}

	params := database.FetchRefreshTokenParams{
		Token:     token,
		ExpiresAt: time.Now(),
	}

	refreshTokenRow, err := config.db.FetchRefreshToken(r.Context(), params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.UnauthorizedError(w, errors.New("refresh token not found in db"))
			return
		}

		response.InternalServerError(w, err)
		return
	}

	jwtToken, err := auth.MakeJWT(refreshTokenRow.UserID, config.jwtSecret)
	if err != nil {
		response.InternalServerError(w, err)
		return
	}

	resData := res{
		Token: jwtToken,
	}

	response.JSON(w, http.StatusOK, resData)
}

func (config *ApiConfig) RevokeHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		response.UnauthorizedError(w, err)
		return
	}

	params := database.RevokeRefreshTokenParams{
		Token: token,
		RevokedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	}

	affectedRows, err := config.db.RevokeRefreshToken(r.Context(), params)
	if err != nil {
		response.InternalServerError(w, err)
		return
	}

	if affectedRows == 0 {
		response.UnauthorizedError(w, errors.New("invalid refresh token"))
		return
	}

	response.JSON(w, http.StatusNoContent, struct {
	}{})
}
