package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
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
		{"Empty as User", false, os.Getenv("CLAIM_ROLE_USER"), http.StatusNotFound},
		{"Empty as Admin", false, os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusNotFound},
		{"NonEmpty as Guest", true, "", http.StatusOK},
		{"NonEmpty as User", true, os.Getenv("CLAIM_ROLE_USER"), http.StatusOK},
		{"NonEmpty as Admin", true, os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusOK},
		{"DBError as Guest", true, "", http.StatusInternalServerError},
		{"DBError as User", true, os.Getenv("CLAIM_ROLE_USER"), http.StatusInternalServerError},
		{"DBError as Admin", true, os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
			SeedUsers(TestAdminDB)
			defer ts.Close()

			if tt.seedData {
				SeedAll(TestAdminDB)
			}

			if strings.Split(tt.name, " ")[0] == "DBError" {
				TestAdminDB.Close()
				TestGuestDB.Close()
				TestUserDB.Close()
				defer InitTestDB()
			}

			req := createRequest(t, "GET", ts.URL+"/seat-types", generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var seatTypes []SeatType
				parseResponseBody(t, resp, &seatTypes)

				if len(seatTypes) == 0 {
					t.Error("Expected non-empty seat types list")
				}
			}
		})
	}
}

func TestGetSeatTypeByID(t *testing.T) {
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
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, SeatTypesData[0].ID
			},
			"",
			http.StatusOK,
		},
		{
			"Valid ID as User",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, SeatTypesData[0].ID
			},
			os.Getenv("CLAIM_ROLE_USER"),
			http.StatusOK,
		},
		{
			"Valid ID as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, SeatTypesData[0].ID
			},
			os.Getenv("CLAIM_ROLE_ADMIN"),
			http.StatusOK,
		},
		{
			"DBError as Guest",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				TestAdminDB.Close()
				TestGuestDB.Close()
				TestUserDB.Close()
				return ts, SeatTypesData[0].ID
			},
			"",
			http.StatusInternalServerError,
		},
		{
			"DBError as User",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				TestAdminDB.Close()
				TestGuestDB.Close()
				TestUserDB.Close()
				return ts, SeatTypesData[0].ID
			},
			os.Getenv("CLAIM_ROLE_USER"),
			http.StatusInternalServerError,
		},
		{
			"DBError as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				TestAdminDB.Close()
				TestGuestDB.Close()
				TestUserDB.Close()
				return ts, SeatTypesData[0].ID
			},
			os.Getenv("CLAIM_ROLE_ADMIN"),
			http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
			defer func() {
				ts.Close()
				if strings.Split(tt.name, " ")[0] == "DBError" {
					InitTestDB()
				}
			}()

			req := createRequest(t, "GET", ts.URL+"/seat-types/"+id, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var seatType SeatType
				parseResponseBody(t, resp, &seatType)

				if seatType.ID != id {
					t.Errorf("Expected ID %v; got %v", id, seatType.ID)
				}
			}
		})
	}
}

