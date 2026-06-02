package dto

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// IssueFilters represents filtering and sorting options for issue queries
type IssueFilters struct {
	Status      []string  `form:"status" json:"status,omitempty"`           // unresolved, resolved, ignored
	Level       []string  `form:"level" json:"level,omitempty"`             // error, warning, info, debug
	AssignedTo  *string   `form:"assigned_to" json:"assigned_to,omitempty"` // user_id
	DateFrom    *string   `form:"date_from" json:"date_from,omitempty"`     // ISO date string
	DateTo      *string   `form:"date_to" json:"date_to,omitempty"`         // ISO date string
	Search      *string   `form:"search" json:"search,omitempty"`           // text search in title/message
	Sort        string    `form:"sort" json:"sort"`                         // frequency, first_seen, last_seen
	Order       string    `form:"order" json:"order"`                       // asc, desc
	Page        int       `form:"page" json:"page"`                         // page number (1-based)
	Limit       int       `form:"limit" json:"limit"`                       // items per page
	Environment *string   `form:"environment" json:"environment,omitempty"` // production, staging, etc
}

// IssueListResponse represents paginated issue list response
type IssueListResponse struct {
	Issues     []IssueResponse `json:"issues"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"total_pages"`
}

// IssueListItemResponse represents a basic issue response for lists
type IssueListItemResponse struct {
	ID          uuid.UUID `json:"id"`
	ProjectID   uuid.UUID `json:"project_id"`
	Fingerprint string    `json:"fingerprint"`
	Title       string    `json:"title"`
	Culprit     *string   `json:"culprit"`
	Type        string    `json:"type"`
	Level       string    `json:"level"`
	Status      string    `json:"status"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
	TimesSeen   int       `json:"times_seen"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// IssueResponse represents detailed issue response
type IssueResponse struct {
	ID           uuid.UUID                `json:"id"`
	ProjectID    uuid.UUID                `json:"project_id"`
	Fingerprint  string                   `json:"fingerprint"`
	Title        string                   `json:"title"`
	Culprit      *string                  `json:"culprit"`
	Type         string                   `json:"type"`
	Level        string                   `json:"level"`
	Status       string                   `json:"status"`
	FirstSeen    time.Time                `json:"first_seen"`
	LastSeen     time.Time                `json:"last_seen"`
	TimesSeen    int                      `json:"times_seen"`
	AssigneeID   *uuid.UUID               `json:"assignee_id"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
	
	// Optional related data
	Assignee     *IssueAssigneeResponse   `json:"assignee,omitempty"`
	Project      *IssueProjectResponse    `json:"project,omitempty"`
	LatestEvent  *IssueEventResponse      `json:"latest_event,omitempty"`
	CommentCount int                      `json:"comment_count,omitempty"`
	Tags         map[string]string        `json:"tags,omitempty"`
}

// IssueAssigneeResponse represents assignee information in issue response
type IssueAssigneeResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
}

// IssueProjectResponse represents project information in issue response
type IssueProjectResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

// IssueEventResponse represents event information in issue response
type IssueEventResponse struct {
	ID             uuid.UUID      `json:"id"`
	EventID        string         `json:"event_id"`
	Timestamp      time.Time      `json:"timestamp"`
	Level          string         `json:"level"`
	Message        *string        `json:"message"`
	ExceptionType  *string        `json:"exception_type"`
	ExceptionValue *string        `json:"exception_value"`
	Environment    string         `json:"environment"`
	ReleaseVersion *string        `json:"release_version"`
	ServerName     *string        `json:"server_name"`
	UserContext    datatypes.JSON `json:"user_context,omitempty"`
	Tags           datatypes.JSON `json:"tags,omitempty"`
}

// IssueUpdateRequest represents request to update issue status or assignment
type IssueUpdateRequest struct {
	Status     *string     `json:"status,omitempty"`     // resolved, ignored, unresolved
	AssigneeID *uuid.UUID  `json:"assignee_id,omitempty"` // null to unassign
	Resolution *string     `json:"resolution,omitempty"`  // resolution reason
}

// IssueCommentRequest represents request to add comment to issue
type IssueCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

// IssueCommentResponse represents issue comment response
type IssueCommentResponse struct {
	ID        uuid.UUID `json:"id"`
	IssueID   uuid.UUID `json:"issue_id"`
	UserID    uuid.UUID `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// User information
	User IssueCommentUserResponse `json:"user"`
}

// IssueCommentUserResponse represents user info in comment response
type IssueCommentUserResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
}

