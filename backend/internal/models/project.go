package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Project struct {
	BaseModel
	OrganizationID uuid.UUID `json:"organization_id" gorm:"not null;index"`
	Name           string    `json:"name" gorm:"not null;size:255"`
	Slug           string    `json:"slug" gorm:"not null;size:100;index:idx_org_project_slug,unique"`
	Description    *string   `json:"description" gorm:"type:text"`
	Platform       string    `json:"platform" gorm:"not null;default:'javascript';size:50"`
	DSN            string    `json:"dsn" gorm:"uniqueIndex;not null;size:255"`
	PublicKey      string    `json:"public_key" gorm:"not null;size:255"`
	SecretKey      string    `json:"-" gorm:"not null;size:255"` // Hidden from JSON
	IsActive       bool      `json:"is_active" gorm:"default:true"`
	
	// Relationships
	Organization Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	Issues       []Issue      `json:"issues,omitempty" gorm:"foreignKey:ProjectID"`
	Events       []Event      `json:"events,omitempty" gorm:"foreignKey:ProjectID"`
	Releases     []Release    `json:"releases,omitempty" gorm:"foreignKey:ProjectID"`
}

// BeforeCreate generates ID before creating the project (keys and DSN are handled by service)
func (p *Project) BeforeCreate(tx *gorm.DB) error {
	// Call parent BeforeCreate to generate ID
	return p.BaseModel.BeforeCreate(tx)
}

// ProjectResponse represents project data with public key but without secret
type ProjectResponse struct {
	Project
	PublicKey string `json:"public_key"`
}