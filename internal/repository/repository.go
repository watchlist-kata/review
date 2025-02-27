package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/watchlist-kata/review/internal/config"
	"github.com/watchlist-kata/review/pkg/utils"
	"gorm.io/gorm"
	"log/slog"
)

var (
	// ErrReviewNotFound возвращается, когда отзыв не найден
	ErrReviewNotFound = errors.New("review not found")
)

type Repository interface {
	Create(ctx context.Context, review *GormReview) error
	GetByID(ctx context.Context, id uint) (*GormReview, error)
	Update(ctx context.Context, review *GormReview) error
	Delete(ctx context.Context, id uint) error
	GetAll(ctx context.Context) ([]GormReview, error)
	GetByRating(ctx context.Context, rating int) ([]GormReview, error)
	GetByUser(ctx context.Context, userID uint) ([]GormReview, error)
	GetByMedia(ctx context.Context, mediaID uint) ([]GormReview, error)
}

type PostgresRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewPostgresRepository создает новый экземпляр PostgresRepository
func NewPostgresRepository(cfg *config.Config, logger *slog.Logger) (*PostgresRepository, error) {
	db, err := utils.ConnectToDatabase(cfg)
	if err != nil {
		logger.Error("failed to connect to database", slog.Any("error", err))
		return nil, err
	}

	return &PostgresRepository{db: db, logger: logger}, nil
}

func (r *PostgresRepository) Create(ctx context.Context, review *GormReview) error {
	select {
	case <-ctx.Done():
		r.logger.ErrorContext(ctx, fmt.Sprintf("Create operation canceled for review with media ID: %d and user ID: %d", review.MediaID, review.UserID), slog.Any("error", ctx.Err()))
		return ctx.Err()
	default:
	}

	if err := r.db.Create(review).Error; err != nil {
		r.logger.ErrorContext(ctx, fmt.Sprintf("failed to create review for media ID: %d and user ID: %d", review.MediaID, review.UserID), slog.Any("error", err))
		return err
	}

	r.logger.InfoContext(ctx, fmt.Sprintf("review created successfully for media ID: %d and user ID: %d", review.MediaID, review.UserID))
	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uint) (*GormReview, error) {
	select {
	case <-ctx.Done():
		r.logger.ErrorContext(ctx, fmt.Sprintf("GetByID operation canceled for review ID: %d", id), slog.Any("error", ctx.Err()))
		return nil, ctx.Err()
	default:
	}

	var review GormReview
	if err := r.db.First(&review, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.WarnContext(ctx, fmt.Sprintf("review not found with ID: %d", id))
			return nil, ErrReviewNotFound
		}
		r.logger.ErrorContext(ctx, fmt.Sprintf("failed to get review by ID: %d", id), slog.Any("error", err))
		return nil, err
	}

	r.logger.InfoContext(ctx, fmt.Sprintf("review fetched successfully with ID: %d", id))
	return &review, nil
}

func (r *PostgresRepository) Update(ctx context.Context, review *GormReview) error {
	select {
	case <-ctx.Done():
		r.logger.ErrorContext(ctx, fmt.Sprintf("Update operation canceled for review ID: %d", review.ID), slog.Any("error", ctx.Err()))
		return ctx.Err()
	default:
	}

	if err := r.db.Save(review).Error; err != nil {
		r.logger.ErrorContext(ctx, fmt.Sprintf("failed to update review with ID: %d", review.ID), slog.Any("error", err))
		return err
	}

	r.logger.InfoContext(ctx, fmt.Sprintf("review updated successfully with ID: %d", review.ID))
	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id uint) error {
	select {
	case <-ctx.Done():
		r.logger.ErrorContext(ctx, fmt.Sprintf("Delete operation canceled for review ID: %d", id), slog.Any("error", ctx.Err()))
		return ctx.Err()
	default:
	}

	if err := r.db.Delete(&GormReview{}, id).Error; err != nil {
		r.logger.ErrorContext(ctx, fmt.Sprintf("failed to delete review with ID: %d", id), slog.Any("error", err))
		return err
	}

	r.logger.InfoContext(ctx, fmt.Sprintf("review deleted successfully with ID: %d", id))
	return nil
}

func (r *PostgresRepository) GetAll(ctx context.Context) ([]GormReview, error) {
	select {
	case <-ctx.Done():
		r.logger.ErrorContext(ctx, "GetAll operation canceled", slog.Any("error", ctx.Err()))
		return nil, ctx.Err()
	default:
	}

	var reviews []GormReview
	if err := r.db.Find(&reviews).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to get all reviews", slog.Any("error", err))
		return nil, err
	}

	r.logger.InfoContext(ctx, "all reviews fetched successfully")
	return reviews, nil
}

func (r *PostgresRepository) GetByRating(ctx context.Context, rating int) ([]GormReview, error) {
	select {
	case <-ctx.Done():
		r.logger.ErrorContext(ctx, fmt.Sprintf("GetByRating operation canceled for rating: %d", rating), slog.Any("error", ctx.Err()))
		return nil, ctx.Err()
	default:
	}

	var reviews []GormReview
	if err := r.db.Where("rating = ?", rating).Find(&reviews).Error; err != nil {
		r.logger.ErrorContext(ctx, fmt.Sprintf("failed to get reviews by rating: %d", rating), slog.Any("error", err))
		return nil, err
	}

	r.logger.InfoContext(ctx, fmt.Sprintf("reviews fetched successfully by rating: %d", rating))
	return reviews, nil
}

func (r *PostgresRepository) GetByUser(ctx context.Context, userID uint) ([]GormReview, error) {
	select {
	case <-ctx.Done():
		r.logger.ErrorContext(ctx, fmt.Sprintf("GetByUser operation canceled for user ID: %d", userID), slog.Any("error", ctx.Err()))
		return nil, ctx.Err()
	default:
	}

	var reviews []GormReview
	if err := r.db.Where("user_id = ?", userID).Find(&reviews).Error; err != nil {
		r.logger.ErrorContext(ctx, fmt.Sprintf("failed to get reviews by user ID: %d", userID), slog.Any("error", err))
		return nil, err
	}

	r.logger.InfoContext(ctx, fmt.Sprintf("reviews fetched successfully by user ID: %d", userID))
	return reviews, nil
}

func (r *PostgresRepository) GetByMedia(ctx context.Context, mediaID uint) ([]GormReview, error) {
	select {
	case <-ctx.Done():
		r.logger.ErrorContext(ctx, fmt.Sprintf("GetByMedia operation canceled for media ID: %d", mediaID), slog.Any("error", ctx.Err()))
		return nil, ctx.Err()
	default:
	}

	var reviews []GormReview
	if err := r.db.Where("media_id = ?", mediaID).Find(&reviews).Error; err != nil {
		r.logger.ErrorContext(ctx, fmt.Sprintf("failed to get reviews by media ID: %d", mediaID), slog.Any("error", err))
		return nil, err
	}

	r.logger.InfoContext(ctx, fmt.Sprintf("reviews fetched successfully by media ID: %d", mediaID))
	return reviews, nil
}
