package ocr

import (
	"context"
)

// DeleteJob soft deletes an OCR job by redacting all page content to [REDACTED],
// deleting associated files from storage, and marking the job as deleted.
// The job will no longer be accessible via normal fetch endpoints but will
// appear in job listings with a deleted flag.
func (s *SDK) DeleteJob(ctx context.Context, jobID string) error {
	if jobID == "" {
		return NewSDKError(ErrorTypeValidationError, "job ID is required", nil)
	}

	// Make the API call to delete the job
	apiRequest := s.client.JobsAPI.DeleteJob(ctx, jobID)
	_, httpResp, err := apiRequest.Execute()
	if err != nil {
		return s.handleAPIError(err, httpResp, "failed to delete job")
	}

	return nil
}
