//go:build integration
// +build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"cw/internal/dataAccess/config"
	domain "cw/internal/domain/models"
	"cw/internal/utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TicketRepositoryIntegrationTestSuite struct {
	suite.Suite
	repo        *TicketRepository
	ctx         context.Context
	testTickets []string
	testData    TestData
}

type TestData struct {
	movieShowID string
	seatID      string
	userID      string
	movieID     string
	hallID      string
}

func TestTicketRepositoryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(TicketRepositoryIntegrationTestSuite))
}

func (s *TicketRepositoryIntegrationTestSuite) SetupSuite() {
	db, err := config.NewTestDatabase()
	if err != nil {
		s.T().Fatalf("Failed to connect to test database: %v", err)
	}
	s.repo = NewTicketRepository(db)
	s.ctx = context.Background()
	s.testData = s.createRequiredTestData()
	s.cleanDatabase()
}

func (s *TicketRepositoryIntegrationTestSuite) createRequiredTestData() TestData {
	db := s.repo.db

	movieID := uuid.New().String()
	hallID := uuid.New().String()
	userID := uuid.New().String()
	seatID := uuid.New().String()
	movieShowID := uuid.New().String()
	screenTypeID := uuid.New().String()
	seatTypeID := uuid.New().String()

	_, err := db.Exec(s.ctx, `
		INSERT INTO screen_types (id, name, description, price_modifier) 
		VALUES ($1, 'Test Screen', 'Test screen type', 1.0)
	`, screenTypeID)
	if err != nil {
		s.T().Fatalf("Failed to create screen type: %v", err)
	}

	_, err = db.Exec(s.ctx, `
		INSERT INTO halls (id, screen_type_id, name, description) 
		VALUES ($1, $2, 'Test Hall', 'Test hall')
	`, hallID, screenTypeID)
	if err != nil {
		s.T().Fatalf("Failed to create hall: %v", err)
	}

	_, err = db.Exec(s.ctx, `
		INSERT INTO movies (id, title, duration, description, age_limit, release_date) 
		VALUES ($1, 'Test Movie', '02:30:00', 'Test description', 12, '2024-01-01')
	`, movieID)
	if err != nil {
		s.T().Fatalf("Failed to create movie: %v", err)
	}

	_, err = db.Exec(s.ctx, `
		INSERT INTO users (id, name, email, password_hash, birth_date) 
		VALUES ($1, 'Test User', 'test@test.com', 'hash', '1990-01-01')
	`, userID)
	if err != nil {
		s.T().Fatalf("Failed to create user: %v", err)
	}

	_, err = db.Exec(s.ctx, `
		INSERT INTO seat_types (id, name, description, price_modifier) 
		VALUES ($1, 'Test Seat Type', 'Test seat type', 1.0)
	`, seatTypeID)
	if err != nil {
		s.T().Fatalf("Failed to create seat type: %v", err)
	}

	_, err = db.Exec(s.ctx, `
		INSERT INTO seats (id, hall_id, seat_type_id, row_number, seat_number) 
		VALUES ($1, $2, $3, 1, 1)
	`, seatID, hallID, seatTypeID)
	if err != nil {
		s.T().Fatalf("Failed to create seat: %v", err)
	}

	futureTime := time.Now().Add(24 * time.Hour)
	_, err = db.Exec(s.ctx, `
		INSERT INTO movie_shows (id, movie_id, hall_id, start_time, language) 
		VALUES ($1, $2, $3, $4, 'Russkiy')
	`, movieShowID, movieID, hallID, futureTime)
	if err != nil {
		s.T().Fatalf("Failed to create movie show: %v", err)
	}

	return TestData{
		movieShowID: movieShowID,
		seatID:      seatID,
		userID:      userID,
		movieID:     movieID,
		hallID:      hallID,
	}
}

func (s *TicketRepositoryIntegrationTestSuite) cleanDatabase() {
	for _, id := range s.testTickets {
		_ = s.repo.Delete(s.ctx, id)
	}
	s.testTickets = []string{}
}

func (s *TicketRepositoryIntegrationTestSuite) SetupTest() {
	s.testTickets = []string{}
}

func (s *TicketRepositoryIntegrationTestSuite) TearDownTest() {
	s.cleanDatabase()
}

func (s *TicketRepositoryIntegrationTestSuite) TearDownSuite() {
	s.cleanDatabase()
}

func (s *TicketRepositoryIntegrationTestSuite) assertNoError(utilsErr *utils.Error) {
	if utilsErr != nil {
		s.T().Errorf("Expected no error, got: %v", utilsErr)
	}
}

