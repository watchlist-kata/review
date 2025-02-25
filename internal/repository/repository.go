package repository

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/watchlist-kata/review/internal/config"
)

// Repository определяет интерфейс для работы с отзывами
type Repository interface {
	Create(review *GormReview) error
	GetByID(id uint) (*GormReview, error)
	Update(review *GormReview) error
	Delete(id uint) error
	GetAll() ([]GormReview, error)
	GetByRating(rating int) ([]GormReview, error)
	GetByUser(userID uint) ([]GormReview, error)
	GetByMedia(mediaID uint) ([]GormReview, error)
}

// PostgresRepository представляет реализацию репозитория для работы с отзывами в PostgreSQL
type PostgresRepository struct {
	db *gorm.DB // Указатель на экземпляр базы данных
}

// NewPostgresRepository создает новый экземпляр PostgresRepository
func NewPostgresRepository(cfg *config.Config) (*PostgresRepository, error) {
	// Подключение к базе данных
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &PostgresRepository{db: db}, nil
}

// Create создает новый отзыв в базе данных
func (r *PostgresRepository) Create(review *GormReview) error {
	return r.db.Create(review).Error
}

// GetByID получает отзыв по ID
func (r *PostgresRepository) GetByID(id uint) (*GormReview, error) {
	var review GormReview
	if err := r.db.First(&review, id).Error; err != nil {
		return nil, err // Возвращаем ошибку, если отзыв не найден
	}
	return &review, nil
}

// Update обновляет существующий отзыв
func (r *PostgresRepository) Update(review *GormReview) error {
	if err := r.db.Save(review).Error; err != nil {
		return err // Возвращаем ошибку при обновлении
	}
	return nil
}

// Delete удаляет отзыв по ID
func (r *PostgresRepository) Delete(id uint) error {
	if err := r.db.Delete(&GormReview{}, id).Error; err != nil {
		return err // Возвращаем ошибку при удалении
	}
	return nil
}

// GetAll получает все отзывы
func (r *PostgresRepository) GetAll() ([]GormReview, error) {
	var reviews []GormReview
	if err := r.db.Find(&reviews).Error; err != nil {
		return nil, err // Возвращаем ошибку при получении всех отзывов
	}
	return reviews, nil
}

// GetByRating получает отзывы по рейтингу
func (r *PostgresRepository) GetByRating(rating int) ([]GormReview, error) {
	var reviews []GormReview
	if err := r.db.Where("rating = ?", rating).Find(&reviews).Error; err != nil {
		return nil, err // Возвращаем ошибку при получении отзывов по рейтингу
	}
	return reviews, nil
}

// GetByUser получает отзывы пользователя по его ID
func (r *PostgresRepository) GetByUser(userID uint) ([]GormReview, error) {
	var reviews []GormReview
	if err := r.db.Where("user_id = ?", userID).Find(&reviews).Error; err != nil {
		return nil, err // Возвращаем ошибку при получении отзывов пользователя
	}
	return reviews, nil
}

// GetByMedia получает отзывы по ID медиа-контента
func (r *PostgresRepository) GetByMedia(mediaID uint) ([]GormReview, error) {
	var reviews []GormReview
	if err := r.db.Where("media_id = ?", mediaID).Find(&reviews).Error; err != nil {
		return nil, err // Возвращаем ошибку при получении отзывов по медиа-контенту
	}
	return reviews, nil
}
