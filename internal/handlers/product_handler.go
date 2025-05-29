package handlers

import (
	"net/http"

	"github.com/NgTruong624/project_backend/internal/models"
	"github.com/NgTruong624/project_backend/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductHandler struct {
	db *gorm.DB
}

func NewProductHandler(db *gorm.DB) *ProductHandler {
	return &ProductHandler{
		db: db,
	}
}

// GetProducts lấy danh sách sản phẩm (Public)
func (h *ProductHandler) GetProducts(c *gin.Context) {
	var query models.ProductQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid query parameters", err.Error()))
		return
	}

	// Set default values
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 {
		query.Limit = 10
	}

	// Build query
	dbQuery := h.db.Model(&models.Product{})

	// Apply filters
	if query.Search != "" {
		dbQuery = dbQuery.Where("name ILIKE ? OR description ILIKE ?", "%"+query.Search+"%", "%"+query.Search+"%")
	}
	if query.Category != "" {
		dbQuery = dbQuery.Where("category = ?", query.Category)
	}

	// Get total count
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Error counting products", err.Error()))
		return
	}

	// Apply sorting
	if query.SortBy != "" {
		order := "ASC"
		if query.Order == "desc" {
			order = "DESC"
		}
		dbQuery = dbQuery.Order(query.SortBy + " " + order)
	}

	// Apply pagination
	offset := (query.Page - 1) * query.Limit
	var products []models.Product
	if err := dbQuery.Offset(offset).Limit(query.Limit).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Error fetching products", err.Error()))
		return
	}

	// Convert to response
	var productResponses []models.ProductResponse
	for _, p := range products {
		productResponses = append(productResponses, models.ProductResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       p.Stock,
			ImageURL:    p.ImageURL,
			Category:    p.Category,
			CreatedAt:   p.CreatedAt,
		})
	}

	totalPages := (int(total) + query.Limit - 1) / query.Limit
	c.JSON(http.StatusOK, utils.NewPaginatedResponse(
		http.StatusOK,
		"Products retrieved successfully",
		productResponses,
		query.Page,
		totalPages,
		total,
		query.Limit,
	))
}

// GetProduct lấy chi tiết sản phẩm (Public)
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product
	if err := h.db.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse(http.StatusNotFound, "Product not found", ""))
		return
	}

	productResponse := models.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		ImageURL:    product.ImageURL,
		Category:    product.Category,
		CreatedAt:   product.CreatedAt,
	}

	c.JSON(http.StatusOK, utils.NewResponse(http.StatusOK, "Product retrieved successfully", productResponse))
}

// CreateProduct tạo sản phẩm mới (Private - Admin only)
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	// Kiểm tra quyền admin
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, utils.NewErrorResponse(http.StatusForbidden, "Permission denied", "Only admin can create products"))
		return
	}

	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid request", err.Error()))
		return
	}

	product := models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		ImageURL:    req.ImageURL,
		Category:    req.Category,
	}

	if err := h.db.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Error creating product", err.Error()))
		return
	}

	productResponse := models.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		ImageURL:    product.ImageURL,
		Category:    product.Category,
		CreatedAt:   product.CreatedAt,
	}

	c.JSON(http.StatusCreated, utils.NewResponse(http.StatusCreated, "Product created successfully", productResponse))
}

// UpdateProduct cập nhật sản phẩm (Private - Admin only)
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	// Kiểm tra quyền admin
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, utils.NewErrorResponse(http.StatusForbidden, "Permission denied", "Only admin can update products"))
		return
	}

	id := c.Param("id")
	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid request", err.Error()))
		return
	}

	var product models.Product
	if err := h.db.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse(http.StatusNotFound, "Product not found", ""))
		return
	}

	// Update fields
	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.Price > 0 {
		product.Price = req.Price
	}
	if req.Stock >= 0 {
		product.Stock = req.Stock
	}
	if req.ImageURL != "" {
		product.ImageURL = req.ImageURL
	}
	if req.Category != "" {
		product.Category = req.Category
	}

	if err := h.db.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Error updating product", err.Error()))
		return
	}

	productResponse := models.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		ImageURL:    product.ImageURL,
		Category:    product.Category,
		CreatedAt:   product.CreatedAt,
	}

	c.JSON(http.StatusOK, utils.NewResponse(http.StatusOK, "Product updated successfully", productResponse))
}

// DeleteProduct xóa sản phẩm (Private - Admin only)
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	// Kiểm tra quyền admin
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, utils.NewErrorResponse(http.StatusForbidden, "Permission denied", "Only admin can delete products"))
		return
	}

	id := c.Param("id")
	if err := h.db.Delete(&models.Product{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Error deleting product", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.NewResponse(http.StatusOK, "Product deleted successfully", nil))
} 