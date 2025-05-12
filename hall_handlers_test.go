package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func setupTestHallsServer(t *testing.T, clearTable bool) *httptest.Server {
	if clearTable {
		_ = ClearTable(TestAdminDB, "halls")
	}
	SeedEquipmentTypes(TestAdminDB)
	return httptest.NewServer(NewRouter())
}

func getFirstHallID(t *testing.T, ts *httptest.Server, token string) string {
	req := createRequest(t, "GET", ts.URL+"/halls", token, nil)
	resp := executeRequest(t, req, http.StatusOK)
	defer resp.Body.Close()

	var halls []Hall
	parseResponseBody(t, resp, &halls)

	if len(halls) == 0 {
		t.Fatal("Expected at least one hall, got none")
	}

	return halls[0].ID
}

func TestGetHalls(t *testing.T) {
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
			ts := setupTestHallsServer(t, true)
			defer ts.Close()

			if tt.seedData {
				SeedHalls(TestAdminDB)
			}

			req := createRequest(t, "GET", ts.URL+"/halls", generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var halls []Hall
				parseResponseBody(t, resp, &halls)

				if len(halls) == 0 {
					t.Error("Expected non-empty halls list")
				}
			}
		})
	}
}

func TestGetHallByID(t *testing.T) {
	setupValidIDTest := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestHallsServer(t, true)
		_ = SeedHalls(TestAdminDB)
		return ts, getFirstHallID(t, ts, "")
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
				ts := setupTestHallsServer(t, true)
				SeedHalls(TestAdminDB)
				return ts, uuid.New().String()
			},
			"",
			http.StatusNotFound,
		},
		{
			"Unknown ID as User",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestHallsServer(t, true)
				SeedHalls(TestAdminDB)
				return ts, uuid.New().String()
			},
			"ruser",
			http.StatusNotFound,
		},
		{
			"Unknown ID as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestHallsServer(t, true)
				SeedHalls(TestAdminDB)
				return ts, uuid.New().String()
			},
			"admin",
			http.StatusNotFound,
		},
		{
			"Invalid ID as Guest",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestHallsServer(t, true)
				SeedHalls(TestAdminDB)
				return ts, "invalid-id"
			},
			"",
			http.StatusBadRequest,
		},
		{
			"Invalid ID as User",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestHallsServer(t, true)
				SeedHalls(TestAdminDB)
				return ts, "invalid-id"
			},
			"ruser",
			http.StatusBadRequest,
		},
		{
			"Invalid ID as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestHallsServer(t, true)
				SeedHalls(TestAdminDB)
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

			req := createRequest(t, "GET", ts.URL+"/halls/"+id, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var hall Hall
				parseResponseBody(t, resp, &hall)

				if hall.ID != id {
					t.Errorf("Expected ID %v; got %v", id, hall.ID)
				}
			}
		})
	}
}

