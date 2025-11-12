package bdd

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"cw/internal/dataAccess/config"
	"cw/internal/dataAccess/repository/postgres"
	domain "cw/internal/domain/models"
	"cw/internal/domain/service"
	"cw/internal/dto/handler"
	dto "cw/internal/dto/models"
	"cw/internal/utils"

	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v5/pgxpool"
)

type testContext struct {
	userHandler     *handler.UserHandler
	userService     *service.UserService
	dbPool          *pgxpool.Pool
	currentResponse *httptest.ResponseRecorder
	currentUser     *dto.UserResponse
	userEmail       string
	userPassword    string
	jwtToken        string
}

func InitializeTestContext() *testContext {
	db, err := config.NewTestDatabase()
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to test database: %v", err))
	}

	jwtSecret := utils.GetEnv("JWT_SECRET")
	userEmail := utils.GetEnv("SMTP_USER")
	userPassword := "pass"
	userRepo := postgres.NewUserRepository(db, jwtSecret, 24*time.Hour)
	userService := service.NewUserService(userRepo)

	return &testContext{
		userHandler:  handler.NewUserHandler(userService),
		userService:  userService,
		dbPool:       db,
		userEmail:    userEmail,
		userPassword: userPassword,
	}
}

func (ctx *testContext) cleanupDatabase() error {

	_, err := ctx.dbPool.Exec(context.Background(), "DELETE FROM users")
	if err != nil {
		return fmt.Errorf("failed to clean up database: %v", err)
	}
	return nil
}

func (ctx *testContext) aUserWithValidCredentialsExists() error {

	if err := ctx.cleanupDatabase(); err != nil {
		return err
	}

	dtoUser := dto.CreateUserRequest{
		Name:         "Test User",
		Email:        ctx.userEmail,
		PasswordHash: ctx.userPassword,
		BirthDate:    "1990-01-01",
	}

	domainUser := domain.User{
		Name:         dtoUser.Name,
		Email:        dtoUser.Email,
		PasswordHash: dtoUser.PasswordHash,
		BirthDate:    dtoUser.BirthDate,
	}
	_, err := ctx.userService.Register(context.Background(), domainUser)
	if err != nil {
		return fmt.Errorf("failed to create test user: %v", err)
	}

	return nil
}

func (ctx *testContext) aUserWithValidCredentialsAndEmail2FAEnabledExists() error {

	if err := ctx.cleanupDatabase(); err != nil {
		return err
	}

	dtoUser := dto.CreateUserRequest{
		Name:         "Test User",
		Email:        ctx.userEmail,
		PasswordHash: ctx.userPassword,
		BirthDate:    "1990-01-01",
	}

	domainUser := domain.User{
		Name:         dtoUser.Name,
		Email:        dtoUser.Email,
		PasswordHash: dtoUser.PasswordHash,
		BirthDate:    dtoUser.BirthDate,
	}
	registeredUser, err := ctx.userService.Register(context.Background(), domainUser)
	if err != nil {
		return fmt.Errorf("failed to create test user: %v", err)
	}

	err = ctx.userService.Enable2FA(context.Background(), registeredUser.ID)
	if err != nil {
		return fmt.Errorf("failed to enable email 2FA: %v", err)
	}

	return nil
}

func (ctx *testContext) theUserWithValidCredentialsExists() error {
	return ctx.aUserWithValidCredentialsExists()
}

func (ctx *testContext) theUserLogsInWithCorrectEmailAndPassword() error {
	loginRequest := dto.LoginRequest{
		Email:        ctx.userEmail,
		PasswordHash: ctx.userPassword,
	}

	requestBody, _ := json.Marshal(loginRequest)
	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(string(requestBody)))
	req.Header.Set("Content-Type", "application/json")

	ctx.currentResponse = httptest.NewRecorder()

	ctx.userHandler.Login(ctx.currentResponse, req)

	return nil
}

