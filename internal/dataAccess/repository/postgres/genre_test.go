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

type GenreRepositoryIntegrationTestSuite struct {
	suite.Suite
	repo       *GenreRepository
	ctx        context.Context
	testGenres []string
	baseGenres []domain.Genre
}

func TestGenreRepositoryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(GenreRepositoryIntegrationTestSuite))
}

func (s *GenreRepositoryIntegrationTestSuite) SetupSuite() {
	db, err := config.NewTestDatabase()
	if err != nil {
		s.T().Fatalf("Failed to connect to test database: %v", err)
	}
	s.repo = NewGenreRepository(db)
	s.ctx = context.Background()

	s.cleanDatabase()
	s.seedDatabase()
}

func (s *GenreRepositoryIntegrationTestSuite) seedDatabase() {

	s.baseGenres = []domain.Genre{
		{Name: "Action", Description: "Action movies"},
		{Name: "Comedy", Description: "Comedy movies"},
		{Name: "Drama", Description: "Drama movies"},
		{Name: "Horror", Description: "Horror movies"},
		{Name: "Sci-Fi", Description: "Science fiction movies"},
	}

	for _, genre := range s.baseGenres {
		created, err := s.repo.Create(s.ctx, genre)
		if err != nil {
			s.T().Logf("Warning: failed to seed genre %s: %v", genre.Name, err)
		} else {
			s.baseGenres[s.getBaseGenreIndex(genre.Name)].ID = created.ID
		}
	}
}

func (s *GenreRepositoryIntegrationTestSuite) getBaseGenreIndex(name string) int {
	for i, genre := range s.baseGenres {
		if genre.Name == name {
			return i
		}
	}
	return -1
}

func (s *GenreRepositoryIntegrationTestSuite) cleanDatabase() {
	_, err := s.repo.db.Exec(s.ctx, "DELETE FROM genres")
	if err != nil {
		s.T().Fatalf("Failed to clean test database: %v", err)
	}
}

func (s *GenreRepositoryIntegrationTestSuite) SetupTest() {
	s.testGenres = []string{}
}

func (s *GenreRepositoryIntegrationTestSuite) TearDownTest() {
	for _, id := range s.testGenres {
		_ = s.repo.Delete(s.ctx, id)
	}
}

func (s *GenreRepositoryIntegrationTestSuite) assertNoError(utilsErr *utils.Error) {
	if utilsErr != nil {
		s.T().Errorf("Expected no error, got: %v", utilsErr)
	}
}

func (s *GenreRepositoryIntegrationTestSuite) assertErrorCode(utilsErr *utils.Error, expectedCode int) {
	if utilsErr == nil {
		s.T().Error("Expected error, got nil")
		return
	}
	assert.Equal(s.T(), expectedCode, utilsErr.Code)
}

func (s *GenreRepositoryIntegrationTestSuite) TestCreateGenre() {
	genre := domain.Genre{
		Name:        "Test Genre",
		Description: "Test Description",
	}

	result, utilsErr := s.repo.Create(s.ctx, genre)

	s.assertNoError(utilsErr)
	assert.NotEmpty(s.T(), result.ID)
	assert.Equal(s.T(), "Test Genre", result.Name)
	assert.Equal(s.T(), "Test Description", result.Description)

	s.testGenres = append(s.testGenres, result.ID)
}

func (s *GenreRepositoryIntegrationTestSuite) TestGetGenreByID() {

	baseGenre := s.baseGenres[0]

	retrieved, utilsErr := s.repo.GetByID(s.ctx, baseGenre.ID)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), baseGenre.ID, retrieved.ID)
	assert.Equal(s.T(), baseGenre.Name, retrieved.Name)
	assert.Equal(s.T(), baseGenre.Description, retrieved.Description)
}

func (s *GenreRepositoryIntegrationTestSuite) TestGetGenreByID_NotFound() {
	_, utilsErr := s.repo.GetByID(s.ctx, "00000000-0000-0000-0000-000000000000")

	s.assertErrorCode(utilsErr, 404)
}

func (s *GenreRepositoryIntegrationTestSuite) TestUpdateGenre() {

	genre := domain.Genre{
		Name:        "Old Name",
		Description: "Old Description",
	}

	created, utilsErr := s.repo.Create(s.ctx, genre)
	s.assertNoError(utilsErr)
	s.testGenres = append(s.testGenres, created.ID)

	updateData := domain.Genre{
		Name:        "Updated Name",
		Description: "Updated Description",
	}

	updated, utilsErr := s.repo.Update(s.ctx, created.ID, updateData)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), created.ID, updated.ID)
	assert.Equal(s.T(), "Updated Name", updated.Name)
	assert.Equal(s.T(), "Updated Description", updated.Description)
}

func (s *GenreRepositoryIntegrationTestSuite) TestUpdateGenre_NotFound() {
	updateData := domain.Genre{
		Name:        "Updated Name",
		Description: "Updated Description",
	}

	_, utilsErr := s.repo.Update(s.ctx, "00000000-0000-0000-0000-000000000000", updateData)

	s.assertErrorCode(utilsErr, 404)
}

