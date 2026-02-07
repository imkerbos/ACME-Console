package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/ACME-Console/internal/auth"
	"github.com/imkerbos/ACME-Console/internal/model"
	"github.com/imkerbos/ACME-Console/internal/response"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db         *gorm.DB
	jwtManager *auth.JWTManager
}

func NewAuthHandler(db *gorm.DB, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		db:         db,
		jwtManager: jwtManager,
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token     string   `json:"token"`
	ExpiresIn int64    `json:"expires_in"`
	User      UserInfo `json:"user"`
}

type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	var user model.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		response.BadRequest(c, "invalid username or password")
		return
	}

	if user.Status != 1 {
		response.BadRequest(c, "user is disabled")
		return
	}

	if !user.CheckPassword(req.Password) {
		response.BadRequest(c, "invalid username or password")
		return
	}

	// Generate token with role
	token, err := h.jwtManager.Generate(user.ID, user.Username, user.Role)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	// Update last login time
	now := time.Now()
	h.db.Model(&user).Update("last_login", now)

	response.Success(c, LoginResponse{
		Token:     token,
		ExpiresIn: 86400, // 24 hours
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Nickname: user.Nickname,
			Email:    user.Email,
			Role:     user.Role,
		},
	})
}

// GetCurrentUser handles GET /api/v1/auth/me
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.NotFound(c, "user not found")
		return
	}

	response.Success(c, UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
		Role:     user.Role,
	})
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type UpdateProfileRequest struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email" binding:"omitempty,email"`
}

// ChangePassword handles POST /api/v1/auth/change-password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.NotFound(c, "user not found")
		return
	}

	if !user.CheckPassword(req.OldPassword) {
		response.BadRequest(c, "old password is incorrect")
		return
	}

	if err := user.SetPassword(req.NewPassword); err != nil {
		response.InternalError(c, err)
		return
	}

	if err := h.db.Save(&user).Error; err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, "password changed successfully")
}

// UpdateProfile handles PUT /api/v1/auth/profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.NotFound(c, "user not found")
		return
	}

	// Update fields if provided
	updates := make(map[string]interface{})
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}

	if len(updates) > 0 {
		if err := h.db.Model(&user).Updates(updates).Error; err != nil {
			response.InternalError(c, err)
			return
		}
	}

	// Reload user to get updated data
	if err := h.db.First(&user, userID).Error; err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
		Role:     user.Role,
	})
}
