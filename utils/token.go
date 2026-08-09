package utils

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

const secretKey = "gin-project-secret-key-use-from-env"

func GenerateToken(id uint) (string, error) {
	claims := jwt.MapClaims{
		"id": id,
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secretKey))
}

func ParseToken(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if id, ok := claims["id"].(float64); ok {
			return uint(id), nil
		}
	}
	return 0, errors.New("invalid token")
}
