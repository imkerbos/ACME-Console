package utils

// PaginatedResult represents a paginated response
type PaginatedResult[T any] struct {
	Items      []T `json:"items"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalPages int `json:"total_pages"`
}

// NewPagination creates a new paginated result
func NewPagination[T any](items []T, total, page, pageSize int) PaginatedResult[T] {
	totalPages := total / pageSize
	if total%pageSize > 0 {
		totalPages++
	}

	return PaginatedResult[T]{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
