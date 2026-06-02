package handler

import (
	"context"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/grpc/protos"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProductGrpcHandler struct {
	protos.UnimplementedProductServiceServer
	productService service.ProductService
}

func (p *ProductGrpcHandler) CreateProduct(ctx context.Context, req *protos.CreateProductRequest) (*protos.ProductResponse, error) {
	createDTO := &dto.CreateProductRequest{
		CategoryId:  uint(req.CategoryId),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       int(req.Stock),
		SKU:         req.Sku,
	}

	resp, err := p.productService.CreateProduct(ctx, createDTO)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create product: %v", err)
	}

	return p.mapToProductResponse(resp), nil
}

func (p *ProductGrpcHandler) ListProducts(ctx context.Context, req *protos.ListProductsRequest) (*protos.ListProductsResponse, error) {
	products, meta, err := p.productService.GetProducts(ctx, int(req.Page), int(req.Limit))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get products: %v", err)
	}

	protoProducts := make([]*protos.ProductResponse, len(products))
	for i, prod := range products {
		protoProducts[i] = p.mapToProductResponse(prod)
	}

	return &protos.ListProductsResponse{
		Products: protoProducts,
		Meta:     p.mapToPaginatedMeta(meta),
	}, nil
}

func (p *ProductGrpcHandler) GetProduct(ctx context.Context, req *protos.GetProductRequest) (*protos.ProductResponse, error) {
	resp, err := p.productService.GetProductById(ctx, uint(req.GetId()))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
	}

	return p.mapToProductResponse(resp), nil
}

func (p *ProductGrpcHandler) UpdateProduct(ctx context.Context, req *protos.UpdateProductRequest) (*protos.ProductResponse, error) {
	updateDTO := &dto.UpdateProductRequest{}

	if req.CategoryId != nil {
		catID := uint(*req.CategoryId)
		updateDTO.CategoryId = &catID
	}
	if req.Name != nil {
		updateDTO.Name = req.Name
	}
	if req.Description != nil {
		updateDTO.Description = req.Description
	}
	if req.Price != nil {
		updateDTO.Price = req.Price
	}
	if req.Stock != nil {
		stockVal := int(*req.Stock)
		updateDTO.Stock = &stockVal
	}
	if req.IsActive != nil {
		updateDTO.IsActive = req.IsActive
	}

	resp, err := p.productService.UpdateProduct(ctx, uint(req.GetId()), updateDTO)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update product: %v", err)
	}

	return p.mapToProductResponse(resp), nil
}

func (p *ProductGrpcHandler) DeleteProduct(ctx context.Context, req *protos.DeleteProductRequest) (*emptypb.Empty, error) {
	err := p.productService.DeleteProduct(ctx, uint(req.GetId()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete product: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (p *ProductGrpcHandler) AddProductImage(ctx context.Context, req *protos.AddProductImageRequest) (*protos.ProductImageResponse, error) {
	resp, err := p.productService.AddProductImage(ctx, uint(req.ProductId), req.Url, req.AltText)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add image: %v", err)
	}

	return &protos.ProductImageResponse{
		Id:        uint64(resp.Id),
		Url:       resp.URL,
		AltText:   resp.AltText,
		IsPrimary: resp.IsPrimary,
		CreatedAt: timestamppb.New(resp.CreatedAt),
	}, nil
}

func (p *ProductGrpcHandler) SearchProducts(ctx context.Context, req *protos.SearchProductsRequest) (*protos.SearchProductsResponse, error) {
	searchDTO := &dto.SearchProductsRequest{
		Query: req.Query,
		Page:  int(req.Page),
		Limit: int(req.Limit),
	}

	if req.CategoryId != nil {
		catID := uint(*req.CategoryId)
		searchDTO.CategoryId = &catID
	}
	if req.MinPrice != nil {
		searchDTO.MinPrice = req.MinPrice
	}
	if req.MaxPrice != nil {
		searchDTO.MaxPrice = req.MaxPrice
	}

	results, meta, err := p.productService.SearchProducts(ctx, searchDTO)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "product search failed: %v", err)
	}

	protoResults := make([]*protos.ProductSearchResult, len(results))
	for i, res := range results {
		protoResults[i] = &protos.ProductSearchResult{
			Product: p.mapToProductResponse(&res.ProductResponse),
			Rank:    res.Rank,
		}
	}

	return &protos.SearchProductsResponse{
		Results: protoResults,
		Meta:    p.mapToPaginatedMeta(meta),
	}, nil
}

func (p *ProductGrpcHandler) mapToProductResponse(prod *dto.ProductResponse) *protos.ProductResponse {
	if prod == nil {
		return nil
	}

	protoImages := make([]*protos.ProductImageResponse, len(prod.Images))
	for i, img := range prod.Images {
		protoImages[i] = &protos.ProductImageResponse{
			Id:        uint64(img.Id),
			Url:       img.URL,
			AltText:   img.AltText,
			IsPrimary: img.IsPrimary,
			CreatedAt: timestamppb.New(img.CreatedAt),
		}
	}

	return &protos.ProductResponse{
		Id:          uint64(prod.Id),
		CategoryId:  uint64(prod.CategoryId),
		Name:        prod.Name,
		Description: prod.Description,
		Price:       prod.Price,
		Stock:       int32(prod.Stock),
		Sku:         prod.SKU,
		IsActive:    prod.IsActive,
		Category: &protos.CategoryResponse{
			Id:          uint64(prod.Category.Id),
			Name:        prod.Category.Name,
			Description: prod.Category.Description,
			IsActive:    prod.Category.IsActive,
			CreatedAt:   timestamppb.New(prod.Category.CreatedAt),
			UpdatedAt:   timestamppb.New(prod.Category.UpdatedAt),
		},
		Images:    protoImages,
		CreatedAt: timestamppb.New(prod.CreatedAt),
		UpdatedAt: timestamppb.New(prod.UpdatedAt),
	}
}

func (p *ProductGrpcHandler) mapToPaginatedMeta(m *helper.PaginatedMeta) *protos.PaginatedMeta {
	if m == nil {
		return nil
	}
	return &protos.PaginatedMeta{
		Page:       int32(m.Page),
		Limit:      int32(m.Limit),
		Total:      m.Total,
		TotalPages: int32(m.TotalPage),
	}
}

func NewProductGrpcHandler(productService service.ProductService) *ProductGrpcHandler {
	return &ProductGrpcHandler{
		productService: productService,
	}
}
