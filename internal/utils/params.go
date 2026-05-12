package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ParseIDParam parses an ID parameter from the URL path
func ParseIDParam(c *gin.Context, param string) (uint, error) {
	idStr := c.Param(param)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// ParseID is a shorthand for ParseIDParam with "id" as the parameter name
func ParseID(c *gin.Context) (uint, error) {
	return ParseIDParam(c, "id")
}

// ParseQueryInt parses an integer query parameter with a default value
func ParseQueryInt(c *gin.Context, key string, defaultVal int) int {
	valStr := c.Query(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

// ParseQueryBool parses a boolean query parameter with a default value
func ParseQueryBool(c *gin.Context, key string, defaultVal bool) bool {
	valStr := c.Query(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

// GetUserID retrieves the user ID from the gin context (set by JWT middleware)
func GetUserID(c *gin.Context) uint {
	userID, exists := c.Get("userID")
	if !exists {
		return 0
	}
	return userID.(uint)
}

// ParseQueryUint parses a uint query parameter
func ParseQueryUint(c *gin.Context, key string) uint {
	valStr := c.Query(key)
	if valStr == "" {
		return 0
	}
	val, err := strconv.ParseUint(valStr, 10, 32)
	if err != nil {
		return 0
	}
	return uint(val)
}
