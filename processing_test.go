package ocr

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestProcessURL_Validation(t *testing.T) {
	// Create a minimal SDK instance for testing
	// We don't need a real client since validation happens before API calls
	sdk := &SDK{
		config: &Config{
			APIKey:  "test-key",
			BaseURL: "https://api.example.com",
		},
	}

	tests := []struct {
		name        string
		url         string
		opts        []ProcessingOption
		expectError string
	}{
		{
			name:        "invalid URL - empty",
			url:         "",
			opts:        nil,
			expectError: "validation failed for url: URL cannot be empty",
		},
		{
			name:        "invalid URL - no scheme",
			url:         "example.com/document.pdf",
			opts:        nil,
			expectError: "validation failed for url: URL must include a scheme",
		},
		{
			name:        "invalid URL - wrong file type",
			url:         "https://example.com/document.txt",
			opts:        nil,
			expectError: "unsupported file type '.txt'",
		},
		{
			name:        "invalid format",
			url:         "https://example.com/document.pdf",
			opts:        []ProcessingOption{WithFormat(Format("invalid"))},
			expectError: "invalid format 'invalid'",
		},
		{
			name:        "model too long",
			url:         "https://example.com/document.pdf",
			opts:        []ProcessingOption{WithModelString(strings.Repeat("a", 101))},
			expectError: "model name too long",
		},
		{
			name: "schema with markdown format",
			url:  "https://example.com/document.pdf",
			opts: []ProcessingOption{
				WithFormat(FormatMarkdown),
				WithSchema(map[string]interface{}{"title": "string"}),
			},
			expectError: "custom schema is not supported with markdown format",
		},
		{
			name: "instructions too long",
			url:  "https://example.com/document.pdf",
			opts: []ProcessingOption{
				WithInstructions(strings.Repeat("a", MaxInstructionsLength+1)),
			},
			expectError: "instructions too long",
		},
		{
			name: "invalid category ID",
			url:  "https://example.com/document.pdf",
			opts: []ProcessingOption{
				WithCategoryID("invalid category"),
			},
			expectError: "category ID can only contain letters, numbers, hyphens, and underscores",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sdk.ProcessURL(context.Background(), tt.url, tt.opts...)

			if err == nil {
				t.Fatal("expected validation error, got none")
			}

			sdkErr, ok := err.(*SDKError)
			if !ok {
				t.Fatalf("expected SDKError, got %T", err)
			}

			if !sdkErr.IsValidationError() {
				t.Errorf("expected validation error type, got %s", sdkErr.Type)
			}

			if !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("expected error containing %q, got %q", tt.expectError, err.Error())
			}
		})
	}
}

func TestProcessFile_Validation(t *testing.T) {
	// Test validation logic directly for ProcessFile
	// We only test validation errors that occur before API calls

	tests := []struct {
		name        string
		filename    string
		content     []byte
		opts        []ProcessingOption
		expectError string
	}{
		{
			name:        "invalid filename - empty",
			filename:    "",
			content:     []byte("fake pdf content"),
			opts:        nil,
			expectError: "validation failed for filename: filename cannot be empty",
		},
		{
			name:        "invalid filename - no extension",
			filename:    "document",
			content:     []byte("fake pdf content"),
			opts:        nil,
			expectError: "file must have an extension",
		},
		{
			name:        "invalid filename - wrong extension",
			filename:    "document.txt",
			content:     []byte("fake pdf content"),
			opts:        nil,
			expectError: "unsupported file type '.txt'",
		},
		{
			name:        "invalid processing config",
			filename:    "document.pdf",
			content:     []byte("fake pdf content"),
			opts:        []ProcessingOption{WithFormat(Format("invalid"))},
			expectError: "invalid format 'invalid'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test filename validation directly
			if tt.filename != "" {
				err := ValidateFileExtension(tt.filename)
				if strings.Contains(tt.expectError, "filename") {
					if err == nil {
						t.Fatal("expected filename validation error, got none")
					}
					if !strings.Contains(err.Error(), strings.Split(tt.expectError, ": ")[1]) {
						t.Errorf("expected error containing filename validation, got %q", err.Error())
					}
					return
				}
			}

			// Test processing config validation
			config := applyProcessingOptions(tt.opts)
			err := ValidateProcessingConfig(config)
			if strings.Contains(tt.expectError, "format") || strings.Contains(tt.expectError, "model") {
				if err == nil {
					t.Fatal("expected processing config validation error, got none")
				}
				if !strings.Contains(err.Error(), strings.Split(tt.expectError, " ")[1]) {
					t.Errorf("expected error containing config validation, got %q", err.Error())
				}
			}
		})
	}
}

