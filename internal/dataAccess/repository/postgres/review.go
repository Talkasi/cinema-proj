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

type ReviewRepository struct {
	db *pgxpool.Pool
}

func NewReviewRepository(db *pgxpool.Pool) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) GetAll(ctx context.Context, filters domain.ReviewFilters, page, limit int) ([]domain.Review, int, *utils.Error) {

	whereClauses, args, err := r.buildFilterClauses(filters)
	if err != nil {
		return nil, 0, err
	}

	total, err := r.getCount(ctx, whereClauses, args)
	if err != nil {
		return nil, 0, err
	}

	query, queryArgs := r.buildGetAllQuery(whereClauses, args, page, limit)

	reviewEntities, err := r.executeGetAllQuery(ctx, query, queryArgs)
	if err != nil {
		return nil, 0, err
	}

	return entity.ReviewsToDomain(reviewEntities), total, nil
}

func (r *ReviewRepository) buildFilterClauses(filters domain.ReviewFilters) ([]string, []interface{}, *utils.Error) {
	var whereClauses []string
	var args []interface{}
	argPos := 1

	if filters.MovieID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("movie_id = $%d", argPos))
		args = append(args, filters.MovieID)
		argPos++
	}

	if filters.UserID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("user_id = $%d", argPos))
		args = append(args, filters.UserID)
		argPos++
	}

	if filters.RatingMin > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("rating >= $%d", argPos))
		args = append(args, filters.RatingMin)
		argPos++
	}

	if filters.RatingMax > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("rating <= $%d", argPos))
		args = append(args, filters.RatingMax)
		argPos++
	}

	if filters.Comment != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("review_comment ILIKE $%d", argPos))
		args = append(args, "%"+filters.Comment+"%")
		argPos++
	}

	_ = argPos  // Mark as used to avoid linter error

	return whereClauses, args, nil
}

func (r *ReviewRepository) getCount(ctx context.Context, whereClauses []string, args []interface{}) (int, *utils.Error) {
	countQuery := "SELECT COUNT(*) FROM reviews"
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

func (r *ReviewRepository) buildGetAllQuery(whereClauses []string, args []interface{}, page, limit int) (string, []interface{}) {
	query := "SELECT id, movie_id, user_id, rating, review_comment FROM reviews"
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}
	query += " ORDER BY id DESC LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)

	args = append(args, limit, (page-1)*limit)

	return query, args
}

func (r *ReviewRepository) executeGetAllQuery(ctx context.Context, query string, args []interface{}) ([]entity.Review, *utils.Error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, utils.ConvertError(err)
	}
	defer rows.Close()

	var reviewEntities []entity.Review
	for rows.Next() {
		var review entity.Review
		var comment *string
		if err := rows.Scan(&review.ID, &review.MovieID, &review.UserID, &review.Rating, &comment); err != nil {
			return nil, utils.ConvertError(err)
		}
		if comment != nil {
			review.Comment = *comment
		}
		reviewEntities = append(reviewEntities, review)
	}

	if err := rows.Err(); err != nil {
		return nil, utils.ConvertError(err)
	}

	return reviewEntities, nil
}

func (r *ReviewRepository) CreateForMovie(ctx context.Context, movieId string, review domain.Review) (domain.Review, *utils.Error) {
	reviewEntity := entity.ReviewFromDomain(review)
	reviewEntity.MovieID = movieId

	query := "INSERT INTO reviews (movie_id, user_id, rating, review_comment) VALUES ($1, $2, $3, $4) RETURNING id"
	err := r.db.QueryRow(ctx, query, reviewEntity.MovieID, reviewEntity.UserID, reviewEntity.Rating, reviewEntity.Comment).Scan(&reviewEntity.ID)
	if err != nil {
		return domain.Review{}, utils.ConvertError(err)
	}

	return entity.ReviewToDomain(reviewEntity), nil
}

func (r *ReviewRepository) Update(ctx context.Context, id string, review domain.Review) (domain.Review, *utils.Error) {
	reviewEntity := entity.ReviewFromDomain(review)

	query := "UPDATE reviews SET rating = $1, review_comment = $2 WHERE id = $3 RETURNING id, movie_id, user_id, rating, review_comment"
	var updatedReview entity.Review
	err := r.db.QueryRow(ctx, query, reviewEntity.Rating, reviewEntity.Comment, id).Scan(
		&updatedReview.ID, &updatedReview.MovieID, &updatedReview.UserID, &updatedReview.Rating, &updatedReview.Comment,
	)
	if err != nil {
		return domain.Review{}, utils.ConvertError(err)
	}

	return entity.ReviewToDomain(updatedReview), nil
}

func (r *ReviewRepository) Delete(ctx context.Context, id string) *utils.Error {
	query := "DELETE FROM reviews WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return utils.ConvertError(err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return utils.NewNotFound("review not found", nil)
	}

	return nil
}
