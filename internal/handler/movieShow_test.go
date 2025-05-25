package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestGetMovieShows(t *testing.T) {
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

			req := createRequest(t, "GET", ts.URL+"/movie-shows", generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var shows []MovieShow
				parseResponseBody(t, resp, &shows)

				if len(shows) == 0 {
					t.Error("Expected non-empty movie shows list")
				}
			}
		})
	}
}

func TestGetMovieShowByID(t *testing.T) {
	setupValidIDTest := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, MovieShowsData[0].ID
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
			"Valid ID as Guest",
			setupValidIDTest,
			"",
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
			SeedUsers(TestAdminDB)
			defer ts.Close()

			req := createRequest(t, "GET", ts.URL+"/movie-shows/"+id, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var show MovieShow
				parseResponseBody(t, resp, &show)

				if show.ID != id {
					t.Errorf("Expected ID %v; got %v", id, show.ID)
				}
			}
		})
	}
}

func TestCreateMovieShow(t *testing.T) {
	validMovieShow := MovieShowAdmin{
		MovieID:   MoviesData[0].ID,
		HallID:    HallsData[1].ID,
		StartTime: time.Now().Add(124 * time.Hour),
		Language:  Russian,
		BasePrice: 300,
	}

	invalidMovieShow := MovieShowAdmin{
		MovieID:   "invalid",
		HallID:    "invalid",
		StartTime: time.Date(1890, 1, 1, 0, 0, 0, 0, time.UTC),
		Language:  "INVALID",
		BasePrice: 300,
	}

	setupConflictTest := func(t *testing.T) {
		_, err := TestAdminDB.Exec(context.Background(),
			"INSERT INTO movie_shows (movie_id, hall_id, start_time, language) VALUES ($1, $2, $3, $4)",
			validMovieShow.MovieID, validMovieShow.HallID, validMovieShow.StartTime, validMovieShow.Language)
		if err != nil {
			t.Fatalf("Failed to insert first movie show: %v", err)
		}
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
			validMovieShow,
			nil,
			http.StatusForbidden,
		},
		{
			"Forbidden User",
			os.Getenv("CLAIM_ROLE_USER"),
			validMovieShow,
			nil,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			validMovieShow,
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
			"Invalid data Guest",
			"",
			invalidMovieShow,
			nil,
			http.StatusBadRequest,
		},
		{
			"Invalid data User",
			os.Getenv("CLAIM_ROLE_USER"),
			invalidMovieShow,
			nil,
			http.StatusBadRequest,
		},
		{
			"Invalid data Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			invalidMovieShow,
			nil,
			http.StatusBadRequest,
		},
		{
			"Conflict Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			validMovieShow,
			setupConflictTest,
			http.StatusConflict,
		},
		{
			"Invalid movie ID",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			MovieShowAdmin{
				MovieID:   "invalid",
				HallID:    HallsData[0].ID,
				StartTime: time.Now().Add(24 * time.Hour),
				Language:  Russian,
				BasePrice: 300,
			},
			func(t *testing.T) {
				SeedAll(TestAdminDB)
			},
			http.StatusBadRequest,
		},
		{
			"Invalid hall ID",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			MovieShowAdmin{
				MovieID:   MoviesData[0].ID,
				HallID:    "invalid",
				StartTime: time.Now().Add(24 * time.Hour),
				Language:  Russian,
				BasePrice: 300,
			},
			nil,
			http.StatusBadRequest,
		},
		{
			"Start time in past",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			MovieShowAdmin{
				MovieID:   MoviesData[0].ID,
				HallID:    HallsData[0].ID,
				StartTime: time.Now().Add(-24 * time.Hour),
				Language:  Russian,
				BasePrice: 300,
			},
			nil,
			http.StatusCreated,
		},
		{
			"Invalid language",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			MovieShowAdmin{
				MovieID:   MoviesData[0].ID,
				HallID:    HallsData[0].ID,
				StartTime: time.Now().Add(24 * time.Hour),
				Language:  "INVALID",
				BasePrice: 300,
			},
			nil,
			http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			defer ts.Close()

			if tt.setup != nil {
				tt.setup(t)
			}

			req := createRequest(t, "POST", ts.URL+"/movie-shows", generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusCreated {
				var created string
				parseResponseBody(t, resp, &created)

				if created == "" {
					t.Error("Expected non-empty ID in response")
				}

				if _, err := uuid.Parse(created); err != nil {
					t.Error("Invalid UUID format in response")
				}
			}
		})
	}
}

