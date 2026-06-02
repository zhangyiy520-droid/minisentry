package services

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"minisentry/internal/dto"
	"minisentry/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IssueService struct {
	db *gorm.DB
}

func NewIssueService(db *gorm.DB) *IssueService {
	return &IssueService{db: db}
}

// GetProjectIssues retrieves issues for a project with filtering, sorting, and pagination
func (s *IssueService) GetProjectIssues(projectID uuid.UUID, filters dto.IssueFilters) (*dto.IssueListResponse, error) {
	query := s.db.Model(&models.Issue{}).Where("project_id = ?", projectID)
	
	// Apply filters
	query = s.applyIssueFilters(query, filters)
	
	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count issues: %w", err)
	}
	
	// Apply sorting
	query = s.applyIssueSorting(query, filters)
	
	// Apply pagination
	page, limit := s.getPaginationDefaults(filters.Page, filters.Limit)
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)
	
	// Preload associations
	query = query.Preload("Assignee").Preload("Project")
	
	var issues []models.Issue
	if err := query.Find(&issues).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve issues: %w", err)
	}
	
	// Convert to response DTOs
	issueResponses := make([]dto.IssueResponse, len(issues))
	for i, issue := range issues {
		issueResponse, err := s.convertIssueToResponse(issue, true)
		if err != nil {
			log.Printf("Error converting issue %s to response: %v", issue.ID, err)
			continue
		}
		issueResponses[i] = *issueResponse
	}
	
	totalPages := dto.CalculateTotalPages(total, limit)
	
	return &dto.IssueListResponse{
		Issues:     issueResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// GetIssue retrieves a single issue with detailed information
func (s *IssueService) GetIssue(issueID uuid.UUID) (*dto.IssueResponse, error) {
	var issue models.Issue
	if err := s.db.Preload("Assignee").Preload("Project").
		First(&issue, issueID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("issue not found")
		}
		return nil, fmt.Errorf("failed to retrieve issue: %w", err)
	}
	
	return s.convertIssueToResponse(issue, true)
}

// UpdateIssueStatus updates the status or assignment of an issue
func (s *IssueService) UpdateIssueStatus(issueID uuid.UUID, userID uuid.UUID, request dto.IssueUpdateRequest) (*dto.IssueResponse, error) {
	var issue models.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("issue not found")
		}
		return nil, fmt.Errorf("failed to retrieve issue: %w", err)
	}
	
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	oldStatus := issue.Status
	oldAssigneeID := issue.AssigneeID
	
	// Update fields
	updates := make(map[string]interface{})
	
	if request.Status != nil {
		status := models.IssueStatus(*request.Status)
		if s.isValidStatusTransition(issue.Status, status) {
			updates["status"] = status
			issue.Status = status
		} else {
			tx.Rollback()
			return nil, fmt.Errorf("invalid status transition from %s to %s", issue.Status, status)
		}
	}
	
	if request.AssigneeID != nil {
		updates["assignee_id"] = *request.AssigneeID
		issue.AssigneeID = request.AssigneeID
	}
	
	// Update issue
	if len(updates) > 0 {
		if err := tx.Model(&issue).Updates(updates).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update issue: %w", err)
		}
	}
	
	// Log activities
	if request.Status != nil && string(oldStatus) != *request.Status {
		if err := s.logStatusChangeActivity(tx, issueID, userID, string(oldStatus), *request.Status, request.Resolution); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to log status change activity: %w", err)
		}
	}
	
	if request.AssigneeID != nil && !s.uuidPtrEqual(oldAssigneeID, request.AssigneeID) {
		if err := s.logAssignmentActivity(tx, issueID, userID, oldAssigneeID, request.AssigneeID); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to log assignment activity: %w", err)
		}
	}
	
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	// Return updated issue
	return s.GetIssue(issueID)
}

// AssignIssue assigns an issue to a user
func (s *IssueService) AssignIssue(issueID uuid.UUID, userID uuid.UUID, assigneeID *uuid.UUID) (*dto.IssueResponse, error) {
	request := dto.IssueUpdateRequest{
		AssigneeID: assigneeID,
	}
	return s.UpdateIssueStatus(issueID, userID, request)
}

