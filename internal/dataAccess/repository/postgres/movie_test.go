package postgres

import (
	"context"
	"testing"
	"time"

	"cw/internal/dataAccess/config"
	domain "cw/internal/domain/models"
	"cw/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MovieRepositoryIntegrationTestSuite struct {
	suite.Suite
	repo       *MovieRepository
	genreRepo  *GenreRepository
	ctx        context.Context
	testMovies []string
	testGenres []string
}

func TestMovieRepositoryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(MovieRepositoryIntegrationTestSuite))
}

func (s *MovieRepositoryIntegrationTestSuite) SetupSuite() {
	db, err := config.NewTestDatabase()
	if err != nil {
		s.T().Fatalf("Failed to connect to test database: %v", err)
	}
	s.repo = NewMovieRepository(db)
	s.genreRepo = NewGenreRepository(db)
	s.ctx = context.Background()

	s.cleanDatabase()
	s.createTestGenres()
}

func (s *MovieRepositoryIntegrationTestSuite) createTestGenres() {
	genres := []domain.Genre{
		{Name: "Action", Description: "Action movies"},
		{Name: "Drama", Description: "Drama movies"},
		{Name: "Comedy", Description: "Comedy movies"},
	}

	for _, genre := range genres {
		created, err := s.genreRepo.Create(s.ctx, genre)
		if err != nil {
			s.T().Fatalf("Failed to create test genre: %v", err)
		}
		s.testGenres = append(s.testGenres, created.ID)
	}
}

func (s *MovieRepositoryIntegrationTestSuite) cleanDatabase() {

	_, err := s.repo.db.Exec(s.ctx, "DELETE FROM movies_genres")
	if err != nil {
		s.T().Fatalf("Failed to clean movies_genres: %v", err)
	}
	_, err = s.repo.db.Exec(s.ctx, "DELETE FROM reviews")
	if err != nil {
		s.T().Fatalf("Failed to clean reviews: %v", err)
	}

	_, err = s.repo.db.Exec(s.ctx, "DELETE FROM movies")
	if err != nil {
		s.T().Fatalf("Failed to clean movies: %v", err)
	}
	_, err = s.repo.db.Exec(s.ctx, "DELETE FROM genres")
	if err != nil {
		s.T().Fatalf("Failed to clean genres: %v", err)
	}
}

func (s *MovieRepositoryIntegrationTestSuite) SetupTest() {
	s.testMovies = []string{}
}

func (s *MovieRepositoryIntegrationTestSuite) TearDownTest() {
	for _, id := range s.testMovies {
		s.repo.Delete(s.ctx, id)
	}
}

func (s *MovieRepositoryIntegrationTestSuite) TearDownSuite() {
	for _, id := range s.testGenres {
		s.genreRepo.Delete(s.ctx, id)
	}
}

func (s *MovieRepositoryIntegrationTestSuite) assertNoError(utilsErr *utils.Error) {
	if utilsErr != nil {
		s.T().Errorf("Expected no error, got: %v", utilsErr)
	}
}

func (s *MovieRepositoryIntegrationTestSuite) assertErrorCode(utilsErr *utils.Error, expectedCode int) {
	if utilsErr == nil {
		s.T().Error("Expected error, got nil")
		return
	}
	assert.Equal(s.T(), expectedCode, utilsErr.Code)
}

func (s *MovieRepositoryIntegrationTestSuite) TestCreateMovie() {
	movie := domain.Movie{
		Title:            "Test Movie",
		Duration:         "02:00:00",
		Description:      "Test Description",
		AgeLimit:         16,
		BoxOfficeRevenue: 1000000.0,
		ReleaseDate:      time.Now(),
		GenreIDs:         []string{s.testGenres[0]},
	}

	result, utilsErr := s.repo.Create(s.ctx, movie)

	s.assertNoError(utilsErr)
	assert.NotEmpty(s.T(), result.ID)
	assert.Equal(s.T(), "Test Movie", result.Title)
	assert.Equal(s.T(), "02:00:00", result.Duration)
	assert.Equal(s.T(), "Test Description", result.Description)
	assert.Equal(s.T(), 16, result.AgeLimit)
	assert.Equal(s.T(), 1000000.0, result.BoxOfficeRevenue)
	assert.Len(s.T(), result.GenreIDs, 1)
	assert.Equal(s.T(), s.testGenres[0], result.GenreIDs[0])

	s.testMovies = append(s.testMovies, result.ID)
}

