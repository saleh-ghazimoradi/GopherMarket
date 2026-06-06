package repository

import (
	"context"
	"errors"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"gorm.io/gorm"
	"time"
)

type DiscountRepository interface {
	CreateDiscount(ctx context.Context, discount *domain.Discount) error
	GetActiveByProductId(ctx context.Context, productId uint) (*domain.Discount, error)
	DeleteDiscount(ctx context.Context, id uint) error
	WithTx(tx *gorm.DB) DiscountRepository
}

type discountRepository struct {
	dbWrite *gorm.DB
	dbRead  *gorm.DB
}

func (d *discountRepository) CreateDiscount(ctx context.Context, discount *domain.Discount) error {
	return d.dbWrite.WithContext(ctx).Create(discount).Error
}

func (d *discountRepository) GetActiveByProductId(ctx context.Context, productId uint) (*domain.Discount, error) {
	var discount domain.Discount
	if err := d.dbRead.WithContext(ctx).Where("product_id = ? AND start_time <= ? AND end_time >= ?", productId, time.Now().UTC(), time.Now().UTC()).First(&discount).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrsNotFound
		default:
			return nil, err
		}
	}
	return &discount, nil
}

func (d *discountRepository) DeleteDiscount(ctx context.Context, id uint) error {
	return d.dbWrite.WithContext(ctx).Delete(&domain.Discount{}, id).Error
}

func (d *discountRepository) WithTx(tx *gorm.DB) DiscountRepository {
	return &discountRepository{
		dbWrite: tx,
		dbRead:  tx,
	}
}

func NewDiscountRepository(dbWrite, dbRead *gorm.DB) DiscountRepository {
	return &discountRepository{
		dbWrite: dbWrite,
		dbRead:  dbRead,
	}
}
