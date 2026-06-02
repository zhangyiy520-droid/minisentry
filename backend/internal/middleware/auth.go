package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"minisentry/internal/services"

	"github.com/google/uuid"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

type AuthMiddleware struct {
	jwtService *services.JWTService
}

type UserContext struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Name  string    `json:"name"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func NewAuthMiddleware(jwtService *services.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

// RequireAuth is a middleware that validates JWT tokens and injects user context
func (am *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			am.writeErrorResponse(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		// Check if header starts with "Bearer "
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			am.writeErrorResponse(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, bearerPrefix)
		if token == "" {
			am.writeErrorResponse(w, http.StatusUnauthorized, "missing token")
			return
		}

		// Validate token
		claims, err := am.jwtService.ValidateToken(token, "access")
		if err != nil {
			switch err {
			case services.ErrTokenExpired:
				am.writeErrorResponse(w, http.StatusUnauthorized, "token expired")
			case services.ErrInvalidTokenType:
				am.writeErrorResponse(w, http.StatusUnauthorized, "invalid token type")
			default:
				am.writeErrorResponse(w, http.StatusUnauthorized, "invalid token")
			}
			return
		}

		// Parse user ID
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			am.writeErrorResponse(w, http.StatusUnauthorized, "invalid user ID in token")
			return
		}

		// Create user context
		userCtx := &UserContext{
			ID:    userID,
			Email: claims.Email,
			Name:  claims.Name,
		}

		// Add user to request context
		ctx := context.WithValue(r.Context(), UserContextKey, userCtx)
		r = r.WithContext(ctx)

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// OptionalAuth is a middleware that validates JWT tokens if present but doesn't require them
func (am *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No auth header, continue without user context
			next.ServeHTTP(w, r)
			return
		}

		// Check if header starts with "Bearer "
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			// Invalid format, continue without user context
			next.ServeHTTP(w, r)
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, bearerPrefix)
		if token == "" {
			// Empty token, continue without user context
			next.ServeHTTP(w, r)
			return
		}

		// Validate token
		claims, err := am.jwtService.ValidateToken(token, "access")
		if err != nil {
			// Invalid token, continue without user context
			next.ServeHTTP(w, r)
			return
		}

		// Parse user ID
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			// Invalid user ID, continue without user context
			next.ServeHTTP(w, r)
			return
		}

		// Create user context
		userCtx := &UserContext{
			ID:    userID,
			Email: claims.Email,
			Name:  claims.Name,
		}

		// Add user to request context
		ctx := context.WithValue(r.Context(), UserContextKey, userCtx)
		r = r.WithContext(ctx)

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext extracts the user from the request context
func GetUserFromContext(ctx context.Context) (*UserContext, bool) {
	user, ok := ctx.Value(UserContextKey).(*UserContext)
	return user, ok
}

// writeErrorResponse writes a JSON error response
func (am *AuthMiddleware) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}