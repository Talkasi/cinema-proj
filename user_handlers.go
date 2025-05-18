package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func validateAllUserData(w http.ResponseWriter, u UserData) bool {
	u.Name = PrepareString(u.Name)
	u.Email = PrepareString(u.Email)
	u.BirthDate = PrepareString(u.BirthDate)

	if err := validateUserName(u.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	if err := validateUserEmail(u.Email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	if err := validateUserBirthDate(u.BirthDate); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	return true
}

func validateUserName(name string) error {
	validNameRegex := regexp.MustCompile(`^[A-Za-zА-Яа-яЁё\s-]+$`)
	if !validNameRegex.MatchString(name) {
		return errors.New("имя пользователя может содержать только буквы, пробелы и дефисы")
	}

	if !regexp.MustCompile(`\S`).MatchString(name) {
		return errors.New("имя пользователя не может состоять только из пробелов")
	}

	if len(name) == 0 || len(name) > 50 {
		return errors.New("имя пользователя не может быть пустым и не может превышать 50 символов")
	}
	return nil
}

func validateUserEmail(email string) error {
	validEmailRegex := regexp.MustCompile(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}$`)
	if !validEmailRegex.MatchString(email) {
		return errors.New("неверный формат email")
	}

	if len(email) > 100 {
		return errors.New("email не может превышать 100 символов")
	}
	return nil
}

func validateUserBirthDate(birthDate string) error {
	parsedDate, err := time.Parse("2006-01-02", birthDate)
	if err != nil {
		return errors.New("неверный формат даты, используйте YYYY-MM-DD")
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	minBirthDate := today.AddDate(-100, 0, 0)

	if parsedDate.After(today) {
		return errors.New("дата рождения не может быть в будущем")
	}

	if parsedDate.Before(minBirthDate) {
		return errors.New("дата рождения не может быть более 100 лет назад")
	}

	return nil
}

func validateUserPassword(password string) error {
	if len(password) < 8 {
		return errors.New("пароль должен содержать не менее 8 символов")
	}
	return nil
}

// @Summary Получить всех пользователей (admin)
// @Description Возвращает список всех пользователей.
// @Tags Пользователи
// @Produce json
// @Security BearerAuth
// @Success 200 {array} User "Список пользователей"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Пользователи не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /users [get]
func GetUsers(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		role := r.Header.Get("Role")
		println("Role", role, os.Getenv("CLAIM_ROLE_USER"), os.Getenv("CLAIM_ROLE_ADMIN"))

		if role != os.Getenv("CLAIM_ROLE_ADMIN") {
			http.Error(w, "Доступ запрещён", http.StatusForbidden)
			return
		}

		rows, err := db.Query(context.Background(),
			"SELECT id, name, email, birth_date, is_blocked, is_admin FROM users")
		if HandleDatabaseError(w, err, "пользователями") {
			return
		}
		defer rows.Close()

		var users []User
		var birthDate time.Time
		for rows.Next() {
			var u User
			if err := rows.Scan(&u.ID, &u.Name, &u.Email, &birthDate, &u.IsBlocked, &u.IsAdmin); HandleDatabaseError(w, err, "пользователем") {
				return
			}
			u.BirthDate = birthDate.Format("2006-01-02")
			users = append(users, u)
		}

		if len(users) == 0 {
			http.Error(w, "Пользователи не найдены", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(users)
	}
}

// @Summary Получить пользователя по ID (user* | admin)
// @Description Возвращает пользователя по ID
// @Tags Пользователи
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID пользователя"
// @Success 200 {object} User "Пользователь"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 404 {object} ErrorResponse "Пользователь не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /users/{id} [get]
func GetUserByID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		role := r.Header.Get("Role")
		user_id := r.Header.Get("UserID")
		if (role != os.Getenv("CLAIM_ROLE_ADMIN")) && (id.String() != user_id) {
			http.Error(w, "Доступ запрещён", http.StatusForbidden)
			return
		}

		var u User
		var birthDate time.Time
		err := db.QueryRow(context.Background(),
			"SELECT id, name, email, birth_date, is_blocked, is_admin FROM users WHERE id = $1", id).
			Scan(&u.ID, &u.Name, &u.Email, &birthDate, &u.IsBlocked, &u.IsAdmin)
		u.BirthDate = birthDate.Format("2006-01-02")

		if IsError(w, err) {
			return
		}

		json.NewEncoder(w).Encode(u)
	}
}

// @Summary Обновить пользователя (user* | admin)
// @Description Обновляет данные пользователя
// @Tags Пользователи
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID пользователя"
// @Param user body UserData true "Новые данные пользователя"
// @Success 200 "Данные пользователя успешно обновлены"
// @Failure 400 {object} ErrorResponse "В запросе предоставлены неверные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Пользователь не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /users/{id} [put]
func UpdateUser(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		role := r.Header.Get("Role")
		user_id := r.Header.Get("UserID")
		if (role != os.Getenv("CLAIM_ROLE_ADMIN")) && (id.String() != user_id) {
			http.Error(w, "Доступ запрещён", http.StatusForbidden)
			return
		}

		var u UserData
		if !DecodeJSONBody(w, r, &u) {
			return
		}

		if !validateAllUserData(w, u) {
			return
		}

		res, err := db.Exec(context.Background(),
			"UPDATE users SET name=$1, email=$2, birth_date=$3 WHERE id=$4",
			u.Name, u.Email, u.BirthDate, id)

		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// @Summary Удалить пользователя (admin)
// @Description Удаляет пользователя по ID
// @Tags Пользователи
// @Param id path string true "ID пользователя"
// @Security BearerAuth
// @Success 204 "Пользователь успешно удалён"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Пользователь не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /users/{id} [delete]
func DeleteUser(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		res, err := db.Exec(context.Background(),
			"DELETE FROM users WHERE id = $1", id)

		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// @Summary Зарегистрировать нового пользователя (guest | user | admin)
// @Description Регистрирует нового пользователя в системе.
// @Tags Пользователи
// @Accept json
// @Produce json
// @Param user body UserRegister true "Данные для регистрации"
// @Success 201 "Пользователь успешно зарегистрирован в системе"
// @Failure 400 {object} ErrorResponse "В запросе предоставлены неверные данные"
// @Failure 409 {object} ErrorResponse "Пользователь с таким email уже существует"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /user/register [post]
func RegisterUser(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user UserRegister
		if !DecodeJSONBody(w, r, &user) {
			return
		}

		user.Name = PrepareString(user.Name)
		user.Email = PrepareString(user.Email)
		user.BirthDate = PrepareString(user.BirthDate)

		if err := validateUserName(user.Name); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := validateUserEmail(user.Email); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := validateUserPassword(user.PasswordHash); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := validateUserBirthDate(user.BirthDate); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		id := uuid.New()
		_, err := db.Exec(context.Background(),
			"INSERT INTO users (id, name, email, password_hash, birth_date) VALUES ($1, $2, $3, $4, $5)",
			id, user.Name, user.Email, user.PasswordHash, user.BirthDate)

		if IsError(w, err) {
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w)
	}
}

// @Summary Вход пользователя (guest | user | admin)
// @Description Аутентифицирует пользователя и возвращает JWT-токен.
// @Tags Пользователи
// @Accept json
// @Produce json
// @Param credentials body UserLogin true "Данные для входа"
// @Success 200 {object} AuthResponse "Данные авторизации"
// @Failure 400 {object} ErrorResponse "В запросе предоставлены неверные данные"
// @Failure 401 {object} ErrorResponse "Неверный email или пароль"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /user/login [post]
func LoginUser(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds UserLogin
		if !DecodeJSONBody(w, r, &creds) {
			return
		}

		creds.Email = PrepareString(creds.Email)

		if err := validateUserEmail(creds.Email); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := validateUserPassword(creds.PasswordHash); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var user struct {
			ID           string
			PasswordHash string
			IsBlocked    bool
			IsAdmin      bool
		}

		err := db.QueryRow(context.Background(),
			"SELECT id, password_hash, is_blocked, is_admin FROM users WHERE email = $1", creds.Email).
			Scan(&user.ID, &user.PasswordHash, &user.IsBlocked, &user.IsAdmin)
		if isNoRows(err) {
			http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
			return
		}
		if IsError(w, err) {
			return
		}

		if creds.PasswordHash != user.PasswordHash {
			http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
			return
		}

		role := os.Getenv("CLAIM_ROLE_USER")
		if user.IsAdmin {
			role = os.Getenv("CLAIM_ROLE_ADMIN")
		}

		token, err := GenerateToken(role, user.ID)
		if err != nil {
			http.Error(w, "Ошибка генерации токена", http.StatusInternalServerError)
			return
		}

		var resp AuthResponse
		resp.Token = token
		resp.UserID = user.ID
		json.NewEncoder(w).Encode(resp)
	}
}
