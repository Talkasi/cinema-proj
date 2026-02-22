package handler

import (
	domain "cw/internal/domain/models"
	"cw/internal/domain/service"
	dto "cw/internal/dto/models"
	"cw/internal/middleware"
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

// @Summary Get a list of users
// @Description Returns a paginated list of users with filtering (administrators only)
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Param name query string false "Filter by name (case-insensitive search)"
// @Param email query string false "Filter by email (case-insensitive search)"
// @Param is_admin query boolean false "Filter by administrator status"
// @Success 200 {object} dto.PaginatedUserResponse "User list"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
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
	if err := json.NewEncoder(w).Encode(result); err != nil {
		utils.WriteError(w, utils.NewInternal("failed to encode JSON", err))
		return
	}
}

// @Summary Get a user by ID
// @Description Returns information about a user (available to the user themself or an administrator)
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Success 200 {object} dto.UserResponse "User information"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
// @Router /users/{id} [get]
func (uh *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := uh.userService.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		utils.WriteError(w, utils.NewInternal("failed to encode JSON", err))
		return
	}
}

// @Summary Update a user
// @Description Updates information about a user (available to the user themself or an administrator)
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Param user body dto.UpdateUserRequest true "New user data"
// @Success 200 "User updated"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
// @Router /users/{id} [put]
func (uh *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateUserRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	domainUser := domain.UpdateUserFromDTO(req)

	_, err := uh.userService.Update(r.Context(), id, domainUser)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Change administrator status
// @Description Changes a user's administrator status (administrators only)
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Param admin body dto.UpdateAdminStatusRequest true "Administrator status"
// @Success 200 "Administrator status updated"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
// @Router /users/{id}/role [patch]
func (uh *UserHandler) UpdateAdminStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateAdminStatusRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	_, err := uh.userService.UpdateAdminStatus(r.Context(), id, req.IsAdmin)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Delete a user
// @Description Deletes a user (administrators only)
// @Tags Users
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Success 204 "User deleted"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
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

// @Summary Register a new user
// @Description Registers a new user in the system
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "Registration data"
// @Success 201 {object} dto.RegisterResponse "ID of the created user"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 409 {object} dto.ErrorResponse "Data conflict"
// @Router /auth/register [post]
func (uh *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
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
	if err := json.NewEncoder(w).Encode(dto.RegisterResponse{ID: result.ID}); err != nil {
		utils.WriteError(w, utils.NewInternal("failed to encode JSON", err))
		return
	}
}

// @Summary Authenticate a user
// @Description Authenticates a user and returns a JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.AuthResponse "Successful authentication"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized access"
// @Router /auth/login [post]
func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	domainUser := domain.LoginFromDTO(req)

	result, err := uh.userService.Login(r.Context(), domainUser)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		utils.WriteError(w, utils.NewInternal("failed to encode JSON", err))
		return
	}
}

// @Summary Enable two-factor authentication
// @Description Enables two-factor authentication for the current user
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 "2FA enabled"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Router /auth/2fa/enable [post]
func (uh *UserHandler) Enable2FA(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	err := uh.userService.Enable2FA(r.Context(), userID)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Disable two-factor authentication
// @Description Disables two-factor authentication for the current user
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 "2FA disabled"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Router /auth/2fa/disable [post]
func (uh *UserHandler) Disable2FA(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	err := uh.userService.Disable2FA(r.Context(), userID)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Get 2FA information
// @Description Gets information about the current user's two-factor authentication status
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.TwoFAInfoResponse "2FA information"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Router /auth/2fa/info [get]
func (uh *UserHandler) Get2FAInfo(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	enabled, err := uh.userService.Get2FAInfo(r.Context(), userID)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	response := dto.TwoFAInfoResponse{
		Enabled: enabled,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.WriteError(w, utils.NewInternal("failed to encode JSON", err))
		return
	}
}

// @Summary Verify a 2FA code
// @Description Verifies the two-factor authentication code and returns a token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param code body dto.Verify2FARequest true "2FA code"
// @Success 200 {object} dto.AuthResponse "Authentication token"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Router /auth/verify [post]
func (uh *UserHandler) Verify2FA(w http.ResponseWriter, r *http.Request) {
	var req dto.Verify2FARequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
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
	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.WriteError(w, utils.NewInternal("failed to encode JSON", err))
		return
	}
}

// @Summary Change a password
// @Description Allows a user to change their password
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param password body dto.UpdatePasswordRequest true "Current and new passwords"
// @Success 200 "Password updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Router /users/password [patch]
func (uh *UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdatePasswordRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)

	err := uh.userService.VerifyCurrentPassword(r.Context(), userID, req.CurrentPassword)
	if err != nil {
		utils.WriteError(w, utils.NewForbidden("Current password is incorrect", nil))
		return
	}

	err = uh.userService.UpdatePassword(r.Context(), userID, req.NewPassword)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, "Password updated successfully")
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
