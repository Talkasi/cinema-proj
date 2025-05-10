package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestGetEquipmentTypesEmptyGuest(t *testing.T) {
	_ = ClearTable(TestAdminDB, "equipment_types")
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL+"/equipment-types", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		fmt.Println(string(body))
		t.Errorf("Expected status %v; got %v", http.StatusNotFound, resp.Status)
	}

	if len(body) == 0 {
		t.Errorf("Expected non-empty body")
	}
}

func TestGetEquipmentTypesEmptyUser(t *testing.T) {
	_ = ClearTable(TestAdminDB, "equipment_types")
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	tokenString, err := GenerateToken("email", "ruser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req, err := http.NewRequest("GET", ts.URL+"/equipment-types", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", tokenString)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		fmt.Println(string(body))
		t.Errorf("Expected status %v; got %v", http.StatusNotFound, resp.Status)
	}

	if len(body) == 0 {
		t.Errorf("Expected non-empty body")
	}
}

func TestGetEquipmentTypesEmptyAdmin(t *testing.T) {
	_ = ClearTable(TestAdminDB, "equipment_types")
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	tokenString, err := GenerateToken("email", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req, err := http.NewRequest("GET", ts.URL+"/equipment-types", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", tokenString)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		fmt.Println(string(body))
		t.Errorf("Expected status %v; got %v", http.StatusNotFound, resp.Status)
	}

	if len(body) == 0 {
		t.Errorf("Expected non-empty body")
	}
}

func TestGetEquipmentTypesNonEmptyGuest(t *testing.T) {
	_ = SeedEquipmentTypes(TestAdminDB)
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL+"/equipment-types", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println(string(body))
		t.Errorf("Expected status OK; got %v", resp.Status)
	}

	if len(body) == 0 {
		t.Errorf("Expected non-empty body")
	}

	// Протестировать содержимое
}

func TestGetEquipmentTypesNonEmptyUser(t *testing.T) {
	_ = SeedEquipmentTypes(TestAdminDB)
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	tokenString, err := GenerateToken("email", "ruser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req, err := http.NewRequest("GET", ts.URL+"/equipment-types", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", tokenString)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println(string(body))
		t.Errorf("Expected status OK; got %v", resp.Status)
	}

	if len(body) == 0 {
		t.Errorf("Expected non-empty body")
	}

	// Протестировать содержимое
}

