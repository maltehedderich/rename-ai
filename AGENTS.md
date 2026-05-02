# Repository Guidelines

## Project Structure & Module Organization

This is a Go CLI module for `rnai`, a Gemini-powered file renamer.

- `cmd/rnai/main.go` contains the Cobra CLI entry point, flags, environment binding, and top-level execution flow.
- `internal/domain/` contains pure filename, validation, and collision logic.
- `internal/adapters/` contains external boundary code: `ai/` for Gemini, `fs/` for filesystem access, and `ui/` for console interaction.
- `docs/` holds design notes, and `RELEASING.md` documents the GoReleaser tag-based release process.
- Tests live next to implementation files as `*_test.go`.

## Build, Test, and Development Commands

- `go test -v ./...` runs the full test suite.
- `go test -v -race ./...` runs tests with the race detector, matching CI.
- `go vet ./...` runs Go static analysis used by CI.
- `go build -v ./...` verifies all packages build.
- `go mod verify` checks downloaded module integrity.
- `go run ./cmd/rnai path/to/file.pdf --dry-run` runs the CLI locally. Set `GEMINI_API_KEY` first; optionally set `GEMINI_MODEL`.

CI also runs `golangci-lint` through GitHub Actions. Run `golangci-lint run ./...` locally if you have it installed.

## Coding Style & Naming Conventions

Use idiomatic Go. Format all Go files with `gofmt`; use `goimports` behavior for import grouping: standard library, third-party, then internal packages.

Use `MixedCaps` for exported names and `mixedCaps` for unexported names. Keep small-scope variables short, but use descriptive names for wider scopes. Wrap propagated errors with context using `fmt.Errorf("...: %w", err)`, and prefer guard clauses over deep nesting.

## Testing Guidelines

Write focused unit tests beside the package they cover, named `TestXxx` in `*_test.go` files. Prefer table-driven tests for domain logic and adapter behavior. Use `t.Parallel()` when tests do not share mutable state, files, environment variables, or console input.

Keep AI-dependent behavior isolated behind adapters so tests can exercise domain and UI logic without requiring live Gemini calls.

## Commit & Pull Request Guidelines

Recent history uses conventional prefixes such as `feat:`, `fix:`, `docs:`, and `chore:`. Keep commits small and imperative, for example `fix: handle uppercase confirmations`.

Pull requests should include a concise summary, linked issue when relevant, test results (`go test -v -race ./...`, `go vet ./...`), and screenshots or terminal output for visible CLI behavior changes.

## Security & Configuration Tips

Never commit API keys or generated credentials. Configure local runs with `GEMINI_API_KEY` and optional `GEMINI_MODEL` environment variables. Prefer `--dry-run` when testing rename behavior on real files.
