package token

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func Create(userID uint, role, secret string, hours int) (string, error) {
	now := time.Now()
	claims := Claims{UserID: userID, Role: role, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(hours) * time.Hour)), IssuedAt: jwt.NewNumericDate(now)}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func Parse(value, secret string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(value, claims, func(parsed *jwt.Token) (any, error) { return []byte(secret), nil }, jwt.WithValidMethods([]string{"HS256"}))
	return claims, err
}
