package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type UserClaim struct {
	jwt.RegisteredClaims
	Username string
	UserID   uint64
}

const secretKey = "AsDfGhJkL"

func CreateToken(username string, userid uint64) (string, error) {
	userClaim := &UserClaim{
		Username: username,
		UserID:   userid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaim)

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetUser(tokenString string) string {
	userClaim := &UserClaim{}
	token, err := jwt.ParseWithClaims(tokenString, userClaim, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return ""
	}
	if !token.Valid {
		return ""
	}
	return userClaim.Username
}