func TestUpdateMovieShow(t *testing.T) {
	validUpdateData := MovieShowData{
		MovieID:   MoviesData[0].ID,
		HallID:    HallsData[0].ID,
		StartTime: time.Now().Add(48 * time.Hour),
		Language:  English,
	}

	invalidUpdateData := MovieShowData{
		MovieID:   "invalid",
		HallID:    "invalid",
		StartTime: time.Date(1890, 1, 1, 0, 0, 0, 0, time.UTC),
		Language:  "INVALID",
	}

	setupExistingShow := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, MovieShowsData[0].ID
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
			"Invalid data as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			invalidUpdateData,
			setupExistingShow,
			http.StatusBadRequest,
		},
		{
			"Forbidden Guest",
			"",
			"",
			validUpdateData,
			setupExistingShow,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			validUpdateData,
			setupExistingShow,
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
			SeedUsers(TestAdminDB)
			defer ts.Close()

			effectiveID := tt.id
			if effectiveID == "" {
				effectiveID = id
			}

			req := createRequest(t, "PUT", ts.URL+"/movie-shows/"+effectiveID, generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestDeleteMovieShow(t *testing.T) {
	setupExistingShow := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, MovieShowsData[3].ID
	}

	tests := []struct {
		name           string
		role           string
		id             string
		setup          func(t *testing.T) (*httptest.Server, string)
		expectedStatus int
	}{
		{
			"Invalid UUID as Guest",
			"",
			"invalid-uuid",
			func(t *testing.T) (*httptest.Server, string) {
				return setupTestServer(), ""
			},
			http.StatusBadRequest,
		},
		{
			"Forbidden as Guest",
			"",
			"",
			setupExistingShow,
			http.StatusForbidden,
		},
		{
			"Success as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			setupExistingShow,
			http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
			SeedUsers(TestAdminDB)
			defer ts.Close()

			effectiveID := tt.id
			if effectiveID == "" {
				effectiveID = id
			}

			req := createRequest(t, "DELETE", ts.URL+"/movie-shows/"+effectiveID, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestCreateMovieShowDBError(t *testing.T) {
	ts := setupTestServer()
	SeedUsers(TestAdminDB)
	defer ts.Close()

	// Create DB error situation
	TestAdminDB.Close()
	TestGuestDB.Close()
	TestUserDB.Close()

	req := createRequest(t, "POST", ts.URL+"/movie-shows",
		generateToken(t, os.Getenv("CLAIM_ROLE_ADMIN")),
		MovieShowData{
			MovieID:   uuid.New().String(),
			HallID:    uuid.New().String(),
			StartTime: time.Now().Add(24 * time.Hour),
			Language:  Russian,
		})
	resp := executeRequest(t, req, http.StatusInternalServerError)
	defer resp.Body.Close()

	if err := InitTestDB(); err != nil {
		log.Fatal("Failed to reconnect to DB: ", err)
	}
}

func TestMovieShowConflictTrigger(t *testing.T) {
	t.Run("No conflict - different halls", func(t *testing.T) {
		ts := setupTestServer()
		SeedAll(TestAdminDB)
		if err := ClearTable(TestAdminDB, "movie_shows"); err != nil {
			t.Fatalf("Failed to clear database")
		}
		defer ts.Close()

		hall1 := HallsData[0].ID
		hall2 := HallsData[1].ID
		movieID := MoviesData[2].ID

		startTime1 := time.Now().Add(24 * time.Hour)
		_, err := TestAdminDB.Exec(context.Background(),
			"INSERT INTO movie_shows (movie_id, hall_id, start_time, language) VALUES ($1, $2, $3, $4)",
			movieID, hall1, startTime1, Russian)
		if err != nil {
			t.Fatalf("Failed to insert first movie show: %v", err)
		}

		_, err = TestAdminDB.Exec(context.Background(),
			"INSERT INTO movie_shows (movie_id, hall_id, start_time, language) VALUES ($1, $2, $3, $4)",
			movieID, hall2, startTime1, Russian)
		if err != nil {
			t.Errorf("Expected no conflict for different halls, got error: %v", err)
		}
	})

	t.Run("No conflict - same hall, different times", func(t *testing.T) {
		ts := setupTestServer()
		SeedAll(TestAdminDB)
		if err := ClearTable(TestAdminDB, "movie_shows"); err != nil {
			t.Fatalf("Failed to clear database")
		}
		defer ts.Close()

		hallID := HallsData[0].ID
		movieID := MoviesData[1].ID

		startTime1 := time.Now().Add(24 * time.Hour)
		_, err := TestAdminDB.Exec(context.Background(),
			"INSERT INTO movie_shows (movie_id, hall_id, start_time, language) VALUES ($1, $2, $3, $4)",
			movieID, hallID, startTime1, Russian)
		if err != nil {
			t.Fatalf("Failed to insert first movie show: %v", err)
		}

		startTime2 := startTime1.Add(3*time.Hour + 30*time.Minute)
		_, err = TestAdminDB.Exec(context.Background(),
			"INSERT INTO movie_shows (movie_id, hall_id, start_time, language) VALUES ($1, $2, $3, $4)",
			movieID, hallID, startTime2, Russian)
		if err != nil {
			t.Errorf("Expected no conflict for sufficient time gap, got error: %v", err)
		}
	})

	t.Run("Conflict - same hall, overlapping times", func(t *testing.T) {
		ts := setupTestServer()
		SeedAll(TestAdminDB)
		if err := ClearTable(TestAdminDB, "movie_shows"); err != nil {
			t.Fatalf("Failed to clear database")
		}
		defer ts.Close()

		hallID := HallsData[0].ID
		movieID := MoviesData[1].ID

		startTime1 := time.Now().Add(24 * time.Hour)
		_, err := TestAdminDB.Exec(context.Background(),
			"INSERT INTO movie_shows (movie_id, hall_id, start_time, language) VALUES ($1, $2, $3, $4)",
			movieID, hallID, startTime1, Russian)
		if err != nil {
			t.Fatalf("Failed to insert first movie show: %v", err)
		}

		startTime2 := startTime1.Add(1 * time.Hour)
		_, err = TestAdminDB.Exec(context.Background(),
			"INSERT INTO movie_shows (movie_id, hall_id, start_time, language) VALUES ($1, $2, $3, $4)",
			movieID, hallID, startTime2, Russian)

		if err == nil {
			t.Error("Expected conflict error, got no error")
		} else {
			if !strings.Contains(err.Error(), "Невозможно запланировать показ") {
				t.Errorf("Expected conflict error message, got: %v", err)
			}
		}
	})

	t.Run("Conflict - cleaning time between shows", func(t *testing.T) {
		ts := setupTestServer()
		SeedAll(TestAdminDB)
		if err := ClearTable(TestAdminDB, "movie_shows"); err != nil {
			t.Fatalf("Failed to clear database")
		}
		defer ts.Close()

		hallID := HallsData[0].ID
		movieID := MoviesData[1].ID

		startTime1 := time.Now().Add(24 * time.Hour)
		_, err := TestAdminDB.Exec(context.Background(),
			"INSERT INTO movie_shows (movie_id, hall_id, start_time, language) VALUES ($1, $2, $3, $4)",
			movieID, hallID, startTime1, Russian)
		if err != nil {
			t.Fatalf("Failed to insert first movie show: %v", err)
		}

		startTime2 := startTime1.Add(1 * time.Hour)
		_, err = TestAdminDB.Exec(context.Background(),
			"INSERT INTO movie_shows (movie_id, hall_id, start_time, language) VALUES ($1, $2, $3, $4)",
			movieID, hallID, startTime2, Russian)

		if err == nil {
			t.Error("Expected conflict error for cleaning time, got no error")
		} else {
			if !strings.Contains(err.Error(), "будет проводиться уборка") {
				t.Errorf("Expected cleaning time conflict error, got: %v", err)
			}
		}
	})

	t.Run("No conflict - exact cleaning time between shows", func(t *testing.T) {
		ts := setupTestServer()
		SeedAll(TestAdminDB)
		if err := ClearTable(TestAdminDB, "movie_shows"); err != nil {
			t.Fatalf("Failed to clear database")
		}
		defer ts.Close()

		hallID := HallsData[0].ID
		movieID := MoviesData[1].ID
		movieDuration := MoviesData[1].Duration

		startTime1 := time.Now().Add(24 * time.Hour)
		_, err := TestAdminDB.Exec(context.Background(),
			"INSERT INTO movie_shows (movie_id, hall_id, start_time, language) VALUES ($1, $2, $3, $4)",
			movieID, hallID, startTime1, Russian)
		if err != nil {
			t.Fatalf("Failed to insert first movie show: %v", err)
		}

		parsedDuration, _ := time.Parse("15:04:05", movieDuration)
		startTime2 := startTime1.Add(parsedDuration.Sub(time.Time{}) + 10*time.Minute)
		_, err = TestAdminDB.Exec(context.Background(),
			"INSERT INTO movie_shows (movie_id, hall_id, start_time, language) VALUES ($1, $2, $3, $4)",
			movieID, hallID, startTime2, Russian)

		if err != nil {
			t.Errorf("Expected no conflict with exact cleaning time gap, got error: %v", err)
		}
	})
}

func TestGetShowsByMovie(t *testing.T) {
	setupWithData := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, MoviesData[0].ID
	}

	tests := []struct {
		name           string
		role           string
		movieID        string
		hours          string
		setup          func(t *testing.T) (*httptest.Server, string)
		expectedStatus int
		expectedCount  int
	}{
		{
			"Valid request with default hours",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			"",
			setupWithData,
			http.StatusOK,
			1,
		},
		{
			"Valid request with custom hours",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			"48",
			setupWithData,
			http.StatusOK,
			1,
		},
		{
			"Invalid movie ID format",
			os.Getenv("CLAIM_ROLE_USER"),
			"invalid-uuid",
			"",
			setupWithData,
			http.StatusBadRequest,
			0,
		},
		{
			"Non-existent movie ID",
			os.Getenv("CLAIM_ROLE_USER"),
			uuid.New().String(),
			"",
			setupWithData,
			http.StatusNotFound,
			0,
		},
		{
			"Invalid hours parameter",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			"invalid",
			setupWithData,
			http.StatusBadRequest,
			0,
		},
		{
			"Negative hours parameter",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			"-10",
			setupWithData,
			http.StatusBadRequest,
			0,
		},
		{
			"Guest access allowed",
			"",
			"",
			"",
			setupWithData,
			http.StatusOK,
			1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, movieID := tt.setup(t)
			SeedUsers(TestAdminDB)
			defer ts.Close()

			// Use provided ID or fallback to id from setup
			effectiveID := tt.movieID
			if effectiveID == "" {
				effectiveID = movieID
			}

			url := fmt.Sprintf("%s/movies/%s/shows", ts.URL, effectiveID)
			if tt.hours != "" {
				url += "?hours=" + tt.hours
			}

			req := createRequest(t, "GET", url, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var shows []MovieShow
				if err := json.NewDecoder(resp.Body).Decode(&shows); err != nil {
					t.Fatalf("Could not decode response: %v", err)
				}

				if len(shows) != tt.expectedCount {
					t.Errorf("Expected %d shows, got %d", tt.expectedCount, len(shows))
				}
			}
		})
	}
}

func TestGetShowsByDate(t *testing.T) {
	setupWithData := func(t *testing.T) *httptest.Server {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts
	}

	tests := []struct {
		name           string
		role           string
		date           string
		setup          func(t *testing.T) *httptest.Server
		expectedStatus int
	}{
		{
			"Valid date",
			os.Getenv("CLAIM_ROLE_USER"),
			time.Now().Add(48 * time.Hour).Format("2006-01-02"),
			setupWithData,
			http.StatusOK,
		},
		{
			"Invalid date format",
			os.Getenv("CLAIM_ROLE_USER"),
			"2023\\01\\01",
			setupWithData,
			http.StatusBadRequest,
		},
		{
			"Date with no shows",
			os.Getenv("CLAIM_ROLE_USER"),
			"1900-01-01",
			setupWithData,
			http.StatusNotFound,
		},
		{
			"Guest access allowed",
			"",
			time.Now().Add(48 * time.Hour).Format("2006-01-02"),
			setupWithData,
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := tt.setup(t)
			SeedUsers(TestAdminDB)
			defer ts.Close()

			req := createRequest(t, "GET", ts.URL+"/movie-shows/by-date/"+tt.date, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var shows []MovieShow
				if err := json.NewDecoder(resp.Body).Decode(&shows); err != nil {
					t.Fatalf("Could not decode response: %v", err)
				}

				if len(shows) == 0 {
					t.Error("Expected at least one show, got zero")
				}
			}
		})
	}
}

func TestGetUpcomingShows(t *testing.T) {
	setupWithData := func(t *testing.T) *httptest.Server {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts
	}

	tests := []struct {
		name           string
		role           string
		hours          string
		setup          func(t *testing.T) *httptest.Server
		expectedStatus int
	}{
		{
			"Default hours",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			setupWithData,
			http.StatusOK,
		},
		{
			"Custom valid hours",
			os.Getenv("CLAIM_ROLE_USER"),
			"72",
			setupWithData,
			http.StatusOK,
		},
		{
			"Invalid hours format",
			os.Getenv("CLAIM_ROLE_USER"),
			"invalid",
			setupWithData,
			http.StatusBadRequest,
		},
		{
			"Negative hours",
			os.Getenv("CLAIM_ROLE_USER"),
			"-5",
			setupWithData,
			http.StatusBadRequest,
		},
		{
			"Zero hours",
			os.Getenv("CLAIM_ROLE_USER"),
			"0",
			setupWithData,
			http.StatusBadRequest,
		},
		{
			"Guest access allowed",
			"",
			"",
			setupWithData,
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := tt.setup(t)
			SeedUsers(TestAdminDB)
			defer ts.Close()

			url := ts.URL + "/movie-shows/upcoming"
			if tt.hours != "" {
				url += "?hours=" + tt.hours
			}

			req := createRequest(t, "GET", url, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var shows []MovieShow
				if err := json.NewDecoder(resp.Body).Decode(&shows); err != nil {
					t.Fatalf("Could not decode response: %v", err)
				}

				if len(shows) == 0 {
					t.Error("Expected at least one show, got zero")
				}
			}
		})
	}
}

func setupTestDBmovie(b *testing.B, withIndex bool, n int) *pgxpool.Pool {
	db, err := pgxpool.New(context.Background(), fmt.Sprintf("user=%s dbname=%s password=%s sslmode=disable",
		os.Getenv("TEST_ADMIN_USER"),
		os.Getenv("TEST_DB_NAME"),
		os.Getenv("TEST_ADMIN_PASSWORD")))
	if err != nil {
		b.Fatalf("Failed to connect to database: %v", err)
	}

	_, err = db.Exec(context.Background(), "DROP INDEX IF EXISTS idx_movie_time;")
	if err != nil {
		b.Error(err)
	}

	if withIndex {
		_, err = db.Exec(context.Background(), "CREATE INDEX IF NOT EXISTS idx_movie_time ON movie_shows(start_time)")
		if err != nil {
			b.Fatalf("Failed to create index: %v", err)
		}
	}

	_, err = db.Exec(context.Background(), "TRUNCATE TABLE movie_shows CASCADE")
	if err != nil {
		b.Fatalf("Failed to truncate table: %v", err)
	}

	SeedAll(db)
	generateMovieShows(db, n)

	return db
}

func BenchmarkDBWithConcurrencyNEWServerINDmovie(b *testing.B) {
	db := setupTestDBmovie(b, false, 1_000_000)
	defer db.Close()

	benchmarks := []struct {
		name    string
		workers int
	}{
		{"1_workers", 1},
		{"2_workers", 2},
		{"4_workers", 4},
		{"6_workers", 6},
		{"10000_workers", 10000},
		{"30000_workers", 30000},
		{"50000_workers", 50000},
		{"70000_workers", 70000},
		{"100000_workers", 100000},
	}

	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			b.SetParallelism(bb.workers)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					req, err := http.NewRequest(
						"GET",
						ts.URL+"/movie-shows/upcoming",
						http.NoBody,
					)
					if err != nil {
						b.Error("Create request error:", err)
						continue
					}
					req.Header.Set("Content-Type", "application/json")

					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						b.Error("Request failed:", err)
						continue
					}

					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						b.Errorf("Unexpected status: %d", resp.StatusCode)
					}
				}
			})
		})
	}
}

