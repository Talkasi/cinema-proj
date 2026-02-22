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

	whereClauses, args, err := r.buildFilterClauses(filters)
	if err != nil {
		return nil, 0, err
	}

	total, err := r.getCount(ctx, whereClauses, args)
	if err != nil {
		return nil, 0, err
	}

	query, args := r.buildGetAllQuery(whereClauses, args, page, limit)

	movieShowEntities, err := r.executeGetAllQuery(ctx, query, args)
	if err != nil {
		return nil, 0, err
	}

	return entity.MovieShowsToDomain(movieShowEntities), total, nil
}

func (r *MovieShowRepository) buildFilterClauses(filters domain.MovieShowFilters) ([]string, []interface{}, *utils.Error) {
	var whereClauses []string
	var args []interface{}
	argPos := 1
	var clause string
	var clauseArgs []interface{}
	var err *utils.Error

	// Handle movie ID filters
	if len(filters.MovieIDs) > 0 {
		clause, clauseArgs, argPos = r.buildInClause(filters.MovieIDs, argPos)
		whereClauses = append(whereClauses, fmt.Sprintf("movie_id IN (%s)", clause))
		args = append(args, clauseArgs...)
	}

	// Handle hall ID filters
	if len(filters.HallIDs) > 0 {
		clause, clauseArgs, argPos = r.buildInClause(filters.HallIDs, argPos)
		whereClauses = append(whereClauses, fmt.Sprintf("hall_id IN (%s)", clause))
		args = append(args, clauseArgs...)
	}

	// Handle language filters
	if len(filters.Languages) > 0 {
		clause, clauseArgs, argPos = r.buildInClause(filters.Languages, argPos)
		whereClauses = append(whereClauses, fmt.Sprintf("language::text IN (%s)", clause))
		args = append(args, clauseArgs...)
	}

	// Handle date filter
	if filters.Date != "" {
		clause, clauseArgs, argPos, err = r.handleDateFilter(filters.StartTimeTo, argPos)
		if err != nil {
			return nil, nil, err
		}
		whereClauses = append(whereClauses, clause)
		args = append(args, clauseArgs...)
	}

	// Handle start time from filter
	if filters.StartTimeFrom != "" {
		clause, clauseArgs, argPos, err = r.handleStartTimeFromFilter(filters.StartTimeTo, argPos)
		if err != nil {
			return nil, nil, err
		}
		whereClauses = append(whereClauses, clause)
		args = append(args, clauseArgs...)
	}

	// Handle start time to filter
	if filters.StartTimeTo != "" {
		clause, clauseArgs, _, err = r.handleStartTimeToFilter(filters.StartTimeTo, argPos)
		if err != nil {
			return nil, nil, err
		}
		whereClauses = append(whereClauses, clause)
		args = append(args, clauseArgs...)
	}

	return whereClauses, args, nil
}

func (r *MovieShowRepository) buildInClause(values []string, startPos int) (string, []interface{}, int) {
	placeholders := make([]string, len(values))
	args := make([]interface{}, len(values))

	for i, val := range values {
		placeholders[i] = fmt.Sprintf("$%d", startPos+i)
		args[i] = val
	}

	return strings.Join(placeholders, ","), args, startPos + len(values)
}

// Handle date filter extraction to reduce complexity
func (r *MovieShowRepository) handleDateFilter(dateStr string, startPos int) (string, []interface{}, int, *utils.Error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", nil, startPos, utils.NewBadRequest("Invalid date format", err)
	}

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	clause := fmt.Sprintf("start_time >= $%d AND start_time < $%d", startPos, startPos+1)
	args := []interface{}{startOfDay, endOfDay}

	return clause, args, startPos + 2, nil
}

// Handle start time from filter extraction to reduce complexity
func (r *MovieShowRepository) handleStartTimeFromFilter(startTimeStr string, startPos int) (string, []interface{}, int, *utils.Error) {
	startTimeFrom, err := time.Parse("15:04", startTimeStr)
	if err != nil {
		return "", nil, startPos, utils.NewBadRequest("Invalid time format", err)
	}

	clause := fmt.Sprintf("EXTRACT(HOUR FROM start_time) * 60 + EXTRACT(MINUTE FROM start_time) >= $%d", startPos)
	minutes := startTimeFrom.Hour()*60 + startTimeFrom.Minute()
	args := []interface{}{minutes}

	return clause, args, startPos + 1, nil
}

// Handle start time to filter extraction to reduce complexity
func (r *MovieShowRepository) handleStartTimeToFilter(startTimeStr string, startPos int) (string, []interface{}, int, *utils.Error) {
	startTimeTo, err := time.Parse("15:04", startTimeStr)
	if err != nil {
		return "", nil, startPos, utils.NewBadRequest("Invalid time format", err)
	}

	clause := fmt.Sprintf("EXTRACT(HOUR FROM start_time) * 60 + EXTRACT(MINUTE FROM start_time) <= $%d", startPos)
	minutes := startTimeTo.Hour()*60 + startTimeTo.Minute()
	args := []interface{}{minutes}

	return clause, args, startPos + 1, nil
}

func (r *MovieShowRepository) getCount(ctx context.Context, whereClauses []string, args []interface{}) (int, *utils.Error) {
	countQuery := "SELECT COUNT(*) FROM movie_shows"
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

func (r *MovieShowRepository) buildGetAllQuery(whereClauses []string, args []interface{}, page, limit int) (string, []interface{}) {
	query := "SELECT id, movie_id, hall_id, start_time, language FROM movie_shows"
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}
	query += " ORDER BY start_time ASC LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)

	args = append(args, limit, (page-1)*limit)

	return query, args
}

func (r *MovieShowRepository) executeGetAllQuery(ctx context.Context, query string, args []interface{}) ([]entity.MovieShow, *utils.Error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, utils.ConvertError(err)
	}
	defer rows.Close()

	var movieShowEntities []entity.MovieShow
	for rows.Next() {
		var movieShow entity.MovieShow
		if err := rows.Scan(&movieShow.ID, &movieShow.MovieID, &movieShow.HallID, &movieShow.StartTime, &movieShow.Language); err != nil {
			return nil, utils.ConvertError(err)
		}
		movieShowEntities = append(movieShowEntities, movieShow)
	}

	if err := rows.Err(); err != nil {
		return nil, utils.ConvertError(err)
	}

	return movieShowEntities, nil
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
