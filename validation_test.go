package ocr

import (
	"fmt"
	"strings"
	"testing"
)

func TestValidateFileExtension(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		expectError bool
	}{
		{"valid PDF", "document.pdf", false},
		{"valid PDF uppercase", "document.PDF", false},
		{"invalid txt", "document.txt", true},
		{"invalid jpg", "document.jpg", true},
		{"invalid png", "document.png", true},
		{"invalid docx", "document.docx", true},
		{"empty filename", "", true},
		{"no extension", "document", true},
		{"just extension", ".pdf", false},
		{"multiple extensions", "document.backup.pdf", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileExtension(tt.filename)
			if tt.expectError && err == nil {
				t.Errorf("expected error for filename %q, got none", tt.filename)
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error for filename %q, got: %v", tt.filename, err)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{"valid https PDF", "https://example.com/document.pdf", false},
		{"valid http PDF", "http://example.com/document.pdf", false},
		{"valid with port", "https://example.com:8080/document.pdf", false},
		{"valid with path", "https://example.com/path/to/document.pdf", false},
		{"valid with query", "https://example.com/document.pdf?version=1", false},
		{"invalid scheme", "ftp://example.com/document.pdf", true},
		{"no scheme", "example.com/document.pdf", true},
		{"empty URL", "", true},
		{"no host", "https:///document.pdf", true},
		{"invalid file type", "https://example.com/document.txt", true},
		{"no file extension", "https://example.com/document", true},
		{"malformed URL", "https://[invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if tt.expectError && err == nil {
				t.Errorf("expected error for URL %q, got none", tt.url)
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error for URL %q, got: %v", tt.url, err)
			}
		})
	}
}

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		name        string
		format      Format
		expectError bool
	}{
		{"valid markdown", FormatMarkdown, false},
		{"valid structured", FormatStructured, false},
		{"valid per_page_structured", FormatPerPageStructured, false},
		{"invalid format", Format("invalid"), true},
		{"empty format", Format(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFormat(tt.format)
			if tt.expectError && err == nil {
				t.Errorf("expected error for format %q, got none", tt.format)
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error for format %q, got: %v", tt.format, err)
			}
		})
	}
}

func TestValidateModel(t *testing.T) {
	tests := []struct {
		name        string
		model       string
		expectError bool
	}{
		{"valid standard-v1", "standard-v1", false},
		{"valid english-pro-v1", "english-pro-v1", false},
		{"valid pro-v1", "pro-v1", false},
		{"valid custom model", "custom-model-v2", false},
		{"empty model (optional)", "", false},
		{"model too long", strings.Repeat("a", 101), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateModel(tt.model)
			if tt.expectError && err == nil {
				t.Errorf("expected error for model %q, got none", tt.model)
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error for model %q, got: %v", tt.model, err)
			}
		})
	}
}

func TestValidateInstructions(t *testing.T) {
	tests := []struct {
		name         string
		instructions string
		expectError  bool
	}{
		{"empty instructions", "", false},
		{"short instructions", "Extract the title", false},
		{"normal instructions", "Extract all invoice details including date, amount, and vendor", false},
		{"max length instructions", strings.Repeat("a", MaxInstructionsLength), false},
		{"too long instructions", strings.Repeat("a", MaxInstructionsLength+1), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInstructions(tt.instructions)
			if tt.expectError && err == nil {
				t.Errorf("expected error for instructions of length %d, got none", len(tt.instructions))
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error for instructions of length %d, got: %v", len(tt.instructions), err)
			}
		})
	}
}

func TestValidateSchema(t *testing.T) {
	tests := []struct {
		name        string
		schema      map[string]interface{}
		format      Format
		expectError bool
	}{
		{
			name:        "nil schema with structured",
			schema:      nil,
			format:      FormatStructured,
			expectError: false,
		},
		{
			name:        "valid schema with structured",
			schema:      map[string]interface{}{"title": "string", "amount": "number"},
			format:      FormatStructured,
			expectError: false,
		},
		{
			name:        "schema with markdown format",
			schema:      map[string]interface{}{"title": "string"},
			format:      FormatMarkdown,
			expectError: true,
		},
		{
			name:        "empty schema object",
			schema:      map[string]interface{}{},
			format:      FormatStructured,
			expectError: true,
		},
		{
			name: "nested schema",
			schema: map[string]interface{}{
				"invoice": map[string]interface{}{
					"date":   "string",
					"amount": "number",
				},
			},
			format:      FormatStructured,
			expectError: false,
		},
		{
			name:        "schema with array",
			schema:      map[string]interface{}{"items": []interface{}{"item1", "item2"}},
			format:      FormatStructured,
			expectError: false,
		},
		{
			name:        "schema with very long key",
			schema:      map[string]interface{}{strings.Repeat("a", 101): "string"},
			format:      FormatStructured,
			expectError: true,
		},
		{
			name:        "schema with empty key",
			schema:      map[string]interface{}{"": "string"},
			format:      FormatStructured,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSchema(tt.schema, tt.format)
			if tt.expectError && err == nil {
				t.Errorf("expected error for schema validation, got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error for schema validation, got: %v", err)
			}
		})
	}
}