func BenchmarkDBWithConcurrencyNEWServerNINDmovie(b *testing.B) {
	db := setupTestDBmovie(b, false, 1_000_000)
	defer db.Close()

	benchmarks := []struct {
		name    string
		workers int
	}{
		{"1_workers", 1},
		{"2_workers", 2},
		{"4_workers", 4},
		{"6_workers", 6},
	}

	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			b.SetParallelism(bb.workers)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					req, err := http.NewRequest(
						"GET",
						ts.URL+"/movie-shows/upcoming",
						http.NoBody,
					)
					if err != nil {
						b.Error("Create request error:", err)
						continue
					}
					req.Header.Set("Content-Type", "application/json")

					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						b.Error("Request failed:", err)
						continue
					}

					// io.Copy(io.Discard, resp.Body)
					resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						b.Errorf("Unexpected status: %d", resp.StatusCode)
					}
				}
			})
		})
	}
}

func BenchmarkDBWithConcurrencyNEW_nind_movie(b *testing.B) {
	db := setupTestDBmovie(b, false, 1_000_000)
	defer db.Close()

	benchmarks := []struct {
		name    string
		workers int
	}{
		{"1_workers", 1},
		{"2_workers", 2},
		{"4_workers", 4},
		{"6_workers", 6},
	}

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			b.SetParallelism(bb.workers)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				hours := rand.Intn(24 * 4)
				now := time.Now()
				endTime := now.Add(time.Duration(hours) * time.Hour)

				for pb.Next() {
					rows, err := db.Query(context.Background(), `
						SELECT id, movie_id, hall_id, start_time, language 
						FROM movie_shows 
						WHERE start_time BETWEEN $1 AND $2
						ORDER BY start_time`, now, endTime)
					rows.Close()

					if err != nil {
						b.Error("Query failed:", err)
						continue
					}
				}
			})
		})
	}
}

