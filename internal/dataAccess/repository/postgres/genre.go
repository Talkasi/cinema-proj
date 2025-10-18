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

type GenreRepository struct {
	db *pgxpool.Pool
}

func NewGenreRepository(db *pgxpool.Pool) *GenreRepository {
	return &GenreRepository{db: db}
}

func (r *GenreRepository) GetAll(ctx context.Context, filters domain.GenreFilters, page, limit int) ([]domain.Genre, int, *utils.Error) {
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

	countQuery := "SELECT COUNT(*) FROM genres"
	if len(whereClauses) > 0 {
		countQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, utils.ConvertError(err)
	}

	query := "SELECT id, name, description FROM genres"
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}
	query += " ORDER BY name"

	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
		args = append(args, limit, offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, utils.ConvertError(err)
	}
	defer rows.Close()

	var genreEntities []entity.Genre
	for rows.Next() {
		var genre entity.Genre
		if err := rows.Scan(&genre.ID, &genre.Name, &genre.Description); err != nil {
			return nil, 0, utils.ConvertError(err)
		}
		genreEntities = append(genreEntities, genre)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, utils.ConvertError(err)
	}

	return entity.GenresToDomain(genreEntities), total, nil
}

func (r *GenreRepository) GetByID(ctx context.Context, id string) (domain.Genre, *utils.Error) {
	var genreEntity entity.Genre

	query := "SELECT id, name, description FROM genres WHERE id = $1"
	err := r.db.QueryRow(ctx, query, id).Scan(&genreEntity.ID, &genreEntity.Name, &genreEntity.Description)
	if err != nil {
		return domain.Genre{}, utils.ConvertError(err)
	}

	return entity.GenreToDomain(genreEntity), nil
}

func (r *GenreRepository) Create(ctx context.Context, genre domain.Genre) (domain.Genre, *utils.Error) {
	genreEntity := entity.GenreFromDomain(genre)

	query := "INSERT INTO genres (name, description) VALUES ($1, $2) RETURNING id"
	err := r.db.QueryRow(ctx, query, genreEntity.Name, genreEntity.Description).Scan(&genreEntity.ID)
	if err != nil {
		return domain.Genre{}, utils.ConvertError(err)
	}

	return entity.GenreToDomain(genreEntity), nil
}

func (r *GenreRepository) Update(ctx context.Context, id string, genre domain.Genre) (domain.Genre, *utils.Error) {
	genreEntity := entity.GenreFromDomain(genre)

	query := "UPDATE genres SET name = $1, description = $2 WHERE id = $3 RETURNING id, name, description"
	var updatedGenre entity.Genre
	err := r.db.QueryRow(ctx, query, genreEntity.Name, genreEntity.Description, id).Scan(
		&updatedGenre.ID, &updatedGenre.Name, &updatedGenre.Description,
	)
	if err != nil {
		return domain.Genre{}, utils.ConvertError(err)
	}

	return entity.GenreToDomain(updatedGenre), nil
}

func (r *GenreRepository) Delete(ctx context.Context, id string) *utils.Error {
	query := "DELETE FROM genres WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return utils.ConvertError(err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return utils.NewNotFound("genre not found", nil)
	}

	return nil
}
