package dto

import "time"

type AddToCartRequest struct {
	ProductId uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity"`
}

type CartResponse struct {
	Id        uint               `json:"id"`
	UserId    uint               `json:"user_id"`
	CartItems []CartItemResponse `json:"cart_items"`
	Total     float64            `json:"total"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type CartItemResponse struct {
	Id        uint            `json:"id"`
	Product   ProductResponse `json:"product"`
	Quantity  int             `json:"quantity"`
	Subtotal  float64         `json:"subtotal"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
