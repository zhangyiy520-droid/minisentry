package handlers

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"minisentry/internal/dto"
	"minisentry/internal/middleware"
	"minisentry/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ErrorHandler struct {
	errorService *services.ErrorService
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(errorService *services.ErrorService) *ErrorHandler {
	return &ErrorHandler{
		errorService: errorService,
	}
}

// RegisterRoutes registers error ingestion routes
func (eh *ErrorHandler) RegisterRoutes(r chi.Router, projectMiddleware *middleware.ProjectMiddleware) {
	// Sentry-compatible error ingestion endpoint (specific path to avoid conflicts)
	r.Group(func(r chi.Router) {
		r.Use(projectMiddleware.DSNAuth) // Use DSN authentication
		r.Post("/api/{project_id}/store/", eh.sentryStoreHandler)
	})

	// Alternative error ingestion endpoints
	r.Route("/api/v1/errors", func(r chi.Router) {
		r.Use(projectMiddleware.DSNAuth) // Use DSN authentication
		r.Post("/ingest", eh.errorIngestHandler)
		r.Get("/stats", eh.errorStatsHandler)
		r.Get("/issues/{issue_id}/events", eh.issueEventsHandler)
	})
}

// sentryStoreHandler handles the Sentry-compatible store endpoint
func (eh *ErrorHandler) sentryStoreHandler(w http.ResponseWriter, r *http.Request) {
	// Get project from context (set by DSN auth middleware)
	projectCtx, ok := middleware.GetProjectFromContext(r.Context())
	if !ok {
		eh.writeErrorResponse(w, http.StatusInternalServerError, "project not found in context")
		return
	}

	// Validate project ID from URL matches authenticated project
	projectIDStr := chi.URLParam(r, "project_id")
	if projectIDStr == "" {
		eh.writeErrorResponse(w, http.StatusBadRequest, "project ID required")
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		eh.writeErrorResponse(w, http.StatusBadRequest, "invalid project ID format")
		return
	}

	if projectID != projectCtx.ID {
		eh.writeErrorResponse(w, http.StatusForbidden, "project ID mismatch")
		return
	}

	eh.handleErrorIngestion(w, r, projectID)
}

// errorIngestHandler handles the alternative error ingestion endpoint
func (eh *ErrorHandler) errorIngestHandler(w http.ResponseWriter, r *http.Request) {
	// Get project from context (set by DSN auth middleware)
	projectCtx, ok := middleware.GetProjectFromContext(r.Context())
	if !ok {
		eh.writeErrorResponse(w, http.StatusInternalServerError, "project not found in context")
		return
	}

	eh.handleErrorIngestion(w, r, projectCtx.ID)
}

