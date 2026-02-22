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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MovieShowRepositoryIntegrationTestSuite struct {
	suite.Suite
	repo            *MovieShowRepository
	movieRepo       *MovieRepository
	hallRepo        *HallRepository
	screenTypeRepo  *ScreenTypeRepository
	ctx             context.Context
	testShows       []string
	testMovies      []string
	testHalls       []string
	testScreenTypes []string
}

func TestMovieShowRepositoryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(MovieShowRepositoryIntegrationTestSuite))
}

func (s *MovieShowRepositoryIntegrationTestSuite) SetupSuite() {
	db, err := config.NewTestDatabase()
	if err != nil {
		s.T().Fatalf("Failed to connect to test database: %v", err)
	}
	s.repo = NewMovieShowRepository(db)
	s.movieRepo = NewMovieRepository(db)
	s.hallRepo = NewHallRepository(db)
	s.screenTypeRepo = NewScreenTypeRepository(db)
	s.ctx = context.Background()

	s.cleanDatabase()
	s.createTestScreenTypes()
	s.createTestMovies()
	s.createTestHalls()
}

func (s *MovieShowRepositoryIntegrationTestSuite) createTestScreenTypes() {
	screenTypes := []domain.ScreenType{
		{Name: "2D", Description: "Standard 2D", PriceModifier: 1.0},
		{Name: "3D", Description: "3D movies", PriceModifier: 1.5},
		{Name: "IMAX", Description: "IMAX format", PriceModifier: 2.0},
	}

	for _, st := range screenTypes {
		created, err := s.screenTypeRepo.Create(s.ctx, st)
		if err != nil {
			s.T().Fatalf("Failed to create test screen type: %v", err)
		}
		s.testScreenTypes = append(s.testScreenTypes, created.ID)
	}
}

func (s *MovieShowRepositoryIntegrationTestSuite) createTestMovies() {
	movies := []domain.Movie{
		{
			Title:            "Action Movie",
			Duration:         "02:00:00",
			Description:      "Exciting action movie",
			AgeLimit:         16,
			BoxOfficeRevenue: 1000000.0,
			ReleaseDate:      time.Now().AddDate(0, -1, 0),
			GenreIDs:         []string{},
		},
		{
			Title:            "Comedy Movie",
			Duration:         "01:30:00",
			Description:      "Funny comedy movie",
			AgeLimit:         12,
			BoxOfficeRevenue: 500000.0,
			ReleaseDate:      time.Now().AddDate(0, -2, 0),
			GenreIDs:         []string{},
		},
		{
			Title:            "Drama Movie",
			Duration:         "02:15:00",
			Description:      "Emotional drama movie",
			AgeLimit:         18,
			BoxOfficeRevenue: 750000.0,
			ReleaseDate:      time.Now().AddDate(0, -3, 0),
			GenreIDs:         []string{},
		},
	}

	for _, movie := range movies {
		created, err := s.movieRepo.Create(s.ctx, movie)
		if err != nil {
			s.T().Fatalf("Failed to create test movie: %v", err)
		}
		s.testMovies = append(s.testMovies, created.ID)
	}
}

