package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"minisentry/internal/dto"
	"minisentry/internal/middleware"
	"minisentry/internal/services"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	userService *services.UserService
	jwtService  *services.JWTService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *services.UserService, jwtService *services.JWTService) *UserHandler {
	return &UserHandler{
		userService: userService,
		jwtService:  jwtService,
	}
}

// RegisterRoutes registers all user-related routes
func (h *UserHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware) {
	// Public routes (no authentication required)
	r.Post("/auth/register", h.Register)
	r.Post("/auth/login", h.Login)
	r.Post("/auth/refresh", h.RefreshToken)

	// Protected routes (authentication required)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequireAuth)
		r.Post("/auth/logout", h.Logout)
		r.Get("/auth/profile", h.GetProfile)
		r.Put("/auth/profile", h.UpdateProfile)
		r.Put("/auth/password", h.ChangePassword)
	})
}

// Register handles user registration
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err)
		return
	}

	// Create user
	user, err := h.userService.CreateUser(&req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEmailExists):
			h.writeErrorResponse(w, http.StatusConflict, "Email already exists", err)
		case errors.Is(err, services.ErrInvalidEmail):
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid email format", err)
		case errors.Is(err, services.ErrInvalidName):
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid name", err)
		case errors.Is(err, services.ErrPasswordTooWeak):
			h.writeErrorResponse(w, http.StatusBadRequest, "Password does not meet requirements", err)
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create user", err)
		}
		return
	}

	// Generate JWT tokens
	tokens, err := h.jwtService.GenerateTokens(user.ID, user.Email, user.Name)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to generate authentication tokens", err)
		return
	}

	// Prepare response
	var userResponse dto.UserResponse
	userResponse.ConvertFromModel(user.ToResponse())
	
	response := dto.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
		User:         userResponse,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Login handles user authentication
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err)
		return
	}

	// Authenticate user
	user, err := h.userService.AuthenticateUser(&req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidPassword):
			h.writeErrorResponse(w, http.StatusUnauthorized, "Invalid email or password", nil)
		case errors.Is(err, services.ErrUserInactive):
			h.writeErrorResponse(w, http.StatusUnauthorized, "Account is inactive", nil)
		case errors.Is(err, services.ErrInvalidEmail):
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid email format", err)
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "Authentication failed", err)
		}
		return
	}

	// Generate JWT tokens
	tokens, err := h.jwtService.GenerateTokens(user.ID, user.Email, user.Name)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to generate authentication tokens", err)
		return
	}

	// Prepare response
	var userResponse dto.UserResponse
	userResponse.ConvertFromModel(user.ToResponse())
	
	response := dto.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
		User:         userResponse,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// RefreshToken handles JWT token refresh
func (h *UserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err)
		return
	}

	// Refresh tokens
	tokens, err := h.jwtService.RefreshToken(req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidToken), errors.Is(err, services.ErrTokenExpired):
			h.writeErrorResponse(w, http.StatusUnauthorized, "Invalid or expired refresh token", nil)
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to refresh token", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokens)
}

// Logout handles user logout
func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Since JWTs are stateless, logout is handled client-side by discarding tokens
	// In a production environment, you might want to maintain a blacklist of tokens
	response := dto.SuccessResponse{
		Success: true,
		Message: "Successfully logged out",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetProfile retrieves the current user's profile
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by auth middleware)
	userClaims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user from context", nil)
		return
	}

	// Get user ID
	userID := userClaims.ID

	// Get user details
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			h.writeErrorResponse(w, http.StatusNotFound, "User not found", nil)
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user profile", err)
		}
		return
	}

	var userResponse dto.UserResponse
	userResponse.ConvertFromModel(user.ToResponse())
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userResponse)
}

// UpdateProfile updates the current user's profile
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userClaims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user from context", nil)
		return
	}

	// Get user ID
	userID := userClaims.ID

	var req dto.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err)
		return
	}

	// Update user profile
	user, err := h.userService.UpdateUserProfile(userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			h.writeErrorResponse(w, http.StatusNotFound, "User not found", nil)
		case errors.Is(err, services.ErrInvalidName):
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid name", err)
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update profile", err)
		}
		return
	}

	var userResponse dto.UserResponse
	userResponse.ConvertFromModel(user.ToResponse())
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userResponse)
}

// ChangePassword changes the current user's password
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userClaims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user from context", nil)
		return
	}

	// Get user ID
	userID := userClaims.ID

	var req dto.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err)
		return
	}

	// Change password
	err := h.userService.ChangePassword(userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			h.writeErrorResponse(w, http.StatusNotFound, "User not found", nil)
		case errors.Is(err, services.ErrInvalidPassword):
			h.writeErrorResponse(w, http.StatusUnauthorized, "Current password is incorrect", nil)
		case errors.Is(err, services.ErrPasswordTooWeak):
			h.writeErrorResponse(w, http.StatusBadRequest, "New password does not meet requirements", err)
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to change password", err)
		}
		return
	}

	response := dto.SuccessResponse{
		Success: true,
		Message: "Password changed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// writeErrorResponse writes a standardized error response
func (h *UserHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	response := dto.ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}

	// Add error details in development mode (in production, be careful about exposing internal errors)
	if err != nil {
		response.Details = map[string]interface{}{
			"error": err.Error(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}