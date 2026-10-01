package handlers

import (
	"fmt"
	"log"
	"net/http"
)

func (config *ApiConfig) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	response := fmt.Sprintf("<html>\n  <body>\n    <h1>Welcome, Chirpy Admin</h1>\n    <p>Chirpy has been visited %d times!</p>\n  </body>\n</html>", config.serverHits.Load())
	_, err := w.Write([]byte(response))
	if err != nil {
		fmt.Printf("writing response in /metrics: %v", err)
	}
}

func (config *ApiConfig) ResetHandler(w http.ResponseWriter, r *http.Request) {
	if config.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	err := config.db.DeleteUsers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("resetting server: %v", err)
		return
	}

	config.serverHits.Swap(0)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Database Reset"))
	if err != nil {
		fmt.Printf("writing response in /reset")
		return
	}
}

func (config *ApiConfig) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		fmt.Printf("writing response in /healthz: %v", err)
	}
}
