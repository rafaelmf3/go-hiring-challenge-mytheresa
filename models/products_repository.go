package models

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type ProductsRepositoryInterface interface {
	// Deprecated: use GetProductsWithFilters instead
	GetAllProducts() ([]Product, error)
	GetProductByCode(ctx context.Context, code string) (*Product, error)
	GetProductsWithFilters(ctx context.Context, offset, limit int, categoryCode *string, priceLessThan *float64) ([]Product, int64, error)
}

type productsRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) ProductsRepositoryInterface {
	return &productsRepository{
		db: db,
	}
}

// Deprecated: use GetProductsWithFilters instead
func (r *productsRepository) GetAllProducts() ([]Product, error) {
	var products []Product
	if err := r.db.Preload("Variants").Preload("Category").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// GetProductByCode retrieves a product by its code along with its variants and category
func (r *productsRepository) GetProductByCode(ctx context.Context, code string) (*Product, error) {
	var product Product
	if err := r.db.WithContext(ctx).Preload("Variants").Preload("Category").
		Where("code = ?", code).
		First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &product, nil
}

// GetProductsWithFilters retrieves products with optional filters for category code and price less than a specified value, along with pagination
func (r *productsRepository) GetProductsWithFilters(ctx context.Context, offset, limit int, categoryCode *string, priceLessThan *float64) ([]Product, int64, error) {
	var products []Product
	var total int64

	query := r.db.WithContext(ctx).Model(&Product{})

	if categoryCode != nil && *categoryCode != "" {
		query = query.Joins("JOIN categories ON products.category_id = categories.id").
			Where("categories.code = ?", *categoryCode)
	}

	if priceLessThan != nil {
		query = query.Where("price < ?", *priceLessThan)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Category").Preload("Variants").
		Offset(offset).
		Limit(limit).
		Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}
