package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func setupTestEquipmentTypesServer(t *testing.T, clearTable bool) *httptest.Server {
	// if clearTable {
	// 	_ = ClearTable(TestAdminDB, "equipment_types")
	// }
	err := ClearTable(TestAdminDB, "equipment_types")
	if err != nil {
		println(err.Error())
	}
	return httptest.NewServer(NewRouter())
}

func getFirstEquipmentTypeID(t *testing.T, ts *httptest.Server, token string) string {
	req := createRequest(t, "GET", ts.URL+"/equipment-types", token, nil)
	resp := executeRequest(t, req, http.StatusOK)
	defer resp.Body.Close()

	var equipmentTypes []EquipmentType
	parseResponseBody(t, resp, &equipmentTypes)

	if len(equipmentTypes) == 0 {
		t.Fatal("Expected at least one equipment type, got none")
	}

	return equipmentTypes[0].ID
}

func TestGetEquipmentTypes(t *testing.T) {
	tests := []struct {
		name           string
		seedData       bool
		role           string
		expectedStatus int
	}{
		{"Empty as Guest", false, "", http.StatusNotFound},
		{"Empty as User", false, "ruser", http.StatusNotFound},
		{"Empty as Admin", false, "admin", http.StatusNotFound},
		{"NonEmpty as Guest", true, "", http.StatusOK},
		{"NonEmpty as User", true, "ruser", http.StatusOK},
		{"NonEmpty as Admin", true, "admin", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestEquipmentTypesServer(t, true)
			defer ts.Close()

			if tt.seedData {
				_ = SeedEquipmentTypes(TestAdminDB)
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
		ts := setupTestEquipmentTypesServer(t, true)
		_ = SeedEquipmentTypes(TestAdminDB)
		return ts, getFirstEquipmentTypeID(t, ts, "")
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
				ts := setupTestEquipmentTypesServer(t, true)
				_ = SeedEquipmentTypes(TestAdminDB)
				return ts, uuid.New().String()
			},
			"",
			http.StatusNotFound,
		},
		{
			"Unknown ID as User",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestEquipmentTypesServer(t, true)
				_ = SeedEquipmentTypes(TestAdminDB)
				return ts, uuid.New().String()
			},
			"ruser",
			http.StatusNotFound,
		},
		{
			"Unknown ID as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestEquipmentTypesServer(t, true)
				_ = SeedEquipmentTypes(TestAdminDB)
				return ts, uuid.New().String()
			},
			"admin",
			http.StatusNotFound,
		},
		{
			"Unknown ID When Empty as Guest",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestEquipmentTypesServer(t, true)
				return ts, uuid.New().String()
			},
			"",
			http.StatusNotFound,
		},
		{
			"Unknown ID When Empty as User",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestEquipmentTypesServer(t, true)
				return ts, uuid.New().String()
			},
			"ruser",
			http.StatusNotFound,
		},
		{
			"Unknown ID When Empty as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestEquipmentTypesServer(t, true)
				return ts, uuid.New().String()
			},
			"admin",
			http.StatusNotFound,
		},
		{
			"Invalid ID as Guest",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestEquipmentTypesServer(t, true)
				_ = SeedEquipmentTypes(TestAdminDB)
				return ts, "invalid-id"
			},
			"",
			http.StatusBadRequest,
		},
		{
			"Invalid ID as User",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestEquipmentTypesServer(t, true)
				_ = SeedEquipmentTypes(TestAdminDB)
				return ts, "invalid-id"
			},
			"ruser",
			http.StatusBadRequest,
		},
		{
			"Invalid ID as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestEquipmentTypesServer(t, true)
				_ = SeedEquipmentTypes(TestAdminDB)
				return ts, "invalid-id"
			},
			"admin",
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
			"ruser",
			http.StatusOK,
		},
		{
			"Valid ID as Admin",
			setupValidIDTest,
			"admin",
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
			"ruser",
			validEquipmentType,
			nil,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			"admin",
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
			"ruser",
			"{invalid json}",
			nil,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON Admin",
			"admin",
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
			"ruser",
			invalidEquipmentType,
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty field in JSON Admin",
			"admin",
			invalidEquipmentType,
			nil,
			http.StatusBadRequest,
		},
		{
			"Insert Error Guest",
			"",
			validEquipmentType,
			func(t *testing.T) {
				_, err := TestAdminDB.Exec(context.Background(), "INSERT INTO equipment_types (name) VALUES ($1)", validEquipmentType.Name)
				if err != nil {
					t.Fatalf("Failed to insert into test database: %v", err)
				}
			},
			http.StatusForbidden,
		},
		{
			"Insert Error User",
			"ruser",
			validEquipmentType,
			func(t *testing.T) {
				_, err := TestAdminDB.Exec(context.Background(), "INSERT INTO equipment_types (name) VALUES ($1)", validEquipmentType.Name)
				if err != nil {
					t.Fatalf("Failed to insert into test database: %v", err)
				}
			},
			http.StatusForbidden,
		},
		{
			"Insert Error Admin",
			"admin",
			validEquipmentType,
			func(t *testing.T) {
				_, err := TestAdminDB.Exec(context.Background(), "INSERT INTO equipment_types (name) VALUES ($1)", validEquipmentType.Name)
				if err != nil {
					t.Fatalf("Failed to insert into test database: %v", err)
				}
			},
			http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestEquipmentTypesServer(t, true)
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
		ts := setupTestEquipmentTypesServer(t, true)
		_ = SeedEquipmentTypes(TestAdminDB)
		return ts, getFirstEquipmentTypeID(t, ts, generateToken(t, "admin"))
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
				ts := setupTestEquipmentTypesServer(t, true)
				_ = SeedEquipmentTypes(TestAdminDB)
				return ts, ""
			},
			http.StatusBadRequest,
		},
		{
			"Invalid UUID as User",
			"ruser",
			"invalid-uuid",
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestEquipmentTypesServer(t, true)
				_ = SeedEquipmentTypes(TestAdminDB)
				return ts, ""
			},
			http.StatusBadRequest,
		},
		{
			"Invalid UUID as Admin",
			"admin",
			"invalid-uuid",
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestEquipmentTypesServer(t, true)
				_ = SeedEquipmentTypes(TestAdminDB)
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
				ts := setupTestEquipmentTypesServer(t, true)
				_ = SeedEquipmentTypes(TestAdminDB)
				return ts, unknown_id
			},
			http.StatusForbidden,
		},
		{
			"Unknown UUID as User",
			"ruser",
			unknown_id,
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestEquipmentTypesServer(t, true)
				return ts, unknown_id
			},
			http.StatusForbidden,
		},
		{
			"Unknown UUID as Admin",
			"admin",
			unknown_id,
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestEquipmentTypesServer(t, true)
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
			"ruser",
			"",
			"invalid-json",
			setupExistingEquipment,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as Admin",
			"admin",
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
			"ruser",
			"",
			EquipmentTypeData{Name: "", Description: "Test"},
			setupExistingEquipment,
			http.StatusBadRequest,
		},
		{
			"Empty Name as Admin",
			"admin",
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
			"ruser",
			"",
			validUpdateData,
			setupExistingEquipment,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			"admin",
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
		ts := setupTestEquipmentTypesServer(t, true)
		_ = SeedEquipmentTypes(TestAdminDB)
		return ts, getFirstEquipmentTypeID(t, ts, generateToken(t, "admin"))
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
				ts := setupTestEquipmentTypesServer(t, true)
				_ = SeedEquipmentTypes(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusForbidden,
		},
		{
			"Not Found as User",
			"ruser",
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestEquipmentTypesServer(t, true)
				_ = SeedEquipmentTypes(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusForbidden,
		},
		{
			"Not Found as Admin",
			"admin",
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestEquipmentTypesServer(t, true)
				_ = SeedEquipmentTypes(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusNotFound,
		},
		{
			"Invalid UUID as Guest",
			"",
			"invalid-uuid",
			func(t *testing.T) (*httptest.Server, string) {
				return setupTestEquipmentTypesServer(t, true), ""
			},
			http.StatusBadRequest,
		},
		{
			"Invalid UUID as User",
			"ruser",
			"invalid-uuid",
			func(t *testing.T) (*httptest.Server, string) {
				return setupTestEquipmentTypesServer(t, true), ""
			},
			http.StatusBadRequest,
		},
		{
			"Invalid UUID as Admin",
			"admin",
			"invalid-uuid",
			func(t *testing.T) (*httptest.Server, string) {
				return setupTestEquipmentTypesServer(t, true), ""
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
			"ruser",
			"",
			setupExistingEquipment,
			http.StatusForbidden,
		},
		{
			"Success as Admin",
			"admin",
			"",
			setupExistingEquipment,
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
