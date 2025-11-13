package postgres

import (
	"context"
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

type ScreenTypeRepositoryIntegrationTestSuite struct {
	suite.Suite
	repo            *ScreenTypeRepository
	ctx             context.Context
	testScreenTypes []string
}

func TestScreenTypeRepositoryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ScreenTypeRepositoryIntegrationTestSuite))
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) SetupSuite() {
	db, err := config.NewTestDatabase()
	if err != nil {
		s.T().Fatalf("Failed to connect to test database: %v", err)
	}
	s.repo = NewScreenTypeRepository(db)
	s.ctx = context.Background()

	s.cleanDatabase()
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) cleanDatabase() {
	for _, id := range s.testScreenTypes {
		_ = s.repo.Delete(s.ctx, id)
	}
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) SetupTest() {
	s.testScreenTypes = []string{}
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) TearDownTest() {
	for _, id := range s.testScreenTypes {
		_ = s.repo.Delete(s.ctx, id)
	}
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) TearDownSuite() {
	s.cleanDatabase()
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) assertNoError(utilsErr *utils.Error) {
	if utilsErr != nil {
		s.T().Errorf("Expected no error, got: %v", utilsErr)
	}
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) assertErrorCode(utilsErr *utils.Error, expectedCode int) {
	if utilsErr == nil {
		s.T().Error("Expected error, got nil")
		return
	}
	assert.Equal(s.T(), expectedCode, utilsErr.Code)
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) TestCreateScreenType() {
	screenType := domain.ScreenType{
		Name:          "Test 4DX " + strconv.FormatInt(time.Now().Unix(), 10),
		Description:   "4D cinema experience with motion seats and effects",
		PriceModifier: 2.5,
	}

	result, utilsErr := s.repo.Create(s.ctx, screenType)

	s.assertNoError(utilsErr)
	assert.NotEmpty(s.T(), result.ID)
	assert.Equal(s.T(), screenType.Name, result.Name)
	assert.Equal(s.T(), "4D cinema experience with motion seats and effects", result.Description)
	assert.Equal(s.T(), 2.5, result.PriceModifier)

	s.testScreenTypes = append(s.testScreenTypes, result.ID)
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) TestCreateScreenTypeWithMinPriceModifier() {
	screenType := domain.ScreenType{
		Name:          "Test Budget " + strconv.FormatInt(time.Now().Unix(), 10),
		Description:   "Budget screen type",
		PriceModifier: 0.1,
	}

	result, utilsErr := s.repo.Create(s.ctx, screenType)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), 0.1, result.PriceModifier)
	s.testScreenTypes = append(s.testScreenTypes, result.ID)
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) TestCreateScreenTypeDuplicateName() {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	screenType := domain.ScreenType{
		Name:          "Test Duplicate " + timestamp,
		Description:   "First screen type",
		PriceModifier: 1.0,
	}

	result1, utilsErr := s.repo.Create(s.ctx, screenType)
	s.assertNoError(utilsErr)
	s.testScreenTypes = append(s.testScreenTypes, result1.ID)

	screenType2 := domain.ScreenType{
		Name:          "Test Duplicate " + timestamp,
		Description:   "Second screen type with same name",
		PriceModifier: 1.5,
	}

	_, utilsErr = s.repo.Create(s.ctx, screenType2)

	s.assertErrorCode(utilsErr, 409)
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) TestGetScreenTypeByID() {
	screenType := domain.ScreenType{
		Name:          "Test Get Screen " + strconv.FormatInt(time.Now().Unix(), 10),
		Description:   "Screen type for get test",
		PriceModifier: 1.8,
	}

	created, utilsErr := s.repo.Create(s.ctx, screenType)
	s.assertNoError(utilsErr)
	s.testScreenTypes = append(s.testScreenTypes, created.ID)

	retrieved, utilsErr := s.repo.GetByID(s.ctx, created.ID)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), created.ID, retrieved.ID)
	assert.Equal(s.T(), screenType.Name, retrieved.Name)
	assert.Equal(s.T(), "Screen type for get test", retrieved.Description)
	assert.Equal(s.T(), 1.8, retrieved.PriceModifier)
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) TestGetScreenTypeByID_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	_, utilsErr := s.repo.GetByID(s.ctx, nonExistentID)

	s.assertErrorCode(utilsErr, 404)
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) TestUpdateScreenType() {
	screenType := domain.ScreenType{
		Name:          "Old Screen Type " + strconv.FormatInt(time.Now().Unix(), 10),
		Description:   "Old description",
		PriceModifier: 1.0,
	}

	created, utilsErr := s.repo.Create(s.ctx, screenType)
	s.assertNoError(utilsErr)
	s.testScreenTypes = append(s.testScreenTypes, created.ID)

	updateData := domain.ScreenType{
		Name:          "Updated Screen Type " + strconv.FormatInt(time.Now().Unix(), 10),
		Description:   "Updated description",
		PriceModifier: 2.0,
	}

	updated, utilsErr := s.repo.Update(s.ctx, created.ID, updateData)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), created.ID, updated.ID)
	assert.Equal(s.T(), updateData.Name, updated.Name)
	assert.Equal(s.T(), "Updated description", updated.Description)
	assert.Equal(s.T(), 2.0, updated.PriceModifier)
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) TestUpdateScreenType_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	updateData := domain.ScreenType{
		Name:          "Should not update",
		Description:   "Should not update",
		PriceModifier: 1.0,
	}

	_, utilsErr := s.repo.Update(s.ctx, nonExistentID, updateData)

	s.assertErrorCode(utilsErr, 404)
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) TestDeleteScreenType() {
	screenType := domain.ScreenType{
		Name:          "Screen to delete " + strconv.FormatInt(time.Now().Unix(), 10),
		Description:   "Will be deleted",
		PriceModifier: 1.5,
	}

	created, utilsErr := s.repo.Create(s.ctx, screenType)
	s.assertNoError(utilsErr)

	utilsErr = s.repo.Delete(s.ctx, created.ID)
	s.assertNoError(utilsErr)

	_, utilsErr = s.repo.GetByID(s.ctx, created.ID)
	s.assertErrorCode(utilsErr, 404)
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) TestDeleteScreenType_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	utilsErr := s.repo.Delete(s.ctx, nonExistentID)

	s.assertErrorCode(utilsErr, 404)
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) TestGetAllScreenTypes() {
	screenTypes := []domain.ScreenType{
		{
			Name:          "Test Screen A " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   "Description A",
			PriceModifier: 1.2,
		},
		{
			Name:          "Test Screen B " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   "Description B",
			PriceModifier: 1.8,
		},
	}

	for _, screenType := range screenTypes {
		created, utilsErr := s.repo.Create(s.ctx, screenType)
		s.assertNoError(utilsErr)
		s.testScreenTypes = append(s.testScreenTypes, created.ID)
	}

	allScreenTypes, total, utilsErr := s.repo.GetAll(s.ctx, domain.ScreenTypeFilters{}, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, len(screenTypes))
	assert.GreaterOrEqual(s.T(), len(allScreenTypes), len(screenTypes))
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) TestGetAllScreenTypesWithPagination() {
	for i := 0; i < 5; i++ {
		screenType := domain.ScreenType{
			Name:          "Test Screen " + string(rune('A'+i)) + " " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   "Description " + string(rune('A'+i)),
			PriceModifier: 1.0 + float64(i)*0.2,
		}
		created, utilsErr := s.repo.Create(s.ctx, screenType)
		s.assertNoError(utilsErr)
		s.testScreenTypes = append(s.testScreenTypes, created.ID)
	}

	screenTypesPage1, total, utilsErr := s.repo.GetAll(s.ctx, domain.ScreenTypeFilters{}, 1, 2)
	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 5)
	assert.Len(s.T(), screenTypesPage1, 2)

	screenTypesPage2, _, utilsErr := s.repo.GetAll(s.ctx, domain.ScreenTypeFilters{}, 2, 2)
	s.assertNoError(utilsErr)
	assert.Len(s.T(), screenTypesPage2, 2)
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) TestGetAllScreenTypesWithNameFilter() {
	uniqueName := "Unique Test Screen XYZ " + strconv.FormatInt(time.Now().Unix(), 10)
	screenType := domain.ScreenType{
		Name:          uniqueName,
		Description:   "Unique description",
		PriceModifier: 1.5,
	}

	created, utilsErr := s.repo.Create(s.ctx, screenType)
	s.assertNoError(utilsErr)
	s.testScreenTypes = append(s.testScreenTypes, created.ID)

	filters := domain.ScreenTypeFilters{
		Name: "XYZ",
	}

	screenTypes, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)
	assert.GreaterOrEqual(s.T(), len(screenTypes), 1)

	for _, st := range screenTypes {
		assert.Contains(s.T(), strings.ToLower(st.Name), "xyz")
	}
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) TestGetAllScreenTypesWithDescriptionFilter() {
	screenType := domain.ScreenType{
		Name:          "Special Screen " + strconv.FormatInt(time.Now().Unix(), 10),
		Description:   "This is a special premium screen type with advanced features",
		PriceModifier: 2.0,
	}

	created, utilsErr := s.repo.Create(s.ctx, screenType)
	s.assertNoError(utilsErr)
	s.testScreenTypes = append(s.testScreenTypes, created.ID)

	filters := domain.ScreenTypeFilters{
		Description: "premium",
	}

	screenTypes, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)
	assert.GreaterOrEqual(s.T(), len(screenTypes), 1)

	for _, st := range screenTypes {
		assert.Contains(s.T(), strings.ToLower(st.Description), "premium")
	}
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) TestGetAllScreenTypesOrderByName() {
	screenTypes := []domain.ScreenType{
		{
			Name:          "Zeta Screen " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   "Zeta description",
			PriceModifier: 1.0,
		},
		{
			Name:          "Alpha Screen " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   "Alpha description",
			PriceModifier: 1.0,
		},
		{
			Name:          "Beta Screen " + strconv.FormatInt(time.Now().Unix(), 10),
			Description:   "Beta description",
			PriceModifier: 1.0,
		},
	}

	for _, screenType := range screenTypes {
		created, utilsErr := s.repo.Create(s.ctx, screenType)
		s.assertNoError(utilsErr)
		s.testScreenTypes = append(s.testScreenTypes, created.ID)
	}

	allScreenTypes, total, utilsErr := s.repo.GetAll(s.ctx, domain.ScreenTypeFilters{}, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, len(screenTypes))

	if len(allScreenTypes) >= len(screenTypes) {

		for i := 1; i < len(allScreenTypes); i++ {
			assert.True(s.T(), allScreenTypes[i-1].Name <= allScreenTypes[i].Name)
		}
	}
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) TestCreateScreenTypeWithReasonablePriceModifier() {
	screenType := domain.ScreenType{
		Name:          "Luxury Screen " + strconv.FormatInt(time.Now().Unix(), 10),
		Description:   "Luxury cinema experience",
		PriceModifier: 5.0,
	}

	result, utilsErr := s.repo.Create(s.ctx, screenType)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), 5.0, result.PriceModifier)
	s.testScreenTypes = append(s.testScreenTypes, result.ID)
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) TestUpdateScreenTypeToSameName() {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	screenType := domain.ScreenType{
		Name:          "Original Name " + timestamp,
		Description:   "Original description",
		PriceModifier: 1.0,
	}

	created, utilsErr := s.repo.Create(s.ctx, screenType)
	s.assertNoError(utilsErr)
	s.testScreenTypes = append(s.testScreenTypes, created.ID)

	updateData := domain.ScreenType{
		Name:          "Original Name " + timestamp,
		Description:   "Updated description",
		PriceModifier: 2.0,
	}

	updated, utilsErr := s.repo.Update(s.ctx, created.ID, updateData)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), "Original Name "+timestamp, updated.Name)
	assert.Equal(s.T(), "Updated description", updated.Description)
	assert.Equal(s.T(), 2.0, updated.PriceModifier)
}

func (s *ScreenTypeRepositoryIntegrationTestSuite) TestGetAllScreenTypesEmptyResult() {
	filters := domain.ScreenTypeFilters{
		Name: "NonExistentScreenTypeNameThatDoesNotExist12345",
	}

	screenTypes, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), 0, total)
	assert.Empty(s.T(), screenTypes)
}
