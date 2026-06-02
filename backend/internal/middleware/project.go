package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"minisentry/internal/models"
	"minisentry/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type projectContextKey string

const (
	ProjectContextKey     projectContextKey = "project"
	ProjectRoleContextKey projectContextKey = "project_role"
)

type ProjectMiddleware struct {
	projectService *services.ProjectService
}

// ProjectContext holds project data in request context
type ProjectContext struct {
	ID             uuid.UUID                `json:"id"`
	OrganizationID uuid.UUID                `json:"organization_id"`
	Name           string                   `json:"name"`
	Slug           string                   `json:"slug"`
	Platform       string                   `json:"platform"`
	DSN            string                   `json:"dsn"`
	PublicKey      string                   `json:"public_key"`
	IsActive       bool                     `json:"is_active"`
	Role           models.OrganizationRole  `json:"role"` // User's role in the organization
}

func NewProjectMiddleware(projectService *services.ProjectService) *ProjectMiddleware {
	return &ProjectMiddleware{
		projectService: projectService,
	}
}

// RequireProjectAccess middleware checks if user has access to project
func (pm *ProjectMiddleware) RequireProjectAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user from context (auth middleware should run first)
		user, ok := GetUserFromContext(r.Context())
		if !ok {
			pm.writeErrorResponse(w, http.StatusUnauthorized, "authentication required")
			return
		}

		// Get project ID from URL
		projectIDStr := chi.URLParam(r, "id")
		if projectIDStr == "" {
			pm.writeErrorResponse(w, http.StatusBadRequest, "project ID required")
			return
		}

		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			pm.writeErrorResponse(w, http.StatusBadRequest, "invalid project ID format")
			return
		}

		// Check if user has access to project and get their role
		role, err := pm.projectService.ValidateProjectAccess(user.ID, projectID)
		if err != nil {
			switch err {
			case services.ErrProjectAccessDenied:
				pm.writeErrorResponse(w, http.StatusForbidden, "access denied to project")
			case services.ErrProjectNotFound:
				pm.writeErrorResponse(w, http.StatusNotFound, "project not found")
			default:
				pm.writeErrorResponse(w, http.StatusInternalServerError, "failed to check project access")
			}
			return
		}

		// Get project details
		project, err := pm.projectService.GetProject(user.ID, projectID)
		if err != nil {
			switch err {
			case services.ErrProjectNotFound:
				pm.writeErrorResponse(w, http.StatusNotFound, "project not found")
			case services.ErrProjectAccessDenied:
				pm.writeErrorResponse(w, http.StatusForbidden, "access denied to project")
			default:
				pm.writeErrorResponse(w, http.StatusInternalServerError, "failed to get project")
			}
			return
		}

		// Add project and role to context
		projectCtx := &ProjectContext{
			ID:             project.ID,
			OrganizationID: project.OrganizationID,
			Name:           project.Name,
			Slug:           project.Slug,
			Platform:       project.Platform,
			DSN:            project.DSN,
			PublicKey:      project.PublicKey,
			IsActive:       project.IsActive,
			Role:           role,
		}

		ctx := context.WithValue(r.Context(), ProjectContextKey, projectCtx)
		ctx = context.WithValue(ctx, ProjectRoleContextKey, role)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// RequireProjectRole middleware checks if user has specific role in project's organization
