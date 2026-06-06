package service

import (
	"context"
	"fmt"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"time"
)

type DiscountService interface {
	CreateDiscount(ctx context.Context, req *dto.CreateDiscountRequest) (*dto.DiscountResponse, error)
	DeleteDiscount(ctx context.Context, id uint, productId uint) error
}

type discountService struct {
	discountRepository repository.DiscountRepository
	redisCache         repository.RedisCache
	tracer             trace.Tracer
}

func (d *discountService) CreateDiscount(ctx context.Context, req *dto.CreateDiscountRequest) (*dto.DiscountResponse, error) {
	ctx, span := d.tracer.Start(ctx, "DiscountService.CreateDiscount",
		trace.WithAttributes(
			attribute.Int64("discount.product_id", int64(req.ProductId)),
			attribute.String("discount.type", string(req.DiscountType)),
		))
	defer span.End()

	discount := &domain.Discount{
		ProductId:     req.ProductId,
		DiscountType:  req.DiscountType,
		DiscountValue: req.DiscountValue,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
	}

	if err := d.discountRepository.CreateDiscount(ctx, discount); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create discount")
		return nil, fmt.Errorf("failed to create discount: %w", err)
	}

	_ = d.redisCache.Delete(ctx, fmt.Sprintf(productByIdKey, req.ProductId))
	_ = d.redisCache.DeletePattern(ctx, productListPattern)

	return &dto.DiscountResponse{
		Id:            discount.Id,
		ProductId:     discount.ProductId,
		DiscountType:  discount.DiscountType,
		DiscountValue: discount.DiscountValue,
		StartTime:     discount.StartTime,
		EndTime:       discount.EndTime,
		CreatedAt:     discount.CreatedAt,
	}, nil
}

func (d *discountService) DeleteDiscount(ctx context.Context, id uint, productId uint) error {
	ctx, span := d.tracer.Start(ctx, "DiscountService.DeleteDiscount",
		trace.WithAttributes(attribute.Int64("discount.id", int64(id))))
	defer span.End()

	if err := d.discountRepository.DeleteDiscount(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete discount")
		return fmt.Errorf("failed to delete discount: %w", err)
	}

	_ = d.redisCache.Delete(ctx, fmt.Sprintf(productByIdKey, productId))
	_ = d.redisCache.DeletePattern(ctx, productListPattern)

	return nil
}

func IsDiscountActive(t time.Time, discount *domain.Discount) bool {
	if discount == nil {
		return false
	}
	tUTC := t.UTC()
	return !tUTC.Before(discount.StartTime.UTC()) && !tUTC.After(discount.EndTime.UTC())
}

func CalculateDiscountPrice(discount *domain.Discount, basePrice float64) float64 {
	if discount == nil || !IsDiscountActive(time.Now(), discount) {
		return basePrice
	}

	switch discount.DiscountType {
	case domain.DiscountPercentage:
		discounted := basePrice * (discount.DiscountValue / 100.0)
		return basePrice - discounted
	case domain.DiscountFixed:
		if discount.DiscountValue >= basePrice {
			return 0.0
		}
		return basePrice - discount.DiscountValue
	default:
		return basePrice
	}
}

func NewDiscountService(discountRepository repository.DiscountRepository, redisCache repository.RedisCache, tracer trace.Tracer) DiscountService {
	return &discountService{
		discountRepository: discountRepository,
		redisCache:         redisCache,
		tracer:             tracer,
	}
}
