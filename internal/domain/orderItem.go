package domain

import (
	"gorm.io/gorm"
	"time"
)

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
