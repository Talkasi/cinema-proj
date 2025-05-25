package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
)

func TestGetSeats(t *testing.T) {
	tests := []struct {
		name           string
		seedData       bool
		role           string
		expectedStatus int
	}{
		{"Empty as Guest", false, "", http.StatusForbidden},
		{"Empty as User", false, os.Getenv("CLAIM_ROLE_USER"), http.StatusForbidden},
		{"Empty as Admin", false, os.Getenv("CLAIM_ROLE_ADMIN"), http.StatusNotFound},
		{"NonEmpty as Guest", true, "", http.StatusForbidden},
		{"NonEmpty as User", true, os.Getenv("CLAIM_ROLE_USER"), http.StatusForbidden},
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

			req := createRequest(t, "GET", ts.URL+"/seats", generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var seats []Seat
				parseResponseBody(t, resp, &seats)

				if len(seats) == 0 {
					t.Error("Expected non-empty seats list")
				}
			}
		})
	}
}

func TestGetSeatByID(t *testing.T) {
	setupValidIDTest := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, SeatsData[0].ID
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

			req := createRequest(t, "GET", ts.URL+"/seats/"+id, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var seat Seat
				parseResponseBody(t, resp, &seat)

				if seat.ID != id {
					t.Errorf("Expected ID %v; got %v", id, seat.ID)
				}
			}
		})
	}
}

func TestCreateSeat(t *testing.T) {
	validSeat := SeatData{
		HallID:     HallsData[2].ID,
		SeatTypeID: SeatTypesData[1].ID,
		RowNumber:  1,
		SeatNumber: 1,
	}

	invalidSeat := SeatData{
		HallID:     "invalid",
		SeatTypeID: uuid.New().String(),
		RowNumber:  0,
		SeatNumber: -1,
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
			validSeat,
			func(t *testing.T) {
				SeedAll(TestAdminDB)
			},
			http.StatusForbidden,
		},
		{
			"Forbidden User",
			os.Getenv("CLAIM_ROLE_USER"),
			validSeat,
			func(t *testing.T) {
				SeedAll(TestAdminDB)
			},
			http.StatusForbidden,
		},
		{
			"Success Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			validSeat,
			func(t *testing.T) {
				SeedAll(TestAdminDB)
			},
			http.StatusCreated,
		},
		{
			"Invalid JSON Guest",
			"",
			"{invalid json}",
			func(t *testing.T) {
				SeedAll(TestAdminDB)
			},
			http.StatusBadRequest,
		},
		{
			"Invalid JSON User",
			os.Getenv("CLAIM_ROLE_USER"),
			"{invalid json}",
			func(t *testing.T) {
				SeedAll(TestAdminDB)
			},
			http.StatusBadRequest,
		},
		{
			"Invalid JSON Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"{invalid json}",
			func(t *testing.T) {
				SeedAll(TestAdminDB)
			},
			http.StatusBadRequest,
		},
		{
			"Invalid data Guest",
			"",
			invalidSeat,
			func(t *testing.T) {
				SeedAll(TestAdminDB)
			},
			http.StatusBadRequest,
		},
		{
			"Invalid data User",
			os.Getenv("CLAIM_ROLE_USER"),
			invalidSeat,
			func(t *testing.T) {
				SeedAll(TestAdminDB)
			},
			http.StatusBadRequest,
		},
		{
			"Invalid data Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			invalidSeat,
			func(t *testing.T) {
				SeedAll(TestAdminDB)
			},
			http.StatusBadRequest,
		},
		{
			"Conflict Admin - duplicate seat",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			validSeat,
			func(t *testing.T) {
				SeedAll(TestAdminDB)

				_, err := TestAdminDB.Exec(context.Background(),
					"INSERT INTO seats (id, hall_id, seat_type_id, row_number, seat_number) VALUES ($1, $2, $3, $4, $5)",
					uuid.New(), validSeat.HallID, validSeat.SeatTypeID, validSeat.RowNumber, validSeat.SeatNumber)
				if err != nil {
					t.Fatal("Failed to create test seat:", err)
				}
			},
			http.StatusConflict,
		},
		{
			"Invalid Hall ID",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			SeatData{
				HallID:     uuid.New().String(), // Несуществующий зал
				SeatTypeID: SeatTypesData[0].ID,
				RowNumber:  1,
				SeatNumber: 1,
			},
			func(t *testing.T) {
				SeedAll(TestAdminDB)
			},
			http.StatusConflict,
		},
		{
			"Invalid Seat Type ID",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			SeatData{
				HallID:     HallsData[0].ID,
				SeatTypeID: uuid.New().String(), // Несуществующий тип
				RowNumber:  1,
				SeatNumber: 99,
			},
			func(t *testing.T) {
				SeedAll(TestAdminDB)
			},
			http.StatusConflict,
		},
		{
			"Row number too big",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			SeatData{
				HallID:     HallsData[0].ID,
				SeatTypeID: SeatTypesData[0].ID,
				RowNumber:  101,
				SeatNumber: 1,
			},
			nil,
			http.StatusBadRequest,
		},
		{
			"Seat number too big",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			SeatData{
				HallID:     HallsData[0].ID,
				SeatTypeID: SeatTypesData[0].ID,
				RowNumber:  1,
				SeatNumber: 101,
			},
			nil,
			http.StatusBadRequest,
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

			req := createRequest(t, "POST", ts.URL+"/seats", generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusCreated {
				var seatID string
				parseResponseBody(t, resp, &seatID)

				if seatID == "" {
					t.Error("Expected non-empty ID in response")
				}

				if _, err := uuid.Parse(seatID); err != nil {
					t.Error("Invalid UUID format in response")
				}
			}
		})
	}
}

