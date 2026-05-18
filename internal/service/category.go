package service

import (
	"context"
	"fmt"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type CategoryService interface {
	CreateCategory(ctx context.Context, req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error)
	GetCategory(ctx context.Context, id uint) (*dto.CategoryResponse, error)
	GetCategories(ctx context.Context) ([]*dto.CategoryResponse, error)
	UpdateCategory(ctx context.Context, id uint, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error)
	DeleteCategory(ctx context.Context, id uint) error
}

type categoryService struct {
	categoryRepository repository.CategoryRepository
	tracer             trace.Tracer
}

func (c *categoryService) CreateCategory(ctx context.Context, req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {
	ctx, span := c.tracer.Start(ctx, "CategoryService.CreateCategory",
		trace.WithAttributes(attribute.String("category.name", req.Name)))
	defer span.End()

	category := &domain.Category{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := c.categoryRepository.CreateCategory(ctx, category); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create category")
		return nil, err
	}
	span.SetAttributes(attribute.Int64("category.id", int64(category.Id)))

	return c.toCategoryReps(category), nil
}

func (c *categoryService) GetCategory(ctx context.Context, id uint) (*dto.CategoryResponse, error) {
	ctx, span := c.tracer.Start(ctx, "CategoryService.GetCategory",
		trace.WithAttributes(attribute.Int64("category.id", int64(id))))
	defer span.End()

	category, err := c.categoryRepository.GetCategoryById(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "category not found")
		return nil, fmt.Errorf("failed to get category by id: %w", err)
	}

	return c.toCategoryReps(category), nil
}

func (c *categoryService) GetCategories(ctx context.Context) ([]*dto.CategoryResponse, error) {
	ctx, span := c.tracer.Start(ctx, "CategoryService.GetCategories")
	defer span.End()

	categories, err := c.categoryRepository.GetCategories(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get categories")
		return nil, err
	}

	resp := make([]*dto.CategoryResponse, len(categories))
	for i := range categories {
		resp[i] = c.toCategoryReps(categories[i])
	}

	return resp, nil
}

func (c *categoryService) UpdateCategory(ctx context.Context, id uint, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error) {
	ctx, span := c.tracer.Start(ctx, "CategoryService.UpdateCategory",
		trace.WithAttributes(attribute.Int64("category.id", int64(id))))
	defer span.End()

	category, err := c.categoryRepository.GetCategoryById(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "category not found")
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
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update category")
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return c.toCategoryReps(category), nil
}

func (c *categoryService) DeleteCategory(ctx context.Context, id uint) error {
	ctx, span := c.tracer.Start(ctx, "CategoryService.DeleteCategory",
		trace.WithAttributes(attribute.Int64("category.id", int64(id))))
	defer span.End()

	if err := c.categoryRepository.DeleteCategory(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete category")
		return err
	}
	return nil
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

func NewCategoryService(categoryRepository repository.CategoryRepository, tracer trace.Tracer) CategoryService {
	return &categoryService{
		categoryRepository: categoryRepository,
		tracer:             tracer,
	}
}
