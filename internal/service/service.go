package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/watchlist-kata/protos/review"
	"github.com/watchlist-kata/review/internal/repository"
)

type ReviewService struct {
	review.UnimplementedReviewServiceServer
	repo   repository.Repository
	logger *slog.Logger
}

func NewReviewService(repo repository.Repository, logger *slog.Logger) *ReviewService {
	return &ReviewService{
		repo:   repo,
		logger: logger,
	}
}

func (s *ReviewService) checkContextCancelled(ctx context.Context, method string) error {
	select {
	case <-ctx.Done():
		s.logger.ErrorContext(ctx, fmt.Sprintf("%s operation canceled", method), slog.Any("error", ctx.Err()))
		return ctx.Err()
	default:
		return nil
	}
}

func (s *ReviewService) Create(ctx context.Context, req *review.CreateReviewRequest) (*review.CreateReviewResponse, error) {
	if err := s.checkContextCancelled(ctx, "Create"); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	if req.Rating < 1 || req.Rating > 10 {
		s.logger.WarnContext(ctx, "invalid rating: must be between 1 and 10")
		return nil, status.Errorf(codes.InvalidArgument, "Rating must be between 1 and 10")
	}

	gormReview := &repository.GormReview{
		MediaID: uint(req.MediaId),
		UserID:  uint(req.UserId),
		Content: req.Content,
		Rating:  int(req.Rating),
	}

	if err := s.repo.Create(ctx, gormReview); err != nil {
		s.logger.ErrorContext(ctx, fmt.Sprintf("failed to create review for media ID: %d and user ID: %d", req.MediaId, req.UserId), slog.Any("error", err))
		return nil, status.Errorf(codes.Internal, "Failed to create review: %v", err)
	}

	protoReview := ConvertToProtoReview(gormReview)

	s.logger.InfoContext(ctx, fmt.Sprintf("review created successfully for media ID: %d and user ID: %d", req.MediaId, req.UserId))
	return &review.CreateReviewResponse{
		Review: protoReview,
	}, nil
}

func (s *ReviewService) GetByID(ctx context.Context, req *review.GetReviewRequest) (*review.GetReviewResponse, error) {
	if err := s.checkContextCancelled(ctx, "GetByID"); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	gormReview, err := s.repo.GetByID(ctx, uint(req.Id))
	if err != nil {
		if errors.Is(err, repository.ErrReviewNotFound) {
			s.logger.WarnContext(ctx, fmt.Sprintf("review not found with ID: %d", req.Id))
			return nil, status.Errorf(codes.NotFound, "Review not found: %v", err)
		}
		s.logger.ErrorContext(ctx, fmt.Sprintf("failed to get review by ID: %d", req.Id), slog.Any("error", err))
		return nil, status.Errorf(codes.Internal, "Failed to get review: %v", err)
	}

	protoReview := ConvertToProtoReview(gormReview)

	s.logger.InfoContext(ctx, fmt.Sprintf("review fetched successfully with ID: %d", req.Id))
	return &review.GetReviewResponse{
		Review: protoReview,
	}, nil
}

func (s *ReviewService) Update(ctx context.Context, req *review.UpdateReviewRequest) (*review.UpdateReviewResponse, error) {
	if err := s.checkContextCancelled(ctx, "Update"); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	gormReview, err := s.repo.GetByID(ctx, uint(req.Id))
	if err != nil {
		if errors.Is(err, repository.ErrReviewNotFound) {
			s.logger.WarnContext(ctx, fmt.Sprintf("review not found with ID: %d", req.Id))
			return nil, status.Errorf(codes.NotFound, "Review not found: %v", err)
		}
		s.logger.ErrorContext(ctx, fmt.Sprintf("failed to get review for update with ID: %d", req.Id), slog.Any("error", err))
		return nil, status.Errorf(codes.Internal, "Failed to get review: %v", err)
	}

	if req.Content != "" {
		gormReview.Content = req.Content
	}

	if req.Rating != 0 {
		if req.Rating < 1 || req.Rating > 10 {
			s.logger.WarnContext(ctx, "invalid rating: must be between 1 and 10")
			return nil, status.Errorf(codes.InvalidArgument, "Rating must be between 1 and 10")
		}
		gormReview.Rating = int(req.Rating)
	}

	if err := s.repo.Update(ctx, gormReview); err != nil {
		s.logger.ErrorContext(ctx, fmt.Sprintf("failed to update review with ID: %d", req.Id), slog.Any("error", err))
		return nil, status.Errorf(codes.Internal, "Failed to update review: %v", err)
	}

	protoReview := ConvertToProtoReview(gormReview)

	s.logger.InfoContext(ctx, fmt.Sprintf("review updated successfully with ID: %d", req.Id))
	return &review.UpdateReviewResponse{
		Review: protoReview,
	}, nil
}

