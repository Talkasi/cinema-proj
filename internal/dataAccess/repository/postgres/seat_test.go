package postgres

import (
	"context"
	"strconv"
	"testing"
	"time"

	"cw/internal/dataAccess/config"
	domain "cw/internal/domain/models"
	"cw/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SeatRepositoryIntegrationTestSuite struct {
	suite.Suite
	repo            *SeatRepository
	hallRepo        *HallRepository
	seatTypeRepo    *SeatTypeRepository
	screenTypeRepo  *ScreenTypeRepository
	ctx             context.Context
	testSeats       []string
	testHalls       []string
	testSeatTypes   []string
	testScreenTypes []string
}

func TestSeatRepositoryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(SeatRepositoryIntegrationTestSuite))
}

func (s *SeatRepositoryIntegrationTestSuite) SetupSuite() {
	db, err := config.NewTestDatabase()
	if err != nil {
		s.T().Fatalf("Failed to connect to test database: %v", err)
	}
	s.repo = NewSeatRepository(db)
	s.hallRepo = NewHallRepository(db)
	s.seatTypeRepo = NewSeatTypeRepository(db)
	s.screenTypeRepo = NewScreenTypeRepository(db)
	s.ctx = context.Background()

	s.cleanDatabase()
	s.createTestScreenTypes()
	s.createTestHalls()
	s.createTestSeatTypes()
}

func (s *SeatRepositoryIntegrationTestSuite) createTestScreenTypes() {
	// Создаем тестовые типы экранов
	screenTypes := []domain.ScreenType{
		{
			Name:          "Test Screen Type 1 " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   "Test screen type for hall",
			PriceModifier: 1.0,
		},
		{
			Name:          "Test Screen Type 2 " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   "Another test screen type",
			PriceModifier: 1.5,
		},
	}

	for _, screenType := range screenTypes {
		created, err := s.screenTypeRepo.Create(s.ctx, screenType)
		if err != nil {
			s.T().Fatalf("Failed to create test screen type: %v", err)
		}
		s.testScreenTypes = append(s.testScreenTypes, created.ID)
	}
}

func (s *SeatRepositoryIntegrationTestSuite) createTestHalls() {
	// Создаем тестовые залы с нашими screen types
	halls := []domain.Hall{
		{
			Name:         "Test Hall 1 " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:  "Test hall for seat testing",
			ScreenTypeID: s.testScreenTypes[0],
		},
		{
			Name:         "Test Hall 2 " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:  "Another test hall for seat testing",
			ScreenTypeID: s.testScreenTypes[1],
		},
	}

	for _, hall := range halls {
		created, err := s.hallRepo.Create(s.ctx, hall)
		if err != nil {
			s.T().Fatalf("Failed to create test hall: %v", err)
		}
		s.testHalls = append(s.testHalls, created.ID)
	}
}

func (s *SeatRepositoryIntegrationTestSuite) createTestSeatTypes() {
	// Создаем тестовые типы мест
	seatTypes := []domain.SeatType{
		{
			Name:          "Test Standard " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   "Test standard seat type",
			PriceModifier: 1.0,
		},
		{
			Name:          "Test Premium " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   "Test premium seat type",
			PriceModifier: 1.5,
		},
	}

	for _, seatType := range seatTypes {
		created, err := s.seatTypeRepo.Create(s.ctx, seatType)
		if err != nil {
			s.T().Fatalf("Failed to create test seat type: %v", err)
		}
		s.testSeatTypes = append(s.testSeatTypes, created.ID)
	}
}

func (s *SeatRepositoryIntegrationTestSuite) cleanDatabase() {
	// Удаляем только тестовые данные в правильном порядке
	for _, id := range s.testSeats {
		s.repo.Delete(s.ctx, id)
	}
	for _, id := range s.testHalls {
		s.hallRepo.Delete(s.ctx, id)
	}
	for _, id := range s.testSeatTypes {
		s.seatTypeRepo.Delete(s.ctx, id)
	}
	for _, id := range s.testScreenTypes {
		s.screenTypeRepo.Delete(s.ctx, id)
	}
}

func (s *SeatRepositoryIntegrationTestSuite) SetupTest() {
	s.testSeats = []string{}
}

func (s *SeatRepositoryIntegrationTestSuite) TearDownTest() {
	for _, id := range s.testSeats {
		s.repo.Delete(s.ctx, id)
	}
}

func (s *SeatRepositoryIntegrationTestSuite) TearDownSuite() {
	s.cleanDatabase()
}

