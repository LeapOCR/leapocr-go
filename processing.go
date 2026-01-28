package ocr

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/leapocr/leapocr-go/internal/generated"
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
	uploadRequest := generated.UploadRemoteURLUploadRequest{
		Url: fileURL,
	}

	if !config.templateSlugSet {
		formatStr := string(config.format)
		uploadRequest.Format = &formatStr
	}

	// Add optional fields if provided
	if config.modelSet {
		uploadRequest.Model = &config.model
	}
	if config.instructionsSet {
		uploadRequest.Instructions = &config.instructions
	}
	if config.schemaSet {
		uploadRequest.Schema = config.schema
	}
	if config.templateSlugSet {
		uploadRequest.TemplateSlug = &config.templateSlug
	}

	// Make the API call using the generated client
	apiRequest := s.client.SDKAPI.UploadFromRemoteURL(ctx)
	apiRequest = apiRequest.UploadFromRemoteURLRequest(
		generated.UploadRemoteURLUploadRequestAsUploadFromRemoteURLRequest(&uploadRequest),
	)

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
	config := applyProcessingOptions(opts)

	// Validate and read file
	fileContent, fileSize32, err := s.validateAndReadFile(file, filename, config)
	if err != nil {
		return nil, err
	}

	// Build initiate request
	initiateRequest := s.buildInitiateRequest(filename, fileSize32, config)

	// Initiate direct upload and get response with presigned URLs
	uploadResp, err := s.initiateDirectUpload(ctx, initiateRequest)
	if err != nil {
		return nil, err
	}

	var jobID string
	if uploadResp.JobId != nil {
		jobID = *uploadResp.JobId
	}

	// Upload file parts to presigned URLs and collect ETags
	completedParts, err := s.uploadFileParts(ctx, uploadResp, io.NopCloser(bytes.NewReader(fileContent)))
	if err != nil {
		return nil, NewSDKError(ErrorTypeUploadError, "failed to upload file", err)
	}

	// Complete the multipart upload
	if err := s.completeDirectUpload(ctx, jobID, completedParts); err != nil {
		return nil, NewSDKError(ErrorTypeUploadError, "failed to complete upload", err)
	}

	return &Job{
		ID:     jobID,
		Status: "processing",
	}, nil
}

// validateAndReadFile validates the file and reads its content
func (s *SDK) validateAndReadFile(file io.Reader, filename string, config *processingConfig) ([]byte, int32, error) {
	if err := ValidateFileExtension(filename); err != nil {
		return nil, 0, NewSDKError(ErrorTypeValidationError, "invalid filename", err)
	}

	if err := ValidateProcessingConfig(config); err != nil {
		return nil, 0, NewSDKError(ErrorTypeValidationError, "invalid processing configuration", err)
	}

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, 0, NewSDKError(ErrorTypeUploadError, "failed to read file content", err)
	}

	fileSize := int64(len(fileContent))
	if fileSize == 0 {
		return nil, 0, NewSDKError(ErrorTypeValidationError, "file is empty", nil)
	}
	if fileSize > MaxFileSizeBytes {
		return nil, 0, NewSDKError(ErrorTypeValidationError,
			fmt.Sprintf("file size (%d bytes) exceeds maximum allowed size (%d bytes)", fileSize, MaxFileSizeBytes), nil)
	}

	const maxInt32 = 2147483647
	if fileSize > maxInt32 {
		return nil, 0, NewSDKError(ErrorTypeValidationError,
			fmt.Sprintf("file size (%d bytes) exceeds API limit (%d bytes)", fileSize, maxInt32), nil)
	}

	return fileContent, int32(fileSize), nil // #nosec G115 - validated above
}

// buildInitiateRequest builds the initiate direct upload request
func (s *SDK) buildInitiateRequest(filename string, fileSize32 int32, config *processingConfig) generated.UploadInitiateDirectUploadRequest {
	initiateRequest := generated.UploadInitiateDirectUploadRequest{
		FileName:    filename,
		ContentType: getContentType(filename),
		FileSize:    &fileSize32,
	}

	formatStr := string(config.format)
	if !config.templateSlugSet {
		initiateRequest.Format = &formatStr
	}

	if config.modelSet {
		initiateRequest.Model = &config.model
	}
	if config.instructionsSet {
		initiateRequest.Instructions = &config.instructions
	}
	if config.schemaSet {
		initiateRequest.Schema = config.schema
	}
	if config.templateSlugSet {
		initiateRequest.TemplateSlug = &config.templateSlug
	}

	return initiateRequest
}

// initiateDirectUpload initiates the direct upload and returns the response
func (s *SDK) initiateDirectUpload(ctx context.Context, initiateRequest generated.UploadInitiateDirectUploadRequest) (*generated.UploadDirectUploadResponse, error) {
	apiRequest := s.client.SDKAPI.DirectUpload(ctx)
	apiRequest = apiRequest.DirectUploadRequest(
		generated.UploadInitiateDirectUploadRequestAsDirectUploadRequest(&initiateRequest),
	)

	resp, httpResp, err := apiRequest.Execute()
	if err != nil {
		return nil, s.handleAPIError(err, httpResp, "failed to initiate file upload")
	}

	return resp, nil
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
