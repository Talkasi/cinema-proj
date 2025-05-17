package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"
)

func setupTestServer() *httptest.Server {
	err := ClearAll(TestAdminDB)
	if err != nil {
		println(err.Error())
	}
	return httptest.NewServer(NewRouter())
}

func generateToken(t *testing.T, role string) string {
	var user_id string
	switch role {
	case os.Getenv("CLAIM_ROLE_ADMIN"):
		user_id = UsersData[len(UsersData)-2].ID
	case os.Getenv("CLAIM_ROLE_USER"):
		user_id = UsersData[len(UsersData)-1].ID
	}
	token, err := GenerateToken(role, user_id)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	return token
}

func executeRequest(t *testing.T, req *http.Request, expectedStatus int) *http.Response {
	t.Helper()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}

	if resp.StatusCode != expectedStatus {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status %v; got %v. Body: %s", expectedStatus, resp.Status, string(body))
	}

	return resp
}

func createRequest(t *testing.T, method, url, token string, body interface{}) *http.Request {
	var buf io.Reader
	if body != nil {
		switch v := body.(type) {
		case string:
			buf = strings.NewReader(v)
		case []byte:
			buf = bytes.NewBuffer(v)
		default:
			jsonData, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("Failed to marshal JSON: %v", err)
			}
			buf = bytes.NewBuffer(jsonData)
		}
	}

	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	if token != "" {
		req.Header.Set("Authorization", token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req
}

func parseResponseBody(t *testing.T, resp *http.Response, v interface{}) {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading response body: %v", err)
		return
	}

	if len(body) == 0 && v != nil {
		t.Error("Expected non-empty body")
		return
	}

	if v != nil {
		if err := json.Unmarshal(body, v); err != nil {
			t.Errorf("Error unmarshalling JSON: %v", err)
		}
	}
}

func TestMain(m *testing.M) {
	IsTestMode = true
	err := godotenv.Load("config.env")
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	if err := InitTestDB(); err != nil {
		log.Fatal("ошибка подключения к БД: ", err)
	}

	m.Run()

	TestAdminDB.Close()
	TestGuestDB.Close()
	TestUserDB.Close()

}