func (s *SeatRepositoryIntegrationTestSuite) assertNoError(utilsErr *utils.Error) {
	if utilsErr != nil {
		s.T().Errorf("Expected no error, got: %v", utilsErr)
	}
}

func (s *SeatRepositoryIntegrationTestSuite) assertErrorCode(utilsErr *utils.Error, expectedCode int) {
	if utilsErr == nil {
		s.T().Error("Expected error, got nil")
		return
	}
	assert.Equal(s.T(), expectedCode, utilsErr.Code)
}

func (s *SeatRepositoryIntegrationTestSuite) TestCreateSeat() {
	seat := domain.Seat{
		HallID:     s.testHalls[0],
		SeatTypeID: s.testSeatTypes[0],
		RowNumber:  1,
		SeatNumber: 1,
	}

	result, utilsErr := s.repo.Create(s.ctx, seat)

	s.assertNoError(utilsErr)
	assert.NotEmpty(s.T(), result.ID)
	assert.Equal(s.T(), s.testHalls[0], result.HallID)
	assert.Equal(s.T(), s.testSeatTypes[0], result.SeatTypeID)
	assert.Equal(s.T(), 1, result.RowNumber)
	assert.Equal(s.T(), 1, result.SeatNumber)

	s.testSeats = append(s.testSeats, result.ID)
}

func (s *SeatRepositoryIntegrationTestSuite) TestCreateSeatDuplicatePosition() {
	seat := domain.Seat{
		HallID:     s.testHalls[0],
		SeatTypeID: s.testSeatTypes[0],
		RowNumber:  2,
		SeatNumber: 2,
	}

	result1, utilsErr := s.repo.Create(s.ctx, seat)
	s.assertNoError(utilsErr)
	s.testSeats = append(s.testSeats, result1.ID)

	seat2 := domain.Seat{
		HallID:     s.testHalls[0],
		SeatTypeID: s.testSeatTypes[1],
		RowNumber:  2,
		SeatNumber: 2,
	}

	_, utilsErr = s.repo.Create(s.ctx, seat2)

	s.assertErrorCode(utilsErr, 409)
}

func (s *SeatRepositoryIntegrationTestSuite) TestGetSeatByID() {
	seat := domain.Seat{
		HallID:     s.testHalls[0],
		SeatTypeID: s.testSeatTypes[0],
		RowNumber:  3,
		SeatNumber: 3,
	}

	created, utilsErr := s.repo.Create(s.ctx, seat)
	s.assertNoError(utilsErr)
	s.testSeats = append(s.testSeats, created.ID)

	retrieved, utilsErr := s.repo.GetByID(s.ctx, created.ID)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), created.ID, retrieved.ID)
	assert.Equal(s.T(), s.testHalls[0], retrieved.HallID)
	assert.Equal(s.T(), s.testSeatTypes[0], retrieved.SeatTypeID)
	assert.Equal(s.T(), 3, retrieved.RowNumber)
	assert.Equal(s.T(), 3, retrieved.SeatNumber)
}

func (s *SeatRepositoryIntegrationTestSuite) TestGetSeatByID_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	_, utilsErr := s.repo.GetByID(s.ctx, nonExistentID)

	s.assertErrorCode(utilsErr, 404)
}

func (s *SeatRepositoryIntegrationTestSuite) TestUpdateSeat() {
	seat := domain.Seat{
		HallID:     s.testHalls[0],
		SeatTypeID: s.testSeatTypes[0],
		RowNumber:  4,
		SeatNumber: 4,
	}

	created, utilsErr := s.repo.Create(s.ctx, seat)
	s.assertNoError(utilsErr)
	s.testSeats = append(s.testSeats, created.ID)

	updateData := domain.Seat{
		HallID:     s.testHalls[0],
		SeatTypeID: s.testSeatTypes[1],
		RowNumber:  5,
		SeatNumber: 5,
	}

	updated, utilsErr := s.repo.Update(s.ctx, created.ID, updateData)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), created.ID, updated.ID)
	assert.Equal(s.T(), s.testHalls[0], updated.HallID)
	assert.Equal(s.T(), s.testSeatTypes[1], updated.SeatTypeID)
	assert.Equal(s.T(), 5, updated.RowNumber)
	assert.Equal(s.T(), 5, updated.SeatNumber)
}

func (s *SeatRepositoryIntegrationTestSuite) TestUpdateSeat_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	updateData := domain.Seat{
		HallID:     s.testHalls[0],
		SeatTypeID: s.testSeatTypes[0],
		RowNumber:  1,
		SeatNumber: 1,
	}

	_, utilsErr := s.repo.Update(s.ctx, nonExistentID, updateData)

	s.assertErrorCode(utilsErr, 404)
}

