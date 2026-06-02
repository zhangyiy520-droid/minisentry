package services

import (
	"errors"
	"fmt"
	"strings"

	"minisentry/internal/database"
	"minisentry/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrOrganizationNotFound    = errors.New("organization not found")
	ErrOrganizationSlugExists  = errors.New("organization slug already exists")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrUserNotMember          = errors.New("user is not a member of this organization")
	ErrMemberNotFound         = errors.New("member not found")
	ErrCannotRemoveOwner      = errors.New("cannot remove organization owner")
	ErrCannotChangeOwnerRole  = errors.New("cannot change owner role")
	ErrOrgUserNotFound        = errors.New("user not found")
	ErrUserAlreadyMember      = errors.New("user is already a member of this organization")
)

type OrganizationService struct {
	db *database.DB
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(db *database.DB) *OrganizationService {
	return &OrganizationService{
		db: db,
	}
}

// CreateOrganization creates a new organization with the user as owner
func (s *OrganizationService) CreateOrganization(userID uuid.UUID, name, slug string, description *string) (*models.Organization, error) {
	// Normalize slug
	slug = strings.ToLower(strings.TrimSpace(slug))
	
	// Check if slug already exists
	var existingOrg models.Organization
	if err := s.db.DB.Where("slug = ?", slug).First(&existingOrg).Error; err == nil {
		return nil, ErrOrganizationSlugExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check slug uniqueness: %w", err)
	}

	// Start transaction
	tx := s.db.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create organization
	org := &models.Organization{
		Name:        name,
		Slug:        slug,
		Description: description,
	}
	
	if err := tx.Create(org).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Add user as owner
	member := &models.OrganizationMember{
		OrganizationID: org.ID,
		UserID:         userID,
		Role:           models.RoleOwner,
	}
	
	if err := tx.Create(member).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to add user as owner: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return org, nil
}

// GetOrganization gets organization by ID with permission check
func (s *OrganizationService) GetOrganization(userID, orgID uuid.UUID) (*models.Organization, models.OrganizationRole, error) {
	// Check if user is a member and get their role
	var member models.OrganizationMember
	if err := s.db.DB.Where("organization_id = ? AND user_id = ?", orgID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrUserNotMember
		}
		return nil, "", fmt.Errorf("failed to check membership: %w", err)
	}

	// Get organization
	var org models.Organization
	if err := s.db.DB.Where("id = ?", orgID).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrOrganizationNotFound
		}
		return nil, "", fmt.Errorf("failed to get organization: %w", err)
	}

	return &org, member.Role, nil
}

// GetUserOrganizations gets all organizations user belongs to
func (s *OrganizationService) GetUserOrganizations(userID uuid.UUID) ([]models.OrganizationWithRole, error) {
	var results []models.OrganizationWithRole
	
	// Join organizations with member roles
	if err := s.db.DB.Raw(`
		SELECT o.*, om.role 
		FROM organizations o 
		INNER JOIN organization_members om ON o.id = om.organization_id 
		WHERE om.user_id = ? 
		ORDER BY o.name ASC
	`, userID).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}

	return results, nil
}

// UpdateOrganization updates organization details
func (s *OrganizationService) UpdateOrganization(userID, orgID uuid.UUID, name *string, description *string) (*models.Organization, error) {
	// Check permissions (owner or admin required)
	role, err := s.getUserRole(userID, orgID)
	if err != nil {
		return nil, err
	}
	
	if role != models.RoleOwner && role != models.RoleAdmin {
		return nil, ErrInsufficientPermissions
	}

	// Get organization
	var org models.Organization
	if err := s.db.DB.Where("id = ?", orgID).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Update fields
	updates := make(map[string]interface{})
	if name != nil {
		updates["name"] = *name
	}
	if description != nil {
		updates["description"] = *description
	}

	if len(updates) > 0 {
		if err := s.db.DB.Model(&org).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update organization: %w", err)
		}
	}

	return &org, nil
}

