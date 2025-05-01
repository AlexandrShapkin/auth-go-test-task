package repositories

import (
	"context"

	"github.com/AlexandrShapkin/auth-go-test-task/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormUserRepo struct {
	DB *gorm.DB
}

type UserRepo interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

func NewUserRepo(db *gorm.DB) UserRepo {
	return &GormUserRepo{
		DB: db,
	}
}

func (r *GormUserRepo) Create(ctx context.Context, user *models.User) error {
	return r.DB.WithContext(ctx).Create(user).Error
}

func (r *GormUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.DB.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepo) Update(ctx context.Context, user *models.User) error {
	return r.DB.WithContext(ctx).Save(user).Error
}

func (r *GormUserRepo) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return r.DB.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error
}
