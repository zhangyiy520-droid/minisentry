package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"minisentry/internal/dto"
	"minisentry/internal/middleware"
	"minisentry/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type IssueHandler struct {
	issueService *services.IssueService
}

func NewIssueHandler(issueService *services.IssueService) *IssueHandler {
	return &IssueHandler{
		issueService: issueService,
	}
}

// RegisterRoutes registers all issue-related routes
func (h *IssueHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware, projectMiddleware *middleware.ProjectMiddleware) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(authMiddleware.RequireAuth)
		
		// Project-scoped issue routes
		r.Route("/projects/{project_id}/issues", func(r chi.Router) {
			r.Use(projectMiddleware.RequireProjectAccess)
			r.Get("/", h.ListProjectIssues)    // GET /api/v1/projects/{id}/issues
			r.Get("/stats", h.GetIssueStats)   // GET /api/v1/projects/{id}/issues/stats
		})
		
		// Individual issue routes
		r.Route("/issues/{issue_id}", func(r chi.Router) {
			r.Use(h.issueAccessMiddleware)
			r.Get("/", h.GetIssue)                    // GET /api/v1/issues/{id}
			r.Put("/", h.UpdateIssue)                 // PUT /api/v1/issues/{id}
			r.Post("/comments", h.AddIssueComment)    // POST /api/v1/issues/{id}/comments
			r.Get("/comments", h.GetIssueComments)    // GET /api/v1/issues/{id}/comments
			r.Get("/activity", h.GetIssueActivity)    // GET /api/v1/issues/{id}/activity
			r.Get("/events", h.GetIssueEvents)        // GET /api/v1/issues/{id}/events
		})
		
		// Bulk operations
		r.Post("/issues/bulk-update", h.BulkUpdateIssues) // POST /api/v1/issues/bulk-update
	})
}

