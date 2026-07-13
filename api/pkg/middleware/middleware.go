package middleware

import (
	"context"
	"net/http"
	"unicode/utf8"

	"alexdenkk/labs/pkg/token/jwt"
)

type Middleware struct {
	tokenManager *jwt.TokenManager
}

func NewMiddleware(tokenManager *jwt.TokenManager) *Middleware {
	return &Middleware{
		tokenManager: tokenManager,
	}
}

func (middleware *Middleware) Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")

		if utf8.RuneCountInString(tokenString) < 8 {
			http.Error(w, "not authorized", http.StatusForbidden)
			return
		}

		claims, err := middleware.tokenManager.ParseAccessToken(tokenString[7:])

		if err != nil {
			http.Error(w, "not authorized", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
