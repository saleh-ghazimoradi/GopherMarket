package repository

import (
	"context"
	"errors"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"gorm.io/gorm"
)

type CartItemRepository interface {
	GetCartItem(ctx context.Context, cartId, productId uint) (*domain.CartItem, error)
	CreateCartItem(ctx context.Context, item *domain.CartItem) error
	UpdateCartItem(ctx context.Context, item *domain.CartItem) error
	GetCartItemWithUser(ctx context.Context, userId, itemId uint) (*domain.CartItem, error)
	DeleteCartItem(ctx context.Context, userId, itemId uint) error
	ClearCartItems(ctx context.Context, cartId uint) error
	WithTx(tx *gorm.DB) CartItemRepository
}

type cartItemRepository struct {
	dbWrite *gorm.DB
	dbRead  *gorm.DB
}

func (c *cartItemRepository) GetCartItem(ctx context.Context, cartId, productId uint) (*domain.CartItem, error) {
	var cartItem domain.CartItem
	err := c.dbRead.WithContext(ctx).Where("cart_id = ? and product_id = ?", cartId, productId).First(&cartItem).Error
	return &cartItem, err
}

func (c *cartItemRepository) CreateCartItem(ctx context.Context, item *domain.CartItem) error {
	return c.dbWrite.WithContext(ctx).Create(item).Error
}

func (c *cartItemRepository) UpdateCartItem(ctx context.Context, item *domain.CartItem) error {
	return c.dbWrite.WithContext(ctx).Save(item).Error
}

func (c *cartItemRepository) GetCartItemWithUser(ctx context.Context, userId, itemId uint) (*domain.CartItem, error) {
	var item domain.CartItem
	if err := c.dbRead.WithContext(ctx).Joins("JOIN carts ON cart_items.cart_id = carts.id").
		Where("cart_items.id = ? AND carts.user_id = ?", itemId, userId).
		First(&item).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrsNotFound
		default:
			return nil, err
		}
	}
	return &item, nil
}

func (c *cartItemRepository) DeleteCartItem(ctx context.Context, userId, itemId uint) error {
	return c.dbWrite.WithContext(ctx).Unscoped().Where(
		"id = ? AND cart_id IN (?)",
		itemId,
		c.dbRead.Select("id").Table("carts").Where("user_id = ?", userId),
	).Delete(&domain.CartItem{}).Error
}

func (c *cartItemRepository) ClearCartItems(ctx context.Context, cartId uint) error {
	return c.dbWrite.WithContext(ctx).Unscoped().Where("cart_id = ?", cartId).Delete(&domain.CartItem{}).Error
}

func (c *cartItemRepository) WithTx(tx *gorm.DB) CartItemRepository {
	return &cartItemRepository{
		dbWrite: tx,
		dbRead:  tx,
	}
}

func NewCartItemRepository(dbWrite, dbRead *gorm.DB) CartItemRepository {
	return &cartItemRepository{
		dbWrite: dbWrite,
		dbRead:  dbRead,
	}
}
