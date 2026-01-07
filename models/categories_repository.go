package models

import (
	"gorm.io/gorm"
)

// CategoriesRepository defines the interface for category operations
type CategoriesRepository interface {
}

type categoriesRepository struct {
	db *gorm.DB
}

func NewCategoriesRepository(db *gorm.DB) CategoriesRepository {
	return &categoriesRepository{
		db: db,
	}
}
