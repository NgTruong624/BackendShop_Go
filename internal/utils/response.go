package utils

// Response là cấu trúc response chung cho tất cả API
type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// PaginatedResponse là cấu trúc response cho các API có phân trang
type PaginatedResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Meta    Meta        `json:"meta"`
}

// Meta chứa thông tin phân trang và metadata
type Meta struct {
	Pagination Pagination             `json:"pagination"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
}

// Pagination chứa thông tin phân trang
type Pagination struct {
	CurrentPage  int   `json:"current_page"`
	TotalPages   int   `json:"total_pages"`
	TotalItems   int64 `json:"total_items"`
	ItemsPerPage int   `json:"items_per_page"`
	HasNext      bool  `json:"has_next"`
	HasPrev      bool  `json:"has_prev"`
}

// NewResponse tạo một response mới
func NewResponse(status int, message string, data interface{}) Response {
	return Response{
		Status:  status,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse tạo một response lỗi
func NewErrorResponse(status int, message string, err interface{}) Response {
	return Response{
		Status:  status,
		Message: message,
		Error:   err,
	}
}

// NewPaginatedResponse tạo một response có phân trang
func NewPaginatedResponse(status int, message string, data interface{}, currentPage, totalPages int, totalItems int64, itemsPerPage int, filters map[string]interface{}) PaginatedResponse {
	return PaginatedResponse{
		Status:  status,
		Message: message,
		Data:    data,
		Meta: Meta{
			Pagination: Pagination{
				CurrentPage:  currentPage,
				TotalPages:   totalPages,
				TotalItems:   totalItems,
				ItemsPerPage: itemsPerPage,
				HasNext:      currentPage < totalPages,
				HasPrev:      currentPage > 1,
			},
			Filters: filters,
		},
	}
}
