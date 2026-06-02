package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	BaseModel
	Email         string    `json:"email" gorm:"uniqueIndex;not null;size:255"`
	PasswordHash  string    `json:"-" gorm:"not null;size:255"`
	Name          string    `json:"name" gorm:"not null;size:255"`
	AvatarURL     *string   `json:"avatar_url" gorm:"size:500"`
	IsActive      bool      `json:"is_active" gorm:"default:true"`
	EmailVerified bool      `json:"email_verified" gorm:"default:false"`
}

// UserResponse represents user data returned to clients (without sensitive fields)
type UserResponse struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	AvatarURL     *string   `json:"avatar_url"`
	IsActive      bool      `json:"is_active"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:            u.ID,
		Email:         u.Email,
		Name:          u.Name,
		AvatarURL:     u.AvatarURL,
		IsActive:      u.IsActive,
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}