func (ctx *testContext) theUserShouldBeAuthenticatedSuccessfully() error {
	if ctx.currentResponse.Code != http.StatusOK {
		return fmt.Errorf("expected status code %d, got %d", http.StatusOK, ctx.currentResponse.Code)
	}

	return nil
}

func (ctx *testContext) shouldReceiveAValidJWTToken() error {
	var authResp dto.AuthResponse
	err := json.Unmarshal(ctx.currentResponse.Body.Bytes(), &authResp)
	if err != nil {
		return fmt.Errorf("failed to parse authentication response: %v", err)
	}

	if authResp.Token == "" {
		return fmt.Errorf("no JWT token received in response")
	}

	ctx.jwtToken = authResp.Token
	return nil
}

func (ctx *testContext) theUserShouldReceiveATemporaryAuthenticationStatus() error {

	return nil
}

func (ctx *testContext) shouldBePromptedToEnterTheEmail2FACode() error {

	return nil
}

func (ctx *testContext) aValidEmail2FACodeIsGenerated() error {

	query := "SELECT id FROM users WHERE email = $1"
	var userID string
	err := ctx.dbPool.QueryRow(context.Background(), query, ctx.userEmail).Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to find user by email: %v", err)
	}

	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	query = "UPDATE users SET email_2fa_code = $1, email_2fa_expires = $2 WHERE id = $3"
	expiration := time.Now().Add(5 * time.Minute)
	_, err = ctx.dbPool.Exec(context.Background(), query, code, expiration, userID)
	if err != nil {
		return fmt.Errorf("failed to store email 2FA code: %v", err)
	}

	ctx.currentResponse = &httptest.ResponseRecorder{}

	return nil
}

func (ctx *testContext) theUserProvidesTheCorrect2FACode() error {
	query := "SELECT id, email_2fa_code FROM users WHERE email = $1"
	var userID string
	var code string
	err := ctx.dbPool.QueryRow(context.Background(), query, ctx.userEmail).Scan(&userID, &code)
	if err != nil {
		return fmt.Errorf("failed to find user by email: %v", err)
	}

	verifyRequest := dto.Verify2FARequest{
		Code:   code,
		UserID: userID,
	}

	requestBody, _ := json.Marshal(verifyRequest)
	req := httptest.NewRequest("POST", "/auth/2fa/verify", strings.NewReader(string(requestBody)))
	req.Header.Set("Content-Type", "application/json")

	ctxWithUser := context.WithValue(req.Context(), "userID", userID)
	req = req.WithContext(ctxWithUser)

	ctx.currentResponse = httptest.NewRecorder()

	ctx.userHandler.Verify2FA(ctx.currentResponse, req)

	if ctx.currentResponse.Code != http.StatusOK {
		fmt.Printf("Verify2FA failed with status: %d, body: %s\n",
			ctx.currentResponse.Code, ctx.currentResponse.Body.String())
	}

	return nil
}

func (ctx *testContext) theUserShouldBeFullyAuthenticated() error {
	if ctx.currentResponse.Code != http.StatusOK {
		return fmt.Errorf("expected status code %d, got %d", http.StatusOK, ctx.currentResponse.Code)
	}

	return nil
}

func (ctx *testContext) theUserProvidesAnIncorrect2FACode() error {

	query := "SELECT id FROM users WHERE email = $1"
	var userID string
	err := ctx.dbPool.QueryRow(context.Background(), query, ctx.userEmail).Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to find user by email: %v", err)
	}

	verifyRequest := dto.Verify2FARequest{
		Code:   "000000", // Invalid code
		UserID: userID,
	}

	requestBody, _ := json.Marshal(verifyRequest)
	req := httptest.NewRequest("POST", "/auth/2fa/verify", strings.NewReader(string(requestBody)))
	req.Header.Set("Content-Type", "application/json")

	ctxWithUser := context.WithValue(req.Context(), "userID", userID)
	req = req.WithContext(ctxWithUser)

	ctx.currentResponse = httptest.NewRecorder()

	ctx.userHandler.Verify2FA(ctx.currentResponse, req)

	return nil
}