func BenchmarkDBWithConcurrencyNEW_ind_movie(b *testing.B) {
	db := setupTestDBmovie(b, false, 1_000_000)
	defer db.Close()
	benchmarks := []struct {
		name    string
		workers int
	}{
		{"10000_workers", 10000},
		{"20000_workers", 20000},
		{"30000_workers", 30000},
		{"40000_workers", 40000},
		{"50000_workers", 50000},
		{"60000_workers", 60000},
		{"70000_workers", 70000},
		{"80000_workers", 80000},
		{"90000_workers", 90000},
		{"100000_workers", 100000},
		// {"200000_workers", 200000},
		// {"300000_workers", 300000},
		// {"400000_workers", 400000},
		// {"500000_workers", 500000},
		// {"600000_workers", 600000},
		// {"700000_workers", 700000},
		// {"800000_workers", 800000},
		// {"900000_workers", 900000},
		// {"1000000_workers", 1000000},
	}

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			b.SetParallelism(bb.workers)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				hours := rand.Intn(24 * 4)
				now := time.Now()
				endTime := now.Add(time.Duration(hours) * time.Hour)

				for pb.Next() {
					rows, err := db.Query(context.Background(), `
						SELECT id, movie_id, hall_id, start_time, language 
						FROM movie_shows 
						WHERE start_time BETWEEN $1 AND $2
						ORDER BY start_time`, now, endTime)

					// println("HERE")

					rows.Close()

					if err != nil {
						b.Error("Query failed:", err)
						continue
					}
				}
			})
		})
	}
}

