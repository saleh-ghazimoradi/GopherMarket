package dto

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"time"
)

type CreateDiscountRequest struct {
	ProductId     uint                `json:"product_id"`
	DiscountType  domain.DiscountType `json:"discount_type"`
	DiscountValue float64             `json:"discount_value"`
	StartTime     time.Time           `json:"start_time"`
	EndTime       time.Time           `json:"end_time"`
}

type DiscountResponse struct {
	Id            uint                `json:"id"`
	ProductId     uint                `json:"product_id"`
	DiscountType  domain.DiscountType `json:"discount_type"`
	DiscountValue float64             `json:"discount_value"`
	StartTime     time.Time           `json:"start_time"`
	EndTime       time.Time           `json:"end_time"`
	CreatedAt     time.Time           `json:"created_at"`
}
