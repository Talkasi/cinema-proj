package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
)

func TestGetUsers(t *testing.T) {
	tests := []struct {
		name           string
		seedData       bool
		role           string
		expectedStatus int
	}{
		{"Empty as Guest", false, "", http.StatusNotFound},
		{"Empty as User", false, os.Getenv("CLAIM_ROLE_USER"), http.StatusForbidden},
		{"Empty as Admin", false, os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusForbidden},
		{"NonEmpty as Guest", true, "", http.StatusOK},
		{"NonEmpty as User", true, os.Getenv("CLAIM_ROLE_USER"), http.StatusOK},
		{"NonEmpty as Admin", true, os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
			defer ts.Close()

			if tt.seedData {
				SeedAll(TestAdminDB)
			}

			req := createRequest(t, "GET", ts.URL+"/users", generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var users []User
				parseResponseBody(t, resp, &users)

				if len(users) == 0 {
					t.Error("Expected non-empty users list")
				}
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	setupValidIDTest := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, UsersData[0].ID
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

			req := createRequest(t, "GET", ts.URL+"/users/"+id, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var user User
				parseResponseBody(t, resp, &user)

				if user.ID != id {
					t.Errorf("Expected ID %v; got %v", id, user.ID)
				}
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	validUpdateData := UserData{
		Name:      "Updated Name",
		Email:     "updated@example.com",
		BirthDate: "2020-12-12",
	}

	setupExistingUser := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, UsersData[0].ID
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
			"Unknown UUID as User",
			os.Getenv("CLAIM_ROLE_USER"),
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
			"Invalid JSON as User",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			"invalid-json",
			setupExistingUser,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			"invalid-json",
			setupExistingUser,
			http.StatusBadRequest,
		},
		{
			"Invalid Name",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			UserData{
				Name:      "Name123!",
				Email:     "valid@example.com",
				BirthDate: "2020-12-12",
			},
			setupExistingUser,
			http.StatusBadRequest,
		},
		{
			"Invalid Email",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			UserData{
				Name:      "Valid Name",
				Email:     "invalid-email",
				BirthDate: "2020-12-12",
			},
			setupExistingUser,
			http.StatusBadRequest,
		},
		{
			"Future Birth Date",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			UserData{
				Name:      "Valid Name",
				Email:     "valid@example.com",
				BirthDate: "2030-12-12",
			},
			setupExistingUser,
			http.StatusBadRequest,
		},
		{
			"Too Old Birth Date",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			UserData{
				Name:      "Valid Name",
				Email:     "valid@example.com",
				BirthDate: "1030-12-30",
			},
			setupExistingUser,
			http.StatusBadRequest,
		},
		{
			"Success User",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			validUpdateData,
			setupExistingUser,
			http.StatusOK,
		},
		{
			"Success Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			validUpdateData,
			setupExistingUser,
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
			SeedUsers(TestAdminDB)
			defer ts.Close()

			effectiveID := tt.id
			if effectiveID == "" {
				effectiveID = id
			}

			req := createRequest(t, "PUT", ts.URL+"/users/"+effectiveID, generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name           string
		role           string
		id             string
		setup          func(t *testing.T) (*httptest.Server, string)
		expectedStatus int
	}{
		{
			"Forbidden as User",
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
			"Success as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
				return ts, UsersData[3].ID
			},
			http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
			SeedUsers(TestAdminDB)
			defer ts.Close()

			effectiveID := tt.id
			if effectiveID == "" {
				effectiveID = id
			}

			req := createRequest(t, "DELETE", ts.URL+"/users/"+effectiveID, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestRegisterUser(t *testing.T) {
	validUser := UserRegister{
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: "PasswordHash123",
		BirthDate:    "2020-12-12",
	}

	tests := []struct {
		name           string
		body           interface{}
		setup          func(t *testing.T)
		expectedStatus int
	}{
		{
			"Success",
			validUser,
			nil,
			http.StatusCreated,
		},
		{
			"Invalid JSON",
			"{invalid json}",
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty Name",
			UserRegister{
				Name:         "",
				Email:        "valid@example.com",
				PasswordHash: "PasswordHash123",
				BirthDate:    "2020-12-12",
			},
			nil,
			http.StatusBadRequest,
		},
		{
			"Invalid Email",
			UserRegister{
				Name:         "Valid Name",
				Email:        "invalid-email",
				PasswordHash: "PasswordHash123",
				BirthDate:    "2020-12-12",
			},
			nil,
			http.StatusBadRequest,
		},
		{
			"Short PasswordHash",
			UserRegister{
				Name:         "Valid Name",
				Email:        "valid@example.com",
				PasswordHash: "short",
				BirthDate:    "2020-12-12",
			},
			nil,
			http.StatusBadRequest,
		},
		{
			"Future Birth Date",
			UserRegister{
				Name:         "Valid Name",
				Email:        "valid@example.com",
				PasswordHash: "PasswordHash123",
				BirthDate:    "2030-12-12",
			},
			nil,
			http.StatusBadRequest,
		},
		{
			"Too Old Birth Date",
			UserRegister{
				Name:         "Valid Name",
				Email:        "valid@example.com",
				PasswordHash: "PasswordHash123",
				BirthDate:    "1030-12-12",
			},
			nil,
			http.StatusBadRequest,
		},
		{
			"Duplicate Email",
			validUser,
			func(t *testing.T) {
				_, err := TestAdminDB.Exec(context.Background(),
					"INSERT INTO users (name, email, password_hash, birth_date) VALUES ($1, $2, $3, $4)",
					validUser.Name, validUser.Email, "hash", validUser.BirthDate)
				if err != nil {
					t.Fatalf("Failed to insert into test database: %v", err)
				}
			},
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

			req := createRequest(t, "POST", ts.URL+"/user/register", "", tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusCreated {
				var created string
				parseResponseBody(t, resp, &created)

				if created == "" {
					t.Error("Expected non-empty ID in response")
				}

				if _, err := uuid.Parse(created); err != nil {
					t.Error("Invalid UUID format in response")
				}
			}
		})
	}
}

func TestLoginUser(t *testing.T) {
	testUser := UserRegister{
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: "PasswordHash123",
		BirthDate:    "2020-12-12",
	}

	tests := []struct {
		name           string
		body           interface{}
		setup          func(t *testing.T)
		expectedStatus int
	}{
		{
			"Success",
			UserLogin{
				Email:        testUser.Email,
				PasswordHash: testUser.PasswordHash,
			},
			func(t *testing.T) {

				_, err := TestAdminDB.Exec(context.Background(),
					"INSERT INTO users (name, email, password_hash, birth_date) VALUES ($1, $2, $3, $4)",
					testUser.Name, testUser.Email, testUser.PasswordHash, testUser.BirthDate)
				if err != nil {
					t.Fatalf("Failed to insert into test database: %v", err)
				}
			},
			http.StatusOK,
		},
		{
			"Invalid JSON",
			"{invalid json}",
			nil,
			http.StatusBadRequest,
		},
		{
			"Invalid Email",
			UserLogin{
				Email:        "invalid-email",
				PasswordHash: testUser.PasswordHash,
			},
			nil,
			http.StatusBadRequest,
		},
		{
			"Short PasswordHash",
			UserLogin{
				Email:        testUser.Email,
				PasswordHash: "short",
			},
			nil,
			http.StatusBadRequest,
		},
		{
			"Wrong PasswordHash",
			UserLogin{
				Email:        testUser.Email,
				PasswordHash: "wrongPasswordHash",
			},
			func(t *testing.T) {

				_, err := TestAdminDB.Exec(context.Background(),
					"INSERT INTO users (name, email, password_hash, birth_date) VALUES ($1, $2, $3, $4)",
					testUser.Name, testUser.Email, testUser.PasswordHash, testUser.BirthDate)
				if err != nil {
					t.Fatalf("Failed to insert into test database: %v", err)
				}
			},
			http.StatusUnauthorized,
		},
		{
			"User Not Found",
			UserLogin{
				Email:        "notfound@example.com",
				PasswordHash: testUser.PasswordHash,
			},
			nil,
			http.StatusUnauthorized,
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

			req := createRequest(t, "POST", ts.URL+"/user/login", "", tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var response map[string]string
				parseResponseBody(t, resp, &response)

				if response["token"] == "" {
					t.Error("Expected token in response")
				}
			}
		})
	}
}

func TestCreateUserDBError(t *testing.T) {
	ts := setupTestServer()
	SeedUsers(TestAdminDB)
	defer ts.Close()

	// Create DB error situation
	TestAdminDB.Close()
	TestGuestDB.Close()
	TestUserDB.Close()

	req := createRequest(t, "POST", ts.URL+"/user/register", "", UserRegister{
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: "PasswordHash123",
		BirthDate:    "2020-12-12",
	})
	resp := executeRequest(t, req, http.StatusInternalServerError)
	defer resp.Body.Close()

	if err := InitTestDB(); err != nil {
		log.Fatal("Failed to reconnect to test database: ", err)
	}
}
