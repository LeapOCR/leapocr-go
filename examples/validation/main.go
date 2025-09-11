package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	ocr "github.com/leapocr/go-sdk"
)

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("LEAPOCR_API_KEY")
	if apiKey == "" {
		log.Fatal("LEAPOCR_API_KEY environment variable is required")
	}

	// Example: Input validation
	if err := inputValidationExample(apiKey); err != nil {
		log.Printf("Input validation example failed: %v", err)
	}

	// Example: Error handling
	if err := errorHandlingExample(apiKey); err != nil {
		log.Printf("Error handling example failed: %v", err)
	}

	// Example: Timeout handling
	if err := timeoutHandlingExample(apiKey); err != nil {
		log.Printf("Timeout handling example failed: %v", err)
	}
}

func inputValidationExample(apiKey string) error {
	fmt.Println("=== Input Validation Example ===")

	// Test creating SDK with empty API key (should fail)
	fmt.Println("1. Testing empty API key...")
	_, err := ocr.New("")
	if err != nil {
		fmt.Printf("✓ Correctly rejected empty API key: %v\n", err)
	} else {
		fmt.Println("✗ Should have rejected empty API key")
	}

	// Test creating SDK with valid configuration
	fmt.Println("2. Testing valid configuration...")
	sdk, err := ocr.New(apiKey)
	if err != nil {
		return fmt.Errorf("failed to create SDK: %w", err)
	}
	fmt.Println("✓ SDK created successfully")

	// Test invalid URL processing
	fmt.Println("3. Testing invalid URL...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = sdk.ProcessURL(ctx, "not-a-valid-url",
		ocr.WithFormat(ocr.FormatStructured),
	)
	if err != nil {
		fmt.Printf("✓ Correctly rejected invalid URL: %v\n", err)
	} else {
		fmt.Println("✗ Should have rejected invalid URL")
	}

	// Test empty job ID
	fmt.Println("4. Testing empty job ID...")
	_, err = sdk.GetJobStatus(ctx, "")
	if err != nil {
		fmt.Printf("✓ Correctly rejected empty job ID: %v\n", err)
	} else {
		fmt.Println("✗ Should have rejected empty job ID")
	}

	fmt.Println()
	return nil
}

func errorHandlingExample(apiKey string) error {
	fmt.Println("=== Error Handling Example ===")

	sdk, err := ocr.New(apiKey)
	if err != nil {
		return fmt.Errorf("failed to create SDK: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test processing with invalid URL (will likely fail)
	fmt.Println("1. Testing error handling for invalid processing...")
	_, err = sdk.ProcessURL(ctx, "https://nonexistent-domain-12345.com/fake.pdf",
		ocr.WithFormat(ocr.FormatStructured),
	)
	if err != nil {
		// Demonstrate error type checking
		if sdkErr, ok := err.(*ocr.SDKError); ok {
			fmt.Printf("✓ Received SDK error: %s\n", sdkErr.Type)
			fmt.Printf("  Message: %s\n", sdkErr.Message)

			if sdkErr.IsHTTPError() {
				fmt.Printf("  HTTP Status: %d\n", sdkErr.StatusCode)
			}

			if sdkErr.IsRetryable() {
				fmt.Println("  This error is retryable")
			} else {
				fmt.Println("  This error is not retryable")
			}
		} else {
			fmt.Printf("✓ Received generic error: %v\n", err)
		}
	}

	// Test getting status for non-existent job
	fmt.Println("2. Testing error handling for non-existent job...")
	_, err = sdk.GetJobStatus(ctx, "non-existent-job-id-12345")
	if err != nil {
		fmt.Printf("✓ Correctly handled non-existent job: %v\n", err)
	}

	fmt.Println()
	return nil
}

func timeoutHandlingExample(apiKey string) error {
	fmt.Println("=== Timeout Handling Example ===")

	sdk, err := ocr.New(apiKey)
	if err != nil {
		return fmt.Errorf("failed to create SDK: %w", err)
	}

	// Test with very short timeout (should fail quickly)
	fmt.Println("1. Testing short timeout...")
	shortCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err = sdk.ProcessURL(shortCtx, "https://httpbin.org/delay/2", // Delayed response
		ocr.WithFormat(ocr.FormatStructured),
	)
	duration := time.Since(start)

	if err != nil {
		if strings.Contains(err.Error(), "context") || strings.Contains(err.Error(), "timeout") {
			fmt.Printf("✓ Correctly timed out after %v: %v\n", duration, err)
		} else {
			fmt.Printf("✓ Failed quickly (%v): %v\n", duration, err)
		}
	} else {
		fmt.Printf("✗ Expected timeout, but succeeded in %v\n", duration)
	}

	// Test custom wait options with short timeout
	fmt.Println("2. Testing custom wait options...")
	normalCtx, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()

	// Create a fake job to test waiting behavior
	// (This will fail, but demonstrates the timeout handling)
	waitOpts := ocr.WaitOptions{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		MaxJitter:    200 * time.Millisecond,
		MaxAttempts:  3, // Very few attempts
	}

	start = time.Now()
	_, err = sdk.WaitUntilDoneWithOptions(normalCtx, "fake-job-id", waitOpts)
	duration = time.Since(start)

	if err != nil {
		fmt.Printf("✓ Wait with custom options failed after %v: %v\n", duration, err)
	} else {
		fmt.Printf("✗ Expected failure, but succeeded in %v\n", duration)
	}

	fmt.Println()
	return nil
}
