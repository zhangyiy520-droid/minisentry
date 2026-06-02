package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"minisentry/internal/database"
	"minisentry/internal/dto"
	"minisentry/internal/models"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrInvalidEventData = errors.New("invalid event data")
	ErrEventExists      = errors.New("event already exists")
)

type ErrorService struct {
	db                 *database.DB
	fingerprintService *FingerprintService
}

// NewErrorService creates a new error processing service
func NewErrorService(db *database.DB) *ErrorService {
	return &ErrorService{
		db:                 db,
		fingerprintService: NewFingerprintService(),
	}
}

// ProcessErrorEvent is the main entry point for error processing
func (es *ErrorService) ProcessErrorEvent(projectID uuid.UUID, eventData *dto.ErrorEventRequest, clientIP, userAgent string) (*dto.ErrorEventResponse, error) {
	// Validate the error payload
	if err := es.ValidateErrorPayload(eventData); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Normalize the error data
	normalizedData, err := es.NormalizeErrorData(projectID, eventData, clientIP, userAgent)
	if err != nil {
		return nil, fmt.Errorf("normalization failed: %w", err)
	}

	// Generate or use custom fingerprint
	fingerprint := es.generateFingerprint(normalizedData, eventData.Fingerprint)
	normalizedData.Fingerprint = fingerprint

	// Find or create issue
	issue, err := es.FindOrCreateIssue(projectID, normalizedData)
	if err != nil {
		return nil, fmt.Errorf("issue management failed: %w", err)
	}

	// Create error event
	event, err := es.CreateErrorEvent(issue.ID, normalizedData)
	if err != nil {
		return nil, fmt.Errorf("event creation failed: %w", err)
	}

	// Update issue statistics
	if err := es.updateIssueStats(issue); err != nil {
		return nil, fmt.Errorf("issue stats update failed: %w", err)
	}

	return &dto.ErrorEventResponse{
		ID:        event.ID.String(),
		EventID:   event.EventID,
		ProjectID: event.ProjectID,
		IssueID:   event.IssueID,
		CreatedAt: event.CreatedAt,
	}, nil
}

