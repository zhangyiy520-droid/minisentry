package services

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"minisentry/internal/database"
	"minisentry/internal/dto"
	"minisentry/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	db              *database.DB
	passwordService *PasswordService
}

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailExists       = errors.New("email already exists")
	ErrInvalidEmail      = errors.New("invalid email format")
	ErrInvalidName       = errors.New("invalid name")
	ErrPasswordTooWeak   = errors.New("password does not meet strength requirements")
	ErrUserInactive      = errors.New("user account is inactive")
	emailRegex           = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// NewUserService creates a new user service
func NewUserService(db *database.DB, passwordService *PasswordService) *UserService {
	return &UserService{
		db:              db,
		passwordService: passwordService,
	}
}

// CreateUser creates a new user account
func (s *UserService) CreateUser(req *dto.RegisterRequest) (*models.User, error) {
	// Validate input
	if err := s.validateRegistrationRequest(req); err != nil {
		return nil, err
	}

	// Check if email already exists
	var existingUser models.User
	if err := s.db.Where("email = ?", strings.ToLower(req.Email)).First(&existingUser).Error; err == nil {
		return nil, ErrEmailExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing email: %w", err)
	}

	// Validate password strength
	if err := s.passwordService.ValidatePasswordStrength(req.Password); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrPasswordTooWeak, err.Error())
	}

	// Hash password
	hashedPassword, err := s.passwordService.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Email:        strings.ToLower(req.Email),
		PasswordHash: hashedPassword,
		Name:         strings.TrimSpace(req.Name),
		IsActive:     true,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// AuthenticateUser authenticates a user with email and password
func (s *UserService) AuthenticateUser(req *dto.LoginRequest) (*models.User, error) {
	// Validate input
	if req.Email == "" || req.Password == "" {
		return nil, ErrInvalidPassword
	}

	if !s.isValidEmail(req.Email) {
		return nil, ErrInvalidEmail
	}

	// Find user by email
	var user models.User
	if err := s.db.Where("email = ?", strings.ToLower(req.Email)).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidPassword // Don't reveal that email doesn't exist
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Verify password
	if err := s.passwordService.ComparePassword(user.PasswordHash, req.Password); err != nil {
		return nil, ErrInvalidPassword
	}

	return &user, nil
}

// GetUserByID retrieves a user by their ID
func (s *UserService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by their email
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	if !s.isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	var user models.User
	if err := s.db.Where("email = ? AND is_active = ?", strings.ToLower(email), true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// UpdateUserProfile updates user profile information
func (s *UserService) UpdateUserProfile(userID uuid.UUID, req *dto.UpdateProfileRequest) (*models.User, error) {
	// Get existing user
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Validate and update fields
	updates := make(map[string]interface{})

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if err := s.validateName(name); err != nil {
			return nil, err
		}
		updates["name"] = name
	}

	if req.AvatarURL != nil {
		avatarURL := strings.TrimSpace(*req.AvatarURL)
		if avatarURL == "" {
			updates["avatar_url"] = nil
		} else {
			if err := s.validateAvatarURL(avatarURL); err != nil {
				return nil, err
			}
			updates["avatar_url"] = avatarURL
		}
	}

	// Update user if there are changes
	if len(updates) > 0 {
		if err := s.db.Model(user).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update user profile: %w", err)
		}

		// Reload user to get updated data
		if err := s.db.Where("id = ?", userID).First(user).Error; err != nil {
			return nil, fmt.Errorf("failed to reload updated user: %w", err)
		}
	}

	return user, nil
}

// ChangePassword changes user's password
func (s *UserService) ChangePassword(userID uuid.UUID, req *dto.ChangePasswordRequest) error {
	// Get user
	user, err := s.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Verify current password
	if err := s.passwordService.ComparePassword(user.PasswordHash, req.CurrentPassword); err != nil {
		return ErrInvalidPassword
	}

	// Validate new password strength
	if err := s.passwordService.ValidatePasswordStrength(req.NewPassword); err != nil {
		return fmt.Errorf("%w: %s", ErrPasswordTooWeak, err.Error())
	}

	// Hash new password
	hashedPassword, err := s.passwordService.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	if err := s.db.Model(user).Update("password_hash", hashedPassword).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// DeactivateUser deactivates a user account
func (s *UserService) DeactivateUser(userID uuid.UUID) error {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return err
	}

	if err := s.db.Model(user).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	return nil
}

// validateRegistrationRequest validates user registration input
func (s *UserService) validateRegistrationRequest(req *dto.RegisterRequest) error {
	if req.Email == "" {
		return ErrInvalidEmail
	}

	if !s.isValidEmail(req.Email) {
		return ErrInvalidEmail
	}

	if err := s.validateName(req.Name); err != nil {
		return err
	}

	if req.Password == "" {
		return ErrInvalidPassword
	}

	return nil
}

// isValidEmail validates email format
func (s *UserService) isValidEmail(email string) bool {
	if len(email) > 255 {
		return false
	}
	return emailRegex.MatchString(strings.ToLower(email))
}

// validateName validates user name
func (s *UserService) validateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrInvalidName
	}
	if len(name) > 255 {
		return fmt.Errorf("%w: name too long", ErrInvalidName)
	}
	return nil
}

// validateAvatarURL validates avatar URL
func (s *UserService) validateAvatarURL(url string) error {
	if len(url) > 500 {
		return errors.New("avatar URL too long")
	}
	// Basic URL validation - in production, you might want more sophisticated validation
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return errors.New("avatar URL must be a valid HTTP or HTTPS URL")
	}
	return nil
}