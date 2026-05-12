package repository

import (
	"context"
	"errors"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"gorm.io/gorm"
)

type CartRepository interface {
	CreateCart(ctx context.Context, cart *domain.Cart) error
	GetCartByUserId(ctx context.Context, userId uint) (*domain.Cart, error)
	GetCartWithItemsAndProducts(ctx context.Context, userId uint) (*domain.Cart, error)
	GetOrCreateCart(ctx context.Context, userId uint) (*domain.Cart, error)
	WithTx(tx *gorm.DB) CartRepository
}

type cartRepository struct {
	dbWrite *gorm.DB
	dbRead  *gorm.DB
}

func (c *cartRepository) CreateCart(ctx context.Context, cart *domain.Cart) error {
	return c.dbWrite.WithContext(ctx).Create(cart).Error
}

func (c *cartRepository) GetCartByUserId(ctx context.Context, userId uint) (*domain.Cart, error) {
	var cart domain.Cart
	if err := c.dbRead.WithContext(ctx).Preload("CartItems.Product.Category").First(&cart, "user_id = ?", userId).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrsNotFound
		default:
			return nil, err
		}
	}
	return &cart, nil
}

func (c *cartRepository) GetCartWithItemsAndProducts(ctx context.Context, userId uint) (*domain.Cart, error) {
	var cart domain.Cart
	if err := c.dbRead.WithContext(ctx).Preload("CartItems.Product").Where("user_id = ?", userId).First(&cart).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrsNotFound
		default:
			return nil, err
		}
	}
	return &cart, nil
}

func (c *cartRepository) GetOrCreateCart(ctx context.Context, userId uint) (*domain.Cart, error) {
	var cart domain.Cart

	if err := c.dbRead.WithContext(ctx).Where("user_id = ?", userId).First(&cart).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			cart = domain.Cart{UserId: userId}
			if err := c.dbWrite.WithContext(ctx).Create(&cart).Error; err != nil {
				return nil, err
			}
		}
		return nil, err
	}

	return &cart, nil
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
