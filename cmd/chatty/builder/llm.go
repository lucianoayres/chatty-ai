package builder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// LLMClient defines the interface for interacting with language models
type LLMClient interface {
	Generate(systemPrompt, userInput string) (string, error)
}

// OllamaClient implements LLMClient for the Ollama API
type OllamaClient struct {
	baseURL string
	model   string
	debug   bool
}

// GenerateRequest represents a request to the generate API
type GenerateRequest struct {
	Model    string         `json:"model"`
	Prompt   string         `json:"prompt"`
	System   string         `json:"system"`
	Format   map[string]any `json:"format"`    // JSON Schema format specification
	Stream   bool          `json:"stream"`
}

// GenerateResponse represents a response from the generate API
type GenerateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// NewOllamaClient creates a new Ollama API client
func NewOllamaClient(baseURL, model string) *OllamaClient {
	return &OllamaClient{
		baseURL: baseURL,
		model:   model,
		debug:   false,
	}
}

// SetDebug enables or disables debug mode
func (c *OllamaClient) SetDebug(debug bool) {
	c.debug = debug
}

// Generate sends a request to the LLM and returns the generated text
func (c *OllamaClient) Generate(systemPrompt, userInput string) (string, error) {
	// Define the JSON schema for the response format
	format := map[string]any{
		"type": "object",
		"required": []string{
			"name",
			"system_message",
			"emoji",
			"description",
		},
		"properties": map[string]any{
			"name": map[string]any{
				"type": "string",
				"description": "The agent's display name",
			},
			"system_message": map[string]any{
				"type": "string",
				"description": "Comprehensive system prompt defining the agent's personality and behavior",
			},
			"emoji": map[string]any{
				"type": "string",
				"description": "A single emoji that best represents the agent",
			},
			"description": map[string]any{
				"type": "string",
				"description": "Brief description of the agent",
			},
		},
	}

	// Prepare the generate request
	req := GenerateRequest{
		Model:  c.model,
		Prompt: userInput,
		System: systemPrompt,
		Format: format,
		Stream: false,
	}

	// Marshal the request to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	if c.debug {
		fmt.Printf("\n📤 Debug Mode: Request to %s:\n", c.baseURL)
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, jsonData, "", "  "); err == nil {
			fmt.Printf("%s\n", prettyJSON.String())
		}
	}

	// Create the HTTP request
	httpReq, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if c.debug {
			fmt.Printf("\n❌ Debug Mode: API Error (Status %d):\n%s\n", resp.StatusCode, string(body))
		}
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if c.debug {
		fmt.Printf("\n📥 Debug Mode: Raw Response:\n%s\n", string(body))
	}

	// Try to parse as GenerateResponse
	var genResp GenerateResponse
	if err := json.Unmarshal(body, &genResp); err != nil {
		// Try to parse error response
		var errResp struct {
			Error string `json:"error"`
		}
		if jsonErr := json.Unmarshal(body, &errResp); jsonErr == nil && errResp.Error != "" {
			return "", fmt.Errorf("API error: %s", errResp.Error)
		}
		return "", fmt.Errorf("failed to parse response: %v\nResponse body: %s", err, string(body))
	}

	if genResp.Response == "" {
		return "", fmt.Errorf("empty response from API")
	}

	if c.debug {
		fmt.Printf("\n📝 Debug Mode: Parsed Response:\n%s\n", genResp.Response)
	}

	return genResp.Response, nil
} 