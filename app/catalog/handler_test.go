package catalog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type mockProductsRepository struct {
	products      []models.Product
	getAllErr     error
	getByCodeErr  error
	getByCodeProd *models.Product
}

func (m *mockProductsRepository) GetAllProducts() ([]models.Product, error) {
	return m.products, m.getAllErr
}

func (m *mockProductsRepository) GetProductByCode(code string) (*models.Product, error) {
	if m.getByCodeErr != nil {
		return nil, m.getByCodeErr
	}
	return m.getByCodeProd, nil
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
