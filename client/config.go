package client

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Config represents the configuration for the OCR API client
type Config struct {
	// APIKey is the API key for authentication
	APIKey string

	// BaseURL is the base URL for the OCR API (default: http://localhost:8080)
	BaseURL *url.URL

	// HTTPClient is the HTTP client to use for requests
	HTTPClient *http.Client

	// UserAgent is the user agent string to send with requests
	UserAgent string

	// Timeout is the request timeout (default: 30s)
	Timeout time.Duration

	// RetryConfig defines retry behavior
	RetryConfig *RetryConfig
}

// RetryConfig defines retry behavior for failed requests
type RetryConfig struct {
	// MaxRetries is the maximum number of retries (default: 3)
	MaxRetries int

	// InitialDelay is the initial delay between retries (default: 1s)
	InitialDelay time.Duration

	// MaxDelay is the maximum delay between retries (default: 30s)
	MaxDelay time.Duration

	// BackoffMultiplier is the multiplier for exponential backoff (default: 2.0)
	BackoffMultiplier float64
}

// NewConfig creates a new configuration with default values
func NewConfig(apiKey string) *Config {
	baseURL, _ := url.Parse("http://localhost:8080")

	return &Config{
		APIKey:     apiKey,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
		UserAgent:  "ocr-go-sdk/1.0.0",
		Timeout:    30 * time.Second,
		RetryConfig: &RetryConfig{
			MaxRetries:        3,
			InitialDelay:      time.Second,
			MaxDelay:          30 * time.Second,
			BackoffMultiplier: 2.0,
		},
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	if c.BaseURL == nil {
		return fmt.Errorf("base URL is required")
	}

	if c.HTTPClient == nil {
		return fmt.Errorf("HTTP client is required")
	}

	return nil
}

// SetBaseURL sets the base URL from a string
func (c *Config) SetBaseURL(baseURL string) error {
	u, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}
	c.BaseURL = u
	return nil
}

// WithTimeout sets the request timeout
func (c *Config) WithTimeout(timeout time.Duration) *Config {
	c.Timeout = timeout
	return c
}

// WithUserAgent sets the user agent
func (c *Config) WithUserAgent(userAgent string) *Config {
	c.UserAgent = userAgent
	return c
}

// WithRetries configures retry behavior
func (c *Config) WithRetries(maxRetries int, initialDelay, maxDelay time.Duration) *Config {
	c.RetryConfig = &RetryConfig{
		MaxRetries:        maxRetries,
		InitialDelay:      initialDelay,
		MaxDelay:          maxDelay,
		BackoffMultiplier: 2.0,
	}
	return c
}
