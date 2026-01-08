package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type mockProductsRepository struct {
	products      []models.Product
	total         int64
	getAllErr     error
	getByCodeErr  error
	getByCodeProd *models.Product
}

func (m *mockProductsRepository) GetAllProducts() ([]models.Product, error) {
	return m.products, m.getAllErr
}

func (m *mockProductsRepository) GetProductByCode(_ context.Context, code string) (*models.Product, error) {
	if m.getByCodeErr != nil {
		return nil, m.getByCodeErr
	}
	return m.getByCodeProd, nil
}

func (m *mockProductsRepository) GetProductsWithFilters(_ context.Context, offset, limit int, categoryCode *string, priceLessThan *float64) ([]models.Product, int64, error) {
	if m.getAllErr != nil {
		return nil, 0, m.getAllErr
	}

	filtered := m.products

	if categoryCode != nil && *categoryCode != "" {
		var categoryFiltered []models.Product
		for _, p := range filtered {
			if p.Category != nil && p.Category.Code == *categoryCode {
				categoryFiltered = append(categoryFiltered, p)
			}
		}
		filtered = categoryFiltered
	}

	if priceLessThan != nil {
		var priceFiltered []models.Product
		for _, p := range filtered {
			if p.Price.InexactFloat64() < *priceLessThan {
				priceFiltered = append(priceFiltered, p)
			}
		}
		filtered = priceFiltered
	}

	total := int64(len(filtered))

	start := offset
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	if start < len(filtered) {
		filtered = filtered[start:end]
	} else {
		filtered = []models.Product{}
	}

	return filtered, total, nil
}

func TestCatalogHandler_HandleGet(t *testing.T) {
	t.Run("successful request with category", func(t *testing.T) {
		clothingID := uint(1)
		category := &models.Category{
			ID:   clothingID,
			Code: "CLOTHING",
			Name: "Clothing",
		}

		products := []models.Product{
			{
				ID:         1,
				Code:       "PROD001",
				Price:      decimal.NewFromFloat(10.99),
				CategoryID: &clothingID,
				Category:   category,
			},
			{
				ID:         2,
				Code:       "PROD002",
				Price:      decimal.NewFromFloat(12.49),
				CategoryID: nil,
				Category:   nil,
			},
		}

		mockRepo := &mockProductsRepository{
			products: products,
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response CatalogResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Products, 2)
		assert.Equal(t, "PROD001", response.Products[0].Code)
		assert.Equal(t, 10.99, response.Products[0].Price)
		assert.NotNil(t, response.Products[0].Category)
		assert.Equal(t, "Clothing", *response.Products[0].Category)
		assert.Nil(t, response.Products[1].Category)
	})
}

