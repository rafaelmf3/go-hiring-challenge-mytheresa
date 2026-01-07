package models

import (
	"gorm.io/gorm"
)

type CategoriesRepositoryInterface interface {
	GetAllCategories() ([]Category, error)
}

type categoriesRepository struct {
	db *gorm.DB
}

func NewCategoriesRepository(db *gorm.DB) CategoriesRepositoryInterface {
	return &categoriesRepository{
		db: db,
	}
}

func (r *categoriesRepository) GetAllCategories() ([]Category, error) {
	var categories []Category
	if err := r.db.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}
