package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestGetHalls(t *testing.T) {
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

				requiredFields := []string{"ID", "Name", "ScreenTypeID", "Description"}
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
		return ts, HallsData[0].ID
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
			os.Getenv("CLAIM_ROLE_USER"),
			http.StatusNotFound,
		},
		{
			"Unknown ID as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			os.Getenv("CLAIM_ROLE_ADMIN"),
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
			os.Getenv("CLAIM_ROLE_USER"),
			http.StatusBadRequest,
		},
		{
			"Invalid ID as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
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
		ScreenTypeID: ScreenTypesData[3].ID,
		Description:  ptr("Test Description"),
	}

	invalidHall := HallData{
		Name:         "",
		ScreenTypeID: "",
		Description:  nil,
	}

	invalidForainKeyHall := HallData{
		Name:         "Test Hall",
		ScreenTypeID: uuid.New().String(),
		Description:  ptr("Test Description"),
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
			os.Getenv("CLAIM_ROLE_USER"),
			validHall,
			nil,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			validHall,
			func(t *testing.T) {
				SeedAll(TestAdminDB)
			},
			http.StatusCreated,
		},
		{
			"Unknown forain key Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
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
			"Empty fields in JSON Guest",
			"",
			invalidHall,
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty fields in JSON User",
			os.Getenv("CLAIM_ROLE_USER"),
			invalidHall,
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty fields in JSON Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			invalidHall,
			nil,
			http.StatusBadRequest,
		},
		{
			"Conflict Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			validHall,
			func(t *testing.T) {
				SeedAll(TestAdminDB)
				_, err := TestAdminDB.Exec(context.Background(),
					"INSERT INTO halls (name, screen_type_id, description) VALUES ($1, $2, $3)",
					validHall.Name, validHall.ScreenTypeID, validHall.Description)
				if err != nil {
					t.Fatalf("Failed to insert into test database: %v", err)
				}
			},
			http.StatusConflict,
		},
		{
			"Capacity 1 as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			HallData{
				Name:         "Min Capacity",
				ScreenTypeID: ScreenTypesData[0].ID,
				Description:  nil,
			},
			func(t *testing.T) { SeedAll(TestAdminDB) },
			http.StatusCreated,
		},
		{
			"Max name length as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			HallData{
				Name:         strings.Repeat("a", 100),
				ScreenTypeID: ScreenTypesData[0].ID,
				Description:  ptr("Test"),
			},
			func(t *testing.T) { SeedAll(TestAdminDB) },
			http.StatusCreated,
		},
		{
			"Name too long as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			HallData{
				Name:         strings.Repeat("a", 101),
				ScreenTypeID: ScreenTypesData[0].ID,
				Description:  nil,
			},
			nil,
			http.StatusBadRequest,
		},
		{
			"Max description length as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			HallData{
				Name:         "Test Hall",
				ScreenTypeID: ScreenTypesData[0].ID,
				Description:  ptr(strings.Repeat("a", 1000)),
			},
			func(t *testing.T) { SeedAll(TestAdminDB) },
			http.StatusCreated,
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
					t.Error("Неверный формат возвращённого UUID")
				}
			}
		})
	}
}

