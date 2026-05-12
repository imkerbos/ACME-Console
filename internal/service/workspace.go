package service

import (
	"errors"
	"fmt"

	"github.com/imkerbos/ACME-Console/internal/model"
	"gorm.io/gorm"
)

var (
	ErrWorkspaceNotFound     = errors.New("workspace not found")
	ErrWorkspaceAccessDenied = errors.New("access denied to workspace")
	ErrMemberNotFound        = errors.New("member not found")
	ErrCannotRemoveOwner     = errors.New("cannot remove workspace owner")
	ErrCannotChangeOwnerRole = errors.New("cannot change owner role")
	ErrUserAlreadyMember     = errors.New("user is already a member")
	ErrInvalidRole           = errors.New("invalid role")
)

type WorkspaceService struct {
	db *gorm.DB
}

func NewWorkspaceService(db *gorm.DB) *WorkspaceService {
	return &WorkspaceService{db: db}
}

// Request/Response types

type CreateWorkspaceRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
}

type UpdateWorkspaceRequest struct {
	Name        string `json:"name" binding:"omitempty,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
	Status      *int   `json:"status" binding:"omitempty,oneof=0 1"`
}

type AddMemberRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=admin member"`
}

type UpdateMemberRequest struct {
	Role string `json:"role" binding:"required,oneof=admin member"`
}

type WorkspaceResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     uint   `json:"owner_id"`
	Status      int    `json:"status"`
	Role        string `json:"role"`        // Current user's role in this workspace
	MemberCount int    `json:"member_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type MemberResponse struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

// Create creates a new workspace and adds the creator as owner
func (s *WorkspaceService) Create(userID uint, req *CreateWorkspaceRequest) (*model.Workspace, error) {
	workspace := &model.Workspace{
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     userID,
		Status:      1,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Create workspace
		if err := tx.Create(workspace).Error; err != nil {
			return fmt.Errorf("failed to create workspace: %w", err)
		}

		// Add creator as owner member
		member := &model.WorkspaceMember{
			WorkspaceID: workspace.ID,
			UserID:      userID,
			Role:        model.WorkspaceRoleOwner,
		}
		if err := tx.Create(member).Error; err != nil {
			return fmt.Errorf("failed to add owner member: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return workspace, nil
}

// ListUserWorkspaces returns all workspaces the user is a member of
func (s *WorkspaceService) ListUserWorkspaces(userID uint) ([]WorkspaceResponse, error) {
	var members []model.WorkspaceMember
	if err := s.db.Where("user_id = ?", userID).Find(&members).Error; err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return []WorkspaceResponse{}, nil
	}

	// Build workspace ID to role map
	workspaceRoles := make(map[uint]string)
	workspaceIDs := make([]uint, 0, len(members))
	for _, m := range members {
		workspaceRoles[m.WorkspaceID] = m.Role
		workspaceIDs = append(workspaceIDs, m.WorkspaceID)
	}

	// Get workspaces
	var workspaces []model.Workspace
	if err := s.db.Where("id IN ?", workspaceIDs).Order("created_at DESC").Find(&workspaces).Error; err != nil {
		return nil, err
	}

	// Get member counts
	type countResult struct {
		WorkspaceID uint
		Count       int64
	}
	var counts []countResult
	s.db.Model(&model.WorkspaceMember{}).
		Select("workspace_id, count(*) as count").
		Where("workspace_id IN ?", workspaceIDs).
		Group("workspace_id").
		Scan(&counts)

	countMap := make(map[uint]int)
	for _, c := range counts {
		countMap[c.WorkspaceID] = int(c.Count)
	}

	// Build response
	result := make([]WorkspaceResponse, 0, len(workspaces))
	for _, w := range workspaces {
		result = append(result, WorkspaceResponse{
			ID:          w.ID,
			Name:        w.Name,
			Description: w.Description,
			OwnerID:     w.OwnerID,
			Status:      w.Status,
			Role:        workspaceRoles[w.ID],
			MemberCount: countMap[w.ID],
			CreatedAt:   w.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   w.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return result, nil
}

// GetByID returns a workspace if the user has access
func (s *WorkspaceService) GetByID(workspaceID, userID uint) (*WorkspaceResponse, error) {
	role, err := s.GetUserRole(workspaceID, userID)
	if err != nil {
		return nil, err
	}

	var workspace model.Workspace
	if err := s.db.First(&workspace, workspaceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkspaceNotFound
		}
		return nil, err
	}

	// Get member count
	var memberCount int64
	s.db.Model(&model.WorkspaceMember{}).Where("workspace_id = ?", workspaceID).Count(&memberCount)

	return &WorkspaceResponse{
		ID:          workspace.ID,
		Name:        workspace.Name,
		Description: workspace.Description,
		OwnerID:     workspace.OwnerID,
		Status:      workspace.Status,
		Role:        role,
		MemberCount: int(memberCount),
		CreatedAt:   workspace.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   workspace.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// Update updates a workspace (owner/admin only)
func (s *WorkspaceService) Update(workspaceID, userID uint, req *UpdateWorkspaceRequest) error {
	if !s.CanManageWorkspace(workspaceID, userID) {
		return ErrWorkspaceAccessDenied
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) == 0 {
		return nil
	}

	return s.db.Model(&model.Workspace{}).Where("id = ?", workspaceID).Updates(updates).Error
}

// Delete deletes a workspace (owner only)
func (s *WorkspaceService) Delete(workspaceID, userID uint) error {
	var workspace model.Workspace
	if err := s.db.First(&workspace, workspaceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWorkspaceNotFound
		}
		return err
	}

	// Only owner can delete
	if workspace.OwnerID != userID {
		return ErrWorkspaceAccessDenied
	}

	// Check if workspace has certificates
	var certCount int64
	s.db.Model(&model.Certificate{}).Where("workspace_id = ?", workspaceID).Count(&certCount)
	if certCount > 0 {
		return fmt.Errorf("cannot delete workspace with %d certificates", certCount)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete all members
		if err := tx.Where("workspace_id = ?", workspaceID).Delete(&model.WorkspaceMember{}).Error; err != nil {
			return err
		}
		// Delete workspace
		return tx.Delete(&workspace).Error
	})
}

// ListMembers returns all members of a workspace
func (s *WorkspaceService) ListMembers(workspaceID, userID uint) ([]MemberResponse, error) {
	// Check access
	if _, err := s.GetUserRole(workspaceID, userID); err != nil {
		return nil, err
	}

	var members []model.WorkspaceMember
	if err := s.db.Preload("User").Where("workspace_id = ?", workspaceID).Order("created_at ASC").Find(&members).Error; err != nil {
		return nil, err
	}

	result := make([]MemberResponse, 0, len(members))
	for _, m := range members {
		resp := MemberResponse{
			ID:        m.ID,
			UserID:    m.UserID,
			Role:      m.Role,
			CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if m.User != nil {
			resp.Username = m.User.Username
			resp.Nickname = m.User.Nickname
			resp.Email = m.User.Email
		}
		result = append(result, resp)
	}

	return result, nil
}

// AddMember adds a user to a workspace (owner/admin only)
func (s *WorkspaceService) AddMember(workspaceID, currentUserID uint, req *AddMemberRequest) error {
	if !s.CanManageMembers(workspaceID, currentUserID) {
		return ErrWorkspaceAccessDenied
	}

	// Validate role
	if req.Role != model.WorkspaceRoleAdmin && req.Role != model.WorkspaceRoleMember {
		return ErrInvalidRole
	}

	// Check if user exists
	var user model.User
	if err := s.db.First(&user, req.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found")
		}
		return err
	}

	// Check if already a member
	var existing model.WorkspaceMember
	err := s.db.Where("workspace_id = ? AND user_id = ?", workspaceID, req.UserID).First(&existing).Error
	if err == nil {
		return ErrUserAlreadyMember
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	member := &model.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      req.UserID,
		Role:        req.Role,
	}

	return s.db.Create(member).Error
}

// UpdateMember updates a member's role (owner/admin only)
func (s *WorkspaceService) UpdateMember(workspaceID, currentUserID, memberUserID uint, req *UpdateMemberRequest) error {
	if !s.CanManageMembers(workspaceID, currentUserID) {
		return ErrWorkspaceAccessDenied
	}

	var member model.WorkspaceMember
	if err := s.db.Where("workspace_id = ? AND user_id = ?", workspaceID, memberUserID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMemberNotFound
		}
		return err
	}

	// Cannot change owner's role
	if member.Role == model.WorkspaceRoleOwner {
		return ErrCannotChangeOwnerRole
	}

	// Validate role
	if req.Role != model.WorkspaceRoleAdmin && req.Role != model.WorkspaceRoleMember {
		return ErrInvalidRole
	}

	return s.db.Model(&member).Update("role", req.Role).Error
}

// RemoveMember removes a user from a workspace (owner/admin only)
func (s *WorkspaceService) RemoveMember(workspaceID, currentUserID, memberUserID uint) error {
	if !s.CanManageMembers(workspaceID, currentUserID) {
		return ErrWorkspaceAccessDenied
	}

	var member model.WorkspaceMember
	if err := s.db.Where("workspace_id = ? AND user_id = ?", workspaceID, memberUserID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMemberNotFound
		}
		return err
	}

	// Cannot remove owner
	if member.Role == model.WorkspaceRoleOwner {
		return ErrCannotRemoveOwner
	}

	return s.db.Delete(&member).Error
}

// GetUserRole returns the user's role in a workspace
func (s *WorkspaceService) GetUserRole(workspaceID, userID uint) (string, error) {
	var member model.WorkspaceMember
	if err := s.db.Where("workspace_id = ? AND user_id = ?", workspaceID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrWorkspaceAccessDenied
		}
		return "", err
	}
	return member.Role, nil
}

// CanManageWorkspace checks if user can manage workspace settings (owner/admin)
func (s *WorkspaceService) CanManageWorkspace(workspaceID, userID uint) bool {
	role, err := s.GetUserRole(workspaceID, userID)
	if err != nil {
		return false
	}
	return role == model.WorkspaceRoleOwner || role == model.WorkspaceRoleAdmin
}

// CanManageMembers checks if user can manage members (owner/admin)
func (s *WorkspaceService) CanManageMembers(workspaceID, userID uint) bool {
	role, err := s.GetUserRole(workspaceID, userID)
	if err != nil {
		return false
	}
	return role == model.WorkspaceRoleOwner || role == model.WorkspaceRoleAdmin
}

// CanManageCertificates checks if user can manage certificates (owner/admin)
func (s *WorkspaceService) CanManageCertificates(workspaceID, userID uint) bool {
	role, err := s.GetUserRole(workspaceID, userID)
	if err != nil {
		return false
	}
	return role == model.WorkspaceRoleOwner || role == model.WorkspaceRoleAdmin
}

// CanViewCertificates checks if user can view certificates (any member)
func (s *WorkspaceService) CanViewCertificates(workspaceID, userID uint) bool {
	_, err := s.GetUserRole(workspaceID, userID)
	return err == nil
}
