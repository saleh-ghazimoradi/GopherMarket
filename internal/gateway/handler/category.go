package handler

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/service"
	"net/http"
)

type CategoryHandler struct {
	categoryService service.CategoryService
}

func (c *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var payload dto.CreateCategoryRequest
	if err := helper.ReadJSON(w, r, &payload); err != nil {
		helper.BadRequestResponse(w, "invalid given payload", err)
		return
	}

	v := helper.NewValidator()
	dto.ValidateCreateCategoryRequest(v, &payload)
	if !v.Valid() {
		helper.FailedValidationResponse(w, "payload is not valid")
		return
	}

	category, err := c.categoryService.CreateCategory(r.Context(), &payload)
	if err != nil {
		helper.InternalServerError(w, "failed to create the category", err)
		return
	}

	helper.CreatedResponse(w, "category successfully created", category)
}

func (c *CategoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := c.categoryService.GetCategories(r.Context())
	if err != nil {
		helper.InternalServerError(w, "failed to get the categories", err)
		return
	}

	helper.SuccessResponse(w, "categories successfully retrieved", categories)
}

func (c *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := helper.ReadParams(r)
	if err != nil {
		helper.BadRequestResponse(w, "invalid id", err)
		return
	}

	var payload dto.UpdateCategoryRequest
	if err := helper.ReadJSON(w, r, &payload); err != nil {
		helper.BadRequestResponse(w, "invalid payload", err)
		return
	}

	v := helper.NewValidator()
	dto.ValidateUpdateCategoryRequest(v, &payload)
	if !v.Valid() {
		helper.FailedValidationResponse(w, "payload is not valid")
		return
	}

	category, err := c.categoryService.UpdateCategory(r.Context(), id, &payload)
	if err != nil {
		helper.InternalServerError(w, "failed to update the category", err)
		return
	}

	helper.SuccessResponse(w, "category successfully updated", category)
}

func (c *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := helper.ReadParams(r)
	if err != nil {
		helper.BadRequestResponse(w, "invalid id", err)
		return
	}

	if err := c.categoryService.DeleteCategory(r.Context(), id); err != nil {
		helper.InternalServerError(w, "failed to delete the category", err)
		return
	}

	helper.SuccessResponse(w, "category successfully deleted", nil)
}

func NewCategoryHandler(categoryService service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}
