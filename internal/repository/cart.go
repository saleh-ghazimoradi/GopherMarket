package repository

import (
	"context"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"gorm.io/gorm"
)

type CartRepository interface {
	CreateCart(ctx context.Context, cart *domain.Cart) error
	WithTx(tx *gorm.DB) CartRepository
}

type cartRepository struct {
	dbWrite *gorm.DB
	dbRead  *gorm.DB
}

func (c *cartRepository) CreateCart(ctx context.Context, cart *domain.Cart) error {
	return c.dbWrite.WithContext(ctx).Create(cart).Error
}

func (c *cartRepository) WithTx(tx *gorm.DB) CartRepository {
	return &cartRepository{
		dbWrite: tx,
		dbRead:  tx,
	}
}

func NewCartRepository(dbWrite, dbRead *gorm.DB) CartRepository {
	return &cartRepository{
		dbWrite: dbWrite,
		dbRead:  dbRead,
	}
}
