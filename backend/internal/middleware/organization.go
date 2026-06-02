package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"minisentry/internal/models"
	"minisentry/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type organizationContextKey string

const (
	OrganizationContextKey     organizationContextKey = "organization"
	OrganizationRoleContextKey organizationContextKey = "organization_role"
)

type OrganizationMiddleware struct {
	orgService *services.OrganizationService
}

// OrganizationContext holds organization data in request context
type OrganizationContext struct {
	ID   uuid.UUID             `json:"id"`
	Name string                `json:"name"`
	Slug string                `json:"slug"`
	Role models.OrganizationRole `json:"role"`
}

func NewOrganizationMiddleware(orgService *services.OrganizationService) *OrganizationMiddleware {
	return &OrganizationMiddleware{
		orgService: orgService,
	}
}

// RequireOrganizationAccess middleware checks if user has access to organization
func (om *OrganizationMiddleware) RequireOrganizationAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user from context (auth middleware should run first)
		user, ok := GetUserFromContext(r.Context())
		if !ok {
			om.writeErrorResponse(w, http.StatusUnauthorized, "authentication required")
			return
		}

		// Get organization ID from URL (try both "id" and "org_id" parameters)
		orgIDStr := chi.URLParam(r, "id")
		if orgIDStr == "" {
			orgIDStr = chi.URLParam(r, "org_id")
		}
		if orgIDStr == "" {
			om.writeErrorResponse(w, http.StatusBadRequest, "organization ID required")
			return
		}

		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			om.writeErrorResponse(w, http.StatusBadRequest, "invalid organization ID")
			return
		}

		// Check if user has access to organization
		org, role, err := om.orgService.GetOrganization(user.ID, orgID)
		if err != nil {
			switch err {
			case services.ErrUserNotMember:
				om.writeErrorResponse(w, http.StatusForbidden, "access denied")
			case services.ErrOrganizationNotFound:
				om.writeErrorResponse(w, http.StatusNotFound, "organization not found")
			default:
				om.writeErrorResponse(w, http.StatusInternalServerError, "failed to check organization access")
			}
			return
		}

		// Add organization and role to context
		orgCtx := &OrganizationContext{
			ID:   org.ID,
			Name: org.Name,
			Slug: org.Slug,
			Role: role,
		}

		ctx := context.WithValue(r.Context(), OrganizationContextKey, orgCtx)
		ctx = context.WithValue(ctx, OrganizationRoleContextKey, role)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// RequireOrganizationRole middleware checks if user has specific role in organization
func (om *OrganizationMiddleware) RequireOrganizationRole(requiredRoles ...models.OrganizationRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get role from context (RequireOrganizationAccess should run first)
			role, ok := GetOrganizationRoleFromContext(r.Context())
			if !ok {
				om.writeErrorResponse(w, http.StatusInternalServerError, "organization role not found in context")
				return
			}

			// Check if user has required role
			hasPermission := false
			for _, requiredRole := range requiredRoles {
				if role == requiredRole {
					hasPermission = true
					break
				}
			}

			// Special case: owners have all permissions
			if role == models.RoleOwner {
				hasPermission = true
			}

			if !hasPermission {
				om.writeErrorResponse(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireOwnerOrAdmin middleware checks if user is owner or admin
func (om *OrganizationMiddleware) RequireOwnerOrAdmin(next http.Handler) http.Handler {
	return om.RequireOrganizationRole(models.RoleOwner, models.RoleAdmin)(next)
}

// RequireOwner middleware checks if user is owner
func (om *OrganizationMiddleware) RequireOwner(next http.Handler) http.Handler {
	return om.RequireOrganizationRole(models.RoleOwner)(next)
}

// RequireMemberAccess middleware for organization member routes
func (om *OrganizationMiddleware) RequireMemberAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user from context
		user, ok := GetUserFromContext(r.Context())
		if !ok {
			om.writeErrorResponse(w, http.StatusUnauthorized, "authentication required")
			return
		}

		// Get organization ID from URL
		orgIDStr := chi.URLParam(r, "id")
		if orgIDStr == "" {
			om.writeErrorResponse(w, http.StatusBadRequest, "organization ID required")
			return
		}

		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			om.writeErrorResponse(w, http.StatusBadRequest, "invalid organization ID")
			return
		}

		// Get target user ID from URL
		userIDStr := chi.URLParam(r, "user_id")
		if userIDStr == "" {
			om.writeErrorResponse(w, http.StatusBadRequest, "user ID required")
			return
		}

		targetUserID, err := uuid.Parse(userIDStr)
		if err != nil {
			om.writeErrorResponse(w, http.StatusBadRequest, "invalid user ID")
			return
		}

		// Check if current user has access to organization
		org, role, err := om.orgService.GetOrganization(user.ID, orgID)
		if err != nil {
			switch err {
			case services.ErrUserNotMember:
				om.writeErrorResponse(w, http.StatusForbidden, "access denied")
			case services.ErrOrganizationNotFound:
				om.writeErrorResponse(w, http.StatusNotFound, "organization not found")
			default:
				om.writeErrorResponse(w, http.StatusInternalServerError, "failed to check organization access")
			}
			return
		}

		// Add organization, role, and target user ID to context
		orgCtx := &OrganizationContext{
			ID:   org.ID,
			Name: org.Name,
			Slug: org.Slug,
			Role: role,
		}

		ctx := context.WithValue(r.Context(), OrganizationContextKey, orgCtx)
		ctx = context.WithValue(ctx, OrganizationRoleContextKey, role)
		ctx = context.WithValue(ctx, "target_user_id", targetUserID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// GetOrganizationFromContext extracts organization from request context
func GetOrganizationFromContext(ctx context.Context) (*OrganizationContext, bool) {
	org, ok := ctx.Value(OrganizationContextKey).(*OrganizationContext)
	return org, ok
}

// GetOrganizationRoleFromContext extracts organization role from request context
func GetOrganizationRoleFromContext(ctx context.Context) (models.OrganizationRole, bool) {
	role, ok := ctx.Value(OrganizationRoleContextKey).(models.OrganizationRole)
	return role, ok
}

// GetTargetUserIDFromContext extracts target user ID from request context
func GetTargetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value("target_user_id").(uuid.UUID)
	return userID, ok
}

// writeErrorResponse writes a JSON error response
func (om *OrganizationMiddleware) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}