package postgres

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	entity "cw/internal/dataAccess/models"
	domain "cw/internal/domain/models"
	"cw/internal/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MovieShowRepository struct {
	db *pgxpool.Pool
}

func NewMovieShowRepository(db *pgxpool.Pool) *MovieShowRepository {
	return &MovieShowRepository{db: db}
}

func (r *MovieShowRepository) GetAll(ctx context.Context, filters domain.MovieShowFilters, page, limit int) ([]domain.MovieShow, int, *utils.Error) {
	var whereClauses []string
	var args []interface{}
	argPos := 1

	if len(filters.MovieIDs) > 0 {
		placeholders := make([]string, len(filters.MovieIDs))
		for i, id := range filters.MovieIDs {
			placeholders[i] = fmt.Sprintf("$%d", argPos)
			args = append(args, id)
			argPos++
		}
		whereClauses = append(whereClauses, fmt.Sprintf("movie_id IN (%s)", strings.Join(placeholders, ",")))
	}
	if len(filters.HallIDs) > 0 {
		placeholders := make([]string, len(filters.HallIDs))
		for i, id := range filters.HallIDs {
			placeholders[i] = fmt.Sprintf("$%d", argPos)
			args = append(args, id)
			argPos++
		}
		whereClauses = append(whereClauses, fmt.Sprintf("hall_id IN (%s)", strings.Join(placeholders, ",")))
	}
	if len(filters.Languages) > 0 {
		placeholders := make([]string, len(filters.Languages))
		for i, lang := range filters.Languages {
			placeholders[i] = fmt.Sprintf("$%d", argPos)
			args = append(args, lang)
			argPos++
		}
		whereClauses = append(whereClauses, fmt.Sprintf("language::text IN (%s)", strings.Join(placeholders, ",")))
	}
	if filters.Date != "" {
		date, err := time.Parse("2006-01-02", filters.Date)
		if err == nil {
			startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
			endOfDay := startOfDay.Add(24 * time.Hour)
			whereClauses = append(whereClauses, fmt.Sprintf("start_time >= $%d AND start_time < $%d", argPos, argPos+1))
			args = append(args, startOfDay, endOfDay)
			argPos += 2
		}
	}
	if filters.StartTimeFrom != "" {
		startTimeFrom, err := time.Parse("15:04", filters.StartTimeFrom)
		if err == nil {
			whereClauses = append(whereClauses, fmt.Sprintf("EXTRACT(HOUR FROM start_time) * 60 + EXTRACT(MINUTE FROM start_time) >= $%d", argPos))
			args = append(args, startTimeFrom.Hour()*60+startTimeFrom.Minute())
			argPos++
		}
	}
	if filters.StartTimeTo != "" {
		startTimeTo, err := time.Parse("15:04", filters.StartTimeTo)
		if err == nil {
			whereClauses = append(whereClauses, fmt.Sprintf("EXTRACT(HOUR FROM start_time) * 60 + EXTRACT(MINUTE FROM start_time) <= $%d", argPos))
			args = append(args, startTimeTo.Hour()*60+startTimeTo.Minute())
			argPos++
		}
	}

	countQuery := "SELECT COUNT(*) FROM movie_shows"
	if len(whereClauses) > 0 {
		countQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, utils.ConvertError(err)
	}

	query := "SELECT id, movie_id, hall_id, start_time, language FROM movie_shows"
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}
	query += " ORDER BY start_time ASC LIMIT $" + strconv.Itoa(argPos) + " OFFSET $" + strconv.Itoa(argPos+1)

	args = append(args, limit, (page-1)*limit)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, utils.ConvertError(err)
	}
	defer rows.Close()

	var movieShowEntities []entity.MovieShow
	for rows.Next() {
		var movieShow entity.MovieShow
		if err := rows.Scan(&movieShow.ID, &movieShow.MovieID, &movieShow.HallID, &movieShow.StartTime, &movieShow.Language); err != nil {
			return nil, 0, utils.ConvertError(err)
		}
		movieShowEntities = append(movieShowEntities, movieShow)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, utils.ConvertError(err)
	}

	return entity.MovieShowsToDomain(movieShowEntities), total, nil
}

func (r *MovieShowRepository) GetByID(ctx context.Context, id string) (domain.MovieShow, *utils.Error) {
	var movieShowEntity entity.MovieShow

	query := "SELECT id, movie_id, hall_id, start_time, language FROM movie_shows WHERE id = $1"
	err := r.db.QueryRow(ctx, query, id).Scan(&movieShowEntity.ID, &movieShowEntity.MovieID, &movieShowEntity.HallID, &movieShowEntity.StartTime, &movieShowEntity.Language)
	if err != nil {
		return domain.MovieShow{}, utils.ConvertError(err)
	}

	return entity.MovieShowToDomain(movieShowEntity), nil
}

func (r *MovieShowRepository) Create(ctx context.Context, movieShow domain.MovieShow) (domain.MovieShow, *utils.Error) {
	movieShowEntity := entity.MovieShowFromDomain(movieShow)

	query := "INSERT INTO movie_shows (movie_id, hall_id, start_time, language) VALUES ($1, $2, $3, $4) RETURNING id"
	err := r.db.QueryRow(ctx, query, movieShowEntity.MovieID, movieShowEntity.HallID, movieShowEntity.StartTime, movieShowEntity.Language).Scan(&movieShowEntity.ID)
	if err != nil {
		return domain.MovieShow{}, utils.ConvertError(err)
	}

	return entity.MovieShowToDomain(movieShowEntity), nil
}

func (r *MovieShowRepository) Update(ctx context.Context, id string, movieShow domain.MovieShow) (domain.MovieShow, *utils.Error) {
	movieShowEntity := entity.MovieShowFromDomain(movieShow)

	query := "UPDATE movie_shows SET movie_id = $1, hall_id = $2, start_time = $3, language = $4 WHERE id = $5 RETURNING id, movie_id, hall_id, start_time, language"
	var updatedMovieShow entity.MovieShow
	err := r.db.QueryRow(ctx, query, movieShowEntity.MovieID, movieShowEntity.HallID, movieShowEntity.StartTime, movieShowEntity.Language, id).Scan(
		&updatedMovieShow.ID, &updatedMovieShow.MovieID, &updatedMovieShow.HallID, &updatedMovieShow.StartTime, &updatedMovieShow.Language,
	)
	if err != nil {
		return domain.MovieShow{}, utils.ConvertError(err)
	}

	return entity.MovieShowToDomain(updatedMovieShow), nil
}

func (r *MovieShowRepository) Delete(ctx context.Context, id string) *utils.Error {
	query := "DELETE FROM movie_shows WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return utils.ConvertError(err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return utils.NewNotFound("movie show not found", nil)
	}

	return nil
}
