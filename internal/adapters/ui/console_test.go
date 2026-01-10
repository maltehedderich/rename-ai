package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/maltehedderich/rename-ai/internal/adapters/ui"
)

func TestConsoleUI_Confirm(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		question string
		want     bool
		wantErr  bool
	}{
		{
			name:     "yes",
			input:    "y\n",
			question: "Proceed?",
			want:     true,
			wantErr:  false,
		},
		{
			name:     "YES",
			input:    "YES\n",
			question: "Proceed?",
			want:     false, // Only simple 'y' check in code: strings.ToLower(response) == "y"
			wantErr:  false,
		},
		{
			name:     "no",
			input:    "n\n",
			question: "Proceed?",
			want:     false,
			wantErr:  false,
		},
		{
			name:     "empty",
			input:    "\n",
			question: "Proceed?",
			want:     false,
			wantErr:  false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			in := strings.NewReader(tc.input)
			out := &bytes.Buffer{}
			c := ui.NewConsoleUIWithStreams(in, out)

			got, err := c.Confirm(tc.question)
			if (err != nil) != tc.wantErr {
				t.Errorf("Confirm() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if got != tc.want {
				t.Errorf("Confirm() = %v, want %v", got, tc.want)
			}

			// Verify question was printed
			if !strings.Contains(out.String(), tc.question) {
				t.Errorf("Expected output to contain %q, got %q", tc.question, out.String())
			}
		})
	}
}

func TestConsoleUI_PrintProposal(t *testing.T) {
	t.Parallel()

	in := strings.NewReader("")
	out := &bytes.Buffer{}
	c := ui.NewConsoleUIWithStreams(in, out)

	c.PrintProposal("old.txt", "new.txt", "because")

	output := out.String()
	if !strings.Contains(output, "old.txt") {
		t.Error("output missing old name")
	}
	if !strings.Contains(output, "new.txt") {
		t.Error("output missing new name")
	}
	if !strings.Contains(output, "because") {
		t.Error("output missing reasoning")
	}
}

func TestConsoleUI_Error(t *testing.T) {
	t.Parallel()

	in := strings.NewReader("")
	out := &bytes.Buffer{}
	c := ui.NewConsoleUIWithStreams(in, out)

	c.Error("something failed")

	if !strings.Contains(out.String(), "something failed") {
		t.Error("output missing error message")
	}
}
