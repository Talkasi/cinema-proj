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

type ReviewRepositoryIntegrationTestSuite struct {
	suite.Suite
	repo        *ReviewRepository
	movieRepo   *MovieRepository
	userRepo    *UserRepository
	ctx         context.Context
	testReviews []string
	testMovies  []string
	testUsers   []string
}

func TestReviewRepositoryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ReviewRepositoryIntegrationTestSuite))
}

func (s *ReviewRepositoryIntegrationTestSuite) SetupSuite() {
	db, err := config.NewTestDatabase()
	if err != nil {
		s.T().Fatalf("Failed to connect to test database: %v", err)
	}
	s.repo = NewReviewRepository(db)
	s.movieRepo = NewMovieRepository(db)

	jwtSecret := "test-secret-key"
	tokenDuration := 24 * time.Hour
	s.userRepo = NewUserRepository(db, jwtSecret, tokenDuration)

	s.ctx = context.Background()

	s.cleanDatabase()
	s.createTestMovies()
	s.createTestUsers()
}

func (s *ReviewRepositoryIntegrationTestSuite) createTestMovies() {

	movies := []domain.Movie{
		{
			Title:            "Test Action Movie " + time.Now().Format("150405"),
			Duration:         "02:00:00",
			Description:      "Exciting action movie for testing",
			AgeLimit:         16,
			BoxOfficeRevenue: 1000000.0,
			ReleaseDate:      time.Now().AddDate(0, -1, 0),
			GenreIDs:         []string{},
		},
		{
			Title:            "Test Comedy Movie " + time.Now().Format("150405"),
			Duration:         "01:30:00",
			Description:      "Funny comedy movie for testing",
			AgeLimit:         12,
			BoxOfficeRevenue: 500000.0,
			ReleaseDate:      time.Now().AddDate(0, -2, 0),
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

func (s *ReviewRepositoryIntegrationTestSuite) createTestUsers() {

	timestamp := time.Now().Format("150405")
	users := []domain.User{
		{
			Name:         "Test User 1",
			Email:        "test1_" + timestamp + "@example.com",
			PasswordHash: "password123",
			BirthDate:    time.Now().AddDate(-25, 0, 0).Format("2006-01-02"),
			IsAdmin:      false,
		},
		{
			Name:         "Test User 2",
			Email:        "test2_" + timestamp + "@example.com",
			PasswordHash: "password456",
			BirthDate:    time.Now().AddDate(-30, 0, 0).Format("2006-01-02"),
			IsAdmin:      false,
		},
	}

	for _, user := range users {
		created, err := s.userRepo.Register(s.ctx, user)
		if err != nil {
			s.T().Fatalf("Failed to create test user: %v", err)
		}
		s.testUsers = append(s.testUsers, created.ID)
	}
}

func (s *ReviewRepositoryIntegrationTestSuite) cleanDatabase() {

	for _, id := range s.testReviews {
		_ = s.repo.Delete(s.ctx, id)
	}
	for _, id := range s.testMovies {
		_ = s.movieRepo.Delete(s.ctx, id)
	}
	for _, id := range s.testUsers {
		_ = s.userRepo.Delete(s.ctx, id)
	}
}

func (s *ReviewRepositoryIntegrationTestSuite) SetupTest() {
	s.testReviews = []string{}
}

func (s *ReviewRepositoryIntegrationTestSuite) TearDownTest() {
	for _, id := range s.testReviews {
		_ = s.repo.Delete(s.ctx, id)
	}
}

func (s *ReviewRepositoryIntegrationTestSuite) TearDownSuite() {
	s.cleanDatabase()
}

func (s *ReviewRepositoryIntegrationTestSuite) assertNoError(utilsErr *utils.Error) {
	if utilsErr != nil {
		s.T().Errorf("Expected no error, got: %v", utilsErr)
	}
}

func (s *ReviewRepositoryIntegrationTestSuite) assertErrorCode(utilsErr *utils.Error, expectedCode int) {
	if utilsErr == nil {
		s.T().Error("Expected error, got nil")
		return
	}
	assert.Equal(s.T(), expectedCode, utilsErr.Code)
}

func (s *ReviewRepositoryIntegrationTestSuite) TestCreateForMovie() {
	review := domain.Review{
		UserID:  s.testUsers[0],
		Rating:  8,
		Comment: "Great movie with amazing action scenes!",
	}

	result, utilsErr := s.repo.CreateForMovie(s.ctx, s.testMovies[0], review)

	s.assertNoError(utilsErr)
	assert.NotEmpty(s.T(), result.ID)
	assert.Equal(s.T(), s.testMovies[0], result.MovieID)
	assert.Equal(s.T(), s.testUsers[0], result.UserID)
	assert.Equal(s.T(), 8, result.Rating)
	assert.Equal(s.T(), "Great movie with amazing action scenes!", result.Comment)

	s.testReviews = append(s.testReviews, result.ID)
}

func (s *ReviewRepositoryIntegrationTestSuite) TestCreateForMovieWithEmptyComment() {
	review := domain.Review{
		UserID:  s.testUsers[1],
		Rating:  7,
		Comment: "Good",
	}

	result, utilsErr := s.repo.CreateForMovie(s.ctx, s.testMovies[0], review)

	s.assertNoError(utilsErr)
	assert.NotEmpty(s.T(), result.ID)
	assert.Equal(s.T(), s.testMovies[0], result.MovieID)
	assert.Equal(s.T(), s.testUsers[1], result.UserID)
	assert.Equal(s.T(), 7, result.Rating)
	assert.Equal(s.T(), "Good", result.Comment)

	s.testReviews = append(s.testReviews, result.ID)
}

func (s *ReviewRepositoryIntegrationTestSuite) TestCreateForMovieDuplicateUserReview() {
	review := domain.Review{
		UserID:  s.testUsers[0],
		Rating:  8,
		Comment: "First review",
	}

	result1, utilsErr := s.repo.CreateForMovie(s.ctx, s.testMovies[0], review)
	s.assertNoError(utilsErr)
	s.testReviews = append(s.testReviews, result1.ID)

	review2 := domain.Review{
		UserID:  s.testUsers[0],
		Rating:  9,
		Comment: "Second review from same user",
	}

	_, utilsErr = s.repo.CreateForMovie(s.ctx, s.testMovies[0], review2)

	s.assertErrorCode(utilsErr, 409)
}

func (s *ReviewRepositoryIntegrationTestSuite) TestUpdateReview() {
	review := domain.Review{
		UserID:  s.testUsers[0],
		Rating:  6,
		Comment: "Average movie",
	}

	created, utilsErr := s.repo.CreateForMovie(s.ctx, s.testMovies[0], review)
	s.assertNoError(utilsErr)
	s.testReviews = append(s.testReviews, created.ID)

	updateData := domain.Review{
		Rating:  9,
		Comment: "Actually, it's a great movie after rewatching!",
	}

	updated, utilsErr := s.repo.Update(s.ctx, created.ID, updateData)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), created.ID, updated.ID)
	assert.Equal(s.T(), s.testMovies[0], updated.MovieID)
	assert.Equal(s.T(), s.testUsers[0], updated.UserID)
	assert.Equal(s.T(), 9, updated.Rating)
	assert.Equal(s.T(), "Actually, it's a great movie after rewatching!", updated.Comment)
}

func (s *ReviewRepositoryIntegrationTestSuite) TestUpdateReviewChangeComment() {
	review := domain.Review{
		UserID:  s.testUsers[0],
		Rating:  7,
		Comment: "Good movie with some flaws",
	}

	created, utilsErr := s.repo.CreateForMovie(s.ctx, s.testMovies[0], review)
	s.assertNoError(utilsErr)
	s.testReviews = append(s.testReviews, created.ID)

	updateData := domain.Review{
		Rating:  8,
		Comment: "Improved after second watch",
	}

	updated, utilsErr := s.repo.Update(s.ctx, created.ID, updateData)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), 8, updated.Rating)
	assert.Equal(s.T(), "Improved after second watch", updated.Comment)
}

