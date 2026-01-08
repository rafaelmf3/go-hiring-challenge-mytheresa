//go:build integration

package catalog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/app/categories"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/mytheresa/go-hiring-challenge/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCatalogHandler_Integration(t *testing.T) {
	tdb := test.SetupTestDB(t)
	defer tdb.Close()

	tdb.SeedTestData(t)

	prodRepo := models.NewProductsRepository(tdb.DB)
	catalogHandler := NewCatalogHandler(prodRepo)
	catRepo := models.NewCategoriesRepository(tdb.DB)
	categoriesHandler := categories.NewCategoriesHandler(catRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /catalog", catalogHandler.HandleGet)
	mux.HandleFunc("GET /catalog/{code}", catalogHandler.HandleGetByCode)
	mux.HandleFunc("GET /categories", categoriesHandler.HandleGet)
	mux.HandleFunc("POST /categories", categoriesHandler.HandlePost)

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Run("GET /catalog - returns all products", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/catalog")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response CatalogResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Greater(t, len(response.Products), 0)
		assert.Greater(t, response.Total, int64(0))
	})

	t.Run("GET /catalog - with pagination", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/catalog?offset=0&limit=2")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response CatalogResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(response.Products), 2)
	})

	t.Run("GET /catalog - filter by category", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/catalog?category=CLOTHING")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response CatalogResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		for _, product := range response.Products {
			assert.NotNil(t, product.Category)
			assert.Equal(t, "Clothing", *product.Category)
		}
	})

	t.Run("GET /catalog - filter by price", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/catalog?price_less_than=12.00")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response CatalogResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		for _, product := range response.Products {
			assert.Less(t, product.Price, 12.00)
		}
	})

	t.Run("GET /catalog/{code} - returns product details", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/catalog/PROD001")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response ProductDetailResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, "PROD001", response.Code)
		assert.NotNil(t, response.Category)
	})

	t.Run("GET /catalog/{code} - product not found", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/catalog/PROD999")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("GET /catalog - invalid offset returns bad request", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/catalog?offset=invalid")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errorResponse map[string]string
		err = json.NewDecoder(resp.Body).Decode(&errorResponse)
		require.NoError(t, err)
		assert.Contains(t, errorResponse["error"], "offset must be a valid integer")
	})

	t.Run("GET /catalog - negative offset returns bad request", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/catalog?offset=-1")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errorResponse map[string]string
		err = json.NewDecoder(resp.Body).Decode(&errorResponse)
		require.NoError(t, err)
		assert.Contains(t, errorResponse["error"], "offset must be non-negative")
	})

	t.Run("GET /catalog - invalid limit returns bad request", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/catalog?limit=invalid")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errorResponse map[string]string
		err = json.NewDecoder(resp.Body).Decode(&errorResponse)
		require.NoError(t, err)
		assert.Contains(t, errorResponse["error"], "limit must be a valid integer")
	})

	t.Run("GET /catalog - invalid price_less_than returns bad request", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/catalog?price_less_than=invalid")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errorResponse map[string]string
		err = json.NewDecoder(resp.Body).Decode(&errorResponse)
		require.NoError(t, err)
		assert.Contains(t, errorResponse["error"], "price_less_than must be a valid number")
	})

	t.Run("GET /catalog - negative price_less_than returns bad request", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/catalog?price_less_than=-10")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errorResponse map[string]string
		err = json.NewDecoder(resp.Body).Decode(&errorResponse)
		require.NoError(t, err)
		assert.Contains(t, errorResponse["error"], "price_less_than must be greater than 0")
	})

	t.Run("GET /catalog - limit clamping works (0 clamped to 1)", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/catalog?limit=0")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response CatalogResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(response.Products), 1)
	})

	t.Run("GET /catalog - limit clamping works (200 clamped to 100)", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/catalog?limit=200")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response CatalogResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(response.Products), 100)
	})
}
