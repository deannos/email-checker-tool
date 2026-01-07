# Contributing to Email Checker Tool

First off, thanks for taking the time to contribute!

This document provides guidelines and instructions for contributing to the Email Checker Tool project.

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating a bug report, check the issue list as you might find out that you don't need to create one. When you are creating a bug report, please include as many details as possible:

* **Use a clear and descriptive title**
* **Describe the exact steps which reproduce the problem**
* **Provide specific examples to demonstrate the steps**
* **Describe the behavior you observed after following the steps**
* **Explain which behavior you expected to see instead and why**
* **Include screenshots and animated GIFs if possible**
* **Include your environment:**
    - Go version (`go version`)
    - Operating system and version
    - Command and flags used
    - Input CSV sample (if applicable)

### Suggesting Enhancements

When creating an enhancement suggestion, please include:

* **Use a clear and descriptive title**
* **Provide a step-by-step description of the suggested enhancement**
* **Provide specific examples to demonstrate the steps**
* **Describe the current behavior and the expected behavior**
* **Explain why this enhancement would be useful**

### Pull Requests

* Fill in the required template
* Follow the Go styleguides
* Include appropriate test cases
* Update documentation as needed
* End all files with a newline

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git

### Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/email-checker-tool.git
   cd email-checker-tool
   ```
3. Add upstream remote
   ```bash
   git remote add upstream https://github.com/deannos/email-checker-tool.git
   ```
4. Create a feature branch
   ```bash
   git checkout -b feature/your-feature-name
   ```
### Building 

```bash
go build -o bin/email-checker ./cmd/email-checker
```
### Running Tests 

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...
```

### Code Style 

- Use **gofmt** to format code: **go fmt ./...**
- Use **go vet** to check for errors: **go vet ./...**
- Use **golangci-lint** for linting (if available)
- Follow Go naming conventions and idioms
- Write clear, descriptive variable and function names
- Add comments for exported functions and packages

### Commit Messages

Use conventional commit format:

```text
feat: add feature description
fix: fix bug description
docs: documentation changes
test: test additions
refactor: code refactoring
chore: maintenance tasks
```

For Example: 

```text
feat: add DNS timeout configuration

- Add configurable DNS timeout flag
- Update worker pool to use new timeout
- Add tests for timeout behavior

```

### Testing Requirements 

- All new code must have accompanying tests
- Maintain or improve code coverage
- All tests must pass before submission
- Run tests locally before pushing:

```bash
go test ./...
```

### Pull Request Process 

1. Update your branch with the latest main:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```
2. Push your branch:
   ```bash
   git push origin feature/your-feature-name
   ```
3. Open a Pull Request on GitHub
4. provide a clear description of your changes 
5. Reference any related issues 
6. Ensure all checks pass

### Review Process

- At least one maintainer review is required
- Address any requested changes promptly
- The PR will be merged after approval

### Recognition 

Contributors will be recognized in:

- CHANGELOG.md for significant contributions
- GitHub contributors page 
- Release notes

### Questions 

Feel free to open a discussion or issue with the tag **question**