# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
