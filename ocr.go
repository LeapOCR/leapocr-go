// Package ocr provides the official Go SDK for the OCR API
package ocr

import (
	"github.com/your-org/ocr-go-sdk/client"
)

// Client is an alias for the main client
type Client = client.Client

// Config is an alias for the client configuration
type Config = client.Config

// New creates a new OCR API client with the given API key
func New(apiKey string) *Client {
	return client.New(apiKey)
}

// NewWithConfig creates a new OCR API client with the given configuration
func NewWithConfig(config *Config) *Client {
	return client.NewWithConfig(config)
}

// NewConfig creates a new configuration with default values
func NewConfig(apiKey string) *Config {
	return client.NewConfig(apiKey)
}

// Re-export common types and constants
type (
	ProcessOption = client.ProcessOption
	ProcessResult = client.ProcessResult
	JobStatus     = client.JobStatus
	JobResult     = client.JobResult
	PageResult    = client.PageResult
)

// Re-export processing options
var (
	WithFormat       = client.WithFormat
	WithTemplateID   = client.WithTemplateID
	WithSchema       = client.WithSchema
	WithInstructions = client.WithInstructions
	WithTier         = client.WithTier
)

// Re-export format constants
const (
	FormatMarkdown          = client.FormatMarkdown
	FormatStructured        = client.FormatStructured
	FormatPerPageStructured = client.FormatPerPageStructured
)

// Re-export tier constants
const (
	TierSwift   = client.TierSwift
	TierCore    = client.TierCore
	TierIntelli = client.TierIntelli
)
