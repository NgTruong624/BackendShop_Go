// File: product_handler.go
package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings" // Thêm import strings
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

// --- GetProducts và GetProduct giữ nguyên như file bạn đã cung cấp ---
// GetProducts lấy danh sách sản phẩm (Public)
func (h *ProductHandler) GetProducts(c *gin.Context) {
	var query models.ProductQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid query parameters", err.Error()))
		return
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 10
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			query.StartDate = t
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			t = t.Add(24*time.Hour - time.Second)
			query.EndDate = t
		}
	}
	if inStock := c.Query("in_stock"); inStock != "" {
		query.InStock = inStock == "true"
	}
	if query.MinPrice > 0 && query.MaxPrice > 0 && query.MinPrice > query.MaxPrice {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid price range", "min_price cannot be greater than max_price"))
		return
	}
	if !query.StartDate.IsZero() && !query.EndDate.IsZero() && query.StartDate.After(query.EndDate) {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid date range", "start_date cannot be after end_date"))
		return
	}

	products, total, err := h.repo.GetAll(&query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Error fetching products", err.Error()))
		return
	}

	var productResponses []models.ProductResponse
	for _, p := range products {
		productResponses = append(productResponses, models.ProductResponse{
			ID: p.ID, Name: p.Name, Description: p.Description, Price: p.Price,
			Stock: p.Stock, ImageURL: p.ImageURL, Category: p.Category, CreatedAt: p.CreatedAt,
		})
	}
	totalPages := (int(total) + query.Limit - 1) / query.Limit
	meta := map[string]interface{}{
		"total": total, "total_pages": totalPages, "current_page": query.Page,
		"per_page": query.Limit, "has_next": query.Page < totalPages, "has_prev": query.Page > 1,
	}
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
		http.StatusOK, "Products retrieved successfully", productResponses,
		query.Page, totalPages, total, query.Limit, meta,
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
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.NewErrorResponse(http.StatusNotFound, "Product not found", ""))
			return
		}
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Error fetching product", err.Error()))
		return
	}
	productResponse := models.ProductResponse{
		ID: product.ID, Name: product.Name, Description: product.Description, Price: product.Price,
		Stock: product.Stock, ImageURL: product.ImageURL, Category: product.Category, CreatedAt: product.CreatedAt,
	}
	c.JSON(http.StatusOK, utils.NewResponse(http.StatusOK, "Product retrieved successfully", productResponse))
}

// CreateProduct tạo sản phẩm mới (Private - Admin only)
func (h *ProductHandler) CreateProduct(c *gin.Context) {
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

	// Sửa: Kiểm tra tên sản phẩm đã tồn tại chưa, xử lý lỗi từ repo
	nameExists, errDb := h.repo.CheckIfNameExists(req.Name, 0) // 0 vì đây là tạo mới, không có ID để loại trừ
	if errDb != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Error checking product name availability", errDb.Error()))
		return
	}
	if nameExists {
		c.JSON(http.StatusConflict, utils.NewErrorResponse(http.StatusConflict, "Product name already exists", "")) // Sử dụng 409 Conflict
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
		// Kiểm tra lỗi từ DB nếu có ràng buộc UNIQUE ở DB
		// (PostgreSQL thường trả về lỗi chứa "23505" hoặc "unique_violation" cho duplicate key)
		if strings.Contains(err.Error(), "unique_violation") || strings.Contains(err.Error(), "23505") {
			c.JSON(http.StatusConflict, utils.NewErrorResponse(http.StatusConflict, "Product name already exists (DB constraint)", err.Error()))
			return
		}
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

	product, err := h.repo.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.NewErrorResponse(http.StatusNotFound, "Product not found", ""))
			return
		}
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Error fetching product", err.Error()))
		return
	}

	// Sửa: Kiểm tra tên sản phẩm trùng (nếu cập nhật tên), xử lý lỗi từ repo
	if req.Name != "" && req.Name != product.Name { // Chỉ kiểm tra nếu tên mới khác tên cũ
		nameExists, errDb := h.repo.CheckIfNameExists(req.Name, product.ID)
		if errDb != nil {
			c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Error checking product name availability", errDb.Error()))
			return
		}
		if nameExists {
			c.JSON(http.StatusConflict, utils.NewErrorResponse(http.StatusConflict, "Another product with this name already exists", "")) // Sử dụng 409 Conflict
			return
		}
		product.Name = req.Name
	}

	// Cập nhật các trường khác
	if req.Description != "" {
		product.Description = req.Description
	}
	// Sửa: Price và Stock nên được kiểm tra xem có được gửi lên không, thay vì chỉ > 0 hoặc >= 0
	// Nếu bạn muốn cho phép cập nhật Price/Stock thành 0, logic hiện tại là ổn.
	// Nếu Price/Stock là optional và chỉ cập nhật nếu được gửi, bạn cần dùng con trỏ trong UpdateProductRequest
	// hoặc kiểm tra sự tồn tại của key trong JSON. Giả sử logic hiện tại là ý muốn.
	if req.Price > 0 { // Hoặc bạn có thể dùng con trỏ để phân biệt 0 và không cung cấp
		product.Price = req.Price
	}
	if req.Stock >= 0 { // Tương tự như Price
		product.Stock = req.Stock
	}
	if req.ImageURL != "" {
		product.ImageURL = req.ImageURL
	}
	if req.Category != "" {
		product.Category = req.Category
	}

	if err := h.repo.Update(product); err != nil {
		// Kiểm tra lỗi từ DB nếu có ràng buộc UNIQUE ở DB
		if strings.Contains(err.Error(), "unique_violation") || strings.Contains(err.Error(), "23505") {
			c.JSON(http.StatusConflict, utils.NewErrorResponse(http.StatusConflict, "Product name already exists (DB constraint)", err.Error()))
			return
		}
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
		CreatedAt:   product.CreatedAt, // Nên là UpdatedAt của product
	}
	c.JSON(http.StatusOK, utils.NewResponse(http.StatusOK, "Product updated successfully", productResponse))
}

// --- DeleteProduct và UploadProductImage giữ nguyên như file bạn đã cung cấp ---
// DeleteProduct xóa sản phẩm (Private - Admin only)
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
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

// UploadProductImage xử lý upload ảnh cho sản phẩm
func (h *ProductHandler) UploadProductImage(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, utils.NewErrorResponse(http.StatusForbidden, "Permission denied", "Only admin can upload product images"))
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid product ID", err.Error()))
		return
	}
	product, err := h.repo.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.NewErrorResponse(http.StatusNotFound, "Product not found", ""))
			return
		}
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Error fetching product", err.Error()))
		return
	}
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "No image file provided", err.Error()))
		return
	}
	if !isValidImageType(file.Header.Get("Content-Type")) {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid file type", "Only JPG, PNG and GIF images are allowed"))
		return
	}
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%d%s", product.ID, time.Now().Unix(), ext)
	uploadPath := filepath.Join("static", "uploads", filename) // Sửa tên biến để tránh nhầm lẫn
	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Error saving file", err.Error()))
		return
	}
	product.ImageURL = "/" + uploadPath // Sử dụng uploadPath đã join
	if err := h.repo.Update(product); err != nil {
		os.Remove(uploadPath)
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Error updating product image URL", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.NewResponse(http.StatusOK, "Image uploaded successfully", gin.H{"image_url": product.ImageURL}))
}

func isValidImageType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true, "image/png": true, "image/gif": true,
	}
	return validTypes[contentType]
}
