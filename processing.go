package ocr

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/leapocr/leapocr-go/gen"
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
	request := gen.UploadRemoteURLUploadRequest{
		Url:    fileURL,
		Format: &formatStr,
	}

	// Add optional fields if provided
	if config.model != "" {
		request.Model = &config.model
	}
	if config.instructions != "" {
		request.Instructions = &config.instructions
	}
	if config.schema != nil {
		request.Schema = config.schema
	}

	// Make the API call using the generated client
	apiRequest := s.client.SDKAPI.UploadFromRemoteURL(ctx)
	apiRequest = apiRequest.UploadRemoteURLUploadRequest(request)

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

	// Read file content to get size (required for chunk calculation)
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, NewSDKError(ErrorTypeUploadError, "failed to read file content", err)
	}

	fileSize := int64(len(fileContent))
	if fileSize == 0 {
		return nil, NewSDKError(ErrorTypeValidationError, "file is empty", nil)
	}
	if fileSize > MaxFileSizeBytes {
		return nil, NewSDKError(ErrorTypeValidationError,
			fmt.Sprintf("file size (%d bytes) exceeds maximum allowed size (%d bytes)", fileSize, MaxFileSizeBytes), nil)
	}

	// Step 1: Get presigned upload URLs for multipart upload
	formatStr := string(config.format)

	// Validate file size fits in int32 (API requirement)
	const maxInt32 = 2147483647
	if fileSize > maxInt32 {
		return nil, NewSDKError(ErrorTypeValidationError,
			fmt.Sprintf("file size (%d bytes) exceeds API limit (%d bytes)", fileSize, maxInt32), nil)
	}
	fileSize32 := int32(fileSize) // #nosec G115 - validated above

	uploadRequest := gen.UploadInitiateDirectUploadRequest{
		FileName:    filename,
		ContentType: getContentType(filename),
		Format:      &formatStr,
		FileSize:    &fileSize32,
	}

	// Add optional fields if provided
	if config.model != "" {
		uploadRequest.Model = &config.model
	}
	if config.instructions != "" {
		uploadRequest.Instructions = &config.instructions
	}
	if config.schema != nil {
		uploadRequest.Schema = config.schema
	}

	// Make the API call to get presigned URLs
	apiRequest := s.client.SDKAPI.DirectUpload(ctx)
	apiRequest = apiRequest.UploadInitiateDirectUploadRequest(uploadRequest)

	resp, httpResp, err := apiRequest.Execute()
	if err != nil {
		return nil, s.handleAPIError(err, httpResp, "failed to initiate file upload")
	}

	var jobID string
	if resp.JobId != nil {
		jobID = *resp.JobId
	}

	// Step 2: Upload file parts to presigned URLs and collect ETags
	// Pass file content as a reader since we already read it
	completedParts, err := s.uploadFileParts(ctx, resp, io.NopCloser(bytes.NewReader(fileContent)))
	if err != nil {
		return nil, NewSDKError(ErrorTypeUploadError, "failed to upload file", err)
	}

	// Step 3: Complete the multipart upload
	if err := s.completeDirectUpload(ctx, jobID, completedParts); err != nil {
		return nil, NewSDKError(ErrorTypeUploadError, "failed to complete upload", err)
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
