// Package ocr provides the official Go SDK for the OCR API
//
// This SDK provides a clean, Go-native interface for processing documents
// with OCR technology. It handles presigned URL uploads, polling for results,
// and provides comprehensive error handling.
//
// Basic usage:
//
//	sdk, err := ocr.New("your-api-key")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Process from URL
//	job, err := sdk.ProcessURL(ctx, "https://example.com/document.pdf",
//		ocr.WithFormat(ocr.FormatStructured),
//		ocr.WithTier(ocr.TierCore))
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Wait for completion
//	result, err := sdk.WaitForCompletion(ctx, job.ID)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("Extracted: %+v\n", result.Data)
package ocr

// All types and functions are defined in their respective files:
// - sdk.go: Main SDK struct and constructors
// - types.go: Result types, enums, and options
// - processing.go: ProcessURL and ProcessFile methods
// - waiter.go: WaitForCompletion and polling logic
// - upload.go: File upload handling
// - errors.go: Error types and handling
