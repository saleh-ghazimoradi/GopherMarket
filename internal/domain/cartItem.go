package domain

import (
	"gorm.io/gorm"
	"time"
)

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
