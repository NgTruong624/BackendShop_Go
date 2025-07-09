package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/NgTruong624/project_backend/internal/models"
	"github.com/NgTruong624/project_backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// ProductRepositoryInterface định nghĩa interface cho repository
type ProductRepositoryInterface interface {
	GetAll(query *models.ProductQueryParams) ([]models.Product, int64, error)
	GetByID(id uint) (*models.Product, error)
	Create(product *models.Product) error
	Update(product *models.Product) error
	Delete(id uint) error
	CheckIfNameExists(name string, excludeID uint) (bool, error)
}

// MockProductRepository implements ProductRepositoryInterface for testing
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) GetAll(query *models.ProductQueryParams) ([]models.Product, int64, error) {
	args := m.Called(query)
	return args.Get(0).([]models.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) GetByID(id uint) (*models.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) Create(product *models.Product) error {
	args := m.Called(product)
	// Simulate ID assignment after creation
	if args.Error(0) == nil {
		product.ID = 1
		product.CreatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *MockProductRepository) Update(product *models.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProductRepository) CheckIfNameExists(name string, excludeID uint) (bool, error) {
	args := m.Called(name, excludeID)
	return args.Bool(0), args.Error(1)
}

// TestProductHandler wrapper để inject mock repository
type TestProductHandler struct {
	repo ProductRepositoryInterface
}

func NewTestProductHandler(repo ProductRepositoryInterface) *TestProductHandler {
	return &TestProductHandler{repo: repo}
}

// Copy all methods from ProductHandler but use interface
func (h *TestProductHandler) GetProducts(c *gin.Context) {
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

	c.JSON(http.StatusOK, utils.NewPaginatedResponse(
		http.StatusOK, "Products retrieved successfully", productResponses,
		query.Page, totalPages, total, query.Limit, meta,
	))
}

func (h *TestProductHandler) GetProduct(c *gin.Context) {
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

func (h *TestProductHandler) CreateProduct(c *gin.Context) {
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

	nameExists, errDb := h.repo.CheckIfNameExists(req.Name, 0)
	if errDb != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Error checking product name availability", errDb.Error()))
		return
	}
	if nameExists {
		c.JSON(http.StatusConflict, utils.NewErrorResponse(http.StatusConflict, "Product name already exists", ""))
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

func (h *TestProductHandler) UpdateProduct(c *gin.Context) {
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

	if req.Name != "" && req.Name != product.Name {
		nameExists, errDb := h.repo.CheckIfNameExists(req.Name, product.ID)
		if errDb != nil {
			c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Error checking product name availability", errDb.Error()))
			return
		}
		if nameExists {
			c.JSON(http.StatusConflict, utils.NewErrorResponse(http.StatusConflict, "Another product with this name already exists", ""))
			return
		}
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

func (h *TestProductHandler) DeleteProduct(c *gin.Context) {
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

func (h *TestProductHandler) UploadProductImage(c *gin.Context) {
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
	uploadPath := filepath.Join("static", "uploads", filename)
	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Error saving file", err.Error()))
		return
	}
	product.ImageURL = "/" + uploadPath
	if err := h.repo.Update(product); err != nil {
		os.Remove(uploadPath)
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Error updating product image URL", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.NewResponse(http.StatusOK, "Image uploaded successfully", gin.H{"image_url": product.ImageURL}))
}

type ProductHandlerTestSuite struct {
	suite.Suite
	handler     *TestProductHandler
	mockRepo    *MockProductRepository
	router      *gin.Engine
	testProduct *models.Product
}

func (suite *ProductHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.mockRepo = new(MockProductRepository)
	suite.handler = NewTestProductHandler(suite.mockRepo)

	suite.router = gin.New()

	// Public routes
	suite.router.GET("/products", suite.handler.GetProducts)
	suite.router.GET("/products/:id", suite.handler.GetProduct)

	// Protected routes (simulate middleware)
	protected := suite.router.Group("/admin")
	protected.Use(func(c *gin.Context) {
		// Simulate JWT middleware setting user context
		c.Set("role", "admin")
		c.Set("user_id", uint(1))
		c.Next()
	})
	protected.POST("/products", suite.handler.CreateProduct)
	protected.PUT("/products/:id", suite.handler.UpdateProduct)
	protected.DELETE("/products/:id", suite.handler.DeleteProduct)
	protected.POST("/products/:id/upload", suite.handler.UploadProductImage)

	// Route for non-admin user
	nonAdmin := suite.router.Group("/user")
	nonAdmin.Use(func(c *gin.Context) {
		c.Set("role", "user")
		c.Set("user_id", uint(2))
		c.Next()
	})
	nonAdmin.POST("/products", suite.handler.CreateProduct)
	nonAdmin.PUT("/products/:id", suite.handler.UpdateProduct)
	nonAdmin.DELETE("/products/:id", suite.handler.DeleteProduct)

	// Test data
	suite.testProduct = &models.Product{
		ID:          1,
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		Stock:       10,
		ImageURL:    "/uploads/test.jpg",
		Category:    "Electronics",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (suite *ProductHandlerTestSuite) SetupTest() {
	// Reset mock expectations before each test
	suite.mockRepo.ExpectedCalls = nil
	suite.mockRepo.Calls = nil
}

// ===================
// GET PRODUCTS TESTS
// ===================

func (suite *ProductHandlerTestSuite) TestGetProducts_Success() {
	// Mock data
	products := []models.Product{*suite.testProduct}

	suite.mockRepo.On("GetAll", mock.AnythingOfType("*models.ProductQueryParams")).
		Return(products, int64(1), nil)

	req, _ := http.NewRequest("GET", "/products", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(200), response["status"])
	assert.Equal(suite.T(), "Products retrieved successfully", response["message"])

	data := response["data"].([]interface{})
	assert.Len(suite.T(), data, 1)

	product := data[0].(map[string]interface{})
	assert.Equal(suite.T(), "Test Product", product["name"])
	assert.Equal(suite.T(), 99.99, product["price"])

	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *ProductHandlerTestSuite) TestGetProducts_WithQueryParams() {
	products := []models.Product{*suite.testProduct}

	suite.mockRepo.On("GetAll", mock.MatchedBy(func(query *models.ProductQueryParams) bool {
		return query.Search == "test" && query.Category == "Electronics" && query.Page == 2 && query.Limit == 5
	})).Return(products, int64(1), nil)

	req, _ := http.NewRequest("GET", "/products?search=test&category=Electronics&page=2&limit=5", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *ProductHandlerTestSuite) TestGetProducts_InvalidPriceRange() {
	req, _ := http.NewRequest("GET", "/products?min_price=100&max_price=50", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid price range", response["message"])
}

func (suite *ProductHandlerTestSuite) TestGetProducts_DatabaseError() {
	suite.mockRepo.On("GetAll", mock.AnythingOfType("*models.ProductQueryParams")).
		Return([]models.Product{}, int64(0), assert.AnError)

	req, _ := http.NewRequest("GET", "/products", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	suite.mockRepo.AssertExpectations(suite.T())
}

// ===================
// GET PRODUCT TESTS
// ===================

func (suite *ProductHandlerTestSuite) TestGetProduct_Success() {
	suite.mockRepo.On("GetByID", uint(1)).Return(suite.testProduct, nil)

	req, _ := http.NewRequest("GET", "/products/1", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Product retrieved successfully", response["message"])

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "Test Product", data["name"])

	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *ProductHandlerTestSuite) TestGetProduct_NotFound() {
	suite.mockRepo.On("GetByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

	req, _ := http.NewRequest("GET", "/products/999", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Product not found", response["message"])

	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *ProductHandlerTestSuite) TestGetProduct_InvalidID() {
	req, _ := http.NewRequest("GET", "/products/invalid", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// ===================
// CREATE PRODUCT TESTS
// ===================

func (suite *ProductHandlerTestSuite) TestCreateProduct_Success() {
	createReq := models.CreateProductRequest{
		Name:        "New Product",
		Description: "New Description",
		Price:       149.99,
		Stock:       20,
		ImageURL:    "/uploads/new.jpg",
		Category:    "Books",
	}

	suite.mockRepo.On("CheckIfNameExists", "New Product", uint(0)).Return(false, nil)
	suite.mockRepo.On("Create", mock.AnythingOfType("*models.Product")).Return(nil)

	jsonData, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/admin/products", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Product created successfully", response["message"])

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "New Product", data["name"])

	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *ProductHandlerTestSuite) TestCreateProduct_NameAlreadyExists() {
	createReq := models.CreateProductRequest{
		Name:        "Existing Product",
		Description: "Description",
		Price:       99.99,
		Stock:       10,
		Category:    "Electronics",
	}

	suite.mockRepo.On("CheckIfNameExists", "Existing Product", uint(0)).Return(true, nil)

	jsonData, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/admin/products", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusConflict, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Product name already exists", response["message"])

	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *ProductHandlerTestSuite) TestCreateProduct_Forbidden() {
	createReq := models.CreateProductRequest{
		Name:        "New Product",
		Description: "Description",
		Price:       99.99,
		Stock:       10,
		Category:    "Electronics",
	}

	jsonData, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/user/products", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Permission denied", response["message"])
}

func (suite *ProductHandlerTestSuite) TestCreateProduct_InvalidRequest() {
	invalidReq := map[string]interface{}{
		"name":  "",  // Empty name should fail validation
		"price": -10, // Negative price should fail
	}

	jsonData, _ := json.Marshal(invalidReq)
	req, _ := http.NewRequest("POST", "/admin/products", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// ===================
// UPDATE PRODUCT TESTS
// ===================

func (suite *ProductHandlerTestSuite) TestUpdateProduct_Success() {
	updateReq := models.UpdateProductRequest{
		Name:        "Updated Product",
		Description: "Updated Description",
		Price:       199.99,
		Stock:       30,
	}

	suite.mockRepo.On("GetByID", uint(1)).Return(suite.testProduct, nil)
	suite.mockRepo.On("CheckIfNameExists", "Updated Product", uint(1)).Return(false, nil)
	suite.mockRepo.On("Update", mock.AnythingOfType("*models.Product")).Return(nil)

	jsonData, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/admin/products/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Product updated successfully", response["message"])

	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *ProductHandlerTestSuite) TestUpdateProduct_NotFound() {
	updateReq := models.UpdateProductRequest{
		Name: "Updated Product",
	}

	suite.mockRepo.On("GetByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

	jsonData, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/admin/products/999", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	suite.mockRepo.AssertExpectations(suite.T())
}

// ===================
// DELETE PRODUCT TESTS
// ===================

func (suite *ProductHandlerTestSuite) TestDeleteProduct_Success() {
	suite.mockRepo.On("Delete", uint(1)).Return(nil)

	req, _ := http.NewRequest("DELETE", "/admin/products/1", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Product deleted successfully", response["message"])

	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *ProductHandlerTestSuite) TestDeleteProduct_DatabaseError() {
	suite.mockRepo.On("Delete", uint(1)).Return(assert.AnError)

	req, _ := http.NewRequest("DELETE", "/admin/products/1", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	suite.mockRepo.AssertExpectations(suite.T())
}

// ===================
// FILE UPLOAD TESTS
// ===================

func (suite *ProductHandlerTestSuite) TestUploadProductImage_Success() {
	// Create uploads directory if it doesn't exist
	os.MkdirAll("static/uploads", 0755)
	defer os.RemoveAll("static") // Clean up after test

	suite.mockRepo.On("GetByID", uint(1)).Return(suite.testProduct, nil)
	suite.mockRepo.On("Update", mock.AnythingOfType("*models.Product")).Return(nil)

	// Create a fake image file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("image", "test.jpg")
	assert.NoError(suite.T(), err)

	// Write fake image data
	part.Write([]byte("fake image data"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/admin/products/1/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Image uploaded successfully", response["message"])

	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *ProductHandlerTestSuite) TestUploadProductImage_NoFile() {
	req, _ := http.NewRequest("POST", "/admin/products/1/upload", nil)
	req.Header.Set("Content-Type", "multipart/form-data")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "No image file provided", response["message"])
}

func (suite *ProductHandlerTestSuite) TestUploadProductImage_ProductNotFound() {
	suite.mockRepo.On("GetByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

	// Create a fake image file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("image", "test.jpg")
	assert.NoError(suite.T(), err)
	part.Write([]byte("fake image data"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/admin/products/999/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	suite.mockRepo.AssertExpectations(suite.T())
}

// Run the test suite
func TestProductHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ProductHandlerTestSuite))
}
