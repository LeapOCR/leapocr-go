package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	ocrsdk "github.com/your-org/ocr-go-sdk"
	"github.com/your-org/ocr-go-sdk/ocr"
)

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("OCR_API_KEY")
	if apiKey == "" {
		log.Fatal("OCR_API_KEY environment variable is required")
	}

	// Example: Custom configuration
	if err := customConfigExample(apiKey); err != nil {
		log.Printf("Custom config example failed: %v", err)
	}

	// Example: Batch processing
	if err := batchProcessingExample(apiKey); err != nil {
		log.Printf("Batch processing example failed: %v", err)
	}

	// Example: Analytics
	if err := analyticsExample(apiKey); err != nil {
		log.Printf("Analytics example failed: %v", err)
	}
}

func customConfigExample(apiKey string) error {
	fmt.Println("=== Custom Configuration Example ===")

	// Create custom configuration
	config := ocrsdk.NewConfig(apiKey)

	// Set custom base URL (if using a different environment)
	if err := config.SetBaseURL("https://api.example.com"); err != nil {
		return fmt.Errorf("failed to set base URL: %w", err)
	}

	// Configure timeouts and retries
	config.
		WithTimeout(60*time.Second).
		WithUserAgent("my-app/1.0.0").
		WithRetries(5, time.Second, 2*time.Minute)

	// Create client with custom config
	client := ocrsdk.NewWithConfig(config)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test health check with custom config
	health, err := client.Health.Check(ctx)
	if err != nil {
		// This might fail with the example URL, which is expected
		fmt.Printf("Health check failed (expected with example URL): %v\n", err)
	} else {
		fmt.Printf("Health status: %+v\n", health)
	}

	fmt.Println()
	return nil
}

func batchProcessingExample(apiKey string) error {
	fmt.Println("=== Batch Processing Example ===")

	client := ocrsdk.New(apiKey)

	// Example files to process (replace with real files)
	files := []string{
		"./document1.pdf",
		"./document2.pdf",
		"./document3.pdf",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Start processing all files
	jobIDs := make([]string, 0, len(files))

	for _, filePath := range files {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Printf("File %s not found, skipping\n", filePath)
			continue
		}

		req := &ocr.ProcessFileRequest{
			FilePath: filePath,
			Mode:     ocr.ModeTextAndImage,
			Priority: ocr.PriorityNormal,
		}

		job, err := client.OCR.ProcessFile(ctx, req)
		if err != nil {
			log.Printf("Failed to start processing %s: %v", filePath, err)
			continue
		}

		fmt.Printf("Started processing %s (Job ID: %s)\n", filePath, job.ID)
		jobIDs = append(jobIDs, job.ID)
	}

	if len(jobIDs) == 0 {
		fmt.Println("No files to process")
		return nil
	}

	// Wait for all jobs to complete
	fmt.Printf("Waiting for %d jobs to complete...\n", len(jobIDs))

	completed := make(map[string]bool)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			allComplete := true

			for _, jobID := range jobIDs {
				if completed[jobID] {
					continue
				}

				status, err := client.OCR.GetStatus(ctx, jobID)
				if err != nil {
					log.Printf("Failed to get status for job %s: %v", jobID, err)
					continue
				}

				switch status.Status {
				case ocr.StatusCompleted:
					result, err := client.OCR.GetResult(ctx, jobID)
					if err != nil {
						log.Printf("Failed to get result for job %s: %v", jobID, err)
					} else {
						fmt.Printf("Job %s completed - Credits: %d, Pages: %d\n",
							jobID, result.CreditsCost, len(result.Pages))
					}
					completed[jobID] = true
				case ocr.StatusFailed:
					errorMsg := "unknown error"
					if status.Error != nil {
						errorMsg = *status.Error
					}
					fmt.Printf("Job %s failed: %s\n", jobID, errorMsg)
					completed[jobID] = true
				default:
					allComplete = false
					fmt.Printf("Job %s: %s", jobID, status.Status)
					if status.Progress != nil {
						fmt.Printf(" (%.1f%%)", status.Progress.Percentage)
					}
					fmt.Println()
				}
			}

			if allComplete {
				fmt.Println("All jobs completed!")
				fmt.Println()
				return nil
			}
		}
	}
}

func analyticsExample(apiKey string) error {
	fmt.Println("=== Analytics Example ===")

	client := ocrsdk.New(apiKey)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get credits analytics
	fmt.Println("Fetching credits analytics...")
	creditsAnalytics, err := client.Analytics.GetCreditsAnalytics(ctx)
	if err != nil {
		return fmt.Errorf("failed to get credits analytics: %w", err)
	}
	fmt.Printf("Credits analytics: %+v\n", creditsAnalytics)

	// Get jobs analytics
	fmt.Println("Fetching jobs analytics...")
	jobsAnalytics, err := client.Analytics.GetJobsAnalytics(ctx)
	if err != nil {
		return fmt.Errorf("failed to get jobs analytics: %w", err)
	}
	fmt.Printf("Jobs analytics: %+v\n", jobsAnalytics)

	// Get job list
	fmt.Println("Fetching job list...")
	jobs, err := client.Jobs.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to get job list: %w", err)
	}
	fmt.Printf("Recent jobs: %+v\n", jobs)

	fmt.Println()
	return nil
}
