# Email Checker Tool

A high-performance, open-source CLI tool to verify DNS configurations (MX, SPF, DMARC) for domain lists. Built in Go, it utilizes concurrent workers and smart rate limiting to process thousands of domains efficiently without overwhelming DNS servers.

## Features

-   **High Performance:** Concurrent processing with configurable worker pools.
-   **Rate Limiting:** Built-in request throttling (RPS) to prevent IP blocks.
-   **Timeout Handling:** Context-aware timeouts prevent hanging on unresponsive servers.
-   **CSV Streaming:** Processes large files efficiently with low memory footprint.
-   **Graceful Shutdown:** Handles SIGINT (Ctrl+C) safely, saving progress before exit.
-   **Comprehensive Checks:** Validates MX records, SPF records, and DMARC policies.

## Installation

### From Source

Ensure you have Go installed (version 1.21 or higher).

```bash
go install github.com/deannos/email-checker-tool@latest
```

From Binary (Linux/macOS/Windows)

Download the latest release for your platform from the Releases  page.

Usage

Basic Usage

Provide a CSV file containing a list of domains (one domain per row, or in the first column).

```bash
email-checker domains.csv
```

Advanced Options 

```bash
email-checker input.csv \
  --workers 20 \
  --rps 50 \
  --timeout 10s \
  --output results.csv
```

Flags 

| Flag | Description | Default |
| :--- | :--- | :--- |
| --workers | Number of concurrent workers to process domains. | 10 |
| --rps | Max DNS requests per second (Rate Limit). | 20 |
| --timeout | Global timeout for the entire operation. | 5s |
| --output | Path to the output CSV file. | output.csv |
| --version | Print the version number. | false |


Input Format

The tool expects a CSV file where the first column contains the domain.

Example input.csv:

```bash
domain
google.com
github.com
example.com
```

Output Format

The tool generates a CSV file with the analysis results.

Example output.csv:

```bash
domain,hasMX,hasSPF,spfRecord,hasDMARC,dmarcRecord,error
google.com,true,true,"v=spf1 ...",true,"v=DMARC1 ...",
github.com,true,true,"v=spf1 ...",true,"v=DMARC1 ...",
example.com,true,false,"",false,"",
```

## Development

Running Tests

Run the full test suite:

```bash
go test ./...
```

Building from Source

```bash
go build -o bin/email-checker ./cmd/email-checker
```

## Architecture

This project follows the Standard Go Project Layout :

- **cmd/email-checker**: Application entry point.
- **internal/checker**: Core logic for DNS lookups (MX, SPF, DMARC).
- **internal/worker**: Worker pool implementation with rate limiting.
- **internal/output**: CSV writing logic.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

Fork the repository.
1. Create your feature <branch_name> 
```bash
git checkout -b feature/<feature_name>
```

2. Commit your changes 
```bash 
git commit -m 'feat: <feature details>'
```

3. Push to the branch 
```bash
git push origin feature/<feature_name>
```
4. Open a Pull Request.

## License

Distributed under the MIT License. See LICENSE for more information.