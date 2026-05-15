package repository

import (
	"context"
	"errors"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"gorm.io/gorm"
	"strings"
)

type ProductRepository interface {
	CreateProduct(ctx context.Context, product *domain.Product) error
	CreateProductImage(ctx context.Context, image *domain.ProductImage) error
	GetProductById(ctx context.Context, id uint) (*domain.Product, error)
	GetProducts(ctx context.Context, offset, limit int) ([]*domain.Product, error)
	GetProductImageCount(ctx context.Context, id uint) (int64, error)
	CountActiveProducts(ctx context.Context) (int64, error)
	UpdateProduct(ctx context.Context, product *domain.Product) error
	DeleteProduct(ctx context.Context, id uint) error
	SearchProducts(ctx context.Context, req *dto.SearchProductsRequest, offset, limit int) ([]*domain.Product, []float32, int64, error)
	WithTx(tx *gorm.DB) ProductRepository
}

type productRepository struct {
	dbWrite *gorm.DB
	dbRead  *gorm.DB
}

func (p *productRepository) CreateProduct(ctx context.Context, product *domain.Product) error {
	return p.dbWrite.WithContext(ctx).Create(product).Error
}

func (p *productRepository) CreateProductImage(ctx context.Context, image *domain.ProductImage) error {
	return p.dbWrite.WithContext(ctx).Create(image).Error
}

func (p *productRepository) GetProductById(ctx context.Context, id uint) (*domain.Product, error) {
	var product domain.Product
	if err := p.dbRead.WithContext(ctx).Preload("Category").Preload("Images").First(&product, id).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrsNotFound
		default:
			return nil, err
		}
	}
	return &product, nil
}

func (p *productRepository) GetProducts(ctx context.Context, offset, limit int) ([]*domain.Product, error) {
	var products []*domain.Product
	if err := p.dbRead.WithContext(ctx).Preload("Category").Preload("Images").Where("is_active = ?", true).Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (p *productRepository) GetProductImageCount(ctx context.Context, id uint) (int64, error) {
	var count int64
	if err := p.dbRead.WithContext(ctx).Model(&domain.ProductImage{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (p *productRepository) CountActiveProducts(ctx context.Context) (int64, error) {
	var total int64
	if err := p.dbRead.WithContext(ctx).Model(&domain.Product{}).Where("is_active = ?", true).Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

func (p *productRepository) UpdateProduct(ctx context.Context, product *domain.Product) error {
	return p.dbWrite.WithContext(ctx).Save(product).Error
}

func (p *productRepository) DeleteProduct(ctx context.Context, id uint) error {
	return p.dbWrite.WithContext(ctx).Delete(&domain.Product{}, id).Error
}

func (p *productRepository) SearchProducts(ctx context.Context, req *dto.SearchProductsRequest, offset, limit int) ([]*domain.Product, []float32, int64, error) {
	db := p.dbRead.
		WithContext(ctx).
		Model(&domain.Product{}).
		Where("is_active = ?", true)

	// Apply full-text search only if the query is not empty.
	// An empty query means "list all, apply filters only".
	hasQuery := strings.TrimSpace(req.Query) != ""
	if hasQuery {
		db = db.Where("search_vector @@ plainto_tsquery('english', ?)", req.Query).
			Select("products.*, ts_rank(search_vector, plainto_tsquery('english', ?)) AS rank", req.Query)
	} else {
		db = db.Select("products.*, 0 AS rank")
	}

	if req.CategoryId != nil {
		db = db.Where("category_id = ?", *req.CategoryId)
	}
	if req.MinPrice != nil {
		db = db.Where("price >= ?", *req.MinPrice)
	}
	if req.MaxPrice != nil {
		db = db.Where("price <= ?", *req.MaxPrice)
	}

	// Count total matching rows (before pagination)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, nil, 0, err
	}

	// Local struct to capture the extra `rank` column
	type productWithRank struct {
		domain.Product
		Rank float32 `gorm:"column:rank"`
	}

	var rows []productWithRank
	if err := db.
		Preload("Category").
		Preload("Images").
		Order("rank DESC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, nil, 0, err
	}

	products := make([]*domain.Product, len(rows))
	ranks := make([]float32, len(rows))
	for i := range rows {
		products[i] = &rows[i].Product
		ranks[i] = rows[i].Rank
	}

	return products, ranks, total, nil
}

func (p *productRepository) WithTx(tx *gorm.DB) ProductRepository {
	return &productRepository{
		dbWrite: tx,
		dbRead:  tx,
	}
}

func NewProductRepository(dbWrite, dbRead *gorm.DB) ProductRepository {
	return &productRepository{
		dbWrite: dbWrite,
		dbRead:  dbRead}
}
