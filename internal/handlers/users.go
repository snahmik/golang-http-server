package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/snahmik/golang-http-server/internal/auth"
	"github.com/snahmik/golang-http-server/internal/database"
	"github.com/snahmik/golang-http-server/internal/response"
)

type userReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type user struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
}

func (config *ApiConfig) CreateUserHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		reqData := userReq{}
		err := decoder.Decode(&reqData)
		if err != nil {
			response.BadRequestError(w, "", err)
			return
		}

		hashedUserPassword, err := auth.HashPassword(reqData.Password)
		if err != nil {
			response.InternalServerError(w, err)
			return
		}

		params := database.CreateUserParams{
			ID:             uuid.New(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			Email:          reqData.Email,
			HashedPassword: hashedUserPassword,
		}

		createdUser, err := config.db.CreateUser(r.Context(), params)
		if err != nil {
			response.InternalServerError(w, err)
			return
		}

		resData := user{
			ID:        createdUser.ID,
			CreatedAt: createdUser.CreatedAt,
			UpdatedAt: createdUser.UpdatedAt,
			Email:     createdUser.Email,
		}

		response.JSON(w, http.StatusCreated, resData)
	})
}

func (config *ApiConfig) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type res struct {
		user
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	reqData := req{}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		response.BadRequestError(w, "", err)
		return
	}

	fetchedUser, err := config.db.FetchUser(r.Context(), reqData.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.UnauthorizedError(w, err)
			return
		}

		response.InternalServerError(w, fmt.Errorf("login user database error: %v", err))
		return
	}

	isValid, err := auth.CheckPassword(reqData.Password, fetchedUser.HashedPassword)
	if err != nil {
		response.InternalServerError(w, err)
		return
	}

	if !isValid {
		response.UnauthorizedError(w, errors.New("invalid email or password"))
		return
	}

	//access token expiry is one hour
	accessToken, err := auth.MakeJWT(fetchedUser.ID, config.jwtSecret)
	if err != nil {
		response.InternalServerError(w, err)
		return
	}

	refreshToken := auth.MakeRefreshToken()

	params := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    fetchedUser.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
	}
	refreshTokenEntry, err := config.db.CreateRefreshToken(r.Context(), params)
	if err != nil {
		response.InternalServerError(w, err)
		return
	}

	resData := res{
		user: user{
			ID:        fetchedUser.ID,
			CreatedAt: fetchedUser.CreatedAt,
			UpdatedAt: fetchedUser.UpdatedAt,
			Email:     fetchedUser.Email,
		},
		Token:        accessToken,
		RefreshToken: refreshTokenEntry.Token,
	}

	response.JSON(w, http.StatusOK, resData)
}
