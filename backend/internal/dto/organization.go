package dto

import (
	"time"

	"minisentry/internal/models"

	"github.com/google/uuid"
)

// CreateOrganizationRequest represents the request payload for creating an organization
type CreateOrganizationRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Slug        string  `json:"slug" validate:"required,min=1,max=100,alphanum"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
}

// UpdateOrganizationRequest represents the request payload for updating an organization
type UpdateOrganizationRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
}

// OrganizationResponse represents the response payload for organization details
type OrganizationResponse struct {
	ID          uuid.UUID             `json:"id"`
	Name        string                `json:"name"`
	Slug        string                `json:"slug"`
	Description *string               `json:"description"`
	Role        models.OrganizationRole `json:"role"` // Current user's role in the organization
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

// OrganizationListResponse represents the response payload for listing organizations
type OrganizationListResponse struct {
	Organizations []OrganizationResponse `json:"organizations"`
}

// AddMemberRequest represents the request payload for adding a member to an organization
type AddMemberRequest struct {
	Email string                `json:"email" validate:"required,email"`
	Role  models.OrganizationRole `json:"role" validate:"required,oneof=admin member"`
}

// UpdateMemberRoleRequest represents the request payload for updating a member's role
type UpdateMemberRoleRequest struct {
	Role models.OrganizationRole `json:"role" validate:"required,oneof=owner admin member"`
}

// OrganizationMemberResponse represents the response payload for organization member details
type OrganizationMemberResponse struct {
	ID             uuid.UUID             `json:"id"`
	OrganizationID uuid.UUID             `json:"organization_id"`
	UserID         uuid.UUID             `json:"user_id"`
	Role           models.OrganizationRole `json:"role"`
	User           UserSummary           `json:"user"`
	JoinedAt       time.Time             `json:"joined_at"`
}

// UserSummary represents a summary of user information for member responses
type UserSummary struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL *string   `json:"avatar_url"`
}

// OrganizationMembersResponse represents the response payload for listing organization members
type OrganizationMembersResponse struct {
	Members []OrganizationMemberResponse `json:"members"`
}

// ToOrganizationResponse converts an Organization model with role to OrganizationResponse
func ToOrganizationResponse(org *models.Organization, role models.OrganizationRole) OrganizationResponse {
	return OrganizationResponse{
		ID:          org.ID,
		Name:        org.Name,
		Slug:        org.Slug,
		Description: org.Description,
		Role:        role,
		CreatedAt:   org.CreatedAt,
		UpdatedAt:   org.UpdatedAt,
	}
}

// ToOrganizationMemberResponse converts an OrganizationMember model to OrganizationMemberResponse
func ToOrganizationMemberResponse(member *models.OrganizationMember) OrganizationMemberResponse {
	return OrganizationMemberResponse{
		ID:             member.ID,
		OrganizationID: member.OrganizationID,
		UserID:         member.UserID,
		Role:           member.Role,
		User: UserSummary{
			ID:        member.User.ID,
			Email:     member.User.Email,
			Name:      member.User.Name,
			AvatarURL: member.User.AvatarURL,
		},
		JoinedAt: member.JoinedAt,
	}
}