package models

import (
	"context"

	"gorm.io/gorm"
)

type CategoriesRepositoryInterface interface {
	GetAllCategories(ctx context.Context) ([]Category, error)
	CreateCategory(ctx context.Context, category *Category) error
}

type categoriesRepository struct {
	db *gorm.DB
}

func NewCategoriesRepository(db *gorm.DB) CategoriesRepositoryInterface {
	return &categoriesRepository{
		db: db,
	}
}

// GetAllCategories retrieves all categories from the database
func (r *categoriesRepository) GetAllCategories(ctx context.Context) ([]Category, error) {
	var categories []Category
	if err := r.db.WithContext(ctx).Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// CreateCategory creates a new category in the database
func (r *categoriesRepository) CreateCategory(ctx context.Context, category *Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}
