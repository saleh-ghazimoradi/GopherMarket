package handler

import (
	"context"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/grpc/protos"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CategoryGrpcHandler struct {
	protos.UnimplementedCategoryServiceServer
	categoryService service.CategoryService
}

func (c *CategoryGrpcHandler) CreateCategory(ctx context.Context, req *protos.CreateCategoryRequest) (*protos.CategoryResponse, error) {
	categoryDTO := &dto.CreateCategoryRequest{
		Name:        req.Name,
		Description: req.Description,
	}

	category, err := c.categoryService.CreateCategory(ctx, categoryDTO)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create category: %v", err)
	}

	return c.mapToCategoryResponse(category), nil
}

func (c *CategoryGrpcHandler) GetCategory(ctx context.Context, req *protos.GetCategoryRequest) (*protos.CategoryResponse, error) {
	category, err := c.categoryService.GetCategory(ctx, uint(req.GetId()))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "category not found: %v", err)
	}

	return c.mapToCategoryResponse(category), nil
}

func (c *CategoryGrpcHandler) ListCategories(ctx context.Context, req *protos.ListCategoriesRequest) (*protos.ListCategoriesResponse, error) {
	categories, err := c.categoryService.GetCategories(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list categories: %v", err)
	}

	protoCategories := make([]*protos.CategoryResponse, len(categories))
	for i, cat := range categories {
		protoCategories[i] = c.mapToCategoryResponse(cat)
	}

	return &protos.ListCategoriesResponse{
		Categories: protoCategories,
	}, nil
}

func (c *CategoryGrpcHandler) UpdateCategory(ctx context.Context, req *protos.UpdateCategoryRequest) (*protos.CategoryResponse, error) {
	updateDTO := &dto.UpdateCategoryRequest{}

	if req.Name != nil {
		updateDTO.Name = req.Name
	}

	if req.Description != nil {
		updateDTO.Description = req.Description
	}

	if req.IsActive != nil {
		updateDTO.IsActive = req.IsActive
	}

	category, err := c.categoryService.UpdateCategory(ctx, uint(req.GetId()), updateDTO)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update category: %v", err)
	}

	return c.mapToCategoryResponse(category), nil
}

func (c *CategoryGrpcHandler) DeleteCategory(ctx context.Context, req *protos.DeleteCategoryRequest) (*emptypb.Empty, error) {
	err := c.categoryService.DeleteCategory(ctx, uint(req.GetId()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete category: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (c *CategoryGrpcHandler) mapToCategoryResponse(cat *dto.CategoryResponse) *protos.CategoryResponse {
	if cat == nil {
		return nil
	}

	return &protos.CategoryResponse{
		Id:          uint64(cat.Id),
		Name:        cat.Name,
		Description: cat.Description,
		IsActive:    cat.IsActive,
		CreatedAt:   timestamppb.New(cat.CreatedAt),
		UpdatedAt:   timestamppb.New(cat.UpdatedAt),
	}
}

func NewCategoryGrpcHandler(categoryService service.CategoryService) *CategoryGrpcHandler {
	return &CategoryGrpcHandler{
		categoryService: categoryService,
	}
}