// AddIssueComment adds a comment to an issue
func (s *IssueService) AddIssueComment(issueID uuid.UUID, userID uuid.UUID, request dto.IssueCommentRequest) (*dto.IssueCommentResponse, error) {
	// Verify issue exists
	var issue models.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("issue not found")
		}
		return nil, fmt.Errorf("failed to verify issue: %w", err)
	}
	
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// Create comment
	comment := models.IssueComment{
		IssueID: issueID,
		UserID:  userID,
		Content: request.Content,
	}
	comment.ID = uuid.New()
	
	if err := tx.Create(&comment).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}
	
	// Log comment activity
	if err := s.logCommentActivity(tx, issueID, userID); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to log comment activity: %w", err)
	}
	
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	// Return comment with user info
	if err := s.db.Preload("User").First(&comment, comment.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve comment with user: %w", err)
	}
	
	return s.convertCommentToResponse(comment), nil
}

// GetIssueComments retrieves paginated comments for an issue
func (s *IssueService) GetIssueComments(issueID uuid.UUID, page, limit int) (*dto.IssueCommentsResponse, error) {
	page, limit = s.getPaginationDefaults(page, limit)
	offset := (page - 1) * limit
	
	// Count total comments
	var total int64
	if err := s.db.Model(&models.IssueComment{}).Where("issue_id = ?", issueID).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count comments: %w", err)
	}
	
	// Get comments with user info
	var comments []models.IssueComment
	if err := s.db.Where("issue_id = ?", issueID).
		Preload("User").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve comments: %w", err)
	}
	
	// Convert to response DTOs
	commentResponses := make([]dto.IssueCommentResponse, len(comments))
	for i, comment := range comments {
		commentResponses[i] = *s.convertCommentToResponse(comment)
	}
	
	totalPages := dto.CalculateTotalPages(total, limit)
	
	return &dto.IssueCommentsResponse{
		Comments:   commentResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// GetIssueActivity retrieves paginated activity timeline for an issue
func (s *IssueService) GetIssueActivity(issueID uuid.UUID, page, limit int) (*dto.IssueActivitiesResponse, error) {
	page, limit = s.getPaginationDefaults(page, limit)
	offset := (page - 1) * limit
	
	// Count total activities
	var total int64
	if err := s.db.Model(&models.IssueActivity{}).Where("issue_id = ?", issueID).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count activities: %w", err)
	}
	
	// Get activities with user info
	var activities []models.IssueActivity
	if err := s.db.Where("issue_id = ?", issueID).
		Preload("User").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&activities).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve activities: %w", err)
	}
	
	// Convert to response DTOs
	activityResponses := make([]dto.IssueActivityResponse, len(activities))
	for i, activity := range activities {
		activityResponses[i] = *s.convertActivityToResponse(activity)
	}
	
	totalPages := dto.CalculateTotalPages(total, limit)
	
	return &dto.IssueActivitiesResponse{
		Activities: activityResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// GetIssueEvents retrieves paginated events for an issue
func (s *IssueService) GetIssueEvents(issueID uuid.UUID, page, limit int) (*dto.IssueEventsResponse, error) {
	page, limit = s.getPaginationDefaults(page, limit)
	offset := (page - 1) * limit
	
	// Count total events
	var total int64
	if err := s.db.Model(&models.Event{}).Where("issue_id = ?", issueID).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count events: %w", err)
	}
	
	// Get events
	var events []models.Event
	if err := s.db.Where("issue_id = ?", issueID).
		Order("timestamp DESC").
		Offset(offset).Limit(limit).
		Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve events: %w", err)
	}
	
	// Convert to response DTOs
	eventResponses := make([]dto.IssueEventResponse, len(events))
	for i, event := range events {
		eventResponses[i] = s.convertEventToResponse(event)
	}
	
	totalPages := dto.CalculateTotalPages(total, limit)
	
	return &dto.IssueEventsResponse{
		Events:     eventResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// GetIssueStats retrieves dashboard statistics for issues in a project
func (s *IssueService) GetIssueStats(projectID uuid.UUID) (*dto.IssueStatsResponse, error) {
	stats := &dto.IssueStatsResponse{
		ByLevel:       make(map[string]int64),
		ByEnvironment: make(map[string]int64),
		Timeline:      make([]dto.IssueTimelineEntry, 0),
	}
	
	// Get total counts by status
	var counts []struct {
		Status string
		Count  int64
	}
	if err := s.db.Model(&models.Issue{}).
		Where("project_id = ?", projectID).
		Select("status, count(*) as count").
		Group("status").
		Scan(&counts).Error; err != nil {
		return nil, fmt.Errorf("failed to get status counts: %w", err)
	}
	
	for _, count := range counts {
		stats.Total += count.Count
		switch count.Status {
		case string(models.StatusUnresolved):
			stats.Unresolved = count.Count
		case string(models.StatusResolved):
			stats.Resolved = count.Count
		case string(models.StatusIgnored):
			stats.Ignored = count.Count
		}
	}
	
	// Get counts by level
	var levelCounts []struct {
		Level string
		Count int64
	}
	if err := s.db.Model(&models.Issue{}).
		Where("project_id = ?", projectID).
		Select("level, count(*) as count").
		Group("level").
		Scan(&levelCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get level counts: %w", err)
	}
	
	for _, count := range levelCounts {
		stats.ByLevel[count.Level] = count.Count
	}
	
	// Get counts by environment (from latest events)
	var envCounts []struct {
		Environment string
		Count       int64
	}
	if err := s.db.Raw(`
		SELECT e.environment, COUNT(DISTINCT i.id) as count
		FROM issues i
		INNER JOIN events e ON e.issue_id = i.id
		WHERE i.project_id = ?
		GROUP BY e.environment
	`, projectID).Scan(&envCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get environment counts: %w", err)
	}
	
	for _, count := range envCounts {
		stats.ByEnvironment[count.Environment] = count.Count
	}
	
	// Get new issues today and this week
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startOfWeek := startOfDay.AddDate(0, 0, -int(startOfDay.Weekday()))
	
	if err := s.db.Model(&models.Issue{}).
		Where("project_id = ? AND first_seen >= ?", projectID, startOfDay).
		Count(&stats.NewToday).Error; err != nil {
		return nil, fmt.Errorf("failed to count new issues today: %w", err)
	}
	
	if err := s.db.Model(&models.Issue{}).
		Where("project_id = ? AND first_seen >= ?", projectID, startOfWeek).
		Count(&stats.NewThisWeek).Error; err != nil {
		return nil, fmt.Errorf("failed to count new issues this week: %w", err)
	}
	
	// Get top issues by frequency
	var topIssues []models.Issue
	if err := s.db.Where("project_id = ?", projectID).
		Preload("Assignee").Preload("Project").
		Order("times_seen DESC").
		Limit(10).
		Find(&topIssues).Error; err != nil {
		return nil, fmt.Errorf("failed to get top issues: %w", err)
	}
	
	stats.TopIssues = make([]dto.IssueResponse, len(topIssues))
	for i, issue := range topIssues {
		issueResponse, err := s.convertIssueToResponse(issue, false)
		if err != nil {
			log.Printf("Error converting top issue %s to response: %v", issue.ID, err)
			continue
		}
		stats.TopIssues[i] = *issueResponse
	}
	
	// Get timeline data (last 30 days)
	var timelineCounts []struct {
		Date  string
		Count int64
	}
	if err := s.db.Raw(`
		SELECT DATE(first_seen) as date, COUNT(*) as count
		FROM issues
		WHERE project_id = ? AND first_seen >= ?
		GROUP BY DATE(first_seen)
		ORDER BY date DESC
		LIMIT 30
	`, projectID, now.AddDate(0, 0, -30)).Scan(&timelineCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get timeline data: %w", err)
	}
	
	for _, count := range timelineCounts {
		stats.Timeline = append(stats.Timeline, dto.IssueTimelineEntry{
			Date:  count.Date,
			Count: count.Count,
		})
	}
	
	return stats, nil
}

// BulkUpdateIssues performs bulk operations on multiple issues
func (s *IssueService) BulkUpdateIssues(userID uuid.UUID, request dto.BulkUpdateIssuesRequest) (*dto.BulkUpdateIssuesResponse, error) {
	if len(request.IssueIDs) == 0 {
		return nil, fmt.Errorf("no issues specified")
	}
	
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	response := &dto.BulkUpdateIssuesResponse{
		UpdatedIDs: make([]uuid.UUID, 0),
		Errors:     make([]string, 0),
	}
	
	for _, issueID := range request.IssueIDs {
		var issue models.Issue
		if err := tx.First(&issue, issueID).Error; err != nil {
			response.FailedCount++
			response.Errors = append(response.Errors, fmt.Sprintf("Issue %s not found", issueID))
			continue
		}
		
		updates := make(map[string]interface{})
		var activityType models.ActivityType
		var activityData map[string]interface{}
		
		switch request.Action {
		case "resolve":
			if issue.Status != models.StatusResolved {
				updates["status"] = models.StatusResolved
				activityType = models.ActivityResolve
				activityData = map[string]interface{}{
					"previous_status": string(issue.Status),
					"new_status":      string(models.StatusResolved),
				}
				if request.Resolution != nil {
					activityData["resolution"] = *request.Resolution
				}
			}
		case "ignore":
			if issue.Status != models.StatusIgnored {
				updates["status"] = models.StatusIgnored
				activityType = models.ActivityIgnore
				activityData = map[string]interface{}{
					"previous_status": string(issue.Status),
					"new_status":      string(models.StatusIgnored),
				}
			}
		case "unresolve":
			if issue.Status != models.StatusUnresolved {
				updates["status"] = models.StatusUnresolved
				activityType = models.ActivityStatusChange
				activityData = map[string]interface{}{
					"previous_status": string(issue.Status),
					"new_status":      string(models.StatusUnresolved),
				}
			}
		case "assign":
			if request.AssigneeID != nil {
				updates["assignee_id"] = *request.AssigneeID
				activityType = models.ActivityAssignment
				activityData = map[string]interface{}{
					"assignee_id": *request.AssigneeID,
				}
				if issue.AssigneeID != nil {
					activityData["previous_assignee_id"] = *issue.AssigneeID
				}
			}
		default:
			response.FailedCount++
			response.Errors = append(response.Errors, fmt.Sprintf("Invalid action: %s", request.Action))
			continue
		}
		
		if len(updates) > 0 {
			if err := tx.Model(&issue).Updates(updates).Error; err != nil {
				response.FailedCount++
				response.Errors = append(response.Errors, fmt.Sprintf("Failed to update issue %s: %v", issueID, err))
				continue
			}
			
			// Log activity
			activityDataJSON, _ := json.Marshal(activityData)
			activity := models.IssueActivity{
				IssueID: issueID,
				UserID:  &userID,
				Type:    activityType,
				Data:    activityDataJSON,
			}
			activity.ID = uuid.New()
			
			if err := tx.Create(&activity).Error; err != nil {
				log.Printf("Failed to log activity for issue %s: %v", issueID, err)
			}
			
			response.UpdatedCount++
			response.UpdatedIDs = append(response.UpdatedIDs, issueID)
		}
	}
	
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit bulk update: %w", err)
	}
	
	return response, nil
}