func (s *MovieShowRepositoryIntegrationTestSuite) createTestHalls() {
	halls := []domain.Hall{
		{
			Name:         "Main Hall",
			Description:  "Large main hall with great acoustics",
			ScreenTypeID: s.testScreenTypes[2],
		},
		{
			Name:         "Comfort Hall",
			Description:  "Medium sized comfortable hall",
			ScreenTypeID: s.testScreenTypes[0],
		},
		{
			Name:         "3D Hall",
			Description:  "Hall specialized for 3D movies",
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

func (s *MovieShowRepositoryIntegrationTestSuite) cleanDatabase() {
	_, err := s.repo.db.Exec(s.ctx, "DELETE FROM tickets")
	if err != nil {
		s.T().Logf("Note: tickets table might not exist: %v", err)
	}

	_, err = s.repo.db.Exec(s.ctx, "DELETE FROM movie_shows")
	if err != nil {
		s.T().Fatalf("Failed to clean movie_shows: %v", err)
	}

	_, err = s.repo.db.Exec(s.ctx, "DELETE FROM halls")
	if err != nil {
		s.T().Fatalf("Failed to clean halls: %v", err)
	}

	_, err = s.repo.db.Exec(s.ctx, "DELETE FROM movies")
	if err != nil {
		s.T().Fatalf("Failed to clean movies: %v", err)
	}

	_, err = s.repo.db.Exec(s.ctx, "DELETE FROM screen_types")
	if err != nil {
		s.T().Logf("Note: screen_types table might not exist: %v", err)
	}
}

func (s *MovieShowRepositoryIntegrationTestSuite) SetupTest() {
	s.testShows = []string{}
}

func (s *MovieShowRepositoryIntegrationTestSuite) TearDownTest() {
	for _, id := range s.testShows {
		_ = s.repo.Delete(s.ctx, id)
	}
}

func (s *MovieShowRepositoryIntegrationTestSuite) TearDownSuite() {
	for _, id := range s.testHalls {
		_ = s.hallRepo.Delete(s.ctx, id)
	}
	for _, id := range s.testMovies {
		_ = s.movieRepo.Delete(s.ctx, id)
	}
	for _, id := range s.testScreenTypes {
		_ = s.screenTypeRepo.Delete(s.ctx, id)
	}
}

func (s *MovieShowRepositoryIntegrationTestSuite) assertNoError(utilsErr *utils.Error) {
	if utilsErr != nil {
		s.T().Errorf("Expected no error, got: %v", utilsErr)
	}
}

func (s *MovieShowRepositoryIntegrationTestSuite) assertErrorCode(utilsErr *utils.Error, expectedCode int) {
	if utilsErr == nil {
		s.T().Error("Expected error, got nil")
		return
	}
	assert.Equal(s.T(), expectedCode, utilsErr.Code)
}

func (s *MovieShowRepositoryIntegrationTestSuite) TestCreateMovieShow() {
	startTime := time.Now().Add(2 * time.Hour).Round(time.Minute).UTC()

	movieShow := domain.MovieShow{
		MovieID:   s.testMovies[0],
		HallID:    s.testHalls[0],
		StartTime: startTime,
		Language:  "English",
	}

	result, utilsErr := s.repo.Create(s.ctx, movieShow)

	s.assertNoError(utilsErr)
	assert.NotEmpty(s.T(), result.ID)
	assert.Equal(s.T(), s.testMovies[0], result.MovieID)
	assert.Equal(s.T(), s.testHalls[0], result.HallID)
	assert.WithinDuration(s.T(), startTime, result.StartTime, time.Minute)
	assert.Equal(s.T(), "English", result.Language)

	s.testShows = append(s.testShows, result.ID)
}

func (s *MovieShowRepositoryIntegrationTestSuite) TestCreateMovieShowWithDifferentLanguages() {
	languages := []string{"English", "Spanish", "French", "German", "Italian"}

	for i, lang := range languages {
		startTime := time.Now().Add(time.Duration(i+1) * 6 * time.Hour).UTC()
		movieShow := domain.MovieShow{
			MovieID:   s.testMovies[1],
			HallID:    s.testHalls[i%len(s.testHalls)],
			StartTime: startTime,
			Language:  lang,
		}

		result, utilsErr := s.repo.Create(s.ctx, movieShow)
		s.assertNoError(utilsErr)
		assert.Equal(s.T(), lang, result.Language)
		s.testShows = append(s.testShows, result.ID)
	}
}

func (s *MovieShowRepositoryIntegrationTestSuite) TestGetMovieShowByID() {
	startTime := time.Now().Add(3 * time.Hour).Round(time.Minute).UTC()
	movieShow := domain.MovieShow{
		MovieID:   s.testMovies[0],
		HallID:    s.testHalls[0],
		StartTime: startTime,
		Language:  "English",
	}

	created, utilsErr := s.repo.Create(s.ctx, movieShow)
	s.assertNoError(utilsErr)
	s.testShows = append(s.testShows, created.ID)

	retrieved, utilsErr := s.repo.GetByID(s.ctx, created.ID)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), created.ID, retrieved.ID)
	assert.Equal(s.T(), s.testMovies[0], retrieved.MovieID)
	assert.Equal(s.T(), s.testHalls[0], retrieved.HallID)
	assert.WithinDuration(s.T(), startTime, retrieved.StartTime, time.Minute)
	assert.Equal(s.T(), "English", retrieved.Language)
}

func (s *MovieShowRepositoryIntegrationTestSuite) TestGetMovieShowByID_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	_, utilsErr := s.repo.GetByID(s.ctx, nonExistentID)

	s.assertErrorCode(utilsErr, 404)
}

