package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func getMovieByID(t *testing.T, ts *httptest.Server, token string, index int) Movie {
	req := createRequest(t, "GET", ts.URL+"/movies", token, nil)
	resp := executeRequest(t, req, http.StatusOK)
	defer resp.Body.Close()

	var movies []Movie
	parseResponseBody(t, resp, &movies)

	if len(movies) == 0 {
		t.Fatal("Expected at least one movie, got none")
	}

	if index >= len(movies) {
		t.Fatal("Index is greater than length of data array")
	}

	return movies[index]
}

func TestGetMovies(t *testing.T) {
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

			req := createRequest(t, "GET", ts.URL+"/movies", generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var movies []Movie
				parseResponseBody(t, resp, &movies)

				if len(movies) == 0 {
					t.Error("Expected non-empty movies list")
				}
			}
		})
	}
}

func TestGetMovieByID(t *testing.T) {
	setupValidIDTest := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)                 // Заполняем базу данных тестовыми данными
		return ts, getMovieByID(t, ts, "", 0).ID // Предполагается, что эта функция возвращает ID существующего фильма
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
				return ts, uuid.New().String() // Генерируем новый UUID
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
				return ts, "invalid-id" // Неверный формат ID
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

			req := createRequest(t, "GET", ts.URL+"/movies/"+id, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var movie Movie
				parseResponseBody(t, resp, &movie)

				if movie.ID != id {
					t.Errorf("Expected ID %v; got %v", id, movie.ID)
				}
			}
		})
	}
}
func TestCreateMovie(t *testing.T) {
	validGenresIds := []string{
		GenresData[0].ID,
		GenresData[2].ID,
		GenresData[3].ID,
		GenresData[5].ID,
	}

	validMovie := MovieData{
		Title:            "Test Movie",
		Duration:         "01:30:00", // Пример корректной продолжительности
		Description:      "Test Description",
		AgeLimit:         12,
		BoxOfficeRevenue: 100000.00,
		ReleaseDate:      time.Now(),
		GenreIDs:         validGenresIds,
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
			validMovie,
			nil,
			http.StatusForbidden,
		},
		{
			"Forbidden User",
			"CLAIM_ROLE_USER",
			validMovie,
			nil,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			"CLAIM_ROLE_ADMIN",
			validMovie,
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
			"Empty title",
			"CLAIM_ROLE_ADMIN",
			MovieData{Title: "", Duration: "01:30:00", Description: "Valid", AgeLimit: 12, BoxOfficeRevenue: 100000.00, ReleaseDate: time.Now(), GenreIDs: validGenresIds},
			nil,
			http.StatusBadRequest,
		},
		{
			"Invalid duration",
			"CLAIM_ROLE_ADMIN",
			MovieData{Title: "Valid", Duration: "00:00:00", Description: "Valid", AgeLimit: 12, BoxOfficeRevenue: 100000.00, ReleaseDate: time.Now(), GenreIDs: validGenresIds},
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty description",
			"CLAIM_ROLE_ADMIN",
			MovieData{Title: "Valid", Duration: "01:30:00", Description: "", AgeLimit: 12, BoxOfficeRevenue: 100000.00, ReleaseDate: time.Now(), GenreIDs: validGenresIds},
			nil,
			http.StatusBadRequest,
		},
		{
			"Invalid age limit",
			"CLAIM_ROLE_ADMIN",
			MovieData{Title: "Valid", Duration: "01:30:00", Description: "Valid", AgeLimit: 10, BoxOfficeRevenue: 100000.00, ReleaseDate: time.Now(), GenreIDs: validGenresIds},
			nil,
			http.StatusBadRequest,
		},
		{
			"Valid age limit (0)",
			"CLAIM_ROLE_ADMIN",
			MovieData{Title: "Valid", Duration: "01:30:00", Description: "Valid", AgeLimit: 0, BoxOfficeRevenue: 100000.00, ReleaseDate: time.Now(), GenreIDs: validGenresIds},
			nil,
			http.StatusCreated,
		},
		{
			"Valid age limit (6)",
			"CLAIM_ROLE_ADMIN",
			MovieData{Title: "Valid", Duration: "01:30:00", Description: "Valid", AgeLimit: 6, BoxOfficeRevenue: 100000.00, ReleaseDate: time.Now(), GenreIDs: validGenresIds},
			nil,
			http.StatusCreated,
		},
		{
			"Valid age limit (12)",
			"CLAIM_ROLE_ADMIN",
			MovieData{Title: "Valid", Duration: "01:30:00", Description: "Valid", AgeLimit: 12, BoxOfficeRevenue: 100000.00, ReleaseDate: time.Now(), GenreIDs: validGenresIds},
			nil,
			http.StatusCreated,
		},
		{
			"Valid age limit (16)",
			"CLAIM_ROLE_ADMIN",
			MovieData{Title: "Valid", Duration: "01:30:00", Description: "Valid", AgeLimit: 16, BoxOfficeRevenue: 100000.00, ReleaseDate: time.Now(), GenreIDs: validGenresIds},
			nil,
			http.StatusCreated,
		},
		{
			"Valid age limit (18)",
			"CLAIM_ROLE_ADMIN",
			MovieData{Title: "Valid", Duration: "01:30:00", Description: "Valid", AgeLimit: 18, BoxOfficeRevenue: 100000.00, ReleaseDate: time.Now(), GenreIDs: validGenresIds},
			nil,
			http.StatusCreated,
		},
		{
			"Invalid age limit (19)",
			"CLAIM_ROLE_ADMIN",
			MovieData{Title: "Valid", Duration: "01:30:00", Description: "Valid", AgeLimit: 19, BoxOfficeRevenue: 100000.00, ReleaseDate: time.Now(), GenreIDs: validGenresIds},
			nil,
			http.StatusBadRequest,
		},
		{
			"Negative box office revenue",
			"CLAIM_ROLE_ADMIN",
			MovieData{Title: "Valid", Duration: "01:30:00", Description: "Valid", AgeLimit: 12, BoxOfficeRevenue: -100.00, ReleaseDate: time.Now(), GenreIDs: validGenresIds},
			nil,
			http.StatusBadRequest,
		},
		{
			"Valid box office revenue (0)",
			"CLAIM_ROLE_ADMIN",
			MovieData{Title: "Valid", Duration: "01:30:00", Description: "Valid", AgeLimit: 12, BoxOfficeRevenue: 0.00, ReleaseDate: time.Now(), GenreIDs: validGenresIds},
			nil,
			http.StatusCreated,
		},
		{
			"Valid box office revenue (positive)",
			"CLAIM_ROLE_ADMIN",
			MovieData{Title: "Valid", Duration: "01:30:00", Description: "Valid", AgeLimit: 12, BoxOfficeRevenue: 100000.00, ReleaseDate: time.Now(), GenreIDs: validGenresIds},
			nil,
			http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
			defer ts.Close()
			err := SeedGenres(TestAdminDB)
			if err != nil {
				t.Fatalf("Seed error")
			}

			if tt.setup != nil {
				tt.setup(t)
			}

			req := createRequest(t, "POST", ts.URL+"/movies", generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusCreated {
				var createdID string
				parseResponseBody(t, resp, &createdID)

				if createdID == "" {
					t.Error("Expected non-empty ID in response")
				}

				if _, err := uuid.Parse(createdID); err != nil {
					t.Error("Неверный формат возвращённого UUID")
				}
			}
		})
	}
}

