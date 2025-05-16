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

func TestGetSeatTypes(t *testing.T) {
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

			req := createRequest(t, "GET", ts.URL+"/seat-types", generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			parseResponseBody(t, resp, nil)
		})
	}
}

func TestGetSeatTypeByID(t *testing.T) {
	// Setup for valid ID tests
	setupValidIDTest := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, SeatTypesData[0].ID
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

			req := createRequest(t, "GET", ts.URL+"/seat-types/"+id, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var seat_type SeatType
				parseResponseBody(t, resp, &seat_type)

				if seat_type.ID != id {
					t.Errorf("Expected ID %v; got %v", id, seat_type.ID)
				}
			}
		})
	}
}

func TestCreateSeatType(t *testing.T) {
	validSeatType := SeatTypeData{
		Name:        "Test Seat",
		Description: "Test Description",
	}

	invalidSeatType := SeatTypeData{
		Name:        "",
		Description: "Test Description",
	}

	setupConflictTest := func(t *testing.T) {
		_, err := TestAdminDB.Exec(context.Background(), "INSERT INTO seat_types (name, description) VALUES ($1, $2)", validSeatType.Name, validSeatType.Description)
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
			validSeatType,
			nil,
			http.StatusForbidden,
		},
		{
			"Forbidden User",
			"CLAIM_ROLE_USER",
			validSeatType,
			nil,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			"CLAIM_ROLE_ADMIN",
			validSeatType,
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
			invalidSeatType,
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty field in JSON User",
			"CLAIM_ROLE_USER",
			invalidSeatType,
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty field in JSON Admin",
			"CLAIM_ROLE_ADMIN",
			invalidSeatType,
			nil,
			http.StatusBadRequest,
		},
		{
			"Insert Error Guest",
			"",
			validSeatType,
			setupConflictTest,
			http.StatusForbidden,
		},
		{
			"Insert Error User",
			"CLAIM_ROLE_USER",
			validSeatType,
			setupConflictTest,
			http.StatusForbidden,
		},
		{
			"Insert Error Admin",
			"CLAIM_ROLE_ADMIN",
			validSeatType,
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

			req := createRequest(t, "POST", ts.URL+"/seat-types", generateToken(t, tt.role), tt.body)
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

func TestUpdateSeatType(t *testing.T) {
	validUpdateData := SeatTypeData{
		Name:        "Updated Seat",
		Description: "Updated Description",
	}

	// Setup function for tests needing existing seat type
	setupExistingSeat := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, SeatTypesData[0].ID
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
			setupExistingSeat,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as User",
			"CLAIM_ROLE_USER",
			"",
			"invalid-json",
			setupExistingSeat,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			"invalid-json",
			setupExistingSeat,
			http.StatusBadRequest,
		},
		{
			"Empty Name as Guest",
			"",
			"",
			SeatTypeData{Name: "", Description: "Test"},
			setupExistingSeat,
			http.StatusBadRequest,
		},
		{
			"Empty Name as User",
			"CLAIM_ROLE_USER",
			"",
			SeatTypeData{Name: "", Description: "Test"},
			setupExistingSeat,
			http.StatusBadRequest,
		},
		{
			"Empty Name as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			SeatTypeData{Name: "", Description: "Test"},
			setupExistingSeat,
			http.StatusBadRequest,
		},
		{
			"Forbidden Guest",
			"",
			"",
			validUpdateData,
			setupExistingSeat,
			http.StatusForbidden,
		},
		{
			"Forbidden User",
			"CLAIM_ROLE_USER",
			"",
			validUpdateData,
			setupExistingSeat,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			validUpdateData,
			setupExistingSeat,
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

			req := createRequest(t, "PUT", ts.URL+"/seat-types/"+effectiveID, generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestDeleteSeatType(t *testing.T) {
	// Setup function for tests needing existing seat type
	setupExistingSeat := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, SeatTypesData[0].ID
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
			setupExistingSeat,
			http.StatusForbidden,
		},
		{
			"Forbidden as User",
			"CLAIM_ROLE_USER",
			"",
			setupExistingSeat,
			http.StatusForbidden,
		},
		{
			"Dependency error as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			setupExistingSeat,
			http.StatusConflict,
		},
		{
			"Success as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
				return ts, SeatTypesData[5].ID
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

			req := createRequest(t, "DELETE", ts.URL+"/seat-types/"+effectiveID, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestSeatTypeConstraintsCreate(t *testing.T) {
	tests := []struct {
		name           string
		role           string
		body           SeatTypeData
		expectedStatus int
	}{
		{
			"Empty name",
			"CLAIM_ROLE_ADMIN",
			SeatTypeData{Name: "   ", Description: "Valid"},
			http.StatusBadRequest,
		},
		{
			"Empty description",
			"CLAIM_ROLE_ADMIN",
			SeatTypeData{Name: "Valid", Description: "   "},
			http.StatusBadRequest,
		},
		{
			"Name too long",
			"CLAIM_ROLE_ADMIN",
			SeatTypeData{Name: strings.Repeat("a", 101), Description: "Valid"},
			http.StatusBadRequest,
		},
		{
			"Description too long",
			"CLAIM_ROLE_ADMIN",
			SeatTypeData{Name: "Valid", Description: strings.Repeat("a", 1001)},
			http.StatusBadRequest,
		},
		{
			"Special characters in name",
			"CLAIM_ROLE_ADMIN",
			SeatTypeData{Name: "Тест <script>", Description: "Valid"},
			http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
			defer ts.Close()

			req := createRequest(t, "POST", ts.URL+"/seat-types", generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestSeatTypeConstraintsUpdate(t *testing.T) {
	tests := []struct {
		name           string
		role           string
		body           SeatTypeData
		expectedStatus int
	}{
		{
			"Empty name",
			"CLAIM_ROLE_ADMIN",
			SeatTypeData{Name: "   ", Description: "Valid"},
			http.StatusBadRequest,
		},
		{
			"Empty description",
			"CLAIM_ROLE_ADMIN",
			SeatTypeData{Name: "Valid", Description: "   "},
			http.StatusBadRequest,
		},
		{
			"Name too long",
			"CLAIM_ROLE_ADMIN",
			SeatTypeData{Name: strings.Repeat("a", 101), Description: "Valid"},
			http.StatusBadRequest,
		},
		{
			"Description too long",
			"CLAIM_ROLE_ADMIN",
			SeatTypeData{Name: "Valid", Description: strings.Repeat("a", 1001)},
			http.StatusBadRequest,
		},
		{
			"Special characters in name",
			"CLAIM_ROLE_ADMIN",
			SeatTypeData{Name: "Тест <script>", Description: "Valid"},
			http.StatusOK,
		},
		{
			"Valid update",
			"CLAIM_ROLE_ADMIN",
			SeatTypeData{Name: "Valid Name", Description: "Valid Description"},
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
			defer ts.Close()
			_ = SeedAll(TestAdminDB)

			seat_typeID := SeatTypesData[0].ID

			req := createRequest(t, "PUT", fmt.Sprintf("%s/seat-types/%s", ts.URL, seat_typeID), generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestUpdateSeatConflict(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	_ = SeedAll(TestAdminDB)
	id1 := SeatTypesData[0]
	id2 := SeatTypesData[1]

	updateData := SeatTypeData{
		Name:        id2.Name,
		Description: "Updated",
	}

	req := createRequest(t, "PUT", ts.URL+"/seat-types/"+id1.ID, generateToken(t, "CLAIM_ROLE_ADMIN"), updateData)
	resp := executeRequest(t, req, http.StatusConflict)
	defer resp.Body.Close()
}

func TestCreateSeatTypeDBError(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	// Создаем ситуацию с ошибкой БД
	TestAdminDB.Close()
	TestGuestDB.Close()
	TestUserDB.Close()

	req := createRequest(t, "POST", ts.URL+"/seat-types",
		generateToken(t, "CLAIM_ROLE_ADMIN"),
		SeatTypeData{
			Name:        "name",
			Description: "Updated",
		})
	resp := executeRequest(t, req, http.StatusInternalServerError)
	defer resp.Body.Close()

	if err := InitTestDB(); err != nil {
		log.Fatal("ошибка подключения к БД: ", err)
	}
}
