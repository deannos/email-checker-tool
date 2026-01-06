# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability, please do **NOT** open a public issue. Instead, please email the maintainers privately:

* Contact: **amjha21122002@gmail.com**
* Subject: [SECURITY] Email Checker Tool - Vulnerability Report

Please include:
* Description of the vulnerability
* Steps to reproduce
* Potential impact
* Suggested fix (if applicable)

We will acknowledge your report within 48 hours and provide an estimated timeline for a fix.

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | ✅ Yes             |
| < 1.0   | ❌ No              |

## Security Considerations

### DNS Queries
- This tool performs DNS queries without caching
- Each domain results in multiple DNS lookups
- Be mindful of DNS rate limiting

### CSV Files
- Input CSV files are processed as-is without sanitization
- Ensure input files come from trusted sources
- No SQL injection risk (not database related)
- No code execution from CSV input

### Resource Usage
- Configurable workers prevent resource exhaustion on local system
- Rate limiting prevents overwhelming remote DNS servers
- Monitor memory usage with large domain lists

EOF
