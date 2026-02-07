package pagination

import (
	"github.com/gin-gonic/gin"
	"github.com/imkerbos/ACME-Console/internal/utils"
	"gorm.io/gorm"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Params represents pagination parameters
type Params struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// Result represents a paginated result
type Result[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

// ParseFromContext extracts pagination parameters from gin context
func ParseFromContext(c *gin.Context) Params {
	page := utils.ParseQueryInt(c, "page", DefaultPage)
	pageSize := utils.ParseQueryInt(c, "page_size", DefaultPageSize)

	if page < 1 {
		page = DefaultPage
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return Params{
		Page:     page,
		PageSize: pageSize,
	}
}

// Offset returns the offset for database queries
func (p Params) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Limit returns the limit for database queries
func (p Params) Limit() int {
	return p.PageSize
}

// Apply applies pagination to a GORM query
func (p Params) Apply(db *gorm.DB) *gorm.DB {
	return db.Offset(p.Offset()).Limit(p.Limit())
}

// NewResult creates a new paginated result
func NewResult[T any](items []T, total int64, params Params) Result[T] {
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	return Result[T]{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}
}

// Paginate is a helper function that handles pagination for a GORM model
func Paginate[T any](db *gorm.DB, params Params, dest *[]T) (*Result[T], error) {
	var total int64

	// Count total records
	if err := db.Model(new(T)).Count(&total).Error; err != nil {
		return nil, err
	}

	// Get paginated records
	if err := params.Apply(db).Find(dest).Error; err != nil {
		return nil, err
	}

	result := NewResult(*dest, total, params)
	return &result, nil
}
