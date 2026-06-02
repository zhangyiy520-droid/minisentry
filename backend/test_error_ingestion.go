package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"minisentry/internal/dto"
)

func main() {
	// Test error ingestion with sample JavaScript error
	testErrorIngestion()
}

func testErrorIngestion() {
	// Sample error event (Sentry-compatible format)
	errorEvent := dto.ErrorEventRequest{
		Timestamp:   &[]time.Time{time.Now()}[0],
		Level:       &[]string{"error"}[0],
		Platform:    &[]string{"javascript"}[0],
		Environment: &[]string{"development"}[0],
		Message: &dto.MessageData{
			Message: "Uncaught TypeError: Cannot read property 'foo' of undefined",
		},
		Exception: &dto.ExceptionData{
			Values: []dto.ExceptionValue{
				{
					Type:  &[]string{"TypeError"}[0],
					Value: &[]string{"Cannot read property 'foo' of undefined"}[0],
					Stacktrace: &dto.StacktraceData{
						Frames: []dto.StackFrame{
							{
								Filename:    &[]string{"http://localhost:3000/static/js/main.js"}[0],
								Function:    &[]string{"handleClick"}[0],
								Lineno:      &[]int{42}[0],
								Colno:       &[]int{15}[0],
								InApp:       &[]bool{true}[0],
								ContextLine: &[]string{"    obj.foo.bar()"}[0],
							},
							{
								Filename:    &[]string{"http://localhost:3000/static/js/main.js"}[0],
								Function:    &[]string{"onClick"}[0],
								Lineno:      &[]int{15}[0],
								Colno:       &[]int{8}[0],
								InApp:       &[]bool{true}[0],
								ContextLine: &[]string{"    handleClick(event)"}[0],
							},
						},
					},
				},
			},
		},
		User: &dto.UserContext{
			ID:    &[]string{"12345"}[0],
			Email: &[]string{"test@example.com"}[0],
		},
		Tags: map[string]string{
			"browser": "Chrome",
			"version": "91.0.4472.124",
		},
		Extra: map[string]interface{}{
			"component": "UserProfile",
			"action":    "click",
		},
		Breadcrumbs: []dto.BreadcrumbData{
			{
				Type:      &[]string{"navigation"}[0],
				Category:  &[]string{"ui.click"}[0],
				Message:   &[]string{"User clicked profile button"}[0],
				Timestamp: &[]time.Time{time.Now().Add(-10 * time.Second)}[0],
				Level:     &[]string{"info"}[0],
			},
		},
	}

	// Convert to JSON
	jsonData, err := json.Marshal(errorEvent)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	// Test different authentication methods
	testCases := []struct {
		name        string
		url         string
		headers     map[string]string
		description string
	}{
		{
			name: "DSN Query Parameter",
			url:  "http://localhost:8080/api/v1/errors/ingest?dsn=https://32characterhexkey123456789abcdef@localhost:8080/project-uuid-here",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			description: "Test with DSN in query parameter",
		},
		{
			name: "Public Key Authentication",
			url:  "http://localhost:8080/api/v1/errors/ingest",
			headers: map[string]string{
				"Content-Type": "application/json",
				"Authorization": "Bearer 32characterhexkey123456789abcdef",
			},
			description: "Test with public key in Authorization header",
		},
		{
			name: "Sentry Auth Header",
			url:  "http://localhost:8080/api/v1/errors/ingest",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Sentry-Auth": "Sentry sentry_version=7, sentry_client=test-client/1.0, sentry_key=32characterhexkey123456789abcdef",
			},
			description: "Test with Sentry-style auth header",
		},
	}

	for _, tc := range testCases {
		fmt.Printf("\n=== %s ===\n", tc.name)
		fmt.Printf("Description: %s\n", tc.description)
		
		// Create request
		req, err := http.NewRequest("POST", tc.url, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			continue
		}

		// Set headers
		for key, value := range tc.headers {
			req.Header.Set(key, value)
		}

		// Make request
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error making request: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		// Read response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			continue
		}

		fmt.Printf("Status: %s\n", resp.Status)
		fmt.Printf("Response: %s\n", string(body))
	}

	// Test multiple errors for grouping
	fmt.Printf("\n=== Testing Error Grouping ===\n")
	testErrorGrouping()
}

func testErrorGrouping() {
	// Test with similar errors that should be grouped together
	similarErrors := []dto.ErrorEventRequest{
		{
			Timestamp: &[]time.Time{time.Now()}[0],
			Level:     &[]string{"error"}[0],
			Platform:  &[]string{"javascript"}[0],
			Exception: &dto.ExceptionData{
				Values: []dto.ExceptionValue{
					{
						Type:  &[]string{"TypeError"}[0],
						Value: &[]string{"Cannot read property 'foo' of undefined"}[0],
						Stacktrace: &dto.StacktraceData{
							Frames: []dto.StackFrame{
								{
									Filename: &[]string{"main.js"}[0],
									Function: &[]string{"handleClick"}[0],
									Lineno:   &[]int{42}[0],
									InApp:    &[]bool{true}[0],
								},
							},
						},
					},
				},
			},
		},
		{
			Timestamp: &[]time.Time{time.Now()}[0],
			Level:     &[]string{"error"}[0],
			Platform:  &[]string{"javascript"}[0],
			Exception: &dto.ExceptionData{
				Values: []dto.ExceptionValue{
					{
						Type:  &[]string{"TypeError"}[0],
						Value: &[]string{"Cannot read property 'foo' of undefined"}[0],
						Stacktrace: &dto.StacktraceData{
							Frames: []dto.StackFrame{
								{
									Filename: &[]string{"main.js"}[0],
									Function: &[]string{"handleClick"}[0],
									Lineno:   &[]int{42}[0],
									InApp:    &[]bool{true}[0],
								},
							},
						},
					},
				},
			},
		},
	}

	for i, errorEvent := range similarErrors {
		fmt.Printf("Sending similar error #%d\n", i+1)
		
		jsonData, err := json.Marshal(errorEvent)
		if err != nil {
			fmt.Printf("Error marshaling JSON: %v\n", err)
			continue
		}

		req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/errors/ingest", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer 32characterhexkey123456789abcdef")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error making request: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			continue
		}

		fmt.Printf("Response #%d: %s - %s\n", i+1, resp.Status, string(body))
	}
}