// handleErrorIngestion processes the error ingestion request
func (eh *ErrorHandler) handleErrorIngestion(w http.ResponseWriter, r *http.Request, projectID uuid.UUID) {
	// Check content type
	contentType := r.Header.Get("Content-Type")
	if !eh.isValidContentType(contentType) {
		eh.writeErrorResponse(w, http.StatusUnsupportedMediaType, 
			"unsupported content type, expected application/json or application/octet-stream")
		return
	}

	// Get request body reader
	bodyReader, err := eh.getBodyReader(r)
	if err != nil {
		eh.writeErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to read request body: %v", err))
		return
	}
	defer bodyReader.Close()

	// Parse the error event data
	var eventData dto.ErrorEventRequest
	if err := json.NewDecoder(bodyReader).Decode(&eventData); err != nil {
		eh.writeErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON payload: %v", err))
		return
	}

	// Get client information
	clientIP := eh.getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Process the error event
	response, err := eh.errorService.ProcessErrorEvent(projectID, &eventData, clientIP, userAgent)
	if err != nil {
		// Handle different types of errors
		switch {
		case strings.Contains(err.Error(), "validation failed"):
			eh.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		case strings.Contains(err.Error(), "project not found"):
			eh.writeErrorResponse(w, http.StatusNotFound, "project not found")
		case strings.Contains(err.Error(), "project is inactive"):
			eh.writeErrorResponse(w, http.StatusForbidden, "project is inactive")
		case strings.Contains(err.Error(), "event already exists"):
			eh.writeErrorResponse(w, http.StatusConflict, "event already exists")
		default:
			eh.writeErrorResponse(w, http.StatusInternalServerError, "failed to process error event")
		}
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// errorStatsHandler returns error statistics for the authenticated project
func (eh *ErrorHandler) errorStatsHandler(w http.ResponseWriter, r *http.Request) {
	// Get project from context
	projectCtx, ok := middleware.GetProjectFromContext(r.Context())
	if !ok {
		eh.writeErrorResponse(w, http.StatusInternalServerError, "project not found in context")
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // Default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get issue statistics
	stats, err := eh.errorService.GetIssueStats(projectCtx.ID, limit, offset)
	if err != nil {
		eh.writeErrorResponse(w, http.StatusInternalServerError, "failed to get error statistics")
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"issues": stats,
		"meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"count":  len(stats),
		},
	})
}

// issueEventsHandler returns events for a specific issue
func (eh *ErrorHandler) issueEventsHandler(w http.ResponseWriter, r *http.Request) {
	// Get project from context
	projectCtx, ok := middleware.GetProjectFromContext(r.Context())
	if !ok {
		eh.writeErrorResponse(w, http.StatusInternalServerError, "project not found in context")
		return
	}

	// Get issue ID from URL
	issueIDStr := chi.URLParam(r, "issue_id")
	if issueIDStr == "" {
		eh.writeErrorResponse(w, http.StatusBadRequest, "issue ID required")
		return
	}

	issueID, err := uuid.Parse(issueIDStr)
	if err != nil {
		eh.writeErrorResponse(w, http.StatusBadRequest, "invalid issue ID format")
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // Default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get events for the issue
	events, err := eh.errorService.GetIssueEvents(issueID, limit, offset)
	if err != nil {
		eh.writeErrorResponse(w, http.StatusInternalServerError, "failed to get issue events")
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"count":  len(events),
			"issue_id": issueID,
			"project_id": projectCtx.ID,
		},
	})
}

// isValidContentType checks if the content type is acceptable
func (eh *ErrorHandler) isValidContentType(contentType string) bool {
	// Remove charset and other parameters
	mediaType := strings.ToLower(strings.Split(contentType, ";")[0])
	mediaType = strings.TrimSpace(mediaType)

	validTypes := []string{
		"application/json",
		"application/octet-stream", // For gzipped payloads
		"text/plain",               // Some clients send this
	}

	for _, validType := range validTypes {
		if mediaType == validType {
			return true
		}
	}

	return false
}

// getBodyReader returns an appropriate reader for the request body
func (eh *ErrorHandler) getBodyReader(r *http.Request) (io.ReadCloser, error) {
	// Check if the content is gzip compressed
	if r.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		return gzipReader, nil
	}

	// For application/octet-stream, also try gzip
	if r.Header.Get("Content-Type") == "application/octet-stream" {
		// Try to read as gzip first
		gzipReader, err := gzip.NewReader(r.Body)
		if err == nil {
			return gzipReader, nil
		}
		// If not gzip, reset and use raw body
		// Note: This is a simplified approach. In production, you might want to buffer and detect format
	}

	return r.Body, nil
}

// getClientIP extracts the client IP address from the request
func (eh *ErrorHandler) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (used by proxies)
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, use the first one
		ips := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header (used by nginx)
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}

	return ip
}

// writeErrorResponse writes a JSON error response
func (eh *ErrorHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":   http.StatusText(statusCode),
		"message": message,
		"status":  statusCode,
	}

	json.NewEncoder(w).Encode(response)
}

// authMiddleware for error ingestion endpoints (uses DSN authentication)
func (eh *ErrorHandler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is handled by the DSN auth middleware in the project middleware
		next.ServeHTTP(w, r)
	})
}

// rateLimitMiddleware could be added for rate limiting error ingestion
func (eh *ErrorHandler) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement rate limiting based on project/DSN
		// For now, just pass through
		next.ServeHTTP(w, r)
	})
}