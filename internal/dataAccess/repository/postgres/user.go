package postgres

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	entity "cw/internal/dataAccess/models"
	domain "cw/internal/domain/models"
	"cw/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db            *pgxpool.Pool
	jwtSecret     string
	tokenDuration time.Duration
}

func NewUserRepository(db *pgxpool.Pool, jwtSecret string, tokenDuration time.Duration) *UserRepository {
	return &UserRepository{
		db:            db,
		jwtSecret:     jwtSecret,
		tokenDuration: tokenDuration,
	}
}

func (r *UserRepository) Login(ctx context.Context, credentials domain.User) (domain.AuthResponse, *utils.Error) {
	var userEntity entity.User

	query := "SELECT id, name, email, password_hash, birth_date, is_admin, two_fa_enabled, email_2fa_code, email_2fa_expires, failed_login_attempts, locked_until FROM users WHERE email = $1"
	err := r.db.QueryRow(ctx, query, credentials.Email).Scan(
		&userEntity.ID, &userEntity.Name, &userEntity.Email, &userEntity.PasswordHash, &userEntity.BirthDate, &userEntity.IsAdmin,
		&userEntity.TwoFANeeded,
		&userEntity.Email2FACode, &userEntity.Email2FAExpires, &userEntity.FailedLoginAttempts, &userEntity.LockedUntil,
	)
	if err != nil {
		return domain.AuthResponse{}, utils.ConvertError(err)
	}

	if userEntity.LockedUntil != nil && time.Now().Before(*userEntity.LockedUntil) {
		return domain.AuthResponse{}, utils.NewForbidden("account is temporarily locked due to multiple failed login attempts", nil)
	}

	if userEntity.PasswordHash != credentials.PasswordHash {
		r.incrementFailedLoginAttempts(ctx, userEntity.ID)
		return domain.AuthResponse{}, utils.NewForbidden("invalid credentials", nil)
	}

	r.resetFailedLoginAttempts(ctx, userEntity.ID)

	if userEntity.TwoFANeeded {
		_, err := r.generateEmail2FACode(ctx, userEntity.ID)
		if err != nil {
			return domain.AuthResponse{}, utils.NewInternal("failed to generate email 2FA code", err)
		}
		return domain.AuthResponse{
			UserID:      userEntity.ID,
			Message:     "Two-factor authentication required. Check your email for the verification code.",
			TwoFANeeded: true,
		}, nil
	}

	token, err := r.generateJWTToken(userEntity.ID, userEntity.IsAdmin)
	if err != nil {
		return domain.AuthResponse{}, utils.NewInternal("failed to generate token", err)
	}

	return domain.AuthResponse{
		UserID: userEntity.ID,
		Token:  token,
	}, nil
}

func (r *UserRepository) Register(ctx context.Context, user domain.User) (domain.User, *utils.Error) {
	userEntity := entity.UserFromDomain(user)

	query := "INSERT INTO users (name, email, password_hash, birth_date, is_admin) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	err := r.db.QueryRow(ctx, query, userEntity.Name, userEntity.Email, userEntity.PasswordHash, userEntity.BirthDate, userEntity.IsAdmin).Scan(&userEntity.ID)
	if err != nil {
		return domain.User{}, utils.ConvertError(err)
	}

	return entity.UserToDomain(userEntity), nil
}

func (r *UserRepository) GetAll(ctx context.Context, filters domain.UserFilters, page, limit int) ([]domain.User, int, *utils.Error) {
	var whereClauses []string
	var args []interface{}
	argPos := 1

	if filters.Name != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", argPos))
		args = append(args, "%"+filters.Name+"%")
		argPos++
	}
	if filters.Email != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("email ILIKE $%d", argPos))
		args = append(args, "%"+filters.Email+"%")
		argPos++
	}
	if filters.IsAdmin != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("is_admin = $%d", argPos))
		args = append(args, *filters.IsAdmin)
		argPos++
	}

	countQuery := "SELECT COUNT(*) FROM users"
	if len(whereClauses) > 0 {
		countQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, utils.ConvertError(err)
	}

	query := "SELECT id, name, email, password_hash, birth_date, is_admin FROM users"
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}
	query += " ORDER BY name LIMIT $" + strconv.Itoa(argPos) + " OFFSET $" + strconv.Itoa(argPos+1)

	args = append(args, limit, (page-1)*limit)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, utils.ConvertError(err)
	}
	defer rows.Close()

	var userEntities []entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.BirthDate, &user.IsAdmin); err != nil {
			return nil, 0, utils.ConvertError(err)
		}
		userEntities = append(userEntities, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, utils.ConvertError(err)
	}

	return entity.UsersToDomain(userEntities), total, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (domain.User, *utils.Error) {
	var userEntity entity.User

	query := "SELECT id, name, email, password_hash, birth_date, is_admin FROM users WHERE id = $1"
	err := r.db.QueryRow(ctx, query, id).Scan(
		&userEntity.ID, &userEntity.Name, &userEntity.Email, &userEntity.PasswordHash, &userEntity.BirthDate, &userEntity.IsAdmin,
	)
	if err != nil {
		return domain.User{}, utils.ConvertError(err)
	}

	return entity.UserToDomain(userEntity), nil
}

