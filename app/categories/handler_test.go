package categories

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
)

type mockCategoriesRepository struct {
	categories      []models.Category
	getAllErr       error
	createErr       error
	createdCategory *models.Category
}

func (m *mockCategoriesRepository) GetAllCategories(_ context.Context) ([]models.Category, error) {
	return m.categories, m.getAllErr
}

func (m *mockCategoriesRepository) CreateCategory(_ context.Context, category *models.Category) error {
	if m.createErr != nil {
		return m.createErr
	}
	if m.createdCategory != nil {
		category.ID = m.createdCategory.ID
	}
	return nil
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

func TestCategoriesHandler_HandlePost(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		mockRepo := &mockCategoriesRepository{
			createdCategory: &models.Category{
				ID:   1,
				Code: "ELECTRONICS",
				Name: "Electronics",
			},
		}

		handler := NewCategoriesHandler(mockRepo)

		reqBody := CreateCategoryRequest{
			Code: "ELECTRONICS",
			Name: "Electronics",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePost(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response CategoryResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "ELECTRONICS", response.Code)
		assert.Equal(t, "Electronics", response.Name)
	})

	t.Run("missing code field", func(t *testing.T) {
		mockRepo := &mockCategoriesRepository{}
		handler := NewCategoriesHandler(mockRepo)

		reqBody := CreateCategoryRequest{
			Name: "Electronics",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePost(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing name field", func(t *testing.T) {
		mockRepo := &mockCategoriesRepository{}
		handler := NewCategoriesHandler(mockRepo)

		reqBody := CreateCategoryRequest{
			Code: "ELECTRONICS",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePost(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid JSON body", func(t *testing.T) {
		mockRepo := &mockCategoriesRepository{}
		handler := NewCategoriesHandler(mockRepo)

		req := httptest.NewRequest("POST", "/categories", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePost(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("database error", func(t *testing.T) {
		mockRepo := &mockCategoriesRepository{
			createErr: assert.AnError,
		}
		handler := NewCategoriesHandler(mockRepo)

		reqBody := CreateCategoryRequest{
			Code: "ELECTRONICS",
			Name: "Electronics",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePost(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
