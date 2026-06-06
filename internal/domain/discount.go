package domain

import (
	"gorm.io/gorm"
	"time"
)

type DiscountType string

const (
	DiscountPercentage DiscountType = "percentage"
	DiscountFixed      DiscountType = "fixed"
)

type Discount struct {
	Id            uint         `gorm:"primaryKey"`
	ProductId     uint         `gorm:"not null;index"`
	DiscountType  DiscountType `gorm:"type:varchar(20);not null;default:'percentage'"`
	DiscountValue float64      `gorm:"type:numeric(10,2);not null"`
	StartTime     time.Time    `gorm:"not null"`
	EndTime       time.Time    `gorm:"not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`

	Product Product
}