// ValidateErrorPayload validates the incoming error payload
func (es *ErrorService) ValidateErrorPayload(eventData *dto.ErrorEventRequest) error {
	if eventData == nil {
		return fmt.Errorf("%w: event data is nil", ErrInvalidEventData)
	}

	// At least one of message or exception should be present
	if eventData.Message == nil && eventData.Exception == nil {
		return fmt.Errorf("%w: neither message nor exception provided", ErrInvalidEventData)
	}

	// If exception is provided, it should have values
	if eventData.Exception != nil && len(eventData.Exception.Values) == 0 {
		return fmt.Errorf("%w: exception provided but no exception values", ErrInvalidEventData)
	}

	// Validate level if provided
	if eventData.Level != nil {
		validLevels := []string{"debug", "info", "warning", "error", "fatal"}
		isValid := false
		for _, level := range validLevels {
			if strings.ToLower(*eventData.Level) == level {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("%w: invalid level '%s'", ErrInvalidEventData, *eventData.Level)
		}
	}

	return nil
}

// NormalizeErrorData cleans and standardizes error data
func (es *ErrorService) NormalizeErrorData(projectID uuid.UUID, eventData *dto.ErrorEventRequest, clientIP, userAgent string) (*dto.NormalizedErrorData, error) {
	normalized := &dto.NormalizedErrorData{
		ProjectID: projectID,
		Platform:  "javascript", // Default platform
	}

	// Generate or use provided event ID
	if eventData.EventID != nil {
		normalized.EventID = *eventData.EventID
	} else {
		normalized.EventID = uuid.New().String()
	}

	// Set timestamp
	if eventData.Timestamp != nil {
		normalized.Timestamp = *eventData.Timestamp
	} else {
		normalized.Timestamp = time.Now()
	}

	// Set level
	if eventData.Level != nil {
		normalized.Level = strings.ToLower(*eventData.Level)
	} else {
		normalized.Level = "error"
	}

	// Set platform
	if eventData.Platform != nil {
		normalized.Platform = *eventData.Platform
	}

	// Set environment
	if eventData.Environment != nil {
		normalized.Environment = *eventData.Environment
	} else {
		normalized.Environment = "production"
	}

	// Set release
	if eventData.Release != nil {
		normalized.Release = eventData.Release
	}

	// Set server name
	if eventData.ServerName != nil {
		normalized.ServerName = eventData.ServerName
	}

	// Extract message
	if eventData.Message != nil {
		if eventData.Message.Formatted != nil {
			normalized.Message = eventData.Message.Formatted
		} else {
			normalized.Message = &eventData.Message.Message
		}
	}

	// Extract exception data
	if eventData.Exception != nil && len(eventData.Exception.Values) > 0 {
		mainException := eventData.Exception.Values[0] // Use the first exception
		normalized.ExceptionType = mainException.Type
		normalized.ExceptionValue = mainException.Value

		// Extract stack trace
		if mainException.Stacktrace != nil {
			normalized.StackTrace = mainException.Stacktrace.Frames
		}
	}

	// Set user context
	if eventData.User != nil {
		normalized.UserContext = eventData.User
	}

	// Set request data
	if eventData.Request != nil {
		normalized.RequestData = eventData.Request
	}

	// Set tags
	if eventData.Tags != nil {
		normalized.Tags = eventData.Tags
	} else {
		normalized.Tags = make(map[string]string)
	}

	// Add client information to tags
	if clientIP != "" {
		normalized.Tags["client_ip"] = clientIP
	}
	if userAgent != "" {
		normalized.Tags["user_agent"] = userAgent
	}

	// Set extra data
	if eventData.Extra != nil {
		normalized.ExtraData = eventData.Extra
	} else {
		normalized.ExtraData = make(map[string]interface{})
	}

	// Set breadcrumbs
	if eventData.Breadcrumbs != nil {
		normalized.Breadcrumbs = eventData.Breadcrumbs
	}

	return normalized, nil
}

// generateFingerprint creates a fingerprint for the error
func (es *ErrorService) generateFingerprint(normalizedData *dto.NormalizedErrorData, customFingerprint []string) string {
	if len(customFingerprint) > 0 {
		return es.fingerprintService.CustomFingerprint(normalizedData, customFingerprint)
	}
	return es.fingerprintService.GenerateErrorFingerprint(normalizedData)
}

// FindOrCreateIssue finds an existing issue or creates a new one
func (es *ErrorService) FindOrCreateIssue(projectID uuid.UUID, normalizedData *dto.NormalizedErrorData) (*models.Issue, error) {
	var issue models.Issue

	// Try to find existing issue by fingerprint
	result := es.db.DB.Where("project_id = ? AND fingerprint = ?", projectID, normalizedData.Fingerprint).First(&issue)
	
	if result.Error == nil {
		// Issue exists, return it
		return &issue, nil
	}

	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Database error
		return nil, fmt.Errorf("failed to query issue: %w", result.Error)
	}

	// Create new issue
	issue = models.Issue{
		ProjectID:   projectID,
		Fingerprint: normalizedData.Fingerprint,
		Title:       es.generateIssueTitle(normalizedData),
		Culprit:     es.generateCulprit(normalizedData),
		Type:        es.determineIssueType(normalizedData),
		Level:       models.IssueLevel(normalizedData.Level),
		Status:      models.StatusUnresolved,
		FirstSeen:   normalizedData.Timestamp,
		LastSeen:    normalizedData.Timestamp,
		TimesSeen:   1,
	}

	if err := es.db.DB.Create(&issue).Error; err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	return &issue, nil
}

// generateIssueTitle creates a descriptive title for the issue
func (es *ErrorService) generateIssueTitle(normalizedData *dto.NormalizedErrorData) string {
	if normalizedData.ExceptionType != nil && normalizedData.ExceptionValue != nil {
		return fmt.Sprintf("%s: %s", *normalizedData.ExceptionType, *normalizedData.ExceptionValue)
	}
	
	if normalizedData.Message != nil {
		return *normalizedData.Message
	}

	if normalizedData.ExceptionType != nil {
		return *normalizedData.ExceptionType
	}

	return "Unknown Error"
}

// generateCulprit identifies the likely source of the error
func (es *ErrorService) generateCulprit(normalizedData *dto.NormalizedErrorData) *string {
	if len(normalizedData.StackTrace) == 0 {
		return nil
	}

	// Find the first in-app frame
	for _, frame := range normalizedData.StackTrace {
		if frame.InApp != nil && *frame.InApp {
			culprit := es.buildCulpritString(frame)
			if culprit != "" {
				return &culprit
			}
		}
	}

	// If no in-app frame found, use the first frame
	culprit := es.buildCulpritString(normalizedData.StackTrace[0])
	if culprit != "" {
		return &culprit
	}

	return nil
}

// buildCulpritString builds a culprit string from a stack frame
func (es *ErrorService) buildCulpritString(frame dto.StackFrame) string {
	var parts []string

	if frame.Function != nil && *frame.Function != "" {
		parts = append(parts, *frame.Function)
	}

	if frame.Filename != nil && *frame.Filename != "" {
		filename := *frame.Filename
		// Extract just the filename from the path
		if lastSlash := strings.LastIndex(filename, "/"); lastSlash != -1 {
			filename = filename[lastSlash+1:]
		}
		
		location := filename
		if frame.Lineno != nil {
			location = fmt.Sprintf("%s:%d", location, *frame.Lineno)
		}
		parts = append(parts, location)
	}

	return strings.Join(parts, " at ")
}