// Helper methods

func (s *IssueService) applyIssueFilters(query *gorm.DB, filters dto.IssueFilters) *gorm.DB {
	// Status filter
	if len(filters.Status) > 0 {
		query = query.Where("status IN ?", filters.Status)
	}
	
	// Level filter
	if len(filters.Level) > 0 {
		query = query.Where("level IN ?", filters.Level)
	}
	
	// Assignee filter
	if filters.AssignedTo != nil {
		if *filters.AssignedTo == "unassigned" {
			query = query.Where("assignee_id IS NULL")
		} else {
			assigneeID, err := uuid.Parse(*filters.AssignedTo)
			if err == nil {
				query = query.Where("assignee_id = ?", assigneeID)
			}
		}
	}
	
	// Date range filter
	if filters.DateFrom != nil {
		if dateFrom, err := time.Parse("2006-01-02", *filters.DateFrom); err == nil {
			query = query.Where("first_seen >= ?", dateFrom)
		}
	}
	if filters.DateTo != nil {
		if dateTo, err := time.Parse("2006-01-02", *filters.DateTo); err == nil {
			query = query.Where("first_seen <= ?", dateTo.Add(24*time.Hour))
		}
	}
	
	// Environment filter (join with events)
	if filters.Environment != nil {
		query = query.Joins("JOIN events ON events.issue_id = issues.id").
			Where("events.environment = ?", *filters.Environment).
			Distinct()
	}
	
	// Text search
	if filters.Search != nil && *filters.Search != "" {
		searchTerm := "%" + strings.ToLower(*filters.Search) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(culprit) LIKE ?", searchTerm, searchTerm)
	}
	
	return query
}

