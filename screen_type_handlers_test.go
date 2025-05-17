package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestGetScreenTypes(t *testing.T) {
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
				_ = SeedAll(TestAdminDB)
			}

			req := createRequest(t, "GET", ts.URL+"/screen-types", generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			parseResponseBody(t, resp, nil)
		})
	}
}

func TestGetScreenTypeByID(t *testing.T) {
	// Setup for valid ID tests
	setupValidIDTest := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, ScreenTypesData[0].ID
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
				_ = SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			"",
			http.StatusNotFound,
		},
		{
			"Unknown ID as User",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			os.Getenv("CLAIM_ROLE_USER"),
			http.StatusNotFound,
		},
		{
			"Unknown ID as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			os.Getenv("CLAIM_ROLE_ADMIN"),
			http.StatusNotFound,
		},
		{
			"Unknown ID When Empty as Guest",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				return ts, uuid.New().String()
			},
			"",
			http.StatusNotFound,
		},
		{
			"Unknown ID When Empty as User",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				return ts, uuid.New().String()
			},
			os.Getenv("CLAIM_ROLE_USER"),
			http.StatusNotFound,
		},
		{
			"Unknown ID When Empty as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				return ts, uuid.New().String()
			},
			os.Getenv("CLAIM_ROLE_ADMIN"),
			http.StatusNotFound,
		},
		{
			"Invalid ID as Guest",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
				return ts, "invalid-id"
			},
			"",
			http.StatusBadRequest,
		},
		{
			"Invalid ID as User",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
				return ts, "invalid-id"
			},
			os.Getenv("CLAIM_ROLE_USER"),
			http.StatusBadRequest,
		},
		{
			"Invalid ID as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
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
			SeedUsers(TestAdminDB)
			defer ts.Close()

			req := createRequest(t, "GET", ts.URL+"/screen-types/"+id, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var screen_type ScreenType
				parseResponseBody(t, resp, &screen_type)

				if screen_type.ID != id {
					t.Errorf("Expected ID %v; got %v", id, screen_type.ID)
				}
			}
		})
	}
}

