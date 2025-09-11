//go:build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/your-org/ocr-go-sdk"
	"github.com/your-org/ocr-go-sdk/ocr"
)

// Integration tests require:
// 1. OCR_API_KEY environment variable
// 2. OCR API server running (default: http://localhost:8080)
// 3. Sample test files in test/fixtures/

func TestIntegration_HealthCheck(t *testing.T) {
	client := createTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	health, err := client.Health.Check(ctx)
	require.NoError(t, err)
	assert.NotNil(t, health)

	t.Logf("Health check response: %+v", health)
}

func TestIntegration_ProcessFile(t *testing.T) {
	client := createTestClient(t)

	// Look for test files
	testFiles := []string{
		"../fixtures/sample-invoice.pdf",
		"../fixtures/sample-document.pdf",
	}

	var testFile string
	for _, file := range testFiles {
		if _, err := os.Stat(file); err == nil {
			testFile = file
			break
		}
	}

	if testFile == "" {
		t.Skip("No test files found in test/fixtures/. Add sample PDF files to run this test.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	req := &ocr.ProcessFileRequest{
		FilePath:           testFile,
		Mode:               ocr.ModeTextAndImage,
		CustomInstructions: "Extract all text and identify key information",
		Priority:           ocr.PriorityNormal,
	}

	t.Logf("Processing file: %s", testFile)
	job, err := client.OCR.ProcessFile(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, job)

	t.Logf("Job created with ID: %s", job.ID)

	// Wait for completion
	result, err := client.OCR.WaitForCompletion(ctx, job.ID)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify results
	assert.Equal(t, ocr.StatusCompleted, result.Status)
	assert.Greater(t, result.CreditsCost, 0)
	assert.Greater(t, len(result.Pages), 0)

	t.Logf("Processing completed successfully!")
	t.Logf("Credits used: %d", result.CreditsCost)
	t.Logf("Processing time: %v", result.ProcessingTime)
	t.Logf("Pages processed: %d", len(result.Pages))

	if len(result.Pages) > 0 {
		t.Logf("First page text length: %d characters", len(result.Pages[0].Text))
		t.Logf("First page confidence: %.2f", result.Pages[0].Confidence)
	}
}

func TestIntegration_ProcessURL(t *testing.T) {
	client := createTestClient(t)

	// This test would need a publicly accessible test document URL
	testURL := os.Getenv("TEST_DOCUMENT_URL")
	if testURL == "" {
		t.Skip("TEST_DOCUMENT_URL environment variable not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	opts := &ocr.ProcessFileRequest{
		Mode:     ocr.ModeTextOnly,
		Priority: ocr.PriorityHigh,
	}

	t.Logf("Processing URL: %s", testURL)
	job, err := client.OCR.ProcessURL(ctx, testURL, opts)
	require.NoError(t, err)
	require.NotNil(t, job)

	t.Logf("Job created with ID: %s", job.ID)

	// Poll for completion
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatal("Timeout waiting for job completion")
		case <-ticker.C:
			status, err := client.OCR.GetStatus(ctx, job.ID)
			require.NoError(t, err)

			t.Logf("Job status: %s", status.Status)
			if status.Progress != nil {
				t.Logf("Progress: %.1f%%", status.Progress.Percentage)
			}

			if status.Status == ocr.StatusCompleted {
				result, err := client.OCR.GetResult(ctx, job.ID)
				require.NoError(t, err)

				assert.Equal(t, ocr.StatusCompleted, result.Status)
				assert.Greater(t, result.CreditsCost, 0)
				assert.Greater(t, len(result.Pages), 0)

				t.Logf("URL processing completed successfully!")
				t.Logf("Credits used: %d", result.CreditsCost)
				t.Logf("Text extracted: %d characters", len(result.Pages[0].Text))
				return
			} else if status.Status == ocr.StatusFailed {
				if status.Error != nil {
					t.Fatalf("Job failed: %s", *status.Error)
				}
				t.Fatal("Job failed with unknown error")
			}
		}
	}
}

func TestIntegration_Analytics(t *testing.T) {
	client := createTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test credits analytics
	t.Run("Credits Analytics", func(t *testing.T) {
		analytics, err := client.Analytics.GetCreditsAnalytics(ctx)
		require.NoError(t, err)
		assert.NotNil(t, analytics)
		t.Logf("Credits analytics: %+v", analytics)
	})

	// Test jobs analytics
	t.Run("Jobs Analytics", func(t *testing.T) {
		analytics, err := client.Analytics.GetJobsAnalytics(ctx)
		require.NoError(t, err)
		assert.NotNil(t, analytics)
		t.Logf("Jobs analytics: %+v", analytics)
	})
}

func TestIntegration_JobManagement(t *testing.T) {
	client := createTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test job listing
	jobs, err := client.Jobs.List(ctx)
	require.NoError(t, err)
	assert.NotNil(t, jobs)
	t.Logf("Job list: %+v", jobs)
}

func createTestClient(t *testing.T) *ocr.Client {
	apiKey := os.Getenv("OCR_API_KEY")
	if apiKey == "" {
		t.Fatal("OCR_API_KEY environment variable is required for integration tests")
	}

	baseURL := os.Getenv("OCR_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	config := ocr.NewConfig(apiKey)
	if err := config.SetBaseURL(baseURL); err != nil {
		t.Fatalf("Invalid base URL: %v", err)
	}

	// Configure for testing
	config.WithTimeout(2 * time.Minute)
	config.WithUserAgent("ocr-go-sdk-test/1.0.0")

	client := ocr.NewWithConfig(config)
	return client
}
