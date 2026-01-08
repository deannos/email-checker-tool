# Email Checker Tool

<div align="center">

[![Go Version](https://img.shields.io/badge/go-1.21%2B-blue.svg)](https://golang.org/doc/devel/release)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/deannos/email-checker-tool.svg)](https://github.com/deannos/email-checker-tool/releases)

A high-performance, production-grade CLI tool for comprehensive email domain validation. Verify DNS configurations (MX, SPF, DMARC) at scale with concurrent processing and intelligent rate limiting.

[Features](#features) • [Installation](#installation) • [Usage](#usage) • [Architecture](#architecture) • [Contributing](#contributing)

</div>

---

## Overview

Email Checker Tool is an open-source command-line utility designed to validate email domain configurations at scale. Whether you're managing mailing lists, validating domains in bulk, or implementing email verification workflows, this tool provides a robust, efficient solution for DNS-based email infrastructure audits.

Built with Go, it leverages concurrent processing patterns to handle thousands of domains rapidly while respecting DNS server limitations through intelligent rate limiting. The tool is designed for both operational simplicity and production reliability.

---

## Features

### High-Performance Processing
- **Concurrent Workers:** Configurable worker pools for parallel domain processing
- **Optimized DNS Lookups:** Efficient resolution of MX, SPF, and DMARC records
- **Minimal Memory Footprint:** Streaming CSV processing prevents memory bloat on large datasets

### Intelligent Rate Limiting
- **Configurable RPS (Requests Per Second):** Prevent IP blocks and DNS server overload
- **Smart Throttling:** Respects rate limits globally across all workers
- **Server-Friendly:** Designed to operate responsibly within DNS infrastructure constraints

### Robust Timeout Management
- **Context-Aware Timeouts:** Prevents hanging on unresponsive DNS servers
- **Configurable Durations:** Global operation timeout with per-request limits
- **Graceful Degradation:** Handles timeouts without crashing the application

### Comprehensive Email Validation
- **MX Record Verification:** Confirms domain has valid mail exchange servers
- **SPF Record Detection:** Identifies Sender Policy Framework configurations
- **DMARC Policy Analysis:** Validates Domain-based Message Authentication, Reporting, and Conformance policies
- **Detailed Error Reporting:** Captures validation issues for troubleshooting

### CSV-Based Workflow
- **Streaming Input Processing:** Handles CSV files of any size efficiently
- **Structured Output:** Machine-readable results for downstream processing
- **Flexible Column Mapping:** Automatically detects domain column in input files

### Safe Shutdown
- **Graceful Signal Handling:** SIGINT (Ctrl+C) captures and safely flushes results
- **Progress Preservation:** Saves processed results before terminating
- **Data Integrity:** Ensures no results are lost during shutdown

---

## Installation

### Prerequisites

- **Go:** Version 1.21 or higher
- **Environment:** Linux, macOS, or Windows

### From Source

```bash
git clone https://github.com/deannos/email-checker-tool.git
cd email-checker-tool
go install ./cmd/email-checker@latest
```

Or directly install the latest release:

```bash
go install github.com/deannos/email-checker-tool@latest
```

### From Binary Release

Download precompiled binaries for your platform from the [Releases](https://github.com/deannos/email-checker-tool/releases) page:

- Linux (x86_64, ARM64)
- macOS (Intel, Apple Silicon)
- Windows (x86_64)

Extract and add to your `$PATH`:

```bash
# Linux/macOS
tar -xzf email-checker-tool_<version>_<os>_<arch>.tar.gz
sudo mv email-checker-tool /usr/local/bin/

# Windows
# Extract the .zip file and add the directory to your PATH
```

### Verify Installation

```bash
email-checker --version
```

---

## Quick Start

### Basic Usage

Create a CSV file with domains to check:

```csv
domain
google.com
github.com
example.com
```

Run the checker:

```bash
email-checker domains.csv
```

Results are saved to `output.csv` by default.

### Advanced Usage

```bash
email-checker input.csv \
  --workers 20 \
  --rps 50 \
  --timeout 10s \
  --output results.csv
```

---

## Configuration

### CLI Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--workers` | `int` | `10` | Number of concurrent workers for parallel processing |
| `--rps` | `int` | `20` | Maximum DNS requests per second (rate limit) |
| `--timeout` | `string` | `5s` | Global timeout for the entire operation (e.g., `30s`, `5m`) |
| `--output` | `string` | `output.csv` | Path to the output CSV file |
| `--version` | `bool` | `false` | Print version information and exit |

### Configuration Guidelines

**Tuning for Performance:**
- Increase `--workers` (20-50) for larger domain lists
- Adjust `--rps` based on your network and target DNS servers (typically 10-100)
- Increase `--timeout` for operations processing 10,000+ domains (consider `30s` or `60s`)

**Tuning for Safety:**
- Reduce `--workers` to 5-10 for shared infrastructure
- Lower `--rps` to 10-15 to be conservative with external DNS servers
- Monitor logs for timeouts and adjust accordingly

---

## Input and Output

### Input Format

The input CSV file must contain a `domain` column. Additional columns are preserved in the output.

**Example Input (domains.csv):**

```csv
domain,company,contact
google.com,Google,support@google.com
github.com,GitHub,support@github.com
example.com,Example Corp,admin@example.com
```

### Output Format

The tool generates a CSV file with validation results for each domain.

**Example Output (output.csv):**

```csv
domain,hasMX,hasSPF,spfRecord,hasDMARC,dmarcRecord,error
google.com,true,true,"v=spf1 include:google.com ~all",true,"v=DMARC1; p=none",
github.com,true,true,"v=spf1 include:github.com ~all",true,"v=DMARC1; p=quarantine",
example.com,true,false,"",false,"","DMARC lookup failed: timeout"
```

### Output Field Reference

| Field | Type | Description |
|-------|------|-------------|
| `domain` | string | The domain being validated |
| `hasMX` | boolean | Whether the domain has MX records |
| `hasSPF` | boolean | Whether the domain has SPF record |
| `spfRecord` | string | Full SPF record value if present |
| `hasDMARC` | boolean | Whether the domain has DMARC policy |
| `dmarcRecord` | string | Full DMARC record value if present |
| `error` | string | Any errors encountered during validation |

---

## Use Cases

### Email List Validation
Verify thousands of domains before importing into email marketing platforms:

```bash
email-checker email_list.csv --workers 30 --rps 50 --timeout 30s
```

### Security Audits
Identify domains missing SPF or DMARC configurations for compliance:

```bash
email-checker company_domains.csv --output security_audit.csv
```

### DNS Infrastructure Testing
Test DNS resolver performance and reliability across a domain list:

```bash
email-checker test_domains.csv --workers 50 --timeout 60s
```

### Bulk Email Sender Validation
Pre-validate sender domains before implementing email authentication:

```bash
email-checker senders.csv --rps 30 --output validated_senders.csv
```

---

## Architecture

The project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout) for maintainability and clarity.

### Directory Structure

```
.
├── cmd/
│   └── email-checker/
│       └── main.go                 # Application entry point
├── internal/
│   ├── checker/
│   │   └── checker.go              # DNS lookup logic (MX, SPF, DMARC)
│   ├── worker/
│   │   └── pool.go                 # Worker pool with rate limiting
│   └── output/
│       └── csv.go                  # CSV writing and result handling
├── go.mod
├── go.sum
├── README.md
├── LICENSE
└── .gitignore
```

### Core Components

#### `cmd/email-checker`
- The application entry point. Handles CLI argument parsing, initializes workers, and orchestrates the validation pipeline.

#### `internal/checker`
- DNS resolution logic for MX, SPF, and DMARC records. Implements clean separation between network operations and business logic.

#### `internal/worker`
- Worker pool implementation with built-in rate limiting. Manages concurrent domain processing while respecting rate limits globally.

#### `internal/output`
- Handles CSV streaming and result persistence. Implements efficient buffering for large-scale operations.

### Design Principles

- **Concurrency:** Goroutines for parallel processing with controlled pooling
- **Rate Limiting:** Token bucket algorithm for predictable rate control
- **Error Handling:** Comprehensive error capture without halting execution
- **Resource Management:** Graceful shutdown and context cancellation

---

## Development

### Prerequisites

- Go 1.21 or higher
- Git
- Make (optional, for build automation)

### Building from Source

```bash
git clone https://github.com/deannos/email-checker-tool.git
cd email-checker-tool
go build -o bin/email-checker ./cmd/email-checker
```

Run the compiled binary:

```bash
./bin/email-checker domains.csv
```

### Running Tests

Execute the full test suite:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test -v ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

Generate coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Quality

Ensure code follows Go conventions:

```bash
# Format code
go fmt ./...

# Lint code (requires golangci-lint)
golangci-lint run ./...

# Run vet
go vet ./...
```

---

## Contributing

We welcome contributions from the community! Whether you're reporting bugs, suggesting features, or submitting code improvements, your help makes this project better.

### Getting Started

1. **Fork the Repository**
   ```bash
   # Click "Fork" on the GitHub repository page
   ```

2. **Clone Your Fork**
   ```bash
   git clone https://github.com/YOUR_USERNAME/email-checker-tool.git
   cd email-checker-tool
   ```

3. **Create a Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

4. **Make Your Changes**
    - Write clear, idiomatic Go code
    - Add or update tests for new functionality
    - Update documentation as needed

5. **Commit Your Changes**
   ```bash
   git commit -m "feat: add your feature description"
   ```

   Use conventional commit messages:
    - `feat:` for new features
    - `fix:` for bug fixes
    - `docs:` for documentation
    - `test:` for test additions
    - `refactor:` for code refactoring

6. **Push to Your Fork**
   ```bash
   git push origin feature/your-feature-name
   ```

7. **Open a Pull Request**
    - Provide a clear description of your changes
    - Reference any related issues
    - Ensure all tests pass

### Development Workflow

- Write tests for new code
- Ensure existing tests pass: `go test ./...`
- Follow Go best practices and idioms
- Keep commits atomic and focused
- Update README.md if adding user-facing features

### Reporting Issues

Found a bug? Please [open an issue](https://github.com/deannos/email-checker-tool/issues) with:
- Clear description of the problem
- Steps to reproduce
- Expected vs. actual behavior
- Go version and OS information

---

## Performance Benchmarks

The tool is designed for high-throughput domain validation. Performance varies based on configuration and network conditions.

### Typical Performance Metrics

| Configuration | Throughput | Notes |
|---|---|---|
| Default (10 workers, 20 RPS) | 200-500 domains/min | Conservative, safe defaults |
| Optimized (20 workers, 50 RPS) | 1,000-2,000 domains/min | Balanced performance |
| Aggressive (50 workers, 100 RPS) | 3,000-5,000 domains/min | Requires careful monitoring |

**Note:** Actual performance depends on DNS resolver response times, network latency, and domain configuration complexity.

---

## Troubleshooting

### High Timeout Rate

**Problem:** Many domains timing out or failing to resolve.

**Solutions:**
- Increase `--timeout` duration
- Reduce `--workers` to decrease concurrent load
- Lower `--rps` to reduce DNS query rate
- Verify network connectivity and DNS resolver availability

### IP Blocks from DNS Servers

**Problem:** DNS queries are being blocked (SERVFAIL responses).

**Solutions:**
- Significantly reduce `--rps` (try 10-15)
- Reduce `--workers` (5-10)
- Implement delays between batches
- Use a different DNS resolver or VPN

### Memory Usage Growing

**Problem:** Memory consumption increases over time.

**Solutions:**
- This is rare with the streaming CSV processor
- Reduce `--workers` to lower concurrent operations
- Process smaller batches of domains
- Check for resource leaks (report as issue)

### Empty or Incorrect Output

**Problem:** Output CSV is missing data or columns.

**Solutions:**
- Verify input CSV has `domain` column
- Check that input file is valid UTF-8
- Ensure output file path is writable
- Check console output for error messages

---

## FAQ

**Q: What DNS records does the tool validate?**
A: The tool checks for MX, SPF, and DMARC records. It validates record presence and captures full record values for manual inspection.

**Q: Can I use this in production?**
A: Yes. The tool is designed for production use with proper configuration and monitoring. Start with conservative settings and scale gradually.

**Q: Does the tool modify any DNS records?**
A: No. Email Checker Tool is read-only—it performs DNS queries only and never makes changes to any records.

**Q: What happens if a domain is invalid?**
A: Invalid domains produce an error entry in the output CSV. The tool continues processing other domains.

**Q: Can I run multiple instances in parallel?**
A: Yes, but be mindful of combined rate limits. A safer approach is increasing `--workers` in a single instance.

**Q: Is there a web API or GUI?**
A: Currently, the tool is CLI-only. A REST API or web interface could be explored in future versions.

---

## License

This project is distributed under the **MIT License**. See the [LICENSE](LICENSE) file for complete terms and conditions.

---

## Support and Community

- **Issues:** [Report bugs or request features](https://github.com/deannos/email-checker-tool/issues)
- **Discussions:** [Start a discussion](https://github.com/deannos/email-checker-tool/discussions)
- **Documentation:** See this README and inline code comments

---

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and release notes.

---

## Acknowledgments

Built with Go's robust standard library and inspired by best practices in the open-source community.

---

<div align="center">

Made with ❤️ by [deannos](https://github.com/deannos)

**Star this repo if you find it useful!** ⭐

</div>
