package models

import (
	"gorm.io/gorm"
)

type CategoriesRepositoryInterface interface {
	GetAllCategories() ([]Category, error)
	CreateCategory(category *Category) error
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
func (r *categoriesRepository) GetAllCategories() ([]Category, error) {
	var categories []Category
	if err := r.db.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// CreateCategory creates a new category in the database
func (r *categoriesRepository) CreateCategory(category *Category) error {
	return r.db.Create(category).Error
}
