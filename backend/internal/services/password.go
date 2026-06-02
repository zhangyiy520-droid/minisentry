package services

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type PasswordService struct {
	cost int
}

var (
	ErrInvalidPassword = errors.New("invalid password")
	ErrPasswordTooLong = errors.New("password too long")
)

// NewPasswordService creates a new password service with the specified bcrypt cost
func NewPasswordService(cost int) *PasswordService {
	// Ensure cost is within valid range
	if cost < bcrypt.MinCost {
		cost = bcrypt.MinCost
	}
	if cost > bcrypt.MaxCost {
		cost = bcrypt.MaxCost
	}

	return &PasswordService{
		cost: cost,
	}
}

// NewDefaultPasswordService creates a new password service with default cost
func NewDefaultPasswordService() *PasswordService {
	return NewPasswordService(bcrypt.DefaultCost)
}

// HashPassword hashes a password using bcrypt
func (p *PasswordService) HashPassword(password string) (string, error) {
	// Check password length to prevent DoS attacks
	if len(password) > 72 {
		return "", ErrPasswordTooLong
	}

	if password == "" {
		return "", ErrInvalidPassword
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), p.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// ComparePassword compares a password with its hash
func (p *PasswordService) ComparePassword(hashedPassword, password string) error {
	// Check password length to prevent DoS attacks
	if len(password) > 72 {
		return ErrPasswordTooLong
	}

	if password == "" || hashedPassword == "" {
		return ErrInvalidPassword
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidPassword
		}
		return fmt.Errorf("failed to compare password: %w", err)
	}

	return nil
}

// ValidatePasswordStrength validates password strength requirements
func (p *PasswordService) ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	if len(password) > 72 {
		return ErrPasswordTooLong
	}

	// Check for at least one uppercase letter
	hasUpper := false
	// Check for at least one lowercase letter
	hasLower := false
	// Check for at least one digit
	hasDigit := false
	// Check for at least one special character
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= 32 && char <= 126: // Printable ASCII range
			// Check if it's a special character (not alphanumeric)
			if !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == ' ') {
				hasSpecial = true
			}
		}
	}

	var missingRequirements []string
	if !hasUpper {
		missingRequirements = append(missingRequirements, "uppercase letter")
	}
	if !hasLower {
		missingRequirements = append(missingRequirements, "lowercase letter")
	}
	if !hasDigit {
		missingRequirements = append(missingRequirements, "digit")
	}
	if !hasSpecial {
		missingRequirements = append(missingRequirements, "special character")
	}

	if len(missingRequirements) > 0 {
		return fmt.Errorf("password must contain at least one: %v", missingRequirements)
	}

	return nil
}

// GetCost returns the bcrypt cost being used
func (p *PasswordService) GetCost() int {
	return p.cost
}

// NeedsRehash checks if a password hash needs to be rehashed with a new cost
func (p *PasswordService) NeedsRehash(hashedPassword string) bool {
	cost, err := bcrypt.Cost([]byte(hashedPassword))
	if err != nil {
		return true // If we can't determine the cost, assume it needs rehashing
	}
	return cost != p.cost
}