func TestUpdateSeat(t *testing.T) {
	validUpdate := SeatData{
		HallID:     HallsData[3].ID,
		SeatTypeID: SeatTypesData[1].ID,
		RowNumber:  2,
		SeatNumber: 2,
	}

	invalidUpdate := SeatData{
		HallID:     "invalid",
		SeatTypeID: uuid.New().String(),
		RowNumber:  0,
		SeatNumber: -1,
	}

	setupExistingSeat := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		SeedAll(TestAdminDB)
		seat := SeatsData[0]
		validUpdate.HallID = seat.HallID
		validUpdate.SeatTypeID = seat.SeatTypeID
		return ts, seat.ID
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
			validUpdate,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, ""
			},
			http.StatusBadRequest,
		},
		{
			"Invalid UUID as User",
			os.Getenv("CLAIM_ROLE_USER"),
			"invalid-uuid",
			validUpdate,
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
			validUpdate,
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
			validUpdate,
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
			validUpdate,
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
			validUpdate,
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
			setupExistingSeat,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as User",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			"invalid-json",
			setupExistingSeat,
			http.StatusBadRequest,
		},
		{
			"Invalid JSON as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			"invalid-json",
			setupExistingSeat,
			http.StatusBadRequest,
		},
		{
			"Invalid data as Guest",
			"",
			"",
			invalidUpdate,
			setupExistingSeat,
			http.StatusBadRequest,
		},
		{
			"Invalid data as User",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			invalidUpdate,
			setupExistingSeat,
			http.StatusBadRequest,
		},
		{
			"Invalid data as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			invalidUpdate,
			setupExistingSeat,
			http.StatusBadRequest,
		},
		{
			"Forbidden Guest",
			"",
			"",
			validUpdate,
			setupExistingSeat,
			http.StatusForbidden,
		},
		{
			"Forbidden User",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			validUpdate,
			setupExistingSeat,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			validUpdate,
			setupExistingSeat,
			http.StatusOK,
		},
		{
			"Conflict Admin - duplicate seat",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			SeatData{SeatsData[0].HallID, SeatsData[2].SeatTypeID, SeatsData[0].RowNumber, SeatsData[0].SeatNumber},
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				SeedAll(TestAdminDB)
				return ts, SeatsData[1].ID
			},
			http.StatusConflict,
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

			req := createRequest(t, "PUT", ts.URL+"/seats/"+effectiveID, generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestDeleteSeat(t *testing.T) {
	setupExistingSeat := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, SeatsData[0].ID
	}

	setupExistingSeatNotTouched := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, SeatsData[3].ID
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
			setupExistingSeat,
			http.StatusForbidden,
		},
		{
			"Forbidden as User",
			os.Getenv("CLAIM_ROLE_USER"),
			"",
			setupExistingSeat,
			http.StatusForbidden,
		},
		{
			"Dependency error as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			setupExistingSeat,
			http.StatusConflict,
		},
		{
			"Success as Admin",
			os.Getenv("CLAIM_ROLE_ADMIN"),
			"",
			setupExistingSeatNotTouched,
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

			req := createRequest(t, "DELETE", ts.URL+"/seats/"+effectiveID, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestCreateSeatDBError(t *testing.T) {
	ts := setupTestServer()
	SeedUsers(TestAdminDB)
	defer ts.Close()

	// Создаем ситуацию с ошибкой БД
	TestAdminDB.Close()
	TestGuestDB.Close()
	TestUserDB.Close()

	req := createRequest(t, "POST", ts.URL+"/seats",
		generateToken(t, os.Getenv("CLAIM_ROLE_ADMIN")),
		SeatData{
			HallID:     uuid.New().String(),
			SeatTypeID: uuid.New().String(),
			RowNumber:  1,
			SeatNumber: 1,
		})
	resp := executeRequest(t, req, http.StatusInternalServerError)
	defer resp.Body.Close()

	if err := InitTestDB(); err != nil {
		log.Fatal("ошибка подключения к БД: ", err)
	}
}

func TestGetSeatsByHallID(t *testing.T) {
	setupWithHallsAndSeats := func(t *testing.T) *httptest.Server {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts
	}

	tests := []struct {
		name           string
		hallID         string
		setup          func(t *testing.T) *httptest.Server
		expectedStatus int
		expectedCount  int
	}{
		{
			"Неверный формат ID зала - ошибка",
			"invalid-uuid",
			setupWithHallsAndSeats,
			http.StatusBadRequest,
			0,
		},
		{
			"Зал не найден - ошибка",
			uuid.New().String(),
			setupWithHallsAndSeats,
			http.StatusNotFound,
			0,
		},
		{
			"Получение мест по ID зала",
			HallsData[0].ID,
			setupWithHallsAndSeats,
			http.StatusOK,
			3,
		},
		{
			"Получение мест по ID малого кинозала",
			HallsData[1].ID,
			setupWithHallsAndSeats,
			http.StatusOK,
			1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := tt.setup(t)
			SeedUsers(TestAdminDB)
			defer ts.Close()

			req := createRequest(t, "GET", ts.URL+"/halls/"+tt.hallID+"/seats", generateToken(t, os.Getenv("CLAIM_ROLE_USER")), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var seats []Seat
				if err := json.NewDecoder(resp.Body).Decode(&seats); err != nil {
					t.Fatalf("Could not decode response: %v", err)
				}

				if len(seats) != tt.expectedCount {
					t.Errorf("Expected %d seats, got %d: %v", tt.expectedCount, len(seats), seats)
				}
			}
		})
	}
}
