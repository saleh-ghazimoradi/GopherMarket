package domain

import (
	"gorm.io/gorm"
	"time"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	Id          uint        `gorm:"primaryKey"`
	UserId      uint        `gorm:"not null"`
	Status      OrderStatus `gorm:"default:pending"`
	TotalAmount float64     `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	User       User
	OrderItems []OrderItem
}

type OrderItem struct {
	Id        uint    `gorm:"primaryKey"`
	OrderId   uint    `gorm:"not null"`
	ProductId uint    `gorm:"not null"`
	Quantity  int     `gorm:"not null"`
	Price     float64 `gorm:"not null"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Order   Order
	Product Product
}

type Cart struct {
	Id        uint `gorm:"primaryKey"`
	UserId    uint `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	CartItems []CartItem
}

type CartItem struct {
	Id        uint `gorm:"primaryKey"`
	CartId    uint `gorm:"not null"`
	ProductId uint `gorm:"not null"`
	Quantity  int  `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Cart    Cart
	Product Product
}
