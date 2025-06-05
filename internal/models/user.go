package models

import (
	"time"
)

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"unique;not null"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Password  string    `json:"-" gorm:"not null"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role" gorm:"default:'user'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserResponse là cấu trúc response khi trả về thông tin user
type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginRequest là cấu trúc request khi đăng nhập
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest là cấu trúc request khi đăng ký
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
}

// UserQueryParams là cấu trúc cho các tham số tìm kiếm và phân trang user
type UserQueryParams struct {
	// Phân trang
	Page  int `form:"page"`
	Limit int `form:"limit" binding:"max=100"`

	// Tìm kiếm (optional - có thể mở rộng sau)
	Search string `form:"search"`
	Role   string `form:"role"` // admin, user
}

// ChangePasswordRequest represents the request body for changing password
type ChangePasswordRequest struct {
	CurrentPassword    string `json:"current_password" binding:"required"`
	NewPassword       string `json:"new_password" binding:"required,min=6"`
	ConfirmNewPassword string `json:"confirm_new_password" binding:"required,eqfield=NewPassword"`
}
