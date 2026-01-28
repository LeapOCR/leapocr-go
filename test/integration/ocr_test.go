//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ocr "github.com/leapocr/leapocr-go"
)

// Integration tests require:
// 1. LEAPOCR_API_KEY environment variable
// 2. OCR API server running (default: http://localhost:8443/api/v1)
// 3. Sample test files in sample/ folder (e.g., test.pdf, A129of19_14.01.22.pdf)

func TestIntegration_ProcessFile(t *testing.T) {
	sdk := createTestSDK(t)
	templateSlug := os.Getenv("LEAPOCR_TEMPLATE_SLUG")

	// Find repo root by looking for go.mod file
	repoRoot := findRepoRoot(t)

	// Look for test files in sample/ folder
	sampleDir := filepath.Join(repoRoot, "sample")
	testFiles := []string{
		filepath.Join(sampleDir, "test.pdf"),
		filepath.Join(sampleDir, "A129of19_14.01.22.pdf"),
		filepath.Join(sampleDir, "A141of21_10.02.22.pdf"),
		filepath.Join(sampleDir, "A29of21&B_31.03.22.pdf"),
		filepath.Join(sampleDir, "A66of20_oral_07.01.22.pdf"),
	}

	var testFile string
	for _, file := range testFiles {
		if _, err := os.Stat(file); err == nil {
			testFile = file
			break
		}
	}

	if testFile == "" {
		t.Skip("No test files found in sample/ folder. Add sample PDF files to run this test.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Open test file
	file, err := os.Open(testFile)
	require.NoError(t, err)
	defer file.Close()

	filename := filepath.Base(testFile)
	t.Logf("Processing PDF file: %s (full path: %s)", filename, testFile)

	// Step 1: ProcessFile handles the full direct upload flow:
	// - Initiates direct upload (gets presigned URLs for chunks)
	// - Uploads chunks to presigned URLs
	// - Completes the upload
	// - Returns job ID
	t.Logf("Step 1: Initiating direct upload (will get presigned URLs for chunks)...")
	structuredOptions := []ocr.ProcessingOption{}
	if templateSlug != "" {
		structuredOptions = append(structuredOptions, ocr.WithTemplateSlug(templateSlug))
	} else {
		structuredOptions = append(
			structuredOptions,
			ocr.WithFormat(ocr.FormatStructured),
			ocr.WithModel(ocr.ModelStandardV1),
			ocr.WithInstructions("Extract all text and identify key information"),
			ocr.WithSchema(map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{"type": "string"},
				},
				"required": []interface{}{"text"},
			}),
		)
	}

	job, err := sdk.ProcessFile(ctx, file, filepath.Base(testFile), structuredOptions...)
	require.NoError(t, err)
	require.NotNil(t, job)

	t.Logf("Step 2: Direct upload completed. Job created with ID: %s", job.ID)
	t.Logf("Step 3: Waiting for OCR processing to complete...")

	// Step 4: Wait for completion after the upload is complete
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
	}

	// Test deletion
	t.Logf("Step 4: Deleting job to test cleanup...")
	err = sdk.DeleteJob(ctx, job.ID)
	require.NoError(t, err)
	t.Logf("Job deleted successfully!")
}

// findRepoRoot finds the repository root by looking for go.mod file
func findRepoRoot(t *testing.T) string {
	// Start from the current working directory
	wd, err := os.Getwd()
	require.NoError(t, err)

	dir := wd
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			t.Fatalf("Could not find repository root (go.mod not found)")
		}
		dir = parent
	}
}

func TestIntegration_ProcessURL(t *testing.T) {
	sdk := createTestSDK(t)

	enabled := strings.ToLower(os.Getenv("LEAPOCR_URL_UPLOAD_ENABLED"))
	if enabled != "1" && enabled != "true" && enabled != "yes" {
		t.Skip("LEAPOCR_URL_UPLOAD_ENABLED not set; skipping URL upload test")
	}

	// Use environment variable if set, otherwise use a hardcoded test PDF URL
	testURL := os.Getenv("TEST_DOCUMENT_URL")
	if testURL == "" {
		// Hardcoded fallback: sample PDF file for testing
		testURL = "https://www.learningcontainer.com/download/sample-50-mb-pdf-file/?wpdmdl=3675&refresh=68c54927697581757759783"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	t.Logf("Processing URL: %s", testURL)
	job, err := sdk.ProcessURL(ctx, testURL,
		ocr.WithFormat(ocr.FormatMarkdown),
		ocr.WithModel(ocr.ModelStandardV1),
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
			ocr.WithFormat(ocr.FormatMarkdown))
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

func TestIntegration_DeleteJob(t *testing.T) {
	sdk := createTestSDK(t)
	templateSlug := os.Getenv("LEAPOCR_TEMPLATE_SLUG")

	// Find repo root and test file
	repoRoot := findRepoRoot(t)
	sampleDir := filepath.Join(repoRoot, "sample")
	testFiles := []string{
		filepath.Join(sampleDir, "test.pdf"),
		filepath.Join(sampleDir, "A129of19_14.01.22.pdf"),
	}

	var testFile string
	for _, file := range testFiles {
		if _, err := os.Stat(file); err == nil {
			testFile = file
			break
		}
	}

	if testFile == "" {
		t.Skip("No test files found in sample/ folder. Add sample PDF files to run this test.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Process a file
	file, err := os.Open(testFile)
	require.NoError(t, err)
	defer file.Close()

	t.Logf("Processing file for deletion test: %s", filepath.Base(testFile))
	structuredOptions := []ocr.ProcessingOption{}
	if templateSlug != "" {
		structuredOptions = append(structuredOptions, ocr.WithTemplateSlug(templateSlug))
	} else {
		structuredOptions = append(
			structuredOptions,
			ocr.WithFormat(ocr.FormatStructured),
			ocr.WithModel(ocr.ModelStandardV1),
			ocr.WithSchema(map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{"type": "string"},
				},
				"required": []interface{}{"text"},
			}),
		)
	}

	job, err := sdk.ProcessFile(ctx, file, filepath.Base(testFile), structuredOptions...)
	require.NoError(t, err)
	t.Logf("Job created: %s", job.ID)

	// Wait for completion
	result, err := sdk.WaitUntilDone(ctx, job.ID)
	require.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
	t.Logf("Job completed successfully")

	// Delete the job
	t.Logf("Deleting job: %s", job.ID)
	err = sdk.DeleteJob(ctx, job.ID)
	require.NoError(t, err)
	t.Logf("Job deleted successfully!")

	// Try to delete again - should fail or succeed (depending on API behavior)
	err = sdk.DeleteJob(ctx, job.ID)
	if err != nil {
		t.Logf("Second delete attempt returned error (expected): %v", err)
	} else {
		t.Logf("Second delete attempt succeeded (idempotent)")
	}
}

func TestIntegration_DeleteNonExistentJob(t *testing.T) {
	sdk := createTestSDK(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try to delete a non-existent job
	err := sdk.DeleteJob(ctx, "non-existent-job-id-12345")
	require.Error(t, err)
	t.Logf("Correctly handled deletion of non-existent job: %v", err)
}

func createTestSDK(t *testing.T) *ocr.SDK {
	apiKey := os.Getenv("LEAPOCR_API_KEY")
	if apiKey == "" {
		t.Fatal("LEAPOCR_API_KEY environment variable is required for integration tests")
	}

	baseURL := os.Getenv("OCR_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8443/api/v1"
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
