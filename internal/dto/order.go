package dto

import "time"

type OrderResponse struct {
	Id          uint                `json:"id"`
	UserId      uint                `json:"user_id"`
	Status      string              `json:"status"`
	TotalAmount float64             `json:"total_amount"`
	OrderItems  []OrderItemResponse `json:"order_items"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

type OrderItemResponse struct {
	Id        uint            `json:"id"`
	Product   ProductResponse `json:"product"`
	Quantity  int             `json:"quantity"`
	Price     float64         `json:"price"`
	CreatedAt time.Time       `json:"created_at"`
}
