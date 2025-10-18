package handler

import (
	domain "cw/internal/domain/models"
	"cw/internal/domain/service"
	dto "cw/internal/dto/models"
	"cw/internal/utils"
	"encoding/json"
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

func (uh *UserHandler) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", uh.Register)
		r.Post("/login", uh.Login)
	})

	r.Route("/users", func(r chi.Router) {
		r.Get("/", uh.GetUsers)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", uh.GetUserByID)
			r.Put("/", uh.UpdateUser)
			r.Patch("/role", uh.UpdateAdminStatus)
			r.Delete("/", uh.DeleteUser)
		})
	})
}
