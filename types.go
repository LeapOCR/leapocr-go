package ocr

import (
	"time"
)

// Format represents the OCR output format
type Format string

const (
	// FormatMarkdown outputs text as markdown
	FormatMarkdown Format = "markdown"
	// FormatStructured outputs structured data as JSON
	FormatStructured Format = "structured"
	// FormatPerPageStructured outputs structured data per page
	FormatPerPageStructured Format = "per_page_structured"
)

// Model represents the OCR model to use for processing
type Model string

const (
	// ModelStandardV1 is the baseline model that handles all cases
	// Credits per page: 1, Priority: 1
	ModelStandardV1 Model = "standard-v1"
	// ModelEnglishProV1 is premium quality for English documents only
	// Credits per page: 2, Priority: 4
	ModelEnglishProV1 Model = "english-pro-v1"
	// ModelProV1 is the highest quality model that handles all cases
	// Credits per page: 5, Priority: 5
	ModelProV1 Model = "pro-v1"
)

// OCRResult represents the final result of OCR processing
type OCRResult struct {
	Text     string                 `json:"text"`
	Data     map[string]interface{} `json:"data"`
	Pages    []PageResult           `json:"pages"`
	Credits  int                    `json:"credits"`
	Duration time.Duration          `json:"duration"`
	JobID    string                 `json:"job_id"`
	Status   string                 `json:"status"`
}

// PageResult represents a single page result
type PageResult struct {
	PageNumber int                    `json:"page_number"`
	Text       string                 `json:"text"`
	Data       map[string]interface{} `json:"data"`
}

// ProcessingOption configures OCR processing
type ProcessingOption func(*processingConfig)

// processingConfig holds all processing configuration
type processingConfig struct {
	format       Format
	model        string // Can be a Model constant or any custom model string
	schema       map[string]interface{}
	instructions string
	categoryID   string
}

// WithFormat sets the output format
func WithFormat(format Format) ProcessingOption {
	return func(c *processingConfig) {
		c.format = format
	}
}

// WithModel sets the OCR model to use for processing
// You can use one of the predefined constants (ModelStandardV1, ModelEnglishProV1, ModelProV1)
// or pass any custom model name as a string
func WithModel(model Model) ProcessingOption {
	return func(c *processingConfig) {
		c.model = string(model)
	}
}

// WithModelString sets a custom model name as a string
// This allows using any model name not defined as a constant
func WithModelString(model string) ProcessingOption {
	return func(c *processingConfig) {
		c.model = model
	}
}

// WithSchema sets a custom extraction schema
func WithSchema(schema map[string]interface{}) ProcessingOption {
	return func(c *processingConfig) {
		c.schema = schema
	}
}

// WithInstructions sets custom processing instructions
func WithInstructions(instructions string) ProcessingOption {
	return func(c *processingConfig) {
		c.instructions = instructions
	}
}

// WithCategoryID sets the document category ID
func WithCategoryID(categoryID string) ProcessingOption {
	return func(c *processingConfig) {
		c.categoryID = categoryID
	}
}

// applyProcessingOptions applies all options to a config
func applyProcessingOptions(opts []ProcessingOption) *processingConfig {
	config := &processingConfig{
		format: FormatStructured,        // default format
		model:  string(ModelStandardV1), // default model
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}
