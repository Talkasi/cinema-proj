//go:build smoke

package smoke

import (
	"bytes"
	dto "cw/internal/dto/models"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"
)

type smokeClient struct {
	baseURL string
	http    *http.Client
	token   string
}

func newSmokeClient() *smokeClient {
	baseURL := os.Getenv("SMOKE_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	return &smokeClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *smokeClient) doJSON(t *testing.T, method, path string, body any, out any, wantStatus int) {
	t.Helper()

	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
	}

	req, err := http.NewRequest(method, c.baseURL+path, bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != wantStatus {
		var errResp dto.ErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		t.Fatalf("%s %s status=%d want=%d message=%q", method, path, resp.StatusCode, wantStatus, errResp.Message)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			t.Fatalf("decode response %s %s: %v", method, path, err)
		}
	}
}

func TestAPISmoke_RegisterLoginGenres(t *testing.T) {
	client := newSmokeClient()

	suffix := fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(1000))
	email := "smoke+" + suffix + "@example.com"
	password := "smokepass123"

	registerReq := dto.CreateUserRequest{
		Name:         "Smoke User",
		Email:        email,
		PasswordHash: password,
		BirthDate:    "1990-01-01",
	}

	var registerResp dto.RegisterResponse
	client.doJSON(t, http.MethodPost, "/api/v1/auth/register", registerReq, &registerResp, http.StatusCreated)
	if registerResp.ID == "" {
		t.Fatal("register response id is empty")
	}

	var loginResp dto.AuthResponse
	client.doJSON(t, http.MethodPost, "/api/v1/auth/login", dto.LoginRequest{
		Email:        email,
		PasswordHash: password,
	}, &loginResp, http.StatusOK)

	if loginResp.UserID == "" {
		t.Fatal("login response user_id is empty")
	}

	if loginResp.Token != "" {
		client.token = loginResp.Token
		var userResp dto.UserResponse
		client.doJSON(t, http.MethodGet, "/api/v1/users/"+registerResp.ID, nil, &userResp, http.StatusOK)
		if userResp.ID != registerResp.ID {
			t.Fatalf("user id mismatch: got %s want %s", userResp.ID, registerResp.ID)
		}
	}

	var genresResp dto.PaginatedGenreResponse
	client.doJSON(t, http.MethodGet, "/api/v1/genres/?limit=1", nil, &genresResp, http.StatusOK)

	if len(genresResp.Data) > 0 {
		var genreResp dto.GenreResponse
		client.doJSON(t, http.MethodGet, "/api/v1/genres/"+genresResp.Data[0].ID, nil, &genreResp, http.StatusOK)
		if genreResp.ID != genresResp.Data[0].ID {
			t.Fatalf("genre id mismatch: got %s want %s", genreResp.ID, genresResp.Data[0].ID)
		}
	}
}
