package handlers

import (
	"sync/atomic"

	"github.com/snahmik/golang-http-server/internal/database"
)

type ApiConfig struct {
	platform   string
	jwtSecret  string
	db         *database.Queries
	serverHits atomic.Int32
}

func NewApiConfig(db *database.Queries, platform string, jwtSecret string) *ApiConfig {
	return &ApiConfig{
		db:        db,
		jwtSecret: jwtSecret,
		platform:  platform,
	}
}

func (config *ApiConfig) IncrementServerHit() {
	config.serverHits.Add(1)
}
