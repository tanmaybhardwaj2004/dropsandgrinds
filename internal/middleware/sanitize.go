package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

var dangerousInputPattern = regexp.MustCompile(`(?i)(<\s*script|javascript:|on\w+\s*=|union\s+select|drop\s+table|insert\s+into|delete\s+from|--|/\*)`)

func SanitizeInput(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := strings.ToLower(r.Header.Get("Content-Type"))
		if r.Body == nil || !strings.Contains(contentType, "application/json") {
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeSanitizeError(w, "Invalid request body")
			return
		}
		_ = r.Body.Close()
		if len(bytes.TrimSpace(body)) == 0 {
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
			return
		}

		var payload interface{}
		if err := json.Unmarshal(body, &payload); err != nil {
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
			return
		}
		clean, ok := sanitizeValue(payload)
		if !ok {
			writeSanitizeError(w, "Request contains unsafe input")
			return
		}
		encoded, err := json.Marshal(clean)
		if err != nil {
			writeSanitizeError(w, "Invalid request body")
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(encoded))
		next.ServeHTTP(w, r)
	})
}

func sanitizeValue(value interface{}) (interface{}, bool) {
	switch v := value.(type) {
	case string:
		v = strings.TrimSpace(v)
		if dangerousInputPattern.MatchString(v) {
			return nil, false
		}
		return v, true
	case []interface{}:
		for i := range v {
			clean, ok := sanitizeValue(v[i])
			if !ok {
				return nil, false
			}
			v[i] = clean
		}
		return v, true
	case map[string]interface{}:
		for key, item := range v {
			clean, ok := sanitizeValue(item)
			if !ok {
				return nil, false
			}
			v[key] = clean
		}
		return v, true
	default:
		return v, true
	}
}

func writeSanitizeError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(models.APIError{Error: message})
}
