package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		Subject:   userID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	})
	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return tokenString, err
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(tokenSecret), nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	userID, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid token subject: %v", err)
	}

	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error parsing userId: %v", err)
	}

	return parsedID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorisation header is missing")
	}

	bearerToken := strings.Fields(authHeader)

	if len(bearerToken) != 2 || !strings.EqualFold(bearerToken[0], "Bearer") {
		return "", fmt.Errorf("malformed authorisation header")
	}

	token := strings.TrimSpace(bearerToken[1])

	if token == "" {
		return "", fmt.Errorf("malformed authorisation header")
	}

	return token, nil
}

func MakeRefreshToken() string {
	key := make([]byte, 32)

	_, err := rand.Read(key)
	//panic used as method fail is OS level issue
	if err != nil {
		panic("Failed to generate refresh token")
	}

	return hex.EncodeToString(key)
}