func (s *MovieShowRepositoryIntegrationTestSuite) TestUpdateMovieShow() {
	originalStartTime := time.Now().Add(4 * time.Hour).Round(time.Minute).UTC()
	movieShow := domain.MovieShow{
		MovieID:   s.testMovies[0],
		HallID:    s.testHalls[0],
		StartTime: originalStartTime,
		Language:  "English",
	}

	created, utilsErr := s.repo.Create(s.ctx, movieShow)
	s.assertNoError(utilsErr)
	s.testShows = append(s.testShows, created.ID)

	newStartTime := time.Now().Add(8 * time.Hour).Round(time.Minute).UTC()
	updateData := domain.MovieShow{
		MovieID:   s.testMovies[1],
		HallID:    s.testHalls[1],
		StartTime: newStartTime,
		Language:  "French",
	}

	updated, utilsErr := s.repo.Update(s.ctx, created.ID, updateData)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), created.ID, updated.ID)
	assert.Equal(s.T(), s.testMovies[1], updated.MovieID)
	assert.Equal(s.T(), s.testHalls[1], updated.HallID)
	assert.WithinDuration(s.T(), newStartTime, updated.StartTime, time.Minute)
	assert.Equal(s.T(), "French", updated.Language)
}

func (s *MovieShowRepositoryIntegrationTestSuite) TestUpdateMovieShow_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	updateData := domain.MovieShow{
		MovieID:   s.testMovies[0],
		HallID:    s.testHalls[0],
		StartTime: time.Now().UTC(),
		Language:  "English",
	}

	_, utilsErr := s.repo.Update(s.ctx, nonExistentID, updateData)

	s.assertErrorCode(utilsErr, 404)
}

func (s *MovieShowRepositoryIntegrationTestSuite) TestDeleteMovieShow() {
	movieShow := domain.MovieShow{
		MovieID:   s.testMovies[0],
		HallID:    s.testHalls[0],
		StartTime: time.Now().Add(12 * time.Hour).UTC(),
		Language:  "German",
	}

	created, utilsErr := s.repo.Create(s.ctx, movieShow)
	s.assertNoError(utilsErr)

	utilsErr = s.repo.Delete(s.ctx, created.ID)
	s.assertNoError(utilsErr)

	_, utilsErr = s.repo.GetByID(s.ctx, created.ID)
	s.assertErrorCode(utilsErr, 404)
}

func (s *MovieShowRepositoryIntegrationTestSuite) TestDeleteMovieShow_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	utilsErr := s.repo.Delete(s.ctx, nonExistentID)

	s.assertErrorCode(utilsErr, 404)
}

func (s *MovieShowRepositoryIntegrationTestSuite) TestGetAllMovieShows() {
	movieShows := []domain.MovieShow{
		{
			MovieID:   s.testMovies[0],
			HallID:    s.testHalls[0],
			StartTime: time.Now().Add(16 * time.Hour).UTC(),
			Language:  "English",
		},
		{
			MovieID:   s.testMovies[1],
			HallID:    s.testHalls[1],
			StartTime: time.Now().Add(20 * time.Hour).UTC(),
			Language:  "Spanish",
		},
		{
			MovieID:   s.testMovies[2],
			HallID:    s.testHalls[2],
			StartTime: time.Now().Add(24 * time.Hour).UTC(),
			Language:  "French",
		},
	}

	for _, show := range movieShows {
		created, utilsErr := s.repo.Create(s.ctx, show)
		s.assertNoError(utilsErr)
		s.testShows = append(s.testShows, created.ID)
	}

	shows, total, utilsErr := s.repo.GetAll(s.ctx, domain.MovieShowFilters{}, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, len(movieShows))
	assert.GreaterOrEqual(s.T(), len(shows), len(movieShows))
}

