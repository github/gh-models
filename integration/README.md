# Integration Tests

This directory contains integration tests for the `gh-models` CLI extension. These tests are separate from the unit tests and use the compiled binary to test actual functionality.

## Overview

The integration tests:
- Use the compiled `gh-models` binary (not mocked clients)
- Test basic functionality of each command (`list`, `run`, `view`, `eval`)
- Are designed to work with or without GitHub authentication
- Skip tests requiring live endpoints when authentication is unavailable
- Keep assertions minimal to avoid brittleness

## Running the Tests

### Prerequisites

1. Build the `gh-models` binary:
   ```bash
   cd ..
   script/build
   ```

2. (Optional) Authenticate with GitHub CLI for full testing:
   ```bash
   gh auth login
   ```

### Running Locally

From the integration directory:
```bash
go test -v
```

Without authentication, some tests will be skipped:
```
=== RUN   TestIntegrationHelp
--- PASS: TestIntegrationHelp (0.05s)
=== RUN   TestIntegrationList
    integration_test.go:90: Skipping integration test - no GitHub authentication available
--- SKIP: TestIntegrationList (0.04s)
```

With authentication, all tests should run and test live endpoints.

## CI/CD

The integration tests run automatically on pushes to `main` via the GitHub Actions workflow `.github/workflows/integration.yml`.

The workflow:
1. Builds the binary
2. Runs tests without authentication (tests basic functionality)
3. On manual dispatch, can also run with authentication for full testing

## Test Structure

Each test follows this pattern:
- Check for binary existence (skip if not built)
- Check for authentication (skip live endpoint tests if unavailable)
- Execute the binary with specific arguments
- Verify basic output format and success/failure

Tests are intentionally simple and focus on:
- Commands execute without errors
- Help text is present and correctly formatted
- Basic output format is as expected
- Authentication requirements are respected

## Adding New Tests

When adding new commands or features:
1. Add a corresponding integration test
2. Follow the existing pattern of checking authentication
3. Keep assertions minimal but meaningful
4. Ensure tests work both with and without authentication