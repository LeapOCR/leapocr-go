package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	ocr "github.com/leapocr/leapocr-go"
)

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("LEAPOCR_API_KEY")
	if apiKey == "" {
		log.Fatal("LEAPOCR_API_KEY environment variable is required")
	}

	// Create OCR SDK
	sdk, err := ocr.New(apiKey)
	if err != nil {
		log.Fatalf("Failed to create SDK: %v", err)
	}

	// Example: Process a local PDF file
	if err := processLocalFile(sdk); err != nil {
		log.Printf("Failed to process local file: %v", err)
	}

	// Example: Process a file from URL
	if err := processFileFromURL(sdk); err != nil {
		log.Printf("Failed to process URL: %v", err)
	}
}

func processLocalFile(sdk *ocr.SDK) error {
	fmt.Println("=== Processing Local File ===")

	// Check if example file exists
	filePath := "./sample-document.pdf"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("Sample file %s not found, skipping local file example\n", filePath)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Start processing with options
	fmt.Printf("Starting OCR processing for %s...\n", filePath)
	job, err := sdk.ProcessFile(ctx, file, "sample-document.pdf",
		ocr.WithFormat(ocr.FormatStructured),
		ocr.WithModel(ocr.ModelStandardV1),
		ocr.WithInstructions("Extract all invoice details including amounts, dates, and vendor information"),
		ocr.WithSchema(map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"invoice_number": map[string]interface{}{"type": "string"},
				"invoice_date":   map[string]interface{}{"type": "string"},
				"total_amount":   map[string]interface{}{"type": "number"},
				"vendor_name":    map[string]interface{}{"type": "string"},
			},
			"required": []interface{}{"invoice_number", "total_amount"},
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to start processing: %w", err)
	}

	fmt.Printf("Job created with ID: %s\n", job.ID)

	// Wait for completion
	fmt.Println("Waiting for processing to complete...")
	result, err := sdk.WaitUntilDone(ctx, job.ID)
	if err != nil {
		return fmt.Errorf("processing failed: %w", err)
	}

	// Print results
	fmt.Printf("Processing completed successfully!\n")
	fmt.Printf("Credits used: %d\n", result.Credits)
	fmt.Printf("Processing time: %v\n", result.Duration)
	fmt.Printf("Pages processed: %d\n", len(result.Pages))

	// Print extracted data
	if result.Data != nil {
		fmt.Printf("Extracted data: %+v\n", result.Data)
	}

	// Print first page text (truncated)
	if len(result.Pages) > 0 && len(result.Pages[0].Text) > 0 {
		text := result.Pages[0].Text
		if len(text) > 200 {
			text = text[:200] + "..."
		}
		fmt.Printf("First page text: %s\n", text)
	}

	// Optional: Delete the job after processing
	fmt.Println("Deleting job...")
	if err := sdk.DeleteJob(ctx, job.ID); err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}
	fmt.Println("Job deleted successfully")

	fmt.Println()
	return nil
}

func processFileFromURL(sdk *ocr.SDK) error {
	fmt.Println("=== Processing File from URL ===")

	// Example URL (replace with a real URL)
	fileURL := "https://example.com/sample-invoice.pdf"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Printf("Starting OCR processing for URL: %s...\n", fileURL)
	job, err := sdk.ProcessURL(ctx, fileURL,
		ocr.WithFormat(ocr.FormatMarkdown),
		ocr.WithModel(ocr.ModelStandardV1),
		ocr.WithInstructions("Extract key financial information"),
	)
	if err != nil {
		return fmt.Errorf("failed to start URL processing: %w", err)
	}

	fmt.Printf("Job created with ID: %s\n", job.ID)

	// Poll for status updates manually (alternative to WaitForCompletion)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			status, err := sdk.GetJobStatus(ctx, job.ID)
			if err != nil {
				return fmt.Errorf("failed to get status: %w", err)
			}

			fmt.Printf("Status: %s", status.Status)
			if status.Progress != 0 {
				fmt.Printf(" (%.1f%% complete)", status.Progress)
			}
			fmt.Println()

			switch status.Status {
			case "completed":
				result, err := sdk.GetJobResult(ctx, job.ID)
				if err != nil {
					return fmt.Errorf("failed to get result: %w", err)
				}

				fmt.Printf("Processing completed!\n")
				fmt.Printf("Credits used: %d\n", result.Credits)
				fmt.Printf("Text length: %d characters\n", len(result.Text))
				return nil
			case "failed", "error":
				errorMsg := "unknown error"
				if status.Error != "" {
					errorMsg = status.Error
				}
				return fmt.Errorf("processing failed: %s", errorMsg)
			}
		}
	}
}
