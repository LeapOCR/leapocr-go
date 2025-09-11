package client

import (
	"context"
	"fmt"

	"github.com/your-org/ocr-go-sdk/ocr"
)

// Client is the main OCR API client that wraps the generated client
type Client struct {
	ocrClient *ocr.APIClient
	config    *Config
	OCR       *OCRService
}

// OCRService wraps OCR operations with a simplified interface
type OCRService struct {
	client *ocr.APIClient
}

// New creates a new OCR API client with the given API key
func New(apiKey string) *Client {
	return NewWithConfig(NewConfig(apiKey))
}

// NewWithConfig creates a new OCR API client with the given configuration
func NewWithConfig(config *Config) *Client {
	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("invalid configuration: %v", err))
	}

	// Create the generated client configuration
	ocrConfig := ocr.NewConfiguration()
	ocrConfig.Servers = []ocr.ServerConfiguration{
		{
			URL: config.BaseURL.String(),
		},
	}

	// Set up authentication
	if config.APIKey != "" {
		ocrConfig.DefaultHeader["Authorization"] = "Bearer " + config.APIKey
	}

	// Configure HTTP client
	ocrConfig.HTTPClient = config.HTTPClient
	ocrConfig.UserAgent = config.UserAgent

	// Create the generated client
	ocrClient := ocr.NewAPIClient(ocrConfig)

	client := &Client{
		ocrClient: ocrClient,
		config:    config,
	}

	// Initialize service wrappers
	client.OCR = &OCRService{client: ocrClient}

	return client
}

// ProcessFileFromPath processes a file by path using the generated client
func (s *OCRService) ProcessFileFromPath(ctx context.Context, filePath string, options ...ProcessOption) (*ProcessResult, error) {
	// This would use the generated client to process files
	// Implementation depends on the exact generated API structure
	return nil, fmt.Errorf("not implemented - requires generated client structure")
}

// ProcessFileFromURL processes a file from URL using the generated client
func (s *OCRService) ProcessFileFromURL(ctx context.Context, url string, options ...ProcessOption) (*ProcessResult, error) {
	// This would use the generated client to process URLs
	// Implementation depends on the exact generated API structure
	return nil, fmt.Errorf("not implemented - requires generated client structure")
}

// GetJobStatus gets job status using the generated client
func (s *OCRService) GetJobStatus(ctx context.Context, jobID string) (*JobStatus, error) {
	// This would use the generated client to get job status
	// Implementation depends on the exact generated API structure
	return nil, fmt.Errorf("not implemented - requires generated client structure")
}

// GetJobResult gets job result using the generated client
func (s *OCRService) GetJobResult(ctx context.Context, jobID string) (*JobResult, error) {
	// This would use the generated client to get job results
	// Implementation depends on the exact generated API structure
	return nil, fmt.Errorf("not implemented - requires generated client structure")
}

// WaitForCompletion waits for job completion using the generated client
func (s *OCRService) WaitForCompletion(ctx context.Context, jobID string) (*JobResult, error) {
	// This would implement polling using the generated client
	// Implementation depends on the exact generated API structure
	return nil, fmt.Errorf("not implemented - requires generated client structure")
}

// ProcessOption represents processing options
type ProcessOption func(*ProcessConfig)

// ProcessConfig represents processing configuration
type ProcessConfig struct {
	Format       string
	TemplateID   string
	Schema       map[string]interface{}
	Instructions string
	Tier         string
}

// ProcessResult represents the result of starting a processing job
type ProcessResult struct {
	JobID     string
	UploadURL string
	Status    string
}

// JobStatus represents job status information
type JobStatus struct {
	JobID         string
	Status        string
	Progress      float64
	EstimatedTime int
	Error         string
}

// JobResult represents the final job result
type JobResult struct {
	JobID          string
	Status         string
	Data           map[string]interface{}
	Pages          []PageResult
	ProcessingTime int
	CreditsUsed    int
	Error          string
}

// PageResult represents a single page result
type PageResult struct {
	PageNumber int                    `json:"page_number"`
	Text       string                 `json:"text"`
	Data       map[string]interface{} `json:"data"`
	Confidence float64                `json:"confidence"`
}

// Processing format options
func WithFormat(format string) ProcessOption {
	return func(c *ProcessConfig) {
		c.Format = format
	}
}

// WithTemplateID sets the template ID
func WithTemplateID(templateID string) ProcessOption {
	return func(c *ProcessConfig) {
		c.TemplateID = templateID
	}
}

// WithSchema sets the processing schema
func WithSchema(schema map[string]interface{}) ProcessOption {
	return func(c *ProcessConfig) {
		c.Schema = schema
	}
}

// WithInstructions sets processing instructions
func WithInstructions(instructions string) ProcessOption {
	return func(c *ProcessConfig) {
		c.Instructions = instructions
	}
}

// WithTier sets the processing tier
func WithTier(tier string) ProcessOption {
	return func(c *ProcessConfig) {
		c.Tier = tier
	}
}

// Format constants
const (
	FormatMarkdown          = "markdown"
	FormatStructured        = "structured"
	FormatPerPageStructured = "per_page_structured"
)

// Tier constants
const (
	TierSwift   = "swift"
	TierCore    = "core"
	TierIntelli = "intelli"
)
