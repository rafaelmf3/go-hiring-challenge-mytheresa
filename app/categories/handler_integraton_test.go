//go:build integration

package categories

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/app/catalog"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/mytheresa/go-hiring-challenge/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategoriesHandler_Integration(t *testing.T) {
	tdb := test.SetupTestDB(t)
	defer tdb.Close()

	prodRepo := models.NewProductsRepository(tdb.DB)
	catalogHandler := catalog.NewCatalogHandler(prodRepo)
	catRepo := models.NewCategoriesRepository(tdb.DB)
	categoriesHandler := NewCategoriesHandler(catRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /catalog", catalogHandler.HandleGet)
	mux.HandleFunc("GET /catalog/{code}", catalogHandler.HandleGetByCode)
	mux.HandleFunc("GET /categories", categoriesHandler.HandleGet)
	mux.HandleFunc("POST /categories", categoriesHandler.HandlePost)

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Run("GET /categories - returns all categories", func(t *testing.T) {
		tdb.SeedTestData(t)

		resp, err := http.Get(server.URL + "/categories")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response []CategoryResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Greater(t, len(response), 0)
	})

	t.Run("POST /categories - creates new category", func(t *testing.T) {
		reqBody := CreateCategoryRequest{
			Code: "ELECTRONICS",
			Name: "Electronics",
		}
		body, _ := json.Marshal(reqBody)

		resp, err := http.Post(
			server.URL+"/categories",
			"application/json",
			bytes.NewBuffer(body),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response CategoryResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, "ELECTRONICS", response.Code)
		assert.Equal(t, "Electronics", response.Name)
		assert.Greater(t, response.ID, uint(0))
	})

	t.Run("POST /categories - validation error", func(t *testing.T) {
		reqBody := CreateCategoryRequest{
			Code: "",
			Name: "Test",
		}
		body, _ := json.Marshal(reqBody)

		resp, err := http.Post(
			server.URL+"/categories",
			"application/json",
			bytes.NewBuffer(body),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
