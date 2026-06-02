package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"minisentry/internal/dto"
	"minisentry/internal/middleware"
	"minisentry/internal/models"
	"minisentry/internal/services"

	"github.com/go-chi/chi/v5"
)

// Validation errors
var (
	ErrEmptyName           = errors.New("name cannot be empty")
	ErrEmptySlug           = errors.New("slug cannot be empty")
	ErrEmptyEmail          = errors.New("email cannot be empty")
	ErrEmptyRole           = errors.New("role cannot be empty")
	ErrNameTooLong         = errors.New("name is too long (max 255 characters)")
	ErrSlugTooLong         = errors.New("slug is too long (max 100 characters)")
	ErrDescriptionTooLong  = errors.New("description is too long (max 1000 characters)")
	ErrInvalidRole         = errors.New("invalid role")
)

type OrganizationHandler struct {
	orgService *services.OrganizationService
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(orgService *services.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{
		orgService: orgService,
	}
}

// RegisterRoutes registers organization routes
func (h *OrganizationHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware, orgMiddleware *middleware.OrganizationMiddleware) {
	r.Route("/organizations", func(r chi.Router) {
		// Require authentication for all organization routes
		r.Use(authMiddleware.RequireAuth)

		// Organization CRUD
		r.Post("/", h.CreateOrganization)
		r.Get("/", h.ListUserOrganizations)

		r.Route("/{id}", func(r chi.Router) {
			// Require organization access for specific organization routes
			r.Use(orgMiddleware.RequireOrganizationAccess)

			r.Get("/", h.GetOrganization)
			r.Put("/", h.UpdateOrganization)
			r.Delete("/", h.DeleteOrganization)

			// Organization members
			r.Route("/members", func(r chi.Router) {
				r.Get("/", h.GetOrganizationMembers)
				r.Post("/", h.AddMember)

				r.Route("/{user_id}", func(r chi.Router) {
					// Require member access for specific member routes
					r.Use(orgMiddleware.RequireMemberAccess)

					r.Put("/", h.UpdateMemberRole)
					r.Delete("/", h.RemoveMember)
				})
			})
		})
	})
}

// CreateOrganization creates a new organization
func (h *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request body
	var req dto.CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if err := h.validateCreateOrganizationRequest(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create organization
	org, err := h.orgService.CreateOrganization(user.ID, req.Name, req.Slug, req.Description)
	if err != nil {
		switch err {
		case services.ErrOrganizationSlugExists:
			h.writeErrorResponse(w, http.StatusConflict, "organization slug already exists")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to create organization")
		}
		return
	}

	// Return response
	response := dto.ToOrganizationResponse(org, models.RoleOwner)
	h.writeJSONResponse(w, http.StatusCreated, response)
}

// ListUserOrganizations lists organizations user belongs to
func (h *OrganizationHandler) ListUserOrganizations(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get user organizations
	orgs, err := h.orgService.GetUserOrganizations(user.ID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get organizations")
		return
	}

	// Convert to response format
	var orgResponses []dto.OrganizationResponse
	for _, org := range orgs {
		orgResponses = append(orgResponses, dto.ToOrganizationResponse(&org.Organization, org.Role))
	}

	response := dto.OrganizationListResponse{
		Organizations: orgResponses,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetOrganization gets organization details
func (h *OrganizationHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	// Get organization from context (middleware handles access control)
	orgCtx, ok := middleware.GetOrganizationFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusInternalServerError, "organization not found in context")
		return
	}

	// Convert to response format
	org := &models.Organization{
		BaseModel: models.BaseModel{
			ID: orgCtx.ID,
		},
		Name: orgCtx.Name,
		Slug: orgCtx.Slug,
	}

	response := dto.ToOrganizationResponse(org, orgCtx.Role)
	h.writeJSONResponse(w, http.StatusOK, response)
}

// UpdateOrganization updates organization details
func (h *OrganizationHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	// Get user and organization from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	orgCtx, ok := middleware.GetOrganizationFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusInternalServerError, "organization not found in context")
		return
	}

	// Check permissions (owner or admin)
	if orgCtx.Role != models.RoleOwner && orgCtx.Role != models.RoleAdmin {
		h.writeErrorResponse(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	// Parse request body
	var req dto.UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if err := h.validateUpdateOrganizationRequest(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update organization
	org, err := h.orgService.UpdateOrganization(user.ID, orgCtx.ID, req.Name, req.Description)
	if err != nil {
		switch err {
		case services.ErrInsufficientPermissions:
			h.writeErrorResponse(w, http.StatusForbidden, "insufficient permissions")
		case services.ErrOrganizationNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "organization not found")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to update organization")
		}
		return
	}

	// Return response
	response := dto.ToOrganizationResponse(org, orgCtx.Role)
	h.writeJSONResponse(w, http.StatusOK, response)
}

// DeleteOrganization deletes organization (owner only)
func (h *OrganizationHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	// Get user and organization from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	orgCtx, ok := middleware.GetOrganizationFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusInternalServerError, "organization not found in context")
		return
	}

	// Check permissions (owner only)
	if orgCtx.Role != models.RoleOwner {
		h.writeErrorResponse(w, http.StatusForbidden, "only organization owners can delete organizations")
		return
	}

	// Delete organization
	if err := h.orgService.DeleteOrganization(user.ID, orgCtx.ID); err != nil {
		switch err {
		case services.ErrInsufficientPermissions:
			h.writeErrorResponse(w, http.StatusForbidden, "insufficient permissions")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to delete organization")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetOrganizationMembers lists organization members
func (h *OrganizationHandler) GetOrganizationMembers(w http.ResponseWriter, r *http.Request) {
	// Get user and organization from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	orgCtx, ok := middleware.GetOrganizationFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusInternalServerError, "organization not found in context")
		return
	}

	// Get organization members
	members, err := h.orgService.GetOrganizationMembers(user.ID, orgCtx.ID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get organization members")
		return
	}

	// Convert to response format
	var memberResponses []dto.OrganizationMemberResponse
	for _, member := range members {
		memberResponses = append(memberResponses, dto.ToOrganizationMemberResponse(&member))
	}

	response := dto.OrganizationMembersResponse{
		Members: memberResponses,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// AddMember adds a member to organization
func (h *OrganizationHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	// Get user and organization from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	orgCtx, ok := middleware.GetOrganizationFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusInternalServerError, "organization not found in context")
		return
	}

	// Check permissions (owner or admin)
	if orgCtx.Role != models.RoleOwner && orgCtx.Role != models.RoleAdmin {
		h.writeErrorResponse(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	// Parse request body
	var req dto.AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if err := h.validateAddMemberRequest(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Add member
	member, err := h.orgService.AddMember(user.ID, orgCtx.ID, req.Email, req.Role)
	if err != nil {
		switch err {
		case services.ErrInsufficientPermissions:
			h.writeErrorResponse(w, http.StatusForbidden, "insufficient permissions")
		case services.ErrOrgUserNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "user not found")
		case services.ErrUserAlreadyMember:
			h.writeErrorResponse(w, http.StatusConflict, "user is already a member")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to add member")
		}
		return
	}

	// Return response
	response := dto.ToOrganizationMemberResponse(member)
	h.writeJSONResponse(w, http.StatusCreated, response)
}

// UpdateMemberRole updates member role
func (h *OrganizationHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	// Get user, organization, and target user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	orgCtx, ok := middleware.GetOrganizationFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusInternalServerError, "organization not found in context")
		return
	}

	targetUserID, ok := middleware.GetTargetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusInternalServerError, "target user ID not found in context")
		return
	}

	// Check permissions (owner only)
	if orgCtx.Role != models.RoleOwner {
		h.writeErrorResponse(w, http.StatusForbidden, "only organization owners can change member roles")
		return
	}

	// Parse request body
	var req dto.UpdateMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if err := h.validateUpdateMemberRoleRequest(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update member role
	member, err := h.orgService.UpdateMemberRole(user.ID, orgCtx.ID, targetUserID, req.Role)
	if err != nil {
		switch err {
		case services.ErrInsufficientPermissions:
			h.writeErrorResponse(w, http.StatusForbidden, "insufficient permissions")
		case services.ErrMemberNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "member not found")
		case services.ErrCannotChangeOwnerRole:
			h.writeErrorResponse(w, http.StatusBadRequest, "cannot change owner role")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to update member role")
		}
		return
	}

	// Return response
	response := dto.ToOrganizationMemberResponse(member)
	h.writeJSONResponse(w, http.StatusOK, response)
}

// RemoveMember removes member from organization
func (h *OrganizationHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	// Get user, organization, and target user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	orgCtx, ok := middleware.GetOrganizationFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusInternalServerError, "organization not found in context")
		return
	}

	targetUserID, ok := middleware.GetTargetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusInternalServerError, "target user ID not found in context")
		return
	}

	// Check permissions (owner or admin)
	if orgCtx.Role != models.RoleOwner && orgCtx.Role != models.RoleAdmin {
		h.writeErrorResponse(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	// Remove member
	if err := h.orgService.RemoveMember(user.ID, orgCtx.ID, targetUserID); err != nil {
		switch err {
		case services.ErrInsufficientPermissions:
			h.writeErrorResponse(w, http.StatusForbidden, "insufficient permissions")
		case services.ErrMemberNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "member not found")
		case services.ErrCannotRemoveOwner:
			h.writeErrorResponse(w, http.StatusBadRequest, "cannot remove organization owner")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to remove member")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Validation methods
func (h *OrganizationHandler) validateCreateOrganizationRequest(req *dto.CreateOrganizationRequest) error {
	if req.Name == "" {
		return ErrEmptyName
	}
	if req.Slug == "" {
		return ErrEmptySlug
	}
	if len(req.Name) > 255 {
		return ErrNameTooLong
	}
	if len(req.Slug) > 100 {
		return ErrSlugTooLong
	}
	if req.Description != nil && len(*req.Description) > 1000 {
		return ErrDescriptionTooLong
	}
	return nil
}

func (h *OrganizationHandler) validateUpdateOrganizationRequest(req *dto.UpdateOrganizationRequest) error {
	if req.Name != nil {
		if *req.Name == "" {
			return ErrEmptyName
		}
		if len(*req.Name) > 255 {
			return ErrNameTooLong
		}
	}
	if req.Description != nil && len(*req.Description) > 1000 {
		return ErrDescriptionTooLong
	}
	return nil
}

func (h *OrganizationHandler) validateAddMemberRequest(req *dto.AddMemberRequest) error {
	if req.Email == "" {
		return ErrEmptyEmail
	}
	if req.Role == "" {
		return ErrEmptyRole
	}
	if req.Role != models.RoleAdmin && req.Role != models.RoleMember {
		return ErrInvalidRole
	}
	return nil
}

func (h *OrganizationHandler) validateUpdateMemberRoleRequest(req *dto.UpdateMemberRoleRequest) error {
	if req.Role == "" {
		return ErrEmptyRole
	}
	if req.Role != models.RoleAdmin && req.Role != models.RoleMember {
		return ErrInvalidRole
	}
	return nil
}

// Helper methods
func (h *OrganizationHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *OrganizationHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := dto.ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}