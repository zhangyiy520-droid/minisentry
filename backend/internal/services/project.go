package services

import (
	"errors"
	"fmt"

	"minisentry/internal/database"
	"minisentry/internal/dto"
	"minisentry/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrProjectNotFound          = errors.New("project not found")
	ErrProjectSlugExists        = errors.New("project slug already exists in organization")
	ErrProjectAccessDenied      = errors.New("access denied to project")
	ErrProjectInvalidPlatform   = errors.New("invalid project platform")
	ErrProjectDSNInvalid        = errors.New("invalid project DSN")
	ErrProjectInactive          = errors.New("project is inactive")
)

type ProjectService struct {
	db      *database.DB
	dsnHost string
}

// NewProjectService creates a new project service
func NewProjectService(db *database.DB, dsnHost string) *ProjectService {
	return &ProjectService{
		db:      db,
		dsnHost: dsnHost,
	}
}

// CreateProject creates a new project within an organization
func (s *ProjectService) CreateProject(userID, orgID uuid.UUID, name, slug, platform string, description *string) (*models.Project, error) {
	// Normalize and validate slug
	normalizedSlug, err := dto.NormalizeProjectSlug(slug)
	if err != nil {
		return nil, fmt.Errorf("invalid slug: %w", err)
	}

	// Validate platform
	if !dto.IsPlatformSupported(platform) {
		return nil, ErrProjectInvalidPlatform
	}

	// Check if user has permission to create projects (owner or admin)
	var member models.OrganizationMember
	if err := s.db.DB.Where("organization_id = ? AND user_id = ?", orgID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotMember
		}
		return nil, fmt.Errorf("failed to check organization membership: %w", err)
	}

	if member.Role != models.RoleOwner && member.Role != models.RoleAdmin {
		return nil, ErrInsufficientPermissions
	}

	// Check if slug already exists in the organization
	var existingProject models.Project
	if err := s.db.DB.Where("organization_id = ? AND slug = ?", orgID, normalizedSlug).First(&existingProject).Error; err == nil {
		return nil, ErrProjectSlugExists
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

	// Generate keys
	publicKey := dto.GenerateProjectKey()
	secretKey := dto.GenerateProjectKey()

	// Create project
	project := &models.Project{
		OrganizationID: orgID,
		Name:           name,
		Slug:           normalizedSlug,
		Description:    description,
		Platform:       platform,
		PublicKey:      publicKey,
		SecretKey:      secretKey,
		IsActive:       true,
	}

	// Generate DSN after ID is set
	if err := tx.Create(project).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Update with generated DSN
	project.DSN = dto.GenerateDSN(publicKey, s.dsnHost, project.ID)
	if err := tx.Save(project).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update project DSN: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return project, nil
}

// GetProject gets project by ID with permission check
func (s *ProjectService) GetProject(userID, projectID uuid.UUID) (*models.Project, error) {
	// Get project with organization
	var project models.Project
	if err := s.db.DB.Preload("Organization").Where("id = ?", projectID).First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Check if user has access through organization membership
	var member models.OrganizationMember
	if err := s.db.DB.Where("organization_id = ? AND user_id = ?", project.OrganizationID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectAccessDenied
		}
		return nil, fmt.Errorf("failed to check project access: %w", err)
	}

	return &project, nil
}

// GetOrganizationProjects gets all projects in an organization
func (s *ProjectService) GetOrganizationProjects(userID, orgID uuid.UUID) ([]models.Project, error) {
	// Check if user is a member of the organization
	var member models.OrganizationMember
	if err := s.db.DB.Where("organization_id = ? AND user_id = ?", orgID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotMember
		}
		return nil, fmt.Errorf("failed to check organization membership: %w", err)
	}

	// Get all projects in the organization
	var projects []models.Project
	if err := s.db.DB.Where("organization_id = ?", orgID).Order("name ASC").Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("failed to get organization projects: %w", err)
	}

	return projects, nil
}

// UpdateProject updates project details
func (s *ProjectService) UpdateProject(userID, projectID uuid.UUID, name, platform *string, description *string) (*models.Project, error) {
	// Get project with organization access check
	project, err := s.GetProject(userID, projectID)
	if err != nil {
		return nil, err
	}

	// Check if user has permission to update (owner or admin)
	var member models.OrganizationMember
	if err := s.db.DB.Where("organization_id = ? AND user_id = ?", project.OrganizationID, userID).First(&member).Error; err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}

	if member.Role != models.RoleOwner && member.Role != models.RoleAdmin {
		return nil, ErrInsufficientPermissions
	}

	// Validate platform if provided
	if platform != nil && !dto.IsPlatformSupported(*platform) {
		return nil, ErrProjectInvalidPlatform
	}

	// Update fields
	updates := make(map[string]interface{})
	if name != nil {
		updates["name"] = *name
	}
	if platform != nil {
		updates["platform"] = *platform
	}
	if description != nil {
		updates["description"] = *description
	}

	if len(updates) > 0 {
		if err := s.db.DB.Model(project).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update project: %w", err)
		}
	}

	return project, nil
}

