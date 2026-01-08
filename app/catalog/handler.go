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
	minLimit      = 1
	maxLimit      = 100
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

	offset, err := parseOffset(r.URL.Query().Get("offset"))
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	limit, err := parseLimit(r.URL.Query().Get("limit"))
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var categoryCode *string
	if cat := r.URL.Query().Get("category"); cat != "" {
		categoryCode = &cat
	}

	priceLessThan, err := parsePriceLessThan(r.URL.Query().Get("price_less_than"))
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	products, total, err := h.productsRepository.GetProductsWithFilters(ctx, offset, limit, categoryCode, priceLessThan)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

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

func parseOffset(query string) (int, error) {
	if query == "" {
		return defaultOffset, nil
	}

	parsed, err := strconv.Atoi(query)
	if err != nil {
		return 0, errors.New("offset must be a valid integer")
	}

	if parsed < 0 {
		return 0, errors.New("offset must be non-negative")
	}

	return parsed, nil
}

func parseLimit(query string) (int, error) {
	if query == "" {
		return defaultLimit, nil
	}

	parsed, err := strconv.Atoi(query)
	if err != nil {
		return 0, errors.New("limit must be a valid integer")
	}

	limit := max(minLimit, min(parsed, maxLimit))
	return limit, nil
}

func parsePriceLessThan(query string) (*float64, error) {
	if query == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseFloat(query, 64)
	if err != nil {
		return nil, errors.New("price_less_than must be a valid number")
	}

	if parsed <= 0 {
		return nil, errors.New("price_less_than must be greater than 0")
	}

	return &parsed, nil
}
