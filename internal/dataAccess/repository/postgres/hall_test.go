package postgres

import (
	"context"
	"testing"

	"cw/internal/dataAccess/config"
	domain "cw/internal/domain/models"
	"cw/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HallRepositoryIntegrationTestSuite struct {
	suite.Suite
	repo            *HallRepository
	screenTypeRepo  *ScreenTypeRepository
	ctx             context.Context
	testHalls       []string
	testScreenTypes []string
}

func TestHallRepositoryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(HallRepositoryIntegrationTestSuite))
}

func (s *HallRepositoryIntegrationTestSuite) SetupSuite() {
	db, err := config.NewTestDatabase()
	if err != nil {
		s.T().Fatalf("Failed to connect to test database: %v", err)
	}
	s.repo = NewHallRepository(db)
	s.screenTypeRepo = NewScreenTypeRepository(db)
	s.ctx = context.Background()

	s.cleanDatabase()
	s.createTestScreenTypes()
}

func (s *HallRepositoryIntegrationTestSuite) createTestScreenTypes() {
	screenTypes := []domain.ScreenType{
		{Name: "2D", Description: "Standard 2D", PriceModifier: 1.0},
		{Name: "3D", Description: "3D Cinema", PriceModifier: 1.5},
		{Name: "IMAX", Description: "IMAX Screen", PriceModifier: 2.0},
	}

	for _, st := range screenTypes {
		created, err := s.screenTypeRepo.Create(s.ctx, st)
		if err != nil {
			s.T().Fatalf("Failed to create test screen type: %v", err)
		}
		s.testScreenTypes = append(s.testScreenTypes, created.ID)
	}
}

func (s *HallRepositoryIntegrationTestSuite) cleanDatabase() {
	_, err := s.repo.db.Exec(s.ctx, "DELETE FROM halls")
	if err != nil {
		s.T().Fatalf("Failed to clean halls: %v", err)
	}
	_, err = s.repo.db.Exec(s.ctx, "DELETE FROM screen_types")
	if err != nil {
		s.T().Fatalf("Failed to clean screen types: %v", err)
	}
}

func (s *HallRepositoryIntegrationTestSuite) SetupTest() {
	s.testHalls = []string{}
}

func (s *HallRepositoryIntegrationTestSuite) TearDownTest() {
	for _, id := range s.testHalls {
		s.repo.Delete(s.ctx, id)
	}
}

func (s *HallRepositoryIntegrationTestSuite) TearDownSuite() {
	// Clean up screen types at the end
	for _, id := range s.testScreenTypes {
		s.screenTypeRepo.Delete(s.ctx, id)
	}
}

func (s *HallRepositoryIntegrationTestSuite) assertNoError(utilsErr *utils.Error) {
	if utilsErr != nil {
		s.T().Errorf("Expected no error, got: %v", utilsErr)
	}
}

func (s *HallRepositoryIntegrationTestSuite) assertErrorCode(utilsErr *utils.Error, expectedCode int) {
	if utilsErr == nil {
		s.T().Error("Expected error, got nil")
		return
	}
	assert.Equal(s.T(), expectedCode, utilsErr.Code)
}

func (s *HallRepositoryIntegrationTestSuite) TestCreateHall() {
	// Use the first test screen type
	screenTypeID := s.testScreenTypes[0]

	hall := domain.Hall{
		Name:         "Test Hall",
		Description:  "Test Description",
		ScreenTypeID: screenTypeID,
	}

	result, utilsErr := s.repo.Create(s.ctx, hall)

	s.assertNoError(utilsErr)
	assert.NotEmpty(s.T(), result.ID)
	assert.Equal(s.T(), "Test Hall", result.Name)
	assert.Equal(s.T(), "Test Description", result.Description)
	assert.Equal(s.T(), screenTypeID, result.ScreenTypeID)

	s.testHalls = append(s.testHalls, result.ID)
}

func (s *HallRepositoryIntegrationTestSuite) TestGetHallByID() {
	screenTypeID := s.testScreenTypes[0]

	hall := domain.Hall{
		Name:         "Test Get Hall",
		Description:  "Test Get Description",
		ScreenTypeID: screenTypeID,
	}

	created, utilsErr := s.repo.Create(s.ctx, hall)
	s.assertNoError(utilsErr)
	s.testHalls = append(s.testHalls, created.ID)

	retrieved, utilsErr := s.repo.GetByID(s.ctx, created.ID)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), created.ID, retrieved.ID)
	assert.Equal(s.T(), "Test Get Hall", retrieved.Name)
	assert.Equal(s.T(), "Test Get Description", retrieved.Description)
	assert.Equal(s.T(), screenTypeID, retrieved.ScreenTypeID)
}

