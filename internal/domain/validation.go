package domain

import (
	"fmt"
	"strings"
)

var allowedMimeTypes = map[string]struct{}{
	// Documents & Text
	"application/pdf": {}, // PDF
	"text/plain":      {}, // Plain Text
	// Code
	"text/x-python":    {},
	"text/javascript":  {},
	"application/json": {},
	"text/html":        {},
	"text/css":         {},
	"text/xml":         {},
	"application/xml":  {}, // XML alias
	"text/markdown":    {}, // Markdown standard
	"text/md":          {}, // Markdown alias (user specified)
	"text/x-markdown":  {}, // Markdown alias
	"text/csv":         {},

	// Images
	"image/png":  {},
	"image/jpeg": {},
	"image/webp": {},
	"image/heic": {},
	"image/heif": {},

	// Video
	"video/mp4":       {},
	"video/mpeg":      {},
	"video/mov":       {}, // User specified
	"video/quicktime": {}, // MOV standard
	"video/avi":       {}, // User specified
	"video/x-msvideo": {}, // AVI standard
	"video/x-flv":     {},
	"video/mpg":       {}, // User specified
	"video/webm":      {},
	"video/wmv":       {}, // User specified
	"video/x-ms-wmv":  {}, // WMV standard
	"video/3gpp":      {},

	// Audio
	"audio/wav":    {},
	"audio/x-wav":  {}, // WAV alias
	"audio/mp3":    {}, // User specified
	"audio/mpeg":   {}, // MP3 standard
	"audio/aiff":   {},
	"audio/x-aiff": {}, // AIFF alias
	"audio/aac":    {},
	"audio/ogg":    {},
	"audio/flac":   {},
}

// IsAllowedMimeType checks if the given mimeType is supported.
// It returns nil if allowed, or a descriptive error if not.
func IsAllowedMimeType(mimeType string) error {
	// Simple normalize: remove parameters like "; charset=utf-8"
	baseMime := strings.Split(mimeType, ";")[0]
	baseMime = strings.TrimSpace(baseMime)

	if _, ok := allowedMimeTypes[baseMime]; ok {
		return nil
	}

	return fmt.Errorf("unsupported file type: %s. Supported categories: Documents, Images, Video, Audio", mimeType)
}
