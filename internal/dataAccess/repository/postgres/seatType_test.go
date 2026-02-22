//go:build integration
// +build integration

package postgres

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"cw/internal/dataAccess/config"
	domain "cw/internal/domain/models"
	"cw/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SeatTypeRepositoryIntegrationTestSuite struct {
	suite.Suite
	repo          *SeatTypeRepository
	ctx           context.Context
	testSeatTypes []string
}

func TestSeatTypeRepositoryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(SeatTypeRepositoryIntegrationTestSuite))
}

func (s *SeatTypeRepositoryIntegrationTestSuite) SetupSuite() {
	db, err := config.NewTestDatabase()
	if err != nil {
		s.T().Fatalf("Failed to connect to test database: %v", err)
	}
	s.repo = NewSeatTypeRepository(db)
	s.ctx = context.Background()
	s.cleanDatabase()
}

func (s *SeatTypeRepositoryIntegrationTestSuite) cleanDatabase() {

	for _, id := range s.testSeatTypes {
		_ = s.repo.Delete(s.ctx, id)
	}
	s.testSeatTypes = []string{}
}

func (s *SeatTypeRepositoryIntegrationTestSuite) SetupTest() {
	s.testSeatTypes = []string{}
}

func (s *SeatTypeRepositoryIntegrationTestSuite) TearDownTest() {
	s.cleanDatabase()
}

func (s *SeatTypeRepositoryIntegrationTestSuite) TearDownSuite() {
	s.cleanDatabase()
}

func (s *SeatTypeRepositoryIntegrationTestSuite) assertNoError(utilsErr *utils.Error) {
	if utilsErr != nil {
		s.T().Errorf("Expected no error, got: %v", utilsErr)
	}
}

func (s *SeatTypeRepositoryIntegrationTestSuite) assertErrorCode(utilsErr *utils.Error, expectedCode int) {
	if utilsErr == nil {
		s.T().Error("Expected error, got nil")
		return
	}
	assert.Equal(s.T(), expectedCode, utilsErr.Code)
}

func (s *SeatTypeRepositoryIntegrationTestSuite) TestCreateSeatType() {
	seatType := domain.SeatType{
		Name:          "Test Standard " + strconv.FormatInt(time.Now().Unix(), 10),
		Description:   "Test standard seat type",
		PriceModifier: 1.0,
	}

	result, utilsErr := s.repo.Create(s.ctx, seatType)

	s.assertNoError(utilsErr)
	assert.NotEmpty(s.T(), result.ID)
	assert.Equal(s.T(), seatType.Name, result.Name)
	assert.Equal(s.T(), seatType.Description, result.Description)
	assert.Equal(s.T(), seatType.PriceModifier, result.PriceModifier)

	s.testSeatTypes = append(s.testSeatTypes, result.ID)
}

func (s *SeatTypeRepositoryIntegrationTestSuite) TestCreateSeatTypeDuplicateName() {
	seatType := domain.SeatType{
		Name:          "Unique Seat Type " + strconv.FormatInt(time.Now().Unix(), 10),
		Description:   "Test seat type",
		PriceModifier: 1.0,
	}

	result1, utilsErr := s.repo.Create(s.ctx, seatType)
	s.assertNoError(utilsErr)
	s.testSeatTypes = append(s.testSeatTypes, result1.ID)

	seatType2 := domain.SeatType{
		Name:          seatType.Name,
		Description:   "Different description",
		PriceModifier: 1.5,
	}

	_, utilsErr = s.repo.Create(s.ctx, seatType2)

	s.assertErrorCode(utilsErr, 409)
}

func (s *SeatTypeRepositoryIntegrationTestSuite) TestGetSeatTypeByID() {
	seatType := domain.SeatType{
		Name:          "Test Get By ID " + strconv.FormatInt(time.Now().Unix(), 10),
		Description:   "Test seat type for GetByID",
		PriceModifier: 1.2,
	}

	created, utilsErr := s.repo.Create(s.ctx, seatType)
	s.assertNoError(utilsErr)
	s.testSeatTypes = append(s.testSeatTypes, created.ID)

	retrieved, utilsErr := s.repo.GetByID(s.ctx, created.ID)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), created.ID, retrieved.ID)
	assert.Equal(s.T(), seatType.Name, retrieved.Name)
	assert.Equal(s.T(), seatType.Description, retrieved.Description)
	assert.Equal(s.T(), seatType.PriceModifier, retrieved.PriceModifier)
}

