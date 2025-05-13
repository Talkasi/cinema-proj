package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// @Summary Получить всех пользователей
// @Tags users
// @Produce json
// @Success 200 {array} User
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /users [get]
func GetUsers(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(context.Background(), "SELECT id, name, email, birth_date, is_blocked FROM users")
		if err != nil {
			http.Error(w, "Ошибка при получении пользователей", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var users []User
		for rows.Next() {
			var u User
			if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.BirthDate, &u.IsBlocked); err != nil {
				http.Error(w, "Ошибка при сканировании", http.StatusInternalServerError)
				return
			}
			users = append(users, u)
		}
		json.NewEncoder(w).Encode(users)
	}
}

// @Summary Получить пользователя по ID
// @Tags users
// @Produce json
// @Param id path string true "ID пользователя"
// @Success 200 {object} User
// @Failure 404 {object} ErrorResponse "Пользователь не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /users/{id} [get]
func GetUserByID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		var u User
		err := db.QueryRow(context.Background(), "SELECT id, name, email, birth_date, is_blocked FROM users WHERE id = $1", id).
			Scan(&u.ID, &u.Name, &u.Email, &u.BirthDate, &u.IsBlocked)

		if err == sql.ErrNoRows {
			http.Error(w, "Пользователь не найден", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Ошибка при получении", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(u)
	}
}

// @Summary Обновить пользователя
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Param user body User true "Обновлённые данные пользователя"
// @Success 200 {object} User
// @Failure 400 {object} ErrorResponse "Неверный JSON"
// @Failure 404 {object} ErrorResponse "Пользователь не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /users/{id} [put]
func UpdateUser(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		var u User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		u.ID = id

		res, err := db.Exec(context.Background(), "UPDATE users SET name=$1, email=$2, birth_date=$3, is_blocked=$4 WHERE id=$5",
			u.Name, u.Email, u.BirthDate, u.IsBlocked, u.ID)
		if err != nil {
			http.Error(w, "Ошибка при обновлёнии", http.StatusInternalServerError)
			return
		}
		rows := res.RowsAffected()
		if rows == 0 {
			http.Error(w, "Пользователь не найден", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(u)
	}
}

// @Summary Удалить пользователя
// @Tags users
// @Param id path string true "ID пользователя"
// @Success 204 {string} string "Удалено"
// @Failure 404 {object} ErrorResponse "Пользователь не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /users/{id} [delete]
func DeleteUser(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		res, err := db.Exec(context.Background(), "DELETE FROM users WHERE id = $1", id)
		if err != nil {
			http.Error(w, fmt.Sprintf("ошибка при удалёнии %v", err), http.StatusInternalServerError)
			return
		}
		rows := res.RowsAffected()
		if rows == 0 {
			http.Error(w, "Пользователь не найден", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// RegisterUser регистрирует нового пользователя
// @Summary Регистрация нового пользователя
// @Description Регистрация нового пользователя с именем пользователя и паролем
// @Accept json
// @Produce json
// @Tags users
// @Param user_register body UserRegister true "Пользователь"
// @Success 201 {string} string "Пользователь зарегистрирован"
// @Failure 400 {object} ErrorResponse "Неверный запрос"
// @Failure 500 {object} ErrorResponse "Ошибка сохранения пользователя"
// @Router /register [post]
func RegisterUser(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user UserRegister
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, fmt.Sprintf("неверный запрос: %v", err), http.StatusBadRequest)
			return
		}

		var err error
		_, err = db.Exec(context.Background(), "INSERT INTO users (email, password_hash, birth_date, name) VALUES ($1, $2, $3, $4)",
			user.Email, user.PasswordHash, user.BirthDate, user.Name)
		if err != nil {
			http.Error(w, fmt.Sprintf("ошибка сохранения пользователя: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

// LoginUser позволяет пользователю войти в систему
// @Summary Логин пользователя
// @Description Логин пользователя с именем пользователя и паролем
// @Accept json
// @Produce json
// @Tags users
// @Param user_login body UserLogin true "Пользователь"
// @Success 200 {string} string "Токен авторизации"
// @Failure 400 {object} ErrorResponse "Неверный запрос"
// @Failure 401 {object} ErrorResponse "Неверное имя пользователя или пароль"
// @Failure 500 {object} ErrorResponse "Ошибка генерации токена"
// @Router /login [post]
func LoginUser(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user UserLogin
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, fmt.Sprintf("неверный запрос %v", err), http.StatusBadRequest)
			return
		}

		var passwordHash string
		err := db.QueryRow(context.Background(), "SELECT password_hash FROM users WHERE email = $1", user.Email).Scan(&passwordHash)
		if err != nil {
			http.Error(w, fmt.Sprintf("неверное имя пользователя или пароль %v", err), http.StatusUnauthorized)
			return
		}

		if passwordHash != user.PasswordHash {
			http.Error(w, "неверное имя пользователя или пароль", http.StatusUnauthorized)
			return
		}

		token, err := GenerateToken(user.Email, "user")
		if err != nil {
			http.Error(w, fmt.Sprintf("ошибка генерации токена %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Authorization", token)
		w.WriteHeader(http.StatusOK)
	}
}