func TestProcessingOptions_Validation(t *testing.T) {
	tests := []struct {
		name        string
		opts        []ProcessingOption
		expectError string
	}{
		{
			name: "valid structured with schema",
			opts: []ProcessingOption{
				WithFormat(FormatStructured),
				WithModel(ModelStandardV1),
				WithSchema(map[string]interface{}{
					"title":  "string",
					"amount": "number",
				}),
				WithInstructions("Extract title and amount"),
				WithCategoryID("invoice"),
			},
			expectError: "",
		},
		{
			name: "valid markdown without schema",
			opts: []ProcessingOption{
				WithFormat(FormatMarkdown),
				WithModel(ModelStandardV1),
				WithInstructions("Extract all text"),
			},
			expectError: "",
		},
		{
			name: "invalid - empty schema object",
			opts: []ProcessingOption{
				WithFormat(FormatStructured),
				WithSchema(map[string]interface{}{}),
			},
			expectError: "schema cannot be empty when provided",
		},
		{
			name: "invalid - deeply nested schema",
			opts: []ProcessingOption{
				WithFormat(FormatStructured),
				WithSchema(createDeeplyNestedSchema(12)), // Exceeds max depth
			},
			expectError: "nesting too deep",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := applyProcessingOptions(tt.opts)
			err := ValidateProcessingConfig(config)

			if tt.expectError == "" {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Fatal("expected validation error, got none")
				}
				if !strings.Contains(err.Error(), tt.expectError) {
					t.Errorf("expected error containing %q, got %q", tt.expectError, err.Error())
				}
			}
		})
	}
}

// Helper function to create deeply nested schema for testing
func createDeeplyNestedSchema(depth int) map[string]interface{} {
	if depth == 0 {
		return map[string]interface{}{"value": "string"}
	}
	return map[string]interface{}{
		"level": createDeeplyNestedSchema(depth - 1),
	}
}

func TestFileSizeValidation(t *testing.T) {
	// Test file size validation in upload function
	sdk := &SDK{
		config: &Config{
			APIKey:  "test-key",
			BaseURL: "https://api.example.com",
		},
	}

	tests := []struct {
		name        string
		content     []byte
		expectError string
	}{
		{
			name:        "empty file",
			content:     []byte{},
			expectError: "file cannot be empty",
		},
		{
			name:        "file too large",
			content:     bytes.Repeat([]byte("a"), MaxFileSizeBytes+1),
			expectError: "file size",
		},
		{
			name:        "valid file size",
			content:     bytes.Repeat([]byte("a"), 1024), // 1KB
			expectError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := bytes.NewReader(tt.content)
			err := sdk.uploadFile(context.Background(), "https://example.com/upload", file, "document.pdf")

			if tt.expectError == "" {
				// We expect this to fail due to invalid upload URL, but not due to file size
				if err != nil && strings.Contains(err.Error(), "file size") {
					t.Errorf("unexpected file size validation error: %v", err)
				}
			} else {
				if err == nil {
					t.Fatal("expected validation error, got none")
				}
				if !strings.Contains(err.Error(), tt.expectError) {
					t.Errorf("expected error containing %q, got %q", tt.expectError, err.Error())
				}
			}
		})
	}
}

func TestValidationError_Integration(t *testing.T) {
	// Test that validation errors are properly wrapped in SDK errors
	sdk := &SDK{
		config: &Config{
			APIKey:  "test-key",
			BaseURL: "https://api.example.com",
		},
	}

	_, err := sdk.ProcessURL(context.Background(), "invalid-url")

	if err == nil {
		t.Fatal("expected error, got none")
	}

	// Check that it's an SDK error
	sdkErr, ok := err.(*SDKError)
	if !ok {
		t.Fatalf("expected SDKError, got %T", err)
	}

	// Check that it's a validation error
	if !sdkErr.IsValidationError() {
		t.Errorf("expected validation error, got %s", sdkErr.Type)
	}

	// Check that the underlying cause is a ValidationError
	cause := sdkErr.Unwrap()
	if cause == nil {
		t.Fatal("expected underlying cause, got none")
	}

	validationErr, ok := cause.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError as cause, got %T", cause)
	}

	if validationErr.Field != "url" {
		t.Errorf("expected field 'url', got '%s'", validationErr.Field)
	}
}
