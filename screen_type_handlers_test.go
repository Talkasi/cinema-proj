package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func getScreenTypeByID(t *testing.T, ts *httptest.Server, token string, index int) ScreenType {
	req := createRequest(t, "GET", ts.URL+"/screen-types", token, nil)
	resp := executeRequest(t, req, http.StatusOK)
	defer resp.Body.Close()

	var screenTypes []ScreenType
	parseResponseBody(t, resp, &screenTypes)

	if len(screenTypes) == 0 {
		t.Fatal("Expected at least one screen type, got none")
	}

	if index >= len(screenTypes) {
		t.Fatal("Index is greater than length of data array")
	}

	return screenTypes[index]
}

func TestGetScreenTypes(t *testing.T) {
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
		return ts, getScreenTypeByID(t, ts, "", 0).ID
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
			"CLAIM_ROLE_USER",
			http.StatusNotFound,
		},
		{
			"Unknown ID as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			"CLAIM_ROLE_ADMIN",
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
			"CLAIM_ROLE_USER",
			http.StatusNotFound,
		},
		{
			"Unknown ID When Empty as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				return ts, uuid.New().String()
			},
			"CLAIM_ROLE_ADMIN",
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
			"CLAIM_ROLE_USER",
			http.StatusBadRequest,
		},
		{
			"Invalid ID as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
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

			req := createRequest(t, "GET", ts.URL+"/screen-types/"+id, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var screenType ScreenType
				parseResponseBody(t, resp, &screenType)

				if screenType.ID != id {
					t.Errorf("Expected ID %v; got %v", id, screenType.ID)
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
			"CLAIM_ROLE_USER",
			validScreenType,
			nil,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			"CLAIM_ROLE_ADMIN",
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
			"Empty field in JSON Guest",
			"",
			invalidScreenType,
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty field in JSON User",
			"CLAIM_ROLE_USER",
			invalidScreenType,
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty field in JSON Admin",
			"CLAIM_ROLE_ADMIN",
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
			"CLAIM_ROLE_USER",
			validScreenType,
			setupConflictTest,
			http.StatusForbidden,
		},
		{
			"Insert Error Admin",
			"CLAIM_ROLE_ADMIN",
			validScreenType,
			setupConflictTest,
			http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
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
		return ts, getScreenTypeByID(t, ts, generateToken(t, "CLAIM_ROLE_ADMIN"), 0).ID
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
			"CLAIM_ROLE_USER",
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
			"CLAIM_ROLE_ADMIN",
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
			"CLAIM_ROLE_USER",
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
			"CLAIM_ROLE_ADMIN",
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
			"CLAIM_ROLE_USER",
			"",
			"invalid-json",
			setupExistingScreen,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as Admin",
			"CLAIM_ROLE_ADMIN",
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
			"CLAIM_ROLE_USER",
			"",
			ScreenTypeData{Name: "", Description: "Test"},
			setupExistingScreen,
			http.StatusBadRequest,
		},
		{
			"Empty Name as Admin",
			"CLAIM_ROLE_ADMIN",
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
			"CLAIM_ROLE_USER",
			"",
			validUpdateData,
			setupExistingScreen,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			validUpdateData,
			setupExistingScreen,
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
		return ts, getScreenTypeByID(t, ts, generateToken(t, "CLAIM_ROLE_ADMIN"), 0).ID
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
			"CLAIM_ROLE_USER",
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
			"CLAIM_ROLE_ADMIN",
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
			"CLAIM_ROLE_USER",
			"invalid-uuid",
			func(t *testing.T) (*httptest.Server, string) {
				return setupTestServer(), ""
			},
			http.StatusBadRequest,
		},
		{
			"Invalid UUID as Admin",
			"CLAIM_ROLE_ADMIN",
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
			"CLAIM_ROLE_USER",
			"",
			setupExistingScreen,
			http.StatusForbidden,
		},
		{
			"Dependency error as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			setupExistingScreen,
			http.StatusFailedDependency,
		},
		{
			"Success as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
				return ts, getScreenTypeByID(t, ts, generateToken(t, "CLAIM_ROLE_ADMIN"), 5).ID
			},
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
			"CLAIM_ROLE_ADMIN",
			ScreenTypeData{Name: "   ", Description: "Valid"},
			http.StatusBadRequest,
		},
		{
			"Empty description",
			"CLAIM_ROLE_ADMIN",
			ScreenTypeData{Name: "Valid", Description: "   "},
			http.StatusBadRequest,
		},
		{
			"Name too long",
			"CLAIM_ROLE_ADMIN",
			ScreenTypeData{Name: strings.Repeat("a", 101), Description: "Valid"},
			http.StatusBadRequest,
		},
		{
			"Description too long",
			"CLAIM_ROLE_ADMIN",
			ScreenTypeData{Name: "Valid", Description: strings.Repeat("a", 1001)},
			http.StatusBadRequest,
		},
		{
			"Special characters in name",
			"CLAIM_ROLE_ADMIN",
			ScreenTypeData{Name: "Тест <script>", Description: "Valid"},
			http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
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
			"CLAIM_ROLE_ADMIN",
			ScreenTypeData{Name: "   ", Description: "Valid"},
			http.StatusBadRequest,
		},
		{
			"Empty description",
			"CLAIM_ROLE_ADMIN",
			ScreenTypeData{Name: "Valid", Description: "   "},
			http.StatusBadRequest,
		},
		{
			"Name too long",
			"CLAIM_ROLE_ADMIN",
			ScreenTypeData{Name: strings.Repeat("a", 101), Description: "Valid"},
			http.StatusBadRequest,
		},
		{
			"Description too long",
			"CLAIM_ROLE_ADMIN",
			ScreenTypeData{Name: "Valid", Description: strings.Repeat("a", 1001)},
			http.StatusBadRequest,
		},
		{
			"Special characters in name",
			"CLAIM_ROLE_ADMIN",
			ScreenTypeData{Name: "Тест <script>", Description: "Valid"},
			http.StatusOK,
		},
		{
			"Valid update",
			"CLAIM_ROLE_ADMIN",
			ScreenTypeData{Name: "Valid Name", Description: "Valid Description"},
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
			defer ts.Close()
			_ = SeedAll(TestAdminDB)

			screenTypeID := getScreenTypeByID(t, ts, generateToken(t, "CLAIM_ROLE_ADMIN"), 0).ID

			req := createRequest(t, "PUT", fmt.Sprintf("%s/screen-types/%s", ts.URL, screenTypeID), generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestUpdateConflict(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	_ = SeedAll(TestAdminDB)
	id1 := getScreenTypeByID(t, ts, generateToken(t, "CLAIM_ROLE_ADMIN"), 0)
	id2 := getScreenTypeByID(t, ts, generateToken(t, "CLAIM_ROLE_ADMIN"), 1)

	updateData := ScreenTypeData{
		Name:        id2.Name,
		Description: "Updated",
	}

	req := createRequest(t, "PUT", ts.URL+"/screen-types/"+id1.ID, generateToken(t, "CLAIM_ROLE_ADMIN"), updateData)
	resp := executeRequest(t, req, http.StatusConflict)
	defer resp.Body.Close()
}

func TestCreateScreenTypeDBError(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	// Создаем ситуацию с ошибкой БД
	TestAdminDB.Close()
	TestGuestDB.Close()
	TestUserDB.Close()

	req := createRequest(t, "POST", ts.URL+"/screen-types",
		generateToken(t, "CLAIM_ROLE_ADMIN"),
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
