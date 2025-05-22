package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestGetUsers(t *testing.T) {
	tests := []struct {
		name           string
		seedData       bool
		role           string
		expectedStatus int
	}{
		{"Empty as Guest", false, "", http.StatusForbidden},
		{"Empty as User", false, os.Getenv("CLAIM_ROLE_USER"), http.StatusForbidden},
		{"Empty as Admin", false, os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusForbidden},
		{"NonEmpty as Guest", true, "", http.StatusForbidden},
		{"NonEmpty as User", true, os.Getenv("CLAIM_ROLE_USER"), http.StatusForbidden},
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
	setupValidIDTestRoleUser := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, UsersData[len(UsersData)-1].ID
	}

	setupValidIDTestRoleAdmin := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, UsersData[len(UsersData)-2].ID
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
			http.StatusForbidden,
		},
		{
			"Unknown ID as User",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			os.Getenv("CLAIM_ROLE_USER"),
			http.StatusForbidden,
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
			setupValidIDTestRoleUser,
			"",
			http.StatusForbidden,
		},
		{
			"Self ID as User",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
				return ts, UsersData[len(UsersData)-1].ID
			},
			os.Getenv("CLAIM_ROLE_USER"),
			http.StatusOK,
		},
		{
			"Others ID as User",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
				return ts, UsersData[0].ID
			},
			os.Getenv("CLAIM_ROLE_USER"),
			http.StatusForbidden,
		},
		{
			"Valid ID as Admin",
			setupValidIDTestRoleAdmin,
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

	setupExistingUserRoleAdmin := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, UsersData[len(UsersData)-2].ID
	}

	setupExistingUserRoleUser := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, UsersData[len(UsersData)-1].ID
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
			"Invalid JSON as User",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			"invalid-json",
			setupExistingUserRoleUser,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			"invalid-json",
			setupExistingUserRoleAdmin,
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
			setupExistingUserRoleUser,
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
			setupExistingUserRoleUser,
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
			setupExistingUserRoleUser,
			http.StatusBadRequest,
		},
		{
			"Success self User",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			validUpdateData,
			setupExistingUserRoleUser,
			http.StatusOK,
		},
		{
			"Forbidden others User",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
				return ts, UsersData[0].ID
			},
			http.StatusForbidden,
		},
		{
			"Forbidden Guest",
			"",
			"",
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
				return ts, UsersData[0].ID
			},
			http.StatusForbidden,
		},
		{
			"Success Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			validUpdateData,
			setupExistingUserRoleAdmin,
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
				return ts, UsersData[len(UsersData)-3].ID
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

func benchmarkLoginWithoutIndex(db *pgxpool.Pool, b *testing.B, userCount int, concurrentRequests int) {
	_, err := db.Exec(context.Background(), "DROP INDEX IF EXISTS idx_users_email;")
	if err != nil {
		println(err.Error())
	}
	err = ClearAll(db)
	if err != nil {
		b.Errorf("Clear" + err.Error())
	}

	rand.Seed(time.Now().UnixNano())
	start := rand.Intn(1000)

	for i := start; i < start+userCount; i++ {
		_, err := db.Exec(context.Background(),
			"INSERT INTO users (name, email, password_hash, birth_date) VALUES ($1, $2, $3, $4)",
			fmt.Sprintf("User%d", i), fmt.Sprintf("user%d@example.com", i), "PasswordHash123", "2000-01-01")
		if err != nil {
			b.Fatalf("Failed to insert user: %v", err)
		}
	}

	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	body := UserLogin{fmt.Sprintf("user%d@example.com", start), "PasswordHash123"}
	b.ResetTimer()

	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(concurrentRequests)
		for j := 0; j < concurrentRequests; j++ {
			go func() {
				defer wg.Done()
				jsonData, err := json.Marshal(body)
				if err != nil {
					b.Errorf("Failed to marshal JSON: %v", err)
					return
				}
				buf := bytes.NewBuffer(jsonData)

				req, err := http.NewRequest("POST", ts.URL+"/user/login", buf)
				if err != nil {
					b.Errorf("Failed to create request: %v", err)
					return
				}
				req.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					b.Errorf("Failed to perform request: %v", err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					b.Errorf("Unexpected status code: %d", resp.StatusCode)
				}
			}()
		}
		wg.Wait()
	}
}
func benchmarkLoginWithIndex(db *pgxpool.Pool, b *testing.B, userCount int, concurrentRequests int) {
	_, err := db.Exec(context.Background(), `
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`)
	if err != nil {
		b.Fatalf("Failed to create table with index: %v", err)
	}
	err = ClearAll(db)
	if err != nil {
		b.Errorf("Clear" + err.Error())
	}

	rand.Seed(time.Now().UnixNano())
	start := rand.Intn(1000)

	for i := start; i < start+userCount; i++ {
		_, err := db.Exec(context.Background(),
			"INSERT INTO users (name, email, password_hash, birth_date) VALUES ($1, $2, $3, $4)",
			fmt.Sprintf("User%d", i), fmt.Sprintf("user%d@example.com", i), "PasswordHash123", "2000-01-01")
		if err != nil {
			b.Fatalf("Failed to insert user: %v", err)
		}
	}

	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	body := UserLogin{fmt.Sprintf("user%d@example.com", start), "PasswordHash123"}
	b.ResetTimer()

	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(concurrentRequests)
		for j := 0; j < concurrentRequests; j++ {
			go func() {
				defer wg.Done()
				jsonData, err := json.Marshal(body)
				if err != nil {
					b.Errorf("Failed to marshal JSON: %v", err)
					return
				}

				req, err := http.NewRequest("POST", ts.URL+"/user/login", bytes.NewBuffer(jsonData))
				if err != nil {
					b.Errorf("Failed to create request: %v", err)
					return
				}
				req.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					b.Errorf("Failed to perform request: %v", err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					b.Errorf("Unexpected status code: %d", resp.StatusCode)
				}
			}()
		}
		wg.Wait()
	}
}

