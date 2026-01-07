package catalog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type mockProductsRepository struct {
	products  []models.Product
	getAllErr error
}

func (m *mockProductsRepository) GetAllProducts() ([]models.Product, error) {
	return m.products, m.getAllErr
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

		var response Response
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
