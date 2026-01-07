package catalog

import (
	"errors"
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
	"gorm.io/gorm"
)

type CatalogResponse struct {
	Products []ProductResponse `json:"products"`
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
	res, err := h.productsRepository.GetAllProducts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Map response
	products := make([]ProductResponse, len(res))
	for i, p := range res {
		var categoryName *string
		if p.Category != nil {
			categoryName = &p.Category.Name
		}
		products[i] = ProductResponse{
			Code:     p.Code,
			Price:    p.Price.InexactFloat64(),
			Category: categoryName,
		}
	}

	response := CatalogResponse{
		Products: products,
	}

	api.OKResponse(w, response)
}

func (h *CatalogHandler) HandleGetByCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "product code is required")
		return
	}

	product, err := h.productsRepository.GetProductByCode(code)
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
