package ocr

import (
	"net/http"
	"time"

	"github.com/leapocr/leapocr-go/internal/generated"
)

// SDK is the main OCR API client that provides a clean, Go-native interface
type SDK struct {
	client *generated.APIClient
	config *Config
}

// Config holds the SDK configuration
type Config struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	UserAgent  string
	Timeout    time.Duration
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig(apiKey string) *Config {
	return &Config{
		APIKey:     apiKey,
		BaseURL:    "https://api.leapocr.com",
		HTTPClient: &http.Client{},
		UserAgent:  "leapocr-go/" + Version,
		Timeout:    30 * time.Second,
	}
}

// NewSDK creates a new SDK instance with the given configuration
func NewSDK(config *Config) (*SDK, error) {
	if config.APIKey == "" {
		return nil, &SDKError{
			Type:    ErrorTypeInvalidConfig,
			Message: "API key is required",
		}
	}

	// Create the generated client configuration
	genConfig := generated.NewConfiguration()
	genConfig.Servers = []generated.ServerConfiguration{
		{
			URL: config.BaseURL,
		},
	}

	// Set up authentication
	genConfig.DefaultHeader["X-API-KEY"] = config.APIKey

	// Configure HTTP client
	if config.HTTPClient != nil {
		genConfig.HTTPClient = config.HTTPClient
	}
	if config.UserAgent != "" {
		genConfig.UserAgent = config.UserAgent
	}

	// Create the generated client
	client := generated.NewAPIClient(genConfig)

	return &SDK{
		client: client,
		config: config,
	}, nil
}

// New creates a new SDK instance with default configuration
func New(apiKey string) (*SDK, error) {
	return NewSDK(DefaultConfig(apiKey))
}

// Job represents a processing job
type Job struct {
	ID     string
	Status string
}
