package services

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"minisentry/internal/dto"
)

type FingerprintService struct {
	// Configuration for fingerprinting
	maxStackFrames       int
	normalizeURLs        bool
	normalizeFilePaths   bool
	ignoreLocalVariables bool
}

// NewFingerprintService creates a new fingerprint service
func NewFingerprintService() *FingerprintService {
	return &FingerprintService{
		maxStackFrames:       5,  // Use top 5 stack frames for fingerprinting
		normalizeURLs:        true,
		normalizeFilePaths:   true,
		ignoreLocalVariables: true,
	}
}

// GenerateErrorFingerprint creates a fingerprint for error grouping
func (fs *FingerprintService) GenerateErrorFingerprint(errorData *dto.NormalizedErrorData) string {
	components := fs.extractFingerprintComponents(errorData)
	fingerprintString := fs.buildFingerprintString(components)
	return fs.hashFingerprint(fingerprintString)
}

// extractFingerprintComponents extracts the key components for fingerprinting
func (fs *FingerprintService) extractFingerprintComponents(errorData *dto.NormalizedErrorData) *dto.FingerprintComponents {
	components := &dto.FingerprintComponents{
		Platform: errorData.Platform,
	}

	// Extract error type and message
	if errorData.ExceptionType != nil {
		components.ErrorType = *errorData.ExceptionType
	}
	
	if errorData.ExceptionValue != nil {
		components.ErrorMessage = fs.normalizeErrorMessage(*errorData.ExceptionValue)
	} else if errorData.Message != nil {
		components.ErrorMessage = fs.normalizeErrorMessage(*errorData.Message)
	}

	// Extract stack frame information
	components.StackFrames = fs.extractStackFrameSignatures(errorData.StackTrace)
	
	// Extract primary filename
	if len(errorData.StackTrace) > 0 && errorData.StackTrace[0].Filename != nil {
		components.Filename = fs.normalizeFilename(*errorData.StackTrace[0].Filename)
	}

	return components
}

// normalizeErrorMessage removes dynamic parts from error messages
func (fs *FingerprintService) normalizeErrorMessage(message string) string {
	// Remove common dynamic patterns
	patterns := []struct {
		regex       string
		replacement string
	}{
		// URLs
		{`https?://[^\s]+`, "<URL>"},
		// Email addresses
		{`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`, "<EMAIL>"},
		// UUIDs
		{`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`, "<UUID>"},
		// Hexadecimal IDs
		{`0x[0-9a-fA-F]+`, "<HEX_ID>"},
		// Numbers (but preserve structure)
		{`\b\d{4,}\b`, "<NUMBER>"},
		// File paths with line numbers
		{`:\d+:\d+`, ":<LINE>:<COL>"},
		// Memory addresses
		{`@[0-9a-fA-F]+`, "@<ADDR>"},
		// Timestamps
		{`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`, "<TIMESTAMP>"},
		// IP addresses
		{`\b(?:\d{1,3}\.){3}\d{1,3}\b`, "<IP>"},
		// JavaScript object references
		{`\[object \w+\]`, "[object <TYPE>]"},
		// Dynamic property access
		{`\[\d+\]`, "[<INDEX>]"},
	}

	normalized := message
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern.regex)
		normalized = re.ReplaceAllString(normalized, pattern.replacement)
	}

	// Trim whitespace and normalize spacing
	normalized = strings.TrimSpace(normalized)
	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ")

	return normalized
}

// normalizeFilename removes dynamic parts from filenames
func (fs *FingerprintService) normalizeFilename(filename string) string {
	if !fs.normalizeFilePaths {
		return filename
	}

	// Remove query parameters and fragments
	if idx := strings.Index(filename, "?"); idx != -1 {
		filename = filename[:idx]
	}
	if idx := strings.Index(filename, "#"); idx != -1 {
		filename = filename[:idx]
	}

	// Normalize common patterns
	patterns := []struct {
		regex       string
		replacement string
	}{
		// Remove webpack module IDs
		{`webpack:///\./`, ""},
		// Remove source map references
		{`\?[a-f0-9]+$`, ""},
		// Normalize chunk names
		{`\.\w+\.chunk\.js$`, ".chunk.js"},
		// Remove hash from bundled files
		{`\.[a-f0-9]{8,}\.js$`, ".js"},
		{`\.[a-f0-9]{8,}\.css$`, ".css"},
		// Normalize node_modules paths
		{`.*/node_modules/`, "node_modules/"},
	}

	normalized := filename
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern.regex)
		normalized = re.ReplaceAllString(normalized, pattern.replacement)
	}

	return normalized
}

// extractStackFrameSignatures extracts meaningful signatures from stack frames
func (fs *FingerprintService) extractStackFrameSignatures(stackTrace []dto.StackFrame) []string {
	var signatures []string
	frameCount := 0

	for _, frame := range stackTrace {
		// Skip frames that are not in-app (if specified)
		if frame.InApp != nil && !*frame.InApp {
			continue
		}

		// Stop if we've collected enough frames
		if frameCount >= fs.maxStackFrames {
			break
		}

		signature := fs.buildFrameSignature(frame)
		if signature != "" {
			signatures = append(signatures, signature)
			frameCount++
		}
	}

	return signatures
}

