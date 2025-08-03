## Contributing

[fork]: https://github.com/github/REPO/fork
[pr]: https://github.com/github/REPO/compare
[style]: https://github.com/github/REPO/blob/main/.golangci.yaml

Hi there! We're thrilled that you'd like to contribute to this project. Your help is essential for keeping it great.

Contributions to this project are [released](https://help.github.com/articles/github-terms-of-service/#6-contributions-under-repository-license) to the public under the [project's open source license](LICENSE.txt).

Please note that this project is released with a [Contributor Code of Conduct](CODE_OF_CONDUCT.md). By participating in this project you agree to abide by its terms.

## Prerequisites for running and testing code

These are one time installations required to be able to test your changes locally as part of the pull request (PR) submission process.

1. Install Go [through download](https://go.dev/doc/install) | [through Homebrew](https://formulae.brew.sh/formula/go) and ensure it's at least version 1.22

## Submitting a pull request

1. [Fork][fork] and clone the repository
1. Make sure the tests pass on your machine: `go test -v ./...` _or_ `make test`
1. Create a new branch: `git checkout -b my-branch-name`
1. Make your change, add tests, and make sure the tests and linter still pass: `make check`
1. For integration testing: `make integration-test` (requires building the binary first)
1. Push to your fork and [submit a pull request][pr]
1. Pat yourself on the back and wait for your pull request to be reviewed and merged.

Here are a few things you can do that will increase the likelihood of your pull request being accepted:

- Follow the [style guide][style].
- Write tests.
- Keep your change as focused as possible. If there are multiple changes you would like to make that are not dependent upon each other, consider submitting them as separate pull requests.
- Write a [good commit message](http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html).

## Resources

- [How to Contribute to Open Source](https://opensource.guide/how-to-contribute/)
- [Using Pull Requests](https://help.github.com/articles/about-pull-requests/)
- [GitHub Help](https://help.github.com)

## Integration Testing

This project includes integration tests that run against the compiled `gh-models` binary and live LLM endpoints.

These tests are excluded from regular test runs and must be run explicitly using:
```bash
make integration-test
```

### Authentication

Some tests require GitHub authentication. Run `gh auth login` before running integration tests to test authenticated scenarios.

Tests are designed to handle both authenticated and unauthenticated scenarios gracefully:

- **Unauthenticated**: Tests validate proper error handling, exit codes, and help functionality
- **Authenticated**: Tests validate actual API interactions, file modifications, and live endpoint behavior

### Test Coverage

The integration test suite covers:

1. **Basic Commands**: Help functionality, error handling, exit codes
2. **File Operations**: Prompt file parsing, validation, modification tracking
3. **Authentication Scenarios**: Both authenticated and unauthenticated flows
4. **Command Chaining**: Sequential execution of multiple commands
5. **Output Formats**: JSON and default output format validation
6. **File System Interaction**: Working directory independence, file permissions
7. **Long-running Commands**: Timeout handling and performance validation

### Running Specific Tests

```bash
# Run all integration tests
make integration-test

# Run specific test patterns
go test -tags=integration -v ./integration/... -run TestBasicCommands

# Run in short mode (skips long-running tests)
go test -tags=integration -short -v ./integration/...
```