func Benchmark_movie_ni(b *testing.B) {
	db := setupTestDBmovie(b, false, 1_000_000)
	defer db.Close()
	benchmarks := []struct {
		name    string
		workers int
	}{
		{"1_workers", 1},
		{"2_workers", 2},
		{"3_workers", 3},
		{"4_workers", 4},
		{"5_workers", 5},
		{"6_workers", 6},
		{"7_workers", 7},
		{"8_workers", 8},
		{"9_workers", 9},
		{"10_workers", 10},
		{"11_workers", 11},
		{"12_workers", 12},

		{"10000_workers", 10000},
		{"20000_workers", 20000},
		{"30000_workers", 30000},
		{"40000_workers", 40000},
		{"50000_workers", 50000},
		{"60000_workers", 60000},
		{"70000_workers", 70000},
		{"80000_workers", 80000},
		{"90000_workers", 90000},
		{"100000_workers", 100000},
	}

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			b.SetParallelism(bb.workers)
			var totalLatency time.Duration
			var queryCount int64

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				hours := rand.Intn(24 * 4)
				now := time.Now()
				endTime := now.Add(time.Duration(hours) * time.Hour)

				for pb.Next() {
					start := time.Now()
					rows, err := db.Query(context.Background(), `
					SELECT id, movie_id, hall_id, start_time, language 
					FROM movie_shows 
					WHERE start_time BETWEEN $1 AND $2
					ORDER BY start_time`, now, endTime)
					if err != nil {
						b.Error("Query failed:", err)
						continue
					}

					var count int
					for rows.Next() {
						count++
					}
					if err := rows.Err(); err != nil {
						b.Error("Rows error:", err)
					}
					rows.Close()

					latency := time.Since(start)
					atomicAddDuration(&totalLatency, latency)
					atomic.AddInt64(&queryCount, 1)
				}

				// for pb.Next() {
				// 	start := time.Now()

				// 	rows, err := db.Query(context.Background(), `
				//         SELECT id, movie_id, hall_id, start_time, language
				//         FROM movie_shows
				//         WHERE start_time BETWEEN $1 AND $2
				//         ORDER BY start_time`, now, endTime)

				// 	latency := time.Since(start)
				// 	atomic.AddInt64(&queryCount, 1)
				// 	atomicAddDuration(&totalLatency, latency)

				// 	if err != nil {
				// 		b.Error("Query failed:", err)
				// 		continue
				// 	}
				// 	rows.Close()
				// }
			})

			avgLatency := time.Duration(0)
			if queryCount > 0 {
				avgLatency = totalLatency / time.Duration(queryCount)
			}
			b.ReportMetric(float64(avgLatency.Nanoseconds())/1e6, "avg_latency_ms")
			b.ReportMetric(float64(queryCount)/b.Elapsed().Seconds(), "queries_per_sec")
		})
	}
}