func BenchmarkLoginWithoutIndexRange(b *testing.B) {
	r := []int{100, 300, 500, 70000, 3000, 5000, 7000000}
	for _, i := range r {
		b.Run(fmt.Sprintf("Users_%d", i), func(b *testing.B) {
			benchmarkLoginWithoutIndex(TestAdminDB, b, i, 100)
		})
	}
}

func BenchmarkLoginWithIndexRange(b *testing.B) {
	r := []int{100, 300, 500, 70000, 3000, 5000, 7000000}
	for _, i := range r {
		b.Run(fmt.Sprintf("Users_%d", i), func(b *testing.B) {
			benchmarkLoginWithIndex(TestAdminDB, b, i, 100)
		})
	}
}

func BenchmarkLoginAllRange(b *testing.B) {
	BenchmarkLoginWithoutIndexRange(b)
	BenchmarkLoginWithIndexRange(b)
}

func benchmarkLogin(db *pgxpool.Pool, b *testing.B, userCount int, concurrentRequests int) {
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        concurrentRequests,
			MaxIdleConnsPerHost: concurrentRequests,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	_, err := db.Exec(context.Background(), "TRUNCATE TABLE users CASCADE")
	if err != nil {
		b.Fatalf("Failed to truncate table: %v", err)
	}

	rand.Seed(time.Now().UnixNano())
	start := rand.Intn(1000)

	for i := start; i < start+userCount; i++ {
		_, err := db.Exec(context.Background(),
			"INSERT INTO users (name, email, password_hash, birth_date) VALUES ($1, $2, $3, $4)",
			fmt.Sprintf("User%d", i), fmt.Sprintf("user%d@example.com", i), "PasswordHash123", "2000-01-01")
		if err != nil {
			b.Fatalf("Failed to insert user: %v", err)
		}
	}

	testEmail := "testuser@example.com"
	testPassword := "validPassword123"
	_, err = db.Exec(context.Background(),
		"INSERT INTO users (name, email, password_hash, birth_date) VALUES ($1, $2, $3, $4)",
		"Test User", testEmail, testPassword, "2000-01-01")
	if err != nil {
		b.Fatalf("Failed to insert test user: %v", err)
	}

	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	loginData := UserLogin{
		Email:        testEmail,
		PasswordHash: testPassword,
	}
	jsonData, err := json.Marshal(loginData)
	if err != nil {
		b.Fatalf("Failed to marshal JSON: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(concurrentRequests)

		for j := 0; j < concurrentRequests; j++ {
			go func() {
				defer wg.Done()

				buf := bytes.NewBuffer(jsonData)
				req, err := http.NewRequest("POST", ts.URL+"/user/login", buf)
				if err != nil {
					b.Errorf("Failed to create request: %v", err)
					return
				}
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				if err != nil {
					b.Errorf("Failed to perform request: %v", err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					body, _ := io.ReadAll(resp.Body)
					b.Errorf("Unexpected status code: %d, body: %s", resp.StatusCode, string(body))
					return
				}
			}()
		}
		wg.Wait()
	}
}

func BenchmarkLoginWithIndex(b *testing.B) {
	db := setupTestDB(b, true)
	defer db.Close()

	benchmarks := []struct {
		name    string
		users   int
		workers int
	}{
		{"Users_10", 10, 10},
		{"Users_100", 100, 10},
		{"Users_300", 300, 10},
		{"Users_500", 500, 10},
		{"Users_700", 700, 10},
		{"Users_1000", 1000, 10},
		{"Users_3000", 3000, 10},
		{"Users_5000", 5000, 10},
		{"Users_7000", 7000, 10},
		{"Users_10000", 10000, 10},
	}

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			benchmarkLogin(db, b, bb.users, bb.workers)
		})
	}
}