func (s *TicketRepositoryIntegrationTestSuite) assertErrorCode(utilsErr *utils.Error, expectedCode int) {
	if utilsErr == nil {
		s.T().Error("Expected error, got nil")
		return
	}
	assert.Equal(s.T(), expectedCode, utilsErr.Code)
}

func (s *TicketRepositoryIntegrationTestSuite) TestCreateTicketForMovieShow() {
	ticket := domain.Ticket{
		MovieShowID: s.testData.movieShowID,
		SeatID:      s.testData.seatID,
		UserID:      &s.testData.userID,
		Price:       25.50,
		Status:      "Available",
	}

	result, utilsErr := s.repo.CreateForMovieShow(s.ctx, s.testData.movieShowID, ticket)

	s.assertNoError(utilsErr)
	assert.NotEmpty(s.T(), result.ID)
	assert.Equal(s.T(), ticket.MovieShowID, result.MovieShowID)
	assert.Equal(s.T(), ticket.SeatID, result.SeatID)
	assert.Equal(s.T(), *ticket.UserID, *result.UserID)
	assert.Equal(s.T(), ticket.Price, result.Price)
	assert.Equal(s.T(), ticket.Status, result.Status)

	s.testTickets = append(s.testTickets, result.ID)
}

func (s *TicketRepositoryIntegrationTestSuite) TestCreateTicketForMovieShow_WithoutUser() {
	ticket := domain.Ticket{
		MovieShowID: s.testData.movieShowID,
		SeatID:      s.testData.seatID,
		UserID:      nil,
		Price:       30.00,
		Status:      "Available",
	}

	result, utilsErr := s.repo.CreateForMovieShow(s.ctx, s.testData.movieShowID, ticket)

	s.assertNoError(utilsErr)
	assert.NotEmpty(s.T(), result.ID)
	assert.Nil(s.T(), result.UserID)
	assert.Equal(s.T(), ticket.Status, result.Status)

	s.testTickets = append(s.testTickets, result.ID)
}

func (s *TicketRepositoryIntegrationTestSuite) TestGetTicketByID() {
	ticket := domain.Ticket{
		MovieShowID: s.testData.movieShowID,
		SeatID:      s.testData.seatID,
		UserID:      &s.testData.userID,
		Price:       20.00,
		Status:      "Available",
	}

	created, utilsErr := s.repo.CreateForMovieShow(s.ctx, s.testData.movieShowID, ticket)
	s.assertNoError(utilsErr)
	s.testTickets = append(s.testTickets, created.ID)

	retrieved, utilsErr := s.repo.GetByID(s.ctx, created.ID)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), created.ID, retrieved.ID)
	assert.Equal(s.T(), ticket.MovieShowID, retrieved.MovieShowID)
	assert.Equal(s.T(), ticket.SeatID, retrieved.SeatID)
	assert.Equal(s.T(), *ticket.UserID, *retrieved.UserID)
	assert.Equal(s.T(), ticket.Price, retrieved.Price)
	assert.Equal(s.T(), ticket.Status, retrieved.Status)
}

func (s *TicketRepositoryIntegrationTestSuite) TestGetTicketByID_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	_, utilsErr := s.repo.GetByID(s.ctx, nonExistentID)

	s.assertErrorCode(utilsErr, 404)
}

func (s *TicketRepositoryIntegrationTestSuite) TestUpdateStatus() {
	ticket := domain.Ticket{
		MovieShowID: s.testData.movieShowID,
		SeatID:      s.testData.seatID,
		UserID:      &s.testData.userID,
		Price:       25.00,
		Status:      "Available",
	}

	created, utilsErr := s.repo.CreateForMovieShow(s.ctx, s.testData.movieShowID, ticket)
	s.assertNoError(utilsErr)
	s.testTickets = append(s.testTickets, created.ID)

	updateData := domain.Ticket{
		Status: "Purchased",
		UserID: &s.testData.userID,
	}

	updated, utilsErr := s.repo.UpdateStatus(s.ctx, created.ID, updateData)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), created.ID, updated.ID)
	assert.Equal(s.T(), updateData.Status, updated.Status)
	assert.Equal(s.T(), created.MovieShowID, updated.MovieShowID)
	assert.Equal(s.T(), created.SeatID, updated.SeatID)
	assert.Equal(s.T(), created.Price, updated.Price)
}

func (s *TicketRepositoryIntegrationTestSuite) TestUpdateStatus_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	updateData := domain.Ticket{
		Status: "Purchased",
		UserID: &s.testData.userID,
	}

	_, utilsErr := s.repo.UpdateStatus(s.ctx, nonExistentID, updateData)

	s.assertErrorCode(utilsErr, 404)
}

