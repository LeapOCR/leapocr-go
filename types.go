package ocr

import (
	"time"
)

// Format represents the OCR output format
type Format string

const (
	// FormatMarkdown outputs text as markdown
	FormatMarkdown          Format = "markdown"
	// FormatStructured outputs structured data as JSON
	FormatStructured        Format = "structured"
	// FormatPerPageStructured outputs structured data per page
	FormatPerPageStructured Format = "per_page_structured"
)

// Tier represents the processing tier
type Tier string

const (
	// TierSwift provides fast processing with basic accuracy
	TierSwift   Tier = "swift"
	// TierCore provides balanced speed and accuracy
	TierCore    Tier = "core"
	// TierIntelli provides highest accuracy with advanced processing
	TierIntelli Tier = "intelli"
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
	Confidence float64                `json:"confidence"`
}

// ProcessingOption configures OCR processing
type ProcessingOption func(*processingConfig)

// processingConfig holds all processing configuration
type processingConfig struct {
	format       Format
	tier         Tier
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

// WithTier sets the processing tier
func WithTier(tier Tier) ProcessingOption {
	return func(c *processingConfig) {
		c.tier = tier
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
		format: FormatStructured, // default format
		tier:   TierCore,         // default tier
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}
