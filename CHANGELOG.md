# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-01-06

### Added
- Initial release of Email Checker Tool
- High-performance concurrent domain validation
- DNS record checking (MX, SPF, DMARC)
- Intelligent rate limiting to prevent DNS server overload
- Graceful shutdown with signal handling
- CSV streaming for efficient large-file processing
- Comprehensive CLI with configurable workers and timeouts
- Complete test suite and documentation

### Features
- Concurrent processing with configurable worker pools
- Built-in rate limiting (requests per second)
- Context-aware timeout handling
- CSV streaming with low memory footprint
- Graceful SIGINT handling
- Detailed validation reports