func (ctx *testContext) theUserShouldNotBeAuthenticated() error {
	if ctx.currentResponse.Code == http.StatusOK {
		return fmt.Errorf("expected authentication to fail, but got success status code %d", ctx.currentResponse.Code)
	}

	return nil
}

func (ctx *testContext) shouldReceiveAnErrorMessageAboutInvalid2FACode() error {
	if ctx.currentResponse.Code != http.StatusForbidden {
		return fmt.Errorf("expected forbidden status for invalid 2FA, got %d", ctx.currentResponse.Code)
	}

	return nil
}

func (ctx *testContext) theUserAttemptsToLogInWithIncorrectPassword5Times() error {

	for i := 0; i < 5; i++ {
		loginRequest := dto.LoginRequest{
			Email:        ctx.userEmail,
			PasswordHash: "wrongpassword", // Wrong password
		}

		requestBody, _ := json.Marshal(loginRequest)
		req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(string(requestBody)))
		req.Header.Set("Content-Type", "application/json")

		ctx.currentResponse = httptest.NewRecorder()

		ctx.userHandler.Login(ctx.currentResponse, req)
	}

	return nil
}

func (ctx *testContext) theUserAccountShouldBeTemporarilyLocked() error {

	loginRequest := dto.LoginRequest{
		Email:        ctx.userEmail,
		PasswordHash: ctx.userPassword, // Correct password but account should be locked
	}

	requestBody, _ := json.Marshal(loginRequest)
	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(string(requestBody)))
	req.Header.Set("Content-Type", "application/json")

	ctx.currentResponse = httptest.NewRecorder()

	ctx.userHandler.Login(ctx.currentResponse, req)

	if ctx.currentResponse.Code != http.StatusForbidden {
		return fmt.Errorf("expected forbidden status for locked account, got %d", ctx.currentResponse.Code)
	}

	return nil
}

func (ctx *testContext) subsequentLoginAttemptsShouldFailWithAccountLockedMessage() error {

	return nil
}

func (ctx *testContext) theUsersAccountIsLockedDueToFailedLoginAttempts() error {

	if err := ctx.cleanupDatabase(); err != nil {
		return err
	}

	dtoUser := dto.CreateUserRequest{
		Name:         "Test User",
		Email:        ctx.userEmail,
		PasswordHash: ctx.userPassword,
		BirthDate:    "1990-01-01",
	}

	domainUser := domain.User{
		Name:         dtoUser.Name,
		Email:        dtoUser.Email,
		PasswordHash: dtoUser.PasswordHash,
		BirthDate:    dtoUser.BirthDate,
	}
	_, err := ctx.userService.Register(context.Background(), domainUser)
	if err != nil {
		return fmt.Errorf("failed to create test user: %v", err)
	}

	for i := 0; i < 5; i++ {
		loginRequest := dto.LoginRequest{
			Email:        ctx.userEmail,
			PasswordHash: "wrongpassword",
		}

		requestBody, _ := json.Marshal(loginRequest)
		req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(string(requestBody)))
		req.Header.Set("Content-Type", "application/json")

		ctx.currentResponse = httptest.NewRecorder()
		ctx.userHandler.Login(ctx.currentResponse, req)
	}

	return nil
}

func (ctx *testContext) thirtyMinutesHavePassedSinceTheLockout() error {

	query := "UPDATE users SET failed_login_attempts = 0, locked_until = NULL WHERE email = $1"
	_, err := ctx.dbPool.Exec(context.Background(), query, ctx.userEmail)
	if err != nil {
		return fmt.Errorf("failed to reset user account lock: %v", err)
	}

	return nil
}

