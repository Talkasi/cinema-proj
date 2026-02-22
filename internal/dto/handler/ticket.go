package handler

import (
	domain "cw/internal/domain/models"
	"cw/internal/domain/service"
	dto "cw/internal/dto/models"
	"cw/internal/utils"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type TicketHandler struct {
	ticketService *service.TicketService
}

func NewTicketHandler(ts *service.TicketService) *TicketHandler {
	return &TicketHandler{ticketService: ts}
}

// @Summary Poluchit bilety
// @Description Vozvraschaet paginirovannyy spisok biletov s filtratsiey
// @Tags Bilety
// @Produce json
// @Param page query int false "Nomer stranitsy" default(1) minimum(1)
// @Param limit query int false "Kolichestvo elementov na stranitse" default(20) minimum(1) maximum(100)
// @Param ticket_Status query string false "Filtr po statusu bileta (mozhno ukazat neskolko cherez zapyatuyu)"
// @Param movie_show_id query string false "UUID seansa"
// @Param price_min query number false "Minimalnaya tsena bileta"
// @Param price_max query number false "Maksimalnaya tsena bileta"
// @Param seat_id query string false "Filtr po ID mesta (mozhno ukazat neskolko cherez zapyatuyu)"
// @Param user_id query string false "Filtr po ID polzovatelya (mozhno ukazat neskolko cherez zapyatuyu)"
// @Success 200 {object} dto.PaginatedTicketResponse "Spisok biletov"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
// @Router /tickets [get]
func (th *TicketHandler) GetTickets(w http.ResponseWriter, r *http.Request) {
	page, limit := parsePaginationParams(r)

	dtoFilters := dto.TicketFilters{
		Status:      splitCommaSeparated(r.URL.Query().Get("ticket_Status")),
		MovieShowID: r.URL.Query().Get("movie_show_id"),
		SeatID:      splitCommaSeparated(r.URL.Query().Get("seat_id")),
		UserID:      splitCommaSeparated(r.URL.Query().Get("user_id")),
	}

	if priceMin := r.URL.Query().Get("price_min"); priceMin != "" {
		if val, err := strconv.ParseFloat(priceMin, 64); err == nil {
			dtoFilters.PriceMin = val
		}
	}

	if priceMax := r.URL.Query().Get("price_max"); priceMax != "" {
		if val, err := strconv.ParseFloat(priceMax, 64); err == nil {
			dtoFilters.PriceMax = val
		}
	}

	domainFilters := domain.TicketFiltersFromDTO(dtoFilters)

	result, err := th.ticketService.GetAll(r.Context(), domainFilters, page, limit)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	if err := writeJSON(w, http.StatusOK, result); err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Poluchit bilet po ID
// @Description Vozvraschaet informatsiyu o bilete po ego identifikatoru
// @Tags Bilety
// @Produce json
// @Param id path string true "UUID bileta"
// @Success 200 {object} dto.TicketResponse "Informatsiya o bilete"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
// @Router /tickets/{id} [get]
func (th *TicketHandler) GetTicketByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ticket, err := th.ticketService.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	if err := writeJSON(w, http.StatusOK, ticket); err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Sozdat bilet na seans
// @Description Sozdaet novyy bilet dlya ukazannogo kinoseansa (tolko dlya avtorizovannykh polzovateley)
// @Tags Bilety
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movie_show_id path string true "UUID kinoseansa"
// @Param ticket body dto.CreateTicketRequest true "Dannye bileta"
// @Success 201 {object} dto.CreateResponse "Bilet sozdan"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Router /movie-shows/{movie_show_id}/tickets [post]
func (th *TicketHandler) CreateTicketForMovieShow(w http.ResponseWriter, r *http.Request) {
	movieShowID := chi.URLParam(r, "movie_show_id")

	var req dto.CreateTicketRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	domainTicket := domain.CreateTicketFromDTO(req)

	result, err := th.ticketService.CreateForMovieShow(r.Context(), movieShowID, domainTicket)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	if err := writeJSON(w, http.StatusCreated, dto.CreateResponse{ID: result.ID}); err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Obnovit status bileta
// @Description Obnovlyaet status bileta (bronirovanie/pokupka)
// @Tags Bilety
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "UUID bileta"
// @Param Status body dto.UpdateStatusRequest true "Dannye statusa bileta"
// @Success 200 "Status bileta obnovlen"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
// @Router /tickets/{id} [patch]
func (th *TicketHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateStatusRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	domainTicket := domain.UpdateStatusFromDTO(req)

	_, err := th.ticketService.UpdateStatus(r.Context(), id, domainTicket)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Udalit bilet
// @Description Udalyaet bilet po identifikatoru (tolko dlya administratorov ili vladeltsa)
// @Tags Bilety
// @Security BearerAuth
// @Param id path string true "UUID bileta"
// @Success 204 "Bilet udalen"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
// @Router /tickets/{id} [delete]
func (th *TicketHandler) DeleteTicket(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := th.ticketService.Delete(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (th *TicketHandler) RegisterRoutes(r chi.Router) {
	r.Route("/tickets", func(r chi.Router) {
		r.Get("/", th.GetTickets)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", th.GetTicketByID)
			r.Patch("/", th.UpdateStatus)
			r.Delete("/", th.DeleteTicket)
		})
	})

	r.Route("/movie-shows/{movie_show_id}/tickets", func(r chi.Router) {
		r.Post("/", th.CreateTicketForMovieShow)
	})
}