func (s *MovieRepositoryIntegrationTestSuite) TestCreateMovieWithoutGenres() {
	movie := domain.Movie{
		Title:            "Test Movie No Genres",
		Duration:         "01:30:00",
		Description:      "Test Description No Genres",
		AgeLimit:         12,
		BoxOfficeRevenue: 500000.0,
		ReleaseDate:      time.Now(),
		GenreIDs:         []string{},
	}

	result, utilsErr := s.repo.Create(s.ctx, movie)

	s.assertNoError(utilsErr)
	assert.NotEmpty(s.T(), result.ID)
	assert.Equal(s.T(), "Test Movie No Genres", result.Title)
	assert.Empty(s.T(), result.GenreIDs)

	s.testMovies = append(s.testMovies, result.ID)
}

func (s *MovieRepositoryIntegrationTestSuite) TestGetMovieByID() {
	movie := domain.Movie{
		Title:            "Test Get Movie",
		Duration:         "01:50:00",
		Description:      "Test Get Description",
		AgeLimit:         18,
		BoxOfficeRevenue: 2000000.0,
		ReleaseDate:      time.Now(),
		GenreIDs:         []string{s.testGenres[0], s.testGenres[1]},
	}

	created, utilsErr := s.repo.Create(s.ctx, movie)
	s.assertNoError(utilsErr)
	s.testMovies = append(s.testMovies, created.ID)

	retrieved, utilsErr := s.repo.GetByID(s.ctx, created.ID)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), created.ID, retrieved.ID)
	assert.Equal(s.T(), "Test Get Movie", retrieved.Title)
	assert.Equal(s.T(), "01:50:00", retrieved.Duration)
	assert.Equal(s.T(), "Test Get Description", retrieved.Description)
	assert.Equal(s.T(), 18, retrieved.AgeLimit)
	assert.Equal(s.T(), 2000000.0, retrieved.BoxOfficeRevenue)
	assert.Len(s.T(), retrieved.GenreIDs, 2)
}

func (s *MovieRepositoryIntegrationTestSuite) TestGetMovieByID_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	_, utilsErr := s.repo.GetByID(s.ctx, nonExistentID)

	s.assertErrorCode(utilsErr, 404)
}

func (s *MovieRepositoryIntegrationTestSuite) TestUpdateMovie() {
	movie := domain.Movie{
		Title:            "Old Title",
		Duration:         "01:40:00",
		Description:      "Old Description",
		AgeLimit:         12,
		BoxOfficeRevenue: 500000.0,
		ReleaseDate:      time.Now(),
		GenreIDs:         []string{s.testGenres[0]},
	}

	created, utilsErr := s.repo.Create(s.ctx, movie)
	s.assertNoError(utilsErr)
	s.testMovies = append(s.testMovies, created.ID)

	updateData := domain.Movie{
		Title:            "Updated Title",
		Duration:         "02:30:00",
		Description:      "Updated Description",
		AgeLimit:         16,
		BoxOfficeRevenue: 1500000.0,
		ReleaseDate:      time.Now().AddDate(0, 0, 1),
		GenreIDs:         []string{s.testGenres[1], s.testGenres[2]},
	}

	updated, utilsErr := s.repo.Update(s.ctx, created.ID, updateData)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), created.ID, updated.ID)
	assert.Equal(s.T(), "Updated Title", updated.Title)
	assert.Equal(s.T(), "02:30:00", updated.Duration)
	assert.Equal(s.T(), "Updated Description", updated.Description)
	assert.Equal(s.T(), 16, updated.AgeLimit)
	assert.Equal(s.T(), 1500000.0, updated.BoxOfficeRevenue)
	assert.Len(s.T(), updated.GenreIDs, 2)
}

func (s *MovieRepositoryIntegrationTestSuite) TestUpdateMovie_RemoveGenres() {
	movie := domain.Movie{
		Title:            "Movie With Genres",
		Duration:         "02:00:00",
		Description:      "Movie with genres",
		AgeLimit:         12,
		BoxOfficeRevenue: 1000000.0,
		ReleaseDate:      time.Now(),
		GenreIDs:         []string{s.testGenres[0], s.testGenres[1]},
	}

	created, utilsErr := s.repo.Create(s.ctx, movie)
	s.assertNoError(utilsErr)
	s.testMovies = append(s.testMovies, created.ID)

	updateData := domain.Movie{
		Title:            "Movie Without Genres",
		Duration:         "02:00:00",
		Description:      "Movie without genres",
		AgeLimit:         12,
		BoxOfficeRevenue: 1000000.0,
		ReleaseDate:      time.Now(),
		GenreIDs:         []string{},
	}

	updated, utilsErr := s.repo.Update(s.ctx, created.ID, updateData)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), "Movie Without Genres", updated.Title)
	assert.Empty(s.T(), updated.GenreIDs)
}

