package ocr

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAPIClient(t *testing.T) {
	cfg := NewConfiguration()
	client := NewAPIClient(cfg)

	assert.NotNil(t, client)
	assert.NotNil(t, client.SDKAPI)
	assert.Equal(t, cfg, client.cfg)
}

func TestGetJobStatus(t *testing.T) {
	// Mock server that returns a status response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, "/ocr/status/test-job-id"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
			"id": "test-job-id",
			"status": "completed",
			"created_at": "2023-01-01T00:00:00Z"
		}`))
	}))
	defer server.Close()

	// Create client with test server
	cfg := NewConfiguration()
	cfg.Servers = []ServerConfiguration{{URL: server.URL}}
	client := NewAPIClient(cfg)

	ctx := context.Background()
	jobID := "test-job-id"

	// Test GetJobStatus
	req := client.SDKAPI.GetJobStatus(ctx, jobID)
	result, resp, err := req.Execute()

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	assert.NotNil(t, result)
	// Note: The actual field name may vary based on the OpenAPI spec
}

func TestGetJobResult(t *testing.T) {
	// Mock server that returns a result response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, "/ocr/result/test-job-id"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
			"job_id": "test-job-id",
			"pages": []
		}`))
	}))
	defer server.Close()

	// Create client with test server
	cfg := NewConfiguration()
	cfg.Servers = []ServerConfiguration{{URL: server.URL}}
	client := NewAPIClient(cfg)

	ctx := context.Background()
	jobID := "test-job-id"

	// Test GetJobResult
	req := client.SDKAPI.GetJobResult(ctx, jobID)
	result, resp, err := req.Execute()

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	assert.NotNil(t, result)
	// Note: The actual field name may vary based on the OpenAPI spec
}

func TestPresignedUpload(t *testing.T) {
	// Mock server that returns a presigned upload response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, "/ocr/uploads/presigned"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
			"job_id": "test-job-id",
			"upload_url": "https://example.com/upload",
			"expires_at": "2023-01-01T01:00:00Z"
		}`))
	}))
	defer server.Close()

	// Create client with test server
	cfg := NewConfiguration()
	cfg.Servers = []ServerConfiguration{{URL: server.URL}}
	client := NewAPIClient(cfg)

	ctx := context.Background()

	// Create upload request
	uploadReq := UploadInitiateUploadRequest{
		FileName:    "test.pdf",
		ContentType: "application/pdf",
		Format:      stringPtr("structured"),
	}

	// Test PresignedUpload
	req := client.SDKAPI.PresignedUpload(ctx).UploadInitiateUploadRequest(uploadReq)
	result, resp, err := req.Execute()

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	assert.NotNil(t, result)
	assert.NotNil(t, result.JobId)
	assert.NotNil(t, result.UploadUrl)
}

func TestUploadFromURL(t *testing.T) {
	// Mock server that returns a URL upload response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, "/ocr/uploads/url"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
			"job_id": "test-job-id",
			"status": "processing"
		}`))
	}))
	defer server.Close()

	// Create client with test server
	cfg := NewConfiguration()
	cfg.Servers = []ServerConfiguration{{URL: server.URL}}
	client := NewAPIClient(cfg)

	ctx := context.Background()

	// Create URL upload request
	urlReq := UploadURLUploadRequest{
		Url:    "https://example.com/document.pdf",
		Format: stringPtr("markdown"),
	}

	// Test UploadFromURL
	req := client.SDKAPI.UploadFromURL(ctx).UploadURLUploadRequest(urlReq)
	result, resp, err := req.Execute()

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	assert.NotNil(t, result)
	assert.NotNil(t, result.JobId)
}

func TestConfiguration(t *testing.T) {
	cfg := NewConfiguration()

	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.UserAgent)
}

func TestServerURLWithContext(t *testing.T) {
	cfg := NewConfiguration()
	cfg.Servers = []ServerConfiguration{
		{URL: "https://api.example.com"},
	}

	ctx := context.Background()
	url, err := cfg.ServerURLWithContext(ctx, "TestOperation")

	assert.NoError(t, err)
	assert.Equal(t, "https://api.example.com", url)
}

func TestErrorHandling(t *testing.T) {
	// Mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		w.Write([]byte(`{
			"error": "job not found",
			"message": "The requested job does not exist"
		}`))
	}))
	defer server.Close()

	// Create client with test server
	cfg := NewConfiguration()
	cfg.Servers = []ServerConfiguration{{URL: server.URL}}
	client := NewAPIClient(cfg)

	ctx := context.Background()
	jobID := "nonexistent-job"

	// Test error handling
	req := client.SDKAPI.GetJobStatus(ctx, jobID)
	result, resp, err := req.Execute()

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 404, resp.StatusCode)
	assert.Nil(t, result)

	// Check if it's a GenericOpenAPIError
	if apiErr, ok := err.(*GenericOpenAPIError); ok {
		assert.Equal(t, "404 Not Found", apiErr.Error())
		assert.NotEmpty(t, apiErr.Body())
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