func Benchmark_movie_i(b *testing.B) {
	db := setupTestDBmovie(b, true, 1_000_000)
	defer db.Close()
	benchmarks := []struct {
		name    string
		workers int
	}{
		{"1_workers", 1},
		{"2_workers", 2},
		{"3_workers", 3},
		{"4_workers", 4},
		{"5_workers", 5},
		{"6_workers", 6},
		{"7_workers", 7},
		{"8_workers", 8},
		{"9_workers", 9},
		{"10_workers", 10},
		{"11_workers", 11},
		{"12_workers", 12},

		{"10000_workers", 10000},
		{"20000_workers", 20000},
		{"30000_workers", 30000},
		{"40000_workers", 40000},
		{"50000_workers", 50000},
		{"60000_workers", 60000},
		{"70000_workers", 70000},
		{"80000_workers", 80000},
		{"90000_workers", 90000},
		{"100000_workers", 100000},
	}

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			b.SetParallelism(bb.workers)
			var totalLatency time.Duration
			var queryCount int64

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				hours := rand.Intn(24 * 4)
				now := time.Now()
				endTime := now.Add(time.Duration(hours) * time.Hour)

				for pb.Next() {
					start := time.Now()
					rows, err := db.Query(context.Background(), `
					SELECT id, movie_id, hall_id, start_time, language 
					FROM movie_shows 
					WHERE start_time BETWEEN $1 AND $2
					ORDER BY start_time`, now, endTime)
					if err != nil {
						b.Error("Query failed:", err)
						continue
					}

					var count int
					for rows.Next() {
						count++
					}
					if err := rows.Err(); err != nil {
						b.Error("Rows error:", err)
					}
					rows.Close()

					latency := time.Since(start)
					atomicAddDuration(&totalLatency, latency)
					atomic.AddInt64(&queryCount, 1)
				}

				// for pb.Next() {
				// 	start := time.Now()

				// 	rows, err := db.Query(context.Background(), `
				//         SELECT id, movie_id, hall_id, start_time, language
				//         FROM movie_shows
				//         WHERE start_time BETWEEN $1 AND $2
				//         ORDER BY start_time`, now, endTime)

				// 	latency := time.Since(start)
				// 	atomic.AddInt64(&queryCount, 1)
				// 	atomicAddDuration(&totalLatency, latency)

				// 	if err != nil {
				// 		b.Error("Query failed:", err)
				// 		continue
				// 	}
				// 	rows.Close()
				// }
			})

			avgLatency := time.Duration(0)
			if queryCount > 0 {
				avgLatency = totalLatency / time.Duration(queryCount)
			}
			b.ReportMetric(float64(avgLatency.Nanoseconds())/1e6, "avg_latency_ms")
			b.ReportMetric(float64(queryCount)/b.Elapsed().Seconds(), "queries_per_sec")
		})
	}
}

