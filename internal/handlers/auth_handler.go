package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/NgTruong624/project_backend/internal/models"
	"github.com/NgTruong624/project_backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db        *gorm.DB
	jwtSecret string
}

func NewAuthHandler(db *gorm.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

// Register xử lý đăng ký user mới
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid request", err.Error()))
		return
	}

	// Kiểm tra email đã tồn tại
	var existingUser models.User
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Email already exists", ""))
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Error hashing password", err.Error()))
		return
	}

	// Tạo user mới
	user := models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		FullName:  req.FullName,
		Role:      "user", // Mặc định là user
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Error creating user", err.Error()))
		return
	}

	// Tạo response không bao gồm password
	userResponse := models.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FullName:  user.FullName,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}

	c.JSON(http.StatusCreated, utils.NewResponse(http.StatusCreated, "User registered successfully", userResponse))
}

// Login xử lý đăng nhập
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid request", err.Error()))
		return
	}

	// Tìm user theo username
	var user models.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse(http.StatusUnauthorized, "Invalid username or password", ""))
		return
	}

	// Kiểm tra password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse(http.StatusUnauthorized, "Invalid username or password", ""))
		return
	}

	// Tạo JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // Token hết hạn sau 24 giờ
	})

	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Error generating token", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.NewResponse(http.StatusOK, "Login successful", gin.H{
		"token": tokenString,
		"user": models.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FullName:  user.FullName,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		},
	}))
}

// ChangePassword handles the password change request
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Handle validation errors
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors := make(map[string]string)
			for _, e := range validationErrors {
				field := e.Field()
				switch field {
				case "CurrentPassword":
					if e.Tag() == "required" {
						errors["current_password"] = "Current password is required."
					}
				case "NewPassword":
					switch e.Tag() {
					case "required":
						errors["new_password"] = "New password is required."
					case "min":
						errors["new_password"] = fmt.Sprintf("New password must be at least %s characters long.", e.Param())
					}
				case "ConfirmNewPassword":
					switch e.Tag() {
					case "required":
						errors["confirm_new_password"] = "Confirm password is required."
					case "eqfield":
						errors["confirm_new_password"] = "Confirm password must match new password."
					}
				}
			}
			c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Validation failed", errors))
			return
		}
		// Handle non-validation errors
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid request data", err.Error()))
		return
	}

	// Get user ID from context (set by JWTMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse(http.StatusUnauthorized, "User not authenticated", ""))
		return
	}

	// Get user from database
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse(http.StatusNotFound, "User not found", ""))
		return
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Current password is incorrect", ""))
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Failed to hash new password", err.Error()))
		return
	}

	// Update password in database
	if err := h.db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Failed to update password", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.NewResponse(http.StatusOK, "Password changed successfully", nil))
}
