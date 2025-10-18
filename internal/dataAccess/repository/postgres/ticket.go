package postgres

import (
	"context"
	"fmt"
	"strings"

	entity "cw/internal/dataAccess/models"
	domain "cw/internal/domain/models"
	"cw/internal/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TicketRepository struct {
	db *pgxpool.Pool
}

func NewTicketRepository(db *pgxpool.Pool) *TicketRepository {
	return &TicketRepository{db: db}
}

func (r *TicketRepository) GetAll(ctx context.Context, filters domain.TicketFilters, page, limit int) ([]domain.Ticket, int, *utils.Error) {
	var whereClauses []string
	var args []interface{}
	argPos := 1

	if len(filters.Status) > 0 {
		placeholders := make([]string, len(filters.Status))
		for i, Status := range filters.Status {
			placeholders[i] = fmt.Sprintf("$%d", argPos)
			args = append(args, Status)
			argPos++
		}
		whereClauses = append(whereClauses, fmt.Sprintf("ticket_Status IN (%s)", strings.Join(placeholders, ",")))
	}
	if filters.MovieShowID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("movie_show_id = $%d", argPos))
		args = append(args, filters.MovieShowID)
		argPos++
	}
	if filters.PriceMin > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("price >= $%d", argPos))
		args = append(args, filters.PriceMin)
		argPos++
	}
	if filters.PriceMax > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("price <= $%d", argPos))
		args = append(args, filters.PriceMax)
		argPos++
	}
	if len(filters.SeatID) > 0 {
		placeholders := make([]string, len(filters.SeatID))
		for i, seatID := range filters.SeatID {
			placeholders[i] = fmt.Sprintf("$%d", argPos)
			args = append(args, seatID)
			argPos++
		}
		whereClauses = append(whereClauses, fmt.Sprintf("seat_id IN (%s)", strings.Join(placeholders, ",")))
	}
	if len(filters.UserID) > 0 {
		placeholders := make([]string, len(filters.UserID))
		for i, userID := range filters.UserID {
			placeholders[i] = fmt.Sprintf("$%d", argPos)
			args = append(args, userID)
			argPos++
		}
		whereClauses = append(whereClauses, fmt.Sprintf("user_id IN (%s)", strings.Join(placeholders, ",")))
	}

	countQuery := "SELECT COUNT(*) FROM tickets"
	if len(whereClauses) > 0 {
		countQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, utils.ConvertError(err)
	}

	// Если нет результатов, возвращаем пустой список
	if total == 0 {
		return []domain.Ticket{}, 0, nil
	}

	// Build the main query
	query := "SELECT id, movie_show_id, seat_id, user_id, ticket_Status, price FROM tickets"
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}
	query += " ORDER BY id DESC LIMIT $%d OFFSET $%d"

	// Calculate offset for pagination
	offset := (page - 1) * limit

	// Create new args slice for the main query that includes LIMIT and OFFSET
	mainQueryArgs := make([]interface{}, len(args))
	copy(mainQueryArgs, args)
	mainQueryArgs = append(mainQueryArgs, limit, offset)

	// Update the query with correct parameter positions
	finalQuery := fmt.Sprintf(query, len(mainQueryArgs)-1, len(mainQueryArgs))

	rows, err := r.db.Query(ctx, finalQuery, mainQueryArgs...)
	if err != nil {
		return nil, 0, utils.ConvertError(err)
	}
	defer rows.Close()

	var ticketEntities []entity.Ticket
	for rows.Next() {
		var ticket entity.Ticket
		var userID *string
		if err := rows.Scan(&ticket.ID, &ticket.MovieShowID, &ticket.SeatID, &userID, &ticket.Status, &ticket.Price); err != nil {
			return nil, 0, utils.ConvertError(err)
		}
		ticket.UserID = userID
		ticketEntities = append(ticketEntities, ticket)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, utils.ConvertError(err)
	}

	return entity.TicketsToDomain(ticketEntities), total, nil
}

func (r *TicketRepository) GetByID(ctx context.Context, id string) (domain.Ticket, *utils.Error) {
	var ticketEntity entity.Ticket
	var userID *string

	query := "SELECT id, movie_show_id, seat_id, user_id, ticket_Status, price FROM tickets WHERE id = $1"
	err := r.db.QueryRow(ctx, query, id).Scan(&ticketEntity.ID, &ticketEntity.MovieShowID, &ticketEntity.SeatID, &userID, &ticketEntity.Status, &ticketEntity.Price)
	if err != nil {
		return domain.Ticket{}, utils.ConvertError(err)
	}
	ticketEntity.UserID = userID

	return entity.TicketToDomain(ticketEntity), nil
}

func (r *TicketRepository) CreateForMovieShow(ctx context.Context, movieShowId string, ticket domain.Ticket) (domain.Ticket, *utils.Error) {
	ticketEntity := entity.TicketFromDomain(ticket)
	ticketEntity.MovieShowID = movieShowId

	query := "INSERT INTO tickets (movie_show_id, seat_id, user_id, ticket_Status, price) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	err := r.db.QueryRow(ctx, query, ticketEntity.MovieShowID, ticketEntity.SeatID, ticketEntity.UserID, ticketEntity.Status, ticketEntity.Price).Scan(&ticketEntity.ID)
	if err != nil {
		return domain.Ticket{}, utils.ConvertError(err)
	}

	return entity.TicketToDomain(ticketEntity), nil
}

func (r *TicketRepository) UpdateStatus(ctx context.Context, id string, StatusData domain.Ticket) (domain.Ticket, *utils.Error) {
	query := "UPDATE tickets SET ticket_Status = $1, user_id = $2 WHERE id = $3 RETURNING id, movie_show_id, seat_id, user_id, ticket_Status, price"
	var updatedTicket entity.Ticket
	var updatedUserID *string
	err := r.db.QueryRow(ctx, query, StatusData.Status, StatusData.UserID, id).Scan(
		&updatedTicket.ID, &updatedTicket.MovieShowID, &updatedTicket.SeatID, &updatedUserID, &updatedTicket.Status, &updatedTicket.Price,
	)
	if err != nil {
		return domain.Ticket{}, utils.ConvertError(err)
	}
	updatedTicket.UserID = updatedUserID

	return entity.TicketToDomain(updatedTicket), nil
}

func (r *TicketRepository) Delete(ctx context.Context, id string) *utils.Error {
	query := "DELETE FROM tickets WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return utils.ConvertError(err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return utils.NewNotFound("ticket not found", nil)
	}

	return nil
}