func BenchmarkLoginWithoutIndex(b *testing.B) {
	db := setupTestDB(b, false)
	defer db.Close()

	benchmarks := []struct {
		name    string
		users   int
		workers int
	}{
		{"Users_10", 10, 10},
		{"Users_100", 100, 10},
		{"Users_300", 300, 10},
		{"Users_500", 500, 10},
		{"Users_700", 700, 10},
		{"Users_1000", 1000, 10},
		{"Users_3000", 3000, 10},
		{"Users_5000", 5000, 10},
		{"Users_7000", 7000, 10},
		{"Users_10000", 10000, 10},
	}

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			benchmarkLogin(db, b, bb.users, bb.workers)
		})
	}
}

func setupTestDB(b *testing.B, withIndex bool) *pgxpool.Pool {
	db, err := pgxpool.New(context.Background(), fmt.Sprintf("user=%s dbname=%s password=%s sslmode=disable",
		os.Getenv("TEST_ADMIN_USER"),
		os.Getenv("TEST_DB_NAME"),
		os.Getenv("TEST_ADMIN_PASSWORD")))
	if err != nil {
		b.Fatalf("Failed to connect to database: %v", err)
	}

	_, err = db.Exec(context.Background(), "DROP INDEX IF EXISTS idx_users_email;")
	if err != nil {
		b.Error(err)
	}

	if withIndex {
		_, err = db.Exec(context.Background(), "CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)")
		if err != nil {
			b.Fatalf("Failed to create index: %v", err)
		}
	}

	return db
}

func benchmarkLoginNEW(db *pgxpool.Pool, b *testing.B, concurrentRequests int) {
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var wg sync.WaitGroup
			errCh := make(chan error, concurrentRequests)

			for j := 0; j < concurrentRequests; j++ {
				wg.Add(1)
				go func(j int) {
					defer wg.Done()

					loginData := UserLogin{
						Email:        fmt.Sprintf("user%d@example.com", j%1000000),
						PasswordHash: "PasswordHash123",
					}

					jsonData, err := json.Marshal(loginData)
					if err != nil {
						errCh <- fmt.Errorf("marshal error: %v", err)
						return
					}

					req, err := http.NewRequest("POST", ts.URL+"/user/login", bytes.NewBuffer(jsonData))
					if err != nil {
						errCh <- fmt.Errorf("create request error: %v", err)
						return
					}
					req.Header.Set("Content-Type", "application/json")

					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						errCh <- fmt.Errorf("request error: %v", err)
						return
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						body, _ := io.ReadAll(resp.Body)
						errCh <- fmt.Errorf("status %d, body: %s", resp.StatusCode, string(body))
						return
					}
				}(j)
			}

			wg.Wait()
			close(errCh)

			for err := range errCh {
				b.Error(err)
			}
		}
	})
}

