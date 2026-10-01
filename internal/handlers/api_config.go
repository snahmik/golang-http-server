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

//func (config *ApiConfig) IncrementServerHit() {
//	config.serverHits.Add(1)
//}
//
//func (config *ApiConfig) GetJWTSecret() (string, error) {
//	if config.jwtSecret == "" {
//		return "", errors.New("session token unavailable")
//	}
//
//	return config.jwtSecret, nil
//}
//
//func (config *ApiConfig) GetSessionToken() (string, error) {
//	if config.sessionToken == "" {
//		return "", errors.New("session token unavailable")
//	}
//
//	return config.sessionToken, nil
//}
//
//func (config *ApiConfig) SetSessionToken(token string) {
//	config.sessionToken = token
//}
//
//func (config *ApiConfig) GetSessionID() (uuid.UUID, error) {
//	if config.sessionToken == "" {
//		return uuid.Nil, errors.New("session token unavailable")
//	}
//
//	return config.sessionID, nil
//}
//
//func (config *ApiConfig) SetSessionID(userID uuid.UUID) {
//	config.sessionID = userID
//}