func (r *UserRepository) Update(ctx context.Context, id string, user domain.User) (domain.User, *utils.Error) {
	userEntity := entity.UserFromDomain(user)

	query := "UPDATE users SET name = $1, email = $2, birth_date = $3 WHERE id = $4 RETURNING id, name, email, password_hash, birth_date, is_admin, two_fa_enabled, email_2fa_code, email_2fa_expires, failed_login_attempts, locked_until"
	var updatedUser entity.User
	err := r.db.QueryRow(ctx, query, userEntity.Name, userEntity.Email, userEntity.BirthDate, id).Scan(
		&updatedUser.ID, &updatedUser.Name, &updatedUser.Email, &updatedUser.PasswordHash, &updatedUser.BirthDate, &updatedUser.IsAdmin,
		&updatedUser.TwoFANeeded, &updatedUser.Email2FACode, &updatedUser.Email2FAExpires, &updatedUser.FailedLoginAttempts, &updatedUser.LockedUntil,
	)
	if err != nil {
		return domain.User{}, utils.ConvertError(err)
	}

	return entity.UserToDomain(updatedUser), nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id string, newPasswordHash string) *utils.Error {
	query := "UPDATE users SET password_hash = $1 WHERE id = $2"
	_, err := r.db.Exec(ctx, query, newPasswordHash, id)
	if err != nil {
		return utils.ConvertError(err)
	}
	return nil
}

func (r *UserRepository) VerifyCurrentPassword(ctx context.Context, id string, currentPassword string) *utils.Error {
	var storedPasswordHash string

	query := "SELECT password_hash FROM users WHERE id = $1"
	err := r.db.QueryRow(ctx, query, id).Scan(&storedPasswordHash)
	if err != nil {
		return utils.ConvertError(err)
	}

	if storedPasswordHash != currentPassword {
		return utils.NewForbidden("Текущий пароль неверен", nil)
	}

	return nil
}

func (r *UserRepository) UpdateAdminStatus(ctx context.Context, id string, isAdmin bool) (domain.User, *utils.Error) {
	query := "UPDATE users SET is_admin = $1 WHERE id = $2 RETURNING id, name, email, password_hash, birth_date, is_admin"
	var updatedUser entity.User
	err := r.db.QueryRow(ctx, query, isAdmin, id).Scan(
		&updatedUser.ID, &updatedUser.Name, &updatedUser.Email, &updatedUser.PasswordHash, &updatedUser.BirthDate, &updatedUser.IsAdmin,
	)
	if err != nil {
		return domain.User{}, utils.ConvertError(err)
	}

	return entity.UserToDomain(updatedUser), nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) *utils.Error {
	query := "DELETE FROM users WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return utils.ConvertError(err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return utils.NewNotFound("user not found", nil)
	}

	return nil
}

func (r *UserRepository) generateJWTToken(userID string, isAdmin bool) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    userID,
		"is_admin":   isAdmin,
		"exp":        time.Now().Add(r.tokenDuration).Unix(),
		"iat":        time.Now().Unix(),
		"auth_level": "full",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(r.jwtSecret))
}

