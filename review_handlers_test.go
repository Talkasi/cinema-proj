package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestGetReviews(t *testing.T) {
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

			req := createRequest(t, "GET", ts.URL+"/reviews", generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var reviews []Review
				parseResponseBody(t, resp, &reviews)

				if len(reviews) == 0 {
					t.Error("Expected non-empty reviews list")
				}
			}
		})
	}
}

func TestGetReviewByID(t *testing.T) {
	setupValidIDTest := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, ReviewsData[0].ID
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
			"Invalid ID as User",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, "invalid-id"
			},
			"CLAIM_ROLE_USER",
			http.StatusBadRequest,
		},
		{
			"Invalid ID as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
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

			req := createRequest(t, "GET", ts.URL+"/reviews/"+id, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var review Review
				parseResponseBody(t, resp, &review)

				if review.ID != id {
					t.Errorf("Expected ID %v; got %v", id, review.ID)
				}
			}
		})
	}
}

func TestCreateReview(t *testing.T) {
	validReview := ReviewData{
		UserID:  UsersData[0].ID,
		MovieID: MoviesData[1].ID,
		Rating:  8,
		Comment: "Great movie!",
	}

	invalidReview := ReviewData{
		UserID:  "invalid",
		MovieID: MoviesData[1].ID,
		Rating:  8,
		Comment: "Great movie!",
	}

	tests := []struct {
		name           string
		role           string
		body           interface{}
		setup          func(t *testing.T)
		expectedStatus int
	}{
		{
			"Success User",
			"CLAIM_ROLE_USER",
			validReview,
			nil,
			http.StatusCreated,
		},
		{
			"Success Admin",
			"CLAIM_ROLE_ADMIN",
			validReview,
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
			"Invalid UserID",
			"CLAIM_ROLE_USER",
			invalidReview,
			nil,
			http.StatusBadRequest,
		},
		{
			"Invalid MovieID",
			"CLAIM_ROLE_USER",
			ReviewData{
				UserID:  UsersData[0].ID,
				MovieID: "invalid",
				Rating:  8,
				Comment: "Great movie!",
			},
			nil,
			http.StatusBadRequest,
		},
		{
			"Rating too low (0)",
			"CLAIM_ROLE_USER",
			ReviewData{
				UserID:  UsersData[0].ID,
				MovieID: MoviesData[1].ID,
				Rating:  0,
				Comment: "Great movie!",
			},
			nil,
			http.StatusBadRequest,
		},
		{
			"Rating too high (11)",
			"CLAIM_ROLE_USER",
			ReviewData{
				UserID:  UsersData[0].ID,
				MovieID: MoviesData[1].ID,
				Rating:  11,
				Comment: "Great movie!",
			},
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty comment",
			"CLAIM_ROLE_USER",
			ReviewData{
				UserID:  UsersData[0].ID,
				MovieID: MoviesData[1].ID,
				Rating:  8,
				Comment: "",
			},
			nil,
			http.StatusBadRequest,
		},
		{
			"Comment with only spaces",
			"CLAIM_ROLE_USER",
			ReviewData{
				UserID:  UsersData[0].ID,
				MovieID: MoviesData[1].ID,
				Rating:  8,
				Comment: "   ",
			},
			nil,
			http.StatusBadRequest,
		},
		{
			"Comment exactly 2000 chars",
			"CLAIM_ROLE_USER",
			ReviewData{
				UserID:  UsersData[0].ID,
				MovieID: MoviesData[1].ID,
				Rating:  8,
				Comment: strings.Repeat("a", 2000),
			},
			nil,
			http.StatusCreated,
		},
		{
			"Comment too long (2001 chars)",
			"CLAIM_ROLE_USER",
			ReviewData{
				UserID:  UsersData[0].ID,
				MovieID: MoviesData[1].ID,
				Rating:  8,
				Comment: strings.Repeat("a", 2001),
			},
			nil,
			http.StatusBadRequest,
		},
		{
			"Duplicate review",
			"CLAIM_ROLE_USER",
			validReview,
			func(t *testing.T) {
				_, err := TestAdminDB.Exec(context.Background(),
					"INSERT INTO reviews (user_id, movie_id, rating, review_comment) VALUES ($1, $2, $3, $4)",
					validReview.UserID, validReview.MovieID, validReview.Rating, validReview.Comment)
				if err != nil {
					t.Fatalf("Failed to insert into test database: %v", err)
				}
			},
			http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
			defer ts.Close()
			SeedAll(TestAdminDB)

			if tt.setup != nil {
				tt.setup(t)
			}

			req := createRequest(t, "POST", ts.URL+"/reviews", generateToken(t, tt.role), tt.body)
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

func TestUpdateReview(t *testing.T) {
	validUpdateData := ReviewData{
		UserID:  UsersData[0].ID,
		MovieID: MoviesData[1].ID,
		Rating:  9,
		Comment: "Updated comment",
	}

	invalidUpdateData := ReviewData{
		UserID:  "invalid",
		MovieID: MoviesData[1].ID,
		Rating:  9,
		Comment: "Updated comment",
	}

	setupExistingReview := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, ReviewsData[0].ID
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
			"Invalid UUID as User",
			"CLAIM_ROLE_USER",
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
			"CLAIM_ROLE_ADMIN",
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
			"Unknown UUID as User",
			"CLAIM_ROLE_USER",
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
			"Invalid JSON as User",
			"CLAIM_ROLE_USER",
			"",
			"invalid-json",
			setupExistingReview,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			"invalid-json",
			setupExistingReview,
			http.StatusBadRequest,
		},
		{
			"Invalid UserID in JSON",
			"CLAIM_ROLE_USER",
			"",
			invalidUpdateData,
			setupExistingReview,
			http.StatusBadRequest,
		},
		{
			"Invalid MovieID in JSON",
			"CLAIM_ROLE_USER",
			"",
			ReviewData{
				UserID:  UsersData[0].ID,
				MovieID: "invalid",
				Rating:  9,
				Comment: "Updated comment",
			},
			setupExistingReview,
			http.StatusBadRequest,
		},
		{
			"Rating too low (0)",
			"CLAIM_ROLE_USER",
			"",
			ReviewData{
				UserID:  UsersData[0].ID,
				MovieID: MoviesData[1].ID,
				Rating:  0,
				Comment: "Updated comment",
			},
			setupExistingReview,
			http.StatusBadRequest,
		},
		{
			"Rating too high (11)",
			"CLAIM_ROLE_USER",
			"",
			ReviewData{
				UserID:  UsersData[0].ID,
				MovieID: MoviesData[1].ID,
				Rating:  11,
				Comment: "Updated comment",
			},
			setupExistingReview,
			http.StatusBadRequest,
		},
		{
			"Empty comment",
			"CLAIM_ROLE_USER",
			"",
			ReviewData{
				UserID:  UsersData[0].ID,
				MovieID: MoviesData[1].ID,
				Rating:  9,
				Comment: "",
			},
			setupExistingReview,
			http.StatusBadRequest,
		},
		{
			"Comment with only spaces",
			"CLAIM_ROLE_USER",
			"",
			ReviewData{
				UserID:  UsersData[0].ID,
				MovieID: MoviesData[1].ID,
				Rating:  9,
				Comment: "   ",
			},
			setupExistingReview,
			http.StatusBadRequest,
		},
		{
			"Comment exactly 2000 chars",
			"CLAIM_ROLE_USER",
			"",
			ReviewData{
				UserID:  UsersData[0].ID,
				MovieID: MoviesData[1].ID,
				Rating:  9,
				Comment: strings.Repeat("a", 2000),
			},
			setupExistingReview,
			http.StatusOK,
		},
		{
			"Comment too long (2001 chars)",
			"CLAIM_ROLE_USER",
			"",
			ReviewData{
				UserID:  UsersData[0].ID,
				MovieID: MoviesData[1].ID,
				Rating:  9,
				Comment: strings.Repeat("a", 2001),
			},
			setupExistingReview,
			http.StatusBadRequest,
		},
		{
			"Success User",
			"CLAIM_ROLE_USER",
			"",
			validUpdateData,
			setupExistingReview,
			http.StatusOK,
		},
		{
			"Success Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			validUpdateData,
			setupExistingReview,
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
			SeedAll(TestAdminDB)
			defer ts.Close()

			// Use provided ID or fallback to id from setup
			effectiveID := tt.id
			if effectiveID == "" {
				effectiveID = id
			}

			req := createRequest(t, "PUT", ts.URL+"/reviews/"+effectiveID, generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestDeleteReview(t *testing.T) {
	// Setup function for tests needing existing review
	setupExistingReview := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, ReviewsData[0].ID
	}

	tests := []struct {
		name           string
		role           string
		id             string
		setup          func(t *testing.T) (*httptest.Server, string)
		expectedStatus int
	}{
		{
			"Not Found as User",
			"CLAIM_ROLE_USER",
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusNotFound,
		},
		{
			"Not Found as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusNotFound,
		},
		{
			"Invalid UUID as User",
			"CLAIM_ROLE_USER",
			"invalid-uuid",
			func(t *testing.T) (*httptest.Server, string) {
				return setupTestServer(), "invalid-uuid"
			},
			http.StatusBadRequest,
		},
		{
			"Invalid UUID as Admin",
			"CLAIM_ROLE_ADMIN",
			"invalid-uuid",
			func(t *testing.T) (*httptest.Server, string) {
				return setupTestServer(), "invalid-uuid"
			},
			http.StatusBadRequest,
		},
		{
			"Success as User",
			"CLAIM_ROLE_USER",
			"",
			setupExistingReview,
			http.StatusNoContent,
		},
		{
			"Success as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			setupExistingReview,
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

			req := createRequest(t, "DELETE", ts.URL+"/reviews/"+effectiveID, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestCreateReviewDBError(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	// Создаем ситуацию с ошибкой БД
	TestAdminDB.Close()
	TestGuestDB.Close()
	TestUserDB.Close()

	req := createRequest(t, "POST", ts.URL+"/reviews",
		generateToken(t, "CLAIM_ROLE_USER"),
		ReviewData{
			UserID:  UsersData[0].ID,
			MovieID: MoviesData[1].ID,
			Rating:  8,
			Comment: "Test comment",
		})
	resp := executeRequest(t, req, http.StatusInternalServerError)
	defer resp.Body.Close()

	if err := InitTestDB(); err != nil {
		log.Fatal("ошибка подключения к БД: ", err)
	}
}

func TestGetReviewsByMovieID(t *testing.T) {
	setupTest := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		err := SeedAll(TestAdminDB)
		if err != nil {
			println(err.Error())
		}
		return ts, MoviesData[0].ID // Используем ID первого фильма из тестовых данных
	}

	tests := []struct {
		name           string
		setup          func(t *testing.T) (*httptest.Server, string)
		movieID        string
		role           string
		expectedStatus int
	}{
		{
			"Valid movie ID with reviews",
			setupTest,
			"",
			"CLAIM_ROLE_USER",
			http.StatusOK,
		},
		{
			"Invalid movie ID format",
			setupTest,
			"invalid-uuid",
			"CLAIM_ROLE_USER",
			http.StatusBadRequest,
		},
		{
			"Non-existent movie ID",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			"",
			"CLAIM_ROLE_USER",
			http.StatusNotFound,
		},
		{
			"Movie without reviews",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				// Используем фильм без отзывов (например, последний в тестовых данных)
				return ts, MoviesData[len(MoviesData)-1].ID
			},
			"",
			"CLAIM_ROLE_USER",
			http.StatusNotFound,
		},
		{
			"As guest user",
			setupTest,
			"",
			"",
			http.StatusOK,
		},
		{
			"As admin user",
			setupTest,
			"",
			"CLAIM_ROLE_ADMIN",
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, movieID := tt.setup(t)
			defer ts.Close()

			// Use provided movieID or fallback to id from setup
			effectiveID := tt.movieID
			if effectiveID == "" {
				effectiveID = movieID
			}

			req := createRequest(t, "GET", ts.URL+"/movies/"+effectiveID+"/reviews", generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var reviews []Review
				parseResponseBody(t, resp, &reviews)

				if len(reviews) == 0 {
					t.Error("Expected non-empty reviews list")
				}

				// Проверяем, что все отзывы относятся к запрошенному фильму
				for _, review := range reviews {
					if review.MovieID != movieID {
						t.Errorf("Expected movie ID %v in review; got %v", movieID, review.MovieID)
					}
				}
			}
		})
	}
}

func TestGetReviewsByUserID(t *testing.T) {
	setupTest := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		SeedAll(TestAdminDB)
		return ts, UsersData[0].ID
	}

	tests := []struct {
		name           string
		setup          func(t *testing.T) (*httptest.Server, string)
		userID         string
		role           string
		expectedStatus int
	}{
		{
			"Valid user ID with reviews",
			setupTest,
			"",
			"CLAIM_ROLE_USER",
			http.StatusOK,
		},
		{
			"Invalid user ID format",
			setupTest,
			"invalid-uuid",
			"CLAIM_ROLE_USER",
			http.StatusBadRequest,
		},
		{
			"Non-existent user ID",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			"",
			"CLAIM_ROLE_USER",
			http.StatusNotFound,
		},
		{
			"User without reviews",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				// Используем пользователя без отзывов (например, последнего в тестовых данных)
				return ts, UsersData[len(UsersData)-1].ID
			},
			"",
			"CLAIM_ROLE_USER",
			http.StatusNotFound,
		},
		{
			"As guest user",
			setupTest,
			"",
			"",
			http.StatusOK,
		},
		{
			"As admin user",
			setupTest,
			"",
			"CLAIM_ROLE_ADMIN",
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, userID := tt.setup(t)
			defer ts.Close()

			// Use provided userID or fallback to id from setup
			effectiveID := tt.userID
			if effectiveID == "" {
				effectiveID = userID
			}

			req := createRequest(t, "GET", ts.URL+"/users/"+effectiveID+"/reviews", generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var reviews []Review
				parseResponseBody(t, resp, &reviews)

				if len(reviews) == 0 {
					t.Error("Expected non-empty reviews list")
				}

				// Проверяем, что все отзывы относятся к запрошенному пользователю
				for _, review := range reviews {
					if review.UserID != userID {
						t.Errorf("Expected user ID %v in review; got %v", userID, review.UserID)
					}
				}
			}
		})
	}
}

func TestGetReviewsByMovieIDDBError(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	// Создаем ситуацию с ошибкой БД
	TestAdminDB.Close()
	TestGuestDB.Close()
	TestUserDB.Close()

	req := createRequest(t, "GET", ts.URL+"/movies/"+MoviesData[0].ID+"/reviews",
		generateToken(t, "CLAIM_ROLE_USER"), nil)
	resp := executeRequest(t, req, http.StatusInternalServerError)
	defer resp.Body.Close()

	if err := InitTestDB(); err != nil {
		log.Fatal("ошибка подключения к БД: ", err)
	}
}

func TestGetReviewsByUserIDDBError(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	// Создаем ситуацию с ошибкой БД
	TestAdminDB.Close()
	TestGuestDB.Close()
	TestUserDB.Close()

	req := createRequest(t, "GET", ts.URL+"/users/"+UsersData[0].ID+"/reviews",
		generateToken(t, "CLAIM_ROLE_USER"), nil)
	resp := executeRequest(t, req, http.StatusInternalServerError)
	defer resp.Body.Close()

	if err := InitTestDB(); err != nil {
		log.Fatal("ошибка подключения к БД: ", err)
	}
}
