# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