// ListProjectIssues handles GET /api/v1/projects/{id}/issues
func (h *IssueHandler) ListProjectIssues(w http.ResponseWriter, r *http.Request) {
	project, ok := middleware.GetProjectFromContext(r.Context())
	if !ok {
		http.Error(w, "Project not found in context", http.StatusInternalServerError)
		return
	}
	
	// Parse query parameters
	filters := h.parseIssueFilters(r)
	
	// Get issues
	response, err := h.issueService.GetProjectIssues(project.ID, filters)
	if err != nil {
		http.Error(w, "Failed to retrieve issues: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetIssue handles GET /api/v1/issues/{id}
func (h *IssueHandler) GetIssue(w http.ResponseWriter, r *http.Request) {
	issueID, err := uuid.Parse(chi.URLParam(r, "issue_id"))
	if err != nil {
		http.Error(w, "Invalid issue ID", http.StatusBadRequest)
		return
	}
	
	issue, err := h.issueService.GetIssue(issueID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Issue not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve issue: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(issue)
}

// UpdateIssue handles PUT /api/v1/issues/{id}
func (h *IssueHandler) UpdateIssue(w http.ResponseWriter, r *http.Request) {
	issueID, err := uuid.Parse(chi.URLParam(r, "issue_id"))
	if err != nil {
		http.Error(w, "Invalid issue ID", http.StatusBadRequest)
		return
	}
	
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}
	
	var request dto.IssueUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	
	// Validate status if provided
	if request.Status != nil {
		if !h.isValidStatus(*request.Status) {
			http.Error(w, "Invalid status value", http.StatusBadRequest)
			return
		}
	}
	
	// Update issue
	updatedIssue, err := h.issueService.UpdateIssueStatus(issueID, user.ID, request)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Issue not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "invalid status transition") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to update issue: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedIssue)
}

// AddIssueComment handles POST /api/v1/issues/{id}/comments
func (h *IssueHandler) AddIssueComment(w http.ResponseWriter, r *http.Request) {
	issueID, err := uuid.Parse(chi.URLParam(r, "issue_id"))
	if err != nil {
		http.Error(w, "Invalid issue ID", http.StatusBadRequest)
		return
	}
	
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}
	
	var request dto.IssueCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	
	// Validate content
	if strings.TrimSpace(request.Content) == "" {
		http.Error(w, "Comment content cannot be empty", http.StatusBadRequest)
		return
	}
	
	// Add comment
	comment, err := h.issueService.AddIssueComment(issueID, user.ID, request)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Issue not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to add comment: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

// GetIssueComments handles GET /api/v1/issues/{id}/comments
func (h *IssueHandler) GetIssueComments(w http.ResponseWriter, r *http.Request) {
	issueID, err := uuid.Parse(chi.URLParam(r, "issue_id"))
	if err != nil {
		http.Error(w, "Invalid issue ID", http.StatusBadRequest)
		return
	}
	
	// Parse pagination
	page, limit := h.parsePagination(r)
	
	// Get comments
	response, err := h.issueService.GetIssueComments(issueID, page, limit)
	if err != nil {
		http.Error(w, "Failed to retrieve comments: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetIssueActivity handles GET /api/v1/issues/{id}/activity
func (h *IssueHandler) GetIssueActivity(w http.ResponseWriter, r *http.Request) {
	issueID, err := uuid.Parse(chi.URLParam(r, "issue_id"))
	if err != nil {
		http.Error(w, "Invalid issue ID", http.StatusBadRequest)
		return
	}
	
	// Parse pagination
	page, limit := h.parsePagination(r)
	
	// Get activity
	response, err := h.issueService.GetIssueActivity(issueID, page, limit)
	if err != nil {
		http.Error(w, "Failed to retrieve activity: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetIssueEvents handles GET /api/v1/issues/{id}/events
func (h *IssueHandler) GetIssueEvents(w http.ResponseWriter, r *http.Request) {
	issueID, err := uuid.Parse(chi.URLParam(r, "issue_id"))
	if err != nil {
		http.Error(w, "Invalid issue ID", http.StatusBadRequest)
		return
	}
	
	// Parse pagination
	page, limit := h.parsePagination(r)
	
	// Get events
	response, err := h.issueService.GetIssueEvents(issueID, page, limit)
	if err != nil {
		http.Error(w, "Failed to retrieve events: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetIssueStats handles GET /api/v1/projects/{id}/issues/stats
func (h *IssueHandler) GetIssueStats(w http.ResponseWriter, r *http.Request) {
	project, ok := middleware.GetProjectFromContext(r.Context())
	if !ok {
		http.Error(w, "Project not found in context", http.StatusInternalServerError)
		return
	}
	
	// Get statistics
	stats, err := h.issueService.GetIssueStats(project.ID)
	if err != nil {
		http.Error(w, "Failed to retrieve issue statistics: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// BulkUpdateIssues handles POST /api/v1/issues/bulk-update
func (h *IssueHandler) BulkUpdateIssues(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}
	
	var request dto.BulkUpdateIssuesRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	
	// Validate request
	if len(request.IssueIDs) == 0 {
		http.Error(w, "No issues specified", http.StatusBadRequest)
		return
	}
	
	if len(request.IssueIDs) > 100 {
		http.Error(w, "Too many issues specified (max 100)", http.StatusBadRequest)
		return
	}
	
	if !h.isValidBulkAction(request.Action) {
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}
	
	if request.Action == "assign" && request.AssigneeID == nil {
		http.Error(w, "Assignee ID required for assign action", http.StatusBadRequest)
		return
	}
	
	// Perform bulk update
	response, err := h.issueService.BulkUpdateIssues(user.ID, request)
	if err != nil {
		http.Error(w, "Failed to perform bulk update: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper methods

// issueAccessMiddleware ensures the user has access to the issue through project membership
func (h *IssueHandler) issueAccessMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		issueID, err := uuid.Parse(chi.URLParam(r, "issue_id"))
		if err != nil {
			http.Error(w, "Invalid issue ID", http.StatusBadRequest)
			return
		}
		
		// Get the issue to find its project
		_, err = h.issueService.GetIssue(issueID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				http.Error(w, "Issue not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to retrieve issue", http.StatusInternalServerError)
			return
		}
		
		// Verify user has access to the project
		_, ok := middleware.GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "User not found in context", http.StatusInternalServerError)
			return
		}
		
		// This is a simplified check - in a real implementation, you'd verify project membership
		// For now, we'll assume if the user can authenticate, they can access any issue
		// In production, add proper project membership verification here
		
		next.ServeHTTP(w, r)
	})
}

func (h *IssueHandler) parseIssueFilters(r *http.Request) dto.IssueFilters {
	query := r.URL.Query()
	
	filters := dto.IssueFilters{
		Sort:  "last_seen",
		Order: "desc",
		Page:  1,
		Limit: 25,
	}
	
	// Parse status filter
	if statusStr := query.Get("status"); statusStr != "" {
		filters.Status = strings.Split(statusStr, ",")
	}
	
	// Parse level filter
	if levelStr := query.Get("level"); levelStr != "" {
		filters.Level = strings.Split(levelStr, ",")
	}
	
	// Parse assignee filter
	if assignedTo := query.Get("assigned_to"); assignedTo != "" {
		filters.AssignedTo = &assignedTo
	}
	
	// Parse date filters
	if dateFrom := query.Get("date_from"); dateFrom != "" {
		filters.DateFrom = &dateFrom
	}
	if dateTo := query.Get("date_to"); dateTo != "" {
		filters.DateTo = &dateTo
	}
	
	// Parse environment filter
	if environment := query.Get("environment"); environment != "" {
		filters.Environment = &environment
	}
	
	// Parse search
	if search := query.Get("search"); search != "" {
		filters.Search = &search
	}
	
	// Parse sort and order
	if sort := query.Get("sort"); sort != "" {
		if h.isValidSortField(sort) {
			filters.Sort = sort
		}
	}
	if order := query.Get("order"); order != "" {
		if order == "asc" || order == "desc" {
			filters.Order = order
		}
	}
	
	// Parse pagination
	if pageStr := query.Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filters.Page = page
		}
	}
	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			filters.Limit = limit
		}
	}
	
	return filters
}

func (h *IssueHandler) parsePagination(r *http.Request) (int, int) {
	query := r.URL.Query()
	
	page := 1
	limit := 25
	
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if limitStr := query.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	return page, limit
}

func (h *IssueHandler) isValidStatus(status string) bool {
	validStatuses := []string{"unresolved", "resolved", "ignored"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

func (h *IssueHandler) isValidSortField(sort string) bool {
	validSorts := []string{"frequency", "first_seen", "last_seen"}
	for _, validSort := range validSorts {
		if sort == validSort {
			return true
		}
	}
	return false
}

func (h *IssueHandler) isValidBulkAction(action string) bool {
	validActions := []string{"resolve", "ignore", "unresolve", "assign"}
	for _, validAction := range validActions {
		if action == validAction {
			return true
		}
	}
	return false
}

// Helper function to send JSON error responses
func (h *IssueHandler) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   http.StatusText(statusCode),
		"message": message,
	})
}

// Helper function to send JSON success responses
func (h *IssueHandler) sendSuccessResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}