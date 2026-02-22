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

// @Summary Poluchit spisok polzovateley
// @Description Vozvraschaet paginirovannyy spisok polzovateley s filtratsiey (tolko dlya administratorov)
// @Tags Polzovateli
// @Produce json
// @Security BearerAuth
// @Param page query int false "Nomer stranitsy" default(1) minimum(1)
// @Param limit query int false "Kolichestvo elementov na stranitse" default(20) minimum(1) maximum(100)
// @Param name query string false "Filtr po imeni (registronezavisimyy poisk)"
// @Param email query string false "Filtr po email (registronezavisimyy poisk)"
// @Param is_admin query boolean false "Filtr po statusu administratora"
// @Success 200 {object} dto.PaginatedUserResponse "Spisok polzovateley"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
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

// @Summary Poluchit polzovatelya po ID
// @Description Vozvraschaet informatsiyu o polzovatele (dostupno samomu polzovatelyu ili administratoru)
// @Tags Polzovateli
// @Produce json
// @Security BearerAuth
// @Param id path string true "UUID polzovatelya"
// @Success 200 {object} dto.UserResponse "Informatsiya o polzovatele"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
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

// @Summary Obnovit polzovatelya
// @Description Obnovlyaet informatsiyu o polzovatele (dostupno samomu polzovatelyu ili administratoru)
// @Tags Polzovateli
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "UUID polzovatelya"
// @Param user body dto.UpdateUserRequest true "Novye dannye polzovatelya"
// @Success 200 "Polzovatel obnovlen"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
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

// @Summary Izmenit status administratora
// @Description Izmenyaet status administratora polzovatelya (tolko dlya administratorov)
// @Tags Polzovateli
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "UUID polzovatelya"
// @Param admin body dto.UpdateAdminStatusRequest true "Status administratora"
// @Success 200 "Status administratora obnovlen"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
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

// @Summary Udalit polzovatelya
// @Description Udalyaet polzovatelya (tolko dlya administratorov)
// @Tags Polzovateli
// @Security BearerAuth
// @Param id path string true "UUID polzovatelya"
// @Success 204 "Polzovatel udalen"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
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

// @Summary Registratsiya novogo polzovatelya
// @Description Registriruet novogo polzovatelya v sisteme
// @Tags Autentifikatsiya
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "Dannye dlya registratsii"
// @Success 201 {object} dto.RegisterResponse "ID sozdannogo polzovatelya"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 409 {object} dto.ErrorResponse "Konflikt dannykh"
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

// @Summary Autentifikatsiya polzovatelya
// @Description Autentifitsiruet polzovatelya i vozvraschaet JWT-token
// @Tags Autentifikatsiya
// @Accept json
// @Produce json
// @Param credentials body dto.LoginRequest true "Dannye dlya vkhoda"
// @Success 200 {object} dto.AuthResponse "Uspeshnaya autentifikatsiya"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 401 {object} dto.ErrorResponse "Neavtorizovannyy dostup"
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

// @Summary Vklyuchit dvukhfaktornuyu autentifikatsiyu
// @Description Vklyuchaet dvukhfaktornuyu autentifikatsiyu dlya tekuschego polzovatelya
// @Tags Autentifikatsiya
// @Produce json
// @Security BearerAuth
// @Success 200 "2FA vklyuchena"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
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

// @Summary Vyklyuchit dvukhfaktornuyu autentifikatsiyu
// @Description Vyklyuchaet dvukhfaktornuyu autentifikatsiyu dlya tekuschego polzovatelya
// @Tags Autentifikatsiya
// @Produce json
// @Security BearerAuth
// @Success 200 "2FA vyklyuchena"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
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

// @Summary Poluchit informatsiyu o 2FA
// @Description Poluchaet informatsiyu o statuse dvukhfaktornoy autentifikatsii tekuschego polzovatelya
// @Tags Autentifikatsiya
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.TwoFAInfoResponse "Informatsiya o 2FA"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
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

// @Summary Podtverdit kod 2FA
// @Description Podtverzhdaet kod dvukhfaktornoy autentifikatsii i vozvraschaet token
// @Tags Autentifikatsiya
// @Accept json
// @Produce json
// @Param code body dto.Verify2FARequest true "Kod 2FA"
// @Success 200 {object} dto.AuthResponse "Token autentifikatsii"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
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

// @Summary Izmenit parol
// @Description Pozvolyaet polzovatelyu izmenit svoy parol
// @Tags Polzovateli
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param password body dto.UpdatePasswordRequest true "Tekuschiy i novyy paroli"
// @Success 200 "Parol uspeshno izmenen"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
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
		utils.WriteError(w, utils.NewForbidden("Tekuschiy parol neveren", nil))
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