func (s *GenreRepositoryIntegrationTestSuite) TestDeleteGenre() {

	genre := domain.Genre{
		Name:        "Delete Test",
		Description: "Delete Description",
	}

	created, utilsErr := s.repo.Create(s.ctx, genre)
	s.assertNoError(utilsErr)

	utilsErr = s.repo.Delete(s.ctx, created.ID)
	s.assertNoError(utilsErr)

	_, utilsErr = s.repo.GetByID(s.ctx, created.ID)
	s.assertErrorCode(utilsErr, 404)
}

func (s *GenreRepositoryIntegrationTestSuite) TestDeleteGenre_NotFound() {
	utilsErr := s.repo.Delete(s.ctx, "00000000-0000-0000-0000-000000000000")
	s.assertErrorCode(utilsErr, 404)
}

func (s *GenreRepositoryIntegrationTestSuite) TestGetAllGenres() {
	genres, total, utilsErr := s.repo.GetAll(s.ctx, domain.GenreFilters{}, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, len(s.baseGenres))
	assert.GreaterOrEqual(s.T(), len(genres), len(s.baseGenres))

	foundGenres := make(map[string]bool)
	for _, genre := range genres {
		foundGenres[genre.Name] = true
	}

	for _, baseGenre := range s.baseGenres {
		assert.True(s.T(), foundGenres[baseGenre.Name], "Base genre %s should be present", baseGenre.Name)
	}
}

func (s *GenreRepositoryIntegrationTestSuite) TestGetAllGenresWithPagination() {

	genres, total, utilsErr := s.repo.GetAll(s.ctx, domain.GenreFilters{}, 1, 2)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, len(s.baseGenres))
	assert.Equal(s.T(), 2, len(genres))

	genresPage2, totalPage2, utilsErr := s.repo.GetAll(s.ctx, domain.GenreFilters{}, 2, 2)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), total, totalPage2)
	assert.GreaterOrEqual(s.T(), len(genresPage2), 2)
}

func (s *GenreRepositoryIntegrationTestSuite) TestGetAllGenresWithNameFilter() {
	filters := domain.GenreFilters{
		Name: "Act",
	}

	genres, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)
	assert.GreaterOrEqual(s.T(), len(genres), 1)

	for _, genre := range genres {
		assert.Contains(s.T(), genre.Name, "Act")
	}
}

func (s *GenreRepositoryIntegrationTestSuite) TestGetAllGenresWithDescriptionFilter() {
	filters := domain.GenreFilters{
		Description: "fiction",
	}

	genres, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)
	assert.GreaterOrEqual(s.T(), len(genres), 1)

	for _, genre := range genres {
		assert.Contains(s.T(), genre.Description, "fiction")
	}
}

func (s *GenreRepositoryIntegrationTestSuite) TestGetAllGenresWithCombinedFilters() {
	filters := domain.GenreFilters{
		Name:        "Com",
		Description: "movies",
	}

	genres, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)
	assert.GreaterOrEqual(s.T(), len(genres), 1)

	foundComedy := false
	for _, genre := range genres {
		if genre.Name == "Comedy" {
			foundComedy = true
			break
		}
	}
	assert.True(s.T(), foundComedy, "Should find Comedy genre with combined filters")
}

func (s *GenreRepositoryIntegrationTestSuite) TestCreateDuplicateGenre() {

	duplicateGenre := domain.Genre{
		Name:        "Action",
		Description: "Another action genre",
	}

	_, utilsErr := s.repo.Create(s.ctx, duplicateGenre)

	s.assertErrorCode(utilsErr, 409)
}

func (s *GenreRepositoryIntegrationTestSuite) TestCreateGenreWithEmptyName() {
	genre := domain.Genre{
		Name:        "",
		Description: "Description",
	}

	_, utilsErr := s.repo.Create(s.ctx, genre)

	s.assertErrorCode(utilsErr, 400)
}

func (s *GenreRepositoryIntegrationTestSuite) TestUpdateGenreToDuplicateName() {

	genre := domain.Genre{
		Name:        "Unique Genre",
		Description: "Unique Description",
	}

	created, utilsErr := s.repo.Create(s.ctx, genre)
	s.assertNoError(utilsErr)
	s.testGenres = append(s.testGenres, created.ID)

	updateData := domain.Genre{
		Name:        "Action",
		Description: "Updated Description",
	}

	_, utilsErr = s.repo.Update(s.ctx, created.ID, updateData)

	s.assertErrorCode(utilsErr, 409)
}

func (s *GenreRepositoryIntegrationTestSuite) TestGetAllGenresEmptyResult() {
	filters := domain.GenreFilters{
		Name: "NonExistentGenreName12345",
	}

	genres, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), 0, total)
	assert.Equal(s.T(), 0, len(genres))
}
