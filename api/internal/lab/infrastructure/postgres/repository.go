package postgres

import (
	"alexdenkk/labs/internal/lab/domain"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type repository struct {
	DB *gorm.DB
}

func New(db *gorm.DB) domain.LabRepository {
	return &repository{
		DB: db,
	}
}

func (repository *repository) Get(ctx context.Context, id uuid.UUID) (domain.Lab, error) {
	var lab domain.Lab
	result := repository.DB.First(&lab, id)
	return lab, result.Error
}

func (repository *repository) GetByUser(ctx context.Context, userID uuid.UUID) ([]domain.Lab, error) {
	var labs []domain.Lab
	result := repository.DB.Where("user_id = ?", userID).Find(&labs)
	return labs, result.Error
}

func (repository *repository) Create(ctx context.Context, lab domain.Lab) error {
	return repository.DB.Create(&lab).Error
}

func (repository *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return repository.DB.Delete(&domain.Lab{}, id).Error
}
