package postgres

import (
	"alexdenkk/labs/internal/user/domain"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type repository struct {
	DB *gorm.DB
}

func New(db *gorm.DB) domain.UserRepository {
	return &repository{
		DB: db,
	}
}

func (repository *repository) Get(ctx context.Context, id uuid.UUID) (domain.User, error) {
	var user domain.User
	result := repository.DB.First(&user, id)
	return user, result.Error
}

func (repository *repository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	var user domain.User
	result := repository.DB.Where("email = ?", email).First(&user)
	return user, result.Error
}

func (repository *repository) Create(ctx context.Context, user domain.User) error {
	return repository.DB.Create(&user).Error
}

func (repository *repository) Update(ctx context.Context, user domain.User) error {
	return repository.DB.Save(&user).Error
}
