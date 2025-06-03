package repository

import (
	"github.com/NgTruong624/project_backend/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetAllUsers lấy danh sách người dùng với phân trang và tìm kiếm
func (r *UserRepository) GetAllUsers(query *models.UserQueryParams) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	dbQuery := r.db.Model(&models.User{})

	// Apply search filters (optional)
	if query.Search != "" {
		dbQuery = dbQuery.Where(
			"username ILIKE ? OR email ILIKE ? OR full_name ILIKE ?",
			"%"+query.Search+"%",
			"%"+query.Search+"%",
			"%"+query.Search+"%",
		)
	}

	// Apply role filter
	if query.Role != "" {
		dbQuery = dbQuery.Where("role = ?", query.Role)
	}

	// Get total count before pagination
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply default sorting by created_at desc
	dbQuery = dbQuery.Order("created_at DESC")

	// Apply pagination
	offset := (query.Page - 1) * query.Limit
	if err := dbQuery.Offset(offset).Limit(query.Limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetByID lấy user theo ID (có thể cần cho các chức năng khác)
func (r *UserRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername lấy user theo username (có thể cần cho các chức năng khác)
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail lấy user theo email (có thể cần cho các chức năng khác)
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