func (s *IssueService) applyIssueSorting(query *gorm.DB, filters dto.IssueFilters) *gorm.DB {
	sortField := "last_seen"
	sortOrder := "DESC"
	
	switch filters.Sort {
	case "frequency":
		sortField = "times_seen"
	case "first_seen":
		sortField = "first_seen"
	case "last_seen":
		sortField = "last_seen"
	}
	
	if filters.Order == "asc" {
		sortOrder = "ASC"
	}
	
	return query.Order(fmt.Sprintf("%s %s", sortField, sortOrder))
}

func (s *IssueService) getPaginationDefaults(page, limit int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	return page, limit
}

func (s *IssueService) convertIssueToResponse(issue models.Issue, includeLatestEvent bool) (*dto.IssueResponse, error) {
	response := &dto.IssueResponse{
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
		AssigneeID:  issue.AssigneeID,
		CreatedAt:   issue.CreatedAt,
		UpdatedAt:   issue.UpdatedAt,
	}
	
	// Add assignee info
	if issue.Assignee != nil {
		response.Assignee = &dto.IssueAssigneeResponse{
			ID:       issue.Assignee.ID,
			Name:     issue.Assignee.Name,
			Email:    issue.Assignee.Email,
			Username: issue.Assignee.Email, // Use email as username since Username field doesn't exist
		}
	}
	
	// Add project info
	if issue.Project.ID != uuid.Nil {
		response.Project = &dto.IssueProjectResponse{
			ID:   issue.Project.ID,
			Name: issue.Project.Name,
			Slug: issue.Project.Slug,
		}
	}
	
	// Get comment count
	var commentCount int64
	if err := s.db.Model(&models.IssueComment{}).Where("issue_id = ?", issue.ID).Count(&commentCount).Error; err == nil {
		response.CommentCount = int(commentCount)
	}
	
	// Get latest event if requested
	if includeLatestEvent {
		var latestEvent models.Event
		if err := s.db.Where("issue_id = ?", issue.ID).Order("timestamp DESC").First(&latestEvent).Error; err == nil {
			response.LatestEvent = &dto.IssueEventResponse{
				ID:             latestEvent.ID,
				EventID:        latestEvent.EventID,
				Timestamp:      latestEvent.Timestamp,
				Level:          string(latestEvent.Level),
				Message:        latestEvent.Message,
				ExceptionType:  latestEvent.ExceptionType,
				ExceptionValue: latestEvent.ExceptionValue,
				Environment:    latestEvent.Environment,
				ReleaseVersion: latestEvent.ReleaseVersion,
				ServerName:     latestEvent.ServerName,
				UserContext:    latestEvent.UserContext,
				Tags:           latestEvent.Tags,
			}
		}
	}
	
	return response, nil
}

