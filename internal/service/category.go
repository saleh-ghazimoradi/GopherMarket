package service

import (
	"context"
	"fmt"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/repository"
)

type CategoryService interface {
	CreateCategory(ctx context.Context, req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error)
	GetCategories(ctx context.Context) ([]*dto.CategoryResponse, error)
	UpdateCategory(ctx context.Context, id uint, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error)
	DeleteCategory(ctx context.Context, id uint) error
}

type categoryService struct {
	categoryRepository repository.CategoryRepository
}

func (c *categoryService) CreateCategory(ctx context.Context, req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {
	category := &domain.Category{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := c.categoryRepository.CreateCategory(ctx, category); err != nil {
		return nil, err
	}

	return c.toCategoryReps(category), nil
}

func (c *categoryService) GetCategories(ctx context.Context) ([]*dto.CategoryResponse, error) {
	categories, err := c.categoryRepository.GetCategories(ctx)
	if err != nil {
		return nil, err
	}

	resp := make([]*dto.CategoryResponse, len(categories))
	for i := range categories {
		resp[i] = c.toCategoryReps(categories[i])
	}

	return resp, nil
}

func (c *categoryService) UpdateCategory(ctx context.Context, id uint, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error) {
	category, err := c.categoryRepository.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get category by id: %w", err)
	}

	if req.Name != nil {
		category.Name = *req.Name
	}

	if req.Description != nil {
		category.Description = *req.Description
	}

	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := c.categoryRepository.UpdateCategory(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return c.toCategoryReps(category), nil
}

func (c *categoryService) DeleteCategory(ctx context.Context, id uint) error {
	return c.categoryRepository.DeleteCategory(ctx, id)
}

func (c *categoryService) toCategoryReps(category *domain.Category) *dto.CategoryResponse {
	return &dto.CategoryResponse{
		Id:          category.Id,
		Name:        category.Name,
		Description: category.Description,
		IsActive:    category.IsActive,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}
}

func NewCategoryService(categoryRepository repository.CategoryRepository) CategoryService {
	return &categoryService{
		categoryRepository: categoryRepository,
	}
}