func (s *SeatTypeRepositoryIntegrationTestSuite) TestGetSeatTypeByID_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	_, utilsErr := s.repo.GetByID(s.ctx, nonExistentID)

	s.assertErrorCode(utilsErr, 404)
}

func (s *SeatTypeRepositoryIntegrationTestSuite) TestUpdateSeatType() {
	seatType := domain.SeatType{
		Name:          "Test Update " + strconv.FormatInt(time.Now().Unix(), 10),
		Description:   "Original description",
		PriceModifier: 1.0,
	}

	created, utilsErr := s.repo.Create(s.ctx, seatType)
	s.assertNoError(utilsErr)
	s.testSeatTypes = append(s.testSeatTypes, created.ID)

	updateData := domain.SeatType{
		Name:          "Updated Name " + strconv.FormatInt(time.Now().Unix(), 10),
		Description:   "Updated description",
		PriceModifier: 1.8,
	}

	updated, utilsErr := s.repo.Update(s.ctx, created.ID, updateData)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), created.ID, updated.ID)
	assert.Equal(s.T(), updateData.Name, updated.Name)
	assert.Equal(s.T(), updateData.Description, updated.Description)
	assert.Equal(s.T(), updateData.PriceModifier, updated.PriceModifier)
}

func (s *SeatTypeRepositoryIntegrationTestSuite) TestUpdateSeatType_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	updateData := domain.SeatType{
		Name:          "Non-existent",
		Description:   "Should not exist",
		PriceModifier: 1.0,
	}

	_, utilsErr := s.repo.Update(s.ctx, nonExistentID, updateData)

	s.assertErrorCode(utilsErr, 404)
}

func (s *SeatTypeRepositoryIntegrationTestSuite) TestDeleteSeatType() {
	seatType := domain.SeatType{
		Name:          "Test Delete " + strconv.FormatInt(time.Now().Unix(), 10),
		Description:   "Test seat type for deletion",
		PriceModifier: 1.0,
	}

	created, utilsErr := s.repo.Create(s.ctx, seatType)
	s.assertNoError(utilsErr)

	utilsErr = s.repo.Delete(s.ctx, created.ID)
	s.assertNoError(utilsErr)

	_, utilsErr = s.repo.GetByID(s.ctx, created.ID)
	s.assertErrorCode(utilsErr, 404)
}

func (s *SeatTypeRepositoryIntegrationTestSuite) TestDeleteSeatType_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	utilsErr := s.repo.Delete(s.ctx, nonExistentID)

	s.assertErrorCode(utilsErr, 404)
}

func (s *SeatTypeRepositoryIntegrationTestSuite) TestGetAllSeatTypes() {

	seatTypes := []domain.SeatType{
		{
			Name:          "Standard " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   "Standard seat type",
			PriceModifier: 1.0,
		},
		{
			Name:          "Premium " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   "Premium seat type",
			PriceModifier: 1.5,
		},
		{
			Name:          "VIP " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   "VIP seat type",
			PriceModifier: 2.0,
		},
	}

	for _, seatType := range seatTypes {
		created, utilsErr := s.repo.Create(s.ctx, seatType)
		s.assertNoError(utilsErr)
		s.testSeatTypes = append(s.testSeatTypes, created.ID)
	}

	result, total, utilsErr := s.repo.GetAll(s.ctx, domain.SeatTypeFilters{}, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, len(seatTypes))
	assert.GreaterOrEqual(s.T(), len(result), len(seatTypes))

	foundCount := 0
	for _, createdType := range seatTypes {
		for _, resultType := range result {
			if resultType.Name == createdType.Name {
				foundCount++
				break
			}
		}
	}
	assert.GreaterOrEqual(s.T(), foundCount, len(seatTypes))
}

