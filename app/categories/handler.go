package categories

import (
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type CategoryResponse struct {
	ID   uint   `json:"id"`
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
	categories, err := h.categories.GetAllCategories()
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
