package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"google.golang.org/genai"
)

type GeminiProvider struct {
	client *genai.Client
	model  string
}

func NewGeminiProvider(ctx context.Context, apiKey string, modelName string) (*GeminiProvider, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	// Use a capable model (no changes to model name passing, but verify default)
	return &GeminiProvider{
		client: client,
		model:  modelName,
	}, nil
}

type aiResponse struct {
	Filename  string `json:"filename"`
	Reasoning string `json:"reasoning"`
}

func (p *GeminiProvider) GenerateName(ctx context.Context, content []byte, mimeType string, currentExt string) (string, string, error) {
	currentDate := time.Now().Format("2006-01-02")
	prompt := fmt.Sprintf(`You are an intelligent file renaming assistant.
		Context:
		- Current Date: %s

		Specific rules:
		1. Analyze the attached content.
		2. Summarize the content to identify its core subject and any relevant date.
		3. Generate a filename adhering to the following structure: YYYY-MM-DD_Subject-Title%s
		   - Always start with a date in ISO 8601 format (YYYY-MM-DD). If no specific date is found in the content, use the Current Date provided above as a fallback.
		   - Use underscores (_) to separate the date from the subject/title.
		   - Use hyphens (-) to separate words within the subject/title.
		   - Alternatively, use CamelCase for the subject (e.g., BudgetReport) if appropriate.
		   - Ensure the filename ends with the extension "%s".
		4. Example: 2023-12-01_Budget-Report%s`, currentDate, currentExt, currentExt, currentExt)

	// Define the schema for structured output
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"filename": map[string]any{
				"type":        "string",
				"description": "The generated filename, starting with YYYY-MM-DD, followed by an underscore and the subject/title, and ending with the correct extension.",
			},
			"reasoning": map[string]any{
				"type":        "string",
				"description": "Brief explanation of why this name was chosen.",
			},
		},
		"required": []string{"filename", "reasoning"},
	}

	config := &genai.GenerateContentConfig{
		ResponseMIMEType:   "application/json",
		ResponseJsonSchema: schema,
		SystemInstruction:  &genai.Content{Parts: []*genai.Part{{Text: prompt}}},
	}

	var part *genai.Part
	if strings.HasPrefix(mimeType, "text/") || mimeType == "application/json" || strings.Contains(mimeType, "xml") {
		part = &genai.Part{Text: string(content)}
	} else {
		part = &genai.Part{InlineData: &genai.Blob{
			MIMEType: mimeType,
			Data:     content,
		}}
	}

	// Construct the content with the part
	userContent := &genai.Content{
		Parts: []*genai.Part{
			{Text: "Analyze the following file content and generate a filename."},
			part,
		},
	}

	resp, err := p.client.Models.GenerateContent(
		ctx,
		p.model,
		[]*genai.Content{userContent},
		config,
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate content: %w", err)
	}

	return parseAIResponse(resp.Text())
}

// parseAIResponse handles the unmarshalling of the JSON response
func parseAIResponse(respText string) (string, string, error) {
	var result aiResponse
	// Clean up potential markdown code blocks if the AI wraps the JSON (SDK might not, but safe to keep)
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
		return "", "", fmt.Errorf("failed to parse AI response: %w (response: %s)", err, respText)
	}

	if result.Filename == "" {
		return "", "", fmt.Errorf("AI response contained empty filename")
	}

	return result.Filename, result.Reasoning, nil
}