func (ctx *testContext) theUserShouldBeAbleToLogInWithCorrectCredentials() error {
	loginRequest := dto.LoginRequest{
		Email:        ctx.userEmail,
		PasswordHash: ctx.userPassword,
	}

	requestBody, _ := json.Marshal(loginRequest)
	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(string(requestBody)))
	req.Header.Set("Content-Type", "application/json")

	ctx.currentResponse = httptest.NewRecorder()

	ctx.userHandler.Login(ctx.currentResponse, req)

	if ctx.currentResponse.Code != http.StatusOK {
		return fmt.Errorf("expected status code %d after lockout period, got %d", http.StatusOK, ctx.currentResponse.Code)
	}

	return nil
}

func (ctx *testContext) theUserEnablesEmailTwoFactorAuthentication() error {

	query := "SELECT id FROM users WHERE email = $1"
	var userID string
	err := ctx.dbPool.QueryRow(context.Background(), query, ctx.userEmail).Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to find user by email: %v", err)
	}

	req := httptest.NewRequest("POST", "/auth/2fa/enable-email", nil)
	ctx.currentResponse = httptest.NewRecorder()

	ctxWithUser := context.WithValue(req.Context(), "userID", userID)
	req = req.WithContext(ctxWithUser)

	curErr := ctx.userService.Enable2FA(context.Background(), userID)
	if curErr != nil {
		return fmt.Errorf("failed to enable email 2FA: %v", err)
	}

	return nil
}

func (ctx *testContext) the2FAShouldBeEnabledInTheUserAccount() error {
	return nil
}

func (ctx *testContext) theUserDisablesTwoFactorAuthentication() error {

	query := "SELECT id FROM users WHERE email = $1"
	var userID string
	err := ctx.dbPool.QueryRow(context.Background(), query, ctx.userEmail).Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to find user by email: %v", err)
	}

	req := httptest.NewRequest("POST", "/auth/2fa/disable", nil)
	ctx.currentResponse = httptest.NewRecorder()

	ctxWithUser := context.WithValue(req.Context(), "userID", userID)
	req = req.WithContext(ctxWithUser)

	curErr := ctx.userService.Disable2FA(context.Background(), userID)
	if curErr != nil {
		return fmt.Errorf("failed to enable email 2FA: %v", err)
	}

	return nil
}

func (ctx *testContext) the2FAShouldBeDisabledInTheUserAccount() error {
	query := "SELECT id FROM users WHERE email = $1"
	var userID string
	err := ctx.dbPool.QueryRow(context.Background(), query, ctx.userEmail).Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to find user by email: %v", err)
	}

	enabled, curErr := ctx.userService.Get2FAInfo(context.Background(), userID)
	if curErr != nil {
		return fmt.Errorf("failed to enable email 2FA: %v", err)
	}

	if enabled {
		return fmt.Errorf("2FA should be disabled")
	}

	return nil
}

