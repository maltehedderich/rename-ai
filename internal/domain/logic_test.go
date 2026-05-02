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

func TestResolveBatchCollisions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		candidates     []domain.RenameCandidate
		existingPaths  []string
		wantFinalNames []string
		wantNoOps      []bool
	}{
		{
			name: "duplicate proposed names are reserved",
			candidates: []domain.RenameCandidate{
				{OriginalPath: "a.txt", ProposedName: "report.txt"},
				{OriginalPath: "b.txt", ProposedName: "report.txt"},
			},
			wantFinalNames: []string{"report.txt", "report-1.txt"},
			wantNoOps:      []bool{false, false},
		},
		{
			name: "existing filesystem collisions increment",
			candidates: []domain.RenameCandidate{
				{OriginalPath: "a.txt", ProposedName: "report.txt"},
			},
			existingPaths:  []string{"report.txt", "report-1.txt"},
			wantFinalNames: []string{"report-2.txt"},
			wantNoOps:      []bool{false},
		},
		{
			name: "target matching another selected source increments",
			candidates: []domain.RenameCandidate{
				{OriginalPath: "a.txt", ProposedName: "b.txt"},
				{OriginalPath: "b.txt", ProposedName: "c.txt"},
			},
			wantFinalNames: []string{"b-1.txt", "c.txt"},
			wantNoOps:      []bool{false, false},
		},
		{
			name: "same proposed name is no-op",
			candidates: []domain.RenameCandidate{
				{OriginalPath: "a.txt", ProposedName: "a.txt"},
			},
			existingPaths:  []string{"a.txt"},
			wantFinalNames: []string{"a.txt"},
			wantNoOps:      []bool{true},
		},
		{
			name: "collisions are scoped by directory",
			candidates: []domain.RenameCandidate{
				{OriginalPath: "one/a.txt", ProposedName: "report.txt"},
				{OriginalPath: "two/a.txt", ProposedName: "report.txt"},
			},
			wantFinalNames: []string{"report.txt", "report.txt"},
			wantNoOps:      []bool{false, false},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			existing := make(map[string]struct{})
			for _, path := range tc.existingPaths {
				existing[path] = struct{}{}
			}

			got := domain.ResolveBatchCollisions(tc.candidates, func(path string) bool {
				_, ok := existing[path]
				return ok
			})

			if len(got) != len(tc.wantFinalNames) {
				t.Fatalf("ResolveBatchCollisions() returned %d results, want %d", len(got), len(tc.wantFinalNames))
			}

			for i, result := range got {
				if result.FinalName != tc.wantFinalNames[i] {
					t.Errorf("result[%d].FinalName = %q; want %q", i, result.FinalName, tc.wantFinalNames[i])
				}
				if result.NoOp != tc.wantNoOps[i] {
					t.Errorf("result[%d].NoOp = %v; want %v", i, result.NoOp, tc.wantNoOps[i])
				}
			}
		})
	}
}