func (s *MovieRepositoryIntegrationTestSuite) TestUpdateMovie_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	updateData := domain.Movie{
		Title:       "Non-existent",
		Duration:    "02:00:00",
		Description: "Should not update",
	}

	_, utilsErr := s.repo.Update(s.ctx, nonExistentID, updateData)

	s.assertErrorCode(utilsErr, 404)
}

func (s *MovieRepositoryIntegrationTestSuite) TestDeleteMovie() {
	movie := domain.Movie{
		Title:            "Delete Test",
		Duration:         "01:30:00",
		Description:      "Delete Description",
		AgeLimit:         12,
		BoxOfficeRevenue: 500000.0,
		ReleaseDate:      time.Now(),
		GenreIDs:         []string{s.testGenres[0]},
	}

	created, utilsErr := s.repo.Create(s.ctx, movie)
	s.assertNoError(utilsErr)

	utilsErr = s.repo.Delete(s.ctx, created.ID)
	s.assertNoError(utilsErr)

	_, utilsErr = s.repo.GetByID(s.ctx, created.ID)
	s.assertErrorCode(utilsErr, 404)
}

func (s *MovieRepositoryIntegrationTestSuite) TestDeleteMovie_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	utilsErr := s.repo.Delete(s.ctx, nonExistentID)

	s.assertErrorCode(utilsErr, 404)
}

func (s *MovieRepositoryIntegrationTestSuite) TestGetAllMovies() {
	movie1 := domain.Movie{
		Title:            "Movie A",
		Duration:         "01:40:00",
		Description:      "Description A",
		AgeLimit:         12,
		BoxOfficeRevenue: 1000000.0,
		ReleaseDate:      time.Now(),
		GenreIDs:         []string{s.testGenres[0]},
	}
	movie2 := domain.Movie{
		Title:            "Movie B",
		Duration:         "02:00:00",
		Description:      "Description B",
		AgeLimit:         16,
		BoxOfficeRevenue: 2000000.0,
		ReleaseDate:      time.Now(),
		GenreIDs:         []string{s.testGenres[1]},
	}

	created1, utilsErr := s.repo.Create(s.ctx, movie1)
	s.assertNoError(utilsErr)
	created2, utilsErr := s.repo.Create(s.ctx, movie2)
	s.assertNoError(utilsErr)

	s.testMovies = append(s.testMovies, created1.ID, created2.ID)

	movies, total, utilsErr := s.repo.GetAll(s.ctx, domain.MovieFilters{}, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 2)
	assert.GreaterOrEqual(s.T(), len(movies), 2)
}

func (s *MovieRepositoryIntegrationTestSuite) TestGetAllMoviesWithTitleFilter() {
	movie := domain.Movie{
		Title:            "Unique Filter Movie",
		Duration:         "01:50:00",
		Description:      "Unique Description",
		AgeLimit:         12,
		BoxOfficeRevenue: 1000000.0,
		ReleaseDate:      time.Now(),
		GenreIDs:         []string{s.testGenres[0]},
	}

	created, utilsErr := s.repo.Create(s.ctx, movie)
	s.assertNoError(utilsErr)
	s.testMovies = append(s.testMovies, created.ID)

	filters := domain.MovieFilters{
		Title: "Unique Filter",
	}

	movies, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)
	assert.GreaterOrEqual(s.T(), len(movies), 1)
}

func (s *MovieRepositoryIntegrationTestSuite) TestGetAllMoviesWithGenreFilter() {
	movie := domain.Movie{
		Title:            "Genre Filter Movie",
		Duration:         "02:00:00",
		Description:      "Genre Description",
		AgeLimit:         12,
		BoxOfficeRevenue: 1000000.0,
		ReleaseDate:      time.Now(),
		GenreIDs:         []string{s.testGenres[0]},
	}

	created, utilsErr := s.repo.Create(s.ctx, movie)
	s.assertNoError(utilsErr)
	s.testMovies = append(s.testMovies, created.ID)

	filters := domain.MovieFilters{
		Genre: "Action",
	}

	movies, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)
	assert.GreaterOrEqual(s.T(), len(movies), 1)
}

func (s *MovieRepositoryIntegrationTestSuite) TestCalculateMovieRating() {
	movie := domain.Movie{
		Title:            "Rating Test Movie",
		Duration:         "01:40:00",
		Description:      "Rating Test Description",
		AgeLimit:         12,
		BoxOfficeRevenue: 1000000.0,
		ReleaseDate:      time.Now(),
		GenreIDs:         []string{s.testGenres[0]},
	}

	created, utilsErr := s.repo.Create(s.ctx, movie)
	s.assertNoError(utilsErr)
	s.testMovies = append(s.testMovies, created.ID)

	rating, utilsErr := s.repo.CalculateMovieRating(s.ctx, created.ID)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), 0.0, rating)
}
