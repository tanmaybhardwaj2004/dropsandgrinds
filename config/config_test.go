package config

import (
	"os"
	"testing"
)

func TestParseEnvInt_UsesFallbackWhenMissing(t *testing.T) {
	const key = "TEST_PARSE_INT_MISSING"
	_ = os.Unsetenv(key)

	got := parseEnvInt(key, 42)
	if got != 42 {
		t.Fatalf("expected fallback 42, got %d", got)
	}
}

func TestParseEnvInt_UsesFallbackWhenInvalid(t *testing.T) {
	const key = "TEST_PARSE_INT_INVALID"
	t.Setenv(key, "invalid")

	got := parseEnvInt(key, 15)
	if got != 15 {
		t.Fatalf("expected fallback 15, got %d", got)
	}
}

func TestParseEnvInt_ReturnsParsedValue(t *testing.T) {
	const key = "TEST_PARSE_INT_VALID"
	t.Setenv(key, "99")

	got := parseEnvInt(key, 0)
	if got != 99 {
		t.Fatalf("expected parsed value 99, got %d", got)
	}
}