// DeleteProject soft deletes a project
func (s *ProjectService) DeleteProject(userID, projectID uuid.UUID) error {
	// Get project with organization access check
	project, err := s.GetProject(userID, projectID)
	if err != nil {
		return err
	}

	// Check if user has permission to delete (owner or admin)
	var member models.OrganizationMember
	if err := s.db.DB.Where("organization_id = ? AND user_id = ?", project.OrganizationID, userID).First(&member).Error; err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}

	if member.Role != models.RoleOwner && member.Role != models.RoleAdmin {
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

	// Soft delete related data (issues, events, releases would be handled by cascade)
	// For now, just delete the project
	if err := tx.Delete(project).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GenerateProjectDSN generates a DSN for the project (used during creation)
func (s *ProjectService) GenerateProjectDSN(publicKey string, projectID uuid.UUID) string {
	return dto.GenerateDSN(publicKey, s.dsnHost, projectID)
}

// RegenerateProjectKey regenerates the project's API key for security
func (s *ProjectService) RegenerateProjectKey(userID, projectID uuid.UUID) (*models.Project, error) {
	// Get project with organization access check
	project, err := s.GetProject(userID, projectID)
	if err != nil {
		return nil, err
	}

	// Check if user has permission to regenerate keys (owner or admin)
	var member models.OrganizationMember
	if err := s.db.DB.Where("organization_id = ? AND user_id = ?", project.OrganizationID, userID).First(&member).Error; err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}

	if member.Role != models.RoleOwner && member.Role != models.RoleAdmin {
		return nil, ErrInsufficientPermissions
	}

	// Generate new keys
	newPublicKey := dto.GenerateProjectKey()
	newSecretKey := dto.GenerateProjectKey()
	newDSN := dto.GenerateDSN(newPublicKey, s.dsnHost, project.ID)

	// Update project with new keys
	updates := map[string]interface{}{
		"public_key": newPublicKey,
		"secret_key": newSecretKey,
		"dsn":        newDSN,
	}

	if err := s.db.DB.Model(project).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to regenerate project keys: %w", err)
	}

	// Update the project struct with new values
	project.PublicKey = newPublicKey
	project.SecretKey = newSecretKey
	project.DSN = newDSN

	return project, nil
}

// GetProjectByDSN finds project by DSN for error ingestion
func (s *ProjectService) GetProjectByDSN(dsn string) (*models.Project, error) {
	// If DSN looks like just a public key (32 hex chars), find by public key
	if len(dsn) == 32 {
		return s.GetProjectByPublicKey(dsn)
	}

	// Parse and validate DSN
	dsnInfo, err := dto.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid DSN: %w", err)
	}

	// Find project by ID and public key
	var project models.Project
	if err := s.db.DB.Where("id = ? AND public_key = ?", dsnInfo.ProjectID, dsnInfo.PublicKey).First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project by DSN: %w", err)
	}

	// Check if project is active
	if !project.IsActive {
		return nil, ErrProjectInactive
	}

	return &project, nil
}

// GetProjectByPublicKey finds project by public key for error ingestion
func (s *ProjectService) GetProjectByPublicKey(publicKey string) (*models.Project, error) {
	if len(publicKey) != 32 {
		return nil, fmt.Errorf("invalid public key length: expected 32 characters")
	}

	var project models.Project
	if err := s.db.DB.Where("public_key = ?", publicKey).First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project by public key: %w", err)
	}

	// Check if project is active
	if !project.IsActive {
		return nil, ErrProjectInactive
	}

	return &project, nil
}

// UpdateProjectConfiguration updates project settings
func (s *ProjectService) UpdateProjectConfiguration(userID, projectID uuid.UUID, isActive *bool, platform *string) (*models.Project, error) {
	// Get project with organization access check
	project, err := s.GetProject(userID, projectID)
	if err != nil {
		return nil, err
	}

	// Check if user has permission to update configuration (owner or admin)
	var member models.OrganizationMember
	if err := s.db.DB.Where("organization_id = ? AND user_id = ?", project.OrganizationID, userID).First(&member).Error; err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}

	if member.Role != models.RoleOwner && member.Role != models.RoleAdmin {
		return nil, ErrInsufficientPermissions
	}

	// Validate platform if provided
	if platform != nil && !dto.IsPlatformSupported(*platform) {
		return nil, ErrProjectInvalidPlatform
	}

	// Update configuration
	updates := make(map[string]interface{})
	if isActive != nil {
		updates["is_active"] = *isActive
	}
	if platform != nil {
		updates["platform"] = *platform
	}

	if len(updates) > 0 {
		if err := s.db.DB.Model(project).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update project configuration: %w", err)
		}
	}

	return project, nil
}

// GetProjectBySlug gets project by organization ID and slug
func (s *ProjectService) GetProjectBySlug(userID, orgID uuid.UUID, slug string) (*models.Project, error) {
	// Check organization membership
	var member models.OrganizationMember
	if err := s.db.DB.Where("organization_id = ? AND user_id = ?", orgID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotMember
		}
		return nil, fmt.Errorf("failed to check organization membership: %w", err)
	}

	// Get project by slug within organization
	var project models.Project
	if err := s.db.DB.Where("organization_id = ? AND slug = ?", orgID, slug).First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project by slug: %w", err)
	}

	return &project, nil
}

// ValidateProjectAccess checks if user has access to project
func (s *ProjectService) ValidateProjectAccess(userID, projectID uuid.UUID) (models.OrganizationRole, error) {
	// Get project
	var project models.Project
	if err := s.db.DB.Where("id = ?", projectID).First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrProjectNotFound
		}
		return "", fmt.Errorf("failed to get project: %w", err)
	}

	// Check organization membership and get role
	var member models.OrganizationMember
	if err := s.db.DB.Where("organization_id = ? AND user_id = ?", project.OrganizationID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrProjectAccessDenied
		}
		return "", fmt.Errorf("failed to check project access: %w", err)
	}

	return member.Role, nil
}