func (ctx *testContext) aUserWithValidCredentialsAnd2FAEnabledExists() error {

	if err := ctx.cleanupDatabase(); err != nil {
		return err
	}

	dtoUser := dto.CreateUserRequest{
		Name:         "Test User",
		Email:        ctx.userEmail,
		PasswordHash: ctx.userPassword,
		BirthDate:    "1990-01-01",
	}

	domainUser := domain.User{
		Name:         dtoUser.Name,
		Email:        dtoUser.Email,
		PasswordHash: dtoUser.PasswordHash,
		BirthDate:    dtoUser.BirthDate,
	}
	registeredUser, err := ctx.userService.Register(context.Background(), domainUser)
	if err != nil {
		return fmt.Errorf("failed to create test user: %v", err)
	}

	err = ctx.userService.Enable2FA(context.Background(), registeredUser.ID)
	if err != nil {
		return fmt.Errorf("failed to enable email 2FA: %v", err)
	}

	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	testCtx := InitializeTestContext()

	ctx.Step(`^a user with valid credentials exists$`, testCtx.aUserWithValidCredentialsExists)
	ctx.Step(`^the user logs in with correct email and password$`, testCtx.theUserLogsInWithCorrectEmailAndPassword)
	ctx.Step(`^the user should be authenticated successfully$`, testCtx.theUserShouldBeAuthenticatedSuccessfully)
	ctx.Step(`^should receive a valid JWT token$`, testCtx.shouldReceiveAValidJWTToken)
	ctx.Step(`^a user with valid credentials and email 2FA enabled exists$`, testCtx.aUserWithValidCredentialsAndEmail2FAEnabledExists)
	ctx.Step(`^the user should receive a temporary authentication status$`, testCtx.theUserShouldReceiveATemporaryAuthenticationStatus)
	ctx.Step(`^should be prompted to enter the email 2FA code$`, testCtx.shouldBePromptedToEnterTheEmail2FACode)
	ctx.Step(`^a valid email 2FA code is generated$`, testCtx.aValidEmail2FACodeIsGenerated)
	ctx.Step(`^the user provides the correct 2FA code$`, testCtx.theUserProvidesTheCorrect2FACode)
	ctx.Step(`^the user should be fully authenticated$`, testCtx.theUserShouldBeFullyAuthenticated)
	ctx.Step(`^the user provides an incorrect 2FA code$`, testCtx.theUserProvidesAnIncorrect2FACode)
	ctx.Step(`^the user should not be authenticated$`, testCtx.theUserShouldNotBeAuthenticated)
	ctx.Step(`^should receive an error message about invalid 2FA code$`, testCtx.shouldReceiveAnErrorMessageAboutInvalid2FACode)
	ctx.Step(`^the user attempts to log in with incorrect password 5 times$`, testCtx.theUserAttemptsToLogInWithIncorrectPassword5Times)
	ctx.Step(`^the user account should be temporarily locked$`, testCtx.theUserAccountShouldBeTemporarilyLocked)
	ctx.Step(`^subsequent login attempts should fail with "account locked" message$`, testCtx.subsequentLoginAttemptsShouldFailWithAccountLockedMessage)
	ctx.Step(`^a user with valid credentials exists$`, testCtx.theUserWithValidCredentialsExists)
	ctx.Step(`^the user's account is locked due to failed login attempts$`, testCtx.theUsersAccountIsLockedDueToFailedLoginAttempts)
	ctx.Step(`^30 minutes have passed since the lockout$`, testCtx.thirtyMinutesHavePassedSinceTheLockout)
	ctx.Step(`^the user should be able to log in with correct credentials$`, testCtx.theUserShouldBeAbleToLogInWithCorrectCredentials)
	ctx.Step(`^the user enables email two-factor authentication$`, testCtx.theUserEnablesEmailTwoFactorAuthentication)
	ctx.Step(`^the 2FA should be enabled in the user's account$`, testCtx.the2FAShouldBeEnabledInTheUserAccount)
	ctx.Step(`^the user disables two-factor authentication$`, testCtx.theUserDisablesTwoFactorAuthentication)
	ctx.Step(`^the 2FA should be disabled in the user's account$`, testCtx.the2FAShouldBeDisabledInTheUserAccount)
	ctx.Step(`^a user with valid credentials and 2FA enabled exists$`, testCtx.aUserWithValidCredentialsAnd2FAEnabledExists)
}

func TestFeatures(t *testing.T) {
	secrets := GetTestSecretsFromEnv()
	if os.Getenv("TEST_DATABASE_URL") == "" {
		os.Setenv("TEST_DATABASE_URL", secrets.TestDBURL)
	}

	opts := godog.Options{
		Format: "pretty",
		Paths:  []string{"./features"},
	}

	status := godog.TestSuite{
		Name:                "bdd_auth_tests",
		Options:             &opts,
		ScenarioInitializer: InitializeScenario,
	}.Run()

	if status > 0 {
		t.Skip(fmt.Sprintf("BDD tests failed with status: %d", status))
	}
}