func TestCreateHall(t *testing.T) {
	validHall := HallData{
		Name:            "Test Hall",
		Capacity:        100,
		EquipmentTypeID: EquipmentTypesData[7].ID,
		Description:     "Test Description",
	}

	invalidHall := HallData{
		Name:            "",
		Capacity:        0,
		EquipmentTypeID: "",
		Description:     "",
	}

	invalidForainKeyHall := HallData{
		Name:            "Test Hall",
		Capacity:        100,
		EquipmentTypeID: uuid.New().String(),
		Description:     "Test Description",
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
			validHall,
			nil,
			http.StatusForbidden,
		},
		{
			"Forbidden User",
			"ruser",
			validHall,
			nil,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			"admin",
			validHall,
			nil,
			http.StatusCreated,
		},
		{
			"Unknown forain key Admin",
			"admin",
			invalidForainKeyHall,
			nil,
			http.StatusFailedDependency,
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
			"Empty fields in JSON Guest",
			"",
			invalidHall,
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty fields in JSON User",
			"ruser",
			invalidHall,
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty fields in JSON Admin",
			"admin",
			invalidHall,
			nil,
			http.StatusBadRequest,
		},
		{
			"Conflict Admin",
			"admin",
			validHall,
			func(t *testing.T) {
				_, err := TestAdminDB.Exec(context.Background(),
					"INSERT INTO halls (name, capacity, equipment_type_id, description) VALUES ($1, $2, $3, $4)",
					validHall.Name, validHall.Capacity, validHall.EquipmentTypeID, validHall.Description)
				if err != nil {
					t.Fatalf("Failed to insert into test database: %v", err)
				}
			},
			http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestHallsServer(t, true)
			defer ts.Close()

			if tt.setup != nil {
				tt.setup(t)
			}

			req := createRequest(t, "POST", ts.URL+"/halls", generateToken(t, tt.role), tt.body)
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

func TestUpdateHall(t *testing.T) {
	validUpdateData := HallData{
		Name:            "Updated Hall",
		Capacity:        150,
		EquipmentTypeID: EquipmentTypesData[7].ID,
		Description:     "Updated Description",
	}

	invalidUpdateData := HallData{
		Name:            "",
		Capacity:        0,
		EquipmentTypeID: "",
		Description:     "",
	}

	// Setup function for tests needing existing hall
	setupExistingHall := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestHallsServer(t, true)
		_ = SeedHalls(TestAdminDB)
		return ts, getFirstHallID(t, ts, "")
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
				ts := setupTestHallsServer(t, true)
				SeedHalls(TestAdminDB)
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
				ts := setupTestHallsServer(t, true)
				SeedHalls(TestAdminDB)
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
				ts := setupTestHallsServer(t, true)
				SeedHalls(TestAdminDB)
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
				ts := setupTestHallsServer(t, true)
				SeedHalls(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusForbidden,
		},
		{
			"Unknown UUID as User",
			"ruser",
			"",
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestHallsServer(t, true)
				SeedHalls(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusForbidden,
		},
		{
			"Unknown UUID as Admin",
			"admin",
			"",
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestHallsServer(t, true)
				SeedHalls(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusNotFound,
		},
		{
			"Invalid JSON as Guest",
			"",
			"",
			"invalid-json",
			setupExistingHall,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as User",
			"ruser",
			"",
			"invalid-json",
			setupExistingHall,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as Admin",
			"admin",
			"",
			"invalid-json",
			setupExistingHall,
			http.StatusBadRequest,
		},
		{
			"Empty fields in JSON as Guest",
			"",
			"",
			invalidUpdateData,
			setupExistingHall,
			http.StatusBadRequest,
		},
		{
			"Empty fields in JSON as User",
			"ruser",
			"",
			invalidUpdateData,
			setupExistingHall,
			http.StatusBadRequest,
		},
		{
			"Empty fields in as Admin",
			"admin",
			"",
			invalidUpdateData,
			setupExistingHall,
			http.StatusBadRequest,
		},
		{
			"Forbidden Guest",
			"",
			"",
			validUpdateData,
			setupExistingHall,
			http.StatusForbidden,
		},
		{
			"Forbidden User",
			"ruser",
			"",
			validUpdateData,
			setupExistingHall,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			"admin",
			"",
			validUpdateData,
			setupExistingHall,
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

			req := createRequest(t, "PUT", ts.URL+"/halls/"+effectiveID, generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestDeleteHall(t *testing.T) {
	// Setup function for tests needing existing hall
	setupExistingHall := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestHallsServer(t, true)
		_ = SeedHalls(TestAdminDB)
		return ts, getFirstHallID(t, ts, "")
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
				ts := setupTestHallsServer(t, true)
				SeedHalls(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusForbidden,
		},
		{
			"Not Found as User",
			"ruser",
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestHallsServer(t, true)
				SeedHalls(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusForbidden,
		},
		{
			"Not Found as Admin",
			"admin",
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestHallsServer(t, true)
				SeedHalls(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusNotFound,
		},
		{
			"Invalid UUID as Guest",
			"",
			"invalid-uuid",
			func(t *testing.T) (*httptest.Server, string) {
				return setupTestHallsServer(t, true), "invalid-uuid"
			},
			http.StatusBadRequest,
		},
		{
			"Invalid UUID as User",
			"ruser",
			"invalid-uuid",
			func(t *testing.T) (*httptest.Server, string) {
				return setupTestHallsServer(t, true), "invalid-uuid"
			},
			http.StatusBadRequest,
		},
		{
			"Invalid UUID as Admin",
			"admin",
			"invalid-uuid",
			func(t *testing.T) (*httptest.Server, string) {
				return setupTestHallsServer(t, true), "invalid-uuid"
			},
			http.StatusBadRequest,
		},
		{
			"Forbidden as Guest",
			"",
			"",
			setupExistingHall,
			http.StatusForbidden,
		},
		{
			"Forbidden as User",
			"ruser",
			"",
			setupExistingHall,
			http.StatusForbidden,
		},
		{
			"Success as Admin",
			"admin",
			"",
			setupExistingHall,
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

			req := createRequest(t, "DELETE", ts.URL+"/halls/"+effectiveID, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}