func TestUpdateHall(t *testing.T) {
	validUpdateData := HallData{
		Name:         "Updated Hall",
		ScreenTypeID: ScreenTypesData[6].ID,
		Description:  ptr("Updated Description"),
	}

	invalidUpdateData := HallData{
		Name:         "",
		ScreenTypeID: "",
		Description:  nil,
	}

	// Setup function for tests needing existing hall
	setupExistingHall := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, HallsData[0].ID
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
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			HallData{
				Name:         "Min Capacity",
				ScreenTypeID: ScreenTypesData[0].ID,
				Description:  nil,
			},
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, HallsData[0].ID
			},
			http.StatusOK,
		},
		{
			"Update With Same Data as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				hall := HallsData[0]
				_, err := TestAdminDB.Exec(context.Background(),
					"UPDATE halls SET name=$1, screen_type_id=$2, description=$3 WHERE id=$4",
					validUpdateData.Name,
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
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			HallData{
				Name:         strings.Repeat("a", 100),
				ScreenTypeID: ScreenTypesData[0].ID,
				Description:  ptr("Test"),
			},
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, HallsData[0].ID
			},
			http.StatusOK,
		},
		{
			"Name too long as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			HallData{
				Name:         strings.Repeat("a", 101),
				ScreenTypeID: ScreenTypesData[0].ID,
				Description:  nil,
			},
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, HallsData[0].ID
			},
			http.StatusBadRequest,
		},
		{
			"Max description length as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			HallData{
				Name:         "Test Hall",
				ScreenTypeID: ScreenTypesData[0].ID,
				Description:  ptr(strings.Repeat("a", 1000)),
			},
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, HallsData[0].ID
			},
			http.StatusOK,
		},
		{
			"Invalid UUID as User",
			os.Getenv("CLAIM_ROLE_USER"),
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
			os.Getenv("CLAIM_ROLE_ADMIN"),
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
			os.Getenv("CLAIM_ROLE_USER"),
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
			os.Getenv("CLAIM_ROLE_ADMIN"),
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
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			"invalid-json",
			setupExistingHall,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
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
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			invalidUpdateData,
			setupExistingHall,
			http.StatusBadRequest,
		},
		{
			"Empty fields in as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
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
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			validUpdateData,
			setupExistingHall,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			validUpdateData,
			setupExistingHall,
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
		return ts, HallsData[0].ID
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
			os.Getenv("CLAIM_ROLE_USER"),
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
			os.Getenv("CLAIM_ROLE_ADMIN"),
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
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				hall := HallsData[4]
				req := createRequest(t, "DELETE", ts.URL+"/halls/"+hall.ID, generateToken(t, os.Getenv("CLAIM_ROLE_ADMIN")), nil)
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
			os.Getenv("CLAIM_ROLE_USER"),
			"invalid-uuid",
			func(t *testing.T) (*httptest.Server, string) {
				return setupTestServer(), "invalid-uuid"
			},
			http.StatusBadRequest,
		},
		{
			"Invalid UUID as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
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
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			setupExistingHall,
			http.StatusForbidden,
		},
		{
			"Dependency error as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			setupExistingHall,
			http.StatusConflict,
		},
		{
			"Success as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
				return ts, HallsData[4].ID
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

			req := createRequest(t, "DELETE", ts.URL+"/halls/"+effectiveID, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestGetHallsByScreenType(t *testing.T) {
	setupWithData := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, ScreenTypesData[0].ID
	}

	tests := []struct {
		name           string
		role           string
		screenTypeID   string
		setup          func(t *testing.T) (*httptest.Server, string)
		expectedStatus int
		expectedCount  int
	}{
		{
			"Guest access allowed",
			"",
			"",
			setupWithData,
			http.StatusOK,
			1,
		},
		{
			"User access allowed",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			setupWithData,
			http.StatusOK,
			1,
		},
		{
			"Admin access allowed",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			setupWithData,
			http.StatusOK,
			1,
		},
		{
			"Invalid UUID format",
			os.Getenv("CLAIM_ROLE_USER"),
			"invalid-uuid",
			func(t *testing.T) (*httptest.Server, string) {
				return setupTestServer(), ""
			},
			http.StatusBadRequest,
			0,
		},
		{
			"Non-existent screen type",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusNotFound,
			0,
		},
		{
			"Empty screen type ID",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			func(t *testing.T) (*httptest.Server, string) {
				return setupTestServer(), ""
			},
			http.StatusBadRequest,
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, screenTypeID := tt.setup(t)
			SeedUsers(TestAdminDB)
			defer ts.Close()

			effectiveID := tt.screenTypeID
			if effectiveID == "" {
				effectiveID = screenTypeID
			}

			req := createRequest(t, "GET", ts.URL+"/halls/by-screen-type?screen_type_id="+effectiveID, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var halls []Hall
				if err := json.NewDecoder(resp.Body).Decode(&halls); err != nil {
					t.Fatalf("Could not decode response: %v", err)
				}

				if len(halls) != tt.expectedCount {
					t.Errorf("Expected %d halls, got %d", tt.expectedCount, len(halls))
				}
			}
		})
	}
}

