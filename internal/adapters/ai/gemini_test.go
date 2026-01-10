package ai

import "testing"

func TestParseAIResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid json",
			input:   `{"filename": "test-file.txt", "reasoning": "it's a test"}`,
			want:    "test-file.txt",
			wantErr: false,
		},
		{
			name:    "json with markdown code block",
			input:   "```json\n{\"filename\": \"test-file.txt\", \"reasoning\": \"it's a test\"}\n```",
			want:    "test-file.txt",
			wantErr: false,
		},
		{
			name:    "json with plain code block",
			input:   "```\n{\"filename\": \"test-file.txt\", \"reasoning\": \"it's a test\"}\n```",
			want:    "test-file.txt",
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   `{invalid json}`,
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty filename",
			input:   `{"filename": "", "reasoning": "fail"}`,
			want:    "",
			wantErr: true,
		},
		{
			name:    "missing filename field",
			input:   `{"reasoning": "fail"}`,
			want:    "",
			wantErr: true,
		},
		{
			name:    "extra fields",
			input:   `{"filename": "extra.txt", "reasoning": "ok", "other": "ignored"}`,
			want:    "extra.txt",
			wantErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseAIResponse(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("parseAIResponse() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if got != tc.want {
				t.Errorf("parseAIResponse() = %v, want %v", got, tc.want)
			}
		})
	}
}
