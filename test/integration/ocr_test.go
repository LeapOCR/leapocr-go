//go:build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ocr "github.com/leapocr/leapocr-go"
)

// Integration tests require:
// 1. LEAPOCR_API_KEY environment variable
// 2. OCR API server running (default: http://localhost:8080)
// 3. Sample test files in test/fixtures/

func TestIntegration_ProcessFile(t *testing.T) {
	sdk := createTestSDK(t)

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

	// Open test file
	file, err := os.Open(testFile)
	require.NoError(t, err)
	defer file.Close()

	t.Logf("Processing file: %s", testFile)
	job, err := sdk.ProcessFile(ctx, file, "test-document.pdf",
		ocr.WithFormat(ocr.FormatStructured),
		ocr.WithTier(ocr.TierCore),
		ocr.WithInstructions("Extract all text and identify key information"),
	)
	require.NoError(t, err)
	require.NotNil(t, job)

	t.Logf("Job created with ID: %s", job.ID)

	// Wait for completion
	result, err := sdk.WaitUntilDone(ctx, job.ID)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify results
	assert.Equal(t, "completed", result.Status)
	assert.Greater(t, result.Credits, 0)
	assert.Greater(t, len(result.Pages), 0)

	t.Logf("Processing completed successfully!")
	t.Logf("Credits used: %d", result.Credits)
	t.Logf("Processing time: %v", result.Duration)
	t.Logf("Pages processed: %d", len(result.Pages))

	if len(result.Pages) > 0 {
		t.Logf("First page text length: %d characters", len(result.Pages[0].Text))
		t.Logf("First page confidence: %.2f", result.Pages[0].Confidence)
	}
}

func TestIntegration_ProcessURL(t *testing.T) {
	sdk := createTestSDK(t)

	// This test would need a publicly accessible test document URL
	testURL := os.Getenv("TEST_DOCUMENT_URL")
	if testURL == "" {
		t.Skip("TEST_DOCUMENT_URL environment variable not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	t.Logf("Processing URL: %s", testURL)
	job, err := sdk.ProcessURL(ctx, testURL,
		ocr.WithFormat(ocr.FormatMarkdown),
		ocr.WithTier(ocr.TierSwift),
	)
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
			status, err := sdk.GetJobStatus(ctx, job.ID)
			require.NoError(t, err)

			t.Logf("Job status: %s", status.Status)
			if status.Progress != 0 {
				t.Logf("Progress: %.1f%%", status.Progress)
			}

			switch status.Status {
			case "completed":
				result, err := sdk.GetJobResult(ctx, job.ID)
				require.NoError(t, err)

				assert.Equal(t, "completed", result.Status)
				assert.Greater(t, result.Credits, 0)
				assert.Greater(t, len(result.Text), 0)

				t.Logf("URL processing completed successfully!")
				t.Logf("Credits used: %d", result.Credits)
				t.Logf("Text extracted: %d characters", len(result.Text))
				return

			case "failed", "error":
				if status.Error != "" {
					t.Fatalf("Job failed: %s", status.Error)
				}
				t.Fatal("Job failed with unknown error")
			}
		}
	}
}

func TestIntegration_ErrorHandling(t *testing.T) {
	sdk := createTestSDK(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test invalid URL
	t.Run("Invalid URL", func(t *testing.T) {
		_, err := sdk.ProcessURL(ctx, "not-a-valid-url",
			ocr.WithFormat(ocr.FormatStructured))
		require.Error(t, err)

		// Check if it's an SDK error
		if sdkErr, ok := err.(*ocr.SDKError); ok {
			t.Logf("Received SDK error: %s - %s", sdkErr.Type, sdkErr.Message)
		}
	})

	// Test non-existent job status
	t.Run("Non-existent Job Status", func(t *testing.T) {
		_, err := sdk.GetJobStatus(ctx, "non-existent-job-id-12345")
		require.Error(t, err)
		t.Logf("Correctly handled non-existent job: %v", err)
	})
}

func TestIntegration_CustomWaitOptions(t *testing.T) {
	sdk := createTestSDK(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test custom wait options with non-existent job (should fail quickly)
	waitOpts := ocr.WaitOptions{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		MaxJitter:    200 * time.Millisecond,
		MaxAttempts:  3, // Very few attempts
	}

	start := time.Now()
	_, err := sdk.WaitUntilDoneWithOptions(ctx, "fake-job-id", waitOpts)
	duration := time.Since(start)

	require.Error(t, err)
	t.Logf("Wait with custom options failed after %v: %v", duration, err)
}

func createTestSDK(t *testing.T) *ocr.SDK {
	apiKey := os.Getenv("LEAPOCR_API_KEY")
	if apiKey == "" {
		t.Fatal("LEAPOCR_API_KEY environment variable is required for integration tests")
	}

	baseURL := os.Getenv("OCR_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Create custom config for testing
	config := ocr.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	config.Timeout = 2 * time.Minute
	config.UserAgent = "leapocr-go-sdk-test/1.0.0"

	sdk, err := ocr.NewSDK(config)
	if err != nil {
		t.Fatalf("Failed to create test SDK: %v", err)
	}

	return sdk
}
