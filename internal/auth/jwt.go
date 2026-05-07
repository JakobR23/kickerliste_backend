package auth

import (
	"errors"
	"fmt"
	"time"

	"bierliste_backend/internal/entity"

	"github.com/golang-jwt/jwt/v5"
)

const tokenTTL = 24 * time.Hour

// Claims are the fields embedded inside each JWT.
type Claims struct {
	UserId   int         `json:"userId"`
	Username string      `json:"username"`
	Role     entity.Role `json:"role"`
	jwt.RegisteredClaims
}

// generateToken signs a new JWT for the given user.
func generateToken(u entity.User, secret string) (string, error) {
	claims := Claims{
		UserId:   u.Id,
		Username: u.Username,
		Role:     u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// parseToken validates tokenStr and returns the embedded claims.
func parseToken(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
