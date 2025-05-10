package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// @Summary Получить всех пользователей
// @Tags users
// @Produce json
// @Success 200 {array} User
// @Failure 500 {string} string "Ошибка сервера"
// @Router /users [get]
func GetUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name, email, birth_date, is_blocked FROM users")
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
// @Failure 404 {string} string "Пользователь не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /users/{id} [get]
func GetUserByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		var u User
		err := db.QueryRow("SELECT id, name, email, birth_date, is_blocked FROM users WHERE id = $1", id).
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

// @Summary Создать пользователя
// @Tags users
// @Accept json
// @Produce json
// @Param user body User true "Пользователь"
// @Success 201 {object} User
// @Failure 400 {string} string "Неверный запрос"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /users [post]
func CreateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		u.ID = uuid.New().String()

		_, err := db.Exec("INSERT INTO users (id, name, email, password_hash, birth_date, is_blocked) VALUES ($1, $2, $3, $4, $5, $6)",
			u.ID, u.Name, u.Email, u.PasswordHash, u.BirthDate, u.IsBlocked)
		if err != nil {
			http.Error(w, "Ошибка при вставке", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(u)
	}
}

// // @Summary Регистрация пользователя
// // @Tags users
// // @Accept json
// // @Produce json
// // @Param user body User true "Пользователь"
// // @Success 201 {object} User
// // @Failure 400 {string} string "Неверный запрос"
// // @Failure 500 {string} string "Ошибка сервера"
// // @Router /register [post]
// func RegisterUser(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		var u User
// 		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
// 			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
// 			return
// 		}

// 		// Проверка на существование пользователя с таким email
// 		var existingUser User
// 		err := db.QueryRow("SELECT id FROM users WHERE email = $1", u.Email).Scan(&existingUser.ID)
// 		if err != sql.ErrNoRows {
// 			http.Error(w, "Пользователь с таким email уже существует", http.StatusBadRequest)
// 			return
// 		}

// 		// Хэширование пароля
// 		hash, err := bcrypt.GenerateFromPassword([]byte(u.PasswordHash), bcrypt.DefaultCost)
// 		if err != nil {
// 			http.Error(w, "Ошибка при хэшировании пароля", http.StatusInternalServerError)
// 			return
// 		}
// 		u.PasswordHash = string(hash)
// 		u.ID = uuid.New().String()

// 		_, err = db.Exec("INSERT INTO users (id, role_id, name, email, password_hash, birth_date, is_blocked) VALUES ($1, $2, $3, $4, $5, $6, $7)",
// 			u.ID, u.RoleID, u.Name, u.Email, u.PasswordHash, u.BirthDate, u.IsBlocked)
// 		if err != nil {
// 			http.Error(w, "Ошибка при вставке", http.StatusInternalServerError)
// 			return
// 		}
// 		w.WriteHeader(http.StatusCreated)
// 		json.NewEncoder(w).Encode(u)
// 	}
// }

// // @Summary Логин пользователя
// // @Tags users
// // @Accept json
// // @Produce json
// // @Param credentials body struct{ Email, Password string } true "Учетные данные"
// // @Success 200 {string} string "Успешный логин"
// // @Failure 400 {string} string "Неверный запрос"
// // @Failure 401 {string} string "Неверный email или пароль"
// // @Failure 500 {string} string "Ошибка сервера"
// // @Router /login [post]
// func LoginUser(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		var credentials struct {
// 			Email    string `json:"email"`
// 			Password string `json:"password"`
// 		}
// 		if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
// 			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
// 			return
// 		}

// 		var u User
// 		err := db.QueryRow("SELECT id, password_hash FROM users WHERE email = $1", credentials.Email).
// 			Scan(&u.ID, &u.PasswordHash)

// 		if err == sql.ErrNoRows {
// 			http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
// 			return
// 		} else if err != nil {
// 			http.Error(w, "Ошибка при получении пользователя", http.StatusInternalServerError)
// 			return
// 		}

// 		// Проверка пароля
// 		if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(credentials.Password)); err != nil {
// 			http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
// 			return
// 		}

// 		// Успешный логин
// 		w.WriteHeader(http.StatusOK)
// 		json.NewEncoder(w).Encode("Успешный логин")
// 	}
// }

// @Summary Обновить пользователя
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Param user body User true "Обновлённые данные пользователя"
// @Success 200 {object} User
// @Failure 400 {string} string "Неверный JSON"
// @Failure 404 {string} string "Пользователь не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /users/{id} [put]
func UpdateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		var u User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		u.ID = id

		res, err := db.Exec("UPDATE users SET name=$1, email=$2, birth_date=$3, is_blocked=$4 WHERE id=$5",
			u.Name, u.Email, u.BirthDate, u.IsBlocked, u.ID)
		if err != nil {
			http.Error(w, "Ошибка при обновлении", http.StatusInternalServerError)
			return
		}
		rows, _ := res.RowsAffected()
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
// @Failure 404 {string} string "Пользователь не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /users/{id} [delete]
func DeleteUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		res, err := db.Exec("DELETE FROM users WHERE id = $1", id)
		if err != nil {
			http.Error(w, fmt.Sprintf("ошибка при удалении %v", err), http.StatusInternalServerError)
			return
		}
		rows, _ := res.RowsAffected()
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
// @Param user_register body UserRegister true "Пользователь"
// @Success 201 {string} string "Пользователь зарегистрирован"
// @Failure 400 {string} string "Неверный запрос"
// @Failure 500 {string} string "Ошибка сохранения пользователя"
// @Router /register [post]
func RegisterUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user UserRegister
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, fmt.Sprintf("неверный запрос: %v", err), http.StatusBadRequest)
			return
		}

		var err error
		_, err = db.Exec("INSERT INTO users (email, password_hash, birth_date, name) VALUES ($1, $2, $3, $4)",
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
// @Param user_login body UserLogin true "Пользователь"
// @Success 200 {string} string "Токен авторизации"
// @Failure 400 {string} string "Неверный запрос"
// @Failure 401 {string} string "Неверное имя пользователя или пароль"
// @Failure 500 {string} string "Ошибка генерации токена"
// @Router /login [post]
func LoginUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user UserLogin
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, fmt.Sprintf("неверный запрос %v", err), http.StatusBadRequest)
			return
		}

		var passwordHash string
		err := db.QueryRow("SELECT password_hash FROM users WHERE email = $1", user.Email).Scan(&passwordHash)
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
