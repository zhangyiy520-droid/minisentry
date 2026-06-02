package dto

import (
	"time"

	"minisentry/internal/models"

	"github.com/google/uuid"
)

// RegisterRequest represents the request payload for user registration
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Name     string `json:"name" validate:"required,min=1,max=255"`
}

// LoginRequest represents the request payload for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse represents the response after successful authentication
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int64        `json:"expires_in"`
	User         UserResponse `json:"user"`
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

// ConvertFromModel converts models.UserResponse to dto.UserResponse
func (ur *UserResponse) ConvertFromModel(modelUser models.UserResponse) {
	ur.ID = modelUser.ID
	ur.Email = modelUser.Email
	ur.Name = modelUser.Name
	ur.AvatarURL = modelUser.AvatarURL
	ur.IsActive = modelUser.IsActive
	ur.EmailVerified = modelUser.EmailVerified
	ur.CreatedAt = modelUser.CreatedAt
	ur.UpdatedAt = modelUser.UpdatedAt
}

// UpdateProfileRequest represents the request payload for updating user profile
type UpdateProfileRequest struct {
	Name      *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	AvatarURL *string `json:"avatar_url,omitempty" validate:"omitempty,max=500,url"`
}

// RefreshTokenRequest represents the request payload for refreshing JWT tokens
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ChangePasswordRequest represents the request payload for changing user password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=72"`
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a standard success response
type SuccessResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}