// buildFrameSignature creates a signature for a single stack frame
func (fs *FingerprintService) buildFrameSignature(frame dto.StackFrame) string {
	var parts []string

	// Add function name if available
	if frame.Function != nil && *frame.Function != "" {
		funcName := fs.normalizeFunctionName(*frame.Function)
		if funcName != "" {
			parts = append(parts, fmt.Sprintf("func:%s", funcName))
		}
	}

	// Add filename if available
	if frame.Filename != nil && *frame.Filename != "" {
		filename := fs.normalizeFilename(*frame.Filename)
		if filename != "" {
			parts = append(parts, fmt.Sprintf("file:%s", filename))
		}
	}

	// Add module if available and different from filename
	if frame.Module != nil && *frame.Module != "" {
		module := *frame.Module
		if frame.Filename == nil || *frame.Filename != module {
			parts = append(parts, fmt.Sprintf("module:%s", module))
		}
	}

	return strings.Join(parts, "|")
}

// normalizeFunctionName normalizes function names for consistent grouping
func (fs *FingerprintService) normalizeFunctionName(funcName string) string {
	// Remove anonymous function markers
	if strings.Contains(funcName, "anonymous") {
		return "<anonymous>"
	}

	// Normalize arrow functions
	if strings.Contains(funcName, "=>") {
		return "<arrow>"
	}

	// Remove object references and dynamic parts
	patterns := []struct {
		regex       string
		replacement string
	}{
		// Remove object instance references
		{`Object\..*?\.`, "Object."},
		// Remove eval contexts
		{`eval at .*? \(`, "eval("},
		// Normalize bound functions
		{`bound `, ""},
		// Remove webpack module references
		{`__webpack_require__.*?`, "<webpack>"},
	}

	normalized := funcName
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern.regex)
		normalized = re.ReplaceAllString(normalized, pattern.replacement)
	}

	return normalized
}

// buildFingerprintString creates the final fingerprint string
func (fs *FingerprintService) buildFingerprintString(components *dto.FingerprintComponents) string {
	var parts []string

	// Add platform
	if components.Platform != "" {
		parts = append(parts, fmt.Sprintf("platform:%s", components.Platform))
	}

	// Add error type
	if components.ErrorType != "" {
		parts = append(parts, fmt.Sprintf("type:%s", components.ErrorType))
	}

	// Add normalized error message
	if components.ErrorMessage != "" {
		parts = append(parts, fmt.Sprintf("message:%s", components.ErrorMessage))
	}

	// Add primary filename
	if components.Filename != "" {
		parts = append(parts, fmt.Sprintf("file:%s", components.Filename))
	}

	// Add stack frame signatures
	if len(components.StackFrames) > 0 {
		// Sort to ensure consistent ordering
		sortedFrames := make([]string, len(components.StackFrames))
		copy(sortedFrames, components.StackFrames)
		sort.Strings(sortedFrames)
		
		parts = append(parts, fmt.Sprintf("stack:%s", strings.Join(sortedFrames, ",")))
	}

	return strings.Join(parts, "||")
}

// hashFingerprint creates a SHA256 hash of the fingerprint string
func (fs *FingerprintService) hashFingerprint(fingerprintString string) string {
	hash := sha256.Sum256([]byte(fingerprintString))
	return fmt.Sprintf("%x", hash)
}

// CustomFingerprint allows for custom fingerprinting rules
func (fs *FingerprintService) CustomFingerprint(errorData *dto.NormalizedErrorData, customRules []string) string {
	if len(customRules) == 0 {
		return fs.GenerateErrorFingerprint(errorData)
	}

	// Apply custom fingerprinting rules
	var parts []string
	for _, rule := range customRules {
		switch rule {
		case "{{ default }}":
			// Use default fingerprinting
			return fs.GenerateErrorFingerprint(errorData)
		case "{{ error.type }}":
			if errorData.ExceptionType != nil {
				parts = append(parts, *errorData.ExceptionType)
			}
		case "{{ error.value }}":
			if errorData.ExceptionValue != nil {
				parts = append(parts, fs.normalizeErrorMessage(*errorData.ExceptionValue))
			}
		case "{{ transaction }}":
			// Use the culprit/transaction name if available
			if len(errorData.StackTrace) > 0 && errorData.StackTrace[0].Function != nil {
				parts = append(parts, *errorData.StackTrace[0].Function)
			}
		default:
			// Treat as literal string
			parts = append(parts, rule)
		}
	}

	customString := strings.Join(parts, "||")
	return fs.hashFingerprint(customString)
}

// SimilarityScore calculates similarity between two fingerprints
func (fs *FingerprintService) SimilarityScore(fingerprint1, fingerprint2 string) float64 {
	if fingerprint1 == fingerprint2 {
		return 1.0
	}

	// For now, exact match only
	// In the future, we could implement fuzzy matching for similar errors
	return 0.0
}