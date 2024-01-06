package util

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type UserClaim struct {
	jwt.RegisteredClaims
	Username string
}

func CreateToken(username string) (string, error) {
	userClaim := &UserClaim{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaim)

	tokenString, err := token.SignedString([]byte("secret_key"))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
