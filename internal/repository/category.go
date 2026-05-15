package repository

import (
	"context"
	"errors"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	CreateCategory(ctx context.Context, category *domain.Category) error
	GetCategoryById(ctx context.Context, id uint) (*domain.Category, error)
	GetCategories(ctx context.Context) ([]*domain.Category, error)
	UpdateCategory(ctx context.Context, category *domain.Category) error
	DeleteCategory(ctx context.Context, id uint) error
	WithTx(tx *gorm.DB) CategoryRepository
}

type categoryRepository struct {
	dbWrite *gorm.DB
	dbRead  *gorm.DB
}

func (c *categoryRepository) CreateCategory(ctx context.Context, category *domain.Category) error {
	return c.dbWrite.WithContext(ctx).Create(category).Error
}

func (c *categoryRepository) GetCategoryById(ctx context.Context, id uint) (*domain.Category, error) {
	var category domain.Category
	if err := c.dbRead.WithContext(ctx).First(&category, id).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrsNotFound
		default:
			return nil, err
		}
	}
	return &category, nil
}

func (c *categoryRepository) GetCategories(ctx context.Context) ([]*domain.Category, error) {
	var categories []*domain.Category

	if err := c.dbRead.WithContext(ctx).Where("is_active = ?", true).Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (c *categoryRepository) UpdateCategory(ctx context.Context, category *domain.Category) error {
	return c.dbWrite.WithContext(ctx).Save(category).Error
}

func (c *categoryRepository) DeleteCategory(ctx context.Context, id uint) error {
	return c.dbWrite.WithContext(ctx).Delete(&domain.Category{}, id).Error
}

func (c *categoryRepository) WithTx(tx *gorm.DB) CategoryRepository {
	return &categoryRepository{
		dbWrite: tx,
		dbRead:  tx,
	}
}

func NewCategoryRepository(dbWrite, dbRead *gorm.DB) CategoryRepository {
	return &categoryRepository{
		dbWrite: dbWrite,
		dbRead:  dbRead,
	}
}
