package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"

	"github.com/OlegVankov/fantastic-engine/internal/util"
)

func Auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

		if auth == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userClaim := util.UserClaim{}

		token, err := jwt.ParseWithClaims(auth, &userClaim, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret_key"), nil
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !token.Valid {
			fmt.Println("token not valid", userClaim.UserID, userClaim.Username)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		r.Header.Add("username", userClaim.Username)
		h.ServeHTTP(w, r)
	})
}
