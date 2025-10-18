package postgres

import (
	"context"
	"fmt"
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

	query := "SELECT id, name, email, password_hash, birth_date, is_admin FROM users WHERE email = $1"
	err := r.db.QueryRow(ctx, query, credentials.Email).Scan(
		&userEntity.ID, &userEntity.Name, &userEntity.Email, &userEntity.PasswordHash, &userEntity.BirthDate, &userEntity.IsAdmin,
	)
	if err != nil {
		return domain.AuthResponse{}, utils.ConvertError(err)
	}

	if userEntity.PasswordHash != credentials.PasswordHash {
		return domain.AuthResponse{}, utils.NewForbidden("invalid credentials", nil)
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

	query := "UPDATE users SET name = $1, email = $2, birth_date = $3 WHERE id = $4 RETURNING id, name, email, password_hash, birth_date, is_admin"
	var updatedUser entity.User
	err := r.db.QueryRow(ctx, query, userEntity.Name, userEntity.Email, userEntity.BirthDate, id).Scan(
		&updatedUser.ID, &updatedUser.Name, &updatedUser.Email, &updatedUser.PasswordHash, &updatedUser.BirthDate, &updatedUser.IsAdmin,
	)
	if err != nil {
		return domain.User{}, utils.ConvertError(err)
	}

	return entity.UserToDomain(updatedUser), nil
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
		"user_id":  userID,
		"is_admin": isAdmin,
		"exp":      time.Now().Add(r.tokenDuration).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(r.jwtSecret))
}
