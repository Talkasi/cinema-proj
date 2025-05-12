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

func getEquipmentTypeByID(t *testing.T, ts *httptest.Server, token string, index int) EquipmentType {
	req := createRequest(t, "GET", ts.URL+"/equipment-types", token, nil)
	resp := executeRequest(t, req, http.StatusOK)
	defer resp.Body.Close()

	var equipmentTypes []EquipmentType
	parseResponseBody(t, resp, &equipmentTypes)

	if len(equipmentTypes) == 0 {
		t.Fatal("Expected at least one equipment type, got none")
	}

	if index >= len(equipmentTypes) {
		t.Fatal("Index is greater than length of data array")
	}

	return equipmentTypes[index]
}

func TestGetEquipmentTypes(t *testing.T) {
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

			req := createRequest(t, "GET", ts.URL+"/equipment-types", generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			parseResponseBody(t, resp, nil)
		})
	}
}

func TestGetEquipmentTypeByID(t *testing.T) {
	// Setup for valid ID tests
	setupValidIDTest := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, getEquipmentTypeByID(t, ts, "", 0).ID
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

			req := createRequest(t, "GET", ts.URL+"/equipment-types/"+id, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var equipmentType EquipmentType
				parseResponseBody(t, resp, &equipmentType)

				if equipmentType.ID != id {
					t.Errorf("Expected ID %v; got %v", id, equipmentType.ID)
				}
			}
		})
	}
}

