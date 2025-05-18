package main

import (
	"context"
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
		{"DBError as Guest", true, "", http.StatusInternalServerError},
		{"DBError as User", true, os.Getenv("CLAIM_ROLE_USER"), http.StatusInternalServerError},
		{"DBError as Admin", true, os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
			SeedUsers(TestAdminDB)
			defer ts.Close()

			if tt.seedData {
				SeedAll(TestAdminDB)
			}

			if strings.Split(tt.name, " ")[0] == "DBError" {
				TestAdminDB.Close()
				TestGuestDB.Close()
				TestUserDB.Close()
				defer InitTestDB()
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
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, GenresData[0].ID
			},
			"",
			http.StatusOK,
		},
		{
			"Valid ID as User",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, GenresData[0].ID
			},
			os.Getenv("CLAIM_ROLE_USER"),
			http.StatusOK,
		},
		{
			"Valid ID as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, GenresData[0].ID
			},
			os.Getenv("CLAIM_ROLE_ADMIN"),
			http.StatusOK,
		},
		{
			"DBError as Guest",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				TestAdminDB.Close()
				TestGuestDB.Close()
				TestUserDB.Close()
				return ts, GenresData[0].ID
			},
			"",
			http.StatusInternalServerError,
		},
		{
			"DBError as User",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				TestAdminDB.Close()
				TestGuestDB.Close()
				TestUserDB.Close()
				return ts, GenresData[0].ID
			},
			os.Getenv("CLAIM_ROLE_USER"),
			http.StatusInternalServerError,
		},
		{
			"DBError as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				TestAdminDB.Close()
				TestGuestDB.Close()
				TestUserDB.Close()
				return ts, GenresData[0].ID
			},
			os.Getenv("CLAIM_ROLE_ADMIN"),
			http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
			defer func() {
				ts.Close()
				if strings.Split(tt.name, " ")[0] == "DBError" {
					InitTestDB()
				}
			}()

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

func TestSearchGenres(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		role           string
		expectedStatus int
		expectedCount  int
	}{
		{"Empty query as Guest", "", "", http.StatusBadRequest, 0},
		{"Empty query as User", "", os.Getenv("CLAIM_ROLE_USER"), http.StatusBadRequest, 0},
		{"Empty query as Admin", "", os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusBadRequest, 0},
		{"Whitespace query as Guest", "    ", "", http.StatusBadRequest, 0},
		{"Whitespace query as User", "    ", os.Getenv("CLAIM_ROLE_USER"), http.StatusBadRequest, 0},
		{"Whitespace query as Admin", "    ", os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusBadRequest, 0},
		{"Short query as Guest", "д", "", http.StatusOK, 3},
		{"Short query as User", "д", os.Getenv("CLAIM_ROLE_USER"), http.StatusOK, 3},
		{"Short query as Admin", "д", os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusOK, 3},
		{"No matches as Guest", "вестерн", "", http.StatusNotFound, 0},
		{"No matches as User", "вестерн", os.Getenv("CLAIM_ROLE_USER"), http.StatusNotFound, 0},
		{"No matches as Admin", "вестерн", os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusNotFound, 0},
		{"Exact match as Guest", "Драма", "", http.StatusOK, 1},
		{"Exact match as User", "Драма", os.Getenv("CLAIM_ROLE_USER"), http.StatusOK, 1},
		{"Exact match as Admin", "Драма", os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusOK, 1},
		{"Partial match as Guest", "рама", "", http.StatusOK, 1},
		{"Partial match as User", "рама", os.Getenv("CLAIM_ROLE_USER"), http.StatusOK, 1},
		{"Partial match as Admin", "рама", os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusOK, 1},
		{"Case insensitive as Guest", "кОмЕдИя", "", http.StatusOK, 1},
		{"Case insensitive as User", "кОмЕдИя", os.Getenv("CLAIM_ROLE_USER"), http.StatusOK, 1},
		{"Case insensitive as Admin", "кОмЕдИя", os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusOK, 1},
		{"With spaces as Guest", "Научная фантастика", "", http.StatusOK, 1},
		{"With spaces as User", "Научная фантастика", os.Getenv("CLAIM_ROLE_USER"), http.StatusOK, 1},
		{"With spaces as Admin", "Научная фантастика", os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusOK, 1},
		{"Partial with spaces as Guest", "  научная    ", "", http.StatusOK, 1},
		{"Partial with spaces as User", "  научная    ", os.Getenv("CLAIM_ROLE_USER"), http.StatusOK, 1},
		{"Partial with spaces as Admin", "  научная    ", os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusOK, 1},
		{"Special chars as Guest", "фэнтези/", "", http.StatusNotFound, 0},
		{"Special chars as User", "фэнтези/", os.Getenv("CLAIM_ROLE_USER"), http.StatusNotFound, 0},
		{"Special chars as Admin", "фэнтези/", os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusNotFound, 0},
		{"DBError as Guest", "Драма", "", http.StatusInternalServerError, 0},
		{"DBError as User", "Драма", os.Getenv("CLAIM_ROLE_USER"), http.StatusInternalServerError, 0},
		{"DBError as Admin", "Драма", os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusInternalServerError, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			defer ts.Close()

			if strings.Split(tt.name, " ")[0] == "DBError" {
				TestAdminDB.Close()
				TestGuestDB.Close()
				TestUserDB.Close()
				defer InitTestDB()
			}

			req := createRequest(t, "GET", ts.URL+"/genres/search?query="+url.QueryEscape(tt.query), generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var genres []Genre
				parseResponseBody(t, resp, &genres)

				if len(genres) != tt.expectedCount {
					t.Errorf("Expected %d genres, got %d", tt.expectedCount, len(genres))
				}
			}
		})
	}
}

