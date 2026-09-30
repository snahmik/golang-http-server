package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/snahmik/golang-http-server/internal/auth"
	"github.com/snahmik/golang-http-server/internal/database"
	"github.com/snahmik/golang-http-server/internal/response"
)

type chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func (config *ApiConfig) CreateChirpHandler() http.Handler {
	type req struct {
		Body string `json:"body"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		reqData := req{}
		err := decoder.Decode(&reqData)
		if err != nil {
			response.BadRequestError(w, "", err)
			return
		}

		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			response.UnauthorizedError(w, err)
			return
		}

		userID, err := auth.ValidateJWT(token, config.jwtSecret)
		if err != nil {
			response.UnauthorizedError(w, err)
			return
		}

		params := database.CreateChirpParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Body:      reqData.Body,
			UserID:    userID,
		}

		createdChirp, err := config.db.CreateChirp(r.Context(), params)
		if err != nil {
			response.InternalServerError(w, err)
			return
		}

		resData := chirp{
			ID:        createdChirp.ID,
			CreatedAt: createdChirp.CreatedAt,
			UpdatedAt: createdChirp.UpdatedAt,
			Body:      createdChirp.Body,
			UserId:    createdChirp.UserID,
		}

		response.JSON(w, http.StatusCreated, resData)
	})
}

func (config *ApiConfig) FetchChirpsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetchedChirps, err := config.db.FetchAllChirps(r.Context())
		if err != nil {
			response.InternalServerError(w, err)
			return
		}

		resData := make([]chirp, 0, len(fetchedChirps))
		for _, fetchedChirp := range fetchedChirps {
			resData = append(resData, chirp{
				ID:        fetchedChirp.ID,
				CreatedAt: fetchedChirp.CreatedAt,
				UpdatedAt: fetchedChirp.UpdatedAt,
				Body:      fetchedChirp.Body,
				UserId:    fetchedChirp.UserID,
			})
		}

		response.JSON(w, http.StatusOK, resData)
	})
}

func (config *ApiConfig) FetchChirpHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chirpId := r.PathValue("id")
		parsedId, err := uuid.Parse(chirpId)
		if err != nil {
			response.BadRequestError(w, "Malformed chirp ID", err)
			return
		}

		fetchedChirp, err := config.db.FetchChirp(r.Context(), parsedId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				response.GenericError(w, http.StatusNotFound, "Chirp Not Found", err)
				return
			}

			response.InternalServerError(w, err)
			return
		}

		resData := chirp{
			ID:        fetchedChirp.ID,
			CreatedAt: fetchedChirp.CreatedAt,
			UpdatedAt: fetchedChirp.UpdatedAt,
			Body:      fetchedChirp.Body,
			UserId:    fetchedChirp.UserID,
		}

		response.JSON(w, http.StatusOK, resData)
	})
}

type validateChirpReq struct {
	Body string `json:"body"`
}

type validateChirpRes struct {
	Error       string `json:"error"`
	Valid       bool   `json:"valid"`
	CleanedBody string `json:"cleaned_body"`
}

func (config *ApiConfig) ValidateChirpHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqData := validateChirpReq{}
		if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
			response.BadRequestError(w, "", err)
			return
		}

		if len(reqData.Body) > 140 {
			response.BadRequestError(w, "Chirp length exceeds 140", nil)
		}

		resData := validateChirpRes{}
		resData.Valid = true
		resData.CleanedBody = cleanInput(reqData.Body)

		response.JSON(w, http.StatusOK, resData)
	})
}

func cleanInput(input string) (cleanedString string) {
	profanities := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	words := strings.Split(input, " ")
	for index, word := range words {
		if _, ok := profanities[strings.ToLower(word)]; ok {
			words[index] = "****"
		}
	}

	return strings.Join(words, " ")
}