func (s *ReviewRepositoryIntegrationTestSuite) TestUpdateReview_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	updateData := domain.Review{
		Rating:  10,
		Comment: "Should not update",
	}

	_, utilsErr := s.repo.Update(s.ctx, nonExistentID, updateData)

	s.assertErrorCode(utilsErr, 404)
}

func (s *ReviewRepositoryIntegrationTestSuite) TestDeleteReview() {
	review := domain.Review{
		UserID:  s.testUsers[0],
		Rating:  5,
		Comment: "Not my favorite",
	}

	created, utilsErr := s.repo.CreateForMovie(s.ctx, s.testMovies[0], review)
	s.assertNoError(utilsErr)

	utilsErr = s.repo.Delete(s.ctx, created.ID)
	s.assertNoError(utilsErr)

	utilsErr = s.repo.Delete(s.ctx, created.ID)
	s.assertErrorCode(utilsErr, 404)
}

func (s *ReviewRepositoryIntegrationTestSuite) TestDeleteReview_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	utilsErr := s.repo.Delete(s.ctx, nonExistentID)

	s.assertErrorCode(utilsErr, 404)
}

func (s *ReviewRepositoryIntegrationTestSuite) TestGetAllReviews() {
	reviews := []domain.Review{
		{
			UserID:  s.testUsers[0],
			Rating:  8,
			Comment: "Great movie!",
		},
		{
			UserID:  s.testUsers[1],
			Rating:  9,
			Comment: "Excellent film",
		},
	}

	for _, review := range reviews {
		created, utilsErr := s.repo.CreateForMovie(s.ctx, s.testMovies[0], review)
		s.assertNoError(utilsErr)
		s.testReviews = append(s.testReviews, created.ID)
	}

	allReviews, total, utilsErr := s.repo.GetAll(s.ctx, domain.ReviewFilters{}, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, len(reviews))
	assert.GreaterOrEqual(s.T(), len(allReviews), len(reviews))
}

