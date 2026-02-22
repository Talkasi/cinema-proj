package handler

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"cw/internal/utils"
)

func TestParsePaginationParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		query     string
		wantPage  int
		wantLimit int
	}{
		{
			name:      "defaults",
			query:     "",
			wantPage:  1,
			wantLimit: 20,
		},
		{
			name:      "valid values",
			query:     "page=3&limit=50",
			wantPage:  3,
			wantLimit: 50,
		},
		{
			name:      "invalid values fallback",
			query:     "page=0&limit=-1",
			wantPage:  1,
			wantLimit: 20,
		},
		{
			name:      "limit above max fallback",
			query:     "page=2&limit=1000",
			wantPage:  2,
			wantLimit: 20,
		},
		{
			name:      "non numeric fallback",
			query:     "page=abc&limit=xyz",
			wantPage:  1,
			wantLimit: 20,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest("GET", "/?"+tt.query, nil)
			page, limit := parsePaginationParams(req)

			if page != tt.wantPage {
				t.Fatalf("page = %d, want %d", page, tt.wantPage)
			}
			if limit != tt.wantLimit {
				t.Fatalf("limit = %d, want %d", limit, tt.wantLimit)
			}
		})
	}
}

func TestDecodeJSONBody(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"cinema"}`))
		var got payload

		if err := decodeJSONBody(req, &got); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "cinema" {
			t.Fatalf("name = %q, want cinema", got.Name)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":`))
		var got payload

		err := decodeJSONBody(req, &got)
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Code != 400 {
			t.Fatalf("code = %d, want 400", err.Code)
		}
	})

	t.Run("empty body returns bad request", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("POST", "/", strings.NewReader(""))
		var got payload

		err := decodeJSONBody(req, &got)
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Code != 400 {
			t.Fatalf("code = %d, want 400", err.Code)
		}
	})

	t.Run("unknown fields are ignored", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"cinema","extra":"ignored"}`))
		var got payload

		if err := decodeJSONBody(req, &got); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "cinema" {
			t.Fatalf("name = %q, want cinema", got.Name)
		}
	})
}

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	err := writeJSON(rr, 201, map[string]string{"id": "123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rr.Code != 201 {
		t.Fatalf("status = %d, want 201", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %q, want application/json", got)
	}
	if !strings.Contains(rr.Body.String(), `"id":"123"`) {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}

func TestDecodeAndValidateJSONBody(t *testing.T) {
	t.Parallel()

	type payload struct {
		Email string `json:"email" validate:"required,email"`
	}

	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"email":"bad"}`))
	var got payload

	err := decodeAndValidateJSONBody(req, &got)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Code != 400 {
		t.Fatalf("code = %d, want 400", err.Code)
	}
	if err.Message != "Invalid input" {
		t.Fatalf("message = %q, want %q", err.Message, "Invalid input")
	}

	// Ensure underlying error remains available for internal logging/debugging.
	if err.Err == nil {
		t.Fatal("expected underlying validation error")
	}
}

func TestValidationErrorResponseIsMasked(t *testing.T) {
	t.Parallel()

	type payload struct {
		Email string `json:"email" validate:"required,email"`
	}

	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"email":"bad"}`))
	var got payload

	err := decodeAndValidateJSONBody(req, &got)
	if err == nil {
		t.Fatal("expected error")
	}

	rr := httptest.NewRecorder()
	utils.WriteError(rr, err)

	if rr.Code != 400 {
		t.Fatalf("status = %d, want 400", rr.Code)
	}

	var resp struct {
		Message string `json:"message"`
	}
	if decodeErr := json.NewDecoder(rr.Body).Decode(&resp); decodeErr != nil {
		t.Fatalf("decode response: %v", decodeErr)
	}

	if resp.Message != "Invalid input" {
		t.Fatalf("message = %q", resp.Message)
	}
	if strings.Contains(strings.ToLower(resp.Message), "email") {
		t.Fatalf("unexpected validation details in response: %q", resp.Message)
	}
}