func (s *ReviewService) Delete(ctx context.Context, req *review.DeleteReviewRequest) (*review.DeleteReviewResponse, error) {
	if err := s.checkContextCancelled(ctx, "Delete"); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	_, err := s.repo.GetByID(ctx, uint(req.Id))
	if err != nil {
		if errors.Is(err, repository.ErrReviewNotFound) {
			s.logger.WarnContext(ctx, fmt.Sprintf("review not found with ID: %d", req.Id))
			return nil, status.Errorf(codes.NotFound, "Review not found: %v", err)
		}
		s.logger.ErrorContext(ctx, fmt.Sprintf("failed to check review existence with ID: %d", req.Id), slog.Any("error", err))
		return nil, status.Errorf(codes.Internal, "Failed to check review existence: %v", err)
	}

	if err := s.repo.Delete(ctx, uint(req.Id)); err != nil {
		s.logger.ErrorContext(ctx, fmt.Sprintf("failed to delete review with ID: %d", req.Id), slog.Any("error", err))
		return nil, status.Errorf(codes.Internal, "Failed to delete review: %v", err)
	}

	s.logger.InfoContext(ctx, fmt.Sprintf("review deleted successfully with ID: %d", req.Id))
	return &review.DeleteReviewResponse{
		Success: true,
	}, nil
}

func (s *ReviewService) GetAll(ctx context.Context, req *review.GetAllReviewsRequest) (*review.GetAllReviewsResponse, error) {
	if err := s.checkContextCancelled(ctx, "GetAll"); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	gormReviews, err := s.repo.GetAll(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get all reviews", slog.Any("error", err))
		return nil, status.Errorf(codes.Internal, "Failed to get reviews: %v", err)
	}

	protoReviews := make([]*review.Review, 0, len(gormReviews))
	for i := range gormReviews {
		protoReviews = append(protoReviews, ConvertToProtoReview(&gormReviews[i]))
	}

	s.logger.InfoContext(ctx, "all reviews fetched successfully")
	return &review.GetAllReviewsResponse{
		Reviews: protoReviews,
	}, nil
}

func (s *ReviewService) GetByRating(ctx context.Context, req *review.GetByRatingRequest) (*review.GetByRatingResponse, error) {
	if err := s.checkContextCancelled(ctx, "GetByRating"); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	if req.Rating < 1 || req.Rating > 10 {
		s.logger.WarnContext(ctx, "invalid rating: must be between 1 and 10")
		return nil, status.Errorf(codes.InvalidArgument, "Rating must be between 1 and 10")
	}

	gormReviews, err := s.repo.GetByRating(ctx, int(req.Rating))
	if err != nil {
		s.logger.ErrorContext(ctx, fmt.Sprintf("failed to get reviews by rating: %d", req.Rating), slog.Any("error", err))
		return nil, status.Errorf(codes.Internal, "Failed to get reviews by rating: %v", err)
	}

	protoReviews := make([]*review.Review, 0, len(gormReviews))
	for i := range gormReviews {
		protoReviews = append(protoReviews, ConvertToProtoReview(&gormReviews[i]))
	}

	s.logger.InfoContext(ctx, fmt.Sprintf("reviews fetched successfully by rating: %d", req.Rating))
	return &review.GetByRatingResponse{
		Reviews: protoReviews,
	}, nil
}

func (s *ReviewService) GetByUser(ctx context.Context, req *review.GetByUserRequest) (*review.GetByUserResponse, error) {
	if err := s.checkContextCancelled(ctx, "GetByUser"); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	gormReviews, err := s.repo.GetByUser(ctx, uint(req.UserId))
	if err != nil {
		s.logger.ErrorContext(ctx, fmt.Sprintf("failed to get reviews by user ID: %d", req.UserId), slog.Any("error", err))
		return nil, status.Errorf(codes.Internal, "Failed to get reviews by user: %v", err)
	}

	protoReviews := make([]*review.Review, 0, len(gormReviews))
	for i := range gormReviews {
		protoReviews = append(protoReviews, ConvertToProtoReview(&gormReviews[i]))
	}

	s.logger.InfoContext(ctx, fmt.Sprintf("reviews fetched successfully by user ID: %d", req.UserId))
	return &review.GetByUserResponse{
		Reviews: protoReviews,
	}, nil
}

func (s *ReviewService) GetByMedia(ctx context.Context, req *review.GetByMediaRequest) (*review.GetByMediaResponse, error) {
	if err := s.checkContextCancelled(ctx, "GetByMedia"); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	gormReviews, err := s.repo.GetByMedia(ctx, uint(req.MediaId))
	if err != nil {
		s.logger.ErrorContext(ctx, fmt.Sprintf("failed to get reviews by media ID: %d", req.MediaId), slog.Any("error", err))
		return nil, status.Errorf(codes.Internal, "Failed to get reviews by media: %v", err)
	}

	protoReviews := make([]*review.Review, 0, len(gormReviews))
	for i := range gormReviews {
		protoReviews = append(protoReviews, ConvertToProtoReview(&gormReviews[i]))
	}

	s.logger.InfoContext(ctx, fmt.Sprintf("reviews fetched successfully by media ID: %d", req.MediaId))
	return &review.GetByMediaResponse{
		Reviews: protoReviews,
	}, nil
}

func ConvertToProtoReview(gormReview *repository.GormReview) *review.Review {
	return &review.Review{
		Id:        int64(gormReview.ID),
		MediaId:   int64(gormReview.MediaID),
		UserId:    int64(gormReview.UserID),
		Content:   gormReview.Content,
		Rating:    int32(gormReview.Rating),
		CreatedAt: gormReview.CreatedAt.Format(time.RFC3339),
		UpdatedAt: gormReview.UpdatedAt.Format(time.RFC3339),
	}
}
