package catalog

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
	"gorm.io/gorm"
)

const (
	defaultOffset = 0
	defaultLimit  = 10
)

type CatalogResponse struct {
	Products []ProductResponse `json:"products"`
	Total    int64             `json:"total"`
}

type ProductResponse struct {
	Code     string  `json:"code"`
	Price    float64 `json:"price"`
	Category *string `json:"category,omitempty"`
}

type ProductDetailResponse struct {
	Code     string            `json:"code"`
	Price    float64           `json:"price"`
	Category *string           `json:"category,omitempty"`
	Variants []VariantResponse `json:"variants"`
}

type VariantResponse struct {
	Name  string  `json:"name"`
	SKU   string  `json:"sku"`
	Price float64 `json:"price"`
}

type CatalogHandler struct {
	productsRepository models.ProductsRepositoryInterface
}

func NewCatalogHandler(r models.ProductsRepositoryInterface) *CatalogHandler {
	return &CatalogHandler{
		productsRepository: r,
	}
}

func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	offset := defaultOffset
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	limit := defaultLimit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			if parsed < 1 {
				limit = 1
			} else if parsed > 100 {
				limit = 100
			} else {
				limit = parsed
			}
		}
	}

	var categoryCode *string
	if cat := r.URL.Query().Get("category"); cat != "" {
		categoryCode = &cat
	}

	var priceLessThan *float64
	if priceStr := r.URL.Query().Get("price_less_than"); priceStr != "" {
		if parsed, err := strconv.ParseFloat(priceStr, 64); err == nil && parsed > 0 {
			priceLessThan = &parsed
		}
	}

	products, total, err := h.productsRepository.GetProductsWithFilters(ctx, offset, limit, categoryCode, priceLessThan)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Map response
	productsResponse := make([]ProductResponse, len(products))
	for i, p := range products {
		var categoryName *string
		if p.Category != nil {
			categoryName = &p.Category.Name
		}
		productsResponse[i] = ProductResponse{
			Code:     p.Code,
			Price:    p.Price.InexactFloat64(),
			Category: categoryName,
		}
	}

	response := CatalogResponse{
		Products: productsResponse,
		Total:    total,
	}

	api.OKResponse(w, response)
}

func (h *CatalogHandler) HandleGetByCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	code := r.PathValue("code")
	if code == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "product code is required")
		return
	}

	product, err := h.productsRepository.GetProductByCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			api.ErrorResponse(w, http.StatusNotFound, "product not found")
			return
		}
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	variants := make([]VariantResponse, len(product.Variants))
	for i, v := range product.Variants {
		price := product.Price
		if !v.Price.IsZero() {
			price = v.Price
		}
		variants[i] = VariantResponse{
			Name:  v.Name,
			SKU:   v.SKU,
			Price: price.InexactFloat64(),
		}
	}

	var categoryName *string
	if product.Category != nil {
		categoryName = &product.Category.Name
	}

	response := ProductDetailResponse{
		Code:     product.Code,
		Price:    product.Price.InexactFloat64(),
		Category: categoryName,
		Variants: variants,
	}

	api.OKResponse(w, response)
}
