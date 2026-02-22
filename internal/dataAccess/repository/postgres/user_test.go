package postgres

import (
	"context"
	"strconv"
	"testing"
	"time"

	"cw/internal/dataAccess/config"
	domain "cw/internal/domain/models"
	"cw/internal/utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryIntegrationTestSuite struct {
	suite.Suite
	repo      *UserRepository
	ctx       context.Context
	testUsers []string
	jwtSecret string
}

func TestUserRepositoryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryIntegrationTestSuite))
}

func (s *UserRepositoryIntegrationTestSuite) SetupSuite() {
	db, err := config.NewTestDatabase()
	if err != nil {
		s.T().Fatalf("Failed to connect to test database: %v", err)
	}
	s.jwtSecret = "test-secret-key"
	tokenDuration := 24 * time.Hour
	s.repo = NewUserRepository(db, s.jwtSecret, tokenDuration)
	s.ctx = context.Background()
	s.cleanDatabase()
}

func (s *UserRepositoryIntegrationTestSuite) cleanDatabase() {
	for _, id := range s.testUsers {
		s.repo.Delete(s.ctx, id)
	}
	s.testUsers = []string{}
}

func (s *UserRepositoryIntegrationTestSuite) SetupTest() {
	s.testUsers = []string{}
}

func (s *UserRepositoryIntegrationTestSuite) TearDownTest() {
	s.cleanDatabase()
}

func (s *UserRepositoryIntegrationTestSuite) TearDownSuite() {
	s.cleanDatabase()
}

func (s *UserRepositoryIntegrationTestSuite) assertNoError(utilsErr *utils.Error) {
	if utilsErr != nil {
		s.T().Errorf("Expected no error, got: %v", utilsErr)
	}
}

func (s *UserRepositoryIntegrationTestSuite) assertErrorCode(utilsErr *utils.Error, expectedCode int) {
	if utilsErr == nil {
		s.T().Error("Expected error, got nil")
		return
	}
	assert.Equal(s.T(), expectedCode, utilsErr.Code)
}

func (s *UserRepositoryIntegrationTestSuite) createTestUser(emailSuffix string, isAdmin bool) domain.User {
	user := domain.User{
		Name:         "Test User " + emailSuffix,
		Email:        "test" + emailSuffix + "@example.com",
		PasswordHash: "hashed_password_" + emailSuffix,
		BirthDate:    "1990-01-01",
		IsAdmin:      isAdmin,
	}

	created, utilsErr := s.repo.Register(s.ctx, user)
	s.assertNoError(utilsErr)
	s.testUsers = append(s.testUsers, created.ID)

	return created
}

func (s *UserRepositoryIntegrationTestSuite) TestRegister() {
	user := domain.User{
		Name:         "John Doe",
		Email:        "john.doe@example.com",
		PasswordHash: "hashed_password",
		BirthDate:    "1990-01-01",
		IsAdmin:      false,
	}

	result, utilsErr := s.repo.Register(s.ctx, user)

	s.assertNoError(utilsErr)
	assert.NotEmpty(s.T(), result.ID)
	assert.Equal(s.T(), user.Name, result.Name)
	assert.Equal(s.T(), user.Email, result.Email)
	assert.Equal(s.T(), user.PasswordHash, result.PasswordHash)
	assert.Equal(s.T(), user.BirthDate, result.BirthDate)
	assert.Equal(s.T(), user.IsAdmin, result.IsAdmin)

	s.testUsers = append(s.testUsers, result.ID)
}

func (s *UserRepositoryIntegrationTestSuite) TestRegister_DuplicateEmail() {
	user := domain.User{
		Name:         "Test User",
		Email:        "duplicate@example.com",
		PasswordHash: "hashed_password",
		BirthDate:    "1990-01-01",
		IsAdmin:      false,
	}

	created, utilsErr := s.repo.Register(s.ctx, user)
	s.assertNoError(utilsErr)
	s.testUsers = append(s.testUsers, created.ID)

	user2 := domain.User{
		Name:         "Another User",
		Email:        "duplicate@example.com",
		PasswordHash: "different_hash",
		BirthDate:    "1995-01-01",
		IsAdmin:      true,
	}

	_, utilsErr = s.repo.Register(s.ctx, user2)
	s.assertErrorCode(utilsErr, 409)
}

func (s *UserRepositoryIntegrationTestSuite) TestLogin_Success() {
	email := "login@example.com"
	passwordHash := "correct_password_hash"

	user := domain.User{
		Name:         "Login User",
		Email:        email,
		PasswordHash: passwordHash,
		BirthDate:    "1990-01-01",
		IsAdmin:      false,
	}

	created, utilsErr := s.repo.Register(s.ctx, user)
	s.assertNoError(utilsErr)
	s.testUsers = append(s.testUsers, created.ID)

	credentials := domain.User{
		Email:        email,
		PasswordHash: passwordHash,
	}

	authResponse, utilsErr := s.repo.Login(s.ctx, credentials)

	s.assertNoError(utilsErr)
	assert.NotEmpty(s.T(), authResponse.Token)
	assert.Equal(s.T(), created.ID, authResponse.UserID)
}

