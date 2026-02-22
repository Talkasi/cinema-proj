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

type HallRepository struct {
	db *pgxpool.Pool
}

func NewHallRepository(db *pgxpool.Pool) *HallRepository {
	return &HallRepository{db: db}
}

func (r *HallRepository) GetAll(ctx context.Context, filters domain.HallFilters, page, limit int) ([]domain.Hall, int, *utils.Error) {

	whereClauses, args, err := r.buildFilterClauses(filters)
	if err != nil {
		return nil, 0, err
	}

	total, err := r.getCount(ctx, whereClauses, args)
	if err != nil {
		return nil, 0, err
	}

	query, queryArgs := r.buildGetAllQuery(whereClauses, args, page, limit)

	hallEntities, err := r.executeGetAllQuery(ctx, query, queryArgs)
	if err != nil {
		return nil, 0, err
	}

	return entity.HallsToDomain(hallEntities), total, nil
}

func (r *HallRepository) buildFilterClauses(filters domain.HallFilters) ([]string, []interface{}, *utils.Error) {
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
		argPos++
	}

	if filters.ScreenTypeID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("screen_type_id = $%d", argPos))
		args = append(args, filters.ScreenTypeID)
		argPos++
	}

	_ = argPos // Mark as used to avoid linter error

	return whereClauses, args, nil
}

func (r *HallRepository) getCount(ctx context.Context, whereClauses []string, args []interface{}) (int, *utils.Error) {
	countQuery := "SELECT COUNT(*) FROM halls"
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

func (r *HallRepository) buildGetAllQuery(whereClauses []string, args []interface{}, page, limit int) (string, []interface{}) {
	query := "SELECT id, name, description, screen_type_id FROM halls"
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}
	query += " ORDER BY name"

	if limit > 0 && page > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
		args = append(args, limit, (page-1)*limit)
	}

	return query, args
}

func (r *HallRepository) executeGetAllQuery(ctx context.Context, query string, args []interface{}) ([]entity.Hall, *utils.Error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, utils.ConvertError(err)
	}
	defer rows.Close()

	var hallEntities []entity.Hall
	for rows.Next() {
		var hall entity.Hall
		if err := rows.Scan(&hall.ID, &hall.Name, &hall.Description, &hall.ScreenTypeID); err != nil {
			return nil, utils.ConvertError(err)
		}
		hallEntities = append(hallEntities, hall)
	}

	if err := rows.Err(); err != nil {
		return nil, utils.ConvertError(err)
	}

	return hallEntities, nil
}

func (r *HallRepository) GetByID(ctx context.Context, id string) (domain.Hall, *utils.Error) {
	var hallEntity entity.Hall

	query := "SELECT id, name, description, screen_type_id FROM halls WHERE id = $1"
	err := r.db.QueryRow(ctx, query, id).Scan(&hallEntity.ID, &hallEntity.Name, &hallEntity.Description, &hallEntity.ScreenTypeID)
	if err != nil {
		return domain.Hall{}, utils.ConvertError(err)
	}

	return entity.HallToDomain(hallEntity), nil
}

func (r *HallRepository) Create(ctx context.Context, hall domain.Hall) (domain.Hall, *utils.Error) {
	hallEntity := entity.HallFromDomain(hall)

	query := "INSERT INTO halls (name, description, screen_type_id) VALUES ($1, $2, $3) RETURNING id"
	err := r.db.QueryRow(ctx, query, hallEntity.Name, hallEntity.Description, hallEntity.ScreenTypeID).Scan(&hallEntity.ID)
	if err != nil {
		return domain.Hall{}, utils.ConvertError(err)
	}

	return entity.HallToDomain(hallEntity), nil
}

func (r *HallRepository) Update(ctx context.Context, id string, hall domain.Hall) (domain.Hall, *utils.Error) {
	hallEntity := entity.HallFromDomain(hall)

	query := "UPDATE halls SET name = $1, description = $2, screen_type_id = $3 WHERE id = $4 RETURNING id, name, description, screen_type_id"
	var updatedHall entity.Hall
	err := r.db.QueryRow(ctx, query, hallEntity.Name, hallEntity.Description, hallEntity.ScreenTypeID, id).Scan(
		&updatedHall.ID, &updatedHall.Name, &updatedHall.Description, &updatedHall.ScreenTypeID,
	)
	if err != nil {
		return domain.Hall{}, utils.ConvertError(err)
	}

	return entity.HallToDomain(updatedHall), nil
}

func (r *HallRepository) Delete(ctx context.Context, id string) *utils.Error {
	query := "DELETE FROM halls WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return utils.ConvertError(err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return utils.NewNotFound("hall not found", nil)
	}

	return nil
}