func BenchmarkLoginNew(b *testing.B) {
	db, err := pgxpool.New(context.Background(), fmt.Sprintf("user=%s dbname=%s password=%s sslmode=disable",
		os.Getenv("TEST_ADMIN_USER"),
		os.Getenv("TEST_DB_NAME"),
		os.Getenv("TEST_ADMIN_PASSWORD")))
	if err != nil {
		b.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(context.Background(), "DROP INDEX IF EXISTS idx_users_email;")
	if err != nil {
		b.Error(err)
	}

	// Очистка и подготовка данных
	_, err = db.Exec(context.Background(), "TRUNCATE TABLE users CASCADE")
	if err != nil {
		b.Fatalf("Failed to truncate table: %v", err)
	}

	// Вставка данных пачками (batch insert)
	batchSize := 10000
	for i := 0; i < 1000000; i += batchSize {
		batch := make([]string, 0, batchSize)
		for j := 0; j < batchSize && i+j < 1000000; j++ {
			batch = append(batch, fmt.Sprintf("('User%d', 'user%d@example.com', 'PasswordHash123', '2000-01-01')", i+j, i+j))
		}

		_, err = db.Exec(context.Background(),
			"INSERT INTO users (name, email, password_hash, birth_date) VALUES "+strings.Join(batch, ","))
		if err != nil {
			b.Fatalf("Failed to insert users: %v", err)
		}
	}

	benchmarks := []struct {
		name  string
		users int
	}{
		{"Users_10", 10},
		{"Users_100", 100},
		{"Users_300", 300},
		{"Users_500", 500},
		{"Users_700", 700},
		{"Users_1000", 1000},
	}

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			benchmarkLoginNEW(db, b, bb.users)
		})
	}
}

func BenchmarkLoginNewIndex(b *testing.B) {
	db, err := pgxpool.New(context.Background(), fmt.Sprintf("user=%s dbname=%s password=%s sslmode=disable",
		os.Getenv("TEST_ADMIN_USER"),
		os.Getenv("TEST_DB_NAME"),
		os.Getenv("TEST_ADMIN_PASSWORD")))
	if err != nil {
		b.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(context.Background(), "CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)")
	if err != nil {
		b.Fatalf("Failed to create index: %v", err)
	}

	// Очистка и подготовка данных
	_, err = db.Exec(context.Background(), "TRUNCATE TABLE users CASCADE")
	if err != nil {
		b.Fatalf("Failed to truncate table: %v", err)
	}

	// Вставка данных пачками (batch insert)
	batchSize := 10000
	for i := 0; i < 1000000; i += batchSize {
		batch := make([]string, 0, batchSize)
		for j := 0; j < batchSize && i+j < 1000000; j++ {
			batch = append(batch, fmt.Sprintf("('User%d', 'user%d@example.com', 'PasswordHash123', '2000-01-01')", i+j, i+j))
		}

		_, err = db.Exec(context.Background(),
			"INSERT INTO users (name, email, password_hash, birth_date) VALUES "+strings.Join(batch, ","))
		if err != nil {
			b.Fatalf("Failed to insert users: %v", err)
		}
	}

	benchmarks := []struct {
		name  string
		users int
	}{
		{"Users_10", 10},
		{"Users_100", 100},
		{"Users_300", 300},
		{"Users_500", 500},
		{"Users_700", 700},
		{"Users_1000", 1000},
	}

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			benchmarkLoginNEW(db, b, bb.users)
		})
	}
}

func setupTestDBNEW(b *testing.B, withIndex bool, n int) *pgxpool.Pool {
	db, err := pgxpool.New(context.Background(), fmt.Sprintf("user=%s dbname=%s password=%s sslmode=disable",
		os.Getenv("TEST_ADMIN_USER"),
		os.Getenv("TEST_DB_NAME"),
		os.Getenv("TEST_ADMIN_PASSWORD")))
	if err != nil {
		b.Fatalf("Failed to connect to database: %v", err)
	}

	_, err = db.Exec(context.Background(), "DROP INDEX IF EXISTS idx_users_email;")
	if err != nil {
		b.Error(err)
	}

	if withIndex {
		_, err = db.Exec(context.Background(), "CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)")
		if err != nil {
			b.Fatalf("Failed to create index: %v", err)
		}
	}

	_, err = db.Exec(context.Background(), "TRUNCATE TABLE users CASCADE")
	if err != nil {
		b.Fatalf("Failed to truncate table: %v", err)
	}

	batchSize := 10000
	for i := 0; i < n; i += batchSize {
		batch := make([]string, 0, batchSize)
		for j := 0; j < batchSize && i+j < n; j++ {
			batch = append(batch, fmt.Sprintf("('User%d', 'user%d@example.com', 'PasswordHash123', '2000-01-01')", i+j, i+j))
		}

		_, err = db.Exec(context.Background(),
			"INSERT INTO users (name, email, password_hash, birth_date) VALUES "+strings.Join(batch, ","))
		if err != nil {
			b.Fatalf("Failed to insert users: %v", err)
		}
	}

	return db
}

