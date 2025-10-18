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

type MovieRepository struct {
	db *pgxpool.Pool
}

func NewMovieRepository(db *pgxpool.Pool) *MovieRepository {
	return &MovieRepository{db: db}
}

func (r *MovieRepository) GetAll(ctx context.Context, filters domain.MovieFilters, page, limit int) ([]domain.Movie, int, *utils.Error) {
	var whereClauses []string
	var countArgs []interface{}
	argPos := 1

	baseQuery := `
		SELECT m.id, m.title, m.duration, m.description, m.age_limit, 
		       m.box_office_revenue, m.release_date,
		       COALESCE(AVG(r.rating), 0) as rating,
		       ARRAY_AGG(DISTINCT mg.genre_id) FILTER (WHERE mg.genre_id IS NOT NULL) as genre_ids
		FROM movies m
		LEFT JOIN reviews r ON m.id = r.movie_id
		LEFT JOIN movies_genres mg ON m.id = mg.movie_id
	`

	if filters.Title != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("m.title ILIKE $%d", argPos))
		countArgs = append(countArgs, "%"+filters.Title+"%")
		argPos++
	}
	if filters.Genre != "" {
		baseQuery += " JOIN movies_genres mg2 ON m.id = mg2.movie_id JOIN genres g ON mg2.genre_id = g.id"
		whereClauses = append(whereClauses, fmt.Sprintf("g.name ILIKE $%d", argPos))
		countArgs = append(countArgs, "%"+filters.Genre+"%")
		argPos++
	}

	if len(whereClauses) > 0 {
		baseQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	baseQuery += " GROUP BY m.id"

	countQuery := "SELECT COUNT(*) FROM (" + baseQuery + ") as counted"
	var total int
	err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, utils.ConvertError(err)
	}

	dataArgs := make([]interface{}, len(countArgs))
	copy(dataArgs, countArgs)

	dataQuery := baseQuery + " ORDER BY m.title"

	if limit > 0 && page > 0 {
		dataQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(dataArgs)+1, len(dataArgs)+2)
		dataArgs = append(dataArgs, limit, (page-1)*limit)
	}

	rows, err := r.db.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, utils.ConvertError(err)
	}
	defer rows.Close()

	var movieEntities []entity.Movie
	for rows.Next() {
		var movieEntity entity.Movie
		var genreIDs []string
		err := rows.Scan(
			&movieEntity.ID, &movieEntity.Title, &movieEntity.Duration, &movieEntity.Description,
			&movieEntity.AgeLimit, &movieEntity.BoxOfficeRevenue, &movieEntity.ReleaseDate,
			&movieEntity.Rating, &genreIDs,
		)
		if err != nil {
			return nil, 0, utils.ConvertError(err)
		}
		movieEntity.GenreIDs = genreIDs
		movieEntities = append(movieEntities, movieEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, utils.ConvertError(err)
	}

	return entity.MoviesToDomain(movieEntities), total, nil
}

func (r *MovieRepository) GetByID(ctx context.Context, id string) (domain.Movie, *utils.Error) {
	query := `
		SELECT m.id, m.title, m.duration, m.description, m.age_limit, 
		       m.box_office_revenue, m.release_date,
		       COALESCE(AVG(r.rating), 0) as rating,
		       ARRAY_AGG(DISTINCT mg.genre_id) FILTER (WHERE mg.genre_id IS NOT NULL) as genre_ids
		FROM movies m
		LEFT JOIN reviews r ON m.id = r.movie_id
		LEFT JOIN movies_genres mg ON m.id = mg.movie_id
		WHERE m.id = $1
		GROUP BY m.id
	`

	var movieEntity entity.Movie
	var genreIDs []string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&movieEntity.ID, &movieEntity.Title, &movieEntity.Duration, &movieEntity.Description,
		&movieEntity.AgeLimit, &movieEntity.BoxOfficeRevenue, &movieEntity.ReleaseDate,
		&movieEntity.Rating, &genreIDs,
	)
	if err != nil {
		return domain.Movie{}, utils.ConvertError(err)
	}
	movieEntity.GenreIDs = genreIDs

	return entity.MovieToDomain(movieEntity), nil
}