func (s *SeatRepositoryIntegrationTestSuite) TestDeleteSeat() {
	seat := domain.Seat{
		HallID:     s.testHalls[0],
		SeatTypeID: s.testSeatTypes[0],
		RowNumber:  6,
		SeatNumber: 6,
	}

	created, utilsErr := s.repo.Create(s.ctx, seat)
	s.assertNoError(utilsErr)

	utilsErr = s.repo.Delete(s.ctx, created.ID)
	s.assertNoError(utilsErr)

	_, utilsErr = s.repo.GetByID(s.ctx, created.ID)
	s.assertErrorCode(utilsErr, 404)
}

func (s *SeatRepositoryIntegrationTestSuite) TestDeleteSeat_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	utilsErr := s.repo.Delete(s.ctx, nonExistentID)

	s.assertErrorCode(utilsErr, 404)
}

func (s *SeatRepositoryIntegrationTestSuite) TestGetByHall() {
	// Создаем несколько мест в одном зале
	seats := []domain.Seat{
		{
			HallID:     s.testHalls[0],
			SeatTypeID: s.testSeatTypes[0],
			RowNumber:  1,
			SeatNumber: 1,
		},
		{
			HallID:     s.testHalls[0],
			SeatTypeID: s.testSeatTypes[0],
			RowNumber:  1,
			SeatNumber: 2,
		},
		{
			HallID:     s.testHalls[0],
			SeatTypeID: s.testSeatTypes[1],
			RowNumber:  2,
			SeatNumber: 1,
		},
	}

	for _, seat := range seats {
		created, utilsErr := s.repo.Create(s.ctx, seat)
		s.assertNoError(utilsErr)
		s.testSeats = append(s.testSeats, created.ID)
	}

	hallSeats, total, utilsErr := s.repo.GetByHall(s.ctx, s.testHalls[0], domain.SeatFilters{}, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, len(seats))
	assert.GreaterOrEqual(s.T(), len(hallSeats), len(seats))

	for _, seat := range hallSeats {
		assert.Equal(s.T(), s.testHalls[0], seat.HallID)
	}
}

func (s *SeatRepositoryIntegrationTestSuite) TestGetByHallWithSeatTypeFilter() {
	// Создаем места разных типов
	seats := []domain.Seat{
		{
			HallID:     s.testHalls[0],
			SeatTypeID: s.testSeatTypes[0],
			RowNumber:  1,
			SeatNumber: 1,
		},
		{
			HallID:     s.testHalls[0],
			SeatTypeID: s.testSeatTypes[1],
			RowNumber:  1,
			SeatNumber: 2,
		},
	}

	for _, seat := range seats {
		created, utilsErr := s.repo.Create(s.ctx, seat)
		s.assertNoError(utilsErr)
		s.testSeats = append(s.testSeats, created.ID)
	}

	filters := domain.SeatFilters{
		SeatTypeID: s.testSeatTypes[0],
	}

	hallSeats, total, utilsErr := s.repo.GetByHall(s.ctx, s.testHalls[0], filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)

	for _, seat := range hallSeats {
		assert.Equal(s.T(), s.testSeatTypes[0], seat.SeatTypeID)
	}
}

func (s *SeatRepositoryIntegrationTestSuite) TestGetByHallWithRowNumberFilter() {
	// Создаем места в разных рядах
	seats := []domain.Seat{
		{
			HallID:     s.testHalls[0],
			SeatTypeID: s.testSeatTypes[0],
			RowNumber:  1,
			SeatNumber: 1,
		},
		{
			HallID:     s.testHalls[0],
			SeatTypeID: s.testSeatTypes[0],
			RowNumber:  5,
			SeatNumber: 1,
		},
		{
			HallID:     s.testHalls[0],
			SeatTypeID: s.testSeatTypes[0],
			RowNumber:  10,
			SeatNumber: 1,
		},
	}

	for _, seat := range seats {
		created, utilsErr := s.repo.Create(s.ctx, seat)
		s.assertNoError(utilsErr)
		s.testSeats = append(s.testSeats, created.ID)
	}

	filters := domain.SeatFilters{
		RowNumberMin: 3,
		RowNumberMax: 8,
	}

	hallSeats, total, utilsErr := s.repo.GetByHall(s.ctx, s.testHalls[0], filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)

	for _, seat := range hallSeats {
		assert.GreaterOrEqual(s.T(), seat.RowNumber, 3)
		assert.LessOrEqual(s.T(), seat.RowNumber, 8)
	}
}