// DeleteOrganization soft deletes organization (owner only)
func (s *OrganizationService) DeleteOrganization(userID, orgID uuid.UUID) error {
	// Check permissions (owner only)
	role, err := s.getUserRole(userID, orgID)
	if err != nil {
		return err
	}
	
	if role != models.RoleOwner {
		return ErrInsufficientPermissions
	}

	// Start transaction
	tx := s.db.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete all members
	if err := tx.Where("organization_id = ?", orgID).Delete(&models.OrganizationMember{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete members: %w", err)
	}

	// Delete organization
	if err := tx.Delete(&models.Organization{}, orgID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// AddMember invites user to organization
func (s *OrganizationService) AddMember(userID, orgID uuid.UUID, email string, role models.OrganizationRole) (*models.OrganizationMember, error) {
	// Check permissions (owner or admin required)
	currentRole, err := s.getUserRole(userID, orgID)
	if err != nil {
		return nil, err
	}
	
	if currentRole != models.RoleOwner && currentRole != models.RoleAdmin {
		return nil, ErrInsufficientPermissions
	}

	// Only owners can add admins
	if role == models.RoleAdmin && currentRole != models.RoleOwner {
		return nil, ErrInsufficientPermissions
	}

	// Cannot add owners
	if role == models.RoleOwner {
		return nil, ErrInsufficientPermissions
	}

	// Find user by email
	var user models.User
	if err := s.db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgUserNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Check if user is already a member
	var existingMember models.OrganizationMember
	if err := s.db.DB.Where("organization_id = ? AND user_id = ?", orgID, user.ID).First(&existingMember).Error; err == nil {
		return nil, ErrUserAlreadyMember
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing membership: %w", err)
	}

	// Create member
	member := &models.OrganizationMember{
		OrganizationID: orgID,
		UserID:         user.ID,
		Role:           role,
	}
	
	if err := s.db.DB.Create(member).Error; err != nil {
		return nil, fmt.Errorf("failed to create member: %w", err)
	}

	// Preload user information
	if err := s.db.DB.Preload("User").First(member, member.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load member with user: %w", err)
	}

	return member, nil
}

// RemoveMember removes user from organization
func (s *OrganizationService) RemoveMember(userID, orgID, targetUserID uuid.UUID) error {
	// Check permissions (owner or admin required)
	currentRole, err := s.getUserRole(userID, orgID)
	if err != nil {
		return err
	}
	
	if currentRole != models.RoleOwner && currentRole != models.RoleAdmin {
		return ErrInsufficientPermissions
	}

	// Get target member
	var targetMember models.OrganizationMember
	if err := s.db.DB.Where("organization_id = ? AND user_id = ?", orgID, targetUserID).First(&targetMember).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMemberNotFound
		}
		return fmt.Errorf("failed to get member: %w", err)
	}

	// Cannot remove owner
	if targetMember.Role == models.RoleOwner {
		return ErrCannotRemoveOwner
	}

	// Only owners can remove admins
	if targetMember.Role == models.RoleAdmin && currentRole != models.RoleOwner {
		return ErrInsufficientPermissions
	}

	// Remove member
	if err := s.db.DB.Delete(&targetMember).Error; err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	return nil
}

// UpdateMemberRole changes member role
func (s *OrganizationService) UpdateMemberRole(userID, orgID, targetUserID uuid.UUID, newRole models.OrganizationRole) (*models.OrganizationMember, error) {
	// Check permissions (owner only)
	currentRole, err := s.getUserRole(userID, orgID)
	if err != nil {
		return nil, err
	}
	
	if currentRole != models.RoleOwner {
		return nil, ErrInsufficientPermissions
	}

	// Get target member
	var targetMember models.OrganizationMember
	if err := s.db.DB.Where("organization_id = ? AND user_id = ?", orgID, targetUserID).First(&targetMember).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMemberNotFound
		}
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	// Cannot change owner role
	if targetMember.Role == models.RoleOwner {
		return nil, ErrCannotChangeOwnerRole
	}

	// Cannot make someone an owner
	if newRole == models.RoleOwner {
		return nil, ErrInsufficientPermissions
	}

	// Update role
	targetMember.Role = newRole
	if err := s.db.DB.Save(&targetMember).Error; err != nil {
		return nil, fmt.Errorf("failed to update member role: %w", err)
	}

	// Preload user information
	if err := s.db.DB.Preload("User").First(&targetMember, targetMember.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load member with user: %w", err)
	}

	return &targetMember, nil
}

// GetOrganizationMembers lists organization members
func (s *OrganizationService) GetOrganizationMembers(userID, orgID uuid.UUID) ([]models.OrganizationMember, error) {
	// Check if user is a member (any role can view members)
	_, err := s.getUserRole(userID, orgID)
	if err != nil {
		return nil, err
	}

	// Get members with user information
	var members []models.OrganizationMember
	if err := s.db.DB.Where("organization_id = ?", orgID).Preload("User").Find(&members).Error; err != nil {
		return nil, fmt.Errorf("failed to get organization members: %w", err)
	}

	return members, nil
}

// getUserRole gets user's role in organization
func (s *OrganizationService) getUserRole(userID, orgID uuid.UUID) (models.OrganizationRole, error) {
	var member models.OrganizationMember
	if err := s.db.DB.Where("organization_id = ? AND user_id = ?", orgID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrUserNotMember
		}
		return "", fmt.Errorf("failed to get user role: %w", err)
	}
	
	return member.Role, nil
}