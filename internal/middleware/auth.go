package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

type authContextKey struct{}

// RequireAuth validates JWT access tokens and injects the user ID into context.
func RequireAuth(secret []byte, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := strings.TrimSpace(r.Header.Get("Authorization"))
		if authorization == "" || !strings.HasPrefix(strings.ToLower(authorization), "bearer ") {
			writeAuthError(w, http.StatusUnauthorized, "Missing bearer token")
			return
		}

		tokenString := strings.TrimSpace(authorization[len("Bearer "):])
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil || !token.Valid {
			writeAuthError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		subject, ok := claims["sub"].(string)
		if !ok || subject == "" {
			writeAuthError(w, http.StatusUnauthorized, "Invalid token subject")
			return
		}

		userID, err := strconv.ParseInt(subject, 10, 64)
		if err != nil || userID <= 0 {
			writeAuthError(w, http.StatusUnauthorized, "Invalid token subject")
			return
		}

		ctx := context.WithValue(r.Context(), authContextKey{}, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserIDFromContext returns the authenticated user ID if present.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	value, ok := ctx.Value(authContextKey{}).(int64)
	return value, ok
}

func writeAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(models.APIError{Error: message})
}
