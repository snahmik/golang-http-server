package handlers

import (
	"fmt"
	"log"
	"net/http"
)

func (config *ApiConfig) MetricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		response := fmt.Sprintf("<html>\n  <body>\n    <h1>Welcome, Chirpy Admin</h1>\n    <p>Chirpy has been visited %d times!</p>\n  </body>\n</html>", config.serverHits.Load())
		_, err := w.Write([]byte(response))
		if err != nil {
			fmt.Printf("error writing response in /healthz: %v", err)
		}
	})
}

func (config *ApiConfig) ResetHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.platform != "dev" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		err := config.db.DeleteUsers(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Error deleting users: %v", err)
			return
		}

		config.serverHits.Swap(0)
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("Database Reset"))
		if err != nil {
			fmt.Printf("error writing response in /reset")
			return
		}
	})
}
func (config *ApiConfig) HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			fmt.Printf("error writing server response in /healthz: %v", err)
		}
	})
}