func (s *MovieShowRepositoryIntegrationTestSuite) TestGetAllMovieShowsWithPagination() {
	for i := 0; i < 5; i++ {
		movieShow := domain.MovieShow{
			MovieID:   s.testMovies[i%len(s.testMovies)],
			HallID:    s.testHalls[i%len(s.testHalls)],
			StartTime: time.Now().Add(time.Duration(i+1) * 8 * time.Hour).UTC(),
			Language:  "English",
		}
		created, utilsErr := s.repo.Create(s.ctx, movieShow)
		s.assertNoError(utilsErr)
		s.testShows = append(s.testShows, created.ID)
	}

	showsPage1, total, utilsErr := s.repo.GetAll(s.ctx, domain.MovieShowFilters{}, 1, 2)
	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 5)
	assert.Len(s.T(), showsPage1, 2)

	showsPage2, _, utilsErr := s.repo.GetAll(s.ctx, domain.MovieShowFilters{}, 2, 2)
	s.assertNoError(utilsErr)
	assert.Len(s.T(), showsPage2, 2)
}

func (s *MovieShowRepositoryIntegrationTestSuite) TestGetAllMovieShowsWithMovieFilter() {
	movieShow := domain.MovieShow{
		MovieID:   s.testMovies[0],
		HallID:    s.testHalls[0],
		StartTime: time.Now().Add(30 * time.Hour).UTC(),
		Language:  "English",
	}

	created, utilsErr := s.repo.Create(s.ctx, movieShow)
	s.assertNoError(utilsErr)
	s.testShows = append(s.testShows, created.ID)

	filters := domain.MovieShowFilters{
		MovieIDs: []string{s.testMovies[0]},
	}

	shows, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)
	assert.GreaterOrEqual(s.T(), len(shows), 1)

	for _, show := range shows {
		assert.Equal(s.T(), s.testMovies[0], show.MovieID)
	}
}

func (s *MovieShowRepositoryIntegrationTestSuite) TestGetAllMovieShowsWithMultipleMovieFilter() {
	for i, movieID := range s.testMovies[:2] {
		movieShow := domain.MovieShow{
			MovieID:   movieID,
			HallID:    s.testHalls[i],
			StartTime: time.Now().Add(time.Duration(32+i) * 10 * time.Hour).UTC(),
			Language:  "English",
		}
		created, utilsErr := s.repo.Create(s.ctx, movieShow)
		s.assertNoError(utilsErr)
		s.testShows = append(s.testShows, created.ID)
	}

	filters := domain.MovieShowFilters{
		MovieIDs: []string{s.testMovies[0], s.testMovies[1]},
	}

	shows, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 2)

	for _, show := range shows {
		assert.True(s.T(), show.MovieID == s.testMovies[0] || show.MovieID == s.testMovies[1])
	}
}

func (s *MovieShowRepositoryIntegrationTestSuite) TestGetAllMovieShowsWithHallFilter() {
	movieShow := domain.MovieShow{
		MovieID:   s.testMovies[0],
		HallID:    s.testHalls[0],
		StartTime: time.Now().Add(36 * time.Hour).UTC(),
		Language:  "English",
	}

	created, utilsErr := s.repo.Create(s.ctx, movieShow)
	s.assertNoError(utilsErr)
	s.testShows = append(s.testShows, created.ID)

	filters := domain.MovieShowFilters{
		HallIDs: []string{s.testHalls[0]},
	}

	shows, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)
	assert.GreaterOrEqual(s.T(), len(shows), 1)

	for _, show := range shows {
		assert.Equal(s.T(), s.testHalls[0], show.HallID)
	}
}