func (s *TicketRepositoryIntegrationTestSuite) TestDeleteTicket() {
	ticket := domain.Ticket{
		MovieShowID: s.testData.movieShowID,
		SeatID:      s.testData.seatID,
		UserID:      &s.testData.userID,
		Price:       15.00,
		Status:      "Available",
	}

	created, utilsErr := s.repo.CreateForMovieShow(s.ctx, s.testData.movieShowID, ticket)
	s.assertNoError(utilsErr)

	utilsErr = s.repo.Delete(s.ctx, created.ID)
	s.assertNoError(utilsErr)

	_, utilsErr = s.repo.GetByID(s.ctx, created.ID)
	s.assertErrorCode(utilsErr, 404)
}

func (s *TicketRepositoryIntegrationTestSuite) TestDeleteTicket_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	utilsErr := s.repo.Delete(s.ctx, nonExistentID)

	s.assertErrorCode(utilsErr, 404)
}

func (s *TicketRepositoryIntegrationTestSuite) TestGetAllTickets() {
	tickets := []domain.Ticket{
		{
			MovieShowID: s.testData.movieShowID,
			SeatID:      uuid.New().String(),
			UserID:      &s.testData.userID,
			Price:       20.00,
			Status:      "Available",
		},
		{
			MovieShowID: s.testData.movieShowID,
			SeatID:      uuid.New().String(),
			UserID:      nil,
			Price:       25.00,
			Status:      "Available",
		},
	}

	for i, ticket := range tickets {
		seatID := uuid.New().String()
		_, err := s.repo.db.Exec(s.ctx, `
			INSERT INTO seats (id, hall_id, seat_type_id, row_number, seat_number) 
			VALUES ($1, $2, (SELECT id FROM seat_types LIMIT 1), $3, $4)
		`, seatID, s.testData.hallID, i+2, i+2)
		if err != nil {
			s.T().Fatalf("Failed to create seat: %v", err)
		}
		ticket.SeatID = seatID

		created, utilsErr := s.repo.CreateForMovieShow(s.ctx, s.testData.movieShowID, ticket)
		s.assertNoError(utilsErr)
		s.testTickets = append(s.testTickets, created.ID)
	}

	result, total, utilsErr := s.repo.GetAll(s.ctx, domain.TicketFilters{}, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, len(tickets))
	assert.GreaterOrEqual(s.T(), len(result), len(tickets))
}

func (s *TicketRepositoryIntegrationTestSuite) TestGetAllWithStatusFilter() {
	seatID1 := uuid.New().String()
	seatID2 := uuid.New().String()

	_, err := s.repo.db.Exec(s.ctx, `
		INSERT INTO seats (id, hall_id, seat_type_id, row_number, seat_number) 
		VALUES ($1, $2, (SELECT id FROM seat_types LIMIT 1), 3, 1)
	`, seatID1, s.testData.hallID)
	if err != nil {
		s.T().Fatalf("Failed to create seat: %v", err)
	}

	_, err = s.repo.db.Exec(s.ctx, `
		INSERT INTO seats (id, hall_id, seat_type_id, row_number, seat_number) 
		VALUES ($1, $2, (SELECT id FROM seat_types LIMIT 1), 3, 2)
	`, seatID2, s.testData.hallID)
	if err != nil {
		s.T().Fatalf("Failed to create seat: %v", err)
	}

	ticket1 := domain.Ticket{
		MovieShowID: s.testData.movieShowID,
		SeatID:      seatID1,
		UserID:      &s.testData.userID,
		Price:       20.00,
		Status:      "Available",
	}

	ticket2 := domain.Ticket{
		MovieShowID: s.testData.movieShowID,
		SeatID:      seatID2,
		UserID:      &s.testData.userID,
		Price:       25.00,
		Status:      "Purchased",
	}

	created1, utilsErr := s.repo.CreateForMovieShow(s.ctx, s.testData.movieShowID, ticket1)
	s.assertNoError(utilsErr)
	s.testTickets = append(s.testTickets, created1.ID)

	created2, utilsErr := s.repo.CreateForMovieShow(s.ctx, s.testData.movieShowID, ticket2)
	s.assertNoError(utilsErr)
	s.testTickets = append(s.testTickets, created2.ID)

	filters := domain.TicketFilters{
		Status: []string{"Available"},
	}

	result, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)

	for _, ticket := range result {
		assert.Equal(s.T(), "Available", ticket.Status)
	}
}

