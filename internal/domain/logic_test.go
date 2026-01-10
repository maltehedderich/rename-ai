package domain_test

import (
	"strings"
	"testing"

	"github.com/maltehedderich/rename-ai/internal/domain"
)

func TestSanitizeFilename(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid simple name",
			input:    "simple.txt",
			expected: "simple.txt",
		},
		{
			name:     "spaces replaced by dashes",
			input:    "my file.txt",
			expected: "my-file.txt",
		},
		{
			name:     "uppercase to lowercase",
			input:    "MyFile.TXT",
			expected: "myfile.txt",
		},
		{
			name:     "multiple special chars",
			input:    "foo@bar#baz!.txt",
			expected: "foo-bar-baz-.txt",
		},
		{
			name:     "preserve underscores and dashes",
			input:    "foo_bar-baz.txt",
			expected: "foo_bar-baz.txt",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := domain.SanitizeFilename(tc.input)
			if got != tc.expected {
				t.Errorf("SanitizeFilename(%q) = %q; want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestResolveCollision(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		baseName      string
		existingFiles []string
		expected      string
	}{
		{
			name:          "no collision",
			baseName:      "file.txt",
			existingFiles: []string{},
			expected:      "file.txt",
		},
		{
			name:          "single collision",
			baseName:      "file.txt",
			existingFiles: []string{"file.txt"},
			expected:      "file-1.txt",
		},
		{
			name:          "multiple collisions",
			baseName:      "file.txt",
			existingFiles: []string{"file.txt", "file-1.txt"},
			expected:      "file-2.txt",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			checkExists := func(path string) bool {
				for _, f := range tc.existingFiles {
					if strings.HasSuffix(path, f) { // simplified check
						return true
					}
				}
				return false
			}

			got := domain.ResolveCollision(tc.baseName, checkExists)
			if got != tc.expected {
				t.Errorf("ResolveCollision(%q) = %q; want %q", tc.baseName, got, tc.expected)
			}
		})
	}
}
