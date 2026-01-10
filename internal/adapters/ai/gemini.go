package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiProvider struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

func NewGeminiProvider(ctx context.Context, apiKey string) (*GeminiProvider, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	// Use flash model by default for speed/cost as per spec
	model := client.GenerativeModel("gemini-flash-latest")
	model.ResponseMIMEType = "application/json"

	return &GeminiProvider{
		client: client,
		model:  model,
	}, nil
}

type aiResponse struct {
	Filename  string `json:"filename"`
	Reasoning string `json:"reasoning"`
}

func (p *GeminiProvider) GenerateName(ctx context.Context, content []byte, mimeType string, currentExt string) (string, error) {
	prompt := fmt.Sprintf(`You are an intelligent file renaming assistant.
		Specific rules:
		1. Analyze the attached content.
		2. Summarize the content to identify its core subject.
		3. Generate a concise, descriptive filename based on that summary.
		4. Use kebab-case.
		5. Ensure the filename ends with the extension "%s".
		6. Return JSON: {"filename": "...", "reasoning": "..."}.`, currentExt)

	var resp *genai.GenerateContentResponse
	var err error

	if strings.HasPrefix(mimeType, "text/") || mimeType == "application/json" || strings.Contains(mimeType, "xml") {
		// Treat as text
		resp, err = p.model.GenerateContent(ctx, genai.Text(prompt), genai.Text(string(content)))
	} else {
		// Treat as blob (e.g. PDF, Image)
		resp, err = p.model.GenerateContent(ctx, genai.Text(prompt), genai.Blob{
			MIMEType: mimeType,
			Data:     content,
		})
	}

	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return "", fmt.Errorf("no candidates returned from AI")
	}

	// Extract text from part
	var respText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			respText += string(txt)
		}
	}

	// Parse JSON using the extracted helper
	return parseAIResponse(respText)
}

// parseAIResponse is extracted to allow unit testing of the parsing logic
func parseAIResponse(respText string) (string, error) {
	var result aiResponse
	// Clean up potential markdown code blocks if the AI wraps the JSON
	cleaned := strings.TrimSpace(respText)
	if strings.HasPrefix(cleaned, "```json") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimSuffix(cleaned, "```")
	} else if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
	}
	cleaned = strings.TrimSpace(cleaned)

	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return "", fmt.Errorf("failed to parse AI response: %w (response: %s)", err, respText)
	}

	if result.Filename == "" {
		return "", fmt.Errorf("AI response contained empty filename")
	}

	return result.Filename, nil
}
