package handlers

import (
	"net/http"

	"github.com/NgTruong624/project_backend/internal/models"
	"github.com/NgTruong624/project_backend/internal/repository"
	"github.com/NgTruong624/project_backend/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminHandler struct {
	userRepo *repository.UserRepository
}

func NewAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{
		userRepo: repository.NewUserRepository(db),
	}
}

// GetUsersList lấy danh sách tất cả người dùng (Admin only)
func (h *AdminHandler) GetUsersList(c *gin.Context) {
	// Kiểm tra quyền admin
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, utils.NewErrorResponse(http.StatusForbidden, "Permission denied", "Only admin can access user list"))
		return
	}

	var query models.UserQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid query parameters", err.Error()))
		return
	}

	// Set default values
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 10
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	// Get users from repository
	users, total, err := h.userRepo.GetAllUsers(&query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Error fetching users", err.Error()))
		return
	}

	// Convert to response (remove password field)
	var userResponses []models.UserResponse
	for _, u := range users {
		userResponses = append(userResponses, models.UserResponse{
			ID:        u.ID,
			Username:  u.Username,
			Email:     u.Email,
			FullName:  u.FullName,
			Role:      u.Role,
			CreatedAt: u.CreatedAt,
		})
	}

	// Calculate pagination info
	totalPages := (int(total) + query.Limit - 1) / query.Limit

	// Prepare metadata
	meta := map[string]interface{}{
		"total":        total,
		"total_pages":  totalPages,
		"current_page": query.Page,
		"per_page":     query.Limit,
		"has_next":     query.Page < totalPages,
		"has_prev":     query.Page > 1,
	}

	// Add filter info to metadata
	if query.Search != "" {
		meta["search"] = query.Search
	}
	if query.Role != "" {
		meta["role"] = query.Role
	}

	c.JSON(http.StatusOK, utils.NewPaginatedResponse(
		http.StatusOK,
		"Users retrieved successfully",
		userResponses,
		query.Page,
		totalPages,
		total,
		query.Limit,
		meta,
	))
}
