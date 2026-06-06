package domain

import (
	"gorm.io/gorm"
	"time"
)

type Product struct {
	Id           uint   `gorm:"primaryKey"`
	CategoryId   uint   `gorm:"not null"`
	Name         string `gorm:"not null"`
	Description  string
	Price        float64 `gorm:"not null"`
	Stock        int     `gorm:"default:0"`
	SKU          string  `gorm:"uniqueIndex;not null"`
	IsActive     bool    `gorm:"default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	SearchVector string         `gorm:"column:search_vector;->"`

	Category   Category
	Images     []ProductImage
	OrderItems []OrderItem
	CartItems  []CartItem
	Discounts  []Discount
}
