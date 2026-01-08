package categories

import (
	"encoding/json"
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type CategoryResponse struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type CreateCategoryRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CategoriesHandler struct {
	categories models.CategoriesRepositoryInterface
}

func NewCategoriesHandler(r models.CategoriesRepositoryInterface) *CategoriesHandler {
	return &CategoriesHandler{
		categories: r,
	}
}

func (h *CategoriesHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	categories, err := h.categories.GetAllCategories(ctx)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses := make([]CategoryResponse, len(categories))
	for i, cat := range categories {
		responses[i] = CategoryResponse{
			ID:   cat.ID,
			Code: cat.Code,
			Name: cat.Name,
		}
	}

	api.OKResponse(w, responses)
}

func (h *CategoriesHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Code == "" || req.Name == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "code and name are required")
		return
	}

	category := &models.Category{
		Code: req.Code,
		Name: req.Name,
	}

	if err := h.categories.CreateCategory(ctx, category); err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := CategoryResponse{
		ID:   category.ID,
		Code: category.Code,
		Name: category.Name,
	}

	api.OKResponse(w, response)
}