func (s *UserRepositoryIntegrationTestSuite) TestLogin_InvalidEmail() {
	credentials := domain.User{
		Email:        "nonexistent@example.com",
		PasswordHash: "any_password",
	}

	_, utilsErr := s.repo.Login(s.ctx, credentials)

	s.assertErrorCode(utilsErr, 404)
}

func (s *UserRepositoryIntegrationTestSuite) TestLogin_InvalidPassword() {
	email := "testlogin@example.com"

	user := domain.User{
		Name:         "Test User",
		Email:        email,
		PasswordHash: "correct_password",
		BirthDate:    "1990-01-01",
		IsAdmin:      false,
	}

	created, utilsErr := s.repo.Register(s.ctx, user)
	s.assertNoError(utilsErr)
	s.testUsers = append(s.testUsers, created.ID)

	credentials := domain.User{
		Email:        email,
		PasswordHash: "wrong_password",
	}

	_, utilsErr = s.repo.Login(s.ctx, credentials)

	s.assertErrorCode(utilsErr, 403)
}

func (s *UserRepositoryIntegrationTestSuite) TestGetUserByID() {
	user := s.createTestUser("getbyid", false)

	retrieved, utilsErr := s.repo.GetByID(s.ctx, user.ID)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), user.ID, retrieved.ID)
	assert.Equal(s.T(), user.Name, retrieved.Name)
	assert.Equal(s.T(), user.Email, retrieved.Email)
	assert.Equal(s.T(), user.PasswordHash, retrieved.PasswordHash)
	assert.Equal(s.T(), user.BirthDate, retrieved.BirthDate)
	assert.Equal(s.T(), user.IsAdmin, retrieved.IsAdmin)
}

func (s *UserRepositoryIntegrationTestSuite) TestGetUserByID_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	_, utilsErr := s.repo.GetByID(s.ctx, nonExistentID)

	s.assertErrorCode(utilsErr, 404)
}

func (s *UserRepositoryIntegrationTestSuite) TestUpdateUser() {
	user := s.createTestUser("update", false)

	updateData := domain.User{
		Name:      "Updated Name",
		Email:     "updated@example.com",
		BirthDate: "1995-01-01",
	}

	updated, utilsErr := s.repo.Update(s.ctx, user.ID, updateData)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), user.ID, updated.ID)
	assert.Equal(s.T(), updateData.Name, updated.Name)
	assert.Equal(s.T(), updateData.Email, updated.Email)
	assert.Equal(s.T(), updateData.BirthDate, updated.BirthDate)
	assert.Equal(s.T(), user.PasswordHash, updated.PasswordHash)
	assert.Equal(s.T(), user.IsAdmin, updated.IsAdmin)
}

func (s *UserRepositoryIntegrationTestSuite) TestUpdateUser_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	updateData := domain.User{
		Name:      "Updated Name",
		Email:     "updated@example.com",
		BirthDate: "1995-01-01",
	}

	_, utilsErr := s.repo.Update(s.ctx, nonExistentID, updateData)

	s.assertErrorCode(utilsErr, 404)
}

func (s *UserRepositoryIntegrationTestSuite) TestUpdateUser_DuplicateEmail() {
	user1 := s.createTestUser("unique1", false)
	user2 := s.createTestUser("unique2", false)

	updateData := domain.User{
		Name:      user2.Name,
		Email:     user1.Email,
		BirthDate: user2.BirthDate,
	}

	_, utilsErr := s.repo.Update(s.ctx, user2.ID, updateData)

	s.assertErrorCode(utilsErr, 409)
}

func (s *UserRepositoryIntegrationTestSuite) TestUpdateAdminStatus() {
	user := s.createTestUser("adminStatus", false)

	updated, utilsErr := s.repo.UpdateAdminStatus(s.ctx, user.ID, true)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), user.ID, updated.ID)
	assert.True(s.T(), updated.IsAdmin)
	assert.Equal(s.T(), user.Name, updated.Name)
	assert.Equal(s.T(), user.Email, updated.Email)
	assert.Equal(s.T(), user.PasswordHash, updated.PasswordHash)
	assert.Equal(s.T(), user.BirthDate, updated.BirthDate)

	updatedBack, utilsErr := s.repo.UpdateAdminStatus(s.ctx, user.ID, false)
	s.assertNoError(utilsErr)
	assert.False(s.T(), updatedBack.IsAdmin)
}

func (s *UserRepositoryIntegrationTestSuite) TestUpdateAdminStatus_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	_, utilsErr := s.repo.UpdateAdminStatus(s.ctx, nonExistentID, true)

	s.assertErrorCode(utilsErr, 404)
}

