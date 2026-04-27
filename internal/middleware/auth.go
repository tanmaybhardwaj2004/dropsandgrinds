package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

type authContextKey struct{}

var logger *slog.Logger

// SetAuthLogger sets the logger for auth middleware
func SetAuthLogger(l *slog.Logger) {
	logger = l
}

// RequireAuth validates JWT access tokens and injects the user ID into context.
func RequireAuth(secret []byte, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := strings.TrimSpace(r.Header.Get("Authorization"))
		if authorization == "" {
			writeAuthError(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		if !strings.HasPrefix(strings.ToLower(authorization), "bearer ") {
			writeAuthError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		tokenString := strings.TrimSpace(authorization[len("Bearer "):])
		if tokenString == "" {
			writeAuthError(w, http.StatusUnauthorized, "Empty bearer token")
			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return secret, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

		if err != nil {
			// Check if token is expired
			if errors.Is(err, jwt.ErrTokenExpired) {
				writeAuthError(w, http.StatusUnauthorized, "Token expired")
				if logger != nil {
					logger.Warn("Token expired", "ip", r.RemoteAddr)
				}
				return
			}
			// Check if token is malformed
			if errors.Is(err, jwt.ErrTokenMalformed) {
				writeAuthError(w, http.StatusUnauthorized, "Malformed token")
				if logger != nil {
					logger.Warn("Malformed token", "ip", r.RemoteAddr)
				}
				return
			}
			// Check if token signature is invalid
			if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
				writeAuthError(w, http.StatusUnauthorized, "Invalid token signature")
				if logger != nil {
					logger.Warn("Invalid token signature", "ip", r.RemoteAddr)
				}
				return
			}
			// Generic error
			writeAuthError(w, http.StatusUnauthorized, "Invalid token")
			if logger != nil {
				logger.Warn("Invalid token", "error", err, "ip", r.RemoteAddr)
			}
			return
		}

		if !token.Valid {
			writeAuthError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Check token expiration explicitly
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				writeAuthError(w, http.StatusUnauthorized, "Token expired")
				if logger != nil {
					logger.Warn("Token expired (exp check)", "ip", r.RemoteAddr)
				}
				return
			}
		}

		// Check token not before (nbf) if present
		if nbf, ok := claims["nbf"].(float64); ok {
			if time.Now().Unix() < int64(nbf) {
				writeAuthError(w, http.StatusUnauthorized, "Token not yet valid")
				if logger != nil {
					logger.Warn("Token not yet valid (nbf check)", "ip", r.RemoteAddr)
				}
				return
			}
		}

		subject, ok := claims["sub"].(string)
		if !ok || subject == "" {
			writeAuthError(w, http.StatusUnauthorized, "Invalid token subject")
			if logger != nil {
				logger.Warn("Invalid token subject", "ip", r.RemoteAddr)
			}
			return
		}

		userID, err := strconv.ParseInt(subject, 10, 64)
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, "Invalid token subject format")
			if logger != nil {
				logger.Warn("Invalid token subject format", "subject", subject, "ip", r.RemoteAddr)
			}
			return
		}

		if userID <= 0 {
			writeAuthError(w, http.StatusUnauthorized, "Invalid user ID in token")
			if logger != nil {
				logger.Warn("Invalid user ID in token", "user_id", userID, "ip", r.RemoteAddr)
			}
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
