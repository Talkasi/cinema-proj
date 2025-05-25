package handler

import (
	"context"
	"cw/internal/service"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(us service.UserService) *UserHandler {
	return &UserHandler{userService: us}
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
func (u *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	seatType, err := u.userService.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w).Encode(seatType)
}

// @Summary Получить пользователя по ID (user* | admin)
// @Description Возвращает пользователя по ID.
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
			"SELECT id, name, email, birth_date, password_hash FROM users WHERE id = $1", id).
			Scan(&u.ID, &u.Name, &u.Email, &birthDate, &u.PasswordHash)
		u.BirthDate = birthDate.Format("2006-01-02")

		if IsError(w, err) {
			return
		}

		json.NewEncoder(w).Encode(u)
	}
}

// @Summary Обновить пользователя (user* | admin)
// @Description Обновляет данные пользователя.
// @Tags Пользователи
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID пользователя"
// @Param user body UserData true "Новые данные пользователя"
// @Success 200 "Данные пользователя успешно обновлены"
// @Failure 400 {object} ErrorResponse "Некорректные данные"
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
			"UPDATE users SET name=$1, email=$2, birth_date=$3, password_hash=$4 WHERE id=$5",
			u.Name, u.Email, u.BirthDate, u.PasswordHash, id)

		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// @Summary Получить nickname по ID пользователя (guest | user | admin)
// @Description Возвращает nickname по ID пользователя.
// @Tags Пользователи
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Success 200 {string} Nickname "Nickname пользователя"
// @Failure 400 {object} ErrorResponse "Некорректные данные"
// @Failure 404 {object} ErrorResponse "Пользователь не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /user/{id} [get]
func GetUserNickname(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		var Nickname string
		err := db.QueryRow(context.Background(),
			"SELECT name FROM users WHERE id = $1", id).
			Scan(&Nickname)

		if IsError(w, err) {
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Nickname)
	}
}

// @Summary Изменить статус администратора для пользователя (admin)
// @Description Изменяет статус администратора для пользователя по ID.
// @Tags Пользователи
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID пользователя"
// @Param user body UserAdmin true "Статус администратора"
// @Success 200 "Статус администрации для пользователя успешно обновлён"
// @Failure 400 {object} ErrorResponse "Некорректные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Пользователь не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /user/admin-status/{id} [put]
func UpdateAdminStatusUser(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		var u UserAdmin
		if !DecodeJSONBody(w, r, &u) {
			return
		}

		role := r.Header.Get("Role")
		user_id := r.Header.Get("UserID")
		if role != os.Getenv("CLAIM_ROLE_ADMIN") || (role == os.Getenv("CLAIM_ROLE_ADMIN") && id.String() == user_id) {
			http.Error(w, "Доступ запрещён", http.StatusForbidden)
			return
		}

		res, err := db.Exec(context.Background(),
			"UPDATE users SET is_admin=$1 WHERE id=$2",
			u.IsAdmin, id)

		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// @Summary Получить статус администратора для пользователя (admin)
// @Description Возвращает статус администратора для пользователя по ID.
// @Tags Пользователи
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID пользователя"
// @Success 200 {object} UserAdmin "Статус администрации для пользователя"
// @Failure 400 {object} ErrorResponse "Некорректные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Пользователь не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /user/admin-status/{id} [get]
func GetAdminStatusUser(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		role := r.Header.Get("Role")
		if role != os.Getenv("CLAIM_ROLE_ADMIN") {
			http.Error(w, "Доступ запрещён", http.StatusForbidden)
			return
		}

		var u UserAdmin

		err := db.QueryRow(context.Background(),
			"SELECT is_admin FROM users WHERE id = $1", id).
			Scan(&u.IsAdmin)

		if IsError(w, err) {
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(u)
	}
}

// @Summary Удалить пользователя (admin)
// @Description Удаляет пользователя по ID.
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

		user_id := r.Header.Get("UserID")
		if id.String() == user_id {
			http.Error(w, "Доступ запрещён", http.StatusForbidden)
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
// @Failure 400 {object} ErrorResponse "Некорректные данные"
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
// @Failure 400 {object} ErrorResponse "Некорректные данные"
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
			IsAdmin      bool
		}

		err := db.QueryRow(context.Background(),
			"SELECT id, password_hash, is_admin FROM users WHERE email = $1", creds.Email).
			Scan(&user.ID, &user.PasswordHash, &user.IsAdmin)
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
