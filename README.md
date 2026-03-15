# Email Checker Tool

<div align="center">

[![Go Version](https://img.shields.io/badge/go-1.21%2B-blue.svg)](https://golang.org/doc/devel/release)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/deannos/email-checker-tool.svg)](https://github.com/deannos/email-checker-tool/releases)
[![CI](https://github.com/deannos/email-checker-tool/actions/workflows/ci.yml/badge.svg)](https://github.com/deannos/email-checker-tool/actions/workflows/ci.yml)

A high-performance, production-grade CLI tool for bulk email domain validation. Verify MX, SPF, and DMARC configurations at scale with concurrent processing, intelligent rate limiting, and flexible output formats.

[Features](#features) • [Installation](#installation) • [Usage](#usage) • [Configuration](#configuration) • [Architecture](#architecture) • [Contributing](#contributing)

</div>

---

## Overview

Email Checker Tool is an open-source CLI utility designed to validate email domain configurations at scale. Whether you're auditing mailing lists, checking DNS health, or enforcing email authentication policies, it provides a robust, efficient pipeline built on Go's standard library.

**v2.0.0 highlights:** custom DNS resolver, automatic retry with backoff, in-memory result caching, JSON output (JSONL), per-domain timing, error categorization, flexible CSV column selection, live progress stats, and a GitHub Actions CI/CD pipeline.

---

## Features

### High-Performance Processing
- **Concurrent Workers:** Configurable goroutine pool for parallel domain processing
- **Streaming I/O:** Processes CSV input of any size with a minimal memory footprint
- **In-Memory Caching:** Optional TTL-based cache deduplicates repeated domain lookups

### Intelligent Rate Limiting
- **Configurable RPS:** Token bucket algorithm enforces a global request-per-second limit across all workers
- **Server-Friendly:** Prevents IP blocks and DNS server overload out of the box

### Robust DNS Handling
- **Fully Context-Aware:** MX, SPF, and DMARC lookups all respect timeout and cancellation
- **Automatic Retry:** Exponential backoff retries on transient errors (configurable attempts)
- **Custom Resolver:** Point the tool at any DNS server (e.g. `8.8.8.8:53`, `1.1.1.1:53`)
- **Error Categorization:** Distinguishes `timeout`, `nxdomain`, `network`, and `unknown` failures

### Flexible Output
- **CSV (default):** Machine-readable, includes all record values plus timing and error type
- **JSON (JSONL):** One JSON object per line — pipe directly into `jq` or any log processor
- **Live Progress:** Logs `Processed / Errors / Rate req/s` every second during a run

### Safe Shutdown
- **Graceful Signal Handling:** `SIGINT`/`SIGTERM` flushes buffered results before exit
- **Data Integrity:** No results are lost on interrupted runs

---

## Installation

### Prerequisites

- Go 1.21 or higher
- Linux, macOS, or Windows

### From Source

```bash
git clone https://github.com/deannos/email-checker-tool.git
cd email-checker-tool
go build -o bin/email-checker ./cmd/email-checker
```

Or install directly:

```bash
go install github.com/deannos/email-checker-tool/cmd/email-checker@latest
```

### From Binary Release

Download precompiled binaries for your platform from the [Releases](https://github.com/deannos/email-checker-tool/releases) page:

- Linux (amd64, arm64)
- macOS (Intel, Apple Silicon)
- Windows (amd64)

```bash
# Linux/macOS
chmod +x email-checker-linux-amd64
sudo mv email-checker-linux-amd64 /usr/local/bin/email-checker
```

### Verify Installation

```bash
email-checker --version
# email-checker 2.0.0
```

---

## Quick Start

Create a CSV file with domains:

```csv
google.com
github.com
example.com
```

Run the checker:

```bash
email-checker domains.csv
```

Results are saved to `output.csv` by default. Progress is printed to stderr every second.

---

## Usage

### Basic

```bash
email-checker domains.csv
```

### Custom output path and format

```bash
# CSV output
email-checker --output results.csv domains.csv

# JSON (JSONL) output
email-checker --format json --output results.jsonl domains.csv
```

### Tune performance

```bash
email-checker --workers 20 --rps 50 --timeout 60s domains.csv
```

### Use a custom DNS resolver with retries

```bash
email-checker --dns 8.8.8.8:53 --retries 3 domains.csv
```

### Enable result caching (useful for lists with duplicates)

```bash
email-checker --cache-ttl 10m domains.csv
```

### Handle a CSV where domain is not the first column

```bash
# Domain is in column index 1 (second column), with a header row
email-checker --col 1 --skip-header contacts.csv
```

### Quiet mode (no progress output)

```bash
email-checker --no-progress domains.csv
```

### Pipe JSON results into jq

```bash
email-checker --format json --output /dev/stdout --no-progress domains.csv \
  | jq 'select(.hasMX == false)'
```

---

## Configuration

### CLI Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--workers` | `int` | `10` | Number of concurrent worker goroutines |
| `--rps` | `int` | `20` | Max DNS requests per second (global rate limit) |
| `--timeout` | `duration` | `30s` | Global operation timeout (e.g. `30s`, `5m`) |
| `--output` | `string` | `output.csv` | Output file path |
| `--format` | `string` | `csv` | Output format: `csv` or `json` |
| `--dns` | `string` | _(system)_ | Custom DNS resolver address (e.g. `8.8.8.8:53`) |
| `--retries` | `int` | `2` | Retry attempts on transient DNS errors |
| `--cache-ttl` | `duration` | `5m` | Result cache TTL; `0` disables caching |
| `--col` | `int` | `0` | Zero-based column index of the domain in the input CSV |
| `--skip-header` | `bool` | `false` | Skip the first row of the input CSV |
| `--no-progress` | `bool` | `false` | Suppress periodic progress stats |
| `--version` | `bool` | `false` | Print version and exit |

### Tuning Guidelines

**For throughput:**
- Increase `--workers` (20–50) and `--rps` (50–100) for large lists
- Increase `--timeout` (`60s` or more) when processing 10,000+ domains
- Enable `--cache-ttl` when the input contains many duplicate domains

**For safety (shared DNS infrastructure):**
- Keep `--workers` at 5–10 and `--rps` at 10–15
- Use `--retries 3` with `--dns 8.8.8.8:53` if the system resolver is unreliable

---

## Input and Output

### Input Format

The input must be a CSV file. By default, the domain is read from the first column (index `0`). Use `--col` to select a different column and `--skip-header` to ignore a header row.

```csv
google.com
github.com
example.com
```

Or with headers and multiple columns:

```csv
company,domain,contact
Google,google.com,support@google.com
GitHub,github.com,support@github.com
```

```bash
email-checker --col 1 --skip-header contacts.csv
```

### CSV Output

```csv
domain,hasMX,hasSPF,spfRecord,hasDMARC,dmarcRecord,error,errorType,durationMs
google.com,true,true,"v=spf1 include:_spf.google.com ~all",true,"v=DMARC1; p=none; ...",,,42
example.com,true,false,,false,,,,31
bad-domain.xyz,false,false,,false,,lookup failed,nxdomain,5
```

### CSV Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `domain` | string | Domain that was checked |
| `hasMX` | bool | Domain has at least one MX record |
| `hasSPF` | bool | Domain has an SPF TXT record |
| `spfRecord` | string | Full SPF record value |
| `hasDMARC` | bool | Domain has a DMARC policy |
| `dmarcRecord` | string | Full DMARC record value |
| `error` | string | Error message if the lookup failed |
| `errorType` | string | Categorized error: `timeout`, `nxdomain`, `network`, `unknown` |
| `durationMs` | int | Total lookup time in milliseconds |

### JSON Output (JSONL)

Each line is a self-contained JSON object:

```json
{"domain":"google.com","hasMX":true,"hasSPF":true,"spfRecord":"v=spf1 include:_spf.google.com ~all","hasDMARC":true,"dmarcRecord":"v=DMARC1; p=none;","durationMs":42}
{"domain":"bad-domain.xyz","hasMX":false,"hasSPF":false,"hasDMARC":false,"error":"lookup failed","errorType":"nxdomain","durationMs":5}
```

---

## Use Cases

### Email List Validation

```bash
email-checker --workers 30 --rps 50 --timeout 60s email_list.csv
```

### Security / Compliance Audit

Find domains missing SPF or DMARC:

```bash
email-checker --format json --output /dev/stdout --no-progress domains.csv \
  | jq 'select(.hasSPF == false or .hasDMARC == false) | .domain'
```

### DNS Infrastructure Testing with Custom Resolver

```bash
email-checker --dns 1.1.1.1:53 --retries 3 domains.csv
```

### Bulk Sender Validation

```bash
email-checker --rps 30 --output validated_senders.csv senders.csv
```

---

## Architecture

The project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout).

### Directory Structure

```
email-checker-tool/
├── .github/
│   └── workflows/
│       └── ci.yml                  # CI: test, lint, cross-platform release builds
├── cmd/
│   └── email-checker/
│       └── main.go                 # CLI entry point
├── internal/
│   ├── checker/
│   │   ├── checker.go              # Checker struct: DNS lookups, retry, error categorization
│   │   ├── checker_test.go
│   │   └── cache.go                # TTL-based in-memory result cache
│   ├── output/
│   │   ├── csv.go                  # Thread-safe CSV writer
│   │   ├── csv_test.go
│   │   └── json.go                 # JSONL writer
│   ├── version/
│   │   └── version.go
│   └── worker/
│       ├── pool.go                 # Worker pool with rate limiting and atomic stats
│       └── pool_test.go
├── go.mod
└── go.sum
```

### Core Components

**`internal/checker`** — DNS resolution layer. The `Checker` struct encapsulates a configurable `net.Resolver`, retry logic with exponential backoff, error categorization, and an optional TTL cache. All lookups (MX, SPF, DMARC) are fully context-aware.

**`internal/worker`** — Concurrent processing pipeline. `Pool` accepts a `CheckFn` function, a `Writer`, and a `rate.Limiter`. Worker goroutines pull domains from a buffered channel, respect the global rate limit, and record processed/error counts with atomic operations.

**`internal/output`** — Result serialization. `CSVWriter` uses a mutex-protected `encoding/csv` writer; `JSONWriter` uses `json.Encoder` for lock-protected JSONL output. Both implement the `worker.Writer` interface.

**`cmd/email-checker`** — Wires everything together: parses flags, constructs the `Checker` and writer, feeds a goroutine with CSV input, starts the pool, and logs live progress stats.

### Design Principles

- **Concurrency:** Goroutines with a bounded channel-based job queue
- **Rate Limiting:** Token bucket (`golang.org/x/time/rate`) enforced globally
- **Resilience:** Retry with exponential backoff; errors captured per-domain without halting
- **Extensibility:** `worker.Writer` interface makes it easy to add new output formats

---

## Development

### Building

```bash
go build -o bin/email-checker ./cmd/email-checker
```

### Testing

```bash
# All packages
go test ./...

# Verbose with race detector
go test -v -race ./...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Linting

```bash
go vet ./...
golangci-lint run ./...
```

### CI/CD

Every push and pull request to `main` runs the full test suite and linter via GitHub Actions (`.github/workflows/ci.yml`). Tagging a release (`v*`) triggers cross-platform binary builds for Linux, macOS, and Windows — automatically attached to the GitHub Release.

---

## Performance Benchmarks

| Configuration | Throughput | Notes |
|---|---|---|
| Default (10 workers, 20 RPS) | ~1,200 domains/min | Safe defaults, good for most use cases |
| Optimized (20 workers, 50 RPS) | ~3,000 domains/min | Balanced performance |
| Aggressive (50 workers, 100 RPS) | ~6,000 domains/min | Monitor for DNS rate limiting |

Actual throughput depends on DNS resolver latency, network conditions, and domain complexity.

---

## Troubleshooting

### High Timeout Rate

- Increase `--timeout` (try `60s` or `120s`)
- Use a faster public resolver: `--dns 8.8.8.8:53`
- Reduce `--workers` and `--rps` to lower concurrent DNS load
- Add `--retries 3` for flaky network conditions

### DNS Rate Limiting / SERVFAIL

- Lower `--rps` to `10`–`15`
- Reduce `--workers` to `5`
- Switch to a different resolver: `--dns 1.1.1.1:53`

### Domain Column Not Detected

- Use `--col <index>` to specify the zero-based column index
- Use `--skip-header` if the CSV has a header row

### Empty Output File

- Verify the input path is correct and the file is readable
- Check that the output path is writable
- Look for error messages in the console log

---

## FAQ

**Q: What DNS records does the tool validate?**
A: MX (mail exchange), SPF (Sender Policy Framework via TXT), and DMARC (`_dmarc.<domain>` TXT). Record values are captured in full.

**Q: Does the tool modify any DNS records?**
A: No. All operations are read-only DNS queries.

**Q: Can I use a private/internal DNS resolver?**
A: Yes. Pass `--dns <host>:<port>` to point at any UDP DNS resolver, including internal ones.

**Q: What does the cache do?**
A: If the same domain appears more than once in the input, the second lookup is served from memory instead of hitting DNS again. Set `--cache-ttl 0` to disable.

**Q: Can I run multiple instances in parallel?**
A: Yes, but be mindful of combined RPS. It's usually safer to increase `--workers` in a single instance.

**Q: Is there a web API or GUI?**
A: The tool is CLI-only. Pipe JSON output to any downstream system as needed.

---

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for the full workflow.

Quick steps:

1. Fork the repo and create a feature branch
2. Make your changes with tests
3. Ensure `go test ./...` and `golangci-lint run` pass
4. Open a pull request — CI will run automatically

Report bugs or request features via [GitHub Issues](https://github.com/deannos/email-checker-tool/issues).

---

## License

Distributed under the **MIT License**. See [LICENSE](LICENSE) for details.

---

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history.

---

<div align="center">

Made with ❤️ by [deannos](https://github.com/deannos)

**Star this repo if you find it useful!** ⭐

</div>
