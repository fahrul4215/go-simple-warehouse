package product

import "time"

type Product struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	SKU         string    `gorm:"unique;not null" json:"sku" validate:"required,alphanum"`
	Name        string    `gorm:"not null" json:"name" validate:"required"`
	Description string    `gorm:"type:text" json:"description"`
	Price       float64   `gorm:"not null" json:"price" validate:"gte=0"`
	Quantity    int       `gorm:"not null" json:"quantity" validate:"gte=0"`
	Location    string    `gorm:"not null" json:"location" validate:"required"`
	Status      string    `gorm:"not null" json:"status" validate:"required,oneof=active inactive"`
	BarcodeURL  string    `gorm:"type:text" json:"barcode_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
