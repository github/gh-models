# Integration Tests

This directory contains integration tests that run against the compiled `gh-models` binary and live LLM endpoints.

These tests are excluded from regular test runs and must be run explicitly using:
```bash
make integration-test
```

## Authentication

Some tests require GitHub authentication. Run `gh auth login` before running integration tests to test authenticated scenarios.

Tests are designed to handle both authenticated and unauthenticated scenarios gracefully:

- **Unauthenticated**: Tests validate proper error handling, exit codes, and help functionality
- **Authenticated**: Tests validate actual API interactions, file modifications, and live endpoint behavior

## Test Coverage

The integration test suite covers:

1. **Basic Commands**: Help functionality, error handling, exit codes
2. **File Operations**: Prompt file parsing, validation, modification tracking
3. **Authentication Scenarios**: Both authenticated and unauthenticated flows
4. **Command Chaining**: Sequential execution of multiple commands
5. **Output Formats**: JSON and default output format validation
6. **File System Interaction**: Working directory independence, file permissions
7. **Long-running Commands**: Timeout handling and performance validation

## Running Specific Tests

```bash
# Run all integration tests
make integration-test

# Run specific test patterns
go test -tags=integration -v ./integration/... -run TestBasicCommands

# Run in short mode (skips long-running tests)
go test -tags=integration -short -v ./integration/...
```