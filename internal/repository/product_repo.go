package repository

import (
	"github.com/NgTruong624/project_backend/internal/models"
	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create tạo sản phẩm mới
func (r *ProductRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

// GetByID lấy sản phẩm theo ID
func (r *ProductRepository) GetByID(id uint) (*models.Product, error) {
	var product models.Product
	err := r.db.First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// GetAll lấy danh sách sản phẩm với các tùy chọn
func (r *ProductRepository) GetAll(query *models.ProductQueryParams) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	// Build query
	dbQuery := r.db.Model(&models.Product{})

	// Apply search filters
	if query.Search != "" {
		dbQuery = dbQuery.Where(
			"name ILIKE ? OR description ILIKE ? OR category ILIKE ?",
			"%"+query.Search+"%",
			"%"+query.Search+"%",
			"%"+query.Search+"%",
		)
	}

	// Apply category filter
	if query.Category != "" {
		dbQuery = dbQuery.Where("category = ?", query.Category)
	}

	// Apply price range filter
	if query.MinPrice > 0 {
		dbQuery = dbQuery.Where("price >= ?", query.MinPrice)
	}
	if query.MaxPrice > 0 {
		dbQuery = dbQuery.Where("price <= ?", query.MaxPrice)
	}

	// Apply stock filter
	if query.InStock {
		dbQuery = dbQuery.Where("stock > 0")
	}

	// Apply date range filter
	if !query.StartDate.IsZero() {
		dbQuery = dbQuery.Where("created_at >= ?", query.StartDate)
	}
	if !query.EndDate.IsZero() {
		dbQuery = dbQuery.Where("created_at <= ?", query.EndDate)
	}

	// Get total count before pagination
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	if query.SortBy != "" {
		// Validate sort field to prevent SQL injection
		validSortFields := map[string]string{
			"name":       "name",
			"price":      "price",
			"stock":      "stock",
			"created_at": "created_at",
			"category":   "category",
		}

		if sortField, ok := validSortFields[query.SortBy]; ok {
			order := "ASC"
			if query.Order == "desc" {
				order = "DESC"
			}
			dbQuery = dbQuery.Order(sortField + " " + order)
		}
	} else {
		// Default sorting by created_at desc
		dbQuery = dbQuery.Order("created_at DESC")
	}

	// Apply pagination
	offset := (query.Page - 1) * query.Limit
	if err := dbQuery.Offset(offset).Limit(query.Limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// Update cập nhật sản phẩm
func (r *ProductRepository) Update(product *models.Product) error {
	return r.db.Save(product).Error
}

// Delete xóa sản phẩm
func (r *ProductRepository) Delete(id uint) error {
	return r.db.Delete(&models.Product{}, id).Error
}

// GetByCategory lấy sản phẩm theo danh mục
func (r *ProductRepository) GetByCategory(category string) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Where("category = ?", category).Find(&products).Error
	return products, err
}

// SearchByName tìm kiếm sản phẩm theo tên
func (r *ProductRepository) SearchByName(name string) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Where("name ILIKE ?", "%"+name+"%").Find(&products).Error
	return products, err
}

// UpdateStock cập nhật số lượng tồn kho
func (r *ProductRepository) UpdateStock(id uint, stock int) error {
	return r.db.Model(&models.Product{}).Where("id = ?", id).Update("stock", stock).Error
}

// GetLowStock lấy danh sách sản phẩm có số lượng tồn kho thấp
func (r *ProductRepository) GetLowStock(threshold int) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Where("stock <= ?", threshold).Find(&products).Error
	return products, err
}
