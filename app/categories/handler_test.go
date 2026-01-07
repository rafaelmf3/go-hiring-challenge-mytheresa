package categories

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
)

type mockCategoriesRepository struct {
	categories []models.Category
	getAllErr  error
}

func (m *mockCategoriesRepository) GetAllCategories() ([]models.Category, error) {
	return m.categories, m.getAllErr
}

func TestCategoriesHandler_HandleGet(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		categories := []models.Category{
			{
				ID:   1,
				Code: "CLOTHING",
				Name: "Clothing",
			},
			{
				ID:   2,
				Code: "SHOES",
				Name: "Shoes",
			},
			{
				ID:   3,
				Code: "ACCESSORIES",
				Name: "Accessories",
			},
		}

		mockRepo := &mockCategoriesRepository{
			categories: categories,
		}

		handler := NewCategoriesHandler(mockRepo)
		req := httptest.NewRequest("GET", "/categories", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response []CategoryResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response, 3)
		assert.Equal(t, "CLOTHING", response[0].Code)
		assert.Equal(t, "Clothing", response[0].Name)
		assert.Equal(t, uint(1), response[0].ID)
	})

	t.Run("empty categories list", func(t *testing.T) {
		mockRepo := &mockCategoriesRepository{
			categories: []models.Category{},
		}

		handler := NewCategoriesHandler(mockRepo)
		req := httptest.NewRequest("GET", "/categories", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []CategoryResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response, 0)
	})

	t.Run("database error", func(t *testing.T) {
		mockRepo := &mockCategoriesRepository{
			getAllErr: assert.AnError,
		}

		handler := NewCategoriesHandler(mockRepo)
		req := httptest.NewRequest("GET", "/categories", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