func TestSearchSeatTypes(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		role           string
		expectedStatus int
		expectedCount  int
	}{
		{"Empty query as Guest", "", "", http.StatusBadRequest, 0},
		{"Empty query as User", "", os.Getenv("CLAIM_ROLE_USER"), http.StatusBadRequest, 0},
		{"Empty query as Admin", "", os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusBadRequest, 0},
		{"Whitespace query as Guest", "    ", "", http.StatusBadRequest, 0},
		{"Whitespace query as User", "    ", os.Getenv("CLAIM_ROLE_USER"), http.StatusBadRequest, 0},
		{"Whitespace query as Admin", "    ", os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusBadRequest, 0},
		{"Short query as Guest", "С", "", http.StatusOK, 7},
		{"Short query as User", "С", os.Getenv("CLAIM_ROLE_USER"), http.StatusOK, 7},
		{"Short query as Admin", "С", os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusOK, 7},
		{"No matches as Guest", "Кресло для отдыха", "", http.StatusNotFound, 0},
		{"No matches as User", "Кресло для отдыха", os.Getenv("CLAIM_ROLE_USER"), http.StatusNotFound, 0},
		{"No matches as Admin", "Кресло для отдыха", os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusNotFound, 0},
		{"Exact match as Guest", "VIP", "", http.StatusOK, 1},
		{"Exact match as User", "VIP", os.Getenv("CLAIM_ROLE_USER"), http.StatusOK, 1},
		{"Exact match as Admin", "VIP", os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusOK, 1},
		{"Partial match as Guest", "стандарт", "", http.StatusOK, 1},
		{"Partial match as User", "стандарт", os.Getenv("CLAIM_ROLE_USER"), http.StatusOK, 1},
		{"Partial match as Admin", "стандарт", os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusOK, 1},
		{"Case insensitive as Guest", "люКС", "", http.StatusOK, 1},
		{"Case insensitive as User", "люКС", os.Getenv("CLAIM_ROLE_USER"), http.StatusOK, 1},
		{"Case insensitive as Admin", "люКС", os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusOK, 1},
		{"With spaces as Guest", "Кресло с подогревом", "", http.StatusOK, 1},
		{"With spaces as User", "Кресло с подогревом", os.Getenv("CLAIM_ROLE_USER"), http.StatusOK, 1},
		{"With spaces as Admin", "Кресло с подогревом", os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusOK, 1},
		{"Partial with spaces as Guest", "  кресло    ", "", http.StatusOK, 3},
		{"Partial with spaces as User", "  кресло    ", os.Getenv("CLAIM_ROLE_USER"), http.StatusOK, 3},
		{"Partial with spaces as Admin", "  кресло    ", os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusOK, 3},
		{"Special chars as Guest", "VIP/", "", http.StatusNotFound, 0},
		{"Special chars as User", "VIP/", os.Getenv("CLAIM_ROLE_USER"), http.StatusNotFound, 0},
		{"Special chars as Admin", "VIP/", os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusNotFound, 0},
		{"DBError as Guest", "VIP", "", http.StatusInternalServerError, 0},
		{"DBError as User", "VIP", os.Getenv("CLAIM_ROLE_USER"), http.StatusInternalServerError, 0},
		{"DBError as Admin", "VIP", os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusInternalServerError, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			defer ts.Close()

			if strings.Split(tt.name, " ")[0] == "DBError" {
				TestAdminDB.Close()
				TestGuestDB.Close()
				TestUserDB.Close()
				defer InitTestDB()
			}

			req := createRequest(t, "GET", ts.URL+"/seat-types/search?query="+url.QueryEscape(tt.query), generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var seatTypes []SeatType
				parseResponseBody(t, resp, &seatTypes)

				if len(seatTypes) != tt.expectedCount {
					t.Errorf("Expected %d seat types, got %d", tt.expectedCount, len(seatTypes))
				}
			}
		})
	}
}

