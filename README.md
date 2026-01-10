# Rename AI (rnai)

`rnai` is a CLI tool that uses Google's Gemini AI to analyze file content and rename files with descriptive, semantic names.

## Features

- **AI-Powered Renaming**: Analyzes text and PDF files to generate names based on content.
- **Safety First**: Includes a `--dry-run` mode to preview changes.
- **Collision Handling**: Automatically handles duplicate filenames by incrementing a counter.

## Prerequisites

- Go 1.22+
- Google Gemini API Key

## Installation

```bash
go install github.com/maltehedderich/rename-ai/cmd/rnai@latest
```

## Usage

1.  **Configure Environment**:
    ```bash
    export GEMINI_API_KEY="your-api-key-here"
    # Optional: Set a specific model (default: gemini-flash-latest)
    export GEMINI_MODEL="gemini-3-pro-preview"
    ```

2.  **Run the tool**:
    ```bash
    rnai path/to/file.pdf
    ```

### Flags

-   `--model`: Specify the Gemini model to use (overrides `GEMINI_MODEL`).
    ```bash
    rnai document.pdf --model gemini-3-pro-preview
    ```
-   `--dry-run`: Simulate the rename operation without making changes.
    ```bash
    rnai document.pdf --dry-run
    ```

## Example

```bash
$ rnai scan_001.pdf

> Detected type: application/pdf
> Analyzing 'scan_001.pdf' with Gemini...
> Proposed Name: q1-financial-report-2024.pdf
> Reasoning: The document summarizes the Q1 financial results for the fiscal year 2024.
> Rename? [y/N]: y
> Success! Renamed to q1-financial-report-2024.pdf
```

## Development

### Running Tests

To run the test suite, use the standard Go test command:

```bash
go test -v ./...
```

