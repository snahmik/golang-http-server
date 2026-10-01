package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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

func (config *ApiConfig) CreateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Body string `json:"body"`
	}

	ctx, err := auth.GetSessionContext(r)
	if err != nil {
		response.UnauthorizedError(w, fmt.Errorf("getting session context for create chirp request: %w", err))
	}

	reqData := req{}
	err = json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		response.BadRequestError(w, "", fmt.Errorf("decoding create chirp request: %w", err))
		return
	}

	params := database.CreateChirpParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      reqData.Body,
		UserID:    ctx.UserID,
	}

	createdChirp, err := config.db.CreateChirp(r.Context(), params)
	if err != nil {
		response.InternalServerError(w, fmt.Errorf("creating chirp %v in db: %w", params.ID, err))
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
}

func (config *ApiConfig) FetchChirpsHandler(w http.ResponseWriter, r *http.Request) {
	fetchedChirps, err := config.db.FetchAllChirps(r.Context())
	if err != nil {
		response.InternalServerError(w, fmt.Errorf("fetching chirps from db: %w", err))
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
}

func (config *ApiConfig) FetchChirpHandler(w http.ResponseWriter, r *http.Request) {
	chirpId := r.PathValue("id")
	parsedId, err := uuid.Parse(chirpId)
	if err != nil {
		response.BadRequestError(w, "Malformed chirp ID", fmt.Errorf("parsing chirp id for fetch chirp request"))
		return
	}

	fetchedChirp, err := config.db.FetchChirp(r.Context(), parsedId)
	if err != nil {
		errMsg := fmt.Errorf("fetching chirp %v from db: %w", parsedId, err)
		if errors.Is(err, sql.ErrNoRows) {
			response.GenericError(w, http.StatusNotFound, "Chirp Not Found", errMsg)
			return
		}

		response.InternalServerError(w, errMsg)
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
}

func (config *ApiConfig) ValidateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Body string `json:"body"`
	}

	type res struct {
		Valid       bool   `json:"valid"`
		CleanedBody string `json:"cleaned_body"`
	}

	reqData := req{}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		response.BadRequestError(w, "", fmt.Errorf("decoding validate chirp request: %w", err))
		return
	}

	if len(reqData.Body) > 140 {
		response.BadRequestError(w, "Chirp length exceeds 140 word length", fmt.Errorf("validating chirp: chirp request exceeds 140 word length"))
		return
	}

	resData := res{
		Valid:       true,
		CleanedBody: cleanInput(reqData.Body),
	}

	response.JSON(w, http.StatusOK, resData)
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
