package repository

import (
	"context"
	"errors"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"gorm.io/gorm"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *domain.Order) error
	GetOrderByUserId(ctx context.Context, userId uint, orderId uint) (*domain.Order, error)
	GetOrderById(ctx context.Context, id uint) (*domain.Order, error)
	GetUserOrders(ctx context.Context, userId uint, offset, limit int) ([]*domain.Order, error)
	CountUserOrders(ctx context.Context, userId uint) (int64, error)
	WithTx(tx *gorm.DB) OrderRepository
}

func (o *orderRepository) CreateOrder(ctx context.Context, order *domain.Order) error {
	return o.dbWrite.WithContext(ctx).Create(order).Error
}

func (o *orderRepository) GetOrderByUserId(ctx context.Context, userId, orderId uint) (*domain.Order, error) {
	var order domain.Order
	if err := o.dbRead.WithContext(ctx).Preload("OrderItems.Product.Category").Where("id = ? AND user_id = ?", orderId, userId).First(&order).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrsNotFound
		default:
			return nil, err
		}
	}
	return &order, nil
}

func (o *orderRepository) GetOrderById(ctx context.Context, id uint) (*domain.Order, error) {
	var order domain.Order
	if err := o.dbRead.WithContext(ctx).Preload("OrderItems.Product.Category").First(&order, id).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrsNotFound
		default:
			return nil, err
		}
	}
	return &order, nil
}

func (o *orderRepository) GetUserOrders(ctx context.Context, userId uint, offset, limit int) ([]*domain.Order, error) {
	var orders []*domain.Order
	if err := o.dbRead.WithContext(ctx).Preload("OrderItems.Product.Category").Where("user_id = ?", userId).Order("created_at desc").Offset(offset).Limit(limit).Find(&orders).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrsNotFound
		default:
			return nil, err
		}
	}
	return orders, nil
}

func (o *orderRepository) CountUserOrders(ctx context.Context, userId uint) (int64, error) {
	var count int64
	if err := o.dbRead.WithContext(ctx).Model(&domain.Order{}).Where("user_id = ?", userId).Count(&count).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return 0, ErrsNotFound
		default:
			return 0, err
		}
	}
	return count, nil
}

type orderRepository struct {
	dbWrite *gorm.DB
	dbRead  *gorm.DB
}

func (o *orderRepository) WithTx(tx *gorm.DB) OrderRepository {
	return &orderRepository{
		dbWrite: tx,
		dbRead:  tx,
	}
}

func NewOrderRepository(dbWrite, dbRead *gorm.DB) OrderRepository {
	return &orderRepository{
		dbWrite: dbWrite,
		dbRead:  dbRead,
	}
}
