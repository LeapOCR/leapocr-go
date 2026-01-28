package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	ocr "github.com/leapocr/leapocr-go"
)

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("LEAPOCR_API_KEY")
	if apiKey == "" {
		log.Fatal("LEAPOCR_API_KEY environment variable is required")
	}

	// Example: Custom configuration
	if err := customConfigExample(apiKey); err != nil {
		log.Printf("Custom config example failed: %v", err)
	}

	// Example: Batch processing with goroutines
	if err := batchProcessingExample(apiKey); err != nil {
		log.Printf("Batch processing example failed: %v", err)
	}

	// Example: Schema-based extraction
	if err := schemaExtractionExample(apiKey); err != nil {
		log.Printf("Schema extraction example failed: %v", err)
	}
}

func customConfigExample(apiKey string) error {
	fmt.Println("=== Custom Configuration Example ===")

	// Create custom configuration
	config := ocr.DefaultConfig(apiKey)
	config.BaseURL = "https://api-staging.ocr.example.com" // Example staging URL
	config.Timeout = 60 * time.Second
	config.UserAgent = "my-app/1.0.0"

	// Create SDK with custom config
	sdk, err := ocr.NewSDK(config)
	if err != nil {
		return fmt.Errorf("failed to create SDK with custom config: %w", err)
	}

	// Test with a dummy operation (this will likely fail, which is expected)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Example URL processing with custom options
	_, err = sdk.ProcessURL(ctx, "https://example.com/test.pdf",
		ocr.WithFormat(ocr.FormatStructured),
		ocr.WithModel(ocr.ModelProV1),
		ocr.WithInstructions("Extract all data with high accuracy"),
		ocr.WithSchema(map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"text": map[string]interface{}{"type": "string"},
			},
			"required": []interface{}{"text"},
		}),
	)
	if err != nil {
		// This is expected to fail with the example URL
		fmt.Printf("Expected failure with example URL: %v\n", err)
	}

	fmt.Println()
	return nil
}

func batchProcessingExample(apiKey string) error {
	fmt.Println("=== Concurrent Batch Processing Example ===")

	sdk, err := ocr.New(apiKey)
	if err != nil {
		return fmt.Errorf("failed to create SDK: %w", err)
	}

	// Example files to process (replace with real files or URLs)
	files := []string{
		"https://example.com/document1.pdf",
		"https://example.com/document2.pdf",
		"https://example.com/document3.pdf",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Process files concurrently
	var wg sync.WaitGroup
	results := make(chan *ocr.OCRResult, len(files))
	errors := make(chan error, len(files))

	for i, fileURL := range files {
		wg.Add(1)
		go func(idx int, url string) {
			defer wg.Done()

			fmt.Printf("Starting processing for file %d: %s\n", idx+1, url)

			// Start processing
			job, err := sdk.ProcessURL(ctx, url,
				ocr.WithFormat(ocr.FormatStructured),
				ocr.WithModel(ocr.ModelStandardV1),
				ocr.WithInstructions(fmt.Sprintf("Process document %d", idx+1)),
				ocr.WithSchema(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"text": map[string]interface{}{"type": "string"},
					},
					"required": []interface{}{"text"},
				}),
			)
			if err != nil {
				errors <- fmt.Errorf("failed to start processing file %d: %w", idx+1, err)
				return
			}

			// Wait for completion with custom options
			waitOpts := ocr.WaitOptions{
				InitialDelay: 2 * time.Second,
				MaxDelay:     30 * time.Second,
				Multiplier:   2.0,
				MaxJitter:    5 * time.Second,
				MaxAttempts:  50,
			}

			result, err := sdk.WaitUntilDoneWithOptions(ctx, job.ID, waitOpts)
			if err != nil {
				errors <- fmt.Errorf("failed to complete processing file %d: %w", idx+1, err)
				return
			}

			fmt.Printf("Completed processing file %d (Job ID: %s)\n", idx+1, job.ID)

			// Optional: Delete the job after processing
			if err := sdk.DeleteJob(ctx, job.ID); err != nil {
				fmt.Printf("Warning: Failed to delete job %s: %v\n", job.ID, err)
			}

			results <- result
		}(i, fileURL)
	}

	// Close channels when all goroutines finish
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	// Collect results
	var successCount int
	var totalCredits int

	// Process results and errors
	for results != nil || errors != nil {
		select {
		case result, ok := <-results:
			if !ok {
				results = nil
				continue
			}
			successCount++
			totalCredits += result.Credits
			fmt.Printf("[SUCCESS] Processing completed - Credits: %d, Pages: %d\n",
				result.Credits, len(result.Pages))

		case err, ok := <-errors:
			if !ok {
				errors = nil
				continue
			}
			fmt.Printf("[FAILED] Processing failed: %v\n", err)

		case <-ctx.Done():
			return fmt.Errorf("batch processing timed out: %w", ctx.Err())
		}
	}

	fmt.Printf("\nBatch processing complete:\n")
	fmt.Printf("  Successfully processed: %d/%d files\n", successCount, len(files))
	fmt.Printf("  Total credits used: %d\n", totalCredits)
	fmt.Println()

	return nil
}

