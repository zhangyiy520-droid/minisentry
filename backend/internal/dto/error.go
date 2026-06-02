package dto

import (
	"time"

	"github.com/google/uuid"
)

// ErrorEventRequest represents the Sentry-compatible error event payload
type ErrorEventRequest struct {
	EventID     *string           `json:"event_id,omitempty"`
	Timestamp   *time.Time        `json:"timestamp,omitempty"`
	Level       *string           `json:"level,omitempty"`
	Logger      *string           `json:"logger,omitempty"`
	Platform    *string           `json:"platform,omitempty"`
	Release     *string           `json:"release,omitempty"`
	Environment *string           `json:"environment,omitempty"`
	ServerName  *string           `json:"server_name,omitempty"`
	Message     *MessageData      `json:"message,omitempty"`
	Exception   *ExceptionData    `json:"exception,omitempty"`
	User        *UserContext      `json:"user,omitempty"`
	Request     *RequestData      `json:"request,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
	Breadcrumbs []BreadcrumbData  `json:"breadcrumbs,omitempty"`
	Contexts    map[string]interface{} `json:"contexts,omitempty"`
	Fingerprint []string          `json:"fingerprint,omitempty"`
	Modules     map[string]string `json:"modules,omitempty"`
}

// MessageData represents structured message information
type MessageData struct {
	Message   string                 `json:"message"`
	Params    []interface{}          `json:"params,omitempty"`
	Formatted *string                `json:"formatted,omitempty"`
}

// ExceptionData represents exception information with stack traces
type ExceptionData struct {
	Values []ExceptionValue `json:"values"`
}

// ExceptionValue represents a single exception with stack trace
type ExceptionValue struct {
	Type       *string        `json:"type,omitempty"`
	Value      *string        `json:"value,omitempty"`
	Module     *string        `json:"module,omitempty"`
	Mechanism  *MechanismData `json:"mechanism,omitempty"`
	Stacktrace *StacktraceData `json:"stacktrace,omitempty"`
}

// MechanismData represents how the exception was captured
type MechanismData struct {
	Type        string                 `json:"type"`
	Description *string                `json:"description,omitempty"`
	HelpLink    *string                `json:"help_link,omitempty"`
	Handled     *bool                  `json:"handled,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// StacktraceData represents stack trace information
type StacktraceData struct {
	Frames         []StackFrame `json:"frames"`
	FramesOmitted  *[]int       `json:"frames_omitted,omitempty"`
}

// StackFrame represents a single stack trace frame
type StackFrame struct {
	Filename      *string                `json:"filename,omitempty"`
	Function      *string                `json:"function,omitempty"`
	Module        *string                `json:"module,omitempty"`
	Lineno        *int                   `json:"lineno,omitempty"`
	Colno         *int                   `json:"colno,omitempty"`
	AbsPath       *string                `json:"abs_path,omitempty"`
	ContextLine   *string                `json:"context_line,omitempty"`
	PreContext    []string               `json:"pre_context,omitempty"`
	PostContext   []string               `json:"post_context,omitempty"`
	InApp         *bool                  `json:"in_app,omitempty"`
	Vars          map[string]interface{} `json:"vars,omitempty"`
	Package       *string                `json:"package,omitempty"`
	Platform      *string                `json:"platform,omitempty"`
	InstructionAddr *string              `json:"instruction_addr,omitempty"`
	Symbol        *string                `json:"symbol,omitempty"`
	SymbolAddr    *string                `json:"symbol_addr,omitempty"`
	ImageAddr     *string                `json:"image_addr,omitempty"`
}

// UserContext represents user information
type UserContext struct {
	ID       *string                `json:"id,omitempty"`
	Email    *string                `json:"email,omitempty"`
	Username *string                `json:"username,omitempty"`
	IPAddress *string               `json:"ip_address,omitempty"`
	Name     *string                `json:"name,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// RequestData represents HTTP request information
type RequestData struct {
	URL         *string               `json:"url,omitempty"`
	Method      *string               `json:"method,omitempty"`
	Data        interface{}           `json:"data,omitempty"`
	QueryString *string               `json:"query_string,omitempty"`
	Headers     map[string]string     `json:"headers,omitempty"`
	Env         map[string]string     `json:"env,omitempty"`
	Cookies     map[string]string     `json:"cookies,omitempty"`
}

// BreadcrumbData represents user action breadcrumbs
type BreadcrumbData struct {
	Type      *string                `json:"type,omitempty"`
	Category  *string                `json:"category,omitempty"`
	Message   *string                `json:"message,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Level     *string                `json:"level,omitempty"`
	Timestamp *time.Time             `json:"timestamp,omitempty"`
}

// TagData represents error tagging information
type TagData struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ErrorEventResponse represents the response after successful error ingestion
type ErrorEventResponse struct {
	ID        string    `json:"id"`
	EventID   string    `json:"event_id"`
	ProjectID uuid.UUID `json:"project_id"`
	IssueID   uuid.UUID `json:"issue_id"`
	CreatedAt time.Time `json:"created_at"`
}


// ErrorProcessingRequest represents internal error processing payload
type ErrorProcessingRequest struct {
	ProjectID uuid.UUID         `json:"project_id"`
	EventData ErrorEventRequest `json:"event_data"`
	ClientIP  string            `json:"client_ip"`
	UserAgent string            `json:"user_agent"`
}

// FingerprintComponents represents the components used for generating fingerprints
type FingerprintComponents struct {
	ErrorType     string   `json:"error_type"`
	ErrorMessage  string   `json:"error_message"`
	StackFrames   []string `json:"stack_frames"`
	Platform      string   `json:"platform"`
	Filename      string   `json:"filename"`
}

// NormalizedErrorData represents cleaned error data ready for storage
type NormalizedErrorData struct {
	EventID         string                 `json:"event_id"`
	ProjectID       uuid.UUID              `json:"project_id"`
	Timestamp       time.Time              `json:"timestamp"`
	Level           string                 `json:"level"`
	Message         *string                `json:"message"`
	ExceptionType   *string                `json:"exception_type"`
	ExceptionValue  *string                `json:"exception_value"`
	StackTrace      []StackFrame           `json:"stack_trace"`
	UserContext     *UserContext           `json:"user_context"`
	RequestData     *RequestData           `json:"request_data"`
	Tags            map[string]string      `json:"tags"`
	ExtraData       map[string]interface{} `json:"extra_data"`
	Breadcrumbs     []BreadcrumbData       `json:"breadcrumbs"`
	Fingerprint     string                 `json:"fingerprint"`
	Environment     string                 `json:"environment"`
	Release         *string                `json:"release"`
	ServerName      *string                `json:"server_name"`
	Platform        string                 `json:"platform"`
}