func (s *IssueService) convertCommentToResponse(comment models.IssueComment) *dto.IssueCommentResponse {
	response := &dto.IssueCommentResponse{
		ID:        comment.ID,
		IssueID:   comment.IssueID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}
	
	if comment.User.ID != uuid.Nil {
		response.User = dto.IssueCommentUserResponse{
			ID:       comment.User.ID,
			Name:     comment.User.Name,
			Email:    comment.User.Email,
			Username: comment.User.Email, // Use email as username since Username field doesn't exist
		}
	}
	
	return response
}

func (s *IssueService) convertActivityToResponse(activity models.IssueActivity) *dto.IssueActivityResponse {
	response := &dto.IssueActivityResponse{
		ID:        activity.ID,
		IssueID:   activity.IssueID,
		UserID:    activity.UserID,
		Type:      string(activity.Type),
		Data:      activity.Data,
		CreatedAt: activity.CreatedAt,
	}
	
	if activity.User != nil && activity.User.ID != uuid.Nil {
		response.User = &dto.IssueActivityUserResponse{
			ID:       activity.User.ID,
			Name:     activity.User.Name,
			Email:    activity.User.Email,
			Username: activity.User.Email, // Use email as username since Username field doesn't exist
		}
	}
	
	return response
}

func (s *IssueService) convertEventToResponse(event models.Event) dto.IssueEventResponse {
	return dto.IssueEventResponse{
		ID:             event.ID,
		EventID:        event.EventID,
		Timestamp:      event.Timestamp,
		Level:          string(event.Level),
		Message:        event.Message,
		ExceptionType:  event.ExceptionType,
		ExceptionValue: event.ExceptionValue,
		Environment:    event.Environment,
		ReleaseVersion: event.ReleaseVersion,
		ServerName:     event.ServerName,
		UserContext:    event.UserContext,
		Tags:           event.Tags,
	}
}

