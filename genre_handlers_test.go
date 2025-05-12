package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func setupTestGenresServer(t *testing.T, clearTable bool) *httptest.Server {
	if clearTable {
		_ = ClearTable(TestAdminDB, "genres")
	}
	return httptest.NewServer(NewRouter())
}

func getFirstGenreID(t *testing.T, ts *httptest.Server, token string) string {
	req := createRequest(t, "GET", ts.URL+"/genres", token, nil)
	resp := executeRequest(t, req, http.StatusOK)
	defer resp.Body.Close()

	var genres []Genre
	parseResponseBody(t, resp, &genres)

	if len(genres) == 0 {
		t.Fatal("Expected at least one genre, got none")
	}

	return genres[0].ID
}

func TestGetGenres(t *testing.T) {
	tests := []struct {
		name           string
		seedData       bool
		role           string
		expectedStatus int
	}{
		{"Empty as Guest", false, "", http.StatusNotFound},
		{"Empty as User", false, "CLAIM_ROLE_USER", http.StatusNotFound},
		{"Empty as Admin", false, "CLAIM_ROLE_ADMIN", http.StatusNotFound},
		{"NonEmpty as Guest", true, "", http.StatusOK},
		{"NonEmpty as User", true, "CLAIM_ROLE_USER", http.StatusOK},
		{"NonEmpty as Admin", true, "CLAIM_ROLE_ADMIN", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestGenresServer(t, true)
			defer ts.Close()

			if tt.seedData {
				SeedGenres(TestAdminDB)
			}

			req := createRequest(t, "GET", ts.URL+"/genres", generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var genres []Genre
				parseResponseBody(t, resp, &genres)

				if len(genres) == 0 {
					t.Error("Expected non-empty genres list")
				}
			}
		})
	}
}

func TestGetGenreByID(t *testing.T) {
	setupValidIDTest := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestGenresServer(t, true)
		_ = SeedGenres(TestAdminDB)
		return ts, getFirstGenreID(t, ts, "")
	}

	tests := []struct {
		name           string
		setup          func(t *testing.T) (*httptest.Server, string)
		role           string
		expectedStatus int
	}{
		{
			"Unknown ID as Guest",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestGenresServer(t, true)
				SeedGenres(TestAdminDB)
				return ts, uuid.New().String()
			},
			"",
			http.StatusNotFound,
		},
		{
			"Unknown ID as User",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestGenresServer(t, true)
				SeedGenres(TestAdminDB)
				return ts, uuid.New().String()
			},
			"CLAIM_ROLE_USER",
			http.StatusNotFound,
		},
		{
			"Unknown ID as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestGenresServer(t, true)
				SeedGenres(TestAdminDB)
				return ts, uuid.New().String()
			},
			"CLAIM_ROLE_ADMIN",
			http.StatusNotFound,
		},
		{
			"Invalid ID as Guest",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestGenresServer(t, true)
				SeedGenres(TestAdminDB)
				return ts, "invalid-id"
			},
			"",
			http.StatusBadRequest,
		},
		{
			"Invalid ID as User",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestGenresServer(t, true)
				SeedGenres(TestAdminDB)
				return ts, "invalid-id"
			},
			"CLAIM_ROLE_USER",
			http.StatusBadRequest,
		},
		{
			"Invalid ID as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestGenresServer(t, true)
				SeedGenres(TestAdminDB)
				return ts, "invalid-id"
			},
			"CLAIM_ROLE_ADMIN",
			http.StatusBadRequest,
		},
		{
			"Valid ID as Guest",
			setupValidIDTest,
			"",
			http.StatusOK,
		},
		{
			"Valid ID as User",
			setupValidIDTest,
			"CLAIM_ROLE_USER",
			http.StatusOK,
		},
		{
			"Valid ID as Admin",
			setupValidIDTest,
			"CLAIM_ROLE_ADMIN",
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
			defer ts.Close()

			req := createRequest(t, "GET", ts.URL+"/genres/"+id, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var genre Genre
				parseResponseBody(t, resp, &genre)

				if genre.ID != id {
					t.Errorf("Expected ID %v; got %v", id, genre.ID)
				}
			}
		})
	}
}

