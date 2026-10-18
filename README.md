# Rename AI (rnai)

`rnai` is a CLI tool that uses Google's Gemini AI to analyze file content and rename files with descriptive, semantic names.

## Features

- **AI-Powered Renaming**: Analyzes text and PDF files to generate names based on content.
- **Safety First**: Includes a `--dry-run` mode to preview changes.
- **Collision Handling**: Automatically handles duplicate filenames by incrementing a counter.

## Installation

### Option 1: Binary (Recommended)
Download the pre-compiled binary for your system from the [Releases](https://github.com/maltehedderich/rename-ai/releases) page.
*No additional dependencies required.*

### Option 2: Go Install
If you are a Go developer or prefer installing via `go`:
```bash
go install github.com/maltehedderich/rename-ai/cmd/rnai@latest
```

Go installs `rnai` into `GOBIN`, or `$(go env GOPATH)/bin` when `GOBIN` is unset (usually `~/go/bin`). This directory must be on your shell's `PATH`.

If you get `command not found: rnai` on macOS or Linux, add the install directory to your current shell:

```bash
rnai_bin="$(go env GOBIN)"
export PATH="${rnai_bin:-$(go env GOPATH)/bin}:$PATH"
rnai --help
```

To make this permanent, add the first two lines to `~/.zshrc` (zsh) or `~/.bashrc` (bash), then open a new terminal.

## Configuration

Before running the tool, you must configure your Gemini API key.

```bash
export GEMINI_API_KEY="your-api-key-here"
```

Optional configuration:
```bash
# Set a specific model (default: gemini-flash-latest)
export GEMINI_MODEL="gemini-3-pro-preview"
```

## Usage

Run the tool in a folder to choose files interactively:

```bash
rnai
```

The picker shows visible regular files in the current folder. Use `a` to select all, `n` to clear the selection, numbers such as `1` or `1,3`, ranges such as `1-4`, Enter to continue, and `q` to cancel.

Run the tool on one supported file:

```bash
rnai path/to/file.pdf
```

Run the tool on multiple explicit files:

```bash
rnai invoice.pdf notes.md screenshot.png
```

### Flags

-   `--model`: Specify the Gemini model to use (overrides `GEMINI_MODEL`).
    ```bash
    rnai document.pdf --model gemini-3-pro-preview
    ```
-   `--dry-run`: Simulate the rename operation without making changes.
    ```bash
    rnai document.pdf --dry-run
    rnai invoice.pdf notes.md --dry-run
    ```

## Example

```bash
$ rnai scan_001.pdf

> Using model: gemini-flash-latest
> Detected type: application/pdf
> Analyzing 'scan_001.pdf' with Gemini...

Proposal
Reasoning:
  The document summarizes the Q1 financial results for the fiscal year 2024.

Rename:
  scan_001.pdf -> 2024-03-24_q1-financial-report-2024.pdf

Rename? [y/N]: y
> Success! Renamed to 2024-03-24_q1-financial-report-2024.pdf
```

---

## Development

If you want to contribute to `rnai`, you will need:

- **Go 1.25+**
- **Google Gemini API Key** for running integration tests (if any).

### Running Tests

To run the test suite:

```bash
go test -v ./...
```
