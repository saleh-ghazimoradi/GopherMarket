package domain

import (
	"gorm.io/gorm"
	"time"
)

type Category struct {
	Id          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Description string
	IsActive    bool `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	Products []Product
}

type Product struct {
	Id          uint   `gorm:"primaryKey"`
	CategoryId  uint   `gorm:"not null"`
	Name        string `gorm:"not null"`
	Description string
	Price       float64 `gorm:"not null"`
	Stock       int     `gorm:"default:0"`
	SKU         string  `gorm:"uniqueIndex;not null"`
	IsActive    bool    `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	Category   Category
	Images     []ProductImage
	OrderItems []OrderItem
	CartItems  []CartItem
}

type ProductImage struct {
	Id        uint   `gorm:"primaryKey"`
	ProductId uint   `gorm:"not null"`
	URL       string `gorm:"not null"`
	AltText   string
	IsPrimary bool `gorm:"default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Product Product
}