func schemaExtractionExample(apiKey string) error {
	fmt.Println("=== Schema-Based Extraction Example ===")

	sdk, err := ocr.New(apiKey)
	if err != nil {
		return fmt.Errorf("failed to create SDK: %w", err)
	}

	// Define custom schema for invoice extraction
	invoiceSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"invoice_number": map[string]interface{}{
				"type":        "string",
				"description": "The invoice number",
			},
			"total_amount": map[string]interface{}{
				"type":        "number",
				"description": "The total amount of the invoice",
			},
			"vendor_name": map[string]interface{}{
				"type":        "string",
				"description": "The name of the vendor/supplier",
			},
			"due_date": map[string]interface{}{
				"type":        "string",
				"format":      "date",
				"description": "The due date for payment",
			},
			"line_items": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"description": map[string]interface{}{"type": "string"},
						"quantity":    map[string]interface{}{"type": "number"},
						"unit_price":  map[string]interface{}{"type": "number"},
						"total":       map[string]interface{}{"type": "number"},
					},
				},
			},
		},
		"required": []interface{}{"invoice_number", "total_amount", "vendor_name"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Example URL (replace with a real invoice URL)
	invoiceURL := "https://example.com/sample-invoice.pdf"

	fmt.Printf("Processing invoice with custom schema: %s\n", invoiceURL)

	job, err := sdk.ProcessURL(ctx, invoiceURL,
		ocr.WithFormat(ocr.FormatStructured),
		ocr.WithModel(ocr.ModelProV1), // Use highest quality model for best accuracy
		ocr.WithSchema(invoiceSchema),
		ocr.WithInstructions("Extract invoice data according to the provided schema. Be precise with numbers and dates."),
	)
	if err != nil {
		// This will likely fail with the example URL, which is expected
		fmt.Printf("Expected failure with example URL: %v\n", err)
		fmt.Println("In a real scenario, this would process the invoice and extract:")
		fmt.Println("- Invoice number")
		fmt.Println("- Total amount")
		fmt.Println("- Vendor name")
		fmt.Println("- Due date")
		fmt.Println("- Line items with quantities and prices")
		fmt.Println()
		return nil
	}

	// If it somehow succeeded, wait for completion
	result, err := sdk.WaitUntilDone(ctx, job.ID)
	if err != nil {
		return fmt.Errorf("processing failed: %w", err)
	}

	fmt.Printf("Schema-based extraction completed!\n")
	fmt.Printf("Credits used: %d\n", result.Credits)
	fmt.Printf("Extracted data: %+v\n", result.Data)

	// Optional: Delete the job after processing
	if err := sdk.DeleteJob(ctx, job.ID); err != nil {
		fmt.Printf("Warning: Failed to delete job: %v\n", err)
	} else {
		fmt.Println("Job deleted successfully")
	}

	fmt.Println()

	return nil
}
