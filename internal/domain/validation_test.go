package domain_test

import (
	"testing"

	"github.com/maltehedderich/rename-ai/internal/domain"
)

func TestIsAllowedMimeType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		mimeType string
		wantErr  bool
	}{
		// Allowed types
		{"PDF", "application/pdf", false},
		{"PlainText", "text/plain", false},
		{"Python", "text/x-python", false},
		{"JSON", "application/json", false},
		{"PNG", "image/png", false},
		{"JPEG", "image/jpeg", false},
		{"MP4", "video/mp4", false},
		{"QuickTime", "video/quicktime", false}, // Standard for .mov
		{"UserMOV", "video/mov", false},         // User valid
		{"MP3", "audio/mpeg", false},            // Standard for .mp3
		{"UserMP3", "audio/mp3", false},         // User valid
		{"WithParams", "text/plain; charset=utf-8", false},

		// Disallowed types
		{"Binary", "application/octet-stream", true},
		{"Executable", "application/x-dosexec", true},
		{"Unknown", "application/unknown", true},
		{"Zip", "application/zip", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := domain.IsAllowedMimeType(tt.mimeType)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsAllowedMimeType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