func (s *IssueService) isValidStatusTransition(from, to models.IssueStatus) bool {
	validTransitions := map[models.IssueStatus][]models.IssueStatus{
		models.StatusUnresolved: {models.StatusResolved, models.StatusIgnored},
		models.StatusResolved:   {models.StatusUnresolved},
		models.StatusIgnored:    {models.StatusUnresolved},
	}
	
	allowed, exists := validTransitions[from]
	if !exists {
		return false
	}
	
	for _, allowedStatus := range allowed {
		if allowedStatus == to {
			return true
		}
	}
	
	return false
}

func (s *IssueService) logStatusChangeActivity(tx *gorm.DB, issueID, userID uuid.UUID, oldStatus, newStatus string, resolution *string) error {
	data := map[string]interface{}{
		"previous_status": oldStatus,
		"new_status":      newStatus,
	}
	if resolution != nil {
		data["resolution"] = *resolution
	}
	
	var activityType models.ActivityType
	switch newStatus {
	case string(models.StatusResolved):
		activityType = models.ActivityResolve
	case string(models.StatusIgnored):
		activityType = models.ActivityIgnore
	default:
		activityType = models.ActivityStatusChange
	}
	
	return s.createActivity(tx, issueID, userID, activityType, data)
}

func (s *IssueService) logAssignmentActivity(tx *gorm.DB, issueID, userID uuid.UUID, oldAssigneeID, newAssigneeID *uuid.UUID) error {
	data := map[string]interface{}{}
	if oldAssigneeID != nil {
		data["previous_assignee_id"] = *oldAssigneeID
	}
	if newAssigneeID != nil {
		data["assignee_id"] = *newAssigneeID
	}
	
	return s.createActivity(tx, issueID, userID, models.ActivityAssignment, data)
}

func (s *IssueService) logCommentActivity(tx *gorm.DB, issueID, userID uuid.UUID) error {
	data := map[string]interface{}{
		"action": "comment_added",
	}
	
	return s.createActivity(tx, issueID, userID, models.ActivityComment, data)
}

func (s *IssueService) createActivity(tx *gorm.DB, issueID, userID uuid.UUID, activityType models.ActivityType, data map[string]interface{}) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal activity data: %w", err)
	}
	
	activity := models.IssueActivity{
		IssueID: issueID,
		UserID:  &userID,
		Type:    activityType,
		Data:    dataJSON,
	}
	activity.ID = uuid.New()
	
	return tx.Create(&activity).Error
}

func (s *IssueService) uuidPtrEqual(a, b *uuid.UUID) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}