func (s *ReviewRepositoryIntegrationTestSuite) TestGetAllReviewsWithMovieFilter() {
	review := domain.Review{
		UserID:  s.testUsers[0],
		Rating:  8,
		Comment: "Good movie",
	}

	created, utilsErr := s.repo.CreateForMovie(s.ctx, s.testMovies[0], review)
	s.assertNoError(utilsErr)
	s.testReviews = append(s.testReviews, created.ID)

	filters := domain.ReviewFilters{
		MovieID: s.testMovies[0],
	}

	reviews, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)
	assert.GreaterOrEqual(s.T(), len(reviews), 1)

	for _, rev := range reviews {
		assert.Equal(s.T(), s.testMovies[0], rev.MovieID)
	}
}

func (s *ReviewRepositoryIntegrationTestSuite) TestGetAllReviewsWithUserFilter() {
	review := domain.Review{
		UserID:  s.testUsers[0],
		Rating:  7,
		Comment: "User specific review",
	}

	created, utilsErr := s.repo.CreateForMovie(s.ctx, s.testMovies[0], review)
	s.assertNoError(utilsErr)
	s.testReviews = append(s.testReviews, created.ID)

	filters := domain.ReviewFilters{
		UserID: s.testUsers[0],
	}

	reviews, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)
	assert.GreaterOrEqual(s.T(), len(reviews), 1)

	for _, rev := range reviews {
		assert.Equal(s.T(), s.testUsers[0], rev.UserID)
	}
}