func TestCreateGenre(t *testing.T) {
	validGenre := GenreData{
		Name:        "Test Genre",
		Description: "Test Description",
	}

	invalidGenre := GenreData{
		Name:        "",
		Description: "Test Description",
	}

	tests := []struct {
		name           string
		role           string
		body           interface{}
		setup          func(t *testing.T)
		expectedStatus int
	}{
		{
			"Forbidden Guest",
			"",
			validGenre,
			nil,
			http.StatusForbidden,
		},
		{
			"Forbidden User",
			"CLAIM_ROLE_USER",
			validGenre,
			nil,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			"CLAIM_ROLE_ADMIN",
			validGenre,
			nil,
			http.StatusCreated,
		},
		{
			"Invalid JSON Guest",
			"",
			"{invalid json}",
			nil,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON User",
			"CLAIM_ROLE_USER",
			"{invalid json}",
			nil,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON Admin",
			"CLAIM_ROLE_ADMIN",
			"{invalid json}",
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty fields in JSON Guest",
			"",
			invalidGenre,
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty fields in JSON User",
			"CLAIM_ROLE_USER",
			invalidGenre,
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty fields in JSON Admin",
			"CLAIM_ROLE_ADMIN",
			invalidGenre,
			nil,
			http.StatusBadRequest,
		},
		{
			"Conflict Admin",
			"CLAIM_ROLE_ADMIN",
			validGenre,
			func(t *testing.T) {
				_, err := TestAdminDB.Exec(context.Background(), "INSERT INTO genres (name, description) VALUES ($1, $2)", validGenre.Name, validGenre.Description)
				if err != nil {
					t.Fatalf("Failed to insert into test database: %v", err)
				}
			},
			http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestGenresServer(t, true)
			defer ts.Close()

			if tt.setup != nil {
				tt.setup(t)
			}

			req := createRequest(t, "POST", ts.URL+"/genres", generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusCreated {
				var created string
				parseResponseBody(t, resp, &created)

				if created == "" {
					t.Error("Expected non-empty ID in response")
				}

				if _, err := uuid.Parse(created); err != nil {
					t.Error("Неверный формат возврещённого UUID")
				}
			}
		})
	}
}