func BenchmarkDBSingleQuery(b *testing.B) {
	db := setupTestDBNEW(b, true, 1_000_000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var user struct {
			ID           string
			PasswordHash string
		}

		email := fmt.Sprintf("user%d@example.com", i%1_000_000)

		err := db.QueryRow(context.Background(),
			"SELECT id, password_hash FROM users WHERE email = $1", email).
			Scan(&user.ID, &user.PasswordHash)

		if err != nil {
			b.Fatalf("Query failed: %v", err)
		}
	}
}

func BenchmarkDBWithConcurrency(b *testing.B) {
	db := setupTestDBNEW(b, false, 1_000_000)
	defer db.Close()

	benchmarks := []struct {
		name    string
		workers int
	}{
		{"10_workers", 10},
		{"200_workers", 200},
		{"400_workers", 400},
		{"600_workers", 600},
		{"800_workers", 800},
		{"1000_workers", 1000},
		{"1200_workers", 1200},
		{"1400_workers", 1400},
		{"1600_workers", 1600},
		{"1800_workers", 1800},
		{"2000_workers", 2000},
	}

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			b.SetParallelism(bb.workers)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					i++
					email := fmt.Sprintf("user%d@example.com", i%1_000_000)

					var user struct {
						ID           string
						PasswordHash string
					}

					err := db.QueryRow(context.Background(),
						"SELECT id, password_hash FROM users WHERE email = $1", email).
						Scan(&user.ID, &user.PasswordHash)

					if err != nil {
						b.Error("Query failed:", err)
						continue
					}
				}
			})
		})
	}
}

func benchmarkAuthServer(b *testing.B, concurrentRequests int) {
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        concurrentRequests,
			MaxIdleConnsPerHost: concurrentRequests,
		},
		Timeout: 10 * time.Second,
	}

	testRequests := make([]UserLogin, concurrentRequests)
	for i := 0; i < concurrentRequests; i++ {
		testRequests[i] = UserLogin{
			Email:        fmt.Sprintf("user%d@example.com", i%1_000_000),
			PasswordHash: "PasswordHash123",
		}
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			jsonData, err := json.Marshal(testRequests[i%concurrentRequests])
			if err != nil {
				b.Error("Marshal error:", err)
				continue
			}

			req, err := http.NewRequest(
				"POST",
				ts.URL+"/user/login",
				bytes.NewBuffer(jsonData),
			)
			if err != nil {
				b.Error("Create request error:", err)
				continue
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				b.Error("Request failed:", err)
				continue
			}

			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				b.Errorf("Unexpected status: %d", resp.StatusCode)
			}

			i++
		}
	})
}

func BenchmarkAuthServer(b *testing.B) {
	db := setupTestDBNEW(b, true, 1_000_000)
	defer db.Close()

	concurrencyLevels := []int{1, 10, 50, 100, 200, 500}
	for _, n := range concurrencyLevels {
		b.Run(fmt.Sprintf("concurrency_%d", n), func(b *testing.B) {
			benchmarkAuthServer(b, n)
		})
	}
}

func Benchmark_db_i(b *testing.B) {
	db := setupTestDBNEW(b, true, 1_000_000)
	defer db.Close()

	benchmarks := []struct {
		name    string
		workers int
	}{
		{"1_workers", 1},
		{"2_workers", 2},
		{"4_workers", 4},
		{"6_workers", 6},
		{"5000_workers", 5000},
		{"10000_workers", 10000},
	}

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			b.SetParallelism(bb.workers)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					i++
					email := fmt.Sprintf("user%d@example.com", i%1_000_000)

					var user struct {
						ID           string
						PasswordHash string
					}

					err := db.QueryRow(context.Background(),
						"SELECT id, password_hash FROM users WHERE email = $1", email).
						Scan(&user.ID, &user.PasswordHash)

					if err != nil {
						b.Error("Query failed:", err)
						continue
					}
				}
			})
		})
	}
}

func Benchmark_db_ni(b *testing.B) {
	db := setupTestDBNEW(b, false, 1_000_000)
	defer db.Close()

	benchmarks := []struct {
		name    string
		workers int
	}{
		{"1_workers", 1},
		{"2_workers", 2},
		{"4_workers", 4},
		{"6_workers", 6},
	}

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			b.SetParallelism(bb.workers)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					i++
					email := fmt.Sprintf("user%d@example.com", i%1_000_000)

					var user struct {
						ID           string
						PasswordHash string
					}

					err := db.QueryRow(context.Background(),
						"SELECT id, password_hash FROM users WHERE email = $1", email).
						Scan(&user.ID, &user.PasswordHash)

					if err != nil {
						b.Error("Query failed:", err)
						continue
					}
				}
			})
		})
	}
}

