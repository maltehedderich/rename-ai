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
