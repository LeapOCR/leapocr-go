package ocr

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// ValidationError represents validation-specific errors
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation failed: %s", e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{Field: field, Message: message}
}

// SupportedFileExtensions lists all supported file extensions
var SupportedFileExtensions = []string{".pdf"}

// MaxFileSizeBytes represents the maximum allowed file size (50MB)
const MaxFileSizeBytes = 50 * 1024 * 1024

// MaxInstructionsLength represents the maximum length for instructions
const MaxInstructionsLength = 5000

// ValidateFileExtension validates that the file extension is supported
func ValidateFileExtension(filename string) error {
	if filename == "" {
		return NewValidationError("filename", "filename cannot be empty")
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return NewValidationError("filename", "file must have an extension")
	}

	for _, supported := range SupportedFileExtensions {
		if ext == supported {
			return nil
		}
	}

	return NewValidationError("filename", fmt.Sprintf("unsupported file type '%s'. Only PDF files are currently supported", ext))
}

// ValidateURL validates that a URL is properly formatted and uses allowed schemes
func ValidateURL(fileURL string) error {
	if fileURL == "" {
		return NewValidationError("url", "URL cannot be empty")
	}

	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return NewValidationError("url", fmt.Sprintf("invalid URL format: %v", err))
	}

	if parsedURL.Scheme == "" {
		return NewValidationError("url", "URL must include a scheme (http or https)")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return NewValidationError("url", "URL must use http or https scheme")
	}

	if parsedURL.Host == "" {
		return NewValidationError("url", "URL must include a host")
	}

	// Validate file extension from URL path
	if err := ValidateFileExtension(parsedURL.Path); err != nil {
		// Re-wrap with URL context
		if validationErr, ok := err.(*ValidationError); ok {
			return NewValidationError("url", fmt.Sprintf("URL path validation failed: %s", validationErr.Message))
		}
		return NewValidationError("url", fmt.Sprintf("URL path validation failed: %v", err))
	}

	return nil
}

// ValidateFormat validates the OCR format
func ValidateFormat(format Format) error {
	switch format {
	case FormatMarkdown, FormatStructured, FormatPerPageStructured:
		return nil
	case "":
		return NewValidationError("format", "format cannot be empty")
	default:
		return NewValidationError("format", fmt.Sprintf("invalid format '%s'. Valid formats are: %s, %s, %s",
			format, FormatMarkdown, FormatStructured, FormatPerPageStructured))
	}
}

// ValidateModel validates the OCR model name
// Model is optional, but if provided, it should be a non-empty string
// Any model name is allowed (not restricted to predefined constants)
func ValidateModel(model string) error {
	// Model is optional, so empty string is allowed
	if model == "" {
		return nil
	}
	
	// Basic validation: model name should be reasonable length
	if len(model) > 100 {
		return NewValidationError("model", "model name too long. Maximum allowed is 100 characters")
	}
	
	return nil
}

// ValidateInstructions validates custom processing instructions
func ValidateInstructions(instructions string) error {
	if len(instructions) > MaxInstructionsLength {
		return NewValidationError("instructions", fmt.Sprintf("instructions too long (%d characters). Maximum allowed is %d characters",
			len(instructions), MaxInstructionsLength))
	}
	return nil
}

// ValidateSchema validates the extraction schema based on format
func ValidateSchema(schema map[string]interface{}, format Format) error {
	if schema == nil {
		return nil // Schema is optional
	}

	// Schema is not allowed with markdown format
	if format == FormatMarkdown {
		return NewValidationError("schema", "custom schema is not supported with markdown format. Use structured format instead")
	}

	// Basic schema structure validation
	if len(schema) == 0 {
		return NewValidationError("schema", "schema cannot be empty when provided")
	}

	// Validate schema doesn't contain invalid keys or structures
	if err := validateSchemaStructure(schema, ""); err != nil {
		return NewValidationError("schema", err.Error())
	}

	return nil
}

// ValidateCategoryID validates document category ID
func ValidateCategoryID(categoryID string) error {
	if categoryID == "" {
		return nil // Category ID is optional
	}

	// Category ID should be a reasonable length and contain valid characters
	if len(categoryID) > 100 {
		return NewValidationError("categoryID", "category ID too long. Maximum allowed is 100 characters")
	}

	// Basic character validation - only allow alphanumeric, hyphens, and underscores
	for _, char := range categoryID {
		if (char < 'a' || char > 'z') &&
			(char < 'A' || char > 'Z') &&
			(char < '0' || char > '9') &&
			char != '-' && char != '_' {
			return NewValidationError("categoryID", "category ID can only contain letters, numbers, hyphens, and underscores")
		}
	}

	return nil
}

// ValidateProcessingConfig validates the entire processing configuration
func ValidateProcessingConfig(config *processingConfig) error {
	// Validate format
	if err := ValidateFormat(config.format); err != nil {
		return err
	}

	// Validate model (optional)
	if err := ValidateModel(config.model); err != nil {
		return err
	}

	// Validate instructions
	if err := ValidateInstructions(config.instructions); err != nil {
		return err
	}

	// Validate schema (depends on format)
	if err := ValidateSchema(config.schema, config.format); err != nil {
		return err
	}

	// Validate category ID
	if err := ValidateCategoryID(config.categoryID); err != nil {
		return err
	}

	return nil
}

// validateSchemaStructure performs deep validation of schema structure
func validateSchemaStructure(schema map[string]interface{}, path string) error {
	if err := validateSchemaLimits(schema, path); err != nil {
		return err
	}

	for key, value := range schema {
		if err := validateSchemaKey(key, path); err != nil {
			return err
		}

		currentPath := buildPath(path, key)
		if err := validateSchemaValue(value, currentPath); err != nil {
			return err
		}
	}

	return nil
}

func validateSchemaLimits(schema map[string]interface{}, path string) error {
	const maxDepth = 10
	const maxKeys = 100

	if len(path) > 0 && strings.Count(path, ".") > maxDepth {
		return fmt.Errorf("schema nesting too deep at %s. Maximum depth is %d levels", path, maxDepth)
	}

	if len(schema) > maxKeys {
		return fmt.Errorf("too many keys in schema object at %s. Maximum allowed is %d", path, maxKeys)
	}

	return nil
}

func validateSchemaKey(key, path string) error {
	if key == "" {
		return fmt.Errorf("empty key not allowed in schema at %s", path)
	}

	if len(key) > 100 {
		return fmt.Errorf("key '%s' too long. Maximum length is 100 characters", key)
	}

	return nil
}

func validateSchemaValue(value interface{}, currentPath string) error {
	switch v := value.(type) {
	case map[string]interface{}:
		return validateSchemaStructure(v, currentPath)
	case []interface{}:
		return validateSchemaArray(v, currentPath)
	case string:
		return validateSchemaString(v, currentPath)
	case nil, bool, float64:
		return nil
	default:
		return fmt.Errorf("unsupported value type %T at %s", v, currentPath)
	}
}

func validateSchemaArray(arr []interface{}, currentPath string) error {
	if len(arr) > 1000 {
		return fmt.Errorf("array at %s too large. Maximum length is 1000 items", currentPath)
	}

	for i, item := range arr {
		if nested, ok := item.(map[string]interface{}); ok {
			if err := validateSchemaStructure(nested, fmt.Sprintf("%s[%d]", currentPath, i)); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateSchemaString(s string, currentPath string) error {
	if len(s) > 1000 {
		return fmt.Errorf("string value at %s too long. Maximum length is 1000 characters", currentPath)
	}
	return nil
}

func buildPath(path, key string) string {
	if path == "" {
		return key
	}
	return path + "." + key
}
