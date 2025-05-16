package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestCreateTicket(t *testing.T) {
	validTicket := TicketData{
		MovieShowID: MovieShowsData[0].ID,
		SeatID:      SeatsData[3].ID,
		Status:      Available,
		Price:       1000,
	}

	invalidTicket := TicketData{
		MovieShowID: "",
		SeatID:      "",
		Status:      "INVALID_STATUS",
		Price:       -100,
	}

	setupConflictTest := func(t *testing.T) {
		_, err := TestAdminDB.Exec(context.Background(),
			"INSERT INTO tickets (id, movie_show_id, seat_id, ticket_status, price) VALUES ($1, $2, $3, $4, $5)",
			uuid.New(), validTicket.MovieShowID, validTicket.SeatID, validTicket.Status, validTicket.Price)
		if err != nil {
			t.Fatalf("Failed to insert into test database: %v", err)
		}
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
			validTicket,
			nil,
			http.StatusForbidden,
		},
		{
			"Forbidden User",
			"CLAIM_ROLE_USER",
			validTicket,
			nil,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			"CLAIM_ROLE_ADMIN",
			validTicket,
			nil,
			http.StatusCreated,
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
			"Invalid data Guest",
			"",
			invalidTicket,
			nil,
			http.StatusBadRequest,
		},
		{
			"Invalid data User",
			"CLAIM_ROLE_USER",
			invalidTicket,
			nil,
			http.StatusBadRequest,
		},
		{
			"Invalid data Admin",
			"CLAIM_ROLE_ADMIN",
			invalidTicket,
			nil,
			http.StatusBadRequest,
		},
		{
			"Conflict Guest",
			"",
			validTicket,
			setupConflictTest,
			http.StatusForbidden,
		},
		{
			"Conflict User",
			"CLAIM_ROLE_USER",
			validTicket,
			setupConflictTest,
			http.StatusForbidden,
		},
		{
			"Conflict Admin",
			"CLAIM_ROLE_ADMIN",
			validTicket,
			setupConflictTest,
			http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			defer ts.Close()

			if tt.setup != nil {
				tt.setup(t)
			}

			req := createRequest(t, "POST", ts.URL+"/tickets", generateToken(t, tt.role), tt.body)
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

func TestGetTicketsByMovieShowID(t *testing.T) {
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
				_ = SeedAll(TestAdminDB)
			}

			movieShowID := MovieShowsData[0].ID

			req := createRequest(t, "GET", ts.URL+"/tickets/movie-show/"+movieShowID, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var tickets []Ticket
				parseResponseBody(t, resp, &tickets)

				if len(tickets) == 0 {
					t.Error("Expected at least one ticket, got none")
				}
			}
		})
	}
}

func TestGetTicketByID(t *testing.T) {
	setupValidIDTest := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, TicketsData[0].ID
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
				_ = SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			"",
			http.StatusNotFound,
		},
		{
			"Unknown ID as User",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			"CLAIM_ROLE_USER",
			http.StatusNotFound,
		},
		{
			"Unknown ID as Admin",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			"CLAIM_ROLE_ADMIN",
			http.StatusNotFound,
		},
		{
			"Invalid ID as Guest",
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
				return ts, "invalid-id"
			},
			"",
			http.StatusBadRequest,
		},
		{
			"Valid ID as Guest",
			setupValidIDTest,
			"",
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
			defer ts.Close()

			req := createRequest(t, "GET", ts.URL+"/tickets/"+id, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()

			if tt.expectedStatus == http.StatusOK {
				var ticket Ticket
				parseResponseBody(t, resp, &ticket)

				if ticket.ID != id {
					t.Errorf("Expected ID %v; got %v", id, ticket.ID)
				}
			}
		})
	}
}