func TestUpdateGenre(t *testing.T) {
	validUpdateData := GenreData{
		Name:        "Updated Genre",
		Description: "Updated Description",
	}

	invalidUpdateData := GenreData{
		Name:        "",
		Description: "Updated Description",
	}

	// Setup function for tests needing existing genre
	setupExistingGenre := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestGenresServer(t, true)
		_ = SeedGenres(TestAdminDB)
		return ts, getFirstGenreID(t, ts, "")
	}

	tests := []struct {
		name           string
		role           string
		id             string
		body           interface{}
		setup          func(t *testing.T) (*httptest.Server, string)
		expectedStatus int
	}{
		{
			"Invalid UUID as Guest",
			"",
			"invalid-uuid",
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestGenresServer(t, true)
				SeedGenres(TestAdminDB)
				return ts, ""
			},
			http.StatusBadRequest,
		},
		{
			"Invalid UUID as User",
			"CLAIM_ROLE_USER",
			"invalid-uuid",
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestGenresServer(t, true)
				SeedGenres(TestAdminDB)
				return ts, ""
			},
			http.StatusBadRequest,
		},
		{
			"Invalid UUID as Admin",
			"CLAIM_ROLE_ADMIN",
			"invalid-uuid",
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestGenresServer(t, true)
				SeedGenres(TestAdminDB)
				return ts, ""
			},
			http.StatusBadRequest,
		},
		{
			"Unknown UUID as Guest",
			"",
			"",
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestGenresServer(t, true)
				SeedGenres(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusForbidden,
		},
		{
			"Unknown UUID as User",
			"CLAIM_ROLE_USER",
			"",
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestGenresServer(t, true)
				SeedGenres(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusForbidden,
		},
		{
			"Unknown UUID as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestGenresServer(t, true)
				SeedGenres(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusNotFound,
		},
		{
			"Invalid JSON as Guest",
			"",
			"",
			"invalid-json",
			setupExistingGenre,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as User",
			"CLAIM_ROLE_USER",
			"",
			"invalid-json",
			setupExistingGenre,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			"invalid-json",
			setupExistingGenre,
			http.StatusBadRequest,
		},
		{
			"Empty fields in JSON as Guest",
			"",
			"",
			invalidUpdateData,
			setupExistingGenre,
			http.StatusBadRequest,
		},
		{
			"Empty fields in JSON as User",
			"CLAIM_ROLE_USER",
			"",
			invalidUpdateData,
			setupExistingGenre,
			http.StatusBadRequest,
		},
		{
			"Empty fields in as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			invalidUpdateData,
			setupExistingGenre,
			http.StatusBadRequest,
		},
		{
			"Forbidden Guest",
			"",
			"",
			validUpdateData,
			setupExistingGenre,
			http.StatusForbidden,
		},
		{
			"Forbidden User",
			"CLAIM_ROLE_USER",
			"",
			validUpdateData,
			setupExistingGenre,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			validUpdateData,
			setupExistingGenre,
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
			defer ts.Close()

			// Use provided ID or fallback to id from setup
			effectiveID := tt.id
			if effectiveID == "" {
				effectiveID = id
			}

			req := createRequest(t, "PUT", ts.URL+"/genres/"+effectiveID, generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestDeleteGenre(t *testing.T) {
	// Setup function for tests needing existing genre
	setupExistingGenre := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestGenresServer(t, true)
		_ = SeedGenres(TestAdminDB)
		return ts, getFirstGenreID(t, ts, "")
	}

	tests := []struct {
		name           string
		role           string
		id             string
		setup          func(t *testing.T) (*httptest.Server, string)
		expectedStatus int
	}{
		{
			"Not Found as Guest",
			"",
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestGenresServer(t, true)
				SeedGenres(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusForbidden,
		},
		{
			"Not Found as User",
			"CLAIM_ROLE_USER",
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestGenresServer(t, true)
				SeedGenres(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusForbidden,
		},
		{
			"Not Found as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestGenresServer(t, true)
				SeedGenres(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusNotFound,
		},
		{
			"Invalid UUID as Guest",
			"",
			"invalid-uuid",
			func(t *testing.T) (*httptest.Server, string) {
				return setupTestGenresServer(t, true), "invalid-uuid"
			},
			http.StatusBadRequest,
		},
		{
			"Invalid UUID as User",
			"CLAIM_ROLE_USER",
			"invalid-uuid",
			func(t *testing.T) (*httptest.Server, string) {
				return setupTestGenresServer(t, true), "invalid-uuid"
			},
			http.StatusBadRequest,
		},
		{
			"Invalid UUID as Admin",
			"CLAIM_ROLE_ADMIN",
			"invalid-uuid",
			func(t *testing.T) (*httptest.Server, string) {
				return setupTestGenresServer(t, true), "invalid-uuid"
			},
			http.StatusBadRequest,
		},
		{
			"Forbidden as Guest",
			"",
			"",
			setupExistingGenre,
			http.StatusForbidden,
		},
		{
			"Forbidden as User",
			"CLAIM_ROLE_USER",
			"",
			setupExistingGenre,
			http.StatusForbidden,
		},
		{
			"Success as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			setupExistingGenre,
			http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
			defer ts.Close()

			// Use provided ID or fallback to id from setup
			effectiveID := tt.id
			if effectiveID == "" {
				effectiveID = id
			}

			req := createRequest(t, "DELETE", ts.URL+"/genres/"+effectiveID, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}
