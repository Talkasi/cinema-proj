package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestGetGenres(t *testing.T) {
	tests := []struct {
		name           string
		seedData       bool
		role           string
		expectedStatus int
	}{
		{"Empty as Guest", false, "", http.StatusNotFound},
		{"Empty as User", false, os.Getenv("CLAIM_ROLE_USER"), http.StatusNotFound},
		{"Empty as Admin", false, os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusNotFound},
		{"NonEmpty as Guest", true, "", http.StatusOK},
		{"NonEmpty as User", true, os.Getenv("CLAIM_ROLE_USER"), http.StatusOK},
		{"NonEmpty as Admin", true, os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
			SeedUsers(TestAdminDB)
			defer ts.Close()

			if tt.seedData {
				SeedAll(TestAdminDB)
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
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, GenresData[0].ID
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
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			"",
			http.StatusNotFound,
		},
		{
			"Unknown ID as User",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			os.Getenv("CLAIM_ROLE_USER"),
			http.StatusNotFound,
		},
		{
			"Unknown ID as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			os.Getenv("CLAIM_ROLE_ADMIN"),
			http.StatusNotFound,
		},
		{
			"Invalid ID as Guest",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, "invalid-id"
			},
			"",
			http.StatusBadRequest,
		},
		{
			"Invalid ID as User",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, "invalid-id"
			},
			os.Getenv("CLAIM_ROLE_USER"),
			http.StatusBadRequest,
		},
		{
			"Invalid ID as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, "invalid-id"
			},
			os.Getenv("CLAIM_ROLE_ADMIN"),
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
			os.Getenv("CLAIM_ROLE_USER"),
			http.StatusOK,
		},
		{
			"Valid ID as Admin",
			setupValidIDTest,
			os.Getenv("CLAIM_ROLE_ADMIN"),
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
			os.Getenv("CLAIM_ROLE_USER"),
			validGenre,
			nil,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
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
			os.Getenv("CLAIM_ROLE_USER"),
			"{invalid json}",
			nil,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
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
			os.Getenv("CLAIM_ROLE_USER"),
			invalidGenre,
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty fields in JSON Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			invalidGenre,
			nil,
			http.StatusBadRequest,
		},
		{
			"Conflict Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			validGenre,
			func(t *testing.T) {
				_, err := TestAdminDB.Exec(context.Background(), "INSERT INTO genres (name, description) VALUES ($1, $2)", validGenre.Name, validGenre.Description)
				if err != nil {
					t.Fatalf("Failed to insert into test database: %v", err)
				}
			},
			http.StatusConflict,
		},
		{
			"Name with invalid characters",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			GenreData{Name: "Action123!", Description: "Valid"},
			nil,
			http.StatusBadRequest,
		},
		{
			"Name with only spaces",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			GenreData{Name: "   ", Description: "Valid"},
			nil,
			http.StatusBadRequest,
		},
		{
			"Name exactly 64 chars",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			GenreData{Name: strings.Repeat("a", 64), Description: "Valid"},
			nil,
			http.StatusCreated,
		},
		{
			"Name too long (65 chars)",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			GenreData{Name: strings.Repeat("a", 65), Description: "Valid"},
			nil,
			http.StatusBadRequest,
		},
		{
			"Name with Unicode (кириллица)",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			GenreData{Name: "Комедия", Description: "Valid"},
			nil,
			http.StatusCreated,
		},
		{
			"Name with hyphen",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			GenreData{Name: "Sci-Fi", Description: "Valid"},
			nil,
			http.StatusCreated,
		},
		{
			"Empty description",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			GenreData{Name: "Valid", Description: ""},
			nil,
			http.StatusBadRequest,
		},
		{
			"Description with only spaces",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			GenreData{Name: "Valid", Description: "   "},
			nil,
			http.StatusBadRequest,
		},
		{
			"Description exactly 1000 chars",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			GenreData{Name: "Valid", Description: strings.Repeat("a", 1000)},
			nil,
			http.StatusCreated,
		},
		{
			"Description too long (1001 chars)",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			GenreData{Name: "Valid", Description: strings.Repeat("a", 1001)},
			nil,
			http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
			SeedUsers(TestAdminDB)
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

	setupExistingGenre := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, GenresData[0].ID
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
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, ""
			},
			http.StatusBadRequest,
		},
		{
			"Invalid UUID as User",
			os.Getenv("CLAIM_ROLE_USER"),
			"invalid-uuid",
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, ""
			},
			http.StatusBadRequest,
		},
		{
			"Invalid UUID as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"invalid-uuid",
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
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
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusForbidden,
		},
		{
			"Unknown UUID as User",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusForbidden,
		},
		{
			"Unknown UUID as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
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
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			"invalid-json",
			setupExistingGenre,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
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
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			invalidUpdateData,
			setupExistingGenre,
			http.StatusBadRequest,
		},
		{
			"Empty fields in as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
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
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			validUpdateData,
			setupExistingGenre,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			validUpdateData,
			setupExistingGenre,
			http.StatusOK,
		},
		{
			"Name with invalid characters",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			GenreData{Name: "Action123!", Description: "Valid"},
			setupExistingGenre,
			http.StatusBadRequest,
		},
		{
			"Name with only spaces",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			GenreData{Name: "   ", Description: "Valid"},
			setupExistingGenre,
			http.StatusBadRequest,
		},
		{
			"Name exactly 64 chars",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			GenreData{Name: strings.Repeat("a", 64), Description: "Valid"},
			setupExistingGenre,
			http.StatusOK,
		},
		{
			"Name too long (65 chars)",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			GenreData{Name: strings.Repeat("a", 65), Description: "Valid"},
			setupExistingGenre,
			http.StatusBadRequest,
		},
		{
			"Name with Unicode (кириллица)",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			GenreData{Name: "Комедия новая", Description: "Valid"},
			setupExistingGenre,
			http.StatusOK,
		},
		{
			"Name with hyphen",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			GenreData{Name: "Sci-Fi", Description: "Valid"},
			setupExistingGenre,
			http.StatusOK,
		},
		{
			"Empty description",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			GenreData{Name: "Valid", Description: ""},
			setupExistingGenre,
			http.StatusBadRequest,
		},
		{
			"Description with only spaces",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			GenreData{Name: "Valid", Description: "   "},
			setupExistingGenre,
			http.StatusBadRequest,
		},
		{
			"Description exactly 1000 chars",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			GenreData{Name: "Valid", Description: strings.Repeat("a", 1000)},
			setupExistingGenre,
			http.StatusOK,
		},
		{
			"Description too long (1001 chars)",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			GenreData{Name: "Valid", Description: strings.Repeat("a", 1001)},
			setupExistingGenre,
			http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
			SeedUsers(TestAdminDB)
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
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, GenresData[0].ID
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
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusForbidden,
		},
		{
			"Not Found as User",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusForbidden,
		},
		{
			"Not Found as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusNotFound,
		},
		{
			"Invalid UUID as Guest",
			"",
			"invalid-uuid",
			func(t *testing.T) (*httptest.Server, string) {
				return setupTestServer(), "invalid-uuid"
			},
			http.StatusBadRequest,
		},
		{
			"Invalid UUID as User",
			os.Getenv("CLAIM_ROLE_USER"),
			"invalid-uuid",
			func(t *testing.T) (*httptest.Server, string) {
				return setupTestServer(), "invalid-uuid"
			},
			http.StatusBadRequest,
		},
		{
			"Invalid UUID as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"invalid-uuid",
			func(t *testing.T) (*httptest.Server, string) {
				return setupTestServer(), "invalid-uuid"
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
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			setupExistingGenre,
			http.StatusForbidden,
		},
		{
			"Dependency error as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			setupExistingGenre,
			http.StatusConflict,
		},
		{
			"Success as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, GenresData[6].ID
			},
			http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
			SeedUsers(TestAdminDB)
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

func TestCreateGenreDBError(t *testing.T) {
	ts := setupTestServer()
	SeedUsers(TestAdminDB)
	defer ts.Close()

	// Создаем ситуацию с ошибкой БД
	TestAdminDB.Close()
	TestGuestDB.Close()
	TestUserDB.Close()

	req := createRequest(t, "POST", ts.URL+"/genres",
		generateToken(t, os.Getenv("CLAIM_ROLE_ADMIN")),
		GenreData{Name: "Test", Description: "Test"})
	resp := executeRequest(t, req, http.StatusInternalServerError)
	defer resp.Body.Close()

	if err := InitTestDB(); err != nil {
		log.Fatal("ошибка подключения к БД: ", err)
	}
}

func TestSearchGenres(t *testing.T) {
	setupWithGenres := func(t *testing.T) *httptest.Server {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts
	}

	tests := []struct {
		name           string
		query          string
		setup          func(t *testing.T) *httptest.Server
		expectedStatus int
		expectedCount  int
		role           string
	}{
		{
			"Пустой запрос - ошибка",
			"",
			setupWithGenres,
			http.StatusBadRequest,
			0,
			os.Getenv("CLAIM_ROLE_USER"),
		},
		{
			"Только пробельные символы - ошибка",
			"          		",
			setupWithGenres,
			http.StatusBadRequest,
			0,
			os.Getenv("CLAIM_ROLE_USER"),
		},
		{
			"Короткий запрос",
			"д",
			setupWithGenres,
			http.StatusOK,
			3,
			os.Getenv("CLAIM_ROLE_USER"),
		},
		{
			"Нет совпадений",
			"вестерн",
			setupWithGenres,
			http.StatusNotFound,
			0,
			os.Getenv("CLAIM_ROLE_USER"),
		},
		{
			"Точное совпадение - Драма",
			"Драма",
			setupWithGenres,
			http.StatusOK,
			1,
			os.Getenv("CLAIM_ROLE_USER"),
		},
		{
			"Частичное совпадение - 'рама'",
			"рама",
			setupWithGenres,
			http.StatusOK,
			1,
			os.Getenv("CLAIM_ROLE_USER"),
		},
		{
			"Частичное совпадение - 'ик'",
			"ик",
			setupWithGenres,
			http.StatusOK,
			4,
			os.Getenv("CLAIM_ROLE_USER"),
		},
		{
			"Поиск без учета регистра - 'кОмЕдИя'",
			"кОмЕдИя",
			setupWithGenres,
			http.StatusOK,
			1,
			os.Getenv("CLAIM_ROLE_USER"),
		},
		{
			"Поиск с пробелами - 'Научная фантастика'",
			"Научная фантастика",
			setupWithGenres,
			http.StatusOK,
			1,
			os.Getenv("CLAIM_ROLE_USER"),
		},
		{
			"Частичное совпадение с пробелами - 'научная'",
			"  научная    ",
			setupWithGenres,
			http.StatusOK,
			1,
			os.Getenv("CLAIM_ROLE_USER"),
		},
		{
			"Админ имеет доступ",
			"Драма",
			setupWithGenres,
			http.StatusOK,
			1,
			os.Getenv("CLAIM_ROLE_ADMIN"),
		},
		{
			"Гость имеет доступ",
			"Драма",
			setupWithGenres,
			http.StatusOK,
			1,
			"",
		},
		{
			"Специальные символы в запросе - 'фэнтези'",
			"фэнтези/",
			setupWithGenres,
			http.StatusNotFound,
			0,
			os.Getenv("CLAIM_ROLE_USER"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := tt.setup(t)
			SeedUsers(TestAdminDB)
			defer ts.Close()

			req := createRequest(t, "GET", ts.URL+"/genres/search?query="+url.QueryEscape(tt.query), generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var genres []Genre
				if err := json.NewDecoder(resp.Body).Decode(&genres); err != nil {
					t.Fatalf("Could not decode response: %v", err)
				}

				if len(genres) != tt.expectedCount {
					t.Errorf("Expected %d genres, got %d: %v", tt.expectedCount, len(genres), genres)
				}

				lowerQuery := strings.ToLower(tt.query)
				lowerQuery = PrepareString(lowerQuery)
				for _, genre := range genres {
					if !strings.Contains(strings.ToLower(genre.Name), lowerQuery) {
						t.Errorf("Genre name '%s' does not contain query '%s'", genre.Name, tt.query)
					}
				}
			}
		})
	}
}