func TestCreateScreenType(t *testing.T) {
	validScreenType := ScreenTypeData{
		Name:        "Test Screen",
		Description: "Test Description",
	}

	invalidScreenType := ScreenTypeData{
		Name:        "",
		Description: "Test Description",
	}

	setupConflictTest := func(t *testing.T) {
		_, err := TestAdminDB.Exec(context.Background(), "INSERT INTO screen_types (name, description) VALUES ($1, $2)", validScreenType.Name, validScreenType.Description)
		if err != nil {
			t.Fatalf("Failed to insert into test database: %v", err)
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
			validScreenType,
			nil,
			http.StatusForbidden,
		},
		{
			"Forbidden User",
			os.Getenv("CLAIM_ROLE_USER"),
			validScreenType,
			nil,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			validScreenType,
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
			"Empty field in JSON Guest",
			"",
			invalidScreenType,
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty field in JSON User",
			os.Getenv("CLAIM_ROLE_USER"),
			invalidScreenType,
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty field in JSON Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			invalidScreenType,
			nil,
			http.StatusBadRequest,
		},
		{
			"Insert Error Guest",
			"",
			validScreenType,
			setupConflictTest,
			http.StatusForbidden,
		},
		{
			"Insert Error User",
			os.Getenv("CLAIM_ROLE_USER"),
			validScreenType,
			setupConflictTest,
			http.StatusForbidden,
		},
		{
			"Insert Error Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			validScreenType,
			setupConflictTest,
			http.StatusConflict,
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

			req := createRequest(t, "POST", ts.URL+"/screen-types", generateToken(t, tt.role), tt.body)
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

func TestUpdateScreenType(t *testing.T) {
	validUpdateData := ScreenTypeData{
		Name:        "Updated Screen",
		Description: "Updated Description",
	}

	// Setup function for tests needing existing screen type
	setupExistingScreen := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, ScreenTypesData[0].ID
	}

	unknown_id := uuid.NewString()

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
				_ = SeedAll(TestAdminDB)
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
				_ = SeedAll(TestAdminDB)
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
				_ = SeedAll(TestAdminDB)
				return ts, ""
			},
			http.StatusBadRequest,
		},
		{
			"Unknown UUID as Guest",
			"",
			unknown_id,
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
				return ts, unknown_id
			},
			http.StatusForbidden,
		},
		{
			"Unknown UUID as User",
			os.Getenv("CLAIM_ROLE_USER"),
			unknown_id,
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				return ts, unknown_id
			},
			http.StatusForbidden,
		},
		{
			"Unknown UUID as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			unknown_id,
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				return ts, unknown_id
			},
			http.StatusNotFound,
		},
		{
			"Invalid JSON as Guest",
			"",
			"",
			"invalid-json",
			setupExistingScreen,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as User",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			"invalid-json",
			setupExistingScreen,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			"invalid-json",
			setupExistingScreen,
			http.StatusBadRequest,
		},
		{
			"Empty Name as Guest",
			"",
			"",
			ScreenTypeData{Name: "", Description: "Test"},
			setupExistingScreen,
			http.StatusBadRequest,
		},
		{
			"Empty Name as User",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			ScreenTypeData{Name: "", Description: "Test"},
			setupExistingScreen,
			http.StatusBadRequest,
		},
		{
			"Empty Name as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			ScreenTypeData{Name: "", Description: "Test"},
			setupExistingScreen,
			http.StatusBadRequest,
		},
		{
			"Forbidden Guest",
			"",
			"",
			validUpdateData,
			setupExistingScreen,
			http.StatusForbidden,
		},
		{
			"Forbidden User",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			validUpdateData,
			setupExistingScreen,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			validUpdateData,
			setupExistingScreen,
			http.StatusOK,
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

			req := createRequest(t, "PUT", ts.URL+"/screen-types/"+effectiveID, generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestDeleteScreenType(t *testing.T) {
	// Setup function for tests needing existing screen type
	setupExistingScreen := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, ScreenTypesData[0].ID
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
				_ = SeedAll(TestAdminDB)
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
				_ = SeedAll(TestAdminDB)
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
				_ = SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusNotFound,
		},
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
			"Invalid UUID as User",
			os.Getenv("CLAIM_ROLE_USER"),
			"invalid-uuid",
			func(t *testing.T) (*httptest.Server, string) {
				return setupTestServer(), ""
			},
			http.StatusBadRequest,
		},
		{
			"Invalid UUID as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
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
			setupExistingScreen,
			http.StatusForbidden,
		},
		{
			"Forbidden as User",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			setupExistingScreen,
			http.StatusForbidden,
		},
		{
			"Dependency error as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			setupExistingScreen,
			http.StatusConflict,
		},
		{
			"Success as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
				return ts, ScreenTypesData[5].ID
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

			req := createRequest(t, "DELETE", ts.URL+"/screen-types/"+effectiveID, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestScreenTypeConstraintsCreate(t *testing.T) {
	tests := []struct {
		name           string
		role           string
		body           ScreenTypeData
		expectedStatus int
	}{
		{
			"Empty name",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			ScreenTypeData{Name: "   ", Description: "Valid"},
			http.StatusBadRequest,
		},
		{
			"Empty description",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			ScreenTypeData{Name: "Valid", Description: "   "},
			http.StatusBadRequest,
		},
		{
			"Name too long",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			ScreenTypeData{Name: strings.Repeat("a", 101), Description: "Valid"},
			http.StatusBadRequest,
		},
		{
			"Description too long",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			ScreenTypeData{Name: "Valid", Description: strings.Repeat("a", 1001)},
			http.StatusBadRequest,
		},
		{
			"Special characters in name",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			ScreenTypeData{Name: "Тест <script>", Description: "Valid"},
			http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
			SeedUsers(TestAdminDB)
			defer ts.Close()

			req := createRequest(t, "POST", ts.URL+"/screen-types", generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestScreenTypeConstraintsUpdate(t *testing.T) {
	tests := []struct {
		name           string
		role           string
		body           ScreenTypeData
		expectedStatus int
	}{
		{
			"Empty name",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			ScreenTypeData{Name: "   ", Description: "Valid"},
			http.StatusBadRequest,
		},
		{
			"Empty description",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			ScreenTypeData{Name: "Valid", Description: "   "},
			http.StatusBadRequest,
		},
		{
			"Name too long",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			ScreenTypeData{Name: strings.Repeat("a", 101), Description: "Valid"},
			http.StatusBadRequest,
		},
		{
			"Description too long",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			ScreenTypeData{Name: "Valid", Description: strings.Repeat("a", 1001)},
			http.StatusBadRequest,
		},
		{
			"Special characters in name",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			ScreenTypeData{Name: "Тест <script>", Description: "Valid"},
			http.StatusOK,
		},
		{
			"Valid update",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			ScreenTypeData{Name: "Valid Name", Description: "Valid Description"},
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
			defer ts.Close()
			_ = SeedAll(TestAdminDB)

			screen_typeID := ScreenTypesData[0].ID

			req := createRequest(t, "PUT", fmt.Sprintf("%s/screen-types/%s", ts.URL, screen_typeID), generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestUpdateConflict(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	_ = SeedAll(TestAdminDB)
	id1 := ScreenTypesData[0]
	id2 := ScreenTypesData[1]

	updateData := ScreenTypeData{
		Name:        id2.Name,
		Description: "Updated",
	}

	req := createRequest(t, "PUT", ts.URL+"/screen-types/"+id1.ID, generateToken(t, os.Getenv("CLAIM_ROLE_ADMIN")), updateData)
	resp := executeRequest(t, req, http.StatusConflict)
	defer resp.Body.Close()
}

func TestCreateScreenTypeDBError(t *testing.T) {
	ts := setupTestServer()
	SeedUsers(TestAdminDB)
	defer ts.Close()

	// Создаем ситуацию с ошибкой БД
	TestAdminDB.Close()
	TestGuestDB.Close()
	TestUserDB.Close()

	req := createRequest(t, "POST", ts.URL+"/screen-types",
		generateToken(t, os.Getenv("CLAIM_ROLE_ADMIN")),
		ScreenTypeData{
			Name:        "name",
			Description: "Updated",
		})
	resp := executeRequest(t, req, http.StatusInternalServerError)
	defer resp.Body.Close()

	if err := InitTestDB(); err != nil {
		log.Fatal("ошибка подключения к БД: ", err)
	}
}

func TestSearchScreenTypes(t *testing.T) {
	setupWithScreenTypes := func(t *testing.T) *httptest.Server {
		ts := setupTestServer()
		_ = SeedScreenTypes(TestAdminDB)
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
			setupWithScreenTypes,
			http.StatusBadRequest,
			0,
			os.Getenv("CLAIM_ROLE_USER"),
		},
		{
			"Только пробельные символы - ошибка",
			"          		",
			setupWithScreenTypes,
			http.StatusBadRequest,
			0,
			os.Getenv("CLAIM_ROLE_USER"),
		},
		{
			"Короткий запрос",
			"Л",
			setupWithScreenTypes,
			http.StatusOK,
			1,
			os.Getenv("CLAIM_ROLE_USER"),
		},
		{
			"Нет совпадений",
			"Плазма",
			setupWithScreenTypes,
			http.StatusNotFound,
			0,
			os.Getenv("CLAIM_ROLE_USER"),
		},
		{
			"Точное совпадение - LED",
			"LED",
			setupWithScreenTypes,
			http.StatusOK,
			2,
			os.Getenv("CLAIM_ROLE_USER"),
		},
		{
			"Частичное совпадение - 'сте'",
			"сте",
			setupWithScreenTypes,
			http.StatusOK,
			1,
			os.Getenv("CLAIM_ROLE_USER"),
		},
		{
			"Поиск без учета регистра - 'oLeD'",
			"oLeD",
			setupWithScreenTypes,
			http.StatusOK,
			1,
			os.Getenv("CLAIM_ROLE_USER"),
		},
		{
			"Поиск с пробелами - 'LCD'",
			"LCD",
			setupWithScreenTypes,
			http.StatusOK,
			1,
			os.Getenv("CLAIM_ROLE_USER"),
		},
		{
			"Частичное совпадение с пробелами - 'система'",
			"  система    ",
			setupWithScreenTypes,
			http.StatusOK,
			1,
			os.Getenv("CLAIM_ROLE_USER"),
		},
		{
			"Админ имеет доступ",
			"OLED",
			setupWithScreenTypes,
			http.StatusOK,
			1,
			os.Getenv("CLAIM_ROLE_ADMIN"),
		},
		{
			"Гость имеет доступ",
			"OLED",
			setupWithScreenTypes,
			http.StatusOK,
			1,
			"",
		},
		{
			"Специальные символы в запросе - 'LED/'",
			"LED/",
			setupWithScreenTypes,
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

			req := createRequest(t, "GET", ts.URL+"/screen-types/search?query="+url.QueryEscape(tt.query), generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var screenTypes []ScreenType
				if err := json.NewDecoder(resp.Body).Decode(&screenTypes); err != nil {
					t.Fatalf("Could not decode response: %v", err)
				}

				if len(screenTypes) != tt.expectedCount {
					t.Errorf("Expected %d screen types, got %d: %v", tt.expectedCount, len(screenTypes), screenTypes)
				}

				lowerQuery := strings.ToLower(tt.query)
				lowerQuery = PrepareString(lowerQuery)
				for _, screenType := range screenTypes {
					if !strings.Contains(strings.ToLower(screenType.Name), lowerQuery) {
						t.Errorf("Screen type name '%s' does not contain query '%s'", screenType.Name, tt.query)
					}
				}
			}
		})
	}
}
