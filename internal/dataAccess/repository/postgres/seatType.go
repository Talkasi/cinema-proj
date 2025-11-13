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

type SeatTypeRepository struct {
	db *pgxpool.Pool
}

func NewSeatTypeRepository(db *pgxpool.Pool) *SeatTypeRepository {
	return &SeatTypeRepository{db: db}
}

func (r *SeatTypeRepository) GetAll(ctx context.Context, filters domain.SeatTypeFilters, page, limit int) ([]domain.SeatType, int, *utils.Error) {

	whereClauses, args, err := r.buildFilterClauses(filters)
	if err != nil {
		return nil, 0, err
	}

	total, err := r.getCount(ctx, whereClauses, args)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []domain.SeatType{}, 0, nil
	}

	query, queryArgs := r.buildGetAllQuery(whereClauses, args, page, limit)

	seatTypeEntities, err := r.executeGetAllQuery(ctx, query, queryArgs)
	if err != nil {
		return nil, 0, err
	}

	return entity.SeatTypesToDomain(seatTypeEntities), total, nil
}

func (r *SeatTypeRepository) buildFilterClauses(filters domain.SeatTypeFilters) ([]string, []interface{}, *utils.Error) {
	var whereClauses []string
	var args []interface{}
	argPos := 1

	if filters.Name != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", argPos))
		args = append(args, "%"+filters.Name+"%")
		argPos++
	}

	if filters.Description != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("description ILIKE $%d", argPos))
		args = append(args, "%"+filters.Description+"%")
	}

	return whereClauses, args, nil
}

func (r *SeatTypeRepository) getCount(ctx context.Context, whereClauses []string, args []interface{}) (int, *utils.Error) {
	countQuery := "SELECT COUNT(*) FROM seat_types"
	if len(whereClauses) > 0 {
		countQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return 0, utils.ConvertError(err)
	}

	return total, nil
}

func (r *SeatTypeRepository) buildGetAllQuery(whereClauses []string, args []interface{}, page, limit int) (string, []interface{}) {
	query := "SELECT id, name, description, price_modifier FROM seat_types"
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}
	query += " ORDER BY name LIMIT $%d OFFSET $%d"

	offset := (page - 1) * limit

	mainQueryArgs := make([]interface{}, len(args))
	copy(mainQueryArgs, args)
	mainQueryArgs = append(mainQueryArgs, limit, offset)

	finalQuery := fmt.Sprintf(query, len(mainQueryArgs)-1, len(mainQueryArgs))

	return finalQuery, mainQueryArgs
}

func (r *SeatTypeRepository) executeGetAllQuery(ctx context.Context, query string, args []interface{}) ([]entity.SeatType, *utils.Error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, utils.ConvertError(err)
	}
	defer rows.Close()

	var seatTypeEntities []entity.SeatType
	for rows.Next() {
		var seatType entity.SeatType
		if err := rows.Scan(&seatType.ID, &seatType.Name, &seatType.Description, &seatType.PriceModifier); err != nil {
			return nil, utils.ConvertError(err)
		}
		seatTypeEntities = append(seatTypeEntities, seatType)
	}

	if err := rows.Err(); err != nil {
		return nil, utils.ConvertError(err)
	}

	return seatTypeEntities, nil
}

func (r *SeatTypeRepository) GetByID(ctx context.Context, id string) (domain.SeatType, *utils.Error) {
	var seatTypeEntity entity.SeatType

	query := "SELECT id, name, description, price_modifier FROM seat_types WHERE id = $1"
	err := r.db.QueryRow(ctx, query, id).Scan(&seatTypeEntity.ID, &seatTypeEntity.Name, &seatTypeEntity.Description, &seatTypeEntity.PriceModifier)
	if err != nil {
		return domain.SeatType{}, utils.ConvertError(err)
	}

	return entity.SeatTypeToDomain(seatTypeEntity), nil
}

func (r *SeatTypeRepository) Create(ctx context.Context, seatType domain.SeatType) (domain.SeatType, *utils.Error) {
	seatTypeEntity := entity.SeatTypeFromDomain(seatType)

	query := "INSERT INTO seat_types (name, description, price_modifier) VALUES ($1, $2, $3) RETURNING id"
	err := r.db.QueryRow(ctx, query, seatTypeEntity.Name, seatTypeEntity.Description, seatTypeEntity.PriceModifier).Scan(&seatTypeEntity.ID)
	if err != nil {
		return domain.SeatType{}, utils.ConvertError(err)
	}

	return entity.SeatTypeToDomain(seatTypeEntity), nil
}

func (r *SeatTypeRepository) Update(ctx context.Context, id string, seatType domain.SeatType) (domain.SeatType, *utils.Error) {
	seatTypeEntity := entity.SeatTypeFromDomain(seatType)

	query := "UPDATE seat_types SET name = $1, description = $2, price_modifier = $3 WHERE id = $4 RETURNING id, name, description, price_modifier"
	var updatedSeatType entity.SeatType
	err := r.db.QueryRow(ctx, query, seatTypeEntity.Name, seatTypeEntity.Description, seatTypeEntity.PriceModifier, id).Scan(
		&updatedSeatType.ID, &updatedSeatType.Name, &updatedSeatType.Description, &updatedSeatType.PriceModifier,
	)
	if err != nil {
		return domain.SeatType{}, utils.ConvertError(err)
	}

	return entity.SeatTypeToDomain(updatedSeatType), nil
}

func (r *SeatTypeRepository) Delete(ctx context.Context, id string) *utils.Error {
	query := "DELETE FROM seat_types WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return utils.ConvertError(err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return utils.NewNotFound("seat type not found", nil)
	}

	return nil
}
