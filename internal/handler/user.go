package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/ACME-Console/internal/model"
	"github.com/imkerbos/ACME-Console/internal/response"
	"github.com/imkerbos/ACME-Console/internal/utils"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

type UserListItem struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Status    int    `json:"status"`
	LastLogin string `json:"last_login,omitempty"`
	CreatedAt string `json:"created_at"`
}

// List handles GET /api/v1/admin/users
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	query := h.db.Model(&model.User{})

	if keyword != "" {
		query = query.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var users []model.User
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&users).Error; err != nil {
		response.InternalError(c, err)
		return
	}

	items := make([]UserListItem, len(users))
	for i, u := range users {
		items[i] = UserListItem{
			ID:        u.ID,
			Username:  u.Username,
			Nickname:  u.Nickname,
			Email:     u.Email,
			Role:      u.Role,
			Status:    u.Status,
			CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if u.LastLogin != nil {
			items[i].LastLogin = u.LastLogin.Format("2006-01-02 15:04:05")
		}
	}

	response.Success(c, utils.NewPagination(items, int(total), page, pageSize))
}

// Get handles GET /api/v1/admin/users/:id
func (h *UserHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var user model.User
	if err := h.db.First(&user, id).Error; err != nil {
		response.NotFound(c, "user not found")
		return
	}

	item := UserListItem{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if user.LastLogin != nil {
		item.LastLogin = user.LastLogin.Format("2006-01-02 15:04:05")
	}

	response.Success(c, item)
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Role     string `json:"role" binding:"required,oneof=admin user"`
}

// Create handles POST /api/v1/admin/users
func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// Check if username exists
	var count int64
	h.db.Model(&model.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		response.BadRequest(c, "username already exists")
		return
	}

	user := &model.User{
		Username: req.Username,
		Nickname: req.Nickname,
		Email:    req.Email,
		Role:     req.Role,
		Status:   1,
	}

	if err := user.SetPassword(req.Password); err != nil {
		response.InternalError(c, err)
		return
	}

	if err := h.db.Create(user).Error; err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{"id": user.ID})
}

type UpdateUserRequest struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Role     string `json:"role" binding:"omitempty,oneof=admin user"`
	Status   *int   `json:"status" binding:"omitempty,oneof=0 1"`
}

// Update handles PUT /api/v1/admin/users/:id
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var user model.User
	if err := h.db.First(&user, id).Error; err != nil {
		response.NotFound(c, "user not found")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// Prevent disabling the last admin
	if req.Status != nil && *req.Status == 0 && user.Role == model.RoleAdmin {
		var adminCount int64
		h.db.Model(&model.User{}).Where("role = ? AND status = 1", model.RoleAdmin).Count(&adminCount)
		if adminCount <= 1 {
			response.BadRequest(c, "cannot disable the last admin user")
			return
		}
	}

	// Prevent changing role of the last admin
	if req.Role != "" && req.Role != model.RoleAdmin && user.Role == model.RoleAdmin {
		var adminCount int64
		h.db.Model(&model.User{}).Where("role = ? AND status = 1", model.RoleAdmin).Count(&adminCount)
		if adminCount <= 1 {
			response.BadRequest(c, "cannot change role of the last admin user")
			return
		}
	}

	updates := map[string]interface{}{}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) > 0 {
		if err := h.db.Model(&user).Updates(updates).Error; err != nil {
			response.InternalError(c, err)
			return
		}
	}

	response.OK(c, "user updated successfully")
}

type ResetPasswordRequest struct {
	Password string `json:"password" binding:"required,min=6"`
}

// ResetPassword handles POST /api/v1/admin/users/:id/reset-password
func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var user model.User
	if err := h.db.First(&user, id).Error; err != nil {
		response.NotFound(c, "user not found")
		return
	}

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := user.SetPassword(req.Password); err != nil {
		response.InternalError(c, err)
		return
	}

	if err := h.db.Model(&user).Update("password", user.Password).Error; err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, "password reset successfully")
}

// Delete handles DELETE /api/v1/admin/users/:id
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	// Get current user from context
	currentUserID, _ := c.Get("userID")
	if currentUserID.(uint) == uint(id) {
		response.BadRequest(c, "cannot delete yourself")
		return
	}

	var user model.User
	if err := h.db.First(&user, id).Error; err != nil {
		response.NotFound(c, "user not found")
		return
	}

	// Prevent deleting the last admin
	if user.Role == model.RoleAdmin {
		var adminCount int64
		h.db.Model(&model.User{}).Where("role = ?", model.RoleAdmin).Count(&adminCount)
		if adminCount <= 1 {
			response.BadRequest(c, "cannot delete the last admin user")
			return
		}
	}

	if err := h.db.Delete(&user).Error; err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, "user deleted successfully")
}
