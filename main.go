package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/snahmik/golang-http-server/internal/database"
	"github.com/snahmik/golang-http-server/internal/handlers"
)

import _ "github.com/lib/pq"

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading ENV")
	}

	dbURL := os.Getenv("DB_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic("Error connecting to db")
	}

	config := handlers.NewApiConfig(database.New(db), os.Getenv("PLATFORM"), jwtSecret)

	mux := http.NewServeMux()
	mux.Handle("/app/", handlers.MiddlewareIncrementHit(config, http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", config.HealthHandler)
	mux.HandleFunc("GET /admin/metrics", config.MetricsHandler)
	mux.HandleFunc("POST /admin/reset", config.ResetHandler)

	mux.Handle("POST /api/chirps", config.MiddlewareAuthJWT(config.CreateChirpHandler))
	mux.HandleFunc("POST /api/validate_chirp", config.ValidateChirpHandler)
	mux.HandleFunc("GET /api/chirps", config.FetchChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{id}", config.FetchChirpHandler)

	mux.Handle("POST /api/refresh", config.MiddlewareAuthRefreshToken(config.RefreshHandler))
	mux.Handle("POST /api/revoke", config.MiddlewareAuthRefreshToken(config.RevokeHandler))
	mux.HandleFunc("POST /api/login", config.LoginUserHandler)

	mux.HandleFunc("POST /api/users", config.CreateUserHandler)
	mux.Handle("PUT /api/users", config.MiddlewareAuthRefreshToken(config.UpdateUserHandler))

	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	fmt.Println("Server listening on port 8080")
	err = server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}
