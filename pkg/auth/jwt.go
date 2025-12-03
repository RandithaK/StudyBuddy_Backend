package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var jwtSecret = []byte("dev-secret") // In production, load from env

type Claims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func GenerateRefreshToken(userID string) (string, error) {
	expirationTime := time.Now().Add(30 * 24 * time.Hour) // 30 days
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

type contextKey string

const UserIDKey contextKey = "userID"

func ForContext(ctx context.Context) string {
	raw, _ := ctx.Value(UserIDKey).(string)
	return raw
}
