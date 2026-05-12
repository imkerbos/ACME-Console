package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/ACME-Console/internal/auth"
	"github.com/imkerbos/ACME-Console/internal/model"
	"github.com/imkerbos/ACME-Console/internal/response"
)

func JWTAuth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		claims, err := jwtManager.Verify(parts[1])
		if err != nil {
			if err == auth.ErrExpiredToken {
				response.Unauthorized(c, "token expired")
			} else {
				response.Unauthorized(c, "invalid token")
			}
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// AdminAuth checks if the user has admin role
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}

		if role.(string) != model.RoleAdmin {
			response.Forbidden(c, "admin access required")
			c.Abort()
			return
		}

		c.Next()
	}
}