func (pm *ProjectMiddleware) RequireProjectRole(requiredRoles ...models.OrganizationRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get role from context (RequireProjectAccess should run first)
			role, ok := GetProjectRoleFromContext(r.Context())
			if !ok {
				pm.writeErrorResponse(w, http.StatusInternalServerError, "project role not found in context")
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
				pm.writeErrorResponse(w, http.StatusForbidden, "insufficient permissions for project")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireProjectOwnerOrAdmin middleware checks if user is owner or admin of project's organization
func (pm *ProjectMiddleware) RequireProjectOwnerOrAdmin(next http.Handler) http.Handler {
	return pm.RequireProjectRole(models.RoleOwner, models.RoleAdmin)(next)
}

// RequireProjectOwner middleware checks if user is owner of project's organization
func (pm *ProjectMiddleware) RequireProjectOwner(next http.Handler) http.Handler {
	return pm.RequireProjectRole(models.RoleOwner)(next)
}

// RequireActiveProject middleware checks if project is active
func (pm *ProjectMiddleware) RequireActiveProject(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get project from context (RequireProjectAccess should run first)
		project, ok := GetProjectFromContext(r.Context())
		if !ok {
			pm.writeErrorResponse(w, http.StatusInternalServerError, "project not found in context")
			return
		}

		if !project.IsActive {
			pm.writeErrorResponse(w, http.StatusForbidden, "project is not active")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// DSNAuth middleware for authenticating requests using DSN (for error ingestion)
func (pm *ProjectMiddleware) DSNAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dsn := pm.extractDSNFromRequest(r)
		if dsn == "" {
			pm.writeErrorResponse(w, http.StatusUnauthorized, "DSN authentication required")
			return
		}

		// Get project by DSN
		project, err := pm.projectService.GetProjectByDSN(dsn)
		if err != nil {
			switch err {
			case services.ErrProjectNotFound:
				pm.writeErrorResponse(w, http.StatusUnauthorized, "invalid DSN")
			case services.ErrProjectInactive:
				pm.writeErrorResponse(w, http.StatusForbidden, "project is inactive")
			default:
				pm.writeErrorResponse(w, http.StatusInternalServerError, "failed to authenticate DSN")
			}
			return
		}

		// Add project to context
		projectCtx := &ProjectContext{
			ID:             project.ID,
			OrganizationID: project.OrganizationID,
			Name:           project.Name,
			Slug:           project.Slug,
			Platform:       project.Platform,
			DSN:            project.DSN,
			PublicKey:      project.PublicKey,
			IsActive:       project.IsActive,
			Role:           "", // No role for DSN auth
		}

		ctx := context.WithValue(r.Context(), ProjectContextKey, projectCtx)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// extractDSNFromRequest extracts DSN from various sources in the request
func (pm *ProjectMiddleware) extractDSNFromRequest(r *http.Request) string {
	// 1. Check X-Sentry-Auth header (Sentry SDK format)
	sentryAuth := r.Header.Get("X-Sentry-Auth")
	if sentryAuth != "" {
		// Parse Sentry auth header format:
		// Sentry sentry_version=7, sentry_client=sentry-javascript/6.0.0, sentry_key=PUBLIC_KEY, sentry_secret=SECRET_KEY
		return pm.parseSentryAuthHeader(sentryAuth)
	}

	// 2. Check Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Support Bearer token format or DSN format
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			// Check if it's a DSN format
			if strings.Contains(token, "@") && strings.Contains(token, "://") {
				return token
			}
		} else if strings.Contains(authHeader, "@") && strings.Contains(authHeader, "://") {
			return authHeader
		}
	}

	// 3. Check query parameter
	if dsn := r.URL.Query().Get("dsn"); dsn != "" {
		return dsn
	}

	// 4. Check sentry_key and construct DSN (for compatibility)
	if sentryKey := r.URL.Query().Get("sentry_key"); sentryKey != "" {
		// Try to construct DSN from key and project ID in URL
		if projectID := chi.URLParam(r, "project_id"); projectID != "" {
			host := r.Host
			if host == "" {
				host = "localhost:8080" // Fallback
			}
			return fmt.Sprintf("https://%s@%s/%s", sentryKey, host, projectID)
		}
	}

	return ""
}

// parseSentryAuthHeader parses Sentry's X-Sentry-Auth header format
func (pm *ProjectMiddleware) parseSentryAuthHeader(authHeader string) string {
	// Format: Sentry sentry_version=7, sentry_client=..., sentry_key=PUBLIC_KEY, sentry_secret=SECRET_KEY
	if !strings.HasPrefix(authHeader, "Sentry ") {
		return ""
	}

	// Parse key-value pairs
	authData := strings.TrimPrefix(authHeader, "Sentry ")
	pairs := strings.Split(authData, ",")
	
	var sentryKey string
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if strings.HasPrefix(pair, "sentry_key=") {
			sentryKey = strings.TrimPrefix(pair, "sentry_key=")
		}
		// Note: We don't use sentry_secret for DSN construction in this simplified version
	}

	if sentryKey == "" {
		return ""
	}

	// We need to construct the DSN. For now, we'll use the request host
	// In a real implementation, you might want to store the full DSN pattern
	// For this MVP, we'll try to extract project ID from the URL path
	return sentryKey // Return just the key for now, the service will match by public key
}

// GetProjectFromContext extracts project from request context
func GetProjectFromContext(ctx context.Context) (*ProjectContext, bool) {
	project, ok := ctx.Value(ProjectContextKey).(*ProjectContext)
	return project, ok
}

// GetProjectRoleFromContext extracts project role from request context
func GetProjectRoleFromContext(ctx context.Context) (models.OrganizationRole, bool) {
	role, ok := ctx.Value(ProjectRoleContextKey).(models.OrganizationRole)
	return role, ok
}

// GetProjectFromContextAsModel extracts project from context and returns as models.Project
func GetProjectFromContextAsModel(ctx context.Context) (*models.Project, bool) {
	projectCtx, ok := GetProjectFromContext(ctx)
	if !ok {
		return nil, false
	}

	project := &models.Project{
		BaseModel: models.BaseModel{
			ID: projectCtx.ID,
		},
		OrganizationID: projectCtx.OrganizationID,
		Name:           projectCtx.Name,
		Slug:           projectCtx.Slug,
		Platform:       projectCtx.Platform,
		DSN:            projectCtx.DSN,
		PublicKey:      projectCtx.PublicKey,
		IsActive:       projectCtx.IsActive,
	}

	return project, true
}

// writeErrorResponse writes a JSON error response
func (pm *ProjectMiddleware) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}