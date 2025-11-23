# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.5] - 2025-11-23

### Changed

- Updated local development server URL from `http://localhost:8080` to `http://localhost:8443`
- Refactored Jobs API to remove deprecated methods and improve RetryJob functionality
- Enhanced validation tests to accommodate changes in API responses and structures

### Added

- New Onboarding API endpoints for managing onboarding status (creation, retrieval, and updates)

### Fixed

- License file formatting

## [0.0.4] - 2025-11-11

- Add Apache License 2.0
- add DeleteJob function and update processing options






## [0.0.4] - 2025-11-11

### Breaking Changes

- **Template Parameter Update**: Replaced `template_id` with `template_slug` in processing options
  - Use `WithTemplateSlug()` instead of template ID references
  - Template slugs provide more readable and stable identifiers

### Added

- `DeleteJob()` method to soft delete OCR jobs and redact sensitive content
- `WithTemplateSlug()` option for using pre-configured templates
- Enhanced validation for template slug parameter
- Automatic cleanup examples for sensitive data handling

### Changed

- Updated processing options to use template slugs for structured extraction
- Enhanced API validation functions to support template slug functionality
- Updated README and examples with DeleteJob usage patterns

## [0.0.3] - 2025-11-07

### Breaking Changes

- **Tier → Model Migration**: Replaced `WithTier()` option with `WithModel()` for OCR model selection
  - `TierSwift`, `TierCore`, `TierIntelli` → `ModelStandardV1`, `ModelEnglishProV1`, `ModelProV1`
  - Added `WithModelString()` for custom model names
- **Removed Confidence Fields**: Removed `PageResult.Confidence` and `OCRResult` confidence-related fields
- **Authentication**: SDK now uses `X-API-KEY` header for authentication (API still supports `Authorization: Bearer` for direct API usage)
- **Upload Mechanism**: Migrated from single presigned URL uploads to multipart direct uploads with ETag handling

### Changed

- Progress calculation now uses `ProcessedPages/TotalPages` instead of `ProgressPercentage`
- Improved error messages to include API error details
- Enhanced status handling to support both string and object formats
- Updated default server configuration to prioritize production HTTPS endpoint
- Improved ETag extraction and handling for multipart uploads

### Added

- Support for multipart direct uploads with automatic chunking
- `WithModel()` and `WithModelString()` options for flexible model selection
- Better error context in upload operations
- Java PATH setup script for OpenAPI generator

### Removed

- `uploadFile()` function (replaced by multipart upload system)
- Confidence-related fields from result types

## [0.0.1] - 2025-09-12

### Added

- Initial version of the LeapOCR Go SDK.
- `README.md` with installation and usage instructions.
- `CONTRIBUTING.md` with guidelines for contributors.
- `RELEASE.md` with instructions for creating a new release.
- Basic and advanced examples in the `examples/` directory.
- CI/CD workflows for linting, testing, and releasing.
- Robust error handling with custom error types.

### Changed

- Updated `BaseURL` to `https://api.leapocr.com`.
- Updated `UserAgent` to be dynamic.
- Improved `README.md` with more details and a better structure.

### Fixed

- Fixed the documentation deployment trigger in the release workflow.
- Fixed the examples to use a real sample PDF and a valid URL.
