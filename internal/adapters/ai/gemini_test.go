package ai

import "testing"

func TestParseAIResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         string
		wantFilename  string
		wantReasoning string
		wantErr       bool
	}{
		{
			name:          "valid json",
			input:         `{"filename": "test-file.txt", "reasoning": "it's a test"}`,
			wantFilename:  "test-file.txt",
			wantReasoning: "it's a test",
			wantErr:       false,
		},
		{
			name:          "json with markdown code block",
			input:         "```json\n{\"filename\": \"test-file.txt\", \"reasoning\": \"it's a test\"}\n```",
			wantFilename:  "test-file.txt",
			wantReasoning: "it's a test",
			wantErr:       false,
		},
		{
			name:          "json with plain code block",
			input:         "```\n{\"filename\": \"test-file.txt\", \"reasoning\": \"it's a test\"}\n```",
			wantFilename:  "test-file.txt",
			wantReasoning: "it's a test",
			wantErr:       false,
		},
		{
			name:    "invalid json",
			input:   `{invalid json}`,
			wantErr: true,
		},
		{
			name:    "empty filename",
			input:   `{"filename": "", "reasoning": "fail"}`,
			wantErr: true,
		},
		{
			name:    "missing filename field",
			input:   `{"reasoning": "fail"}`,
			wantErr: true,
		},
		{
			name:          "extra fields",
			input:         `{"filename": "extra.txt", "reasoning": "ok", "other": "ignored"}`,
			wantFilename:  "extra.txt",
			wantReasoning: "ok",
			wantErr:       false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotFilename, gotReasoning, err := parseAIResponse(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("parseAIResponse() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if gotFilename != tc.wantFilename {
				t.Errorf("parseAIResponse() filename = %v, want %v", gotFilename, tc.wantFilename)
			}
			if gotReasoning != tc.wantReasoning {
				t.Errorf("parseAIResponse() reasoning = %v, want %v", gotReasoning, tc.wantReasoning)
			}
		})
	}
}
