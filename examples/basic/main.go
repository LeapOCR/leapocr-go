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

	// Create OCR client
	client := ocrsdk.New(apiKey)

	// Example: Process a local PDF file
	if err := processLocalFile(client); err != nil {
		log.Printf("Failed to process local file: %v", err)
	}

	// Example: Process a file from URL
	if err := processFileFromURL(client); err != nil {
		log.Printf("Failed to process URL: %v", err)
	}

	// Example: Check health
	if err := checkHealth(client); err != nil {
		log.Printf("Health check failed: %v", err)
	}
}

func processLocalFile(client *ocrsdk.Client) error {
	fmt.Println("=== Processing Local File ===")

	// Check if example file exists (you'll need to provide one)
	filePath := "./sample-document.pdf"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("Sample file %s not found, skipping local file example\n", filePath)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create process request
	req := &ocr.ProcessFileRequest{
		FilePath:           filePath,
		Mode:               ocr.ModeTextAndImage,
		Schema:             "invoice", // optional: use predefined schema
		CustomInstructions: "Extract all invoice details including amounts, dates, and vendor information",
		Priority:           ocr.PriorityNormal,
	}

	// Start processing
	fmt.Printf("Starting OCR processing for %s...\n", filePath)
	job, err := client.OCR.ProcessFile(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to start processing: %w", err)
	}

	fmt.Printf("Job created with ID: %s\n", job.ID)

	// Wait for completion
	fmt.Println("Waiting for processing to complete...")
	result, err := client.OCR.WaitForCompletion(ctx, job.ID)
	if err != nil {
		return fmt.Errorf("processing failed: %w", err)
	}

	// Print results
	fmt.Printf("Processing completed successfully!\n")
	fmt.Printf("Credits used: %d\n", result.CreditsCost)
	fmt.Printf("Processing time: %v\n", result.ProcessingTime)
	fmt.Printf("Pages processed: %d\n", len(result.Pages))

	// Print extracted data
	if result.Data != nil {
		fmt.Printf("Extracted data: %+v\n", result.Data)
	}

	fmt.Println()
	return nil
}

func processFileFromURL(client *ocrsdk.Client) error {
	fmt.Println("=== Processing File from URL ===")

	// Example URL (replace with a real URL)
	fileURL := "https://example.com/sample-invoice.pdf"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Options for URL processing
	opts := &ocr.ProcessFileRequest{
		Mode:               ocr.ModeTextOnly,
		CustomInstructions: "Extract key financial information",
		Priority:           ocr.PriorityHigh,
	}

	fmt.Printf("Starting OCR processing for URL: %s...\n", fileURL)
	job, err := client.OCR.ProcessURL(ctx, fileURL, opts)
	if err != nil {
		return fmt.Errorf("failed to start URL processing: %w", err)
	}

	fmt.Printf("Job created with ID: %s\n", job.ID)

	// Poll for status updates
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			status, err := client.OCR.GetStatus(ctx, job.ID)
			if err != nil {
				return fmt.Errorf("failed to get status: %w", err)
			}

			fmt.Printf("Status: %s", status.Status)
			if status.Progress != nil {
				fmt.Printf(" (%.1f%% complete)", status.Progress.Percentage)
			}
			fmt.Println()

			if status.Status == ocr.StatusCompleted {
				result, err := client.OCR.GetResult(ctx, job.ID)
				if err != nil {
					return fmt.Errorf("failed to get result: %w", err)
				}

				fmt.Printf("Processing completed!\n")
				fmt.Printf("Credits used: %d\n", result.CreditsCost)
				fmt.Printf("Text length: %d characters\n", len(result.Pages[0].Text))
				return nil
			} else if status.Status == ocr.StatusFailed {
				errorMsg := "unknown error"
				if status.Error != nil {
					errorMsg = *status.Error
				}
				return fmt.Errorf("processing failed: %s", errorMsg)
			}
		}
	}
}

func checkHealth(client *ocrsdk.Client) error {
	fmt.Println("=== Health Check ===")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	health, err := client.Health.Check(ctx)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	fmt.Printf("Health status: %+v\n", health)
	fmt.Println()
	return nil
}
