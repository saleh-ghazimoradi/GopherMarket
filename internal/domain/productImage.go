package domain

import (
	"gorm.io/gorm"
	"time"
)

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
