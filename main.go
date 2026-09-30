package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/snahmik/golang-http-server/internal/database"
	"github.com/snahmik/golang-http-server/internal/handlers"
	"github.com/snahmik/golang-http-server/internal/middleware"
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
	mux.Handle("/app/", middleware.IncrementHit(config, http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.Handle("GET /api/healthz", config.HealthHandler())
	mux.Handle("POST /api/validate_chirp", config.ValidateChirpHandler())
	mux.Handle("POST /api/users", config.CreateUserHandler())
	mux.Handle("POST /api/chirps", config.CreateChirpHandler())
	mux.Handle("GET /api/chirps", config.FetchChirpsHandler())
	mux.Handle("GET /admin/metrics", config.MetricsHandler())
	mux.Handle("POST /admin/reset", config.ResetHandler())
	mux.Handle("GET /api/chirps/{id}", config.FetchChirpHandler())
	mux.HandleFunc("POST /api/login", config.LoginUserHandler)
	mux.HandleFunc("POST /api/refresh", config.RefreshHandler)
	mux.HandleFunc("POST /api/revoke", config.RevokeHandler)

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
