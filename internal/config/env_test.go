package config

import (
	"os"
	"testing"
)

func TestGetEnvDefault(t *testing.T) {
	key := "TEST_CFG_DEFAULT"
	t.Setenv(key, "")

	if got := GetEnvDefault(key, "fallback"); got != "fallback" {
		t.Fatalf("got %q, want fallback", got)
	}

	t.Setenv(key, "value")
	if got := GetEnvDefault(key, "fallback"); got != "value" {
		t.Fatalf("got %q, want value", got)
	}
}

func TestGetEnvRequired(t *testing.T) {
	key := "TEST_CFG_REQUIRED"
	_ = os.Unsetenv(key)

	if _, err := GetEnvRequired(key); err == nil {
		t.Fatal("expected error for missing env")
	} else if err.Error() != "required environment variable "+key+" is not set" {
		t.Fatalf("unexpected error message: %q", err.Error())
	}

	t.Setenv(key, "set")
	got, err := GetEnvRequired(key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "set" {
		t.Fatalf("got %q, want set", got)
	}
}
