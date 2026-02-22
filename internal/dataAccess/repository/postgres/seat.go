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

type SeatRepository struct {
	db *pgxpool.Pool
}

func NewSeatRepository(db *pgxpool.Pool) *SeatRepository {
	return &SeatRepository{db: db}
}

func (r *SeatRepository) GetByHall(ctx context.Context, hallId string, filters domain.SeatFilters, page, limit int) ([]domain.Seat, int, *utils.Error) {
	var whereClauses []string
	var args []interface{}
	argPos := 1

	whereClauses = append(whereClauses, fmt.Sprintf("hall_id = $%d", argPos))
	args = append(args, hallId)
	argPos++

	if filters.SeatTypeID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("seat_type_id = $%d", argPos))
		args = append(args, filters.SeatTypeID)
		argPos++
	}
	if filters.RowNumberMin > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("row_number >= $%d", argPos))
		args = append(args, filters.RowNumberMin)
		argPos++
	}
	if filters.RowNumberMax > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("row_number <= $%d", argPos))
		args = append(args, filters.RowNumberMax)
		argPos++
	}
	if filters.SeatNumberMin > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("seat_number >= $%d", argPos))
		args = append(args, filters.SeatNumberMin)
		argPos++
	}
	if filters.SeatNumberMax > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("seat_number <= $%d", argPos))
		args = append(args, filters.SeatNumberMax)
		argPos++
	}

	countQuery := "SELECT COUNT(*) FROM seats"
	if len(whereClauses) > 0 {
		countQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, utils.ConvertError(err)
	}

	query := "SELECT id, hall_id, seat_type_id, row_number, seat_number FROM seats"
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}
	query += " ORDER BY row_number, seat_number LIMIT $%d OFFSET $%d"

	offset := (page - 1) * limit

	mainQueryArgs := make([]interface{}, len(args))
	copy(mainQueryArgs, args)
	mainQueryArgs = append(mainQueryArgs, limit, offset)

	finalQuery := fmt.Sprintf(query, len(mainQueryArgs)-1, len(mainQueryArgs))

	rows, err := r.db.Query(ctx, finalQuery, mainQueryArgs...)
	if err != nil {
		return nil, 0, utils.ConvertError(err)
	}
	defer rows.Close()

	var seatEntities []entity.Seat
	for rows.Next() {
		var seat entity.Seat
		if err := rows.Scan(&seat.ID, &seat.HallID, &seat.SeatTypeID, &seat.RowNumber, &seat.SeatNumber); err != nil {
			return nil, 0, utils.ConvertError(err)
		}
		seatEntities = append(seatEntities, seat)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, utils.ConvertError(err)
	}

	return entity.SeatsToDomain(seatEntities), total, nil
}

func (r *SeatRepository) GetByID(ctx context.Context, id string) (domain.Seat, *utils.Error) {
	var seatEntity entity.Seat

	query := "SELECT id, hall_id, seat_type_id, row_number, seat_number FROM seats WHERE id = $1"
	err := r.db.QueryRow(ctx, query, id).Scan(&seatEntity.ID, &seatEntity.HallID, &seatEntity.SeatTypeID, &seatEntity.RowNumber, &seatEntity.SeatNumber)
	if err != nil {
		return domain.Seat{}, utils.ConvertError(err)
	}

	return entity.SeatToDomain(seatEntity), nil
}

func (r *SeatRepository) Create(ctx context.Context, seat domain.Seat) (domain.Seat, *utils.Error) {
	seatEntity := entity.SeatFromDomain(seat)

	query := "INSERT INTO seats (hall_id, seat_type_id, row_number, seat_number) VALUES ($1, $2, $3, $4) RETURNING id"
	err := r.db.QueryRow(ctx, query, seatEntity.HallID, seatEntity.SeatTypeID, seatEntity.RowNumber, seatEntity.SeatNumber).Scan(&seatEntity.ID)
	if err != nil {
		return domain.Seat{}, utils.ConvertError(err)
	}

	return entity.SeatToDomain(seatEntity), nil
}

func (r *SeatRepository) Update(ctx context.Context, id string, seat domain.Seat) (domain.Seat, *utils.Error) {
	seatEntity := entity.SeatFromDomain(seat)

	query := "UPDATE seats SET hall_id = $1, seat_type_id = $2, row_number = $3, seat_number = $4 WHERE id = $5 RETURNING id, hall_id, seat_type_id, row_number, seat_number"
	var updatedSeat entity.Seat
	err := r.db.QueryRow(ctx, query, seatEntity.HallID, seatEntity.SeatTypeID, seatEntity.RowNumber, seatEntity.SeatNumber, id).Scan(
		&updatedSeat.ID, &updatedSeat.HallID, &updatedSeat.SeatTypeID, &updatedSeat.RowNumber, &updatedSeat.SeatNumber,
	)
	if err != nil {
		return domain.Seat{}, utils.ConvertError(err)
	}

	return entity.SeatToDomain(updatedSeat), nil
}

func (r *SeatRepository) Delete(ctx context.Context, id string) *utils.Error {
	query := "DELETE FROM seats WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return utils.ConvertError(err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return utils.NewNotFound("seat not found", nil)
	}

	return nil
}
