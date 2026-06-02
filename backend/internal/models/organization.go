package models

import (
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	BaseModel
	Name        string    `json:"name" gorm:"not null;size:255"`
	Slug        string    `json:"slug" gorm:"uniqueIndex;not null;size:100"`
	Description *string   `json:"description" gorm:"type:text"`
	
	// Relationships
	Members []OrganizationMember `json:"members,omitempty" gorm:"foreignKey:OrganizationID"`
	Projects []Project           `json:"projects,omitempty" gorm:"foreignKey:OrganizationID"`
}

type OrganizationRole string

const (
	RoleOwner  OrganizationRole = "owner"
	RoleAdmin  OrganizationRole = "admin"
	RoleMember OrganizationRole = "member"
)

type OrganizationMember struct {
	ID             uuid.UUID        `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrganizationID uuid.UUID        `json:"organization_id" gorm:"not null;index"`
	UserID         uuid.UUID        `json:"user_id" gorm:"not null;index"`
	Role           OrganizationRole `json:"role" gorm:"not null;default:'member';size:50"`
	JoinedAt       time.Time        `json:"joined_at" gorm:"default:now()"`
	
	// Relationships
	Organization Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	User         User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// OrganizationWithRole represents an organization with the user's role
type OrganizationWithRole struct {
	Organization
	Role OrganizationRole `json:"role"`
}