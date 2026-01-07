package models

import (
	"gorm.io/gorm"
)

type ProductsRepositoryInterface interface {
	GetAllProducts() ([]Product, error)
}

type productsRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) ProductsRepositoryInterface {
	return &productsRepository{
		db: db,
	}
}

func (r *productsRepository) GetAllProducts() ([]Product, error) {
	var products []Product
	if err := r.db.Preload("Variants").Preload("Category").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}