func (s *HallRepositoryIntegrationTestSuite) TestGetHallByID_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	_, utilsErr := s.repo.GetByID(s.ctx, nonExistentID)

	s.assertErrorCode(utilsErr, 404)
}

func (s *HallRepositoryIntegrationTestSuite) TestUpdateHall() {
	screenTypeID1 := s.testScreenTypes[0]
	screenTypeID2 := s.testScreenTypes[1]

	hall := domain.Hall{
		Name:         "Old Name",
		Description:  "Old Description",
		ScreenTypeID: screenTypeID1,
	}

	created, utilsErr := s.repo.Create(s.ctx, hall)
	s.assertNoError(utilsErr)
	s.testHalls = append(s.testHalls, created.ID)

	updateData := domain.Hall{
		Name:         "Updated Name",
		Description:  "Updated Description",
		ScreenTypeID: screenTypeID2,
	}

	updated, utilsErr := s.repo.Update(s.ctx, created.ID, updateData)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), created.ID, updated.ID)
	assert.Equal(s.T(), "Updated Name", updated.Name)
	assert.Equal(s.T(), "Updated Description", updated.Description)
	assert.Equal(s.T(), screenTypeID2, updated.ScreenTypeID)
}

func (s *HallRepositoryIntegrationTestSuite) TestDeleteHall() {
	screenTypeID := s.testScreenTypes[0]

	hall := domain.Hall{
		Name:         "Delete Test",
		Description:  "Delete Description",
		ScreenTypeID: screenTypeID,
	}

	created, utilsErr := s.repo.Create(s.ctx, hall)
	s.assertNoError(utilsErr)

	utilsErr = s.repo.Delete(s.ctx, created.ID)
	s.assertNoError(utilsErr)

	_, utilsErr = s.repo.GetByID(s.ctx, created.ID)
	s.assertErrorCode(utilsErr, 404)
}

func (s *HallRepositoryIntegrationTestSuite) TestGetAllHalls() {
	screenTypeID1 := s.testScreenTypes[0]
	screenTypeID2 := s.testScreenTypes[1]

	hall1 := domain.Hall{
		Name:         "Hall A",
		Description:  "Description A",
		ScreenTypeID: screenTypeID1,
	}
	hall2 := domain.Hall{
		Name:         "Hall B",
		Description:  "Description B",
		ScreenTypeID: screenTypeID2,
	}

	created1, utilsErr := s.repo.Create(s.ctx, hall1)
	s.assertNoError(utilsErr)
	created2, utilsErr := s.repo.Create(s.ctx, hall2)
	s.assertNoError(utilsErr)

	s.testHalls = append(s.testHalls, created1.ID, created2.ID)

	halls, total, utilsErr := s.repo.GetAll(s.ctx, domain.HallFilters{}, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 2)
	assert.GreaterOrEqual(s.T(), len(halls), 2)
}

func (s *HallRepositoryIntegrationTestSuite) TestGetAllHallsWithNameFilter() {
	screenTypeID := s.testScreenTypes[0]

	hall := domain.Hall{
		Name:         "Unique Filter Hall",
		Description:  "Unique Description",
		ScreenTypeID: screenTypeID,
	}

	created, utilsErr := s.repo.Create(s.ctx, hall)
	s.assertNoError(utilsErr)
	s.testHalls = append(s.testHalls, created.ID)

	filters := domain.HallFilters{
		Name: "Unique Filter",
	}

	halls, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)
	assert.GreaterOrEqual(s.T(), len(halls), 1)
}

func (s *HallRepositoryIntegrationTestSuite) TestGetAllHallsWithScreenTypeFilter() {
	screenTypeID := s.testScreenTypes[0]

	hall := domain.Hall{
		Name:         "Screen Type Hall",
		Description:  "Screen Type Description",
		ScreenTypeID: screenTypeID,
	}

	created, utilsErr := s.repo.Create(s.ctx, hall)
	s.assertNoError(utilsErr)
	s.testHalls = append(s.testHalls, created.ID)

	filters := domain.HallFilters{
		ScreenTypeID: screenTypeID,
	}

	halls, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)
	assert.GreaterOrEqual(s.T(), len(halls), 1)
}

func (s *HallRepositoryIntegrationTestSuite) TestCreateDuplicateHall() {
	screenTypeID := s.testScreenTypes[0]

	hall := domain.Hall{
		Name:         "Duplicate Hall",
		Description:  "Duplicate Description",
		ScreenTypeID: screenTypeID,
	}

	created, utilsErr := s.repo.Create(s.ctx, hall)
	s.assertNoError(utilsErr)
	s.testHalls = append(s.testHalls, created.ID)

	_, utilsErr = s.repo.Create(s.ctx, hall)

	s.assertErrorCode(utilsErr, 409)
}
