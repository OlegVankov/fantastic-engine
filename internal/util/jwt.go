package util

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type UserClaim struct {
	jwt.RegisteredClaims
	Username string
	UserID   uint64
}

func CreateToken(username string, userid uint64) (string, error) {
	userClaim := &UserClaim{
		Username: username,
		UserID:   userid,
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
