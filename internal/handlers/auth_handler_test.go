package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/NgTruong624/project_backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type AuthHandlerTestSuite struct {
	suite.Suite
	db      *gorm.DB
	sqlDB   *sql.DB
	mock    sqlmock.Sqlmock
	handler *AuthHandler
	router  *gin.Engine
}

// SetupSuite được gọi một lần trước khi chạy tất cả tests
func (suite *AuthHandlerTestSuite) SetupSuite() {
	// Tạo mock database
	var err error
	suite.sqlDB, suite.mock, err = sqlmock.New()
	assert.NoError(suite.T(), err)

	// Tạo GORM DB instance với mock
	suite.db, err = gorm.Open(postgres.New(postgres.Config{
		Conn: suite.sqlDB,
	}), &gorm.Config{})
	assert.NoError(suite.T(), err)

	// Tạo auth handler với JWT secret cho testing
	suite.handler = NewAuthHandler(suite.db, "test-secret-key")

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	suite.router.POST("/login", suite.handler.Login)
	suite.router.POST("/register", suite.handler.Register)
	suite.router.POST("/change-password", suite.handler.ChangePassword)
}

// TearDownSuite được gọi sau khi tất cả tests hoàn thành
func (suite *AuthHandlerTestSuite) TearDownSuite() {
	suite.sqlDB.Close()
}

// SetupTest được gọi trước mỗi test
func (suite *AuthHandlerTestSuite) SetupTest() {
	// Reset mock expectations - không cần làm gì đặc biệt
}

// TestLogin_Success kiểm tra đăng nhập thành công
func (suite *AuthHandlerTestSuite) TestLogin_Success() {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	assert.NoError(suite.T(), err)

	// Mock database query để tìm user
	suite.mock.ExpectQuery(`SELECT \* FROM "users" WHERE username = \$1 ORDER BY "users"\."id" LIMIT \$2`).
		WithArgs("testuser", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password", "full_name", "role", "created_at", "updated_at"}).
			AddRow(1, "testuser", "test@example.com", string(hashedPassword), "Test User", "user", time.Now(), time.Now()))

	loginRequest := models.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	jsonData, err := json.Marshal(loginRequest)
	assert.NoError(suite.T(), err)
	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	assert.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(200), response["status"])
	assert.Equal(suite.T(), "Login successful", response["message"])
	assert.NotNil(suite.T(), response["data"])

	// Kiểm tra có token trong response
	data := response["data"].(map[string]interface{})
	assert.NotEmpty(suite.T(), data["token"])
	assert.NotNil(suite.T(), data["user"])

	// Kiểm tra thông tin user
	user := data["user"].(map[string]interface{})
	assert.Equal(suite.T(), "testuser", user["username"])
	assert.Equal(suite.T(), "test@example.com", user["email"])
	assert.Equal(suite.T(), "Test User", user["full_name"])
	assert.Equal(suite.T(), "user", user["role"])

	// Verify all expectations were met
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// TestLogin_InvalidUsername kiểm tra đăng nhập với username không tồn tại
func (suite *AuthHandlerTestSuite) TestLogin_InvalidUsername() {
	// Mock database query trả về record not found
	suite.mock.ExpectQuery(`SELECT \* FROM "users" WHERE username = \$1 ORDER BY "users"\."id" LIMIT \$2`).
		WithArgs("nonexistentuser", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	loginRequest := models.LoginRequest{
		Username: "nonexistentuser",
		Password: "password123",
	}

	jsonData, err := json.Marshal(loginRequest)
	assert.NoError(suite.T(), err)
	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	assert.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(401), response["status"])
	assert.Equal(suite.T(), "Invalid username or password", response["message"])
	assert.Nil(suite.T(), response["data"])

	// Verify all expectations were met
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// TestLogin_InvalidPassword kiểm tra đăng nhập với password sai
func (suite *AuthHandlerTestSuite) TestLogin_InvalidPassword() {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	assert.NoError(suite.T(), err)

	// Mock database query để tìm user
	suite.mock.ExpectQuery(`SELECT \* FROM "users" WHERE username = \$1 ORDER BY "users"\."id" LIMIT \$2`).
		WithArgs("testuser", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password", "full_name", "role", "created_at", "updated_at"}).
			AddRow(1, "testuser", "test@example.com", string(hashedPassword), "Test User", "user", time.Now(), time.Now()))

	loginRequest := models.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	jsonData, err := json.Marshal(loginRequest)
	assert.NoError(suite.T(), err)
	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	assert.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(401), response["status"])
	assert.Equal(suite.T(), "Invalid username or password", response["message"])
	assert.Nil(suite.T(), response["data"])

	// Verify all expectations were met
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// TestLogin_InvalidJSON kiểm tra đăng nhập với JSON không hợp lệ
func (suite *AuthHandlerTestSuite) TestLogin_InvalidJSON() {
	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer([]byte("invalid json")))
	assert.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(400), response["status"])
	assert.Equal(suite.T(), "Invalid request", response["message"])
}

