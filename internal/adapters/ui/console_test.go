package ui_test

import (
	"bytes"
	"errors"
	"reflect"
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
			want:     true,
			wantErr:  false,
		},
		{
			name:     "yes_full",
			input:    "yes\n",
			question: "Proceed?",
			want:     true,
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

func TestConsoleUI_SelectFiles(t *testing.T) {
	t.Parallel()

	files := []string{"a.txt", "b.txt", "c.txt"}
	tests := []struct {
		name       string
		input      string
		want       []string
		wantErr    error
		wantOutput string
	}{
		{
			name:  "select all",
			input: "a\n\n",
			want:  []string{"a.txt", "b.txt", "c.txt"},
		},
		{
			name:  "select none after all",
			input: "a\nn\n3\n\n",
			want:  []string{"c.txt"},
		},
		{
			name:  "comma list",
			input: "1,3\n\n",
			want:  []string{"a.txt", "c.txt"},
		},
		{
			name:  "range",
			input: "1-2\n\n",
			want:  []string{"a.txt", "b.txt"},
		},
		{
			name:       "empty selection asks again",
			input:      "\n2\n\n",
			want:       []string{"b.txt"},
			wantOutput: "Select at least one file",
		},
		{
			name:       "invalid selection asks again",
			input:      "4\n2\n\n",
			want:       []string{"b.txt"},
			wantOutput: "Invalid selection",
		},
		{
			name:    "cancel",
			input:   "q\n",
			wantErr: ui.ErrSelectionCancelled,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			in := strings.NewReader(tc.input)
			out := &bytes.Buffer{}
			c := ui.NewConsoleUIWithStreams(in, out)

			got, err := c.SelectFiles(files)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("SelectFiles() error = %v; want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("SelectFiles() error = %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("SelectFiles() = %#v; want %#v", got, tc.want)
			}
			if tc.wantOutput != "" && !strings.Contains(out.String(), tc.wantOutput) {
				t.Fatalf("output missing %q: %q", tc.wantOutput, out.String())
			}
		})
	}
}

func TestConsoleUI_SharedInputBuffer(t *testing.T) {
	t.Parallel()

	in := strings.NewReader("1\n\ny\n")
	out := &bytes.Buffer{}
	c := ui.NewConsoleUIWithStreams(in, out)

	selected, err := c.SelectFiles([]string{"a.txt"})
	if err != nil {
		t.Fatalf("SelectFiles() error = %v", err)
	}
	if !reflect.DeepEqual(selected, []string{"a.txt"}) {
		t.Fatalf("SelectFiles() = %#v; want a.txt", selected)
	}

	confirmed, err := c.Confirm("Rename?")
	if err != nil {
		t.Fatalf("Confirm() error = %v", err)
	}
	if !confirmed {
		t.Fatal("Confirm() = false; want true")
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
