package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func getHallByID(t *testing.T, ts *httptest.Server, token string, index int) Hall {
	req := createRequest(t, "GET", ts.URL+"/halls", token, nil)
	resp := executeRequest(t, req, http.StatusOK)
	defer resp.Body.Close()

	var halls []Hall
	parseResponseBody(t, resp, &halls)

	if len(halls) == 0 {
		t.Fatal("Expected at least one hall, got none")
	}

	if index >= len(halls) {
		t.Fatal("Index is greater than length of data array")
	}

	return halls[index]
}

func TestGetHalls(t *testing.T) {
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

			req := createRequest(t, "GET", ts.URL+"/halls", generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var halls []Hall
				parseResponseBody(t, resp, &halls)

				if len(halls) == 0 {
					t.Error("Expected non-empty halls list")
				}

				requiredFields := []string{"ID", "Name", "Capacity", "ScreenTypeID", "Description"}
				for _, hall := range halls {
					v := reflect.ValueOf(hall)
					for _, field := range requiredFields {
						if v.FieldByName(field).IsZero() {
							t.Errorf("Missing or zero field %s in response", field)
						}
					}
				}
			}
		})
	}
}

func TestGetHallByID(t *testing.T) {
	setupValidIDTest := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, getHallByID(t, ts, "", 0).ID
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
		Name:         "Test Hall",
		Capacity:     100,
		ScreenTypeID: ScreenTypesData[3].ID,
		Description:  "Test Description",
	}

	invalidHall := HallData{
		Name:         "",
		Capacity:     0,
		ScreenTypeID: "",
		Description:  "",
	}

	invalidForainKeyHall := HallData{
		Name:         "Test Hall",
		Capacity:     100,
		ScreenTypeID: uuid.New().String(),
		Description:  "Test Description",
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
			"CLAIM_ROLE_USER",
			validHall,
			nil,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			"CLAIM_ROLE_ADMIN",
			validHall,
			func(t *testing.T) {
				SeedAll(TestAdminDB)
			},
			http.StatusCreated,
		},
		{
			"Unknown forain key Admin",
			"CLAIM_ROLE_ADMIN",
			invalidForainKeyHall,
			nil,
			http.StatusConflict,
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
			"Empty fields in JSON Guest",
			"",
			invalidHall,
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty fields in JSON User",
			"CLAIM_ROLE_USER",
			invalidHall,
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty fields in JSON Admin",
			"CLAIM_ROLE_ADMIN",
			invalidHall,
			nil,
			http.StatusBadRequest,
		},
		{
			"Conflict Admin",
			"CLAIM_ROLE_ADMIN",
			validHall,
			func(t *testing.T) {
				SeedAll(TestAdminDB)
				_, err := TestAdminDB.Exec(context.Background(),
					"INSERT INTO halls (name, capacity, screen_type_id, description) VALUES ($1, $2, $3, $4)",
					validHall.Name, validHall.Capacity, validHall.ScreenTypeID, validHall.Description)
				if err != nil {
					t.Fatalf("Failed to insert into test database: %v", err)
				}
			},
			http.StatusConflict,
		},
		{
			"Capacity 1 as Admin",
			"CLAIM_ROLE_ADMIN",
			HallData{
				Name:         "Min Capacity",
				Capacity:     1,
				ScreenTypeID: ScreenTypesData[0].ID,
				Description:  "Test",
			},
			func(t *testing.T) { SeedAll(TestAdminDB) },
			http.StatusCreated,
		},
		{
			"Max name length as Admin",
			"CLAIM_ROLE_ADMIN",
			HallData{
				Name:         strings.Repeat("a", 100),
				Capacity:     100,
				ScreenTypeID: ScreenTypesData[0].ID,
				Description:  "Test",
			},
			func(t *testing.T) { SeedAll(TestAdminDB) },
			http.StatusCreated,
		},
		{
			"Name too long as Admin",
			"CLAIM_ROLE_ADMIN",
			HallData{
				Name:         strings.Repeat("a", 101),
				Capacity:     100,
				ScreenTypeID: ScreenTypesData[0].ID,
				Description:  "Test",
			},
			nil,
			http.StatusBadRequest,
		},
		{
			"Max description length as Admin",
			"CLAIM_ROLE_ADMIN",
			HallData{
				Name:         "Test Hall",
				Capacity:     100,
				ScreenTypeID: ScreenTypesData[0].ID,
				Description:  strings.Repeat("a", 1000),
			},
			func(t *testing.T) { SeedAll(TestAdminDB) },
			http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
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
		Name:         "Updated Hall",
		Capacity:     150,
		ScreenTypeID: ScreenTypesData[7].ID,
		Description:  "Updated Description",
	}

	invalidUpdateData := HallData{
		Name:         "",
		Capacity:     0,
		ScreenTypeID: "",
		Description:  "",
	}

	// Setup function for tests needing existing hall
	setupExistingHall := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, getHallByID(t, ts, "", 0).ID
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
			"Capacity 1 as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			HallData{
				Name:         "Min Capacity",
				Capacity:     1,
				ScreenTypeID: ScreenTypesData[0].ID,
				Description:  "Test",
			},
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, getHallByID(t, ts, generateToken(t, "CLAIM_ROLE_ADMIN"), 0).ID
			},
			http.StatusOK,
		},
		{
			"Update With Same Data as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				hall := getHallByID(t, ts, "", 0)
				_, err := TestAdminDB.Exec(context.Background(),
					"UPDATE halls SET name=$1, capacity=$2, screen_type_id=$3, description=$4 WHERE id=$5",
					validUpdateData.Name, validUpdateData.Capacity,
					validUpdateData.ScreenTypeID, validUpdateData.Description, hall.ID)
				if err != nil {
					t.Fatal(err)
				}
				return ts, hall.ID
			},
			http.StatusOK,
		},
		{
			"Max name length as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			HallData{
				Name:         strings.Repeat("a", 100),
				Capacity:     100,
				ScreenTypeID: ScreenTypesData[0].ID,
				Description:  "Test",
			},
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, getHallByID(t, ts, generateToken(t, "CLAIM_ROLE_ADMIN"), 0).ID
			},
			http.StatusOK,
		},
		{
			"Name too long as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			HallData{
				Name:         strings.Repeat("a", 101),
				Capacity:     100,
				ScreenTypeID: ScreenTypesData[0].ID,
				Description:  "Test",
			},
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, getHallByID(t, ts, generateToken(t, "CLAIM_ROLE_ADMIN"), 0).ID
			},
			http.StatusBadRequest,
		},
		{
			"Max description length as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			HallData{
				Name:         "Test Hall",
				Capacity:     100,
				ScreenTypeID: ScreenTypesData[0].ID,
				Description:  strings.Repeat("a", 1000),
			},
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, getHallByID(t, ts, generateToken(t, "CLAIM_ROLE_ADMIN"), 0).ID
			},
			http.StatusOK,
		},
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
			"Unknown UUID as User",
			"CLAIM_ROLE_USER",
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
			"Invalid JSON as Guest",
			"",
			"",
			"invalid-json",
			setupExistingHall,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as User",
			"CLAIM_ROLE_USER",
			"",
			"invalid-json",
			setupExistingHall,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as Admin",
			"CLAIM_ROLE_ADMIN",
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
			"CLAIM_ROLE_USER",
			"",
			invalidUpdateData,
			setupExistingHall,
			http.StatusBadRequest,
		},
		{
			"Empty fields in as Admin",
			"CLAIM_ROLE_ADMIN",
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
			"CLAIM_ROLE_USER",
			"",
			validUpdateData,
			setupExistingHall,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			"CLAIM_ROLE_ADMIN",
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
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, getHallByID(t, ts, "", 0).ID
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
			"Double Delete as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				hall := getHallByID(t, ts, "", 4)
				req := createRequest(t, "DELETE", ts.URL+"/halls/"+hall.ID, generateToken(t, "CLAIM_ROLE_ADMIN"), nil)
				resp := executeRequest(t, req, http.StatusNoContent)
				resp.Body.Close()
				return ts, hall.ID
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
			setupExistingHall,
			http.StatusForbidden,
		},
		{
			"Forbidden as User",
			"CLAIM_ROLE_USER",
			"",
			setupExistingHall,
			http.StatusForbidden,
		},
		{
			"Dependency error as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			setupExistingHall,
			http.StatusConflict,
		},
		{
			"Success as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
				return ts, getHallByID(t, ts, "", 4).ID
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

			req := createRequest(t, "DELETE", ts.URL+"/halls/"+effectiveID, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}