func TestCatalogHandler_HandleGetByCode(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		clothingID := uint(1)
		category := &models.Category{
			ID:   clothingID,
			Code: "CLOTHING",
			Name: "Clothing",
		}

		product := &models.Product{
			ID:         1,
			Code:       "PROD001",
			Price:      decimal.NewFromFloat(10.99),
			CategoryID: &clothingID,
			Category:   category,
			Variants: []models.Variant{
				{
					ID:        1,
					ProductID: 1,
					Name:      "Variant A",
					SKU:       "SKU001A",
					Price:     decimal.NewFromFloat(11.99),
				},
				{
					ID:        2,
					ProductID: 1,
					Name:      "Variant B",
					SKU:       "SKU001B",
					Price:     decimal.Zero,
				},
			},
		}

		mockRepo := &mockProductsRepository{
			getByCodeProd: product,
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog/PROD001", nil)
		req.SetPathValue("code", "PROD001")
		w := httptest.NewRecorder()

		handler.HandleGetByCode(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response ProductDetailResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "PROD001", response.Code)
		assert.Equal(t, 10.99, response.Price)
		assert.NotNil(t, response.Category)
		assert.Equal(t, "Clothing", *response.Category)
		assert.Len(t, response.Variants, 2)
		assert.Equal(t, "Variant A", response.Variants[0].Name)
		assert.Equal(t, 11.99, response.Variants[0].Price)
		assert.Equal(t, "Variant B", response.Variants[1].Name)
		assert.Equal(t, 10.99, response.Variants[1].Price)
	})

	t.Run("missing code parameter", func(t *testing.T) {
		mockRepo := &mockProductsRepository{}
		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog/", nil)
		req.SetPathValue("code", "")
		w := httptest.NewRecorder()

		handler.HandleGetByCode(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("product not found", func(t *testing.T) {
		mockRepo := &mockProductsRepository{
			getByCodeErr: gorm.ErrRecordNotFound,
		}
		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog/PROD999", nil)
		req.SetPathValue("code", "PROD999")
		w := httptest.NewRecorder()

		handler.HandleGetByCode(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestCatalogHandler_HandleGet_GetProductsWithFilters(t *testing.T) {
	clothingID := uint(1)
	shoesID := uint(2)
	accessoriesID := uint(3)

	clothingCategory := &models.Category{
		ID:   clothingID,
		Code: "CLOTHING",
		Name: "Clothing",
	}
	shoesCategory := &models.Category{
		ID:   shoesID,
		Code: "SHOES",
		Name: "Shoes",
	}
	accessoriesCategory := &models.Category{
		ID:   accessoriesID,
		Code: "ACCESSORIES",
		Name: "Accessories",
	}

	allProducts := []models.Product{
		{
			ID:         1,
			Code:       "PROD001",
			Price:      decimal.NewFromFloat(10.99),
			CategoryID: &clothingID,
			Category:   clothingCategory,
		},
		{
			ID:         2,
			Code:       "PROD002",
			Price:      decimal.NewFromFloat(12.49),
			CategoryID: &shoesID,
			Category:   shoesCategory,
		},
		{
			ID:         3,
			Code:       "PROD003",
			Price:      decimal.NewFromFloat(8.99),
			CategoryID: &accessoriesID,
			Category:   accessoriesCategory,
		},
		{
			ID:         4,
			Code:       "PROD004",
			Price:      decimal.NewFromFloat(15.99),
			CategoryID: &clothingID,
			Category:   clothingCategory,
		},
		{
			ID:         5,
			Code:       "PROD005",
			Price:      decimal.NewFromFloat(20.00),
			CategoryID: nil,
			Category:   nil,
		},
	}

	t.Run("no filters - returns all products with pagination", func(t *testing.T) {
		mockRepo := &mockProductsRepository{
			products: allProducts,
			total:    int64(len(allProducts)),
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog?offset=0&limit=3", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response CatalogResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), response.Total)
		assert.Len(t, response.Products, 3)
		assert.Equal(t, "PROD001", response.Products[0].Code)
		assert.Equal(t, "PROD002", response.Products[1].Code)
		assert.Equal(t, "PROD003", response.Products[2].Code)
	})

	t.Run("filter by category only", func(t *testing.T) {
		mockRepo := &mockProductsRepository{
			products: allProducts,
			total:    2,
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog?category=CLOTHING", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response CatalogResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), response.Total)
		assert.Len(t, response.Products, 2)
		assert.Equal(t, "PROD001", response.Products[0].Code)
		assert.Equal(t, "Clothing", *response.Products[0].Category)
		assert.Equal(t, "PROD004", response.Products[1].Code)
		assert.Equal(t, "Clothing", *response.Products[1].Category)
	})

	t.Run("filter by price_less_than only", func(t *testing.T) {
		mockRepo := &mockProductsRepository{
			products: allProducts,
			total:    3,
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog?price_less_than=12.00", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response CatalogResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), response.Total)
		assert.Len(t, response.Products, 2)
		for _, product := range response.Products {
			assert.Less(t, product.Price, 12.00)
		}
	})

	t.Run("filter by both category and price", func(t *testing.T) {
		mockRepo := &mockProductsRepository{
			products: allProducts,
			total:    1,
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog?category=CLOTHING&price_less_than=12.00", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response CatalogResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), response.Total)
		assert.Len(t, response.Products, 1)
		assert.Equal(t, "PROD001", response.Products[0].Code)
		assert.Equal(t, "Clothing", *response.Products[0].Category)
		assert.Less(t, response.Products[0].Price, 12.00)
	})

	t.Run("pagination with offset", func(t *testing.T) {
		mockRepo := &mockProductsRepository{
			products: allProducts,
			total:    int64(len(allProducts)),
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog?offset=2&limit=2", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response CatalogResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), response.Total)
		assert.Len(t, response.Products, 2)
		assert.Equal(t, "PROD003", response.Products[0].Code)
		assert.Equal(t, "PROD004", response.Products[1].Code)
	})

	t.Run("pagination with limit exceeding available products", func(t *testing.T) {
		mockRepo := &mockProductsRepository{
			products: allProducts,
			total:    int64(len(allProducts)),
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog?offset=3&limit=10", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response CatalogResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), response.Total)
		assert.Len(t, response.Products, 2)
	})

	t.Run("pagination with offset beyond available products", func(t *testing.T) {
		mockRepo := &mockProductsRepository{
			products: allProducts,
			total:    int64(len(allProducts)),
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog?offset=10&limit=5", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response CatalogResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), response.Total)
		assert.Len(t, response.Products, 0)
	})

	t.Run("default pagination values", func(t *testing.T) {
		mockRepo := &mockProductsRepository{
			products: allProducts,
			total:    int64(len(allProducts)),
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response CatalogResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), response.Total)
		assert.Len(t, response.Products, 5)
	})

	t.Run("limit validation - minimum (clamped)", func(t *testing.T) {
		mockRepo := &mockProductsRepository{
			products: allProducts,
			total:    int64(len(allProducts)),
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog?limit=0", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response CatalogResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Products, 1)
	})

	t.Run("limit validation - maximum (clamped)", func(t *testing.T) {
		largeProductList := make([]models.Product, 150)
		for i := range largeProductList {
			largeProductList[i] = models.Product{
				ID:    uint(i + 1),
				Code:  "PROD" + strconv.Itoa(i+1),
				Price: decimal.NewFromFloat(10.0 + float64(i)),
			}
		}

		mockRepo := &mockProductsRepository{
			products: largeProductList,
			total:    int64(len(largeProductList)),
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog?limit=200", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response CatalogResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Products, 100)
	})

	t.Run("invalid offset parameter - returns bad request", func(t *testing.T) {
		mockRepo := &mockProductsRepository{
			products: allProducts,
			total:    int64(len(allProducts)),
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog?offset=invalid", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Contains(t, errorResponse["error"], "offset must be a valid integer")
	})

	t.Run("negative offset parameter - returns bad request", func(t *testing.T) {
		mockRepo := &mockProductsRepository{
			products: allProducts,
			total:    int64(len(allProducts)),
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog?offset=-1", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Contains(t, errorResponse["error"], "offset must be non-negative")
	})

	t.Run("invalid limit parameter - returns bad request", func(t *testing.T) {
		mockRepo := &mockProductsRepository{
			products: allProducts,
			total:    int64(len(allProducts)),
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog?limit=invalid", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Contains(t, errorResponse["error"], "limit must be a valid integer")
	})

	t.Run("invalid price_less_than parameter - returns bad request", func(t *testing.T) {
		mockRepo := &mockProductsRepository{
			products: allProducts,
			total:    int64(len(allProducts)),
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog?price_less_than=invalid", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Contains(t, errorResponse["error"], "price_less_than must be a valid number")
	})

	t.Run("negative price_less_than parameter - returns bad request", func(t *testing.T) {
		mockRepo := &mockProductsRepository{
			products: allProducts,
			total:    int64(len(allProducts)),
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog?price_less_than=-10", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Contains(t, errorResponse["error"], "price_less_than must be greater than 0")
	})

	t.Run("zero price_less_than parameter - returns bad request", func(t *testing.T) {
		mockRepo := &mockProductsRepository{
			products: allProducts,
			total:    int64(len(allProducts)),
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog?price_less_than=0", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Contains(t, errorResponse["error"], "price_less_than must be greater than 0")
	})

	t.Run("repository error handling", func(t *testing.T) {
		mockRepo := &mockProductsRepository{
			products:  allProducts,
			getAllErr: errors.New("database error"),
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "database error")
	})

	t.Run("empty category filter", func(t *testing.T) {
		mockRepo := &mockProductsRepository{
			products: allProducts,
			total:    int64(len(allProducts)),
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog?category=", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response CatalogResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), response.Total)
	})
}
