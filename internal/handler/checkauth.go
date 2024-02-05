package handler

import (
	"net/http"
	"strings"

	"github.com/OlegVankov/fantastic-engine/internal/util/jwt"
)

func Auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		username := jwt.GetUser(token)
		if username == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		r.Header.Add("username", username)
		h.ServeHTTP(w, r)
	})
}
