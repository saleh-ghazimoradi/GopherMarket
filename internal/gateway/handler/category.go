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

// CreateCategory docs
// @Summary Create a new category
// @Description Create a new product category (Admin only)
// @Tags Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateCategoryRequest true "Category data"
// @Success 201 {object} helper.Response{data=dto.CategoryResponse} "Category created successfully"
// @Failure 400 {object} helper.Response "Invalid request data"
// @Failure 401 {object} helper.Response "Unauthorized"
// @Failure 403 {object} helper.Response "Admin access required"
// @Router /categories [post]
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

// GetCategories docs
// @Summary Get all categories
// @Description Retrieve all active categories
// @Tags Categories
// @Produce json
// @Success 200 {object} helper.Response{data=[]dto.CategoryResponse} "Categories retrieved successfully"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /categories [get]
func (c *CategoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := c.categoryService.GetCategories(r.Context())
	if err != nil {
		helper.InternalServerError(w, "failed to get the categories", err)
		return
	}

	helper.SuccessResponse(w, "categories successfully retrieved", categories)
}

// UpdateCategory docs
// @Summary Update a category
// @Description Update an existing category (Admin only)
// @Tags Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Param request body dto.UpdateCategoryRequest true "Category update data"
// @Success 200 {object} helper.Response{data=dto.CategoryResponse} "Category updated successfully"
// @Failure 400 {object} helper.Response "Invalid request data"
// @Failure 401 {object} helper.Response "Unauthorized"
// @Failure 403 {object} helper.Response "Admin access required"
// @Router /categories/{id} [put]
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

// DeleteCategory docs
// @Summary Delete a category
// @Description Delete a category (Admin only)
// @Tags Categories
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Success 200 {object} helper.Response "Category deleted successfully"
// @Failure 400 {object} helper.Response "Invalid category ID"
// @Failure 401 {object} helper.Response "Unauthorized"
// @Failure 403 {object} helper.Response "Admin access required"
// @Router /categories/{id} [delete]
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