func TestCreateEquipmentType(t *testing.T) {
	validEquipmentType := EquipmentTypeData{
		Name:        "Test Equipment",
		Description: "Test Description",
	}

	invalidEquipmentType := EquipmentTypeData{
		Name:        "",
		Description: "Test Description",
	}

	setupConflictTest := func(t *testing.T) {
		_, err := TestAdminDB.Exec(context.Background(), "INSERT INTO equipment_types (name, description) VALUES ($1, $2)", validEquipmentType.Name, validEquipmentType.Description)
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
			validEquipmentType,
			nil,
			http.StatusForbidden,
		},
		{
			"Forbidden User",
			"CLAIM_ROLE_USER",
			validEquipmentType,
			nil,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			"CLAIM_ROLE_ADMIN",
			validEquipmentType,
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
			invalidEquipmentType,
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty field in JSON User",
			"CLAIM_ROLE_USER",
			invalidEquipmentType,
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty field in JSON Admin",
			"CLAIM_ROLE_ADMIN",
			invalidEquipmentType,
			nil,
			http.StatusBadRequest,
		},
		{
			"Insert Error Guest",
			"",
			validEquipmentType,
			setupConflictTest,
			http.StatusForbidden,
		},
		{
			"Insert Error User",
			"CLAIM_ROLE_USER",
			validEquipmentType,
			setupConflictTest,
			http.StatusForbidden,
		},
		{
			"Insert Error Admin",
			"CLAIM_ROLE_ADMIN",
			validEquipmentType,
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

			req := createRequest(t, "POST", ts.URL+"/equipment-types", generateToken(t, tt.role), tt.body)
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

func TestUpdateEquipmentType(t *testing.T) {
	validUpdateData := EquipmentTypeData{
		Name:        "Updated Equipment",
		Description: "Updated Description",
	}

	// Setup function for tests needing existing equipment type
	setupExistingEquipment := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, getEquipmentTypeByID(t, ts, generateToken(t, "CLAIM_ROLE_ADMIN"), 0).ID
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
			setupExistingEquipment,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as User",
			"CLAIM_ROLE_USER",
			"",
			"invalid-json",
			setupExistingEquipment,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			"invalid-json",
			setupExistingEquipment,
			http.StatusBadRequest,
		},
		{
			"Empty Name as Guest",
			"",
			"",
			EquipmentTypeData{Name: "", Description: "Test"},
			setupExistingEquipment,
			http.StatusBadRequest,
		},
		{
			"Empty Name as User",
			"CLAIM_ROLE_USER",
			"",
			EquipmentTypeData{Name: "", Description: "Test"},
			setupExistingEquipment,
			http.StatusBadRequest,
		},
		{
			"Empty Name as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			EquipmentTypeData{Name: "", Description: "Test"},
			setupExistingEquipment,
			http.StatusBadRequest,
		},
		{
			"Forbidden Guest",
			"",
			"",
			validUpdateData,
			setupExistingEquipment,
			http.StatusForbidden,
		},
		{
			"Forbidden User",
			"CLAIM_ROLE_USER",
			"",
			validUpdateData,
			setupExistingEquipment,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			validUpdateData,
			setupExistingEquipment,
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

			req := createRequest(t, "PUT", ts.URL+"/equipment-types/"+effectiveID, generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestDeleteEquipmentType(t *testing.T) {
	// Setup function for tests needing existing equipment type
	setupExistingEquipment := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, getEquipmentTypeByID(t, ts, generateToken(t, "CLAIM_ROLE_ADMIN"), 0).ID
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
			setupExistingEquipment,
			http.StatusForbidden,
		},
		{
			"Forbidden as User",
			"CLAIM_ROLE_USER",
			"",
			setupExistingEquipment,
			http.StatusForbidden,
		},
		{
			"Dependency error as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			setupExistingEquipment,
			http.StatusFailedDependency,
		},
		{
			"Success as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
				return ts, getEquipmentTypeByID(t, ts, generateToken(t, "CLAIM_ROLE_ADMIN"), 5).ID
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

			req := createRequest(t, "DELETE", ts.URL+"/equipment-types/"+effectiveID, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestEquipmentTypeConstraintsCreate(t *testing.T) {
	tests := []struct {
		name           string
		role           string
		body           EquipmentTypeData
		expectedStatus int
	}{
		{
			"Empty name",
			"CLAIM_ROLE_ADMIN",
			EquipmentTypeData{Name: "   ", Description: "Valid"},
			http.StatusBadRequest,
		},
		{
			"Empty description",
			"CLAIM_ROLE_ADMIN",
			EquipmentTypeData{Name: "Valid", Description: "   "},
			http.StatusBadRequest,
		},
		{
			"Name too long",
			"CLAIM_ROLE_ADMIN",
			EquipmentTypeData{Name: strings.Repeat("a", 101), Description: "Valid"},
			http.StatusBadRequest,
		},
		{
			"Description too long",
			"CLAIM_ROLE_ADMIN",
			EquipmentTypeData{Name: "Valid", Description: strings.Repeat("a", 1001)},
			http.StatusBadRequest,
		},
		{
			"Special characters in name",
			"CLAIM_ROLE_ADMIN",
			EquipmentTypeData{Name: "Тест <script>", Description: "Valid"},
			http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
			defer ts.Close()

			req := createRequest(t, "POST", ts.URL+"/equipment-types", generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestEquipmentTypeConstraintsUpdate(t *testing.T) {
	tests := []struct {
		name           string
		role           string
		body           EquipmentTypeData
		expectedStatus int
	}{
		{
			"Empty name",
			"CLAIM_ROLE_ADMIN",
			EquipmentTypeData{Name: "   ", Description: "Valid"},
			http.StatusBadRequest,
		},
		{
			"Empty description",
			"CLAIM_ROLE_ADMIN",
			EquipmentTypeData{Name: "Valid", Description: "   "},
			http.StatusBadRequest,
		},
		{
			"Name too long",
			"CLAIM_ROLE_ADMIN",
			EquipmentTypeData{Name: strings.Repeat("a", 101), Description: "Valid"},
			http.StatusBadRequest,
		},
		{
			"Description too long",
			"CLAIM_ROLE_ADMIN",
			EquipmentTypeData{Name: "Valid", Description: strings.Repeat("a", 1001)},
			http.StatusBadRequest,
		},
		{
			"Special characters in name",
			"CLAIM_ROLE_ADMIN",
			EquipmentTypeData{Name: "Тест <script>", Description: "Valid"},
			http.StatusOK,
		},
		{
			"Valid update",
			"CLAIM_ROLE_ADMIN",
			EquipmentTypeData{Name: "Valid Name", Description: "Valid Description"},
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
			defer ts.Close()
			_ = SeedAll(TestAdminDB)

			equipmentTypeID := getEquipmentTypeByID(t, ts, generateToken(t, "CLAIM_ROLE_ADMIN"), 0).ID

			req := createRequest(t, "PUT", fmt.Sprintf("%s/equipment-types/%s", ts.URL, equipmentTypeID), generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestUpdateConflict(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	_ = SeedAll(TestAdminDB)
	id1 := getEquipmentTypeByID(t, ts, generateToken(t, "CLAIM_ROLE_ADMIN"), 0)
	id2 := getEquipmentTypeByID(t, ts, generateToken(t, "CLAIM_ROLE_ADMIN"), 1)

	updateData := EquipmentTypeData{
		Name:        id2.Name,
		Description: "Updated",
	}

	req := createRequest(t, "PUT", ts.URL+"/equipment-types/"+id1.ID, generateToken(t, "CLAIM_ROLE_ADMIN"), updateData)
	resp := executeRequest(t, req, http.StatusConflict)
	defer resp.Body.Close()
}

func TestCreateEquipmentTypeDBError(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	// Создаем ситуацию с ошибкой БД
	TestAdminDB.Close()
	TestGuestDB.Close()
	TestUserDB.Close()

	req := createRequest(t, "POST", ts.URL+"/equipment-types",
		generateToken(t, "CLAIM_ROLE_ADMIN"),
		EquipmentTypeData{
			Name:        "name",
			Description: "Updated",
		})
	resp := executeRequest(t, req, http.StatusInternalServerError)
	defer resp.Body.Close()

	if err := InitTestDB(); err != nil {
		log.Fatal("ошибка подключения к БД: ", err)
	}
}
