package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "cw/docs"
)

var IsTestMode bool

var (
	TestGuestDB *pgxpool.Pool
	TestUserDB  *pgxpool.Pool
	TestAdminDB *pgxpool.Pool
)

func InitTestDB() error {
	var err error
	ctx := context.Background()

	TestGuestDB, err = pgxpool.New(ctx, "user=guest_test dbname=cinema_test password=guest111 sslmode=disable")
	if err != nil {
		return fmt.Errorf("ошибка подключения гостя: %v", err)
	}

	TestUserDB, err = pgxpool.New(ctx, "user=ruser_test dbname=cinema_test password=user111 sslmode=disable")
	if err != nil {
		return fmt.Errorf("ошибка подключения пользователя: %v", err)
	}

	TestAdminDB, err = pgxpool.New(ctx, "user=admin_test dbname=cinema_test password=admin555 sslmode=disable")
	if err != nil {
		return fmt.Errorf("ошибка подключения администратора: %v", err)
	}

	return nil
}

// @title Your API Title
// @version 1.0
// @description This is a sample
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	IsTestMode = false
	if err := InitDB(); err != nil {
		log.Fatal("ошибка подключения к БД: ", err)
	}

	defer AdminDB.Close()
	defer UserDB.Close()
	defer GuestDB.Close()

	if err := CreateAll(AdminDB); err != nil {
		log.Fatal("ошибка создания таблиц БД: ", err)
	}

	if err := SeedAll(AdminDB); err != nil {
		log.Fatal("ошибка вставки данных: ", err)
	}

	log.Println("Сервер запущен на http://localhost:8080")
	http.ListenAndServe(":8080", NewRouter())
}

type Claims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.StandardClaims
}

var SecretKey = []byte("secret_key")

var (
	GuestDB *pgxpool.Pool
	UserDB  *pgxpool.Pool
	AdminDB *pgxpool.Pool
)

func InitDB() error {
	var err error

	ctx := context.Background()

	GuestDB, err = pgxpool.New(ctx, "user=guest dbname=cinema password=guest111 sslmode=disable")
	if err != nil {
		return fmt.Errorf("ошибка подключения гостя: %v", err)
	}
	// defer GuestDB.Close()

	UserDB, err = pgxpool.New(ctx, "user=ruser dbname=cinema password=user111 sslmode=disable")
	if err != nil {
		return fmt.Errorf("ошибка подключения пользователя: %v", err)
	}
	// defer UserDB.Close()

	AdminDB, err = pgxpool.New(ctx, "user=admin dbname=cinema password=admin555 sslmode=disable")
	if err != nil {
		return fmt.Errorf("ошибка подключения администратора: %v", err)
	}
	// defer AdminDB.Close()

	return nil
}

func Midleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString != "" {
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return SecretKey, nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Неверный токен", http.StatusForbidden)
				return
			}

			r.Header.Set("Role", claims.Role)
		} else {
			r.Header.Set("Role", "guest")
		}

		next.ServeHTTP(w, r)
	}
}

func ClearTable(db *pgxpool.Pool, tableName string) error {
	query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", tableName)

	_, err := db.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to clear table %s: %w", tableName, err)
	}
	return nil
}

func GenerateToken(email, role string) (string, error) {
	claims := Claims{
		Email: email,
		Role:  role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(SecretKey)
}

func RoleBasedHandler(handler func(db *pgxpool.Pool) http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		role := r.Header.Get("Role")
		var db *pgxpool.Pool

		switch role {
		case "admin":
			db = AdminDB
			if IsTestMode {
				db = TestAdminDB
			}
		case "ruser":
			db = UserDB
			if IsTestMode {
				db = TestUserDB
			}
		default:
			db = GuestDB
			if IsTestMode {
				db = TestGuestDB
			}
		}

		if db == nil {
			http.Error(w, "Database connection not available", http.StatusInternalServerError)
			return
		}

		handler(db)(w, r)
	}
}

func ParseUUIDFromPath(w http.ResponseWriter, pathValue string) (uuid.UUID, bool) {
	id, err := uuid.Parse(pathValue)
	if err != nil {
		http.Error(w, "Неверный формат UUID", http.StatusBadRequest)
		return uuid.Nil, false
	}
	return id, true
}

func DecodeJSONBody(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return false
	}
	return true
}

func HandleDatabaseError(w http.ResponseWriter, err error, entity string) bool {
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при работе с %s: %v", entity, err), http.StatusInternalServerError)
		return true
	}
	return false
}

func CheckRowsAffected(w http.ResponseWriter, rowsAffected int64) bool {
	if rowsAffected == 0 {
		http.Error(w, "Данные не найдены", http.StatusNotFound)
		return false
	}
	return true
}

func ValidateRequiredFields(w http.ResponseWriter, fields map[string]string) bool {
	for field, value := range fields {
		if value == "" {
			http.Error(w, fmt.Sprintf("Поле '%s' не может быть пустым", field), http.StatusBadRequest)
			return false
		}
	}
	return true
}
