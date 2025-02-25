package repository

import (
	"time"
)

// GormReview представляет модель отзыва в базе данных
type GormReview struct {
	ID        uint      `gorm:"primaryKey"`     // Уникальный идентификатор отзыва
	MediaID   uint      `gorm:"not null"`       // ID медиа, на которое оставлен отзыв
	UserID    uint      `gorm:"not null"`       // ID пользователя, оставившего отзыв
	Content   string    `gorm:"not null"`       // Содержимое отзыва
	Rating    int       `gorm:"default:0"`      // Оценка отзыва
	CreatedAt time.Time `gorm:"autoCreateTime"` // Дата создания
	UpdatedAt time.Time `gorm:"autoUpdateTime"` // Дата обновления
}

// TableName указывает GORM использовать имя таблицы "review"
func (GormReview) TableName() string {
	return "review"
}
