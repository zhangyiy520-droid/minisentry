package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type IssueStatus string
type IssueLevel string
type IssueType string

const (
	StatusUnresolved IssueStatus = "unresolved"
	StatusResolved   IssueStatus = "resolved"
	StatusIgnored    IssueStatus = "ignored"
)

const (
	LevelDebug   IssueLevel = "debug"
	LevelInfo    IssueLevel = "info"
	LevelWarning IssueLevel = "warning"
	LevelError   IssueLevel = "error"
	LevelFatal   IssueLevel = "fatal"
)

const (
	TypeError   IssueType = "error"
	TypeCSP     IssueType = "csp"
	TypeDefault IssueType = "default"
)

type Issue struct {
	BaseModel
	ProjectID   uuid.UUID    `json:"project_id" gorm:"not null;index"`
	Fingerprint string       `json:"fingerprint" gorm:"not null;size:255;index:idx_project_fingerprint,unique"`
	Title       string       `json:"title" gorm:"not null;size:500"`
	Culprit     *string      `json:"culprit" gorm:"size:500"`
	Type        IssueType    `json:"type" gorm:"not null;size:100"`
	Level       IssueLevel   `json:"level" gorm:"not null;default:'error';size:50"`
	Status      IssueStatus  `json:"status" gorm:"default:'unresolved';size:50"`
	FirstSeen   time.Time    `json:"first_seen" gorm:"default:now()"`
	LastSeen    time.Time    `json:"last_seen" gorm:"default:now()"`
	TimesSeen   int          `json:"times_seen" gorm:"default:1"`
	AssigneeID  *uuid.UUID   `json:"assignee_id"`
	
	// Relationships
	Project   Project        `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Assignee  *User          `json:"assignee,omitempty" gorm:"foreignKey:AssigneeID"`
	Events    []Event        `json:"events,omitempty" gorm:"foreignKey:IssueID"`
	Comments  []IssueComment `json:"comments,omitempty" gorm:"foreignKey:IssueID"`
	Activities []IssueActivity `json:"activities,omitempty" gorm:"foreignKey:IssueID"`
}

type Event struct {
	BaseModel
	IssueID         uuid.UUID      `json:"issue_id" gorm:"not null;index"`
	ProjectID       uuid.UUID      `json:"project_id" gorm:"not null;index"`
	EventID         string         `json:"event_id" gorm:"not null;size:255;index:idx_project_event_id,unique"`
	Timestamp       time.Time      `json:"timestamp" gorm:"default:now();index"`
	Level           IssueLevel     `json:"level" gorm:"not null;default:'error';size:50"`
	Message         *string        `json:"message" gorm:"type:text"`
	ExceptionType   *string        `json:"exception_type" gorm:"size:255"`
	ExceptionValue  *string        `json:"exception_value" gorm:"type:text"`
	StackTrace      datatypes.JSON `json:"stack_trace" gorm:"type:jsonb"`
	RequestData     datatypes.JSON `json:"request_data" gorm:"type:jsonb"`
	UserContext     datatypes.JSON `json:"user_context" gorm:"type:jsonb"`
	Tags            datatypes.JSON `json:"tags" gorm:"type:jsonb"`
	ExtraData       datatypes.JSON `json:"extra_data" gorm:"type:jsonb"`
	Fingerprint     string         `json:"fingerprint" gorm:"not null;size:255"`
	ReleaseVersion  *string        `json:"release_version" gorm:"size:100"`
	Environment     string         `json:"environment" gorm:"default:'production';size:100"`
	ServerName      *string        `json:"server_name" gorm:"size:255"`
	
	// Relationships
	Issue   Issue   `json:"issue,omitempty" gorm:"foreignKey:IssueID"`
	Project Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}

type IssueComment struct {
	BaseModel
	IssueID uuid.UUID `json:"issue_id" gorm:"not null;index"`
	UserID  uuid.UUID `json:"user_id" gorm:"not null"`
	Content string    `json:"content" gorm:"not null;type:text"`
	
	// Relationships
	Issue Issue `json:"issue,omitempty" gorm:"foreignKey:IssueID"`
	User  User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type ActivityType string

const (
	ActivityStatusChange ActivityType = "status_change"
	ActivityAssignment   ActivityType = "assignment"
	ActivityComment      ActivityType = "comment"
	ActivityResolve      ActivityType = "resolve"
	ActivityIgnore       ActivityType = "ignore"
)

type IssueActivity struct {
	BaseModel
	IssueID uuid.UUID      `json:"issue_id" gorm:"not null;index"`
	UserID  *uuid.UUID     `json:"user_id"`
	Type    ActivityType   `json:"type" gorm:"not null;size:100"`
	Data    datatypes.JSON `json:"data" gorm:"type:jsonb"`
	
	// Relationships
	Issue Issue `json:"issue,omitempty" gorm:"foreignKey:IssueID"`
	User  *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type Release struct {
	BaseModel
	ProjectID    uuid.UUID  `json:"project_id" gorm:"not null;index"`
	Version      string     `json:"version" gorm:"not null;size:100;index:idx_project_version,unique"`
	Ref          *string    `json:"ref" gorm:"size:255"`
	URL          *string    `json:"url" gorm:"size:500"`
	DateReleased *time.Time `json:"date_released"`
	
	// Relationships
	Project Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}