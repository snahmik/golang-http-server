package auth_test

import (
	"testing"

	"github.com/snahmik/golang-http-server/internal/auth"
)

func TestPassword(t *testing.T) {
	tests := []struct {
		name                 string
		creationPassword     string
		verificationPassword string
		wantIsValid          bool
	}{
		{name: "Happy path", creationPassword: "test123", verificationPassword: "test123", wantIsValid: true},
		{name: "Different password", creationPassword: "test123", verificationPassword: "test321", wantIsValid: false},
	}

	for _, tt := range tests {
		hashedPassword, err := auth.HashPassword(tt.creationPassword)

		if err != nil {
			t.Errorf("Test :%s\nError: %v\nWant: No error", tt.name, err)
		}

		isValid, err := auth.CheckPassword(tt.verificationPassword, hashedPassword)

		if isValid != tt.wantIsValid {
			t.Errorf("Test :%s\nGot password is valid: %v\nWant password is valid: %v", tt.name, isValid, tt.wantIsValid)
		}

		if err != nil {
			t.Errorf("Test :%s\nError: %v\nWant: No error", tt.name, err)
		}
	}
}
