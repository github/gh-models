# Copilot Instructions for AI Coding Agents

## Project Overview
This repository implements the GitHub Models CLI extension, enabling users to interact with various AI models via the `gh` CLI. The codebase is organized for extensibility, supporting prompt evaluation, model listing, and inference workflows. It uses Go.

## Architecture & Key Components
- **cmd/**: Main CLI commands. Subfolders (e.g., `generate/`, `eval/`, `list/`, `run/`, `view/`) encapsulate distinct features.
- **internal/**: Contains integrations (e.g., Azure model clients) and shared logic (e.g., SSE, model keys).
- **pkg/**: Utility packages for config, prompt parsing, and general helpers.
- **examples/**: Sample prompt files and GitHub Actions for reference and testing.
- **script/**: Build and release scripts.

## Developer Workflows
- **Build**: Use the provided `Makefile` or scripts in `script/` for building and packaging. Example: `make build` or `bash script/build`.
- **Test**: Run Go tests with `go test ./...`. Individual command tests are in `cmd/*/*_test.go`.
- **Debug**: Logging is handled via the standard library (`log`). Most command structs accept a logger for debugging output.
- **CLI Usage**: The extension is invoked via `gh models <command>`. See `README.md` for usage patterns and examples.

## External Dependencies & Integration
- **gh CLI**: Required for extension operation. Authenticate via `gh auth`.
- **Azure AI Content Safety**: Integrated for output filtering; cannot be disabled.
- **OpenAI API**: Used for model inference and evaluation (see `openai.ChatCompletionRequest`).

## Conventions & Recommendations
- Keep new features modular by adding new subfolders under `cmd/`.
- Use the provided types and utility functions for consistency.
- Persist results and context to output directories for reproducibility.
- Reference `README.md` and `examples/` for usage and integration patterns.
- Follow Go best practices for naming.

## Generating Test Files
- **Test File Location**: For each CLI command, place its tests in the same subfolder, named as `<command>_test.go` (e.g., `cmd/generate/generate_test.go`).
- **Test Structure**: Use Go's standard `testing` package. Each test should cover a distinct scenario, including edge cases and error handling. 
- **Manual Tests**: For manual unit tests, follow the pattern in existing test files. Use table-driven tests for coverage and clarity.
- **Running Tests**: Execute all tests with `go test ./...` or run specific files with `go test cmd/generate/generate_test.go`.
- **Examples**: See `cmd/generate/generate_test.go` and `examples/` for sample test prompts and expected outputs.

---

For questions or unclear patterns, review the `README.md` and key files in `cmd/generate/`, or ask for clarification.
