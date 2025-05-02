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

// Репозиторий сущности пользователя, описывающий основные необходимые CRUD операции
type UserRepo interface {
	// Создает новую запись пользователя. Передавать обьект созданный при помощи NewUser, генерация uuid происходит в нем
	Create(ctx context.Context, user *models.User) error
	// Находит запись пользователя по его uuid
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	// Обертка вокруг FindByID, но не требует предварительного парсинга uuid
	FindByIDString(ctx context.Context, idString string) (*models.User, error)
	// Обновляет запись пользователя исходя из его uuid переданного в обьекте (обновляет все поля)
	Update(ctx context.Context, user *models.User) error
	// Удаляет пользователя по его uuid
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

// Конструктор для создания экземпляра репозитория. Более предпочтительно, чем создание из голой структуры
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
	err := r.DB.WithContext(ctx).First(&user, "user_id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepo) FindByIDString(ctx context.Context, idString string) (*models.User, error) {
	id, err := uuid.Parse(idString)
	if err != nil {
		return nil, err
	}

	user, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *GormUserRepo) Update(ctx context.Context, user *models.User) error {
	return r.DB.WithContext(ctx).Save(user).Error
}

func (r *GormUserRepo) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return r.DB.WithContext(ctx).Delete(&models.User{}, "user_id = ?", id).Error
}
