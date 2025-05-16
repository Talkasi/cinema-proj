package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGetMovieShows(t *testing.T) {
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
			ts := setupTestServer()
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
			"CLAIM_ROLE_USER",
			http.StatusNotFound,
		},
		{
			"Unknown ID as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			"CLAIM_ROLE_ADMIN",
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
	validMovieShow := MovieShowData{
		MovieID:   MoviesData[0].ID,
		HallID:    HallsData[1].ID,
		StartTime: time.Now().Add(124 * time.Hour),
		Language:  Russian,
	}

	invalidMovieShow := MovieShowData{
		MovieID:   "invalid",
		HallID:    "invalid",
		StartTime: time.Date(1890, 1, 1, 0, 0, 0, 0, time.UTC),
		Language:  "INVALID",
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
			"CLAIM_ROLE_USER",
			validMovieShow,
			nil,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			"CLAIM_ROLE_ADMIN",
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
			"Invalid data Guest",
			"",
			invalidMovieShow,
			nil,
			http.StatusBadRequest,
		},
		{
			"Invalid data User",
			"CLAIM_ROLE_USER",
			invalidMovieShow,
			nil,
			http.StatusBadRequest,
		},
		{
			"Invalid data Admin",
			"CLAIM_ROLE_ADMIN",
			invalidMovieShow,
			nil,
			http.StatusBadRequest,
		},
		{
			"Conflict Admin",
			"CLAIM_ROLE_ADMIN",
			validMovieShow,
			setupConflictTest,
			http.StatusConflict,
		},
		{
			"Invalid movie ID",
			"CLAIM_ROLE_ADMIN",
			MovieShowData{
				MovieID:   "invalid",
				HallID:    HallsData[0].ID,
				StartTime: time.Now().Add(24 * time.Hour),
				Language:  Russian,
			},
			func(t *testing.T) {
				SeedAll(TestAdminDB)
			},
			http.StatusBadRequest,
		},
		{
			"Invalid hall ID",
			"CLAIM_ROLE_ADMIN",
			MovieShowData{
				MovieID:   MoviesData[0].ID,
				HallID:    "invalid",
				StartTime: time.Now().Add(24 * time.Hour),
				Language:  Russian,
			},
			nil,
			http.StatusBadRequest,
		},
		{
			"Start time in past",
			"CLAIM_ROLE_ADMIN",
			MovieShowData{
				MovieID:   MoviesData[0].ID,
				HallID:    HallsData[0].ID,
				StartTime: time.Now().Add(-24 * time.Hour),
				Language:  Russian,
			},
			nil,
			http.StatusCreated,
		},
		{
			"Invalid language",
			"CLAIM_ROLE_ADMIN",
			MovieShowData{
				MovieID:   MoviesData[0].ID,
				HallID:    HallsData[0].ID,
				StartTime: time.Now().Add(24 * time.Hour),
				Language:  "INVALID",
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
			"CLAIM_ROLE_ADMIN",
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
			"CLAIM_ROLE_ADMIN",
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
			"CLAIM_ROLE_ADMIN",
			"",
			validUpdateData,
			setupExistingShow,
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
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
			"CLAIM_ROLE_ADMIN",
			"",
			setupExistingShow,
			http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
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
	defer ts.Close()

	// Create DB error situation
	TestAdminDB.Close()
	TestGuestDB.Close()
	TestUserDB.Close()

	req := createRequest(t, "POST", ts.URL+"/movie-shows",
		generateToken(t, "CLAIM_ROLE_ADMIN"),
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
			"CLAIM_ROLE_USER",
			"",
			"",
			setupWithData,
			http.StatusOK,
			1,
		},
		{
			"Valid request with custom hours",
			"CLAIM_ROLE_USER",
			"",
			"48",
			setupWithData,
			http.StatusOK,
			1,
		},
		{
			"Invalid movie ID format",
			"CLAIM_ROLE_USER",
			"invalid-uuid",
			"",
			setupWithData,
			http.StatusBadRequest,
			0,
		},
		{
			"Non-existent movie ID",
			"CLAIM_ROLE_USER",
			uuid.New().String(),
			"",
			setupWithData,
			http.StatusNotFound,
			0,
		},
		{
			"Invalid hours parameter",
			"CLAIM_ROLE_USER",
			"",
			"invalid",
			setupWithData,
			http.StatusBadRequest,
			0,
		},
		{
			"Negative hours parameter",
			"CLAIM_ROLE_USER",
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
			"CLAIM_ROLE_USER",
			time.Now().Add(48 * time.Hour).Format("2006-01-02"),
			setupWithData,
			http.StatusOK,
		},
		{
			"Invalid date format",
			"CLAIM_ROLE_USER",
			"2023\\01\\01",
			setupWithData,
			http.StatusBadRequest,
		},
		{
			"Date with no shows",
			"CLAIM_ROLE_USER",
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
			"CLAIM_ROLE_USER",
			"",
			setupWithData,
			http.StatusOK,
		},
		{
			"Custom valid hours",
			"CLAIM_ROLE_USER",
			"72",
			setupWithData,
			http.StatusOK,
		},
		{
			"Invalid hours format",
			"CLAIM_ROLE_USER",
			"invalid",
			setupWithData,
			http.StatusBadRequest,
		},
		{
			"Negative hours",
			"CLAIM_ROLE_USER",
			"-5",
			setupWithData,
			http.StatusBadRequest,
		},
		{
			"Zero hours",
			"CLAIM_ROLE_USER",
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