func (r *MovieRepository) Create(ctx context.Context, movie domain.Movie) (domain.Movie, *utils.Error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Movie{}, utils.ConvertError(err)
	}
	defer tx.Rollback(ctx)

	movieEntity := entity.MovieFromDomain(movie)

	query := `
		INSERT INTO movies (title, duration, description, age_limit, box_office_revenue, release_date)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	err = tx.QueryRow(ctx, query,
		movieEntity.Title, movieEntity.Duration, movieEntity.Description,
		movieEntity.AgeLimit, movieEntity.BoxOfficeRevenue, movieEntity.ReleaseDate,
	).Scan(&movieEntity.ID)
	if err != nil {
		return domain.Movie{}, utils.ConvertError(err)
	}

	if len(movieEntity.GenreIDs) > 0 {
		genreQuery := "INSERT INTO movies_genres (movie_id, genre_id) VALUES "
		var genreArgs []interface{}
		var placeholders []string

		for i, genreID := range movieEntity.GenreIDs {
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
			genreArgs = append(genreArgs, movieEntity.ID, genreID)
		}

		genreQuery += strings.Join(placeholders, ", ")
		_, err = tx.Exec(ctx, genreQuery, genreArgs...)
		if err != nil {
			return domain.Movie{}, utils.ConvertError(err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return domain.Movie{}, utils.ConvertError(err)
	}

	return r.GetByID(ctx, movieEntity.ID)
}

func (r *MovieRepository) Update(ctx context.Context, id string, movie domain.Movie) (domain.Movie, *utils.Error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Movie{}, utils.ConvertError(err)
	}
	defer tx.Rollback(ctx)

	movieEntity := entity.MovieFromDomain(movie)

	query := `
		UPDATE movies 
		SET title = $1, duration = $2, description = $3, age_limit = $4, 
		    box_office_revenue = $5, release_date = $6
		WHERE id = $7
	`
	result, err := tx.Exec(ctx, query,
		movieEntity.Title, movieEntity.Duration, movieEntity.Description,
		movieEntity.AgeLimit, movieEntity.BoxOfficeRevenue, movieEntity.ReleaseDate,
		id,
	)
	if err != nil {
		return domain.Movie{}, utils.ConvertError(err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.Movie{}, utils.NewNotFound("movie not found", nil)
	}

	_, err = tx.Exec(ctx, "DELETE FROM movies_genres WHERE movie_id = $1", id)
	if err != nil {
		return domain.Movie{}, utils.ConvertError(err)
	}

	if len(movieEntity.GenreIDs) > 0 {
		genreQuery := "INSERT INTO movies_genres (movie_id, genre_id) VALUES "
		var genreArgs []interface{}
		var placeholders []string

		for i, genreID := range movieEntity.GenreIDs {
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
			genreArgs = append(genreArgs, id, genreID)
		}

		genreQuery += strings.Join(placeholders, ", ")
		_, err = tx.Exec(ctx, genreQuery, genreArgs...)
		if err != nil {
			return domain.Movie{}, utils.ConvertError(err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return domain.Movie{}, utils.ConvertError(err)
	}

	return r.GetByID(ctx, id)
}

func (r *MovieRepository) Delete(ctx context.Context, id string) *utils.Error {
	query := "DELETE FROM movies WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return utils.ConvertError(err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return utils.NewNotFound("movie not found", nil)
	}

	return nil
}

func (r *MovieRepository) CalculateMovieRating(ctx context.Context, movieID string) (float64, *utils.Error) {
	query := "SELECT COALESCE(AVG(rating), 0) FROM reviews WHERE movie_id = $1"
	var rating float64
	err := r.db.QueryRow(ctx, query, movieID).Scan(&rating)
	if err != nil {
		return 0, utils.ConvertError(err)
	}
	return rating, nil
}
