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
	var args []interface{}
	argPos := 1

	// Build a single query using window functions to get both data and total count
	query := `
		WITH filtered_movies AS (
			SELECT 
				m.id, 
				m.title, 
				m.duration, 
				m.description, 
				m.age_limit, 
				m.box_office_revenue, 
				m.release_date,
				COUNT(*) OVER() AS total_count
			FROM movies m`

	// Add filters to the query
	if filters.Title != "" {
		whereClause := fmt.Sprintf("m.title ILIKE $%d", argPos)
		whereClauses = append(whereClauses, whereClause)
		args = append(args, "%"+filters.Title+"%")
		argPos++
	}
	if filters.Genre != "" {
		query += " JOIN movies_genres mg2 ON m.id = mg2.movie_id JOIN genres g ON mg2.genre_id = g.id"
		whereClause := fmt.Sprintf("g.name ILIKE $%d", argPos)
		whereClauses = append(whereClauses, whereClause)
		args = append(args, "%"+filters.Genre+"%")
		argPos++
	}

	// Apply WHERE clause
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += " ORDER BY m.title"

	if limit > 0 && page > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
		args = append(args, limit, (page-1)*limit)
	}

	query += ") SELECT id, title, duration, description, age_limit, box_office_revenue, release_date, total_count FROM filtered_movies"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, utils.ConvertError(err)
	}
	defer rows.Close()

	var movieIDs []string
	var movieEntities []entity.Movie
	var total int

	for rows.Next() {
		var movieEntity entity.Movie
		var rowCount int
		err := rows.Scan(
			&movieEntity.ID, 
			&movieEntity.Title, 
			&movieEntity.Duration, 
			&movieEntity.Description,
			&movieEntity.AgeLimit, 
			&movieEntity.BoxOfficeRevenue, 
			&movieEntity.ReleaseDate,
			&rowCount,
		)
		if err != nil {
			return nil, 0, utils.ConvertError(err)
		}
		// Only set the total once from the first row
		if total == 0 {
			total = rowCount
		}
		movieIDs = append(movieIDs, movieEntity.ID)
		movieEntities = append(movieEntities, movieEntity)
	}
	rows.Close()

	if err := rows.Err(); err != nil {
		return nil, 0, utils.ConvertError(err)
	}

	// If no movies found, return with total 0
	if len(movieEntities) == 0 {
		return []domain.Movie{}, 0, nil
	}

	// If we have movies to return, fetch ratings and genre IDs separately
	if len(movieIDs) > 0 {
		// Fetch ratings for the returned movies
		ratingQuery := fmt.Sprintf(`
			SELECT movie_id, COALESCE(AVG(rating), 0)
			FROM reviews
			WHERE movie_id = ANY($1)
			GROUP BY movie_id`)
		
		ratingRows, err := r.db.Query(ctx, ratingQuery, movieIDs)
		if err != nil {
			return nil, 0, utils.ConvertError(err)
		}
		defer ratingRows.Close()

		ratingMap := make(map[string]float64)
		for ratingRows.Next() {
			var movieID string
			var rating float64
			if err := ratingRows.Scan(&movieID, &rating); err != nil {
				return nil, 0, utils.ConvertError(err)
			}
			ratingMap[movieID] = rating
		}
		ratingRows.Close()

		// Fetch genre IDs for the returned movies
		genreQuery := fmt.Sprintf(`
			SELECT movie_id, ARRAY_AGG(genre_id) as genre_ids
			FROM movies_genres
			WHERE movie_id = ANY($1)
			GROUP BY movie_id`)
		
		genreRows, err := r.db.Query(ctx, genreQuery, movieIDs)
		if err != nil {
			return nil, 0, utils.ConvertError(err)
		}
		defer genreRows.Close()

		genreMap := make(map[string][]string)
		for genreRows.Next() {
			var movieID string
			var genreIDs []string
			if err := genreRows.Scan(&movieID, &genreIDs); err != nil {
				return nil, 0, utils.ConvertError(err)
			}
			genreMap[movieID] = genreIDs
		}
		genreRows.Close()

		// Combine the data in the correct order
		for i := range movieEntities {
			movieID := movieEntities[i].ID
			// Set rating
			if rating, exists := ratingMap[movieID]; exists {
				movieEntities[i].Rating = rating
			} else {
				movieEntities[i].Rating = 0
			}
			// Set genre IDs
			if genreIDs, exists := genreMap[movieID]; exists {
				movieEntities[i].GenreIDs = genreIDs
			} else {
				movieEntities[i].GenreIDs = []string{}
			}
		}
	}

	return entity.MoviesToDomain(movieEntities), total, nil
}

func (r *MovieRepository) GetByID(ctx context.Context, id string) (domain.Movie, *utils.Error) {
	// First, get the basic movie data without joins
	query := `
		SELECT m.id, m.title, m.duration, m.description, m.age_limit, 
		       m.box_office_revenue, m.release_date
		FROM movies m
		WHERE m.id = $1
	`

	var movieEntity entity.Movie
	err := r.db.QueryRow(ctx, query, id).Scan(
		&movieEntity.ID, &movieEntity.Title, &movieEntity.Duration, &movieEntity.Description,
		&movieEntity.AgeLimit, &movieEntity.BoxOfficeRevenue, &movieEntity.ReleaseDate,
	)
	if err != nil {
		return domain.Movie{}, utils.ConvertError(err)
	}

	// Get the rating separately to avoid expensive JOIN
	ratingQuery := `SELECT COALESCE(AVG(rating), 0) FROM reviews WHERE movie_id = $1`
	var rating float64
	err = r.db.QueryRow(ctx, ratingQuery, id).Scan(&rating)
	if err != nil {
		// If no reviews exist, rating will be 0, which is fine
		rating = 0
	}
	movieEntity.Rating = rating

	// Get genre IDs separately to avoid expensive JOIN
	genreQuery := `SELECT ARRAY_AGG(genre_id) FROM movies_genres WHERE movie_id = $1`
	var genreIDs []string
	err = r.db.QueryRow(ctx, genreQuery, id).Scan(&genreIDs)
	if err != nil || genreIDs == nil {
		// If no genres exist, set empty array
		genreIDs = []string{}
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
