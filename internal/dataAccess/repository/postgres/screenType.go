package postgres

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	entity "cw/internal/dataAccess/models"
	domain "cw/internal/domain/models"
	"cw/internal/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ScreenTypeRepository struct {
	db *pgxpool.Pool
}

func NewScreenTypeRepository(db *pgxpool.Pool) *ScreenTypeRepository {
	return &ScreenTypeRepository{db: db}
}

func (r *ScreenTypeRepository) GetAll(ctx context.Context, filters domain.ScreenTypeFilters, page, limit int) ([]domain.ScreenType, int, *utils.Error) {
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

	countQuery := "SELECT COUNT(*) FROM screen_types"
	if len(whereClauses) > 0 {
		countQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, utils.ConvertError(err)
	}

	query := "SELECT id, name, description, price_modifier FROM screen_types"
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}
	query += " ORDER BY name LIMIT $" + strconv.Itoa(argPos) + " OFFSET $" + strconv.Itoa(argPos+1)

	args = append(args, limit, (page-1)*limit)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, utils.ConvertError(err)
	}
	defer rows.Close()

	var screenTypeEntities []entity.ScreenType
	for rows.Next() {
		var screenType entity.ScreenType
		if err := rows.Scan(&screenType.ID, &screenType.Name, &screenType.Description, &screenType.PriceModifier); err != nil {
			return nil, 0, utils.ConvertError(err)
		}
		screenTypeEntities = append(screenTypeEntities, screenType)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, utils.ConvertError(err)
	}

	return entity.ScreenTypesToDomain(screenTypeEntities), total, nil
}

func (r *ScreenTypeRepository) GetByID(ctx context.Context, id string) (domain.ScreenType, *utils.Error) {
	var screenTypeEntity entity.ScreenType

	query := "SELECT id, name, description, price_modifier FROM screen_types WHERE id = $1"
	err := r.db.QueryRow(ctx, query, id).Scan(&screenTypeEntity.ID, &screenTypeEntity.Name, &screenTypeEntity.Description, &screenTypeEntity.PriceModifier)
	if err != nil {
		return domain.ScreenType{}, utils.ConvertError(err)
	}

	return entity.ScreenTypeToDomain(screenTypeEntity), nil
}

func (r *ScreenTypeRepository) Create(ctx context.Context, screenType domain.ScreenType) (domain.ScreenType, *utils.Error) {
	screenTypeEntity := entity.ScreenTypeFromDomain(screenType)

	query := "INSERT INTO screen_types (name, description, price_modifier) VALUES ($1, $2, $3) RETURNING id"
	err := r.db.QueryRow(ctx, query, screenTypeEntity.Name, screenTypeEntity.Description, screenTypeEntity.PriceModifier).Scan(&screenTypeEntity.ID)
	if err != nil {
		return domain.ScreenType{}, utils.ConvertError(err)
	}

	return entity.ScreenTypeToDomain(screenTypeEntity), nil
}

func (r *ScreenTypeRepository) Update(ctx context.Context, id string, screenType domain.ScreenType) (domain.ScreenType, *utils.Error) {
	screenTypeEntity := entity.ScreenTypeFromDomain(screenType)

	query := "UPDATE screen_types SET name = $1, description = $2, price_modifier = $3 WHERE id = $4 RETURNING id, name, description, price_modifier"
	var updatedScreenType entity.ScreenType
	err := r.db.QueryRow(ctx, query, screenTypeEntity.Name, screenTypeEntity.Description, screenTypeEntity.PriceModifier, id).Scan(
		&updatedScreenType.ID, &updatedScreenType.Name, &updatedScreenType.Description, &updatedScreenType.PriceModifier,
	)
	if err != nil {
		return domain.ScreenType{}, utils.ConvertError(err)
	}

	return entity.ScreenTypeToDomain(updatedScreenType), nil
}

func (r *ScreenTypeRepository) Delete(ctx context.Context, id string) *utils.Error {
	query := "DELETE FROM screen_types WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return utils.ConvertError(err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return utils.NewNotFound("screen type not found", nil)
	}

	return nil
}
