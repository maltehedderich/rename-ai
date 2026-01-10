# Design Specification: AI Content Renamer (MVP)
**Version:** 0.2.0 (Refined MVP)
**Date:** 2024-05-20
**Author:** Principal Go Engineer

## 1. Executive Summary
The content-renamer is a CLI tool written in Go that reads a specific file (Text or PDF), analyzes its content using Google Gemini, and suggests a semantic filename based on a summary of the content. The tool ensures file safety via dry-runs and auto-incrementing naming collisions.

## 2. MVP Scope & Constraints
- **Input:** Single file. Supported types: Text-based files (.txt, .md, source code) and PDFs.
- **AI Provider:** Google Gemini (leveraging native document parsing).
- **Naming Strategy:** Names generated based on a summary of the file's content.
- **Collision Handling:** Auto-increment (e.g., file.pdf -> file-1.pdf).
- **Safety:** Dry-run by default. Explicit confirmation required before filesystem write.
- **Limit:** Respect Gemini's token limits; for large PDFs, the tool will utilize Gemini's File API or Blob input features.

## 3. High-Level Architecture (Hexagonal)
We will use a Ports and Adapters (Hexagonal) architecture to separate the core logic from the Gemini API and the filesystem.

### 3.1 Domain (Core Logic)
- **Entities:** `FileMetadata` (includes MIME type), `RenameSuggestion`.
- **Business Logic:**
    - MIME type detection.
    - Filename sanitization (kebab-case enforcement).
    - Collision resolution logic (Increment counter).

### 3.2 Ports (Interfaces)
These interfaces define how the application interacts with the outside world.

```go
package ports

import "context"

// FileSystem handles OS-level operations
type FileSystem interface {
    ReadFile(path string) ([]byte, error)
    Rename(oldPath, newPath string) error
    Exists(path string) bool
    GetMimeType(path string) (string, error)
}

// AIProvider handles interaction with the LLM
type AIProvider interface {
    // GenerateName accepts raw content and mimeType to support Multimodal inputs (PDFs)
    GenerateName(ctx context.Context, content []byte, mimeType string, currentExt string) (string, error)
}

// UI handles interaction with the user
type UI interface {
    PrintProposal(old, new, reasoning string)
    Confirm(question string) (bool, error)
    Error(msg string)
}
```

### 3.3 Adapters (Implementations)
- **CLI Adapter:** `spf13/cobra`.
- **OS Adapter:** `os` lib + `github.com/gabriel-vasile/mimetype` for detection.
- **Gemini Adapter:** Uses `github.com/google/generative-ai-go`. For PDFs, it sends the data as a `genai.Blob` or uses the File API for larger documents.

## 4. Detailed Design

### 4.1 CLI Interface (UX)
```bash
# Basic usage
$ renamer ./scan_001.pdf

# Output:
# > Detected type: application/pdf
# > Analyzing 'scan_001.pdf' with Gemini...
# > Proposed Name: q1-financial-report-2024.pdf
# > Reasoning: The document summarizes the Q1 financial results for the fiscal year 2024.
# > Rename? [y/N]:
```

**Flags:**
- `--dry-run`: (Default: `false`) If true, only print, do not ask to rename.
- `--style`: (Default: `kebab`) Options: `snake`, `camel`, `pascal`.

### 4.2 Application Flow
1. **Bootstrap:** Load API Key (`GEMINI_API_KEY`) and initialize Adapters.
2. **Validation:** Check if input file exists.
3. **Detection:** Detect MIME type using magic numbers (via library).
4. **Logic:** If PDF, prepare for Blob upload. If Text, prepare for text prompt.
5. **Read:** Read file content.
6. **Prompt Construction:**
    - System Prompt: "You are an intelligent file renaming assistant. specific rules: 1. Analyze the attached content. 2. Summarize the content to identify its core subject. 3. Generate a concise, descriptive filename based on that summary. 4. Use kebab-case. 5. Return JSON: `{\"filename\": \"...\", \"reasoning\": \"...\"}`."
7. **Input:** The file content (Text or PDF Blob).
8. **API Call:** Send to Gemini Pro (1.5 Flash or Pro recommended for PDF window size).
9. **Parsing:** Decode the JSON response.
10. **Collision Check:** Check if `proposed_name` exists.
11. **Loop:** If exists, append `-1`, `-2`, etc., until unique. Example: `invoice.pdf` -> `invoice-1.pdf`.
12. **Interaction:** Present Old Name, New Name, and Reasoning.
13. **Execution:** If `y`, perform `os.Rename`.

### 4.3 Data Structures
```go
type RenameRequest struct {
    OriginalPath string
    Content      []byte
    MimeType     string
    Extension    string
}

type RenameResult struct {
    OriginalName string
    ProposedName string
    Reasoning    string
}
```

## 5. Technology Stack
- **Language:** Go 1.22+
- **CLI Framework:** `github.com/spf13/cobra`
- **Configuration:** `github.com/spf13/viper`
- **AI SDK:** `github.com/google/generative-ai-go` + `google.golang.org/api/option`
- **MIME Detection:** `github.com/gabriel-vasile/mimetype`
- **Logging:** `log/slog`

## 6. Security & Safety Considerations
- **PDF Size:** Large PDFs (>20MB) might hit gRPC limits if sent inline.
    - **Mitigation:** For MVP, limit file size to ~10MB. Post-MVP, use Gemini File API (Upload -> Reference URI -> Generate -> Delete).
- **Path Traversal:** Sanitize the returned filename to remove `/` or `\` characters.
- **Cost Control:** Use Gemini Flash model by default for speed and lower cost.

## 7. Future Proofing (Post-MVP)
- **Batch Mode:** Accepting `*.pdf` or directories.
- **Large File Support:** Implementing the full Gemini File API upload flow for large documents.
- **Undo functionality:** Keeping a transaction log.