func TestGetEquipmentTypesNonEmptyAdmin(t *testing.T) {
	_ = SeedEquipmentTypes(TestAdminDB)
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	tokenString, err := GenerateToken("email", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req, err := http.NewRequest("GET", ts.URL+"/equipment-types", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", tokenString)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println(string(body))
		t.Errorf("Expected status OK; got %v", resp.Status)
	}

	if len(body) == 0 {
		t.Errorf("Expected non-empty body")
	}

	// Протестировать содержимое
}

func TestGetEquipmentTypeByIDUnknownIDNonEmptyGuest(t *testing.T) {
	_ = SeedEquipmentTypes(TestAdminDB)
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	equipmentTypeID := uuid.New().String()
	reqByID, err := http.NewRequest("GET", ts.URL+"/equipment-types/"+string(equipmentTypeID), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	respByID, err := http.DefaultClient.Do(reqByID)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer respByID.Body.Close()

	if respByID.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %v; got %v", http.StatusNotFound, respByID.Status)
	}
}

func TestGetEquipmentTypeByIDUnknownIDNonEmptyUser(t *testing.T) {
	_ = SeedEquipmentTypes(TestAdminDB)
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	tokenString, err := GenerateToken("email", "ruser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	equipmentTypeID := uuid.New().String()
	reqByID, err := http.NewRequest("GET", ts.URL+"/equipment-types/"+string(equipmentTypeID), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	reqByID.Header.Set("Authorization", tokenString)

	respByID, err := http.DefaultClient.Do(reqByID)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer respByID.Body.Close()

	if respByID.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %v; got %v", http.StatusNotFound, respByID.Status)
	}
}

func TestGetEquipmentTypeByIDUnknownIDNonEmptyAdmin(t *testing.T) {
	_ = SeedEquipmentTypes(TestAdminDB)
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	tokenString, err := GenerateToken("email", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	equipmentTypeID := uuid.New().String()
	reqByID, err := http.NewRequest("GET", ts.URL+"/equipment-types/"+string(equipmentTypeID), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	reqByID.Header.Set("Authorization", tokenString)

	respByID, err := http.DefaultClient.Do(reqByID)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer respByID.Body.Close()

	if respByID.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %v; got %v", http.StatusNotFound, respByID.Status)
	}
}

func TestGetEquipmentTypeByIDUnknownIDWhenEmptyGuest(t *testing.T) {
	_ = ClearTable(TestAdminDB, "equipment_types")
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	equipmentTypeID := uuid.New().String()
	reqByID, err := http.NewRequest("GET", ts.URL+"/equipment-types/"+string(equipmentTypeID), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	respByID, err := http.DefaultClient.Do(reqByID)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer respByID.Body.Close()

	if respByID.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %v; got %v", http.StatusNotFound, respByID.Status)
	}
}

func TestGetEquipmentTypeByIDUnknownIDWhenEmptyUser(t *testing.T) {
	_ = ClearTable(TestAdminDB, "equipment_types")
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	tokenString, err := GenerateToken("email", "ruser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	equipmentTypeID := uuid.New().String()
	reqByID, err := http.NewRequest("GET", ts.URL+"/equipment-types/"+string(equipmentTypeID), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	reqByID.Header.Set("Authorization", tokenString)

	respByID, err := http.DefaultClient.Do(reqByID)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer respByID.Body.Close()

	if respByID.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %v; got %v", http.StatusNotFound, respByID.Status)
	}
}

func TestGetEquipmentTypeByIDUnknownIDWhenEmptyAdmin(t *testing.T) {
	_ = ClearTable(TestAdminDB, "equipment_types")
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	tokenString, err := GenerateToken("email", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	equipmentTypeID := uuid.New().String()
	reqByID, err := http.NewRequest("GET", ts.URL+"/equipment-types/"+string(equipmentTypeID), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	reqByID.Header.Set("Authorization", tokenString)

	respByID, err := http.DefaultClient.Do(reqByID)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer respByID.Body.Close()

	if respByID.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %v; got %v", http.StatusNotFound, respByID.Status)
	}
}

func TestGetEquipmentTypeByIDInvalidIDWhenEmptyGuest(t *testing.T) {
	_ = ClearTable(TestAdminDB, "equipment_types")
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	equipmentTypeID := "845837954"
	reqByID, err := http.NewRequest("GET", ts.URL+"/equipment-types/"+string(equipmentTypeID), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	respByID, err := http.DefaultClient.Do(reqByID)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer respByID.Body.Close()

	if respByID.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %v; got %v", http.StatusBadRequest, respByID.Status)
	}
}

func TestGetEquipmentTypeByIDInvalidIDWhenEmptyUser(t *testing.T) {
	_ = ClearTable(TestAdminDB, "equipment_types")
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	tokenString, err := GenerateToken("email", "ruser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	equipmentTypeID := "345345345"
	reqByID, err := http.NewRequest("GET", ts.URL+"/equipment-types/"+string(equipmentTypeID), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	reqByID.Header.Set("Authorization", tokenString)

	respByID, err := http.DefaultClient.Do(reqByID)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer respByID.Body.Close()

	if respByID.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %v; got %v", http.StatusBadRequest, respByID.Status)
	}
}

func TestGetEquipmentTypeByIDInvalidIDWhenEmptyAdmin(t *testing.T) {
	_ = ClearTable(TestAdminDB, "equipment_types")
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	tokenString, err := GenerateToken("email", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	equipmentTypeID := "345345345"
	reqByID, err := http.NewRequest("GET", ts.URL+"/equipment-types/"+string(equipmentTypeID), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	reqByID.Header.Set("Authorization", tokenString)

	respByID, err := http.DefaultClient.Do(reqByID)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer respByID.Body.Close()

	if respByID.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %v; got %v", http.StatusBadRequest, respByID.Status)
	}
}
func TestGetEquipmentTypeByIDInvalidIDNonEmptyGuest(t *testing.T) {
	_ = SeedEquipmentTypes(TestAdminDB)
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	equipmentTypeID := "yu5yieuyi4"
	reqByID, err := http.NewRequest("GET", ts.URL+"/equipment-types/"+string(equipmentTypeID), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	respByID, err := http.DefaultClient.Do(reqByID)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer respByID.Body.Close()

	if respByID.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %v; got %v", http.StatusBadRequest, respByID.Status)
	}
}

func TestGetEquipmentTypeByIDInvalidIDNonEmptyUser(t *testing.T) {
	_ = SeedEquipmentTypes(TestAdminDB)
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	tokenString, err := GenerateToken("email", "ruser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	equipmentTypeID := "yu5yieuyi4"
	reqByID, err := http.NewRequest("GET", ts.URL+"/equipment-types/"+string(equipmentTypeID), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	reqByID.Header.Set("Authorization", tokenString)

	respByID, err := http.DefaultClient.Do(reqByID)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer respByID.Body.Close()

	if respByID.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %v; got %v", http.StatusBadRequest, respByID.Status)
	}
}

func TestGetEquipmentTypeByIDInvalidIDNonEmptyAdmin(t *testing.T) {
	_ = SeedEquipmentTypes(TestAdminDB)
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	tokenString, err := GenerateToken("email", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	equipmentTypeID := "yu5yieuyi4"
	reqByID, err := http.NewRequest("GET", ts.URL+"/equipment-types/"+string(equipmentTypeID), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	reqByID.Header.Set("Authorization", tokenString)

	respByID, err := http.DefaultClient.Do(reqByID)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer respByID.Body.Close()

	if respByID.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %v; got %v", http.StatusBadRequest, respByID.Status)
	}
}

func TestGetEquipmentTypeByIDNonEmptyGuest(t *testing.T) {
	_ = SeedEquipmentTypes(TestAdminDB)
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL+"/equipment-types", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %v in GetAll; got %v", http.StatusOK, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading response body: %v", err)
	}

	var equipmentTypes []EquipmentType
	err = json.Unmarshal(body, &equipmentTypes)
	if err != nil {
		t.Errorf("Error unmarshalling JSON: %v", err)
	}

	if len(equipmentTypes) == 0 {
		t.Fatal("Expected at least one equipment type, got none")
	}

	equipmentTypeID := equipmentTypes[0].ID
	reqByID, err := http.NewRequest("GET", ts.URL+"/equipment-types/"+string(equipmentTypeID), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	respByID, err := http.DefaultClient.Do(reqByID)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer respByID.Body.Close()

	if respByID.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", respByID.Status)
	}

	id_body, err := io.ReadAll(respByID.Body)
	if err != nil {
		t.Errorf("Error reading response body: %v", err)
	}

	var equipmentType EquipmentType
	err = json.Unmarshal(id_body, &equipmentType)
	if err != nil {
		t.Errorf("Error unmarshalling JSON: %v", err)
		return
	}

	if equipmentType.ID != equipmentTypeID {
		t.Errorf("Expected ID %v; got %v", equipmentTypeID, equipmentType.ID)
	}
}

func TestGetEquipmentTypeByIDNonEmptyUser(t *testing.T) {
	_ = SeedEquipmentTypes(TestAdminDB)
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL+"/equipment-types", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	tokenString, err := GenerateToken("email", "ruser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	req.Header.Set("Authorization", tokenString)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading response body: %v", err)
	}

	var equipmentTypes []EquipmentType
	err = json.Unmarshal(body, &equipmentTypes)
	if err != nil {
		t.Errorf("Error unmarshalling JSON: %v", err)
	}

	if len(equipmentTypes) == 0 {
		t.Fatal("Expected at least one equipment type, got none")
	}

	equipmentTypeID := equipmentTypes[0].ID
	reqByID, err := http.NewRequest("GET", ts.URL+"/equipment-types/"+string(equipmentTypeID), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	reqByID.Header.Set("Authorization", tokenString)
	respByID, err := http.DefaultClient.Do(reqByID)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer respByID.Body.Close()

	if respByID.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", respByID.Status)
	}

	id_body, err := io.ReadAll(respByID.Body)
	if err != nil {
		t.Errorf("Error reading response body: %v", err)
	}

	var equipmentType EquipmentType
	err = json.Unmarshal(id_body, &equipmentType)
	if err != nil {
		t.Errorf("Error unmarshalling JSON: %v", err)
		return
	}

	if equipmentType.ID != equipmentTypeID {
		t.Errorf("Expected ID %v; got %v", equipmentTypeID, equipmentType.ID)
	}
}

func TestGetEquipmentTypeByIDNonEmptyAdmin(t *testing.T) {
	_ = SeedEquipmentTypes(TestAdminDB)
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL+"/equipment-types", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	tokenString, err := GenerateToken("email", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	req.Header.Set("Authorization", tokenString)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading response body: %v", err)
	}

	var equipmentTypes []EquipmentType
	err = json.Unmarshal(body, &equipmentTypes)
	if err != nil {
		t.Errorf("Error unmarshalling JSON: %v", err)
	}

	if len(equipmentTypes) == 0 {
		t.Fatal("Expected at least one equipment type, got none")
	}

	equipmentTypeID := equipmentTypes[0].ID
	reqByID, err := http.NewRequest("GET", ts.URL+"/equipment-types/"+string(equipmentTypeID), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	reqByID.Header.Set("Authorization", tokenString)
	respByID, err := http.DefaultClient.Do(reqByID)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer respByID.Body.Close()

	if respByID.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", respByID.Status)
	}

	id_body, err := io.ReadAll(respByID.Body)
	if err != nil {
		t.Errorf("Error reading response body: %v", err)
	}

	var equipmentType EquipmentType
	err = json.Unmarshal(id_body, &equipmentType)
	if err != nil {
		t.Errorf("Error unmarshalling JSON: %v", err)
		return
	}

	if equipmentType.ID != equipmentTypeID {
		t.Errorf("Expected ID %v; got %v", equipmentTypeID, equipmentType.ID)
	}
}

func TestCreateEquipmentTypeForbiddenGuest(t *testing.T) {
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	equipmentType := EquipmentType{
		Name:        "Test Equipment",
		Description: "Test Description",
	}
	jsonData, err := json.Marshal(equipmentType)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	req, err := http.NewRequest("POST", ts.URL+"/equipment-types", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status %v; got %v", http.StatusForbidden, resp.Status)
	}
}

func TestCreateEquipmentTypeForbiddenUser(t *testing.T) {
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	equipmentType := EquipmentType{
		Name:        "Test Equipment",
		Description: "Test Description",
	}
	jsonData, err := json.Marshal(equipmentType)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	tokenString, err := GenerateToken("email", "ruser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	req, err := http.NewRequest("POST", ts.URL+"/equipment-types", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", tokenString)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status %v; got %v", http.StatusForbidden, resp.Status)
	}
}

func TestCreateEquipmentTypeSuccessAdmin(t *testing.T) {
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	equipmentType := EquipmentType{
		Name:        "Test Equipment",
		Description: "Test Description",
	}
	jsonData, err := json.Marshal(equipmentType)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	tokenString, err := GenerateToken("email", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	req, err := http.NewRequest("POST", ts.URL+"/equipment-types", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", tokenString)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status %v; got %v", http.StatusCreated, resp.Status)
	}

	var createdEquipmentType EquipmentType
	if err := json.NewDecoder(resp.Body).Decode(&createdEquipmentType); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if createdEquipmentType.ID == "" {
		t.Error("Expected non-empty ID in response")
	}
}

func TestCreateEquipmentTypeInvalidJSONGuest(t *testing.T) {
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	invalidJSON := []byte("{invalid json}")

	req, err := http.NewRequest("POST", ts.URL+"/equipment-types", bytes.NewBuffer(invalidJSON))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %v; got %v", http.StatusBadRequest, resp.Status)
	}
}

func TestCreateEquipmentTypeInvalidJSONUser(t *testing.T) {
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	invalidJSON := []byte("{invalid json}")

	tokenString, err := GenerateToken("email", "user")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	req, err := http.NewRequest("POST", ts.URL+"/equipment-types", bytes.NewBuffer(invalidJSON))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", tokenString)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %v; got %v", http.StatusBadRequest, resp.Status)
	}
}

func TestCreateEquipmentTypeInvalidJSONAdmin(t *testing.T) {
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	invalidJSON := []byte("{invalid json}")

	tokenString, err := GenerateToken("email", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	req, err := http.NewRequest("POST", ts.URL+"/equipment-types", bytes.NewBuffer(invalidJSON))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", tokenString)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %v; got %v", http.StatusBadRequest, resp.Status)
	}
}

func TestCreateEquipmentTypeInsertErrorGuest(t *testing.T) {
	_ = ClearTable(TestAdminDB, "equipment_types")
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	equipmentType := EquipmentType{
		Name:        "Test Equipment",
		Description: "This is a test equipment type",
	}
	body, err := json.Marshal(equipmentType)
	if err != nil {
		t.Fatalf("Failed to marshal equipment type: %v", err)
	}

	_, err = TestAdminDB.Exec("INSERT INTO equipment_types (name) VALUES ($1)", equipmentType.Name)
	if err != nil {
		t.Fatalf("Failed to insert into test database: %v", err)
	}

	req, err := http.NewRequest("POST", ts.URL+"/equipment-types", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status %v; got %v", http.StatusForbidden, resp.Status)
	}
}

func TestCreateEquipmentTypeInsertErrorUser(t *testing.T) {
	_ = ClearTable(TestAdminDB, "equipment_types")
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	equipmentType := EquipmentType{
		Name:        "Test Equipment",
		Description: "This is a test equipment type",
	}
	body, err := json.Marshal(equipmentType)
	if err != nil {
		t.Fatalf("Failed to marshal equipment type: %v", err)
	}

	_, err = TestAdminDB.Exec("INSERT INTO equipment_types (name) VALUES ($1)", equipmentType.Name)
	if err != nil {
		t.Fatalf("Failed to insert into test database: %v", err)
	}

	tokenString, err := GenerateToken("email", "user")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	req, err := http.NewRequest("POST", ts.URL+"/equipment-types", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", tokenString)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status %v; got %v", http.StatusForbidden, resp.Status)
	}
}

func TestCreateEquipmentTypeInsertErrorAdmin(t *testing.T) {
	_ = ClearTable(TestAdminDB, "equipment_types")
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	equipmentType := EquipmentType{
		Name:        "Test Equipment",
		Description: "This is a test equipment type",
	}
	body, err := json.Marshal(equipmentType)
	if err != nil {
		t.Fatalf("Failed to marshal equipment type: %v", err)
	}

	_, err = TestAdminDB.Exec("INSERT INTO equipment_types (name) VALUES ($1)", equipmentType.Name)
	if err != nil {
		t.Fatalf("Failed to insert into test database: %v", err)
	}

	tokenString, err := GenerateToken("email", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	req, err := http.NewRequest("POST", ts.URL+"/equipment-types", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", tokenString)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Expected status %v; got %v", http.StatusConflict, resp.Status)
	}
}
