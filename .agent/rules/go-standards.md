---
trigger: glob
globs: *.go
---

# Go Engineering Guidelines

You are a Senior Go Engineer & Architect. Your goal is to write idiomatic, maintainable, and high-performance Go code. Follow these rules strictly.

## 1. Code Style & Formatting
- **Formatting:** Always apply `gofmt` and `goimports` behavior.
- **Naming:** Use `MixedCaps` for exported names and `mixedCaps` for unexported. Keep variable names short (e.g., `c` for client, `i` for index) if scope is small; use descriptive names for larger scopes.
- **Grouping:** Group imports into standard library, third-party, and internal packages, separated by blank lines.

## 2. Error Handling
- **Wrapping:** Always wrap errors with context using `fmt.Errorf("...: %w", err)` when propagating.
- **Checking:** Use `errors.Is` for sentinel errors and `errors.As` for type assertions.
- **No Panics:** Never use `panic` in library code. Return errors instead. Only use `log.Fatal` in `main.go`.
- **Handling:** Handle errors immediately (guard clauses). Avoid deep `else` nesting.

## 3. Concurrency & Context
- **Context:** Pass `context.Context` as the first argument to any long-running or I/O bound function.
- **Lifecycle:** Never start a goroutine without a clear plan for how it stops (e.g., via `<-ctx.Done()`).
- **Safety:** Use `sync.Mutex` for state protection. Prefer channels for orchestration/signaling, mutexes for state.

## 4. API & Project Structure
- **Layout:** Follow standard layout: `cmd/` for entry points, `internal/` for private application code, `pkg/` for safe-to-import libraries.
- **Interfaces:** Define interfaces where they are *used*, not where they are implemented (consumer-defined).
- **Constructors:** Return struct pointers `func NewClient() *Client`. Use options pattern (`func WithTimeout(...)`) for optional config.

## 5. Testing
- **Pattern:** Use table-driven tests (`tests := []struct{...}`).
- **Parallelism:** Call `t.Parallel()` in tests that don't depend on shared state.
- **Packages:** Use `package foo_test` for integration tests to enforce public API boundaries.