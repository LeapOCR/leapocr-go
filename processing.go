package ocr

import (
	"context"
	"fmt"
	"io"

	"github.com/leapocr/go-sdk/gen"
)

// ProcessURL starts OCR processing for a file at the given URL
func (s *SDK) ProcessURL(ctx context.Context, fileURL string, opts ...ProcessingOption) (*Job, error) {
	// Validate URL
	if err := ValidateURL(fileURL); err != nil {
		return nil, NewSDKError(ErrorTypeValidationError, "invalid URL", err)
	}

	config := applyProcessingOptions(opts)

	// Validate processing configuration
	if err := ValidateProcessingConfig(config); err != nil {
		return nil, NewSDKError(ErrorTypeValidationError, "invalid processing configuration", err)
	}

	// Create the URL upload request
	formatStr := string(config.format)
	tierStr := string(config.tier)
	request := gen.UploadURLUploadRequest{
		Url:    fileURL,
		Format: &formatStr,
		Tier:   &tierStr,
	}

	// Add optional fields if provided
	if config.instructions != "" {
		request.Instructions = &config.instructions
	}
	if config.schema != nil {
		request.Schema = config.schema
	}
	if config.schema != nil {
		// Convert schema to the format expected by the generated client
		// This would need to be adapted based on the actual generated types
		// For now, we'll skip this as it depends on the generated schema structure
	}

	// Make the API call using the generated client
	apiRequest := s.client.SDKAPI.UploadFromURL(ctx)
	apiRequest = apiRequest.UploadURLUploadRequest(request)

	resp, httpResp, err := apiRequest.Execute()
	if err != nil {
		return nil, s.handleAPIError(err, httpResp, "failed to start processing from URL")
	}

	// Extract job ID from response
	var jobID string
	if resp.JobId != nil {
		jobID = *resp.JobId
	}

	return &Job{
		ID:     jobID,
		Status: "processing",
	}, nil
}

// ProcessFile starts OCR processing for a file from an io.Reader
func (s *SDK) ProcessFile(ctx context.Context, file io.Reader, filename string, opts ...ProcessingOption) (*Job, error) {
	// Validate filename and extension
	if err := ValidateFileExtension(filename); err != nil {
		return nil, NewSDKError(ErrorTypeValidationError, "invalid filename", err)
	}

	config := applyProcessingOptions(opts)

	// Validate processing configuration
	if err := ValidateProcessingConfig(config); err != nil {
		return nil, NewSDKError(ErrorTypeValidationError, "invalid processing configuration", err)
	}

	// Step 1: Get presigned upload URL
	formatStr := string(config.format)
	tierStr := string(config.tier)
	uploadRequest := gen.UploadInitiateUploadRequest{
		FileName:    filename,
		ContentType: getContentType(filename),
		Format:      &formatStr,
		Tier:        &tierStr,
	}

	// Add optional fields if provided
	if config.instructions != "" {
		uploadRequest.Instructions = &config.instructions
	}
	if config.schema != nil {
		uploadRequest.Schema = config.schema
	}

	// Make the API call to get presigned URL
	apiRequest := s.client.SDKAPI.PresignedUpload(ctx)
	apiRequest = apiRequest.UploadInitiateUploadRequest(uploadRequest)

	resp, httpResp, err := apiRequest.Execute()
	if err != nil {
		return nil, s.handleAPIError(err, httpResp, "failed to initiate file upload")
	}

	var presignedURL, jobID string
	if resp.UploadUrl != nil {
		presignedURL = *resp.UploadUrl
	}
	if resp.JobId != nil {
		jobID = *resp.JobId
	}

	// Step 2: Upload file to presigned URL
	if err := s.uploadFile(ctx, presignedURL, file, filename); err != nil {
		return nil, NewSDKError(ErrorTypeUploadError, "failed to upload file", err)
	}

	return &Job{
		ID:     jobID,
		Status: "processing",
	}, nil
}

// handleAPIError converts generated client errors to SDK errors
func (s *SDK) handleAPIError(err error, httpResp interface{}, message string) *SDKError {
	// This would need to be implemented based on the actual generated error types
	// For now, we'll create a generic API error
	return NewSDKError(ErrorTypeAPIError, fmt.Sprintf("%s: %v", message, err), err)
}

// getContentType returns the content type based on filename
func getContentType(filename string) string {
	// Simple content type detection - could be enhanced with mime type detection
	if len(filename) > 4 && filename[len(filename)-4:] == ".pdf" {
		return "application/pdf"
	}
	if len(filename) > 4 && filename[len(filename)-4:] == ".png" {
		return "image/png"
	}
	if len(filename) > 4 && filename[len(filename)-4:] == ".jpg" {
		return "image/jpeg"
	}
	if len(filename) > 5 && filename[len(filename)-5:] == ".jpeg" {
		return "image/jpeg"
	}
	return "application/octet-stream"
}
