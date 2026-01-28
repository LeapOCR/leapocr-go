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
	Text     string         `json:"text"`
	Data     map[string]any `json:"data"`
	Pages    []PageResult   `json:"pages"`
	Credits  int            `json:"credits"`
	Duration time.Duration  `json:"duration"`
	JobID    string         `json:"job_id"`
	Status   string         `json:"status"`
}

// PageResult represents a single page result
type PageResult struct {
	PageNumber int            `json:"page_number"`
	Text       string         `json:"text"`
	Data       map[string]any `json:"data"`
	Confidence *float64       `json:"confidence,omitempty"`
}

// ProcessingOption configures OCR processing
type ProcessingOption func(*processingConfig)

// processingConfig holds all processing configuration
type processingConfig struct {
	format          Format
	model           string // Can be a Model constant or any custom model string
	schema          map[string]any
	instructions    string
	templateSlug    string
	formatSet       bool
	modelSet        bool
	schemaSet       bool
	instructionsSet bool
	templateSlugSet bool
}

// WithFormat sets the output format
func WithFormat(format Format) ProcessingOption {
	return func(c *processingConfig) {
		c.format = format
		c.formatSet = true
	}
}

// WithModel sets the OCR model to use for processing
// You can use one of the predefined constants (ModelStandardV1, ModelEnglishProV1, ModelProV1)
// or pass any custom model name as a string
func WithModel(model Model) ProcessingOption {
	return func(c *processingConfig) {
		c.model = string(model)
		c.modelSet = true
	}
}

// WithModelString sets a custom model name as a string
// This allows using any model name not defined as a constant
func WithModelString(model string) ProcessingOption {
	return func(c *processingConfig) {
		c.model = model
		c.modelSet = true
	}
}

// WithSchema sets a custom extraction schema
func WithSchema(schema map[string]any) ProcessingOption {
	return func(c *processingConfig) {
		c.schema = schema
		c.schemaSet = true
	}
}

// WithInstructions sets custom processing instructions
func WithInstructions(instructions string) ProcessingOption {
	return func(c *processingConfig) {
		c.instructions = instructions
		c.instructionsSet = true
	}
}

// WithTemplateSlug sets the template slug for structured extraction
func WithTemplateSlug(templateSlug string) ProcessingOption {
	return func(c *processingConfig) {
		c.templateSlug = templateSlug
		c.templateSlugSet = true
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