func (s *MovieShowRepositoryIntegrationTestSuite) TestGetAllMovieShowsWithLanguageFilter() {
	movieShow := domain.MovieShow{
		MovieID:   s.testMovies[0],
		HallID:    s.testHalls[0],
		StartTime: time.Now().Add(40 * time.Hour).UTC(),
		Language:  "French",
	}

	created, utilsErr := s.repo.Create(s.ctx, movieShow)
	s.assertNoError(utilsErr)
	s.testShows = append(s.testShows, created.ID)

	filters := domain.MovieShowFilters{
		Languages: []string{"French"},
	}

	shows, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)
	assert.GreaterOrEqual(s.T(), len(shows), 1)

	for _, show := range shows {
		assert.Equal(s.T(), "French", show.Language)
	}
}

func (s *MovieShowRepositoryIntegrationTestSuite) TestGetAllMovieShowsWithMultipleLanguageFilter() {
	languages := []string{"English", "French"}

	for i, lang := range languages {
		movieShow := domain.MovieShow{
			MovieID:   s.testMovies[0],
			HallID:    s.testHalls[i],
			StartTime: time.Now().Add(time.Duration(44+i) * 12 * time.Hour).UTC(),
			Language:  lang,
		}
		created, utilsErr := s.repo.Create(s.ctx, movieShow)
		s.assertNoError(utilsErr)
		s.testShows = append(s.testShows, created.ID)
	}

	filters := domain.MovieShowFilters{
		Languages: languages,
	}

	shows, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, len(languages))

	for _, show := range shows {
		assert.Contains(s.T(), languages, show.Language)
	}
}

func (s *MovieShowRepositoryIntegrationTestSuite) TestGetAllMovieShowsWithDateFilter() {
	tomorrow := time.Now().Add(24 * time.Hour)
	specificTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 14, 30, 0, 0, time.UTC)

	movieShow := domain.MovieShow{
		MovieID:   s.testMovies[0],
		HallID:    s.testHalls[0],
		StartTime: specificTime,
		Language:  "English",
	}

	created, utilsErr := s.repo.Create(s.ctx, movieShow)
	s.assertNoError(utilsErr)
	s.testShows = append(s.testShows, created.ID)

	filters := domain.MovieShowFilters{
		Date: tomorrow.Format("2006-01-02"),
	}

	shows, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)
	assert.GreaterOrEqual(s.T(), len(shows), 1)

	for _, show := range shows {
		showDate := show.StartTime.Format("2006-01-02")
		assert.Equal(s.T(), tomorrow.Format("2006-01-02"), showDate)
	}
}

func (s *MovieShowRepositoryIntegrationTestSuite) TestGetAllMovieShowsWithComplexFilters() {
	tomorrow := time.Now().Add(24 * time.Hour)
	specificTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 16, 0, 0, 0, time.UTC)

	movieShow := domain.MovieShow{
		MovieID:   s.testMovies[0],
		HallID:    s.testHalls[0],
		StartTime: specificTime,
		Language:  "English",
	}

	created, utilsErr := s.repo.Create(s.ctx, movieShow)
	s.assertNoError(utilsErr)
	s.testShows = append(s.testShows, created.ID)

	filters := domain.MovieShowFilters{
		MovieIDs:      []string{s.testMovies[0]},
		HallIDs:       []string{s.testHalls[0]},
		Languages:     []string{"English"},
		Date:          tomorrow.Format("2006-01-02"),
		StartTimeFrom: "15:00",
		StartTimeTo:   "17:00",
	}

	shows, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)
	assert.GreaterOrEqual(s.T(), len(shows), 1)

	if len(shows) > 0 {
		show := shows[0]
		assert.Equal(s.T(), s.testMovies[0], show.MovieID)
		assert.Equal(s.T(), s.testHalls[0], show.HallID)
		assert.Equal(s.T(), "English", show.Language)
		showDate := show.StartTime.Format("2006-01-02")
		assert.Equal(s.T(), tomorrow.Format("2006-01-02"), showDate)
	}
}
