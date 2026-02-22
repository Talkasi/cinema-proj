package main

import "testing"

func TestParseOptionalInt(t *testing.T) {
	t.Parallel()

	if got := parseOptionalInt("", 7); got != 7 {
		t.Fatalf("empty: got %d, want 7", got)
	}
	if got := parseOptionalInt(" 42 ", 7); got != 42 {
		t.Fatalf("valid: got %d, want 42", got)
	}
	if got := parseOptionalInt("bad", 7); got != 7 {
		t.Fatalf("invalid: got %d, want 7", got)
	}
}

func TestParseOptionalFloat(t *testing.T) {
	t.Parallel()

	if got := parseOptionalFloat("", 1.5); got != 1.5 {
		t.Fatalf("empty: got %v, want 1.5", got)
	}
	if got := parseOptionalFloat(" 3.25 ", 1.5); got != 3.25 {
		t.Fatalf("valid: got %v, want 3.25", got)
	}
	if got := parseOptionalFloat("bad", 1.5); got != 1.5 {
		t.Fatalf("invalid: got %v, want 1.5", got)
	}
}