func TestSearchHallsByName(t *testing.T) {
	setupWithData := func(t *testing.T) *httptest.Server {
		ts := setupTestServer()
		SeedUsers(TestAdminDB)
		db := TestAdminDB

		_ = SeedAll(db)
		ClearTable(db, "halls")

		_, err := db.Exec(context.Background(), `
            INSERT INTO halls (id, name, screen_type_id, description)
            VALUES 
                ($1, 'IMAX Premium', (SELECT id FROM screen_types WHERE name = 'IMAX'), 'Премиум зал IMAX'),
                ($2, 'LED Standard', (SELECT id FROM screen_types WHERE name = 'LED'), 'Стандартный LED зал'),
                ($3, '3D Atmos', (SELECT id FROM screen_types WHERE name = 'Система 3D'), 'Зал с 3D и звуком Atmos')
            `,
			uuid.New(), uuid.New(), uuid.New(),
		)
		if err != nil {
			t.Fatal(err)
		}

		return ts
	}

	tests := []struct {
		name           string
		role           string
		query          string
		setup          func(t *testing.T) *httptest.Server
		expectedStatus int
		expectedCount  int
	}{
		{
			"Guest access forbidden",
			"",
			"IMAX",
			setupWithData,
			http.StatusOK,
			1,
		},
		{
			"Search 'IMAX' (user)",
			os.Getenv("CLAIM_ROLE_USER"),
			"IMAX",
			setupWithData,
			http.StatusOK,
			1,
		},
		{
			"Search 'Standard' (admin)",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"Standard",
			setupWithData,
			http.StatusOK,
			1,
		},
		{
			"Search '3D' (partial match)",
			os.Getenv("CLAIM_ROLE_USER"),
			"3D",
			setupWithData,
			http.StatusOK,
			1,
		},
		{
			"Search 'premium' (case insensitive)",
			os.Getenv("CLAIM_ROLE_USER"),
			"premium",
			setupWithData,
			http.StatusOK,
			1,
		},
		{
			"Search non-existent hall",
			os.Getenv("CLAIM_ROLE_USER"),
			"VIP Lounge",
			setupWithData,
			http.StatusNotFound,
			0,
		},
		{
			"Search is short",
			os.Getenv("CLAIM_ROLE_USER"),
			"at",
			setupWithData,
			http.StatusOK,
			1,
		},
		{
			"Search with special chars (#)",
			os.Getenv("CLAIM_ROLE_USER"),
			"#1",
			func(t *testing.T) *httptest.Server {
				ts := setupTestServer()
				db := TestAdminDB
				_ = SeedAll(db)
				ClearTable(db, "halls")
				_, err := db.Exec(context.Background(),
					"INSERT INTO halls (id, name, screen_type_id) VALUES ($1, 'Hall #1', $2)",
					uuid.New(), ScreenTypesData[0].ID)
				if err != nil {
					t.Fatal(err)
				}
				return ts
			},
			http.StatusOK,
			1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := tt.setup(t)
			SeedUsers(TestAdminDB)
			defer ts.Close()

			req := createRequest(t, "GET", ts.URL+"/halls/search?query="+url.QueryEscape(tt.query),
				generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var halls []Hall
				if err := json.NewDecoder(resp.Body).Decode(&halls); err != nil {
					t.Fatalf("Could not decode response: %v", err)
				}

				if len(halls) != tt.expectedCount {
					t.Errorf("Expected %d halls, got %d", tt.expectedCount, len(halls))
				}
			}
		})
	}
}
