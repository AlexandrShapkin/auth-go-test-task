package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Модель сущности пользователя
type User struct {
	// GUID (uuid) записи пользователя. Генерируется при создании через NewUser
	UserID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	// Email пользователя
	Email        string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	// Хешированный при помощи bcrypt пароль
	Password     string    `gorm:"type:varchar(60);not null"`
	// Хешированный при помощи bcrupt refresh токен
	RefreshToken string    `gorm:"type:varchar(60)"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// Конструктор нового обьекта модели пользователя. Наиболее предпочтителен, так как генерирует еще и его GUID
func NewUser(email string, hashedPassword string) *User {
	return &User{
		UserID:   uuid.New(),
		Email:    email,
		Password: hashedPassword,
	}
}
