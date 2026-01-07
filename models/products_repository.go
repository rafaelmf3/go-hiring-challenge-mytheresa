package models

import (
	"errors"

	"gorm.io/gorm"
)

type ProductsRepositoryInterface interface {
	GetAllProducts() ([]Product, error)
	GetProductByCode(code string) (*Product, error)
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

func (r *productsRepository) GetProductByCode(code string) (*Product, error) {
	var product Product
	if err := r.db.Preload("Variants").Preload("Category").
		Where("code = ?", code).
		First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &product, nil
}