func (s *ReviewRepositoryIntegrationTestSuite) TestGetAllReviewsWithRatingMinFilter() {
	reviews := []domain.Review{
		{
			UserID:  s.testUsers[0],
			Rating:  6,
			Comment: "Good",
		},
		{
			UserID:  s.testUsers[1],
			Rating:  9,
			Comment: "Excellent",
		},
	}

	for _, review := range reviews {
		created, utilsErr := s.repo.CreateForMovie(s.ctx, s.testMovies[0], review)
		s.assertNoError(utilsErr)
		s.testReviews = append(s.testReviews, created.ID)
	}

	filters := domain.ReviewFilters{
		RatingMin: 8,
	}

	reviewsResult, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)

	for _, rev := range reviewsResult {
		assert.GreaterOrEqual(s.T(), rev.Rating, 8)
	}
}

func (s *ReviewRepositoryIntegrationTestSuite) TestGetAllReviewsWithRatingMaxFilter() {
	reviews := []domain.Review{
		{
			UserID:  s.testUsers[0],
			Rating:  4,
			Comment: "Average",
		},
		{
			UserID:  s.testUsers[1],
			Rating:  7,
			Comment: "Good",
		},
	}

	for _, review := range reviews {
		created, utilsErr := s.repo.CreateForMovie(s.ctx, s.testMovies[0], review)
		s.assertNoError(utilsErr)
		s.testReviews = append(s.testReviews, created.ID)
	}

	filters := domain.ReviewFilters{
		RatingMax: 5,
	}

	reviewsResult, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)

	for _, rev := range reviewsResult {
		assert.LessOrEqual(s.T(), rev.Rating, 5)
	}
}

func (s *ReviewRepositoryIntegrationTestSuite) TestCreateReviewWithMinRating() {
	review := domain.Review{
		UserID:  s.testUsers[0],
		Rating:  1,
		Comment: "Min",
	}

	result, utilsErr := s.repo.CreateForMovie(s.ctx, s.testMovies[0], review)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), 1, result.Rating)
	s.testReviews = append(s.testReviews, result.ID)
}

func (s *ReviewRepositoryIntegrationTestSuite) TestCreateReviewWithMaxRating() {
	review := domain.Review{
		UserID:  s.testUsers[0],
		Rating:  10,
		Comment: "Max",
	}

	result, utilsErr := s.repo.CreateForMovie(s.ctx, s.testMovies[0], review)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), 10, result.Rating)
	s.testReviews = append(s.testReviews, result.ID)
}

func (s *ReviewRepositoryIntegrationTestSuite) TestGetAllReviewsWithPagination() {
	reviews := []domain.Review{
		{
			UserID:  s.testUsers[0],
			Rating:  6,
			Comment: "Review A",
		},
		{
			UserID:  s.testUsers[1],
			Rating:  7,
			Comment: "Review B",
		},
	}

	for _, review := range reviews {
		created, utilsErr := s.repo.CreateForMovie(s.ctx, s.testMovies[0], review)
		s.assertNoError(utilsErr)
		s.testReviews = append(s.testReviews, created.ID)
	}

	reviewsPage1, total, utilsErr := s.repo.GetAll(s.ctx, domain.ReviewFilters{}, 1, 1)
	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 2)
	assert.Len(s.T(), reviewsPage1, 1)
}

func (s *ReviewRepositoryIntegrationTestSuite) TestGetAllReviewsWithRatingRangeFilter() {
	reviews := []domain.Review{
		{
			UserID:  s.testUsers[0],
			Rating:  3,
			Comment: "Poor",
		},
		{
			UserID:  s.testUsers[1],
			Rating:  6,
			Comment: "Good",
		},
	}

	for i, review := range reviews {
		created, utilsErr := s.repo.CreateForMovie(s.ctx, s.testMovies[i], review)
		s.assertNoError(utilsErr)
		s.testReviews = append(s.testReviews, created.ID)
	}

	filters := domain.ReviewFilters{
		RatingMin: 5,
		RatingMax: 8,
	}

	reviewsResult, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)

	for _, rev := range reviewsResult {
		assert.GreaterOrEqual(s.T(), rev.Rating, 5)
		assert.LessOrEqual(s.T(), rev.Rating, 8)
	}
}
