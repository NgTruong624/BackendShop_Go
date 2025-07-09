package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NgTruong624/project_backend/internal/models"
	"github.com/NgTruong624/project_backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// UserRepositoryInterface định nghĩa interface cho user repository
type UserRepositoryInterface interface {
	GetAllUsers(query *models.UserQueryParams) ([]models.User, int64, error)
	GetByID(id uint) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
}

// MockUserRepository implements UserRepositoryInterface for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetAllUsers(query *models.UserQueryParams) ([]models.User, int64, error) {
	args := m.Called(query)
	return args.Get(0).([]models.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) GetByID(id uint) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// TestAdminHandler wrapper để inject mock repository
type TestAdminHandler struct {
	userRepo UserRepositoryInterface
}

func NewTestAdminHandler(userRepo UserRepositoryInterface) *TestAdminHandler {
	return &TestAdminHandler{userRepo: userRepo}
}

// GetUsersList implementation for testing
func (h *TestAdminHandler) GetUsersList(c *gin.Context) {
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

type AdminHandlerTestSuite struct {
	suite.Suite
	handler   *TestAdminHandler
	mockRepo  *MockUserRepository
	router    *gin.Engine
	testUsers []models.User
}

func (suite *AdminHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.mockRepo = new(MockUserRepository)
	suite.handler = NewTestAdminHandler(suite.mockRepo)

	suite.router = gin.New()

	// Admin routes
	adminGroup := suite.router.Group("/admin")
	adminGroup.Use(func(c *gin.Context) {
		// Simulate JWT middleware setting admin user context
		c.Set("role", "admin")
		c.Set("user_id", uint(1))
		c.Next()
	})
	adminGroup.GET("/users", suite.handler.GetUsersList)

	// Non-admin routes
	userGroup := suite.router.Group("/user")
	userGroup.Use(func(c *gin.Context) {
		// Simulate JWT middleware setting regular user context
		c.Set("role", "user")
		c.Set("user_id", uint(2))
		c.Next()
	})
	userGroup.GET("/users", suite.handler.GetUsersList)

	// Test data
	suite.testUsers = []models.User{
		{
			ID:        1,
			Username:  "admin",
			Email:     "admin@example.com",
			FullName:  "Admin User",
			Role:      "admin",
			CreatedAt: time.Now(),
		},
		{
			ID:        2,
			Username:  "user1",
			Email:     "user1@example.com",
			FullName:  "Regular User 1",
			Role:      "user",
			CreatedAt: time.Now(),
		},
		{
			ID:        3,
			Username:  "user2",
			Email:     "user2@example.com",
			FullName:  "Regular User 2",
			Role:      "user",
			CreatedAt: time.Now(),
		},
	}
}

func (suite *AdminHandlerTestSuite) SetupTest() {
	// Reset mock expectations before each test
	suite.mockRepo.ExpectedCalls = nil
	suite.mockRepo.Calls = nil
}

// ===================
// GET USERS LIST TESTS
// ===================

func (suite *AdminHandlerTestSuite) TestGetUsersList_Success() {
	// Mock repository response
	suite.mockRepo.On("GetAllUsers", mock.AnythingOfType("*models.UserQueryParams")).
		Return(suite.testUsers, int64(3), nil)

	req, _ := http.NewRequest("GET", "/admin/users", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(200), response["status"])
	assert.Equal(suite.T(), "Users retrieved successfully", response["message"])

	data := response["data"].([]interface{})
	assert.Len(suite.T(), data, 3)

	// Verify first user
	firstUser := data[0].(map[string]interface{})
	assert.Equal(suite.T(), "admin", firstUser["username"])
	assert.Equal(suite.T(), "admin@example.com", firstUser["email"])
	assert.Equal(suite.T(), "Admin User", firstUser["full_name"])
	assert.Equal(suite.T(), "admin", firstUser["role"])

	// Verify pagination metadata
	meta := response["meta"].(map[string]interface{})
	pagination := meta["pagination"].(map[string]interface{})
	assert.Equal(suite.T(), float64(3), pagination["total_items"])
	assert.Equal(suite.T(), float64(1), pagination["current_page"])
	assert.Equal(suite.T(), float64(1), pagination["total_pages"])
	assert.Equal(suite.T(), false, pagination["has_next"])
	assert.Equal(suite.T(), false, pagination["has_prev"])

	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AdminHandlerTestSuite) TestGetUsersList_WithPagination() {
	// Mock repository response for paginated data
	suite.mockRepo.On("GetAllUsers", mock.MatchedBy(func(query *models.UserQueryParams) bool {
		return query.Page == 2 && query.Limit == 2
	})).Return([]models.User{suite.testUsers[2]}, int64(3), nil)

	req, _ := http.NewRequest("GET", "/admin/users?page=2&limit=2", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	data := response["data"].([]interface{})
	assert.Len(suite.T(), data, 1)

	// Verify pagination metadata
	meta := response["meta"].(map[string]interface{})
	pagination := meta["pagination"].(map[string]interface{})
	assert.Equal(suite.T(), float64(3), pagination["total_items"])
	assert.Equal(suite.T(), float64(2), pagination["current_page"])
	assert.Equal(suite.T(), float64(2), pagination["total_pages"])
	assert.Equal(suite.T(), false, pagination["has_next"])
	assert.Equal(suite.T(), true, pagination["has_prev"])

	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AdminHandlerTestSuite) TestGetUsersList_WithSearch() {
	// Mock repository response for search
	filteredUsers := []models.User{suite.testUsers[1]}

	suite.mockRepo.On("GetAllUsers", mock.MatchedBy(func(query *models.UserQueryParams) bool {
		return query.Search == "user1"
	})).Return(filteredUsers, int64(1), nil)

	req, _ := http.NewRequest("GET", "/admin/users?search=user1", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	data := response["data"].([]interface{})
	assert.Len(suite.T(), data, 1)

	user := data[0].(map[string]interface{})
	assert.Equal(suite.T(), "user1", user["username"])

	// Verify search metadata
	meta := response["meta"].(map[string]interface{})
	assert.Equal(suite.T(), "user1", meta["search"])

	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AdminHandlerTestSuite) TestGetUsersList_WithRoleFilter() {
	// Mock repository response for role filter
	userRoleUsers := []models.User{suite.testUsers[1], suite.testUsers[2]}

	suite.mockRepo.On("GetAllUsers", mock.MatchedBy(func(query *models.UserQueryParams) bool {
		return query.Role == "user"
	})).Return(userRoleUsers, int64(2), nil)

	req, _ := http.NewRequest("GET", "/admin/users?role=user", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	data := response["data"].([]interface{})
	assert.Len(suite.T(), data, 2)

	// Verify both users have role "user"
	for _, userData := range data {
		user := userData.(map[string]interface{})
		assert.Equal(suite.T(), "user", user["role"])
	}

	// Verify role filter metadata
	meta := response["meta"].(map[string]interface{})
	assert.Equal(suite.T(), "user", meta["role"])

	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AdminHandlerTestSuite) TestGetUsersList_Forbidden() {
	// Test with non-admin user
	req, _ := http.NewRequest("GET", "/user/users", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(403), response["status"])
	assert.Equal(suite.T(), "Permission denied", response["message"])
	assert.Equal(suite.T(), "Only admin can access user list", response["error"])

	// No repository calls should be made for forbidden access
	suite.mockRepo.AssertNotCalled(suite.T(), "GetAllUsers")
}

func (suite *AdminHandlerTestSuite) TestGetUsersList_DatabaseError() {
	// Mock repository to return error
	suite.mockRepo.On("GetAllUsers", mock.AnythingOfType("*models.UserQueryParams")).
		Return([]models.User{}, int64(0), assert.AnError)

	req, _ := http.NewRequest("GET", "/admin/users", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(500), response["status"])
	assert.Equal(suite.T(), "Error fetching users", response["message"])

	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AdminHandlerTestSuite) TestGetUsersList_InvalidQueryParams() {
	// Test with invalid query parameters that exceed limits
	req, _ := http.NewRequest("GET", "/admin/users?limit=150", nil) // Exceeds max limit of 100
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Should still work but limit should be capped at 100
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Mock should receive query with limit = 100 (capped)
	suite.mockRepo.On("GetAllUsers", mock.MatchedBy(func(query *models.UserQueryParams) bool {
		return query.Limit == 100
	})).Return([]models.User{}, int64(0), nil)

	// Make the request again to trigger the mock expectation
	w2 := httptest.NewRecorder()
	suite.router.ServeHTTP(w2, req)

	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AdminHandlerTestSuite) TestGetUsersList_DefaultPagination() {
	// Mock repository response with default pagination
	suite.mockRepo.On("GetAllUsers", mock.MatchedBy(func(query *models.UserQueryParams) bool {
		return query.Page == 1 && query.Limit == 10 // Default values
	})).Return(suite.testUsers, int64(3), nil)

	req, _ := http.NewRequest("GET", "/admin/users", nil) // No pagination params
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	meta := response["meta"].(map[string]interface{})
	pagination := meta["pagination"].(map[string]interface{})
	assert.Equal(suite.T(), float64(1), pagination["current_page"])
	assert.Equal(suite.T(), float64(10), pagination["items_per_page"])

	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AdminHandlerTestSuite) TestGetUsersList_EmptyResult() {
	// Mock repository to return empty result
	suite.mockRepo.On("GetAllUsers", mock.AnythingOfType("*models.UserQueryParams")).
		Return([]models.User{}, int64(0), nil)

	req, _ := http.NewRequest("GET", "/admin/users", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Users retrieved successfully", response["message"])

	data := response["data"].([]interface{})
	assert.Len(suite.T(), data, 0)

	meta := response["meta"].(map[string]interface{})
	pagination := meta["pagination"].(map[string]interface{})
	assert.Equal(suite.T(), float64(0), pagination["total_items"])
	assert.Equal(suite.T(), float64(0), pagination["total_pages"])

	suite.mockRepo.AssertExpectations(suite.T())
}

// Run the test suite
func TestAdminHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AdminHandlerTestSuite))
}