func (r *UserRepository) incrementFailedLoginAttempts(ctx context.Context, userID string) *utils.Error {
	// Lock account after 5 failed attempts for 30 minutes
	query := `
		UPDATE users 
		SET failed_login_attempts = failed_login_attempts + 1,
			locked_until = CASE 
				WHEN failed_login_attempts + 1 >= 5 
				THEN NOW() + INTERVAL '30 minutes'
				ELSE locked_until
			END
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return utils.ConvertError(err)
	}
	return nil
}

func (r *UserRepository) resetFailedLoginAttempts(ctx context.Context, userID string) *utils.Error {
	query := `
		UPDATE users 
		SET failed_login_attempts = 0, locked_until = NULL
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return utils.ConvertError(err)
	}
	return nil
}

func (r *UserRepository) Enable2FA(ctx context.Context, userID string) *utils.Error {
	query := "UPDATE users SET two_fa_enabled = true WHERE id = $1"
	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return utils.ConvertError(err)
	}
	return nil
}

func (r *UserRepository) Disable2FA(ctx context.Context, userID string) *utils.Error {
	query := "UPDATE users SET two_fa_enabled = false WHERE id = $1"
	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return utils.ConvertError(err)
	}
	return nil
}

func (r *UserRepository) Get2FAInfo(ctx context.Context, userID string) (bool, *utils.Error) {
	var enabled bool

	query := "SELECT two_fa_enabled FROM users WHERE id = $1"
	err := r.db.QueryRow(ctx, query, userID).Scan(&enabled)
	if err != nil {
		return false, utils.ConvertError(err)
	}

	return enabled, nil
}

func (r *UserRepository) generateEmail2FACode(ctx context.Context, userID string) (string, *utils.Error) {
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	// Get user email to send the code
	query := "SELECT email FROM users WHERE id = $1"
	var userEmail string
	err := r.db.QueryRow(ctx, query, userID).Scan(&userEmail)
	if err != nil {
		return "", utils.ConvertError(err)
	}

	expiration := time.Now().Add(5 * time.Minute)

	query = "UPDATE users SET email_2fa_code = $1, email_2fa_expires = $2 WHERE id = $3"
	_, err = r.db.Exec(ctx, query, code, expiration, userID)
	if err != nil {
		return "", utils.ConvertError(err)
	}

	// Send the 2FA code via email
	err2 := r.sendEmail2FACode(userEmail, code)
	if err2 != nil {
		return "", utils.NewInternal("failed to send 2FA code via email", err2)
	}

	return code, nil
}

func (r *UserRepository) sendEmail2FACode(userEmail, code string) error {
	emailConfig := utils.GetEmailConfigFromEnv()

	subject := "Your 2FA Code for Cinema Management System"
	body := fmt.Sprintf("Your 2FA code is: %s. This code is valid for 5 minutes.", code)

	sender := utils.NewSMTPSender(emailConfig)

	err := sender.SendEmail(userEmail, subject, body)
	if err != nil {
		fmt.Printf("Failed to send 2FA email to %s: %v\n", userEmail, err)
		return err
	}

	return nil
}

func (r *UserRepository) Verify2FACode(ctx context.Context, userID string, code string) (string, *utils.Error) {
	var emailCode *string
	var emailExpires *time.Time

	query := "SELECT email_2fa_code, email_2fa_expires FROM users WHERE id = $1 AND two_fa_enabled = true"
	err := r.db.QueryRow(ctx, query, userID).Scan(&emailCode, &emailExpires)
	if err != nil {
		return "", utils.ConvertError(err)
	}

	actualCode := ""
	if emailCode != nil {
		actualCode = *emailCode
	}

	if actualCode != code {
		return "", utils.NewForbidden("Invalid 2FA code", nil)
	}
	if emailExpires == nil || time.Now().After(*emailExpires) {
		return "", utils.NewForbidden("2FA code has expired", nil)
	}

	clearQuery := "UPDATE users SET email_2fa_code = NULL, email_2fa_expires = NULL WHERE id = $1"
	_, err = r.db.Exec(ctx, clearQuery, userID)
	if err != nil {
		return "", utils.ConvertError(err)
	}

	var isAdmin bool
	adminQuery := "SELECT is_admin FROM users WHERE id = $1"
	err = r.db.QueryRow(ctx, adminQuery, userID).Scan(&isAdmin)
	if err != nil {
		return "", utils.ConvertError(err)
	}

	token, err := r.generateJWTToken(userID, isAdmin)
	if err != nil {
		return "", utils.NewInternal("failed to generate token", err)
	}

	return token, nil
}