func TestUpdateMovie(t *testing.T) {
	setupExistingMovieWithGenresID := func(t *testing.T) (*httptest.Server, string, []string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, getMovieByID(t, ts, "", 0).ID, []string{
			GenresData[0].ID,
			GenresData[3].ID,
			GenresData[2].ID,
			GenresData[4].ID,
		}
	}

	setupInvalidID := func(t *testing.T) (*httptest.Server, string, []string) {
		ts := setupTestServer()
		SeedAll(TestAdminDB)
		return ts, "invalid-id", []string{
			GenresData[0].ID,
			GenresData[3].ID,
			GenresData[2].ID,
			GenresData[4].ID,
		}
	}

	setupUnknownID := func(t *testing.T) (*httptest.Server, string, []string) {
		ts := setupTestServer()
		SeedAll(TestAdminDB)
		return ts, uuid.NewString(), []string{
			GenresData[0].ID,
			GenresData[3].ID,
			GenresData[2].ID,
			GenresData[4].ID,
		}
	}

	validUpdateData := MovieData{
		Title:            "Updated Movie Title",
		Duration:         "01:30:00",
		Description:      "Updated Description",
		AgeLimit:         16,
		BoxOfficeRevenue: 1000000.00,
		ReleaseDate:      time.Now(),
		GenreIDs:         []string{},
	}

	tests := []struct {
		name           string
		role           string
		id             string
		body           interface{}
		setup          func(t *testing.T) (*httptest.Server, string, []string)
		expectedStatus int
	}{
		{
			"Invalid UUID as Guest",
			"",
			"invalid-uuid",
			validUpdateData,
			setupInvalidID,
			http.StatusBadRequest,
		},
		{
			"Invalid UUID as User",
			"CLAIM_ROLE_USER",
			"invalid-uuid",
			validUpdateData,
			setupInvalidID,
			http.StatusBadRequest,
		},
		{
			"Invalid UUID as Admin",
			"CLAIM_ROLE_ADMIN",
			"invalid-uuid",
			validUpdateData,
			setupInvalidID,
			http.StatusBadRequest,
		},
		{
			"Unknown UUID as Guest",
			"",
			"",
			validUpdateData,
			setupUnknownID,
			http.StatusForbidden,
		},
		{
			"Unknown UUID as User",
			"CLAIM_ROLE_USER",
			"",
			validUpdateData,
			setupUnknownID,
			http.StatusForbidden,
		},
		{
			"Unknown UUID as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			validUpdateData,
			setupUnknownID,
			http.StatusNotFound,
		},
		{
			"Invalid JSON as Guest",
			"",
			"",
			"invalid-json",
			setupExistingMovieWithGenresID,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as User",
			"CLAIM_ROLE_USER",
			"",
			"invalid-json",
			setupExistingMovieWithGenresID,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			"invalid-json",
			setupExistingMovieWithGenresID,
			http.StatusBadRequest,
		},
		{
			"Empty title as Guest",
			"",
			"",
			MovieData{Title: "", Duration: "01:30:00", Description: "Valid", AgeLimit: 16, BoxOfficeRevenue: 1000000.00, ReleaseDate: time.Now()},
			setupExistingMovieWithGenresID,
			http.StatusBadRequest,
		},
		{
			"Invalid duration as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			MovieData{Title: "Valid Title", Duration: "invalid-duration", Description: "Valid", AgeLimit: 16, BoxOfficeRevenue: 1000000.00, ReleaseDate: time.Now()},
			setupExistingMovieWithGenresID,
			http.StatusBadRequest,
		},
		{
			"Invalid age limit as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			MovieData{Title: "Valid Title", Duration: "01:30:00", Description: "Valid", AgeLimit: 99, BoxOfficeRevenue: 1000000.00, ReleaseDate: time.Now()},
			setupExistingMovieWithGenresID,
			http.StatusBadRequest,
		},
		{
			"Success Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			validUpdateData,
			setupExistingMovieWithGenresID,
			http.StatusOK,
		},
		{
			"Title with only spaces",
			"CLAIM_ROLE_ADMIN",
			"",
			MovieData{Title: "   ", Duration: "01:30:00", Description: "Valid", AgeLimit: 16, BoxOfficeRevenue: 1000000.00, ReleaseDate: time.Now()},
			setupExistingMovieWithGenresID,
			http.StatusBadRequest,
		},
		{
			"Description too long (1001 chars)",
			"CLAIM_ROLE_ADMIN",
			"",
			MovieData{Title: "Valid Title", Duration: "01:30:00", Description: strings.Repeat("a", 1001), AgeLimit: 16, BoxOfficeRevenue: 1000000.00, ReleaseDate: time.Now()},
			setupExistingMovieWithGenresID,
			http.StatusBadRequest,
		},
		{
			"Box office revenue negative",
			"CLAIM_ROLE_ADMIN",
			"",
			MovieData{Title: "Valid Title", Duration: "01:30:00", Description: "Valid", AgeLimit: 16, BoxOfficeRevenue: -100.00, ReleaseDate: time.Now()},
			setupExistingMovieWithGenresID,
			http.StatusBadRequest,
		},
		{
			"Valid update with future release date",
			"CLAIM_ROLE_ADMIN",
			"",
			MovieData{Title: "Valid Title", Duration: "01:30:00", Description: "Valid", AgeLimit: 16, BoxOfficeRevenue: 1000000.00, ReleaseDate: time.Now()},
			setupExistingMovieWithGenresID,
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id, genres_ids := tt.setup(t)
			defer ts.Close()

			effectiveID := tt.id
			if effectiveID == "" {
				effectiveID = id
			}

			if movieData, ok := tt.body.(*MovieData); ok {
				movieData.GenreIDs = genres_ids
				tt.body = movieData
			}

			req := createRequest(t, "PUT", ts.URL+"/movies/"+effectiveID, generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestDeleteMovie(t *testing.T) {
	// Setup function for tests needing existing movie
	setupExistingMovie := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, getMovieByID(t, ts, "", 0).ID
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
			"CLAIM_ROLE_USER",
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
			"Forbidden as Guest",
			"",
			"",
			setupExistingMovie,
			http.StatusForbidden,
		},
		{
			"Forbidden as User",
			"CLAIM_ROLE_USER",
			"",
			setupExistingMovie,
			http.StatusForbidden,
		},
		{
			"Dependency error as Admin (У фильма были сеансы)",
			"CLAIM_ROLE_ADMIN",
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, getMovieByID(t, ts, generateToken(t, "CLAIM_ROLE_ADMIN"), 0).ID
			},
			http.StatusFailedDependency,
		},
		{
			"Success as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, getMovieByID(t, ts, generateToken(t, "CLAIM_ROLE_ADMIN"), 5).ID
			},
			http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
			defer ts.Close()

			// Используем предоставленный ID или берем ID из setup
			effectiveID := tt.id
			if effectiveID == "" {
				effectiveID = id
			}

			req := createRequest(t, "DELETE", ts.URL+"/movies/"+effectiveID, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			// Проверяем статус ответа
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestCreateMovieDBError(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	validMovieData := MovieData{
		Title:            "Updated Movie Title",
		Duration:         "01:30:00", // Пример корректного времени
		Description:      "Updated Description",
		AgeLimit:         16,
		BoxOfficeRevenue: 1000000.00,
		ReleaseDate:      time.Now(), // Завтра
		GenreIDs:         []string{"genre1", "genre2"},
	}

	// Создаем ситуацию с ошибкой БД
	TestAdminDB.Close()
	TestGuestDB.Close()
	TestUserDB.Close()

	req := createRequest(t, "POST", ts.URL+"/movies",
		generateToken(t, "CLAIM_ROLE_ADMIN"),
		validMovieData)
	resp := executeRequest(t, req, http.StatusInternalServerError)
	defer resp.Body.Close()

	if err := InitTestDB(); err != nil {
		log.Fatal("ошибка подключения к БД: ", err)
	}
}

func TestSearchMovies(t *testing.T) {
	// Функция для настройки тестового сервера с данными
	setupTestServerWithData := func(t *testing.T) *httptest.Server {
		ts := setupTestServer()

		_ = SeedAll(TestAdminDB)

		return ts
	}

	tests := []struct {
		name           string
		query          string
		expectedTitles []string // Ожидаемые названия фильмов в результате
		expectedStatus int
	}{
		{
			"Точное совпадение",
			"Интерстеллар",
			[]string{"Интерстеллар"},
			http.StatusOK,
		},
		{
			"Частичное совпадение",
			"стеллар",
			[]string{"Интерстеллар"},
			http.StatusOK,
		},
		{
			"Несколько результатов",
			"книга",
			[]string{"Зеленая книга", "Черная книга"},
			http.StatusOK,
		},
		{
			"Регистронезависимый поиск",
			"довод",
			[]string{"Довод"},
			http.StatusOK,
		},
		{
			"Пустой запрос",
			"",
			nil,
			http.StatusBadRequest,
		},
		{
			"Только пробелы",
			"   ",
			nil,
			http.StatusBadRequest,
		},
		{
			"Нет результатов",
			"Несуществующий фильм",
			[]string{},
			http.StatusNotFound,
		},
		{
			"Поиск по описанию (не должен находить)",
			"червоточину",
			[]string{},
			http.StatusNotFound,
		},
		{
			"Поиск с SQL-инъекцией",
			"; DROP TABLE movies;--",
			[]string{},
			http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServerWithData(t)
			defer ts.Close()

			req := createRequest(t, "GET", ts.URL+"/movies/by-title/search?query="+url.QueryEscape(tt.query), "", nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var movies []Movie
				parseResponseBody(t, resp, &movies)

				// Проверяем количество результатов
				if len(movies) != len(tt.expectedTitles) {
					t.Errorf("Expected %d movies, got %d", len(tt.expectedTitles), len(movies))
				}

				// Проверяем, что получили именно те фильмы, которые ожидали
				for _, expectedTitle := range tt.expectedTitles {
					found := false
					for _, movie := range movies {
						if movie.Title == expectedTitle {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected movie %q not found in results", expectedTitle)
					}
				}

				// Дополнительные проверки для найденных фильмов
				for _, movie := range movies {
					// Проверяем, что структура фильма заполнена правильно
					if movie.ID == "" {
						t.Error("Movie ID is empty")
					}
					if movie.Title == "" {
						t.Error("Movie title is empty")
					}
					if movie.Duration == "" {
						t.Error("Movie duration is empty")
					}

					// Проверяем, что название содержит поисковый запрос (регистронезависимо)
					if !strings.Contains(strings.ToLower(movie.Title), strings.ToLower(tt.query)) {
						t.Errorf("Movie title %q doesn't contain query %q", movie.Title, tt.query)
					}
				}
			}
		})
	}

	t.Run("Ошибка базы данных", func(t *testing.T) {
		ts := setupTestServer()
		defer ts.Close()

		TestGuestDB.Close()
		TestUserDB.Close()
		TestAdminDB.Close()

		req := createRequest(t, "GET", ts.URL+"/movies/by-title/search?query=книга", "", nil)
		resp := executeRequest(t, req, http.StatusInternalServerError)
		defer resp.Body.Close()

		if err := InitTestDB(); err != nil {
			t.Fatal("Failed to reinitialize test DB:", err)
		}
	})
}

func TestGetMoviesByAllGenres(t *testing.T) {
	// Настройка тестовых данных
	setupTestData := func(t *testing.T) (*httptest.Server, []string) {
		ts := setupTestServer()
		SeedAll(TestAdminDB)
		genres := []string{
			GenresData[0].ID,
			GenresData[3].ID,
			GenresData[2].ID,
		}

		return ts, genres
	}

	tests := []struct {
		name           string
		genreIDs       []string
		expectedStatus int
		expectedCount  int
		setup          func(t *testing.T) []string
	}{
		{
			"Успешный поиск по одному жанру",
			nil,
			http.StatusOK,
			3,
			func(t *testing.T) []string {
				ts, genres := setupTestData(t)
				defer ts.Close()
				return []string{genres[0]}
			},
		},
		{
			"Успешный поиск по двум жанрам",
			nil,
			http.StatusOK,
			1,
			func(t *testing.T) []string {
				ts, genres := setupTestData(t)
				defer ts.Close()
				return []string{genres[0], genres[1]}
			},
		},
		{
			"Пустой список жанров",
			[]string{},
			http.StatusBadRequest,
			0,
			nil,
		},
		{
			"Неверный формат ID",
			[]string{"invalid-uuid"},
			http.StatusBadRequest,
			0,
			nil,
		},
		{
			"Несуществующие жанры",
			[]string{uuid.New().String(), uuid.New().String()},
			http.StatusNotFound,
			0,
			nil,
		},
		{
			"Часть жанров не существует",
			nil,
			http.StatusNotFound,
			0,
			func(t *testing.T) []string {
				ts, genres := setupTestData(t)
				defer ts.Close()
				return []string{genres[0], uuid.New().String()}
			},
		},
		{
			"SQL-инъекция в параметрах",
			[]string{"1) OR 1=1--"},
			http.StatusBadRequest,
			0,
			nil,
		},
		{
			"Специальные символы в UUID",
			[]string{"00000000-0000-0000-0000-000000000001'--"},
			http.StatusBadRequest,
			0,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
			defer ts.Close()

			var genreIDs []string
			if tt.setup != nil {
				genreIDs = tt.setup(t)
			} else {
				genreIDs = tt.genreIDs
			}

			// Формируем URL с параметрами
			my_url := ts.URL + "/movies/by-genres/search?"
			for _, id := range genreIDs {
				my_url += "genre_ids=" + url.QueryEscape(id) + "&"
			}
			my_url = strings.TrimSuffix(my_url, "&")

			req := createRequest(t, "GET", my_url, "", nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var movies []Movie
				parseResponseBody(t, resp, &movies)

				if len(movies) != tt.expectedCount {
					t.Errorf("Expected %d movies, got %d", tt.expectedCount, len(movies))
				}

				for _, movie := range movies {
					var genreCount int
					err := TestAdminDB.QueryRow(context.Background(), `
                        SELECT COUNT(*) FROM movies_genres 
                        WHERE movie_id = $1 AND genre_id = ANY($2)`,
						movie.ID, genreIDs).Scan(&genreCount)
					if err != nil {
						t.Fatal("Failed to verify genres:", err)
					}
					if genreCount != len(genreIDs) {
						t.Errorf("Movie %s doesn't have all required genres", movie.Title)
					}
				}
			}
		})
	}

	t.Run("Ошибка базы данных", func(t *testing.T) {
		ts := setupTestServer()
		defer ts.Close()

		TestGuestDB.Close()
		TestUserDB.Close()
		TestAdminDB.Close()

		req := createRequest(t, "GET", ts.URL+"/movies/by-genres/search?genre_ids="+uuid.New().String(), "", nil)
		resp := executeRequest(t, req, http.StatusInternalServerError)
		defer resp.Body.Close()

		if err := InitTestDB(); err != nil {
			t.Fatal("Failed to reinitialize test DB:", err)
		}
	})

	t.Run("Проверка прав доступа", func(t *testing.T) {
		ts := setupTestServer()
		defer ts.Close()
		SeedAll(TestAdminDB)

		genreID := GenresData[0].ID

		tests := []struct {
			name           string
			role           string
			expectedStatus int
		}{
			{"Доступ гостя", "", http.StatusOK},
			{"Доступ пользователя", "CLAIM_ROLE_USER", http.StatusOK},
			{"Доступ администратора", "CLAIM_ROLE_ADMIN", http.StatusOK},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := createRequest(t, "GET", ts.URL+"/movies/by-genres/search?genre_ids="+genreID, generateToken(t, tt.role), nil)
				resp := executeRequest(t, req, tt.expectedStatus)
				defer resp.Body.Close()
			})
		}
	})
}
