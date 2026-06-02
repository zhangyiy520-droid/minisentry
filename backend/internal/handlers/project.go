package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"minisentry/internal/dto"
	"minisentry/internal/middleware"
	"minisentry/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Project validation errors
var (
	ErrProjectEmptyName        = errors.New("project name cannot be empty")
	ErrProjectEmptySlug        = errors.New("project slug cannot be empty")
	ErrProjectEmptyPlatform    = errors.New("project platform cannot be empty")
	ErrProjectNameTooLong      = errors.New("project name is too long (max 255 characters)")
	ErrProjectSlugTooLong      = errors.New("project slug is too long (max 100 characters)")
	ErrProjectDescTooLong      = errors.New("project description is too long (max 1000 characters)")
	ErrProjectInvalidPlatform  = errors.New("invalid project platform")
)

type ProjectHandler struct {
	projectService *services.ProjectService
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(projectService *services.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

// RegisterRoutes registers project routes
func (h *ProjectHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware, orgMiddleware *middleware.OrganizationMiddleware, projectMiddleware *middleware.ProjectMiddleware) {
	// Organization project routes
	r.Route("/organizations/{org_id}/projects", func(r chi.Router) {
		r.Use(authMiddleware.RequireAuth)
		r.Use(orgMiddleware.RequireOrganizationAccess)

		r.Post("/", h.CreateProject)
		r.Get("/", h.ListOrganizationProjects)
	})

	// Individual project routes
	r.Route("/projects/{id}", func(r chi.Router) {
		r.Use(authMiddleware.RequireAuth)
		r.Use(projectMiddleware.RequireProjectAccess)

		r.Get("/", h.GetProject)
		r.Put("/", h.UpdateProject)
		r.Delete("/", h.DeleteProject)
		r.Put("/configuration", h.UpdateProjectConfiguration)
		
		r.Route("/keys", func(r chi.Router) {
			r.Post("/regenerate", h.RegenerateProjectKey)
		})
	})
}

// CreateProject creates a new project within an organization
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	// Get organization from context
	org, ok := middleware.GetOrganizationFromContext(r.Context())
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusInternalServerError)
		return
	}

	// Parse request body
	var req dto.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateCreateProjectRequest(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create project
	project, err := h.projectService.CreateProject(user.ID, org.ID, req.Name, req.Slug, req.Platform, req.Description)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrProjectSlugExists):
			http.Error(w, "Project slug already exists in organization", http.StatusConflict)
		case errors.Is(err, services.ErrInsufficientPermissions):
			http.Error(w, "Insufficient permissions to create project", http.StatusForbidden)
		case errors.Is(err, services.ErrProjectInvalidPlatform):
			http.Error(w, "Invalid project platform", http.StatusBadRequest)
		default:
			http.Error(w, "Failed to create project", http.StatusInternalServerError)
		}
		return
	}

	// Return project response
	response := dto.ToProjectResponse(project)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// ListOrganizationProjects lists all projects in an organization
func (h *ProjectHandler) ListOrganizationProjects(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	// Get organization from context
	org, ok := middleware.GetOrganizationFromContext(r.Context())
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusInternalServerError)
		return
	}

	// Get projects
	projects, err := h.projectService.GetOrganizationProjects(user.ID, org.ID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotMember):
			http.Error(w, "Access denied to organization", http.StatusForbidden)
		default:
			http.Error(w, "Failed to get projects", http.StatusInternalServerError)
		}
		return
	}

	// Return projects response
	response := dto.ToProjectListResponse(projects)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetProject gets project details
func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	// Get project from context (injected by middleware)
	project, ok := middleware.GetProjectFromContextAsModel(r.Context())
	if !ok {
		http.Error(w, "Project not found in context", http.StatusInternalServerError)
		return
	}

	// Return project response
	response := dto.ToProjectResponse(project)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateProject updates project details
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	// Get project from context
	project, ok := middleware.GetProjectFromContextAsModel(r.Context())
	if !ok {
		http.Error(w, "Project not found in context", http.StatusInternalServerError)
		return
	}

	// Parse request body
	var req dto.UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateUpdateProjectRequest(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update project
	updatedProject, err := h.projectService.UpdateProject(user.ID, project.ID, req.Name, req.Platform, req.Description)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInsufficientPermissions):
			http.Error(w, "Insufficient permissions to update project", http.StatusForbidden)
		case errors.Is(err, services.ErrProjectInvalidPlatform):
			http.Error(w, "Invalid project platform", http.StatusBadRequest)
		default:
			http.Error(w, "Failed to update project", http.StatusInternalServerError)
		}
		return
	}

	// Return updated project response
	response := dto.ToProjectResponse(updatedProject)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteProject deletes a project
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	// Get project from context
	project, ok := middleware.GetProjectFromContextAsModel(r.Context())
	if !ok {
		http.Error(w, "Project not found in context", http.StatusInternalServerError)
		return
	}

	// Delete project
	if err := h.projectService.DeleteProject(user.ID, project.ID); err != nil {
		switch {
		case errors.Is(err, services.ErrInsufficientPermissions):
			http.Error(w, "Insufficient permissions to delete project", http.StatusForbidden)
		default:
			http.Error(w, "Failed to delete project", http.StatusInternalServerError)
		}
		return
	}

	// Return success response
	w.WriteHeader(http.StatusNoContent)
}