func atomicAddDuration(addr *time.Duration, delta time.Duration) {
	for {
		old := atomic.LoadInt64((*int64)(unsafe.Pointer(addr)))
		new := old + int64(delta)
		if atomic.CompareAndSwapInt64((*int64)(unsafe.Pointer(addr)), old, new) {
			return
		}
	}
}

func generateMovieShows(db *pgxpool.Pool, n int) error {
	ctx := context.Background()

	movies, err := getMovies(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to get movies: %v", err)
	}

	halls, err := getHalls(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to get halls: %v", err)
	}

	languages := []string{"English", "Spanish", "French", "German", "Italian", "Русский"}

	startTime := time.Now()

	println(n)

	_, err = db.Exec(ctx, "ALTER TABLE movie_shows DISABLE TRIGGER check_movie_show_on_insert;")

	batchSize := 1000

	rng := rand.New(rand.NewSource(6585))

	for i := 0; i < n; i += batchSize {
		batch := make([]string, 0, batchSize)
		currentBatchSize := min(batchSize, n-i)

		for j := 0; j < currentBatchSize; j++ {
			movie := movies[rng.Intn(len(movies))]
			hall := halls[rng.Intn(len(halls))]
			lang := languages[rng.Intn(len(languages))]

			// timeOffset := time.Duration(rng.Intn(60*24*14)) * time.Minute
			showTime := startTime

			batch = append(batch, fmt.Sprintf(
				"(uuid_generate_v4(), '%s', '%s', '%s', '%s')",
				movie.ID, hall.ID, showTime.Format(time.RFC3339), lang,
			))
		}

		_, err = db.Exec(ctx,
			"INSERT INTO movie_shows (id, movie_id, hall_id, start_time, language) VALUES "+
				strings.Join(batch, ","))
		if err != nil {
			println(err.Error())
			return fmt.Errorf("failed to insert batch %d: %v", i/batchSize, err)
		}

		startTime = startTime.Add(time.Hour / 3)

		if (i/batchSize)%100 == 0 {
			log.Printf("Inserted %d of %d records", i, n)
		}
	}

	// for i := 0; i < n; i += 1 {
	// 	movie := movies[rand.Intn(len(movies))]
	// 	hall := halls[rand.Intn(len(halls))]
	// 	lang := languages[rand.Intn(len(languages))]

	// 	timeOffset := time.Duration(rand.Intn(60*24*14)) * time.Minute
	// 	showTime := startTime.Add(timeOffset)

	// 	// println(i)
	// 	_, err = db.Exec(ctx,
	// 		fmt.Sprintf("INSERT INTO movie_shows (id, movie_id, hall_id, start_time, language) VALUES ('%s', '%s', '%s', '%s', '%s')",
	// 			uuid.New(), movie.ID, hall.ID, showTime.Format(time.RFC3339), lang))
	// 	if err != nil {
	// 		println(err.Error())
	// 		return fmt.Errorf("failed to insert %d: %v", i, err)
	// 	}

	// 	startTime = startTime.Add(1 * time.Hour)
	// }

	_, err = db.Exec(ctx, "ALTER TABLE movie_shows ENABLE TRIGGER check_movie_show_on_insert;")
	println("DONE")

	return nil
}

func getMovies(ctx context.Context, db *pgxpool.Pool) ([]struct{ ID string }, error) {
	rows, err := db.Query(ctx, "SELECT id FROM movies")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []struct{ ID string }
	for rows.Next() {
		var m struct{ ID string }
		if err := rows.Scan(&m.ID); err != nil {
			return nil, err
		}
		movies = append(movies, m)
	}
	return movies, nil
}

func getHalls(ctx context.Context, db *pgxpool.Pool) ([]struct{ ID string }, error) {
	rows, err := db.Query(ctx, "SELECT id FROM halls")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var halls []struct{ ID string }
	for rows.Next() {
		var h struct{ ID string }
		if err := rows.Scan(&h.ID); err != nil {
			return nil, err
		}
		halls = append(halls, h)
	}
	return halls, nil
}