func TestUpdateTicket(t *testing.T) {
	validUpdateData := TicketData{
		MovieShowID: MovieShowsData[0].ID,
		SeatID:      SeatsData[3].ID,
		Status:      Reserved,
		Price:       1500,
	}

	setupExistingTicket := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, TicketsData[1].ID
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
				_ = SeedAll(TestAdminDB)
				return ts, ""
			},
			http.StatusBadRequest,
		},
		{
			"Unknown UUID as Guest",
			"",
			uuid.New().String(),
			validUpdateData,
			func(t *testing.T) (*httptest.Server, string) {
				ts := setupTestServer()
				_ = SeedAll(TestAdminDB)
				return ts, uuid.New().String()
			},
			http.StatusForbidden,
		},
		{
			"Invalid data as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			Ticket{Status: "INVALID_STATUS", Price: -100},
			setupExistingTicket,
			http.StatusBadRequest,
		},
		{
			"Forbidden Guest",
			"",
			"",
			validUpdateData,
			setupExistingTicket,
			http.StatusForbidden,
		},
		{
			"Success Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			validUpdateData,
			setupExistingTicket,
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
			defer ts.Close()

			effectiveID := tt.id
			if effectiveID == "" {
				effectiveID = id
			}

			req := createRequest(t, "PUT", ts.URL+"/tickets/"+effectiveID, generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestDeleteTicket(t *testing.T) {
	setupExistingTicket := func(t *testing.T) (*httptest.Server, string) {
		ts := setupTestServer()
		_ = SeedAll(TestAdminDB)
		return ts, TicketsData[0].ID
	}

	tests := []struct {
		name           string
		role           string
		id             string
		setup          func(t *testing.T) (*httptest.Server, string)
		expectedStatus int
	}{
		{
			"Invalid UUID as Guest",
			"",
			"invalid-uuid",
			func(t *testing.T) (*httptest.Server, string) {
				return setupTestServer(), ""
			},
			http.StatusBadRequest,
		},
		{
			"Forbidden as Guest",
			"",
			"",
			setupExistingTicket,
			http.StatusForbidden,
		},
		{
			"Success as Admin",
			"CLAIM_ROLE_ADMIN",
			"",
			setupExistingTicket,
			http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, id := tt.setup(t)
			defer ts.Close()

			effectiveID := tt.id
			if effectiveID == "" {
				effectiveID = id
			}

			req := createRequest(t, "DELETE", ts.URL+"/tickets/"+effectiveID, generateToken(t, tt.role), nil)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

// func TestGetTicketsByUserID(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		seedData       bool
// 		role           string
// 		expectedStatus int
// 	}{
// 		{"Empty as Guest", false, "", http.StatusNotFound},
// 		{"Empty as User", false, "CLAIM_ROLE_USER", http.StatusNotFound},
// 		{"NonEmpty as User", true, "CLAIM_ROLE_USER", http.StatusOK},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ts := setupTestServer()
// 			defer ts.Close()

// 			if tt.seedData {
// 				_ = SeedAll(TestAdminDB)
// 			}

// 			userID := uuid.New().String()
// 			if tt.seedData {
// 				userID = TicketsData[0]
// 			}

// 			req := createRequest(t, "GET", ts.URL+"/tickets/user/"+userID, generateToken(t, tt.role), nil)
// 			resp := executeRequest(t, req, tt.expectedStatus)
// 			defer resp.Body.Close()

// 			if tt.expectedStatus == http.StatusOK {
// 				var tickets []Ticket
// 				parseResponseBody(t, resp, &tickets)

// 				if len(tickets) == 0 {
// 					t.Error("Expected at least one ticket, got none")
// 				}
// 			}
// 		})
// 	}
// }

func TestTicketConstraints(t *testing.T) {
	tests := []struct {
		name           string
		role           string
		body           TicketData
		expectedStatus int
	}{
		{
			"Negative price",
			"CLAIM_ROLE_ADMIN",
			TicketData{
				MovieShowID: MovieShowsData[0].ID,
				SeatID:      SeatsData[3].ID,
				Status:      Available,
				Price:       -100,
			},
			http.StatusBadRequest,
		},
		{
			"Invalid status",
			"CLAIM_ROLE_ADMIN",
			TicketData{
				MovieShowID: MovieShowsData[0].ID,
				SeatID:      SeatsData[3].ID,
				Status:      "INVALID_STATUS",
				Price:       1000,
			},
			http.StatusBadRequest,
		},
		{
			"Empty movie show ID",
			"CLAIM_ROLE_ADMIN",
			TicketData{
				MovieShowID: "",
				SeatID:      SeatsData[3].ID,
				Status:      Available,
				Price:       1000,
			},
			http.StatusBadRequest,
		},
		{
			"Valid data",
			"CLAIM_ROLE_ADMIN",
			TicketData{
				MovieShowID: MovieShowsData[0].ID,
				SeatID:      SeatsData[3].ID,
				Status:      Available,
				Price:       1000,
			},
			http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTestServer()
			SeedAll(TestAdminDB)
			defer ts.Close()

			req := createRequest(t, "POST", ts.URL+"/tickets", generateToken(t, tt.role), tt.body)
			resp := executeRequest(t, req, tt.expectedStatus)
			defer resp.Body.Close()
		})
	}
}

func TestTicketDBError(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	// Create DB error situation
	TestAdminDB.Close()
	TestGuestDB.Close()
	TestUserDB.Close()

	t.Run("Create ticket DB error", func(t *testing.T) {
		req := createRequest(t, "POST", ts.URL+"/tickets",
			generateToken(t, "CLAIM_ROLE_ADMIN"),
			TicketData{
				MovieShowID: MovieShowsData[0].ID,
				SeatID:      SeatsData[3].ID,
				Status:      Available,
				Price:       1000,
			})
		resp := executeRequest(t, req, http.StatusInternalServerError)
		defer resp.Body.Close()
	})

	t.Run("Get ticket DB error", func(t *testing.T) {
		req := createRequest(t, "GET", ts.URL+"/tickets/"+uuid.New().String(),
			generateToken(t, "CLAIM_ROLE_ADMIN"), nil)
		resp := executeRequest(t, req, http.StatusInternalServerError)
		defer resp.Body.Close()
	})

	if err := InitTestDB(); err != nil {
		log.Fatal("Failed to reconnect to DB: ", err)
	}
}
