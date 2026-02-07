package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Standard response codes
const (
	CodeSuccess       = 0
	CodeBadRequest    = 40000
	CodeUnauthorized  = 40100
	CodeForbidden     = 40300
	CodeNotFound      = 40400
	CodeValidation    = 42200
	CodeInternalError = 50000
)

// Response is the standard API response structure
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// PageData is the standard paginated response data
type PageData struct {
	Items      any   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

// Success returns a successful response with data
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// Created returns a successful creation response
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Response{
		Code:    CodeSuccess,
		Message: "created",
		Data:    data,
	})
}

// OK returns a simple success message
func OK(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
	})
}

// BadRequest returns a 400 error response
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    CodeBadRequest,
		Message: message,
	})
}

// NotFound returns a 404 error response
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Code:    CodeNotFound,
		Message: message,
	})
}

// Unauthorized returns a 401 error response
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    CodeUnauthorized,
		Message: message,
	})
}

// Forbidden returns a 403 error response
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    CodeForbidden,
		Message: message,
	})
}

// InternalError returns a 500 error response
func InternalError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    CodeInternalError,
		Message: err.Error(),
	})
}

// ValidationError returns a 422 validation error response
func ValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusUnprocessableEntity, Response{
		Code:    CodeValidation,
		Message: err.Error(),
	})
}

// Error returns a custom error response
func Error(c *gin.Context, httpCode int, code int, message string) {
	c.JSON(httpCode, Response{
		Code:    code,
		Message: message,
	})
}
