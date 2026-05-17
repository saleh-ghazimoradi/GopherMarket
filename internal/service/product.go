package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/repository"
	"time"
)

const (
	productByIdKey     = "product:id:%d"
	productListKey     = "products:list:page=%d:limit=%d"
	productListPattern = "products:list:*"
	defaultTTL         = 10 * time.Minute
)

type ProductService interface {
	CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error)
	GetProducts(ctx context.Context, page, limit int) ([]*dto.ProductResponse, *helper.PaginatedMeta, error)
	GetProductById(ctx context.Context, id uint) (*dto.ProductResponse, error)
	UpdateProduct(ctx context.Context, id uint, req *dto.UpdateProductRequest) (*dto.ProductResponse, error)
	DeleteProduct(ctx context.Context, id uint) error
	AddProductImage(ctx context.Context, productId uint, url, altText string) (*dto.ProductImageResponse, error)
	SearchProducts(ctx context.Context, req *dto.SearchProductsRequest) ([]*dto.ProductSearchResult, *helper.PaginatedMeta, error)
}

type productService struct {
	productRepository repository.ProductRepository
	redisCache        repository.RedisCache
}

func (p *productService) CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error) {
	product := &domain.Product{
		CategoryId:  req.CategoryId,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		SKU:         req.SKU,
	}

	if err := p.productRepository.CreateProduct(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	_ = p.redisCache.DeletePattern(ctx, productListPattern)

	return p.GetProductById(ctx, product.Id)
}

func (p *productService) GetProducts(ctx context.Context, page, limit int) ([]*dto.ProductResponse, *helper.PaginatedMeta, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	key := fmt.Sprintf(productListKey, page, limit)

	type cachedPage struct {
		Products []*domain.Product
		Total    int64
	}
	var cp cachedPage
	if err := p.redisCache.Get(ctx, key, &cp); err == nil {
		return p.toProductRespList(cp.Products), p.buildMeta(page, limit, cp.Total), nil
	}

	offset := (page - 1) * limit

	total, err := p.productRepository.CountActiveProducts(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count products: %w", err)
	}

	products, err := p.productRepository.GetProducts(ctx, offset, limit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get products: %w", err)
	}

	_ = p.redisCache.Set(ctx, key, cachedPage{Products: products, Total: total}, 5*time.Minute)

	return p.toProductRespList(products), p.buildMeta(page, limit, total), nil
}

func (p *productService) GetProductById(ctx context.Context, id uint) (*dto.ProductResponse, error) {
	key := fmt.Sprintf(productByIdKey, id)

	var productCache domain.Product
	if err := p.redisCache.Get(ctx, key, &productCache); !errors.Is(err, redis.Nil) {
		return p.toProductResp(&productCache), nil
	}

	product, err := p.productRepository.GetProductById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	_ = p.redisCache.Set(ctx, key, product, defaultTTL)

	return p.toProductResp(product), nil
}

func (p *productService) UpdateProduct(ctx context.Context, id uint, req *dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	product, err := p.productRepository.GetProductById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if req.CategoryId != nil {
		product.CategoryId = *req.CategoryId
	}

	if req.Name != nil {
		product.Name = *req.Name
	}

	if req.Description != nil {
		product.Description = *req.Description
	}

	if req.Price != nil {
		product.Price = *req.Price
	}

	if req.Stock != nil {
		product.Stock = *req.Stock
	}

	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if err := p.productRepository.UpdateProduct(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	_ = p.redisCache.Delete(ctx, fmt.Sprintf(productByIdKey, id))
	_ = p.redisCache.DeletePattern(ctx, productListPattern)

	return p.toProductResp(product), nil
}

func (p *productService) DeleteProduct(ctx context.Context, id uint) error {
	if err := p.productRepository.DeleteProduct(ctx, id); err != nil {
		return err
	}

	_ = p.redisCache.Delete(ctx, fmt.Sprintf(productByIdKey, id))
	_ = p.redisCache.DeletePattern(ctx, productListPattern)

	return nil
}

func (p *productService) AddProductImage(ctx context.Context, productId uint, url, altText string) (*dto.ProductImageResponse, error) {
	count, err := p.productRepository.GetProductImageCount(ctx, productId)
	if err != nil {
		return nil, fmt.Errorf("failed to count images: %w", err)
	}

	image := &domain.ProductImage{
		ProductId: productId,
		URL:       url,
		AltText:   altText,
		IsPrimary: count == 0,
	}

	if err := p.productRepository.CreateProductImage(ctx, image); err != nil {
		return nil, fmt.Errorf("failed to create image: %w", err)
	}

	return &dto.ProductImageResponse{
		Id:        image.Id,
		URL:       image.URL,
		AltText:   image.AltText,
		IsPrimary: image.IsPrimary,
		CreatedAt: image.CreatedAt,
	}, nil
}

func (p *productService) SearchProducts(ctx context.Context, req *dto.SearchProductsRequest) ([]*dto.ProductSearchResult, *helper.PaginatedMeta, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}

	offset := (req.Page - 1) * req.Limit

	products, ranks, total, err := p.productRepository.SearchProducts(ctx, req, offset, req.Limit)
	if err != nil {
		return nil, nil, err
	}

	results := make([]*dto.ProductSearchResult, len(products))
	for i := range products {
		results[i] = &dto.ProductSearchResult{
			// Use the existing helper (renamed for clarity if you prefer, but here it's toProductResp)
			ProductResponse: *p.toProductResp(products[i]),
			Rank:            ranks[i],
		}
	}

	// Reuse buildMeta – eliminates duplicate pagination arithmetic
	meta := p.buildMeta(req.Page, req.Limit, total)

	return results, meta, nil
}

func (p *productService) toProductResp(product *domain.Product) *dto.ProductResponse {
	images := make([]dto.ProductImageResponse, len(product.Images))
	for i := range product.Images {
		images[i] = dto.ProductImageResponse{
			Id:        product.Images[i].Id,
			URL:       product.Images[i].URL,
			AltText:   product.Images[i].AltText,
			IsPrimary: product.Images[i].IsPrimary,
			CreatedAt: product.Images[i].CreatedAt,
		}
	}

	return &dto.ProductResponse{
		Id:          product.Id,
		CategoryId:  product.CategoryId,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		SKU:         product.SKU,
		IsActive:    product.IsActive,
		Category: dto.CategoryResponse{
			Id:          product.Category.Id,
			Name:        product.Category.Name,
			Description: product.Category.Description,
			IsActive:    product.Category.IsActive,
			CreatedAt:   product.Category.CreatedAt,
			UpdatedAt:   product.Category.UpdatedAt,
		},
		Images:    images,
		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
	}
}

func (p *productService) toProductRespList(products []*domain.Product) []*dto.ProductResponse {
	resp := make([]*dto.ProductResponse, len(products))
	for i, prod := range products {
		resp[i] = p.toProductResp(prod)
	}
	return resp
}

func (p *productService) buildMeta(page, limit int, total int64) *helper.PaginatedMeta {
	return &helper.PaginatedMeta{
		Page:      page,
		Limit:     limit,
		Total:     total,
		TotalPage: int((total + int64(limit) - 1) / int64(limit)),
	}
}

func NewProductService(productRepository repository.ProductRepository, redisCache repository.RedisCache) ProductService {
	return &productService{
		productRepository: productRepository,
		redisCache:        redisCache,
	}
}
