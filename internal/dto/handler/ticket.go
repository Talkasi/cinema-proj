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

// @Summary Get tickets
// @Description Returns a paginated list of tickets with filtering
// @Tags Tickets
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Param ticket_Status query string false "Filter by ticket status (multiple comma-separated values allowed)"
// @Param movie_show_id query string false "Movie show UUID"
// @Param price_min query number false "Minimum ticket price"
// @Param price_max query number false "Maximum ticket price"
// @Param seat_id query string false "Filter by seat ID (multiple comma-separated values allowed)"
// @Param user_id query string false "Filter by user ID (multiple comma-separated values allowed)"
// @Success 200 {object} dto.PaginatedTicketResponse "Ticket list"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
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

// @Summary Get a ticket by ID
// @Description Returns information about a ticket by its identifier
// @Tags Tickets
// @Produce json
// @Param id path string true "Ticket UUID"
// @Success 200 {object} dto.TicketResponse "Ticket information"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
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

// @Summary Create a ticket for a movie show
// @Description Creates a new ticket for the specified movie show (authorized users only)
// @Tags Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movie_show_id path string true "Movie show UUID"
// @Param ticket body dto.CreateTicketRequest true "Ticket data"
// @Success 201 {object} dto.CreateResponse "Ticket created"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
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

// @Summary Update ticket status
// @Description Updates ticket status (reservation/purchase)
// @Tags Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Ticket UUID"
// @Param Status body dto.UpdateStatusRequest true "Ticket status data"
// @Success 200 "Ticket status updated"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
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

// @Summary Delete a ticket
// @Description Deletes a ticket by identifier (administrators only or owner)
// @Tags Tickets
// @Security BearerAuth
// @Param id path string true "Ticket UUID"
// @Success 204 "Ticket deleted"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
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