// RegenerateProjectKey regenerates project API key
func (h *ProjectHandler) RegenerateProjectKey(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	// Get project from context
	project, ok := middleware.GetProjectFromContextAsModel(r.Context())
	if !ok {
		http.Error(w, "Project not found in context", http.StatusInternalServerError)
		return
	}

	// Regenerate key
	updatedProject, err := h.projectService.RegenerateProjectKey(user.ID, project.ID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInsufficientPermissions):
			http.Error(w, "Insufficient permissions to regenerate project key", http.StatusForbidden)
		default:
			http.Error(w, "Failed to regenerate project key", http.StatusInternalServerError)
		}
		return
	}

	// Return key response
	response := dto.ProjectKeyResponse{
		PublicKey: updatedProject.PublicKey,
		DSN:       updatedProject.DSN,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateProjectConfiguration updates project configuration
func (h *ProjectHandler) UpdateProjectConfiguration(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	// Get project from context
	project, ok := middleware.GetProjectFromContextAsModel(r.Context())
	if !ok {
		http.Error(w, "Project not found in context", http.StatusInternalServerError)
		return
	}

	// Parse request body
	var req dto.ProjectConfigurationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate platform if provided
	if req.Platform != nil && !dto.IsPlatformSupported(*req.Platform) {
		http.Error(w, "Invalid project platform", http.StatusBadRequest)
		return
	}

	// Update configuration
	updatedProject, err := h.projectService.UpdateProjectConfiguration(user.ID, project.ID, req.IsActive, req.Platform)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInsufficientPermissions):
			http.Error(w, "Insufficient permissions to update project configuration", http.StatusForbidden)
		case errors.Is(err, services.ErrProjectInvalidPlatform):
			http.Error(w, "Invalid project platform", http.StatusBadRequest)
		default:
			http.Error(w, "Failed to update project configuration", http.StatusInternalServerError)
		}
		return
	}

	// Return updated project response
	response := dto.ToProjectResponse(updatedProject)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Validation helpers

func (h *ProjectHandler) validateCreateProjectRequest(req *dto.CreateProjectRequest) error {
	if req.Name == "" {
		return ErrProjectEmptyName
	}
	if len(req.Name) > 255 {
		return ErrProjectNameTooLong
	}

	if req.Slug == "" {
		return ErrProjectEmptySlug
	}
	if len(req.Slug) > 100 {
		return ErrProjectSlugTooLong
	}

	if req.Platform == "" {
		return ErrProjectEmptyPlatform
	}
	if !dto.IsPlatformSupported(req.Platform) {
		return ErrProjectInvalidPlatform
	}

	if req.Description != nil && len(*req.Description) > 1000 {
		return ErrProjectDescTooLong
	}

	return nil
}

func (h *ProjectHandler) validateUpdateProjectRequest(req *dto.UpdateProjectRequest) error {
	if req.Name != nil {
		if *req.Name == "" {
			return ErrProjectEmptyName
		}
		if len(*req.Name) > 255 {
			return ErrProjectNameTooLong
		}
	}

	if req.Platform != nil {
		if *req.Platform == "" {
			return ErrProjectEmptyPlatform
		}
		if !dto.IsPlatformSupported(*req.Platform) {
			return ErrProjectInvalidPlatform
		}
	}

	if req.Description != nil && len(*req.Description) > 1000 {
		return ErrProjectDescTooLong
	}

	return nil
}

// GetProjectIDFromURL extracts project ID from URL parameter
func GetProjectIDFromURL(r *http.Request) (uuid.UUID, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		return uuid.Nil, errors.New("project ID is required")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, errors.New("invalid project ID format")
	}

	return id, nil
}