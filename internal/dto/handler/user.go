package handler

import (
	domain "cw/internal/domain/models"
	"cw/internal/domain/service"
	dto "cw/internal/dto/models"
	"cw/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(us *service.UserService) *UserHandler {
	return &UserHandler{userService: us}
}

// @Summary Получить список пользователей
// @Description Возвращает пагинированный список пользователей с фильтрацией (только для администраторов)
// @Tags Пользователи
// @Produce json
// @Security BearerAuth
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Количество элементов на странице" default(20) minimum(1) maximum(100)
// @Param name query string false "Фильтр по имени (регистронезависимый поиск)"
// @Param email query string false "Фильтр по email (регистронезависимый поиск)"
// @Param is_admin query boolean false "Фильтр по статусу администратора"
// @Success 200 {object} dto.PaginatedUserResponse "Список пользователей"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Router /users [get]
func (uh *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	dtoFilters := dto.UserFilters{
		Name:  r.URL.Query().Get("name"),
		Email: r.URL.Query().Get("email"),
	}

	if isAdmin := r.URL.Query().Get("is_admin"); isAdmin != "" {
		if val, err := strconv.ParseBool(isAdmin); err == nil {
			dtoFilters.IsAdmin = &val
		}
	}

	domainFilters := domain.UserFiltersFromDTO(dtoFilters)

	result, err := uh.userService.GetAll(r.Context(), domainFilters, page, limit)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// @Summary Получить пользователя по ID
// @Description Возвращает информацию о пользователе (доступно самому пользователю или администратору)
// @Tags Пользователи
// @Produce json
// @Security BearerAuth
// @Param id path string true "UUID пользователя"
// @Success 200 {object} dto.UserResponse "Информация о пользователе"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /users/{id} [get]
func (uh *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := uh.userService.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// @Summary Обновить пользователя
// @Description Обновляет информацию о пользователе (доступно самому пользователю или администратору)
// @Tags Пользователи
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "UUID пользователя"
// @Param user body dto.UpdateUserRequest true "Новые данные пользователя"
// @Success 200 "Пользователь обновлен"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /users/{id} [put]
func (uh *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
		return
	}

	domainUser := domain.UpdateUserFromDTO(req)

	_, err := uh.userService.Update(r.Context(), id, domainUser)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Изменить статус администратора
// @Description Изменяет статус администратора пользователя (только для администраторов)
// @Tags Пользователи
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "UUID пользователя"
// @Param admin body dto.UpdateAdminStatusRequest true "Статус администратора"
// @Success 200 "Статус администратора обновлен"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /users/{id}/role [patch]
func (uh *UserHandler) UpdateAdminStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateAdminStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
		return
	}

	_, err := uh.userService.UpdateAdminStatus(r.Context(), id, req.IsAdmin)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Удалить пользователя
// @Description Удаляет пользователя (только для администраторов)
// @Tags Пользователи
// @Security BearerAuth
// @Param id path string true "UUID пользователя"
// @Success 204 "Пользователь удален"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /users/{id} [delete]
func (uh *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := uh.userService.Delete(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary Регистрация нового пользователя
// @Description Регистрирует нового пользователя в системе
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "Данные для регистрации"
// @Success 201 {object} dto.RegisterResponse "ID созданного пользователя"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 409 {object} dto.ErrorResponse "Конфликт данных"
// @Router /auth/register [post]
func (uh *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
		return
	}

	domainUser := domain.CreateUserFromDTO(req)

	result, err := uh.userService.Register(r.Context(), domainUser)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.RegisterResponse{ID: result.ID})
}

// @Summary Аутентификация пользователя
// @Description Аутентифицирует пользователя и возвращает JWT-токен
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param credentials body dto.LoginRequest true "Данные для входа"
// @Success 200 {object} dto.AuthResponse "Успешная аутентификация"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 401 {object} dto.ErrorResponse "Неавторизованный доступ"
// @Router /auth/login [post]
func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
		return
	}

	domainUser := domain.LoginFromDTO(req)

	result, err := uh.userService.Login(r.Context(), domainUser)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// @Summary Включить двухфакторную аутентификацию
// @Description Включает двухфакторную аутентификацию для текущего пользователя
// @Tags Аутентификация
// @Produce json
// @Security BearerAuth
// @Success 200 "2FA включена"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Router /auth/2fa/enable [post]
func (uh *UserHandler) Enable2FA(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	err := uh.userService.Enable2FA(r.Context(), userID)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Выключить двухфакторную аутентификацию
// @Description Выключает двухфакторную аутентификацию для текущего пользователя
// @Tags Аутентификация
// @Produce json
// @Security BearerAuth
// @Success 200 "2FA выключена"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Router /auth/2fa/disable [post]
func (uh *UserHandler) Disable2FA(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	err := uh.userService.Disable2FA(r.Context(), userID)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Получить информацию о 2FA
// @Description Получает информацию о статусе двухфакторной аутентификации текущего пользователя
// @Tags Аутентификация
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.TwoFAInfoResponse "Информация о 2FA"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Router /auth/2fa/info [get]
func (uh *UserHandler) Get2FAInfo(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	enabled, err := uh.userService.Get2FAInfo(r.Context(), userID)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	response := dto.TwoFAInfoResponse{
		Enabled: enabled,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// @Summary Подтвердить код 2FA
// @Description Подтверждает код двухфакторной аутентификации и возвращает токен
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param code body dto.Verify2FARequest true "Код 2FA"
// @Success 200 {object} dto.AuthResponse "Токен аутентификации"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Router /auth/verify [post]
func (uh *UserHandler) Verify2FA(w http.ResponseWriter, r *http.Request) {
	var req dto.Verify2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
		return
	}

	token, err := uh.userService.Verify2FACode(r.Context(), req.UserID, req.Code)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	response := dto.AuthResponse{
		Token:       token,
		UserID:      req.UserID,
		TwoFANeeded: false,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// @Summary Изменить пароль
// @Description Позволяет пользователю изменить свой пароль
// @Tags Пользователи
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param password body dto.UpdatePasswordRequest true "Текущий и новый пароли"
// @Success 200 "Пароль успешно изменен"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Router /users/password [patch]
func (uh *UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
		return
	}

	userID := r.Context().Value("userID").(string)

	err := uh.userService.VerifyCurrentPassword(r.Context(), userID, req.CurrentPassword)
	if err != nil {
		utils.WriteError(w, utils.NewForbidden("Текущий пароль неверен", nil))
		return
	}

	err = uh.userService.UpdatePassword(r.Context(), userID, req.NewPassword)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Password updated successfully")
}

func (uh *UserHandler) RegisterRoutesNoAuth(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", uh.Register)
		r.Post("/login", uh.Login)
		r.Post("/verify", uh.Verify2FA)
	})
}

func (uh *UserHandler) RegisterRoutesWithAuth(r chi.Router) {
	r.Route("/auth/2fa", func(r chi.Router) {
		r.Post("/enable", uh.Enable2FA)
		r.Post("/disable", uh.Disable2FA)
		r.Get("/info", uh.Get2FAInfo)
	})

	r.Route("/users", func(r chi.Router) {
		r.Get("/", uh.GetUsers)
		r.Patch("/password", uh.UpdatePassword)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", uh.GetUserByID)
			r.Put("/", uh.UpdateUser)
			r.Patch("/role", uh.UpdateAdminStatus)
			r.Delete("/", uh.DeleteUser)
		})
	})
}