// IssueCommentsResponse represents paginated comments response
type IssueCommentsResponse struct {
	Comments   []IssueCommentResponse `json:"comments"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
	TotalPages int                    `json:"total_pages"`
}

// IssueActivityResponse represents issue activity timeline entry
type IssueActivityResponse struct {
	ID        uuid.UUID      `json:"id"`
	IssueID   uuid.UUID      `json:"issue_id"`
	UserID    *uuid.UUID     `json:"user_id"`
	Type      string         `json:"type"`
	Data      datatypes.JSON `json:"data"`
	CreatedAt time.Time      `json:"created_at"`
	
	// Optional user information
	User *IssueActivityUserResponse `json:"user,omitempty"`
}

// IssueActivityUserResponse represents user info in activity response
type IssueActivityUserResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
}

// IssueActivitiesResponse represents paginated activities response
type IssueActivitiesResponse struct {
	Activities []IssueActivityResponse `json:"activities"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	Limit      int                     `json:"limit"`
	TotalPages int                     `json:"total_pages"`
}

// IssueEventsResponse represents paginated events response
type IssueEventsResponse struct {
	Events     []IssueEventResponse `json:"events"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	Limit      int                  `json:"limit"`
	TotalPages int                  `json:"total_pages"`
}

// IssueStatsResponse represents dashboard statistics for issues
type IssueStatsResponse struct {
	Total         int64                    `json:"total"`
	Unresolved    int64                    `json:"unresolved"`
	Resolved      int64                    `json:"resolved"`
	Ignored       int64                    `json:"ignored"`
	NewToday      int64                    `json:"new_today"`
	NewThisWeek   int64                    `json:"new_this_week"`
	ByLevel       map[string]int64         `json:"by_level"`
	ByEnvironment map[string]int64         `json:"by_environment"`
	TopIssues     []IssueResponse          `json:"top_issues"`
	Timeline      []IssueTimelineEntry     `json:"timeline"`
}

// IssueTimelineEntry represents issue count over time
type IssueTimelineEntry struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// BulkUpdateIssuesRequest represents request to bulk update issues
type BulkUpdateIssuesRequest struct {
	IssueIDs   []uuid.UUID `json:"issue_ids" binding:"required"`
	Action     string      `json:"action" binding:"required"`     // resolve, ignore, unresolve, assign
	AssigneeID *uuid.UUID  `json:"assignee_id,omitempty"`         // for assign action
	Resolution *string     `json:"resolution,omitempty"`          // resolution reason
}

// BulkUpdateIssuesResponse represents response from bulk update operation
type BulkUpdateIssuesResponse struct {
	UpdatedCount int         `json:"updated_count"`
	FailedCount  int         `json:"failed_count"`
	Errors       []string    `json:"errors,omitempty"`
	UpdatedIDs   []uuid.UUID `json:"updated_ids"`
}

// IssueSearchResponse represents search results for issues
type IssueSearchResponse struct {
	Results    []IssueResponse `json:"results"`
	Total      int64           `json:"total"`
	Query      string          `json:"query"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"total_pages"`
}

// Common pagination request
type PaginationRequest struct {
	Page  int `form:"page" json:"page"`   // 1-based page number
	Limit int `form:"limit" json:"limit"` // items per page
}

// GetDefaultPagination returns default pagination values
func (p *PaginationRequest) GetDefaults() (int, int) {
	page := 1
	limit := 25
	
	if p.Page > 0 {
		page = p.Page
	}
	if p.Limit > 0 && p.Limit <= 100 {
		limit = p.Limit
	}
	
	return page, limit
}

// GetOffset calculates the offset for database queries
func (p *PaginationRequest) GetOffset() int {
	page, limit := p.GetDefaults()
	return (page - 1) * limit
}

// CalculateTotalPages calculates total pages from total count and limit
func CalculateTotalPages(total int64, limit int) int {
	if total == 0 || limit == 0 {
		return 0
	}
	return int((total + int64(limit) - 1) / int64(limit))
}