func (s *SeatRepositoryIntegrationTestSuite) TestGetByHallWithSeatNumberFilter() {
	// Создаем места с разными номерами
	seats := []domain.Seat{
		{
			HallID:     s.testHalls[0],
			SeatTypeID: s.testSeatTypes[0],
			RowNumber:  1,
			SeatNumber: 1,
		},
		{
			HallID:     s.testHalls[0],
			SeatTypeID: s.testSeatTypes[0],
			RowNumber:  1,
			SeatNumber: 5,
		},
		{
			HallID:     s.testHalls[0],
			SeatTypeID: s.testSeatTypes[0],
			RowNumber:  1,
			SeatNumber: 10,
		},
	}

	for _, seat := range seats {
		created, utilsErr := s.repo.Create(s.ctx, seat)
		s.assertNoError(utilsErr)
		s.testSeats = append(s.testSeats, created.ID)
	}

	filters := domain.SeatFilters{
		SeatNumberMin: 3,
		SeatNumberMax: 8,
	}

	hallSeats, total, utilsErr := s.repo.GetByHall(s.ctx, s.testHalls[0], filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)

	for _, seat := range hallSeats {
		assert.GreaterOrEqual(s.T(), seat.SeatNumber, 3)
		assert.LessOrEqual(s.T(), seat.SeatNumber, 8)
	}
}

func (s *SeatRepositoryIntegrationTestSuite) TestGetByHallOrderByRowAndSeat() {
	// Создаем места в разном порядке
	seats := []domain.Seat{
		{
			HallID:     s.testHalls[0],
			SeatTypeID: s.testSeatTypes[0],
			RowNumber:  2,
			SeatNumber: 3,
		},
		{
			HallID:     s.testHalls[0],
			SeatTypeID: s.testSeatTypes[0],
			RowNumber:  1,
			SeatNumber: 2,
		},
		{
			HallID:     s.testHalls[0],
			SeatTypeID: s.testSeatTypes[0],
			RowNumber:  1,
			SeatNumber: 1,
		},
	}

	for _, seat := range seats {
		created, utilsErr := s.repo.Create(s.ctx, seat)
		s.assertNoError(utilsErr)
		s.testSeats = append(s.testSeats, created.ID)
	}

	hallSeats, total, utilsErr := s.repo.GetByHall(s.ctx, s.testHalls[0], domain.SeatFilters{}, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, len(seats))

	if len(hallSeats) >= len(seats) {
		// Проверяем сортировку: сначала по ряду, потом по номеру места
		for i := 1; i < len(hallSeats); i++ {
			assert.True(s.T(),
				hallSeats[i-1].RowNumber < hallSeats[i].RowNumber ||
					(hallSeats[i-1].RowNumber == hallSeats[i].RowNumber &&
						hallSeats[i-1].SeatNumber <= hallSeats[i].SeatNumber))
		}
	}
}

func (s *SeatRepositoryIntegrationTestSuite) TestGetByHallWithPagination() {
	// Создаем несколько мест для тестирования пагинации
	for i := 1; i <= 5; i++ {
		seat := domain.Seat{
			HallID:     s.testHalls[0],
			SeatTypeID: s.testSeatTypes[i%len(s.testSeatTypes)],
			RowNumber:  i,
			SeatNumber: i,
		}
		created, utilsErr := s.repo.Create(s.ctx, seat)
		s.assertNoError(utilsErr)
		s.testSeats = append(s.testSeats, created.ID)
	}

	seatsPage1, total, utilsErr := s.repo.GetByHall(s.ctx, s.testHalls[0], domain.SeatFilters{}, 1, 2)
	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 5)
	assert.Len(s.T(), seatsPage1, 2)

	seatsPage2, _, utilsErr := s.repo.GetByHall(s.ctx, s.testHalls[0], domain.SeatFilters{}, 2, 2)
	s.assertNoError(utilsErr)
	assert.Len(s.T(), seatsPage2, 2)
}

func (s *SeatRepositoryIntegrationTestSuite) TestGetByHallEmptyResult() {
	filters := domain.SeatFilters{
		SeatTypeID: "00000000-0000-0000-0000-000000000000",
	}

	seats, total, utilsErr := s.repo.GetByHall(s.ctx, s.testHalls[0], filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), 0, total)
	assert.Empty(s.T(), seats)
}