func (s *TicketRepositoryIntegrationTestSuite) TestGetAllWithMultipleStatusFilter() {
	seatID1 := uuid.New().String()
	seatID2 := uuid.New().String()
	seatID3 := uuid.New().String()

	for i, seatID := range []string{seatID1, seatID2, seatID3} {
		_, err := s.repo.db.Exec(s.ctx, `
			INSERT INTO seats (id, hall_id, seat_type_id, row_number, seat_number) 
			VALUES ($1, $2, (SELECT id FROM seat_types LIMIT 1), 4, $3)
		`, seatID, s.testData.hallID, i+1)
		if err != nil {
			s.T().Fatalf("Failed to create seat: %v", err)
		}
	}

	tickets := []domain.Ticket{
		{
			MovieShowID: s.testData.movieShowID,
			SeatID:      seatID1,
			UserID:      &s.testData.userID,
			Price:       20.00,
			Status:      "Available",
		},
		{
			MovieShowID: s.testData.movieShowID,
			SeatID:      seatID2,
			UserID:      &s.testData.userID,
			Price:       25.00,
			Status:      "Purchased",
		},
		{
			MovieShowID: s.testData.movieShowID,
			SeatID:      seatID3,
			UserID:      &s.testData.userID,
			Price:       30.00,
			Status:      "Reserved",
		},
	}

	for _, ticket := range tickets {
		created, utilsErr := s.repo.CreateForMovieShow(s.ctx, s.testData.movieShowID, ticket)
		s.assertNoError(utilsErr)
		s.testTickets = append(s.testTickets, created.ID)
	}

	filters := domain.TicketFilters{
		Status: []string{"Available", "Purchased"},
	}

	result, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 2)

	for _, ticket := range result {
		assert.True(s.T(), ticket.Status == "Available" || ticket.Status == "Purchased")
	}
}

func (s *TicketRepositoryIntegrationTestSuite) TestGetAllWithPriceFilter() {
	seatID1 := uuid.New().String()
	seatID2 := uuid.New().String()
	seatID3 := uuid.New().String()

	for i, seatID := range []string{seatID1, seatID2, seatID3} {
		_, err := s.repo.db.Exec(s.ctx, `
			INSERT INTO seats (id, hall_id, seat_type_id, row_number, seat_number) 
			VALUES ($1, $2, (SELECT id FROM seat_types LIMIT 1), 5, $3)
		`, seatID, s.testData.hallID, i+1)
		if err != nil {
			s.T().Fatalf("Failed to create seat: %v", err)
		}
	}

	tickets := []domain.Ticket{
		{
			MovieShowID: s.testData.movieShowID,
			SeatID:      seatID1,
			UserID:      &s.testData.userID,
			Price:       15.00,
			Status:      "Available",
		},
		{
			MovieShowID: s.testData.movieShowID,
			SeatID:      seatID2,
			UserID:      &s.testData.userID,
			Price:       25.00,
			Status:      "Available",
		},
		{
			MovieShowID: s.testData.movieShowID,
			SeatID:      seatID3,
			UserID:      &s.testData.userID,
			Price:       35.00,
			Status:      "Available",
		},
	}

	for _, ticket := range tickets {
		created, utilsErr := s.repo.CreateForMovieShow(s.ctx, s.testData.movieShowID, ticket)
		s.assertNoError(utilsErr)
		s.testTickets = append(s.testTickets, created.ID)
	}

	filters := domain.TicketFilters{
		PriceMin: 20.00,
		PriceMax: 30.00,
	}

	result, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)

	for _, ticket := range result {
		assert.GreaterOrEqual(s.T(), ticket.Price, 20.00)
		assert.LessOrEqual(s.T(), ticket.Price, 30.00)
	}
}

func (s *TicketRepositoryIntegrationTestSuite) TestGetAllWithMovieShowFilter() {
	seatID := uuid.New().String()

	_, err := s.repo.db.Exec(s.ctx, `
		INSERT INTO seats (id, hall_id, seat_type_id, row_number, seat_number) 
		VALUES ($1, $2, (SELECT id FROM seat_types LIMIT 1), 6, 1)
	`, seatID, s.testData.hallID)
	if err != nil {
		s.T().Fatalf("Failed to create seat: %v", err)
	}

	ticket := domain.Ticket{
		MovieShowID: s.testData.movieShowID,
		SeatID:      seatID,
		UserID:      &s.testData.userID,
		Price:       20.00,
		Status:      "Available",
	}

	created, utilsErr := s.repo.CreateForMovieShow(s.ctx, s.testData.movieShowID, ticket)
	s.assertNoError(utilsErr)
	s.testTickets = append(s.testTickets, created.ID)

	filters := domain.TicketFilters{
		MovieShowID: s.testData.movieShowID,
	}

	result, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)

	for _, ticket := range result {
		assert.Equal(s.T(), s.testData.movieShowID, ticket.MovieShowID)
	}
}