func BenchmarkDBWithConcurrencyNEWServerIND(b *testing.B) {
	db := setupTestDBNEW(b, true, 1_000_000)
	defer db.Close()

	benchmarks := []struct {
		name    string
		workers int
	}{
		{"1_workers", 1},
		{"2_workers", 2},
		{"4_workers", 4},
		{"6_workers", 6},
	}

	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			b.SetParallelism(bb.workers)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					i++
					jsonData, err := json.Marshal(UserLogin{fmt.Sprintf("user%d@example.com", i%1_000_000), "PasswordHash123"})
					if err != nil {
						b.Error("Marshal error:", err)
						continue
					}

					req, err := http.NewRequest(
						"POST",
						ts.URL+"/user/login",
						bytes.NewBuffer(jsonData),
					)
					if err != nil {
						b.Error("Create request error:", err)
						continue
					}
					req.Header.Set("Content-Type", "application/json")

					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						b.Error("Request failed:", err)
						continue
					}

					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						b.Errorf("Unexpected status: %d", resp.StatusCode)
					}
				}
			})
		})
	}
}

func BenchmarkDBWithConcurrencyNEWServerNIND(b *testing.B) {
	db := setupTestDBNEW(b, false, 1_000_000)
	defer db.Close()

	benchmarks := []struct {
		name    string
		workers int
	}{
		{"1_workers", 1},
		{"2_workers", 2},
		{"4_workers", 4},
		{"6_workers", 6},
	}

	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			b.SetParallelism(bb.workers)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					i++
					jsonData, err := json.Marshal(UserLogin{fmt.Sprintf("user%d@example.com", i%1_000_000), "PasswordHash123"})
					if err != nil {
						b.Error("Marshal error:", err)
						continue
					}

					req, err := http.NewRequest(
						"POST",
						ts.URL+"/user/login",
						bytes.NewBuffer(jsonData),
					)
					if err != nil {
						b.Error("Create request error:", err)
						continue
					}
					req.Header.Set("Content-Type", "application/json")

					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						b.Error("Request failed:", err)
						continue
					}

					// io.Copy(io.Discard, resp.Body)
					resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						b.Errorf("Unexpected status: %d", resp.StatusCode)
					}
				}
			})
		})
	}
}

func Benchmark_s_ni(b *testing.B) {
	db := setupTestDBNEW(b, false, 1_000_000)
	defer db.Close()

	benchmarks := []struct {
		name    string
		workers int
	}{
		{"1_workers", 1},
		{"2_workers", 2},
		{"4_workers", 4},
		{"6_workers", 6},
	}

	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			b.SetParallelism(bb.workers)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					i++
					jsonData, err := json.Marshal(UserLogin{fmt.Sprintf("user%d@example.com", i%1_000_000), "PasswordHash123"})
					if err != nil {
						b.Error("Marshal error:", err)
						continue
					}

					req, err := http.NewRequest(
						"POST",
						ts.URL+"/user/login",
						bytes.NewBuffer(jsonData),
					)
					if err != nil {
						b.Error("Create request error:", err)
						continue
					}
					req.Header.Set("Content-Type", "application/json")

					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						b.Error("Request failed:", err)
						continue
					}

					// io.Copy(io.Discard, resp.Body)
					resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						b.Errorf("Unexpected status: %d", resp.StatusCode)
					}
				}
			})
		})
	}
}

func Benchmark_s_i(b *testing.B) {
	db := setupTestDBNEW(b, true, 1_000_000)
	defer db.Close()

	benchmarks := []struct {
		name    string
		workers int
	}{
		{"1_workers", 1},
		{"2_workers", 2},
		{"4_workers", 4},
		{"6_workers", 6},
	}

	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			b.SetParallelism(bb.workers)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					i++
					jsonData, err := json.Marshal(UserLogin{fmt.Sprintf("user%d@example.com", i%1_000_000), "PasswordHash123"})
					if err != nil {
						b.Error("Marshal error:", err)
						continue
					}

					req, err := http.NewRequest(
						"POST",
						ts.URL+"/user/login",
						bytes.NewBuffer(jsonData),
					)
					if err != nil {
						b.Error("Create request error:", err)
						continue
					}
					req.Header.Set("Content-Type", "application/json")

					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						b.Error("Request failed:", err)
						continue
					}

					// io.Copy(io.Discard, resp.Body)
					resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						b.Errorf("Unexpected status: %d", resp.StatusCode)
					}
				}
			})
		})
	}
}
