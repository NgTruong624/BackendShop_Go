package models

import (
	"time"
)

type Product struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Price       float64   `json:"price" gorm:"not null"`
	Stock       int       `json:"stock" gorm:"not null"`
	ImageURL    string    `json:"image_url"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProductResponse là cấu trúc response khi trả về thông tin sản phẩm
type ProductResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	ImageURL    string    `json:"image_url"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateProductRequest là cấu trúc request khi tạo sản phẩm mới
type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,min=0"`
	Stock       int     `json:"stock" binding:"required,min=0"`
	ImageURL    string  `json:"image_url"`
	Category    string  `json:"category"`
}

// UpdateProductRequest là cấu trúc request khi cập nhật sản phẩm
type UpdateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"min=0"`
	Stock       int     `json:"stock" binding:"min=0"`
	ImageURL    string  `json:"image_url"`
	Category    string  `json:"category"`
}

// ProductQueryParams là cấu trúc cho các tham số tìm kiếm và phân trang
type ProductQueryParams struct {
	// Tìm kiếm cơ bản
	Search   string `form:"search"`
	Category string `form:"category"`

	// Tìm kiếm theo giá
	MinPrice float64 `form:"min_price"`
	MaxPrice float64 `form:"max_price"`

	// Tìm kiếm theo tồn kho
	InStock bool `form:"in_stock"`

	// Tìm kiếm theo thời gian
	StartDate time.Time `form:"start_date"`
	EndDate   time.Time `form:"end_date"`

	// Sắp xếp
	SortBy string `form:"sort_by"` // price, name, created_at, stock, category
	Order  string `form:"order"`   // asc, desc

	// Phân trang
	Page  int `form:"page"`
	Limit int `form:"limit" binding:"max=100"`
}