func (s *SeatTypeRepositoryIntegrationTestSuite) TestGetAllWithNameFilter() {

	seatTypes := []domain.SeatType{
		{
			Name:          "Standard Test " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   "Standard seat",
			PriceModifier: 1.0,
		},
		{
			Name:          "Premium Test " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   "Premium seat",
			PriceModifier: 1.5,
		},
		{
			Name:          "VIP Special " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   "VIP seat",
			PriceModifier: 2.0,
		},
	}

	for _, seatType := range seatTypes {
		created, utilsErr := s.repo.Create(s.ctx, seatType)
		s.assertNoError(utilsErr)
		s.testSeatTypes = append(s.testSeatTypes, created.ID)
	}

	filters := domain.SeatTypeFilters{
		Name: "Premium",
	}

	result, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)

	for _, seatType := range result {
		assert.Contains(s.T(), seatType.Name, "Premium")
	}
}

func (s *SeatTypeRepositoryIntegrationTestSuite) TestGetAllWithDescriptionFilter() {
	seatTypes := []domain.SeatType{
		{
			Name:          "Type A " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   "Comfortable standard seat",
			PriceModifier: 1.0,
		},
		{
			Name:          "Type B " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   "Luxury premium seat",
			PriceModifier: 1.5,
		},
		{
			Name:          "Type C " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   "Extra wide VIP seat",
			PriceModifier: 2.0,
		},
	}

	for _, seatType := range seatTypes {
		created, utilsErr := s.repo.Create(s.ctx, seatType)
		s.assertNoError(utilsErr)
		s.testSeatTypes = append(s.testSeatTypes, created.ID)
	}

	filters := domain.SeatTypeFilters{
		Description: "luxury",
	}

	result, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)

	for _, seatType := range result {
		assert.Contains(s.T(), strings.ToLower(seatType.Description), "luxury")
	}
}

func (s *SeatTypeRepositoryIntegrationTestSuite) TestGetAllWithPagination() {

	for i := 1; i <= 5; i++ {
		seatType := domain.SeatType{
			Name:          fmt.Sprintf("Seat Type %d ", i) + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   fmt.Sprintf("Description for seat type %d", i),
			PriceModifier: float64(i) * 0.5,
		}
		created, utilsErr := s.repo.Create(s.ctx, seatType)
		s.assertNoError(utilsErr)
		s.testSeatTypes = append(s.testSeatTypes, created.ID)
	}

	page1, total, utilsErr := s.repo.GetAll(s.ctx, domain.SeatTypeFilters{}, 1, 2)
	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 5)
	assert.Len(s.T(), page1, 2)

	page2, _, utilsErr := s.repo.GetAll(s.ctx, domain.SeatTypeFilters{}, 2, 2)
	s.assertNoError(utilsErr)
	assert.Len(s.T(), page2, 2)

	page3, _, utilsErr := s.repo.GetAll(s.ctx, domain.SeatTypeFilters{}, 3, 2)
	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), len(page3), 1)
}

func (s *SeatTypeRepositoryIntegrationTestSuite) TestGetAllOrderByName() {

	seatTypes := []domain.SeatType{
		{
			Name:          "Zebra " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   "Z type",
			PriceModifier: 1.0,
		},
		{
			Name:          "Alpha " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   "A type",
			PriceModifier: 1.0,
		},
		{
			Name:          "Charlie " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   "C type",
			PriceModifier: 1.0,
		},
	}

	for _, seatType := range seatTypes {
		created, utilsErr := s.repo.Create(s.ctx, seatType)
		s.assertNoError(utilsErr)
		s.testSeatTypes = append(s.testSeatTypes, created.ID)
	}

	result, total, utilsErr := s.repo.GetAll(s.ctx, domain.SeatTypeFilters{}, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, len(seatTypes))

	if len(result) >= len(seatTypes) {
		for i := 1; i < len(result); i++ {
			assert.True(s.T(), result[i-1].Name <= result[i].Name,
				"Results should be sorted by name: %s should come before %s",
				result[i-1].Name, result[i].Name)
		}
	}
}

func (s *SeatTypeRepositoryIntegrationTestSuite) TestGetAllEmptyResult() {
	filters := domain.SeatTypeFilters{
		Name: "NonExistentSeatTypeNameThatShouldNotExist",
	}

	result, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), 0, total)
	assert.Empty(s.T(), result)
}
