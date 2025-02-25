package service

import (
	"context"
	"time"

	"github.com/watchlist-kata/protos/review"
	"github.com/watchlist-kata/review/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ReviewService реализует gRPC-сервис для работы с отзывами
type ReviewService struct {
	review.UnimplementedReviewServiceServer
	repo repository.Repository
}

// NewReviewService создает новый экземпляр ReviewService
func NewReviewService(repo repository.Repository) *ReviewService {
	return &ReviewService{
		repo: repo,
	}
}

// ConvertToProtoReview преобразует GormReview в proto Review
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

// Create создает новый отзыв
func (s *ReviewService) Create(ctx context.Context, req *review.CreateReviewRequest) (*review.CreateReviewResponse, error) {
	// Проверка валидности рейтинга
	if req.Rating < 1 || req.Rating > 10 {
		return nil, status.Errorf(codes.InvalidArgument, "Rating must be between 1 and 10")
	}

	// Создаем GORM модель из запроса
	gormReview := &repository.GormReview{
		MediaID: uint(req.MediaId),
		UserID:  uint(req.UserId),
		Content: req.Content,
		Rating:  int(req.Rating),
	}

	// Сохраняем в базу данных
	if err := s.repo.Create(gormReview); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create review: %v", err)
	}

	// Преобразуем GORM модель в proto модель для ответа
	protoReview := ConvertToProtoReview(gormReview)

	return &review.CreateReviewResponse{
		Review: protoReview,
	}, nil
}

// GetByID получает отзыв по его ID
func (s *ReviewService) GetByID(ctx context.Context, req *review.GetReviewRequest) (*review.GetReviewResponse, error) {
	// Получаем отзыв из репозитория
	gormReview, err := s.repo.GetByID(uint(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Review not found: %v", err)
	}

	// Преобразуем GORM модель в proto модель для ответа
	protoReview := ConvertToProtoReview(gormReview)

	return &review.GetReviewResponse{
		Review: protoReview,
	}, nil
}

// Update обновляет существующий отзыв
func (s *ReviewService) Update(ctx context.Context, req *review.UpdateReviewRequest) (*review.UpdateReviewResponse, error) {
	// Получаем существующий отзыв
	gormReview, err := s.repo.GetByID(uint(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Review not found: %v", err)
	}

	// Проверяем и обновляем поля, если они предоставлены
	if req.Content != "" {
		gormReview.Content = req.Content
	}

	if req.Rating != 0 {
		// Проверка валидности рейтинга
		if req.Rating < 1 || req.Rating > 10 {
			return nil, status.Errorf(codes.InvalidArgument, "Rating must be between 1 and 10")
		}
		gormReview.Rating = int(req.Rating)
	}

	// Сохраняем обновленный отзыв
	if err := s.repo.Update(gormReview); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update review: %v", err)
	}

	// Преобразуем GORM модель в proto модель для ответа
	protoReview := ConvertToProtoReview(gormReview)

	return &review.UpdateReviewResponse{
		Review: protoReview,
	}, nil
}

// Delete удаляет отзыв по его ID
func (s *ReviewService) Delete(ctx context.Context, req *review.DeleteReviewRequest) (*review.DeleteReviewResponse, error) {
	// Проверяем, существует ли отзыв
	_, err := s.repo.GetByID(uint(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Review not found: %v", err)
	}

	// Удаляем отзыв
	if err := s.repo.Delete(uint(req.Id)); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete review: %v", err)
	}

	return &review.DeleteReviewResponse{
		Success: true,
	}, nil
}

// GetAll получает все отзывы
func (s *ReviewService) GetAll(ctx context.Context, req *review.GetAllReviewsRequest) (*review.GetAllReviewsResponse, error) {
	// Получаем все отзывы из репозитория
	gormReviews, err := s.repo.GetAll()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get reviews: %v", err)
	}

	// Преобразуем GORM модели в proto модели для ответа
	protoReviews := make([]*review.Review, 0, len(gormReviews))
	for i := range gormReviews {
		protoReviews = append(protoReviews, ConvertToProtoReview(&gormReviews[i]))
	}

	return &review.GetAllReviewsResponse{
		Reviews: protoReviews,
	}, nil
}

// GetByRating получает отзывы по заданному рейтингу
func (s *ReviewService) GetByRating(ctx context.Context, req *review.GetByRatingRequest) (*review.GetByRatingResponse, error) {
	// Проверка валидности рейтинга
	if req.Rating < 1 || req.Rating > 10 {
		return nil, status.Errorf(codes.InvalidArgument, "Rating must be between 1 and 10")
	}

	// Получаем отзывы с заданным рейтингом
	gormReviews, err := s.repo.GetByRating(int(req.Rating))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get reviews by rating: %v", err)
	}

	// Преобразуем GORM модели в proto модели для ответа
	protoReviews := make([]*review.Review, 0, len(gormReviews))
	for i := range gormReviews {
		protoReviews = append(protoReviews, ConvertToProtoReview(&gormReviews[i]))
	}

	return &review.GetByRatingResponse{
		Reviews: protoReviews,
	}, nil
}

// GetByUser получает отзывы конкретного пользователя
func (s *ReviewService) GetByUser(ctx context.Context, req *review.GetByUserRequest) (*review.GetByUserResponse, error) {
	// Получаем отзывы пользователя
	gormReviews, err := s.repo.GetByUser(uint(req.UserId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get reviews by user: %v", err)
	}

	// Преобразуем GORM модели в proto модели для ответа
	protoReviews := make([]*review.Review, 0, len(gormReviews))
	for i := range gormReviews {
		protoReviews = append(protoReviews, ConvertToProtoReview(&gormReviews[i]))
	}

	return &review.GetByUserResponse{
		Reviews: protoReviews,
	}, nil
}

// GetByMedia получает отзывы для конкретного медиа-контента
func (s *ReviewService) GetByMedia(ctx context.Context, req *review.GetByMediaRequest) (*review.GetByMediaResponse, error) {
	// Получаем отзывы для медиа-контента
	gormReviews, err := s.repo.GetByMedia(uint(req.MediaId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get reviews by media: %v", err)
	}

	// Преобразуем GORM модели в proto модели для ответа
	protoReviews := make([]*review.Review, 0, len(gormReviews))
	for i := range gormReviews {
		protoReviews = append(protoReviews, ConvertToProtoReview(&gormReviews[i]))
	}

	return &review.GetByMediaResponse{
		Reviews: protoReviews,
	}, nil
}