func TestCreateGenre(t *testing.T) {
	tests := []struct {
		name           string
		role           string
		body           interface{}
		setup          func(t *testing.T)
		expectedStatus int
	}{
		{"Valid as Guest", "", GenreData{Name: "Test", Description: "Test"}, nil, http.StatusForbidden},
		{"Valid as User", os.Getenv("CLAIM_ROLE_USER"), GenreData{Name: "Test", Description: "Test"}, nil, http.StatusForbidden},
		{"Valid as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), GenreData{Name: "Test", Description: "Test"}, nil, http.StatusCreated},
		{"Invalid JSON as Guest", "", "{invalid json}", nil, http.StatusBadRequest},
		{"Invalid JSON as User", os.Getenv("CLAIM_ROLE_USER"), "{invalid json}", nil, http.StatusBadRequest},
		{"Invalid JSON as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "{invalid json}", nil, http.StatusBadRequest},
		{"Empty fields as Guest", "", GenreData{Name: "", Description: "Test"}, nil, http.StatusBadRequest},
		{"Empty fields as User", os.Getenv("CLAIM_ROLE_USER"), GenreData{Name: "", Description: "Test"}, nil, http.StatusBadRequest},
		{"Empty fields as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), GenreData{Name: "", Description: "Test"}, nil, http.StatusBadRequest},
		{"Duplicate name as Guest", "", GenreData{Name: "Action", Description: "Test"}, func(t *testing.T) {
			_, err := TestAdminDB.Exec(context.Background(), "INSERT INTO genres (name, description) VALUES ($1, $2)", "Action", "Test")
			if err != nil {
				t.Fatal(err)
			}
		}, http.StatusForbidden},
		{"Duplicate name as User", os.Getenv("CLAIM_ROLE_USER"), GenreData{Name: "Action", Description: "Test"}, func(t *testing.T) {
			_, err := TestAdminDB.Exec(context.Background(), "INSERT INTO genres (name, description) VALUES ($1, $2)", "Action", "Test")
			if err != nil {
				t.Fatal(err)
			}
		}, http.StatusForbidden},
		{"Duplicate name as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), GenreData{Name: "Action", Description: "Test"}, func(t *testing.T) {
			_, err := TestAdminDB.Exec(context.Background(), "INSERT INTO genres (name, description) VALUES ($1, $2)", "Action", "Test")
			if err != nil {
				t.Fatal(err)
			}
		}, http.StatusConflict},
		{"Invalid chars as Guest", "", GenreData{Name: "Action123!", Description: "Test"}, nil, http.StatusBadRequest},
		{"Invalid chars as User", os.Getenv("CLAIM_ROLE_USER"), GenreData{Name: "Action123!", Description: "Test"}, nil, http.StatusBadRequest},
		{"Invalid chars as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), GenreData{Name: "Action123!", Description: "Test"}, nil, http.StatusBadRequest},
		{"Whitespace name as Guest", "", GenreData{Name: "   ", Description: "Test"}, nil, http.StatusBadRequest},
		{"Whitespace name as User", os.Getenv("CLAIM_ROLE_USER"), GenreData{Name: "   ", Description: "Test"}, nil, http.StatusBadRequest},
		{"Whitespace name as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), GenreData{Name: "   ", Description: "Test"}, nil, http.StatusBadRequest},
		{"64 chars name as Guest", "", GenreData{Name: strings.Repeat("a", 64), Description: "Test"}, nil, http.StatusForbidden},
		{"64 chars name as User", os.Getenv("CLAIM_ROLE_USER"), GenreData{Name: strings.Repeat("a", 64), Description: "Test"}, nil, http.StatusForbidden},
		{"64 chars name as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), GenreData{Name: strings.Repeat("a", 64), Description: "Test"}, nil, http.StatusCreated},
		{"65 chars name as Guest", "", GenreData{Name: strings.Repeat("a", 65), Description: "Test"}, nil, http.StatusBadRequest},
		{"65 chars name as User", os.Getenv("CLAIM_ROLE_USER"), GenreData{Name: strings.Repeat("a", 65), Description: "Test"}, nil, http.StatusBadRequest},
		{"65 chars name as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), GenreData{Name: strings.Repeat("a", 65), Description: "Test"}, nil, http.StatusBadRequest},
		{"Hyphen name as Guest", "", GenreData{Name: "Sci-Fi", Description: "Test"}, nil, http.StatusForbidden},
		{"Hyphen name as User", os.Getenv("CLAIM_ROLE_USER"), GenreData{Name: "Sci-Fi", Description: "Test"}, nil, http.StatusForbidden},
		{"Hyphen name as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), GenreData{Name: "Sci-Fi", Description: "Test"}, nil, http.StatusCreated},
		{"Empty desc as Guest", "", GenreData{Name: "Test", Description: ""}, nil, http.StatusBadRequest},
		{"Empty desc as User", os.Getenv("CLAIM_ROLE_USER"), GenreData{Name: "Test", Description: ""}, nil, http.StatusBadRequest},
		{"Empty desc as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), GenreData{Name: "Test", Description: ""}, nil, http.StatusBadRequest},
		{"Whitespace desc as Guest", "", GenreData{Name: "Test", Description: "   "}, nil, http.StatusBadRequest},
		{"Whitespace desc as User", os.Getenv("CLAIM_ROLE_USER"), GenreData{Name: "Test", Description: "   "}, nil, http.StatusBadRequest},
		{"Whitespace desc as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), GenreData{Name: "Test", Description: "   "}, nil, http.StatusBadRequest},
		{"1000 chars desc as Guest", "", GenreData{Name: "Test", Description: strings.Repeat("a", 1000)}, nil, http.StatusForbidden},
		{"1000 chars desc as User", os.Getenv("CLAIM_ROLE_USER"), GenreData{Name: "Test", Description: strings.Repeat("a", 1000)}, nil, http.StatusForbidden},
		{"1000 chars desc as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), GenreData{Name: "Test", Description: strings.Repeat("a", 1000)}, nil, http.StatusCreated},
		{"1001 chars desc as Guest", "", GenreData{Name: "Test", Description: strings.Repeat("a", 1001)}, nil, http.StatusBadRequest},
		{"1001 chars desc as User", os.Getenv("CLAIM_ROLE_USER"), GenreData{Name: "Test", Description: strings.Repeat("a", 1001)}, nil, http.StatusBadRequest},
		{"1001 chars desc as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), GenreData{Name: "Test", Description: strings.Repeat("a", 1001)}, nil, http.StatusBadRequest},
		{"DBError as Guest", "", GenreData{Name: "Test", Description: "Test"}, func(t *testing.T) {
			TestAdminDB.Close()
			TestGuestDB.Close()
			TestUserDB.Close()
		}, http.StatusInternalServerError},
		{"DBError as User", os.Getenv("CLAIM_ROLE_USER"), GenreData{Name: "Test", Description: "Test"}, func(t *testing.T) {
			TestAdminDB.Close()
			TestGuestDB.Close()
			TestUserDB.Close()
		}, http.StatusInternalServerError},
		{"DBError as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), GenreData{Name: "Test", Description: "Test"}, func(t *testing.T) {
			TestAdminDB.Close()
			TestGuestDB.Close()
			TestUserDB.Close()
		}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
			SeedUsers(TestAdminDB)
			defer func() {
				ts.Close()
				if strings.Split(tt.name, " ")[0] == "DBError" {
					InitTestDB()
				}
			}()

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
					t.Error("Expected non-empty ID")
				}

				if _, err := uuid.Parse(created); err != nil {
					t.Error("Неверный формат возвращённого UUID")
				}
			}
		})
	}
}

func TestUpdateGenre(t *testing.T) {
	validTestPreparator := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		SeedAll(TestAdminDB)
		return ts, GenresData[0].ID
	}

	invalidTestPreparator := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		SeedAll(TestAdminDB)
		return ts, ""
	}

	unknownTestPreparator := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		SeedAll(TestAdminDB)
		return ts, uuid.New().String()
	}

	tests := []struct {
		name           string
		role           string
		id             string
		body           interface{}
		setup          func(t *testing.T) (*httptest.Server, string)
		expectedStatus int
	}{
		{"Invalid ID as Guest", "", "invalid-uuid", GenreData{Name: "Test", Description: "Test"}, invalidTestPreparator, http.StatusBadRequest},
		{"Invalid ID as User", os.Getenv("CLAIM_ROLE_USER"), "invalid-uuid", GenreData{Name: "Test", Description: "Test"}, invalidTestPreparator, http.StatusBadRequest},
		{"Invalid ID as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "invalid-uuid", GenreData{Name: "Test", Description: "Test"}, invalidTestPreparator, http.StatusBadRequest},
		{"Unknown ID as Guest", "", "", GenreData{Name: "Test", Description: "Test"}, unknownTestPreparator, http.StatusForbidden},
		{"Unknown ID as User", os.Getenv("CLAIM_ROLE_USER"), "", GenreData{Name: "Test", Description: "Test"}, unknownTestPreparator, http.StatusForbidden},
		{"Unknown ID as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", GenreData{Name: "Test", Description: "Test"}, unknownTestPreparator, http.StatusNotFound},
		{"Invalid JSON as Guest", "", "", "invalid-json", validTestPreparator, http.StatusBadRequest},
		{"Invalid JSON as User", os.Getenv("CLAIM_ROLE_USER"), "", "invalid-json", validTestPreparator, http.StatusBadRequest},
		{"Invalid JSON as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", "invalid-json", validTestPreparator, http.StatusBadRequest},
		{"Duplicate name as Guest", "", "", GenreData{Name: "Action", Description: "Test"}, func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			_, err := TestAdminDB.Exec(context.Background(), "INSERT INTO genres (name, description) VALUES ($1, $2)", "Action", "Test")
			if err != nil {
				t.Fatal(err)
			}
			return ts, GenresData[0].ID
		}, http.StatusForbidden},
		{"Duplicate name as User", os.Getenv("CLAIM_ROLE_USER"), "", GenreData{Name: "Action", Description: "Test"}, func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			_, err := TestAdminDB.Exec(context.Background(), "INSERT INTO genres (name, description) VALUES ($1, $2)", "Action", "Test")
			if err != nil {
				t.Fatal(err)
			}
			return ts, GenresData[0].ID
		}, http.StatusForbidden},
		{"Duplicate name as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", GenreData{Name: "Action", Description: "Test"}, func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			_, err := TestAdminDB.Exec(context.Background(), "INSERT INTO genres (name, description) VALUES ($1, $2)", "Action", "Test")
			if err != nil {
				t.Fatal(err)
			}
			return ts, GenresData[0].ID
		}, http.StatusConflict},
		{"Empty fields as Guest", "", "", GenreData{Name: "", Description: "Test"}, validTestPreparator, http.StatusBadRequest},
		{"Empty fields as User", os.Getenv("CLAIM_ROLE_USER"), "", GenreData{Name: "", Description: "Test"}, validTestPreparator, http.StatusBadRequest},
		{"Empty fields as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", GenreData{Name: "", Description: "Test"}, validTestPreparator, http.StatusBadRequest},
		{"Valid as Guest", "", "", GenreData{Name: "Test", Description: "Test"}, validTestPreparator, http.StatusForbidden},
		{"Valid as User", os.Getenv("CLAIM_ROLE_USER"), "", GenreData{Name: "Test", Description: "Test"}, validTestPreparator, http.StatusForbidden},
		{"Valid as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", GenreData{Name: "Test", Description: "Test"}, validTestPreparator, http.StatusOK},
		{"Invalid chars as Guest", "", "", GenreData{Name: "Action123!", Description: "Test"}, validTestPreparator, http.StatusBadRequest},
		{"Invalid chars as User", os.Getenv("CLAIM_ROLE_USER"), "", GenreData{Name: "Action123!", Description: "Test"}, validTestPreparator, http.StatusBadRequest},
		{"Invalid chars as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", GenreData{Name: "Action123!", Description: "Test"}, validTestPreparator, http.StatusBadRequest},
		{"Whitespace name as Guest", "", "", GenreData{Name: "   ", Description: "Test"}, validTestPreparator, http.StatusBadRequest},
		{"Whitespace name as User", os.Getenv("CLAIM_ROLE_USER"), "", GenreData{Name: "   ", Description: "Test"}, validTestPreparator, http.StatusBadRequest},
		{"Whitespace name as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", GenreData{Name: "   ", Description: "Test"}, validTestPreparator, http.StatusBadRequest},
		{"64 chars name as Guest", "", "", GenreData{Name: strings.Repeat("a", 64), Description: "Test"}, validTestPreparator, http.StatusForbidden},
		{"64 chars name as User", os.Getenv("CLAIM_ROLE_USER"), "", GenreData{Name: strings.Repeat("a", 64), Description: "Test"}, validTestPreparator, http.StatusForbidden},
		{"64 chars name as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", GenreData{Name: strings.Repeat("a", 64), Description: "Test"}, validTestPreparator, http.StatusOK},
		{"65 chars name as Guest", "", "", GenreData{Name: strings.Repeat("a", 65), Description: "Test"}, validTestPreparator, http.StatusBadRequest},
		{"65 chars name as User", os.Getenv("CLAIM_ROLE_USER"), "", GenreData{Name: strings.Repeat("a", 65), Description: "Test"}, validTestPreparator, http.StatusBadRequest},
		{"65 chars name as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", GenreData{Name: strings.Repeat("a", 65), Description: "Test"}, validTestPreparator, http.StatusBadRequest},
		{"Hyphen name as Guest", "", "", GenreData{Name: "Sci-Fi", Description: "Test"}, validTestPreparator, http.StatusForbidden},
		{"Hyphen name as User", os.Getenv("CLAIM_ROLE_USER"), "", GenreData{Name: "Sci-Fi", Description: "Test"}, validTestPreparator, http.StatusForbidden},
		{"Hyphen name as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", GenreData{Name: "Sci-Fi", Description: "Test"}, validTestPreparator, http.StatusOK},
		{"Empty desc as Guest", "", "", GenreData{Name: "Test", Description: ""}, validTestPreparator, http.StatusBadRequest},
		{"Empty desc as User", os.Getenv("CLAIM_ROLE_USER"), "", GenreData{Name: "Test", Description: ""}, validTestPreparator, http.StatusBadRequest},
		{"Empty desc as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", GenreData{Name: "Test", Description: ""}, validTestPreparator, http.StatusBadRequest},
		{"Whitespace desc as Guest", "", "", GenreData{Name: "Test", Description: "   "}, validTestPreparator, http.StatusBadRequest},
		{"Whitespace desc as User", os.Getenv("CLAIM_ROLE_USER"), "", GenreData{Name: "Test", Description: "   "}, validTestPreparator, http.StatusBadRequest},
		{"Whitespace desc as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", GenreData{Name: "Test", Description: "   "}, validTestPreparator, http.StatusBadRequest},
		{"1000 chars desc as Guest", "", "", GenreData{Name: "Test", Description: strings.Repeat("a", 1000)}, validTestPreparator, http.StatusForbidden},
		{"1000 chars desc as User", os.Getenv("CLAIM_ROLE_USER"), "", GenreData{Name: "Test", Description: strings.Repeat("a", 1000)}, validTestPreparator, http.StatusForbidden},
		{"1000 chars desc as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", GenreData{Name: "Test", Description: strings.Repeat("a", 1000)}, validTestPreparator, http.StatusOK},
		{"1001 chars desc as Guest", "", "", GenreData{Name: "Test", Description: strings.Repeat("a", 1001)}, validTestPreparator, http.StatusBadRequest},
		{"1001 chars desc as User", os.Getenv("CLAIM_ROLE_USER"), "", GenreData{Name: "Test", Description: strings.Repeat("a", 1001)}, validTestPreparator, http.StatusBadRequest},
		{"1001 chars desc as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", GenreData{Name: "Test", Description: strings.Repeat("a", 1001)}, validTestPreparator, http.StatusBadRequest},
		{"DBError as Guest", "", "", GenreData{Name: "Test", Description: "Test"}, func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			TestAdminDB.Close()
			TestGuestDB.Close()
			TestUserDB.Close()
			return ts, GenresData[0].ID
		}, http.StatusInternalServerError},
		{"DBError as User", os.Getenv("CLAIM_ROLE_USER"), "", GenreData{Name: "Test", Description: "Test"}, func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			TestAdminDB.Close()
			TestGuestDB.Close()
			TestUserDB.Close()
			return ts, GenresData[0].ID
		}, http.StatusInternalServerError},
		{"DBError as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", GenreData{Name: "Test", Description: "Test"}, func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			TestAdminDB.Close()
			TestGuestDB.Close()
			TestUserDB.Close()
			return ts, GenresData[0].ID
		}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
			SeedUsers(TestAdminDB)
			defer func() {
				ts.Close()
				if strings.Split(tt.name, " ")[0] == "DBError" {
					InitTestDB()
				}
			}()

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
	tests := []struct {
		name           string
		role           string
		id             string
		setup          func(t *testing.T) (*httptest.Server, string)
		expectedStatus int
	}{
		{"Unknown ID as Guest", "", "", func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			return ts, uuid.New().String()
		}, http.StatusForbidden},
		{"Unknown ID as User", os.Getenv("CLAIM_ROLE_USER"), "", func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			return ts, uuid.New().String()
		}, http.StatusForbidden},
		{"Unknown ID as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			return ts, uuid.New().String()
		}, http.StatusNotFound},
		{"Invalid ID as Guest", "", "invalid-uuid", func(t *testing.T) (*httptest.Server, string) {
			return setupTestServer(), "invalid-uuid"
		}, http.StatusBadRequest},
		{"Invalid ID as User", os.Getenv("CLAIM_ROLE_USER"), "invalid-uuid", func(t *testing.T) (*httptest.Server, string) {
			return setupTestServer(), "invalid-uuid"
		}, http.StatusBadRequest},
		{"Invalid ID as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "invalid-uuid", func(t *testing.T) (*httptest.Server, string) {
			return setupTestServer(), "invalid-uuid"
		}, http.StatusBadRequest},
		{"Valid as Guest", "", "", func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			return ts, GenresData[0].ID
		}, http.StatusForbidden},
		{"Valid as User", os.Getenv("CLAIM_ROLE_USER"), "", func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			return ts, GenresData[0].ID
		}, http.StatusForbidden},
		{"Valid as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			return ts, GenresData[6].ID
		}, http.StatusNoContent},
		{"DBError as Guest", "", "", func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			TestAdminDB.Close()
			TestGuestDB.Close()
			TestUserDB.Close()
			return ts, GenresData[6].ID
		}, http.StatusInternalServerError},
		{"DBError as User", os.Getenv("CLAIM_ROLE_USER"), "", func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			TestAdminDB.Close()
			TestGuestDB.Close()
			TestUserDB.Close()
			return ts, GenresData[6].ID
		}, http.StatusInternalServerError},
		{"DBError as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			TestAdminDB.Close()
			TestGuestDB.Close()
			TestUserDB.Close()
			return ts, GenresData[6].ID
		}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
			SeedUsers(TestAdminDB)
			defer func() {
				ts.Close()
				if strings.Split(tt.name, " ")[0] == "DBError" {
					InitTestDB()
				}
			}()

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
