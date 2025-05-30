package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/NgTruong624/project_backend/internal/models"
	"github.com/NgTruong624/project_backend/internal/repository"
	"github.com/NgTruong624/project_backend/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductHandler struct {
	repo *repository.ProductRepository
}

func NewProductHandler(db *gorm.DB) *ProductHandler {
	return &ProductHandler{
		repo: repository.NewProductRepository(db),
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
	if query.Limit > 100 {
		query.Limit = 100
	}

	// Parse date parameters if provided
	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			query.StartDate = t
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			// Set end date to end of day
			t = t.Add(24*time.Hour - time.Second)
			query.EndDate = t
		}
	}

	// Parse boolean parameter
	if inStock := c.Query("in_stock"); inStock != "" {
		query.InStock = inStock == "true"
	}

	// Validate price range
	if query.MinPrice > 0 && query.MaxPrice > 0 && query.MinPrice > query.MaxPrice {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid price range", "min_price cannot be greater than max_price"))
		return
	}

	// Validate date range
	if !query.StartDate.IsZero() && !query.EndDate.IsZero() && query.StartDate.After(query.EndDate) {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid date range", "start_date cannot be after end_date"))
		return
	}

	// Get products from repository
	products, total, err := h.repo.GetAll(&query)
	if err != nil {
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
	if query.Category != "" {
		meta["category"] = query.Category
	}
	if query.MinPrice > 0 {
		meta["min_price"] = query.MinPrice
	}
	if query.MaxPrice > 0 {
		meta["max_price"] = query.MaxPrice
	}
	if query.InStock {
		meta["in_stock"] = true
	}
	if !query.StartDate.IsZero() {
		meta["start_date"] = query.StartDate.Format("2006-01-02")
	}
	if !query.EndDate.IsZero() {
		meta["end_date"] = query.EndDate.Format("2006-01-02")
	}
	if query.SortBy != "" {
		meta["sort_by"] = query.SortBy
		meta["order"] = query.Order
	}

	c.JSON(http.StatusOK, utils.NewPaginatedResponse(
		http.StatusOK,
		"Products retrieved successfully",
		productResponses,
		query.Page,
		totalPages,
		total,
		query.Limit,
		meta,
	))
}

// GetProduct lấy chi tiết sản phẩm (Public)
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid product ID", err.Error()))
		return
	}

	product, err := h.repo.GetByID(uint(id))
	if err != nil {
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

	product := &models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		ImageURL:    req.ImageURL,
		Category:    req.Category,
	}

	if err := h.repo.Create(product); err != nil {
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

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid product ID", err.Error()))
		return
	}

	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid request", err.Error()))
		return
	}

	// Lấy sản phẩm hiện tại
	product, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse(http.StatusNotFound, "Product not found", ""))
		return
	}

	// Cập nhật các trường
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

	if err := h.repo.Update(product); err != nil {
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

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid product ID", err.Error()))
		return
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Error deleting product", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.NewResponse(http.StatusOK, "Product deleted successfully", nil))
}