// TestLogin_MissingFields kiểm tra đăng nhập với các field bắt buộc bị thiếu
func (suite *AuthHandlerTestSuite) TestLogin_MissingFields() {
	loginRequest := models.LoginRequest{
		Username: "", // Missing username
		Password: "password123",
	}

	jsonData, err := json.Marshal(loginRequest)
	assert.NoError(suite.T(), err)
	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	assert.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(400), response["status"])
	assert.Equal(suite.T(), "Invalid request", response["message"])
}

// TestRegister_Success kiểm tra đăng ký thành công
func (suite *AuthHandlerTestSuite) TestRegister_Success() {
	// Mock query để check email đã tồn tại (trả về not found)
	suite.mock.ExpectQuery(`SELECT \* FROM "users" WHERE email = \$1 ORDER BY "users"\."id" LIMIT \$2`).
		WithArgs("newuser@example.com", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// Mock insert operation
	suite.mock.ExpectBegin()
	suite.mock.ExpectQuery(`INSERT INTO "users"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	suite.mock.ExpectCommit()

	registerRequest := models.RegisterRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "password123",
		FullName: "New User",
	}

	jsonData, err := json.Marshal(registerRequest)
	assert.NoError(suite.T(), err)
	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
	assert.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(201), response["status"])
	assert.Equal(suite.T(), "User registered successfully", response["message"])
	assert.NotNil(suite.T(), response["data"])

	// Kiểm tra thông tin user trong response
	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "newuser", data["username"])
	assert.Equal(suite.T(), "newuser@example.com", data["email"])
	assert.Equal(suite.T(), "New User", data["full_name"])
	assert.Equal(suite.T(), "user", data["role"])

	// Verify all expectations were met
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// TestRegister_EmailAlreadyExists kiểm tra đăng ký với email đã tồn tại
func (suite *AuthHandlerTestSuite) TestRegister_EmailAlreadyExists() {
	// Mock query để check email đã tồn tại (trả về user)
	suite.mock.ExpectQuery(`SELECT \* FROM "users" WHERE email = \$1 ORDER BY "users"\."id" LIMIT \$2`).
		WithArgs("test@example.com", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password", "full_name", "role", "created_at", "updated_at"}).
			AddRow(1, "existinguser", "test@example.com", "hashedpass", "Existing User", "user", time.Now(), time.Now()))

	registerRequest := models.RegisterRequest{
		Username: "newuser",
		Email:    "test@example.com", // Email đã tồn tại
		Password: "password123",
		FullName: "New User",
	}

	jsonData, err := json.Marshal(registerRequest)
	assert.NoError(suite.T(), err)
	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
	assert.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(400), response["status"])
	assert.Equal(suite.T(), "Email already exists", response["message"])

	// Verify all expectations were met
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// TestRegister_InvalidEmail kiểm tra đăng ký với email không hợp lệ
func (suite *AuthHandlerTestSuite) TestRegister_InvalidEmail() {
	registerRequest := models.RegisterRequest{
		Username: "newuser",
		Email:    "invalid-email", // Email không hợp lệ
		Password: "password123",
		FullName: "New User",
	}

	jsonData, err := json.Marshal(registerRequest)
	assert.NoError(suite.T(), err)
	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
	assert.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(400), response["status"])
	assert.Equal(suite.T(), "Invalid request", response["message"])
}

// TestRegister_PasswordTooShort kiểm tra đăng ký với password quá ngắn
func (suite *AuthHandlerTestSuite) TestRegister_PasswordTooShort() {
	registerRequest := models.RegisterRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "123", // Password quá ngắn
		FullName: "New User",
	}

	jsonData, err := json.Marshal(registerRequest)
	assert.NoError(suite.T(), err)
	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
	assert.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(400), response["status"])
	assert.Equal(suite.T(), "Invalid request", response["message"])
}

// Chạy test suite
func TestAuthHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}
