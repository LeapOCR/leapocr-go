package client

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	client := New("test-api-key")
	assert.NotNil(t, client)
	assert.Equal(t, "test-api-key", client.config.APIKey)
	assert.NotNil(t, client.OCR)
}

func TestNewWithConfig(t *testing.T) {
	config := NewConfig("test-api-key")
	config.WithTimeout(60 * time.Second)

	client := NewWithConfig(config)
	assert.NotNil(t, client)
	assert.Equal(t, "test-api-key", client.config.APIKey)
	assert.Equal(t, 60*time.Second, client.config.Timeout)
}

func TestConfigValidation(t *testing.T) {
	// Test valid config
	config := NewConfig("test-api-key")
	err := config.Validate()
	assert.NoError(t, err)

	// Test invalid config (empty API key)
	config2 := NewConfig("")
	err2 := config2.Validate()
	assert.Error(t, err2)
}

func TestConfigWithMethods(t *testing.T) {
	config := NewConfig("test-api-key")

	// Test WithTimeout
	config.WithTimeout(30 * time.Second)
	assert.Equal(t, 30*time.Second, config.Timeout)

	// Test WithRetries
	config.WithRetries(5, 1*time.Second, 30*time.Second)
	assert.Equal(t, 5, config.RetryConfig.MaxRetries)
	assert.Equal(t, 1*time.Second, config.RetryConfig.InitialDelay)
}

func TestSetBaseURL(t *testing.T) {
	config := NewConfig("test-api-key")

	// Test valid URL
	err := config.SetBaseURL("https://api.example.com")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.example.com", config.BaseURL.String())

	// Test invalid URL (Go's url.Parse is quite permissive, so this may not error)
	err2 := config.SetBaseURL("://invalid-url")
	assert.Error(t, err2)
}

func TestClientCreation(t *testing.T) {
	// Test with default config
	client := New("test-api-key")
	assert.NotNil(t, client)
	assert.NotNil(t, client.ocrClient)
	assert.NotNil(t, client.config)
	assert.NotNil(t, client.OCR)

	// Verify the underlying OCR client is configured
	assert.NotNil(t, client.ocrClient.SDKAPI)
}

func TestOCRServiceMethods(t *testing.T) {
	client := New("test-api-key")
	ctx := context.Background()

	// Test that the methods exist and return "not implemented" errors
	// This validates the interface without requiring full implementation

	t.Run("ProcessFileFromPath", func(t *testing.T) {
		result, err := client.OCR.ProcessFileFromPath(ctx, "/path/to/file.pdf")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")
		assert.Nil(t, result)
	})

	t.Run("ProcessFileFromURL", func(t *testing.T) {
		result, err := client.OCR.ProcessFileFromURL(ctx, "https://example.com/file.pdf")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")
		assert.Nil(t, result)
	})

	t.Run("GetJobStatus", func(t *testing.T) {
		result, err := client.OCR.GetJobStatus(ctx, "test-job-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")
		assert.Nil(t, result)
	})

	t.Run("GetJobResult", func(t *testing.T) {
		result, err := client.OCR.GetJobResult(ctx, "test-job-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")
		assert.Nil(t, result)
	})

	t.Run("WaitForCompletion", func(t *testing.T) {
		result, err := client.OCR.WaitForCompletion(ctx, "test-job-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")
		assert.Nil(t, result)
	})
}

func TestProcessOptions(t *testing.T) {
	// Test process option functions
	config := &ProcessConfig{}

	WithFormat("structured")(config)
	assert.Equal(t, "structured", config.Format)

	WithTemplateID("template-123")(config)
	assert.Equal(t, "template-123", config.TemplateID)

	WithInstructions("Extract invoice data")(config)
	assert.Equal(t, "Extract invoice data", config.Instructions)

	WithTier("swift")(config)
	assert.Equal(t, "swift", config.Tier)

	schema := map[string]interface{}{"field": "value"}
	WithSchema(schema)(config)
	assert.Equal(t, schema, config.Schema)
}

func TestConstants(t *testing.T) {
	// Test format constants
	assert.Equal(t, "markdown", FormatMarkdown)
	assert.Equal(t, "structured", FormatStructured)
	assert.Equal(t, "per_page_structured", FormatPerPageStructured)

	// Test tier constants
	assert.Equal(t, "swift", TierSwift)
	assert.Equal(t, "core", TierCore)
	assert.Equal(t, "intelli", TierIntelli)
}

func TestConfigWithHTTPClient(t *testing.T) {
	// Create custom HTTP client
	customClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	config := NewConfig("test-api-key")
	config.HTTPClient = customClient

	client := NewWithConfig(config)
	assert.NotNil(t, client)

	// The custom HTTP client should be used in the underlying OCR client
	assert.Equal(t, customClient, client.ocrClient.GetConfig().HTTPClient)
}

func TestUserAgent(t *testing.T) {
	config := NewConfig("test-api-key")
	config.UserAgent = "MyApp/1.0"

	client := NewWithConfig(config)
	assert.Equal(t, "MyApp/1.0", client.ocrClient.GetConfig().UserAgent)
}

func TestAuthenticationSetup(t *testing.T) {
	client := New("test-api-key")

	// Check that the Authorization header is set correctly
	authHeader := client.ocrClient.GetConfig().DefaultHeader["Authorization"]
	assert.Equal(t, "Bearer test-api-key", authHeader)
}

func TestServerConfiguration(t *testing.T) {
	config := NewConfig("test-api-key")
	err := config.SetBaseURL("https://custom-api.example.com")
	require.NoError(t, err)

	client := NewWithConfig(config)

	// Verify the server configuration is set correctly
	servers := client.ocrClient.GetConfig().Servers
	assert.Len(t, servers, 1)
	assert.Equal(t, "https://custom-api.example.com", servers[0].URL)
}

func TestRetryConfig(t *testing.T) {
	config := NewConfig("test-api-key")

	// Test default retry config
	assert.NotNil(t, config.RetryConfig)
	assert.Equal(t, 3, config.RetryConfig.MaxRetries)

	// Test custom retry config
	config.WithRetries(5, 2*time.Second, 30*time.Second)

	assert.Equal(t, 5, config.RetryConfig.MaxRetries)
	assert.Equal(t, 2*time.Second, config.RetryConfig.InitialDelay)
	assert.Equal(t, 30*time.Second, config.RetryConfig.MaxDelay)
}

func TestPanicOnInvalidConfig(t *testing.T) {
	// Test that invalid config causes panic
	invalidConfig := &Config{
		APIKey: "", // Invalid: empty API key
	}

	assert.Panics(t, func() {
		NewWithConfig(invalidConfig)
	})
}
