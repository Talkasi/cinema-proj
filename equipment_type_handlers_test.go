package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
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

	tokenString, err := GenerateToken("email", "ruser")
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

	_, err = TestAdminDB.Exec(context.Background(), "INSERT INTO equipment_types (name) VALUES ($1)", equipmentType.Name)
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

	_, err = TestAdminDB.Exec(context.Background(), "INSERT INTO equipment_types (name) VALUES ($1)", equipmentType.Name)
	if err != nil {
		t.Fatalf("Failed to insert into test database: %v", err)
	}

	tokenString, err := GenerateToken("email", "ruser")
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

	_, err = TestAdminDB.Exec(context.Background(), "INSERT INTO equipment_types (name) VALUES ($1)", equipmentType.Name)
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

func TestUpdateEquipmentTypeInvalidUUIDGuest(t *testing.T) {
	_ = SeedEquipmentTypes(TestAdminDB)
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

	req, err := http.NewRequest("PUT", ts.URL+"/equipment-types/invalid-uuid", bytes.NewBuffer(body))
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

func TestUpdateEquipmentTypeInvalidUUIDUser(t *testing.T) {
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

	tokenString, err := GenerateToken("email", "ruser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	req, err := http.NewRequest("PUT", ts.URL+"/equipment-types/invalid-uuid", bytes.NewBuffer(body))
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

func TestUpdateEquipmentTypeInvalidUUIDAdmin(t *testing.T) {
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

	tokenString, err := GenerateToken("email", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	req, err := http.NewRequest("PUT", ts.URL+"/equipment-types/invalid-uuid", bytes.NewBuffer(body))
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

func TestUpdateEquipmentTypeUnknownUUIDGuest(t *testing.T) {
	_ = SeedEquipmentTypes(TestAdminDB)
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	equipmentType := EquipmentTypeData{
		Name:        "Test Equipment",
		Description: "This is a test equipment type",
	}
	body, err := json.Marshal(equipmentType)
	if err != nil {
		t.Fatalf("Failed to marshal equipment type: %v", err)
	}

	unknown_id := uuid.New().String()
	req, err := http.NewRequest("PUT", ts.URL+"/equipment-types/"+unknown_id, bytes.NewBuffer(body))
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

func TestUpdateEquipmentTypeUnknownUUIDUser(t *testing.T) {
	_ = ClearTable(TestAdminDB, "equipment_types")
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	equipmentType := EquipmentTypeData{
		Name:        "Test Equipment",
		Description: "This is a test equipment type",
	}
	body, err := json.Marshal(equipmentType)
	if err != nil {
		t.Fatalf("Failed to marshal equipment type: %v", err)
	}

	tokenString, err := GenerateToken("email", "ruser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	unknown_id := uuid.New().String()
	req, err := http.NewRequest("PUT", ts.URL+"/equipment-types/"+unknown_id, bytes.NewBuffer(body))
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

func TestUpdateEquipmentTypeUnknownUUIDAdmin(t *testing.T) {
	_ = ClearTable(TestAdminDB, "equipment_types")
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	equipmentType := EquipmentTypeData{
		Name:        "Test Equipment",
		Description: "This is a test equipment type",
	}
	body, err := json.Marshal(equipmentType)
	if err != nil {
		t.Fatalf("Failed to marshal equipment type: %v", err)
	}

	tokenString, err := GenerateToken("email", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	unknown_id := uuid.New().String()
	req, err := http.NewRequest("PUT", ts.URL+"/equipment-types/"+unknown_id, bytes.NewBuffer(body))
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

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %v; got %v", http.StatusNotFound, resp.Status)
	}
}

func TestUpdateEquipmentTypeInvalidJSONGuest(t *testing.T) {
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
	reqUpdate, err := http.NewRequest("PUT", ts.URL+"/equipment-types/"+equipmentTypeID, strings.NewReader("invalid-json"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	reqUpdate.Header.Set("Content-Type", "application/json")

	respUpdate, err := http.DefaultClient.Do(reqUpdate)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer respUpdate.Body.Close()

	if respUpdate.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %v; got %v", http.StatusBadRequest, respUpdate.Status)
	}
}

func TestUpdateEquipmentTypeInvalidJSONUser(t *testing.T) {
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
	reqUpdate, err := http.NewRequest("PUT", ts.URL+"/equipment-types/"+equipmentTypeID, strings.NewReader("invalid-json"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	reqUpdate.Header.Set("Authorization", tokenString)
	reqUpdate.Header.Set("Content-Type", "application/json")

	respUpdate, err := http.DefaultClient.Do(reqUpdate)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer respUpdate.Body.Close()

	if respUpdate.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %v; got %v", http.StatusBadRequest, respUpdate.Status)
	}
}

func TestUpdateEquipmentTypeInvalidJSONAdmin(t *testing.T) {
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
	reqUpdate, err := http.NewRequest("PUT", ts.URL+"/equipment-types/"+equipmentTypeID, strings.NewReader("invalid-json"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	reqUpdate.Header.Set("Authorization", tokenString)
	reqUpdate.Header.Set("Content-Type", "application/json")

	respUpdate, err := http.DefaultClient.Do(reqUpdate)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer respUpdate.Body.Close()

	if respUpdate.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %v; got %v", http.StatusBadRequest, respUpdate.Status)
	}
}

func TestUpdateEquipmentTypeNotFoundGuest(t *testing.T) {
	_ = SeedEquipmentTypes(TestAdminDB)
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	equipmentType := EquipmentTypeData{
		Name:        "Test Equipment",
		Description: "This is a test equipment type",
	}
	body, err := json.Marshal(equipmentType)
	if err != nil {
		t.Fatalf("Failed to marshal equipment type: %v", err)
	}

	req, err := http.NewRequest("PUT", ts.URL+"/equipment-types/"+uuid.New().String(), bytes.NewBuffer(body))
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

func TestUpdateEquipmentTypeNotFoundUser(t *testing.T) {
	_ = SeedEquipmentTypes(TestAdminDB)
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	equipmentType := EquipmentTypeData{
		Name:        "Test Equipment",
		Description: "This is a test equipment type",
	}
	body, err := json.Marshal(equipmentType)
	if err != nil {
		t.Fatalf("Failed to marshal equipment type: %v", err)
	}

	tokenString, err := GenerateToken("email", "ruser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	req, err := http.NewRequest("PUT", ts.URL+"/equipment-types/"+uuid.New().String(), bytes.NewBuffer(body))
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

func TestUpdateEquipmentTypeNotFoundAdmin(t *testing.T) {
	_ = SeedEquipmentTypes(TestAdminDB)
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	equipmentType := EquipmentTypeData{
		Name:        "Test Equipment",
		Description: "This is a test equipment type",
	}
	body, err := json.Marshal(equipmentType)
	if err != nil {
		t.Fatalf("Failed to marshal equipment type: %v", err)
	}

	tokenString, err := GenerateToken("email", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	req, err := http.NewRequest("PUT", ts.URL+"/equipment-types/"+uuid.New().String(), bytes.NewBuffer(body))
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

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %v; got %v", http.StatusNotFound, resp.Status)
	}
}

func TestUpdateEquipmentTypeEmptyNameAdmin(t *testing.T) {
	_ = SeedEquipmentTypes(TestAdminDB)
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	equipmentType := EquipmentTypeData{
		Name:        "",
		Description: "This is a test equipment type",
	}
	body, err := json.Marshal(equipmentType)
	if err != nil {
		t.Fatalf("Failed to marshal equipment type: %v", err)
	}

	tokenString, err := GenerateToken("email", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	req, err := http.NewRequest("PUT", ts.URL+"/equipment-types/"+uuid.New().String(), bytes.NewBuffer(body))
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

func TestUpdateEquipmentTypeForbiddenGuest(t *testing.T) {
	_ = ClearTable(TestAdminDB, "equipment_types")
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

	equipmentType := EquipmentTypeData{
		Name:        "Updated Equipment",
		Description: "This is an updated equipment type",
	}

	bodyUpdate, err := json.Marshal(equipmentType)
	if err != nil {
		t.Fatalf("Failed to marshal updated equipment type: %v", err)
	}

	reqUpdate, err := http.NewRequest("PUT", ts.URL+"/equipment-types/"+equipmentTypeID, bytes.NewBuffer(bodyUpdate))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	respUpdate, err := http.DefaultClient.Do(reqUpdate)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer respUpdate.Body.Close()

	if respUpdate.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status %v; got %v", http.StatusForbidden, respUpdate.Status)
	}
}

func TestUpdateEquipmentTypeForbiddenUser(t *testing.T) {
	_ = ClearTable(TestAdminDB, "equipment_types")
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

	equipmentType := EquipmentTypeData{
		Name:        "Updated Equipment",
		Description: "This is an updated equipment type",
	}

	bodyUpdate, err := json.Marshal(equipmentType)
	if err != nil {
		t.Fatalf("Failed to marshal updated equipment type: %v", err)
	}

	reqUpdate, err := http.NewRequest("PUT", ts.URL+"/equipment-types/"+equipmentTypeID, bytes.NewBuffer(bodyUpdate))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", tokenString)
	req.Header.Set("Content-Type", "application/json")

	respUpdate, err := http.DefaultClient.Do(reqUpdate)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer respUpdate.Body.Close()

	if respUpdate.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status %v; got %v", http.StatusForbidden, respUpdate.Status)
	}
}

func TestUpdateEquipmentTypeSuccessAdmin(t *testing.T) {
	_ = ClearTable(TestAdminDB, "equipment_types")
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

	equipmentType := EquipmentTypeData{
		Name:        "Updated Equipment",
		Description: "This is an updated equipment type",
	}

	bodyUpdate, err := json.Marshal(equipmentType)
	if err != nil {
		t.Fatalf("Failed to marshal updated equipment type: %v", err)
	}

	reqUpdate, err := http.NewRequest("PUT", ts.URL+"/equipment-types/"+equipmentTypeID, bytes.NewBuffer(bodyUpdate))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	reqUpdate.Header.Set("Authorization", tokenString)
	reqUpdate.Header.Set("Content-Type", "application/json")

	respUpdate, err := http.DefaultClient.Do(reqUpdate)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer respUpdate.Body.Close()

	if respUpdate.StatusCode != http.StatusOK {
		t.Errorf("Expected status %v; got %v", http.StatusOK, respUpdate.Status)
	}
}

func TestDeleteEquipmentTypeNotFoundGuest(t *testing.T) {
	_ = SeedEquipmentTypes(TestAdminDB)
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	nonExistentID := uuid.New().String()
	req, err := http.NewRequest("DELETE", ts.URL+"/equipment-types/"+nonExistentID, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status %v; got %v", http.StatusForbidden, resp.Status)
	}
}

func TestDeleteEquipmentTypeNotFoundUser(t *testing.T) {
	_ = SeedEquipmentTypes(TestAdminDB)
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	nonExistentID := uuid.New().String()
	tokenString, err := GenerateToken("email", "ruser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	req, err := http.NewRequest("DELETE", ts.URL+"/equipment-types/"+nonExistentID, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", tokenString)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status %v; got %v", http.StatusForbidden, resp.Status)
	}
}

func TestDeleteEquipmentTypeNotFoundAdmin(t *testing.T) {
	_ = SeedEquipmentTypes(TestAdminDB)
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	nonExistentID := uuid.New().String()
	tokenString, err := GenerateToken("email", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	req, err := http.NewRequest("DELETE", ts.URL+"/equipment-types/"+nonExistentID, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", tokenString)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %v; got %v", http.StatusNotFound, resp.Status)
	}
}

func TestDeleteEquipmentTypeInvalidUUIDGuest(t *testing.T) {
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	invalidID := "invalid-uuid"
	req, err := http.NewRequest("DELETE", ts.URL+"/equipment-types/"+invalidID, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %v; got %v", http.StatusBadRequest, resp.Status)
	}
}
func TestDeleteEquipmentTypeInvalidUUIDUser(t *testing.T) {
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	invalidID := "invalid-uuid"
	tokenString, err := GenerateToken("email", "ruser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	req, err := http.NewRequest("DELETE", ts.URL+"/equipment-types/"+invalidID, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", tokenString)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %v; got %v", http.StatusBadRequest, resp.Status)
	}
}

func TestDeleteEquipmentTypeInvalidUUIDAdmin(t *testing.T) {
	ts := httptest.NewServer(NewRouter())
	defer ts.Close()

	invalidID := "invalid-uuid"
	tokenString, err := GenerateToken("email", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	req, err := http.NewRequest("DELETE", ts.URL+"/equipment-types/"+invalidID, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", tokenString)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %v; got %v", http.StatusBadRequest, resp.Status)
	}
}

func TestDeleteEquipmentTypeSuccessAdmin(t *testing.T) {
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
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	reqDelete, err := http.NewRequest("DELETE", ts.URL+"/equipment-types/"+equipmentTypeID, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	reqDelete.Header.Set("Authorization", tokenString)

	respDelete, err := http.DefaultClient.Do(reqDelete)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer respDelete.Body.Close()

	if respDelete.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status %v; got %v", http.StatusNoContent, respDelete.Status)
	}
}
