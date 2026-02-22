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

type ScreenTypeHandler struct {
	screenTypeService *service.ScreenTypeService
}

func NewScreenTypeHandler(sts *service.ScreenTypeService) *ScreenTypeHandler {
	return &ScreenTypeHandler{screenTypeService: sts}
}

// @Summary Get a list of screen types
// @Description Returns a paginated list of all screen types with filtering
// @Tags Screen Types
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Param name query string false "Search by screen type name (case-insensitive substring search)"
// @Param description query string false "Search by screen type description (case-insensitive substring search)"
// @Success 200 {object} dto.PaginatedScreenTypeResponse "Screen type list"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
// @Router /screen-types [get]
func (st *ScreenTypeHandler) GetScreenTypes(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	dtoFilters := dto.ScreenTypeFilters{
		Name:        r.URL.Query().Get("name"),
		Description: r.URL.Query().Get("description"),
	}

	domainFilters := domain.ScreenTypeFiltersFromDTO(dtoFilters)

	result, err := st.screenTypeService.GetAll(r.Context(), domainFilters, page, limit)
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

// @Summary Get a screen type by ID
// @Description Returns information about a screen type by its identifier
// @Tags Screen Types
// @Produce json
// @Param id path string true "Screen type UUID"
// @Success 200 {object} dto.ScreenTypeResponse "Screen type information"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
// @Router /screen-types/{id} [get]
func (st *ScreenTypeHandler) GetScreenTypeByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	screenType, err := st.screenTypeService.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(screenType); err != nil {
		utils.WriteError(w, utils.NewInternal("failed to encode JSON", err))
		return
	}
}

// @Summary Create a screen type
// @Description Creates a new screen type (administrators only)
// @Tags Screen Types
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param screen_type body dto.CreateScreenTypeRequest true "Screen type data"
// @Success 201 {object} dto.CreateResponse "Screen type created"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Router /screen-types [post]
func (st *ScreenTypeHandler) CreateScreenType(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateScreenTypeRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	domainScreenType := domain.CreateScreenTypeFromDTO(req)

	result, err := st.screenTypeService.Create(r.Context(), domainScreenType)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(dto.CreateResponse{ID: result.ID}); err != nil {
		utils.WriteError(w, utils.NewInternal("failed to encode JSON", err))
		return
	}
}

// @Summary Update a screen type
// @Description Fully updates information about a screen type (administrators only)
// @Tags Screen Types
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Screen type UUID"
// @Param screen_type body dto.UpdateScreenTypeRequest true "New screen type data"
// @Success 200 "Screen type updated"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
// @Router /screen-types/{id} [put]
func (st *ScreenTypeHandler) UpdateScreenType(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateScreenTypeRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	domainScreenType := domain.UpdateScreenTypeFromDTO(req)

	_, err := st.screenTypeService.Update(r.Context(), id, domainScreenType)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Delete a screen type
// @Description Deletes a screen type by identifier (administrators only)
// @Tags Screen Types
// @Security BearerAuth
// @Param id path string true "Screen type UUID"
// @Success 204 "Screen type deleted"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
// @Router /screen-types/{id} [delete]
func (st *ScreenTypeHandler) DeleteScreenType(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := st.screenTypeService.Delete(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (st *ScreenTypeHandler) RegisterRoutes(r chi.Router) {
	r.Route("/screen-types", func(r chi.Router) {
		r.Get("/", st.GetScreenTypes)
		r.Post("/", st.CreateScreenType)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", st.GetScreenTypeByID)
			r.Put("/", st.UpdateScreenType)
			r.Delete("/", st.DeleteScreenType)
		})
	})
}
