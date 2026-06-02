package dto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	"minisentry/internal/models"

	"github.com/google/uuid"
)

// CreateProjectRequest represents the request payload for creating a project
type CreateProjectRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Slug        string  `json:"slug" validate:"required,min=1,max=100,alphanum"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Platform    string  `json:"platform" validate:"required,oneof=javascript python go java dotnet php ruby"`
}

// UpdateProjectRequest represents the request payload for updating a project
type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Platform    *string `json:"platform,omitempty" validate:"omitempty,oneof=javascript python go java dotnet php ruby"`
}

// ProjectResponse represents the response payload for project details
type ProjectResponse struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Description    *string   `json:"description"`
	Platform       string    `json:"platform"`
	DSN            string    `json:"dsn"`
	PublicKey      string    `json:"public_key"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ProjectListResponse represents the response payload for listing projects
type ProjectListResponse struct {
	Projects []ProjectResponse `json:"projects"`
}

// ProjectConfigurationRequest represents the request payload for updating project configuration
type ProjectConfigurationRequest struct {
	IsActive *bool `json:"is_active,omitempty"`
	Platform *string `json:"platform,omitempty" validate:"omitempty,oneof=javascript python go java dotnet php ruby"`
}

// ProjectKeyResponse represents the response after regenerating project key
type ProjectKeyResponse struct {
	PublicKey string `json:"public_key"`
	DSN       string `json:"dsn"`
}

// DSNInfo represents parsed DSN information
type DSNInfo struct {
	PublicKey string    `json:"public_key"`
	Host      string    `json:"host"`
	ProjectID uuid.UUID `json:"project_id"`
	Scheme    string    `json:"scheme"`
}

// GenerateProjectKey generates a new 32-character hex key
func GenerateProjectKey() string {
	bytes := make([]byte, 16) // 16 bytes = 32 hex characters
	if _, err := rand.Read(bytes); err != nil {
		// Fallback in case of crypto/rand failure
		panic("failed to generate random key: " + err.Error())
	}
	return hex.EncodeToString(bytes)
}

// GenerateDSN creates a DSN in the format: https://{public_key}@{host}/{project_id}
func GenerateDSN(publicKey string, host string, projectID uuid.UUID) string {
	// Ensure host doesn't include scheme
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")
	
	// Use HTTPS as required
	return fmt.Sprintf("https://%s@%s/%s", publicKey, host, projectID.String())
}

// ParseDSN parses a DSN string and returns DSN info
func ParseDSN(dsn string) (*DSNInfo, error) {
	if dsn == "" {
		return nil, fmt.Errorf("DSN cannot be empty")
	}

	parsedURL, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid DSN format: %w", err)
	}

	if parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("DSN must use HTTPS scheme")
	}

	if parsedURL.User == nil {
		return nil, fmt.Errorf("DSN missing public key")
	}

	publicKey := parsedURL.User.Username()
	if publicKey == "" {
		return nil, fmt.Errorf("DSN missing public key")
	}

	if len(publicKey) != 32 {
		return nil, fmt.Errorf("invalid public key length, expected 32 characters")
	}

	// Extract project ID from path (should be /{project_id})
	path := strings.Trim(parsedURL.Path, "/")
	if path == "" {
		return nil, fmt.Errorf("DSN missing project ID")
	}

	projectID, err := uuid.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID in DSN: %w", err)
	}

	return &DSNInfo{
		PublicKey: publicKey,
		Host:      parsedURL.Host,
		ProjectID: projectID,
		Scheme:    parsedURL.Scheme,
	}, nil
}

// ValidateDSN validates DSN format and returns project ID if valid
func ValidateDSN(dsn string) (uuid.UUID, error) {
	info, err := ParseDSN(dsn)
	if err != nil {
		return uuid.Nil, err
	}
	return info.ProjectID, nil
}

// ToProjectResponse converts a Project model to ProjectResponse
func ToProjectResponse(project *models.Project) ProjectResponse {
	return ProjectResponse{
		ID:             project.ID,
		OrganizationID: project.OrganizationID,
		Name:           project.Name,
		Slug:           project.Slug,
		Description:    project.Description,
		Platform:       project.Platform,
		DSN:            project.DSN,
		PublicKey:      project.PublicKey,
		IsActive:       project.IsActive,
		CreatedAt:      project.CreatedAt,
		UpdatedAt:      project.UpdatedAt,
	}
}

// ToProjectListResponse converts a slice of Project models to ProjectListResponse
func ToProjectListResponse(projects []models.Project) ProjectListResponse {
	responses := make([]ProjectResponse, len(projects))
	for i, project := range projects {
		responses[i] = ToProjectResponse(&project)
	}
	return ProjectListResponse{
		Projects: responses,
	}
}

// ValidateProjectSlug validates project slug format
func ValidateProjectSlug(slug string) error {
	if slug == "" {
		return fmt.Errorf("slug cannot be empty")
	}
	
	if len(slug) > 100 {
		return fmt.Errorf("slug too long, maximum 100 characters")
	}
	
	// Allow alphanumeric characters, hyphens, and underscores
	for _, char := range slug {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '-' || char == '_') {
			return fmt.Errorf("slug contains invalid characters, only alphanumeric, hyphens, and underscores allowed")
		}
	}
	
	return nil
}

// NormalizeProjectSlug normalizes and validates project slug
func NormalizeProjectSlug(slug string) (string, error) {
	// Convert to lowercase and trim whitespace
	normalized := strings.ToLower(strings.TrimSpace(slug))
	
	// Validate the normalized slug
	if err := ValidateProjectSlug(normalized); err != nil {
		return "", err
	}
	
	return normalized, nil
}

// SupportedPlatforms returns list of supported project platforms
func SupportedPlatforms() []string {
	return []string{
		"javascript",
		"python", 
		"go",
		"java",
		"dotnet",
		"php",
		"ruby",
	}
}

// IsPlatformSupported checks if platform is supported
func IsPlatformSupported(platform string) bool {
	for _, supported := range SupportedPlatforms() {
		if platform == supported {
			return true
		}
	}
	return false
}