package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"
	"github.com/snahmik/golang-http-server/internal/auth"
	"github.com/snahmik/golang-http-server/internal/database"
	"github.com/snahmik/golang-http-server/internal/response"
)

type user struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
}

func (config *ApiConfig) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	reqData := req{}
	err := decoder.Decode(&reqData)
	if err != nil {
		response.BadRequestError(w, "", fmt.Errorf("decoding create user request: %w", err))
		return
	}

	hashedUserPassword, err := auth.HashPassword(reqData.Password)
	if err != nil {
		response.InternalServerError(w, fmt.Errorf("hashing password for create user request: %w", err))
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
		errMsg := fmt.Errorf("creating user with id %v in db: %w", params.ID, err)

		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case pqerror.UniqueViolation:
				response.GenericError(w, http.StatusConflict, "Email already exists", errMsg)
				return
			}
		}

		response.InternalServerError(w, errMsg)
		return
	}

	resData := user{
		ID:        createdUser.ID,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
		Email:     createdUser.Email,
	}

	response.JSON(w, http.StatusCreated, resData)
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
		response.BadRequestError(w, "", fmt.Errorf("decoding login request: %w", err))
		return
	}

	fetchedUser, err := config.db.FetchUser(r.Context(), reqData.Email)
	if err != nil {
		errMsg := fmt.Errorf("fetching user %v from db: %w", err)
		if errors.Is(err, sql.ErrNoRows) {
			response.UnauthorizedError(w, errMsg)
			return
		}

		response.InternalServerError(w, errMsg)
		return
	}

	isValid, err := auth.CheckPassword(reqData.Password, fetchedUser.HashedPassword)
	if err != nil {
		response.InternalServerError(w, fmt.Errorf("checking password from login request: %w", err))
		return
	}

	if !isValid {
		response.UnauthorizedError(w, errors.New("checking password from login request: invalid email or password"))
		return
	}

	accessToken, err := auth.MakeJWT(fetchedUser.ID, config.jwtSecret)
	if err != nil {
		response.InternalServerError(w, fmt.Errorf("creating jwt for login request: %w", err))
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
		response.InternalServerError(w, fmt.Errorf("creating refresh token for user %v in db: %w", fetchedUser.ID, err))
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

//func (config *ApiConfig) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
//	type req struct {
//		Email    string `json:"email"`
//		Password string `json:"password"`
//	}
//
//	ctx, err := auth.GetSessionContext(r)
//	if err != nil {
//		response.UnauthorizedError(w,fmt.Errorf("getting session context for update user request: %w",err))
//		return
//	}
//
//	params := database.UpdateUserParams{
//		Email:          "",
//		HashedPassword: "",
//		ID:             ctx.UserID,
//	}
//	affectedRows, err := config.db.UpdateUser()
//}