// determineIssueType determines the type of issue based on the error data
func (es *ErrorService) determineIssueType(normalizedData *dto.NormalizedErrorData) models.IssueType {
	if normalizedData.ExceptionType != nil {
		exceptionType := strings.ToLower(*normalizedData.ExceptionType)
		if strings.Contains(exceptionType, "csp") {
			return models.TypeCSP
		}
	}

	return models.TypeError
}

// CreateErrorEvent creates a new error event
func (es *ErrorService) CreateErrorEvent(issueID uuid.UUID, normalizedData *dto.NormalizedErrorData) (*models.Event, error) {
	// Check if event already exists
	var existingEvent models.Event
	result := es.db.DB.Where("project_id = ? AND event_id = ?", normalizedData.ProjectID, normalizedData.EventID).First(&existingEvent)
	if result.Error == nil {
		return nil, ErrEventExists
	}

	// Serialize complex data to JSON
	stackTraceJSON, err := json.Marshal(normalizedData.StackTrace)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal stack trace: %w", err)
	}

	requestDataJSON, err := json.Marshal(normalizedData.RequestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	userContextJSON, err := json.Marshal(normalizedData.UserContext)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user context: %w", err)
	}

	tagsJSON, err := json.Marshal(normalizedData.Tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %w", err)
	}

	extraDataJSON, err := json.Marshal(normalizedData.ExtraData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal extra data: %w", err)
	}

	// Create event
	event := models.Event{
		IssueID:         issueID,
		ProjectID:       normalizedData.ProjectID,
		EventID:         normalizedData.EventID,
		Timestamp:       normalizedData.Timestamp,
		Level:           models.IssueLevel(normalizedData.Level),
		Message:         normalizedData.Message,
		ExceptionType:   normalizedData.ExceptionType,
		ExceptionValue:  normalizedData.ExceptionValue,
		StackTrace:      datatypes.JSON(stackTraceJSON),
		RequestData:     datatypes.JSON(requestDataJSON),
		UserContext:     datatypes.JSON(userContextJSON),
		Tags:            datatypes.JSON(tagsJSON),
		ExtraData:       datatypes.JSON(extraDataJSON),
		Fingerprint:     normalizedData.Fingerprint,
		ReleaseVersion:  normalizedData.Release,
		Environment:     normalizedData.Environment,
		ServerName:      normalizedData.ServerName,
	}

	if err := es.db.DB.Create(&event).Error; err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return &event, nil
}

// updateIssueStats updates issue statistics
func (es *ErrorService) updateIssueStats(issue *models.Issue) error {
	updates := map[string]interface{}{
		"last_seen":   time.Now(),
		"times_seen":  gorm.Expr("times_seen + ?", 1),
		"updated_at":  time.Now(),
	}

	if err := es.db.DB.Model(issue).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update issue stats: %w", err)
	}

	return nil
}

// GetIssueStats retrieves issue statistics for a project
func (es *ErrorService) GetIssueStats(projectID uuid.UUID, limit int, offset int) ([]dto.IssueListItemResponse, error) {
	var issues []models.Issue

	query := es.db.DB.Where("project_id = ?", projectID).
		Order("last_seen DESC").
		Limit(limit).
		Offset(offset)

	if err := query.Find(&issues).Error; err != nil {
		return nil, fmt.Errorf("failed to get issues: %w", err)
	}

	var stats []dto.IssueListItemResponse
	for _, issue := range issues {
		stats = append(stats, dto.IssueListItemResponse{
			ID:          issue.ID,
			ProjectID:   issue.ProjectID,
			Fingerprint: issue.Fingerprint,
			Title:       issue.Title,
			Culprit:     issue.Culprit,
			Type:        string(issue.Type),
			Level:       string(issue.Level),
			Status:      string(issue.Status),
			FirstSeen:   issue.FirstSeen,
			LastSeen:    issue.LastSeen,
			TimesSeen:   issue.TimesSeen,
			CreatedAt:   issue.CreatedAt,
			UpdatedAt:   issue.UpdatedAt,
		})
	}

	return stats, nil
}

// GetIssueEvents retrieves events for a specific issue
func (es *ErrorService) GetIssueEvents(issueID uuid.UUID, limit int, offset int) ([]models.Event, error) {
	var events []models.Event

	query := es.db.DB.Where("issue_id = ?", issueID).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset)

	if err := query.Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	return events, nil
}