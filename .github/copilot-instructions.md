# Copilot Instructions for AI Coding Agents

## Project Overview
This repository implements the GitHub Models CLI extension (`gh models`), enabling users to interact with AI models via the `gh` CLI. The extension supports inference, prompt evaluation, model listing, and test generation using the PromptPex methodology. Built in Go using Cobra CLI framework and Azure Models API.

## Architecture & Key Components

### Building and Testing

- `make build`: Compiles the CLI binary
- `make check`: Runs format, vet, tidy, tests, golang-ci. Always run when you are done with changes. Use this command to validate that the build and the tests are still ok.
- `make test`: Runs the tests.

### Command Structure
- **cmd/root.go**: Entry point that initializes all subcommands and handles GitHub authentication
- **cmd/{command}/**: Each subcommand (generate, eval, list, run, view) is self-contained with its own types and tests
- **pkg/command/config.go**: Shared configuration pattern - all commands accept a `*command.Config` with terminal, client, and output settings

### Core Services
- **internal/azuremodels/**: Azure API client with streaming support via SSE. Key pattern: commands use `azuremodels.Client` interface, not concrete types
- **pkg/prompt/**: `.prompt.yml` file parsing with template substitution using `{{variable}}` syntax
- **internal/sse/**: Server-sent events for streaming responses

### Data Flow
1. Commands parse `.prompt.yml` files via `prompt.LoadFromFile()`
2. Templates are resolved using `prompt.TemplateString()` with `testData` variables  
3. Azure client converts to `azuremodels.ChatCompletionOptions` and makes API calls
4. Results are formatted using terminal-aware table printers from `command.Config`

### Generate Command (PromptPex)
- **Location**: `cmd/generate/` - self-contained implementation of test generation
- **Key files**: `pipeline.go` (orchestration), `prompts.go` (phase prompts), `evaluators.go` (test evaluation)
- **Flow**: Intent → InputSpec → OutputRules → InverseOutputRules → Tests → Evaluation
- **Session state**: `types.go` defines `PromptPexContext` - serialized to JSON for resumability
- **Cleanup**: `cleaner.go` handles post-generation test refinement and deduplication

## Developer Workflows

### Building & Testing
- **Local build**: `make build` or `script/build` (creates `gh-models` binary)
- **Cross-platform**: `script/build all|windows|linux|darwin` for release builds
  - Builds for windows/amd64, linux/amd64, android/arm64, android/amd64, darwin/amd64, darwin/arm64
- **Testing**: `make check` runs format, vet, tidy, and tests. Use `go test ./...` directly for faster iteration
- **Quality gates**: `make check` - required before commits
- **Integration tests**: `make integration` - builds binary and runs tests against live endpoints (requires `gh auth login`)
  - Located in `integration/` directory with separate go.mod
  - Tests gracefully skip when authentication unavailable
  - CI runs these via `.github/workflows/integration.yml`

### Authentication & Setup
- Extension requires `gh auth login` before use - unauthenticated clients show helpful error messages
- Client initialization pattern in `cmd/root.go`: check token, create appropriate client (authenticated vs unauthenticated)

### Release Process
- Create git tag: `git tag v0.0.x main && git push origin tag v0.0.x`
- This triggers `.github/workflows/release.yml` for production builds
- Users install with `gh extension install github/gh-models` (pulls latest release, not latest commit)

## Prompt File Conventions

### Structure (.prompt.yml)
```yaml
name: "Test Name"
model: "openai/gpt-4o-mini" 
messages:
  - role: system|user|assistant
    content: "{{variable}} templating supported"
testData:
  - variable: "value1"
  - variable: "value2"
evaluators:
  - name: "test-name"
    string: {contains: "{{expected}}"} # String matching
    # OR
    llm: {modelId: "...", prompt: "...", choices: [{choice: "good", score: 1.0}]}
```

### Response Formats
- **JSON Schema**: Use `responseFormat: json_schema` with `jsonSchema` field containing strict JSON schema
- **Templates**: All message content supports `{{variable}}` substitution from `testData` entries

### PromptPex Test Generation
- **Command**: `gh models generate` - implements [PromptPex](https://github.com/microsoft/promptpex) methodology
- **Effort levels**: `min` (quick validation), `low` (limits to 3 rules), `medium` (better coverage), `high` (complex inputs, more tokens)
- **Custom instructions**: Use `--instruction-{phase}` flags for intent, inputspec, outputrules, inverseoutputrules, tests
- **Session persistence**: `--session-file` to save/load generation state
- **Documentation**: See `cmd/generate/README.md` and https://microsoft.github.io/promptpex/reference/test-generation/

## Testing Patterns

### Unit Tests (Command Tests)
- **Location**: `cmd/{command}/{command}_test.go` 
- **Pattern**: Create mock client via `azuremodels.NewMockClient()`, inject into `command.Config`
- **Structure**: Table-driven tests with subtests using `t.Run()`
- **Assertions**: Use `testify/require` for cleaner error messages
- **Run**: `make test` or `go test -race -cover ./...`

### Mock Usage
```go
client := azuremodels.NewMockClient()
cfg := command.NewConfig(new(bytes.Buffer), new(bytes.Buffer), client, true, 80)
```

### Integration Tests
- **Location**: `integration/integration_test.go` (separate go.mod)
- **Pattern**: Execute compiled binary via `exec.Command()` with timeout protection
- **Prerequisites**: Binary must exist (`make build` first)
- **Authentication**: Tests skip gracefully when `gh auth` unavailable
- **Focus**: Verify basic functionality, command execution, output format - not full feature testing
- **Run**: `make integration` (combines `make check`, `make build`, and integration tests)

## Integration Points

### GitHub Authentication
- Uses `github.com/cli/go-gh/v2/pkg/auth` for token management
- Pattern: `auth.TokenForHost("github.com")` to get tokens

### Azure Models API
- Streaming via SSE with custom `sse.EventReader`
- Rate limiting handled automatically by client
- Content safety filtering always enabled (cannot be disabled)

### Terminal Handling  
- All output uses `command.Config` terminal-aware writers
- Table formatting via `cfg.NewTablePrinter()` with width detection

---

**Key Files**: `cmd/root.go` (command registration), `pkg/prompt/prompt.go` (file parsing), `internal/azuremodels/azure_client.go` (API integration), `examples/` (prompt file patterns)

## Instructions

Omit the final summary.