func (s *UserRepositoryIntegrationTestSuite) TestDeleteUser() {
	user := s.createTestUser("delete", false)

	utilsErr := s.repo.Delete(s.ctx, user.ID)
	s.assertNoError(utilsErr)

	_, utilsErr = s.repo.GetByID(s.ctx, user.ID)
	s.assertErrorCode(utilsErr, 404)
}

func (s *UserRepositoryIntegrationTestSuite) TestDeleteUser_NotFound() {
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	utilsErr := s.repo.Delete(s.ctx, nonExistentID)

	s.assertErrorCode(utilsErr, 404)
}

func (s *UserRepositoryIntegrationTestSuite) TestGetAllUsers() {
	users := []domain.User{
		s.createTestUser("all1", false),
		s.createTestUser("all2", true),
		s.createTestUser("all3", false),
	}

	result, total, utilsErr := s.repo.GetAll(s.ctx, domain.UserFilters{}, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, len(users))
	assert.GreaterOrEqual(s.T(), len(result), len(users))
}

func (s *UserRepositoryIntegrationTestSuite) TestGetAllUsersWithNameFilter() {
	_ = s.createTestUser("namefilter1", false)
	_ = s.createTestUser("namefilter2", false)

	filters := domain.UserFilters{
		Name: "namefilter1",
	}

	result, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)

	for _, user := range result {
		assert.Contains(s.T(), user.Name, "namefilter1")
	}
}

func (s *UserRepositoryIntegrationTestSuite) TestGetAllUsersWithEmailFilter() {
	_ = s.createTestUser("emailfilter1", false)
	_ = s.createTestUser("emailfilter2", false)

	filters := domain.UserFilters{
		Email: "emailfilter1",
	}

	result, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)

	for _, user := range result {
		assert.Contains(s.T(), user.Email, "emailfilter1")
	}
}

func (s *UserRepositoryIntegrationTestSuite) TestGetAllUsersWithAdminFilter() {
	_ = s.createTestUser("admin", true)
	_ = s.createTestUser("regular", false)

	trueVal := true
	filters := domain.UserFilters{
		IsAdmin: &trueVal,
	}

	result, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 1)

	for _, user := range result {
		assert.True(s.T(), user.IsAdmin)
	}

	falseVal := false
	filters2 := domain.UserFilters{
		IsAdmin: &falseVal,
	}

	result2, total2, utilsErr := s.repo.GetAll(s.ctx, filters2, 1, 10)
	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total2, 1)

	for _, user := range result2 {
		assert.False(s.T(), user.IsAdmin)
	}
}

func (s *UserRepositoryIntegrationTestSuite) TestGetAllUsersWithPagination() {
	for i := 1; i <= 5; i++ {
		s.createTestUser(strconv.Itoa(i), false)
	}

	page1, total, utilsErr := s.repo.GetAll(s.ctx, domain.UserFilters{}, 1, 2)
	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, 5)
	assert.Len(s.T(), page1, 2)

	page2, _, utilsErr := s.repo.GetAll(s.ctx, domain.UserFilters{}, 2, 2)
	s.assertNoError(utilsErr)
	assert.Len(s.T(), page2, 2)

	page3, _, utilsErr := s.repo.GetAll(s.ctx, domain.UserFilters{}, 3, 2)
	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), len(page3), 1)
}

func (s *UserRepositoryIntegrationTestSuite) TestGetAllUsersOrderByName() {
	users := []domain.User{
		s.createTestUser("Zebra", false),
		s.createTestUser("Alpha", false),
		s.createTestUser("Charlie", false),
	}

	result, total, utilsErr := s.repo.GetAll(s.ctx, domain.UserFilters{}, 1, 10)

	s.assertNoError(utilsErr)
	assert.GreaterOrEqual(s.T(), total, len(users))

	if len(result) >= len(users) {
		for i := 1; i < len(result); i++ {
			assert.True(s.T(), result[i-1].Name <= result[i].Name,
				"Results should be sorted by name: %s should come before %s",
				result[i-1].Name, result[i].Name)
		}
	}
}

func (s *UserRepositoryIntegrationTestSuite) TestGetAllUsersEmptyResult() {
	filters := domain.UserFilters{
		Name: "NonExistentUserNameThatShouldNotExist",
	}

	result, total, utilsErr := s.repo.GetAll(s.ctx, filters, 1, 10)

	s.assertNoError(utilsErr)
	assert.Equal(s.T(), 0, total)
	assert.Empty(s.T(), result)
}

func (s *UserRepositoryIntegrationTestSuite) TestGenerateJWTToken() {
	userID := uuid.New().String()
	isAdmin := true

	token, err := s.repo.generateJWTToken(userID, isAdmin)

	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), token)
}