func (s *TicketRepositoryIntegrationTestSuite) TestGetAllWithUserFilter() {
	seatID := uuid.New().String()

	_, err := s.repo.db.Exec(s.ctx, `
		INSERT INTO seats (id, hall_id, seat_type_id, row_number, seat_number) 
		VALUES ($1, $2, (SELECT id FROM seat_types LIMIT 1), 7, 1)
	`, seatID, s.testData.hallID)
	if err != nil {
		s.T().Fatalf("Failed to create seat: %v", err)
	}

	ticket := domain.Ticket{
		MovieShowID: s.testData.movieShowID,
		SeatID:      seatID,
		UserID:      &s.testData.userID,
		Price:       20.00,
		Status:      "Available",
	}

	created, utilsErr := s.repo.CreateForMovieShow(s.ctx, s.testData.movieShowID, ticket)
	s.assertNoError(utilsErr)
	s.testTickets = append(s.testTickets, created.ID)

	filters := domain.TicketFilters{
		UserID: []string{s.testData.userID},
	}

	result, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)

	for _, ticket := range result {
		assert.Equal(s.T(), s.testData.userID, *ticket.UserID)
	}
}

func (s *TicketRepositoryIntegrationTestSuite) TestGetAllWithSeatFilter() {
	seatID := uuid.New().String()

	_, err := s.repo.db.Exec(s.ctx, `
		INSERT INTO seats (id, hall_id, seat_type_id, row_number, seat_number) 
		VALUES ($1, $2, (SELECT id FROM seat_types LIMIT 1), 8, 1)
	`, seatID, s.testData.hallID)
	if err != nil {
		s.T().Fatalf("Failed to create seat: %v", err)
	}

	ticket := domain.Ticket{
		MovieShowID: s.testData.movieShowID,
		SeatID:      seatID,
		UserID:      &s.testData.userID,
		Price:       20.00,
		Status:      "Available",
	}

	created, utilsErr := s.repo.CreateForMovieShow(s.ctx, s.testData.movieShowID, ticket)
	s.assertNoError(utilsErr)
	s.testTickets = append(s.testTickets, created.ID)

	filters := domain.TicketFilters{
		SeatID: []string{seatID},
	}

	result, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)

	for _, ticket := range result {
		assert.Equal(s.T(), seatID, ticket.SeatID)
	}
}

func (s *TicketRepositoryIntegrationTestSuite) TestGetAllWithPagination() {
	for i := 1; i <= 5; i++ {
		seatID := uuid.New().String()
		_, err := s.repo.db.Exec(s.ctx, `
			INSERT INTO seats (id, hall_id, seat_type_id, row_number, seat_number) 
			VALUES ($1, $2, (SELECT id FROM seat_types LIMIT 1), 9, $3)
		`, seatID, s.testData.hallID, i)
		if err != nil {
			s.T().Fatalf("Failed to create seat: %v", err)
		}

		ticket := domain.Ticket{
			MovieShowID: s.testData.movieShowID,
			SeatID:      seatID,
			UserID:      &s.testData.userID,
			Price:       float64(i) * 10.00,
			Status:      "Available",
		}
		created, utilsErr := s.repo.CreateForMovieShow(s.ctx, s.testData.movieShowID, ticket)
		s.assertNoError(utilsErr)
		s.testTickets = append(s.testTickets, created.ID)
	}

	page1, total, utilsErr := s.repo.GetAll(s.ctx, domain.TicketFilters{}, 1, 2)
	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 5)
	assert.Len(s.T(), page1, 2)

	page2, _, utilsErr := s.repo.GetAll(s.ctx, domain.TicketFilters{}, 2, 2)
	s.assertNoError(utilsErr)
	assert.Len(s.T(), page2, 2)

	page3, _, utilsErr := s.repo.GetAll(s.ctx, domain.TicketFilters{}, 3, 2)
	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), len(page3), 1)
}

func (s *TicketRepositoryIntegrationTestSuite) TestGetAllEmptyResult() {
	filters := domain.TicketFilters{
		MovieShowID: uuid.New().String(),
	}

	result, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), 0, total)
	assert.Empty(s.T(), result)
}