func TestValidateCategoryID(t *testing.T) {
	tests := []struct {
		name        string
		categoryID  string
		expectError bool
	}{
		{"empty category ID", "", false},
		{"valid alphanumeric", "invoice123", false},
		{"valid with hyphens", "invoice-type", false},
		{"valid with underscores", "invoice_type", false},
		{"valid mixed", "invoice_type-123", false},
		{"invalid with spaces", "invoice type", true},
		{"invalid with symbols", "invoice@type", true},
		{"invalid with dots", "invoice.type", true},
		{"too long", strings.Repeat("a", 101), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCategoryID(tt.categoryID)
			if tt.expectError && err == nil {
				t.Errorf("expected error for category ID %q, got none", tt.categoryID)
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error for category ID %q, got: %v", tt.categoryID, err)
			}
		})
	}
}

func TestValidateProcessingConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *processingConfig
		expectError bool
	}{
		{
			name: "valid structured config",
			config: &processingConfig{
				format:       FormatStructured,
				model:         string(ModelStandardV1),
				schema:       map[string]interface{}{"title": "string"},
				instructions: "Extract the title",
				categoryID:   "invoice",
			},
			expectError: false,
		},
		{
			name: "valid markdown config without schema",
			config: &processingConfig{
				format:       FormatMarkdown,
				model:         string(ModelStandardV1),
				instructions: "Extract all text",
			},
			expectError: false,
		},
		{
			name: "invalid - schema with markdown",
			config: &processingConfig{
				format: FormatMarkdown,
				model:  string(ModelStandardV1),
				schema: map[string]interface{}{"title": "string"},
			},
			expectError: true,
		},
		{
			name: "invalid format",
			config: &processingConfig{
				format: Format("invalid"),
				model:  string(ModelStandardV1),
			},
			expectError: true,
		},
		{
			name: "invalid model too long",
			config: &processingConfig{
				format: FormatStructured,
				model:  strings.Repeat("a", 101),
			},
			expectError: true,
		},
		{
			name: "invalid instructions too long",
			config: &processingConfig{
				format:       FormatStructured,
				model:         string(ModelStandardV1),
				instructions: strings.Repeat("a", MaxInstructionsLength+1),
			},
			expectError: true,
		},
		{
			name: "invalid category ID",
			config: &processingConfig{
				format:     FormatStructured,
				model:      string(ModelStandardV1),
				categoryID: "invalid category",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProcessingConfig(tt.config)
			if tt.expectError && err == nil {
				t.Errorf("expected error for config validation, got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error for config validation, got: %v", err)
			}
		})
	}
}

func TestValidateSchemaStructure_DeepNesting(t *testing.T) {
	// Test maximum nesting depth
	deeply_nested := map[string]interface{}{}
	current := deeply_nested
	for i := 0; i < 12; i++ { // Exceeds max depth of 10
		next := map[string]interface{}{}
		current["level"] = next
		current = next
	}

	err := validateSchemaStructure(deeply_nested, "")
	if err == nil {
		t.Error("expected error for deeply nested schema, got none")
	}
	if !strings.Contains(err.Error(), "nesting too deep") {
		t.Errorf("expected nesting error, got: %v", err)
	}
}

func TestValidateSchemaStructure_TooManyKeys(t *testing.T) {
	// Test too many keys
	schema := map[string]interface{}{}
	for i := 0; i < 101; i++ { // Exceeds max keys of 100
		schema[fmt.Sprintf("key%d", i)] = "string"
	}

	err := validateSchemaStructure(schema, "")
	if err == nil {
		t.Error("expected error for schema with too many keys, got none")
	}
	if !strings.Contains(err.Error(), "too many keys") {
		t.Errorf("expected too many keys error, got: %v", err)
	}
}

func TestValidateSchemaStructure_LargeArray(t *testing.T) {
	// Test array that's too large
	largeArray := make([]interface{}, 1001) // Exceeds max of 1000
	for i := range largeArray {
		largeArray[i] = "item"
	}

	schema := map[string]interface{}{
		"items": largeArray,
	}

	err := validateSchemaStructure(schema, "")
	if err == nil {
		t.Error("expected error for schema with large array, got none")
	}
	if !strings.Contains(err.Error(), "array") && !strings.Contains(err.Error(), "too large") {
		t.Errorf("expected array size error, got: %v", err)
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		message  string
		expected string
	}{
		{
			name:     "with field",
			field:    "filename",
			message:  "invalid extension",
			expected: "validation failed for filename: invalid extension",
		},
		{
			name:     "without field",
			field:    "",
			message:  "general validation error",
			expected: "validation failed: general validation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewValidationError(tt.field, tt.message)
			if err.Error() != tt.expected {
				t.Errorf("expected error message %q, got %q", tt.expected, err.Error())
			}
		})
	}
}