func TestCreateSeatType(t *testing.T) {
	tests := []struct {
		name           string
		role           string
		body           interface{}
		setup          func(t *testing.T)
		expectedStatus int
	}{
		{"Valid as Guest", "", SeatTypeData{Name: "Test", Description: "Test"}, nil, http.StatusForbidden},
		{"Valid as User", os.Getenv("CLAIM_ROLE_USER"), SeatTypeData{Name: "Test", Description: "Test"}, nil, http.StatusForbidden},
		{"Valid as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), SeatTypeData{Name: "Test", Description: "Test"}, nil, http.StatusCreated},
		{"Invalid JSON as Guest", "", "{invalid json}", nil, http.StatusBadRequest},
		{"Invalid JSON as User", os.Getenv("CLAIM_ROLE_USER"), "{invalid json}", nil, http.StatusBadRequest},
		{"Invalid JSON as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "{invalid json}", nil, http.StatusBadRequest},
		{"Empty fields as Guest", "", SeatTypeData{Name: "", Description: "Test"}, nil, http.StatusBadRequest},
		{"Empty fields as User", os.Getenv("CLAIM_ROLE_USER"), SeatTypeData{Name: "", Description: "Test"}, nil, http.StatusBadRequest},
		{"Empty fields as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), SeatTypeData{Name: "", Description: "Test"}, nil, http.StatusBadRequest},
		{"Duplicate name as Guest", "", SeatTypeData{Name: "VIP", Description: "Test"}, func(t *testing.T) {
			_, err := TestAdminDB.Exec(context.Background(), "INSERT INTO seat_types (name, description) VALUES ($1, $2)", "VIP", "Test")
			if err != nil {
				t.Fatal(err)
			}
		}, http.StatusForbidden},
		{"Duplicate name as User", os.Getenv("CLAIM_ROLE_USER"), SeatTypeData{Name: "VIP", Description: "Test"}, func(t *testing.T) {
			_, err := TestAdminDB.Exec(context.Background(), "INSERT INTO seat_types (name, description) VALUES ($1, $2)", "VIP", "Test")
			if err != nil {
				t.Fatal(err)
			}
		}, http.StatusForbidden},
		{"Duplicate name as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), SeatTypeData{Name: "VIP", Description: "Test"}, func(t *testing.T) {
			_, err := TestAdminDB.Exec(context.Background(), "INSERT INTO seat_types (name, description) VALUES ($1, $2)", "VIP", "Test")
			if err != nil {
				t.Fatal(err)
			}
		}, http.StatusConflict},
		{"Whitespace name as Guest", "", SeatTypeData{Name: "   ", Description: "Test"}, nil, http.StatusBadRequest},
		{"Whitespace name as User", os.Getenv("CLAIM_ROLE_USER"), SeatTypeData{Name: "   ", Description: "Test"}, nil, http.StatusBadRequest},
		{"Whitespace name as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), SeatTypeData{Name: "   ", Description: "Test"}, nil, http.StatusBadRequest},
		{"100 chars name as Guest", "", SeatTypeData{Name: strings.Repeat("a", 100), Description: "Test"}, nil, http.StatusForbidden},
		{"100 chars name as User", os.Getenv("CLAIM_ROLE_USER"), SeatTypeData{Name: strings.Repeat("a", 100), Description: "Test"}, nil, http.StatusForbidden},
		{"100 chars name as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), SeatTypeData{Name: strings.Repeat("a", 100), Description: "Test"}, nil, http.StatusCreated},
		{"101 chars name as Guest", "", SeatTypeData{Name: strings.Repeat("a", 101), Description: "Test"}, nil, http.StatusBadRequest},
		{"101 chars name as User", os.Getenv("CLAIM_ROLE_USER"), SeatTypeData{Name: strings.Repeat("a", 101), Description: "Test"}, nil, http.StatusBadRequest},
		{"101 chars name as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), SeatTypeData{Name: strings.Repeat("a", 101), Description: "Test"}, nil, http.StatusBadRequest},
		{"Hyphen name as Guest", "", SeatTypeData{Name: "VIP-Plus", Description: "Test"}, nil, http.StatusForbidden},
		{"Hyphen name as User", os.Getenv("CLAIM_ROLE_USER"), SeatTypeData{Name: "VIP-Plus", Description: "Test"}, nil, http.StatusForbidden},
		{"Hyphen name as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), SeatTypeData{Name: "VIP-Plus", Description: "Test"}, nil, http.StatusCreated},
		{"Empty desc as Guest", "", SeatTypeData{Name: "Test", Description: ""}, nil, http.StatusBadRequest},
		{"Empty desc as User", os.Getenv("CLAIM_ROLE_USER"), SeatTypeData{Name: "Test", Description: ""}, nil, http.StatusBadRequest},
		{"Empty desc as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), SeatTypeData{Name: "Test", Description: ""}, nil, http.StatusBadRequest},
		{"Whitespace desc as Guest", "", SeatTypeData{Name: "Test", Description: "   "}, nil, http.StatusBadRequest},
		{"Whitespace desc as User", os.Getenv("CLAIM_ROLE_USER"), SeatTypeData{Name: "Test", Description: "   "}, nil, http.StatusBadRequest},
		{"Whitespace desc as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), SeatTypeData{Name: "Test", Description: "   "}, nil, http.StatusBadRequest},
		{"1000 chars desc as Guest", "", SeatTypeData{Name: "Test", Description: strings.Repeat("a", 1000)}, nil, http.StatusForbidden},
		{"1000 chars desc as User", os.Getenv("CLAIM_ROLE_USER"), SeatTypeData{Name: "Test", Description: strings.Repeat("a", 1000)}, nil, http.StatusForbidden},
		{"1000 chars desc as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), SeatTypeData{Name: "Test", Description: strings.Repeat("a", 1000)}, nil, http.StatusCreated},
		{"1001 chars desc as Guest", "", SeatTypeData{Name: "Test", Description: strings.Repeat("a", 1001)}, nil, http.StatusBadRequest},
		{"1001 chars desc as User", os.Getenv("CLAIM_ROLE_USER"), SeatTypeData{Name: "Test", Description: strings.Repeat("a", 1001)}, nil, http.StatusBadRequest},
		{"1001 chars desc as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), SeatTypeData{Name: "Test", Description: strings.Repeat("a", 1001)}, nil, http.StatusBadRequest},
		{"DBError as Guest", "", SeatTypeData{Name: "Test", Description: "Test"}, func(t *testing.T) {
			TestAdminDB.Close()
			TestGuestDB.Close()
			TestUserDB.Close()
		}, http.StatusInternalServerError},
		{"DBError as User", os.Getenv("CLAIM_ROLE_USER"), SeatTypeData{Name: "Test", Description: "Test"}, func(t *testing.T) {
			TestAdminDB.Close()
			TestGuestDB.Close()
			TestUserDB.Close()
		}, http.StatusInternalServerError},
		{"DBError as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), SeatTypeData{Name: "Test", Description: "Test"}, func(t *testing.T) {
			TestAdminDB.Close()
			TestGuestDB.Close()
			TestUserDB.Close()
		}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
			SeedUsers(TestAdminDB)
			defer func() {
				ts.Close()
				if strings.Split(tt.name, " ")[0] == "DBError" {
					InitTestDB()
				}
			}()

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
					t.Error("Expected non-empty ID")
				}

				if _, err := uuid.Parse(created); err != nil {
					t.Error("Неверный формат возвращённого UUID")
				}
			}
		})
	}
}

func TestUpdateSeatType(t *testing.T) {
	validTestPreparator := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		SeedAll(TestAdminDB)
		return ts, SeatTypesData[0].ID
	}

	invalidTestPreparator := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		SeedAll(TestAdminDB)
		return ts, ""
	}

	unknownTestPreparator := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		SeedAll(TestAdminDB)
		return ts, uuid.New().String()
	}

	tests := []struct {
		name           string
		role           string
		id             string
		body           interface{}
		setup          func(t *testing.T) (*httptest.Server, string)
		expectedStatus int
	}{
		{"Invalid ID as Guest", "", "invalid-uuid", SeatTypeData{Name: "Test", Description: "Test"}, invalidTestPreparator, http.StatusBadRequest},
		{"Invalid ID as User", os.Getenv("CLAIM_ROLE_USER"), "invalid-uuid", SeatTypeData{Name: "Test", Description: "Test"}, invalidTestPreparator, http.StatusBadRequest},
		{"Invalid ID as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "invalid-uuid", SeatTypeData{Name: "Test", Description: "Test"}, invalidTestPreparator, http.StatusBadRequest},
		{"Unknown ID as Guest", "", "", SeatTypeData{Name: "Test", Description: "Test"}, unknownTestPreparator, http.StatusForbidden},
		{"Unknown ID as User", os.Getenv("CLAIM_ROLE_USER"), "", SeatTypeData{Name: "Test", Description: "Test"}, unknownTestPreparator, http.StatusForbidden},
		{"Unknown ID as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", SeatTypeData{Name: "Test", Description: "Test"}, unknownTestPreparator, http.StatusNotFound},
		{"Invalid JSON as Guest", "", "", "invalid-json", validTestPreparator, http.StatusBadRequest},
		{"Invalid JSON as User", os.Getenv("CLAIM_ROLE_USER"), "", "invalid-json", validTestPreparator, http.StatusBadRequest},
		{"Invalid JSON as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", "invalid-json", validTestPreparator, http.StatusBadRequest},
		{"Duplicate name as Guest", "", "", SeatTypeData{Name: "VIP_", Description: "Test"}, func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			_, err := TestAdminDB.Exec(context.Background(), "INSERT INTO seat_types (name, description) VALUES ($1, $2)", "VIP_", "Test")
			if err != nil {
				t.Fatal(err)
			}
			return ts, SeatTypesData[0].ID
		}, http.StatusForbidden},
		{"Duplicate name as User", os.Getenv("CLAIM_ROLE_USER"), "", SeatTypeData{Name: "VIP_", Description: "Test"}, func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			_, err := TestAdminDB.Exec(context.Background(), "INSERT INTO seat_types (name, description) VALUES ($1, $2)", "VIP_", "Test")
			if err != nil {
				t.Fatal(err)
			}
			return ts, SeatTypesData[0].ID
		}, http.StatusForbidden},
		{"Duplicate name as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", SeatTypeData{Name: "VIP_", Description: "Test"}, func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			_, err := TestAdminDB.Exec(context.Background(), "INSERT INTO seat_types (name, description) VALUES ($1, $2)", "VIP_", "Test")
			if err != nil {
				t.Fatal(err)
			}
			return ts, SeatTypesData[0].ID
		}, http.StatusConflict},
		{"Empty fields as Guest", "", "", SeatTypeData{Name: "", Description: "Test"}, validTestPreparator, http.StatusBadRequest},
		{"Empty fields as User", os.Getenv("CLAIM_ROLE_USER"), "", SeatTypeData{Name: "", Description: "Test"}, validTestPreparator, http.StatusBadRequest},
		{"Empty fields as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", SeatTypeData{Name: "", Description: "Test"}, validTestPreparator, http.StatusBadRequest},
		{"Valid as Guest", "", "", SeatTypeData{Name: "Test", Description: "Test"}, validTestPreparator, http.StatusForbidden},
		{"Valid as User", os.Getenv("CLAIM_ROLE_USER"), "", SeatTypeData{Name: "Test", Description: "Test"}, validTestPreparator, http.StatusForbidden},
		{"Valid as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", SeatTypeData{Name: "Test", Description: "Test"}, validTestPreparator, http.StatusOK},
		{"Whitespace name as Guest", "", "", SeatTypeData{Name: "   ", Description: "Test"}, validTestPreparator, http.StatusBadRequest},
		{"Whitespace name as User", os.Getenv("CLAIM_ROLE_USER"), "", SeatTypeData{Name: "   ", Description: "Test"}, validTestPreparator, http.StatusBadRequest},
		{"Whitespace name as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", SeatTypeData{Name: "   ", Description: "Test"}, validTestPreparator, http.StatusBadRequest},
		{"100 chars name as Guest", "", "", SeatTypeData{Name: strings.Repeat("a", 100), Description: "Test"}, validTestPreparator, http.StatusForbidden},
		{"100 chars name as User", os.Getenv("CLAIM_ROLE_USER"), "", SeatTypeData{Name: strings.Repeat("a", 100), Description: "Test"}, validTestPreparator, http.StatusForbidden},
		{"100 chars name as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", SeatTypeData{Name: strings.Repeat("a", 100), Description: "Test"}, validTestPreparator, http.StatusOK},
		{"101 chars name as Guest", "", "", SeatTypeData{Name: strings.Repeat("a", 101), Description: "Test"}, validTestPreparator, http.StatusBadRequest},
		{"101 chars name as User", os.Getenv("CLAIM_ROLE_USER"), "", SeatTypeData{Name: strings.Repeat("a", 101), Description: "Test"}, validTestPreparator, http.StatusBadRequest},
		{"101 chars name as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", SeatTypeData{Name: strings.Repeat("a", 101), Description: "Test"}, validTestPreparator, http.StatusBadRequest},
		{"Hyphen name as Guest", "", "", SeatTypeData{Name: "VIP-Plus", Description: "Test"}, validTestPreparator, http.StatusForbidden},
		{"Hyphen name as User", os.Getenv("CLAIM_ROLE_USER"), "", SeatTypeData{Name: "VIP-Plus", Description: "Test"}, validTestPreparator, http.StatusForbidden},
		{"Hyphen name as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", SeatTypeData{Name: "VIP-Plus", Description: "Test"}, validTestPreparator, http.StatusOK},
		{"Empty desc as Guest", "", "", SeatTypeData{Name: "Test", Description: ""}, validTestPreparator, http.StatusBadRequest},
		{"Empty desc as User", os.Getenv("CLAIM_ROLE_USER"), "", SeatTypeData{Name: "Test", Description: ""}, validTestPreparator, http.StatusBadRequest},
		{"Empty desc as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", SeatTypeData{Name: "Test", Description: ""}, validTestPreparator, http.StatusBadRequest},
		{"Whitespace desc as Guest", "", "", SeatTypeData{Name: "Test", Description: "   "}, validTestPreparator, http.StatusBadRequest},
		{"Whitespace desc as User", os.Getenv("CLAIM_ROLE_USER"), "", SeatTypeData{Name: "Test", Description: "   "}, validTestPreparator, http.StatusBadRequest},
		{"Whitespace desc as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", SeatTypeData{Name: "Test", Description: "   "}, validTestPreparator, http.StatusBadRequest},
		{"1000 chars desc as Guest", "", "", SeatTypeData{Name: "Test", Description: strings.Repeat("a", 1000)}, validTestPreparator, http.StatusForbidden},
		{"1000 chars desc as User", os.Getenv("CLAIM_ROLE_USER"), "", SeatTypeData{Name: "Test", Description: strings.Repeat("a", 1000)}, validTestPreparator, http.StatusForbidden},
		{"1000 chars desc as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", SeatTypeData{Name: "Test", Description: strings.Repeat("a", 1000)}, validTestPreparator, http.StatusOK},
		{"1001 chars desc as Guest", "", "", SeatTypeData{Name: "Test", Description: strings.Repeat("a", 1001)}, validTestPreparator, http.StatusBadRequest},
		{"1001 chars desc as User", os.Getenv("CLAIM_ROLE_USER"), "", SeatTypeData{Name: "Test", Description: strings.Repeat("a", 1001)}, validTestPreparator, http.StatusBadRequest},
		{"1001 chars desc as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", SeatTypeData{Name: "Test", Description: strings.Repeat("a", 1001)}, validTestPreparator, http.StatusBadRequest},
		{"DBError as Guest", "", "", SeatTypeData{Name: "Test", Description: "Test"}, func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			TestAdminDB.Close()
			TestGuestDB.Close()
			TestUserDB.Close()
			return ts, SeatTypesData[0].ID
		}, http.StatusInternalServerError},
		{"DBError as User", os.Getenv("CLAIM_ROLE_USER"), "", SeatTypeData{Name: "Test", Description: "Test"}, func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			TestAdminDB.Close()
			TestGuestDB.Close()
			TestUserDB.Close()
			return ts, SeatTypesData[0].ID
		}, http.StatusInternalServerError},
		{"DBError as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", SeatTypeData{Name: "Test", Description: "Test"}, func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			TestAdminDB.Close()
			TestGuestDB.Close()
			TestUserDB.Close()
			return ts, SeatTypesData[0].ID
		}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
			SeedUsers(TestAdminDB)
			defer func() {
				ts.Close()
				if strings.Split(tt.name, " ")[0] == "DBError" {
					InitTestDB()
				}
			}()

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
	tests := []struct {
		name           string
		role           string
		id             string
		setup          func(t *testing.T) (*httptest.Server, string)
		expectedStatus int
	}{
		{"Unknown ID as Guest", "", "", func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			return ts, uuid.New().String()
		}, http.StatusForbidden},
		{"Unknown ID as User", os.Getenv("CLAIM_ROLE_USER"), "", func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			return ts, uuid.New().String()
		}, http.StatusForbidden},
		{"Unknown ID as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			return ts, uuid.New().String()
		}, http.StatusNotFound},
		{"Invalid ID as Guest", "", "invalid-uuid", func(t *testing.T) (*httptest.Server, string) {
			return setupTestServer(), "invalid-uuid"
		}, http.StatusBadRequest},
		{"Invalid ID as User", os.Getenv("CLAIM_ROLE_USER"), "invalid-uuid", func(t *testing.T) (*httptest.Server, string) {
			return setupTestServer(), "invalid-uuid"
		}, http.StatusBadRequest},
		{"Invalid ID as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "invalid-uuid", func(t *testing.T) (*httptest.Server, string) {
			return setupTestServer(), "invalid-uuid"
		}, http.StatusBadRequest},
		{"Valid as Guest", "", "", func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			return ts, SeatTypesData[0].ID
		}, http.StatusForbidden},
		{"Valid as User", os.Getenv("CLAIM_ROLE_USER"), "", func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			return ts, SeatTypesData[0].ID
		}, http.StatusForbidden},
		{"Valid as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			return ts, SeatTypesData[5].ID
		}, http.StatusNoContent},
		{"DBError as Guest", "", "", func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			TestAdminDB.Close()
			TestGuestDB.Close()
			TestUserDB.Close()
			return ts, SeatTypesData[5].ID
		}, http.StatusInternalServerError},
		{"DBError as User", os.Getenv("CLAIM_ROLE_USER"), "", func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			TestAdminDB.Close()
			TestGuestDB.Close()
			TestUserDB.Close()
			return ts, SeatTypesData[5].ID
		}, http.StatusInternalServerError},
		{"DBError as Admin", os.Getenv("CLAIM_ROLE_ADMIN"), "", func(t *testing.T) (*httptest.Server, string) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			TestAdminDB.Close()
			TestGuestDB.Close()
			TestUserDB.Close()
			return ts, SeatTypesData[5].ID
		}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
			SeedUsers(TestAdminDB)
			defer func() {
				ts.Close()
				if strings.Split(tt.name, " ")[0] == "DBError" {
					InitTestDB()
				}
			}()

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
