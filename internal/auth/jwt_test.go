package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/snahmik/golang-http-server/internal/auth"
)

func TestJWT(t *testing.T) {
	tests := []struct {
		name                  string
		userID                uuid.UUID
		creationTokenSecret   string
		validationTokenSecret string
		expiresIn             time.Duration
		wantError             bool
	}{
		{name: "Happy path", userID: uuid.New(), creationTokenSecret: "secret", validationTokenSecret: "secret", expiresIn: time.Minute, wantError: false},
		{name: "Expired token", userID: uuid.New(), creationTokenSecret: "secret", validationTokenSecret: "secret", expiresIn: time.Minute * 0, wantError: true},
		{name: "Different token secret", userID: uuid.New(), creationTokenSecret: "secret", validationTokenSecret: "different-secret", expiresIn: time.Minute, wantError: true},
	}

	for _, tt := range tests {
		tokenString, err := auth.MakeJWT(tt.userID, tt.creationTokenSecret, tt.expiresIn)

		if err != nil {
			t.Errorf("Test :%s\nError: %v\nWant: no error during token creation", tt.name, err)
		}

		userID, err := auth.ValidateJWT(tokenString, tt.validationTokenSecret)

		if (err != nil) != tt.wantError {
			t.Errorf("Test: %s\nError: %v\nWant: %v", tt.name, err, tt.wantError)
			continue
		}

		if err != nil {
			continue
		}

		if userID != tt.userID {
			t.Errorf("Test: %s\nGot userID: %v\nWant userID: %v", tt.name, userID, tt.userID)
		}

	}
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name          string
		authHeader    string
		wantSetHeader bool
		wantError     bool
	}{
		{name: "Happy path", authHeader: "Bearer testToken", wantSetHeader: true, wantError: false},
		{name: "Malformed header (Missing Token)", authHeader: "Bearer ", wantSetHeader: true, wantError: true},
		{name: "Malformed header (Missing Bearer)", authHeader: " testToken", wantSetHeader: true, wantError: true},
		{name: "Missing header", authHeader: "Bearer testToken", wantSetHeader: false, wantError: true},
	}

	checkedToken := "testToken"

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		if tt.wantSetHeader {
			req.Header.Set("Authorization", tt.authHeader)
		}

		token, err := auth.GetBearerToken(req.Header)

		if (err != nil) != tt.wantError {
			t.Errorf("Test: %s\nError: %v\nWant: %v", tt.name, err, tt.wantError)
			continue
		}

		if err != nil {
			continue
		}

		if token != checkedToken {
			t.Errorf("Test: %s\nGot token: %v\nWant token: %v", tt.name, token, checkedToken)
		}
	}
}
