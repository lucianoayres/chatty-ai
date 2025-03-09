package builder

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

// AgentSchema represents the JSON schema for agent configuration
type AgentSchema struct {
	Name          string   `json:"name" yaml:"name"`
	SystemMessage string   `json:"system_message" yaml:"system_message"`
	Emoji         string   `json:"emoji" yaml:"emoji"`
	LabelColor    string   `json:"label_color" yaml:"label_color"`
	TextColor     string   `json:"text_color" yaml:"text_color"`
	Description   string   `json:"description" yaml:"description"`
	IsDefault     bool     `json:"is_default" yaml:"is_default"`
	Tags          []string `json:"tags" yaml:"tags"`
}

// BuilderConfig holds configuration for the agent builder
type BuilderConfig struct {
	MaxExampleAgents int      // Maximum number of example agents to include
	ExampleFiles     []string // List of example agent filenames (without .yaml extension)
	BuiltinDir       string   // Directory containing built-in agents
}

// DefaultBuilderConfig returns the default builder configuration
func DefaultBuilderConfig() BuilderConfig {
	return BuilderConfig{
		MaxExampleAgents: 3,
		ExampleFiles: []string{
			"ada",      // Technical expert
			"gandalf",  // Fictional character
			"cleopatra", // Historical figure
			"tesla",    // Scientist
			"shakespeare", // Artist
		},
		BuiltinDir: "builtin",
	}
}

// Builder handles agent generation functionality
type Builder struct {
	config BuilderConfig
	llm    LLMClient
}

// NewBuilder creates a new agent builder with the given configuration
func NewBuilder(config BuilderConfig, llm LLMClient) *Builder {
	return &Builder{
		config: config,
		llm:    llm,
	}
}

// loadExampleAgents loads example agent configurations from files
func (b *Builder) loadExampleAgents() ([]AgentSchema, error) {
	_, filename, _, _ := runtime.Caller(0)
	builtinPath := filepath.Join(filepath.Dir(filename), "..", "agents", b.config.BuiltinDir)

	var examples []AgentSchema
	for _, name := range b.config.ExampleFiles {
		if len(examples) >= b.config.MaxExampleAgents {
			break
		}

		path := filepath.Join(builtinPath, name+".yaml")
		data, err := os.ReadFile(path)
		if err != nil {
			continue // Skip if file not found
		}

		var agent AgentSchema
		if err := yaml.Unmarshal(data, &agent); err != nil {
			continue // Skip if invalid YAML
		}

		// Only include the fields we want to show as examples
		examples = append(examples, AgentSchema{
			Name:          agent.Name,
			SystemMessage: agent.SystemMessage,
			Emoji:         agent.Emoji,
			Description:   agent.Description,
		})
	}

	return examples, nil
}

// getSystemPrompt returns the system prompt for agent generation
func (b *Builder) getSystemPrompt() (string, error) {
	examples, err := b.loadExampleAgents()
	if err != nil {
		return "", fmt.Errorf("failed to load example agents: %v", err)
	}

	var examplesStr strings.Builder
	for i, ex := range examples {
		// Create a simplified example with only the fields we want the LLM to generate
		example := struct {
			Name          string `json:"name"`
			SystemMessage string `json:"system_message"`
			Emoji         string `json:"emoji"`
			Description   string `json:"description"`
		}{
			Name:          ex.Name,
			SystemMessage: ex.SystemMessage,
			Emoji:         ex.Emoji,
			Description:   ex.Description,
		}

		exampleJSON, err := json.MarshalIndent(example, "    ", "    ")
		if err != nil {
			continue
		}
		examplesStr.WriteString(fmt.Sprintf("Example %d:\n%s\n\n", i+1, string(exampleJSON)))
	}

	return fmt.Sprintf(`You are an expert AI Agent Generator, specializing in creating well-defined agent personalities for AI interactions.

Your task is to generate a new agent configuration based on the user's description.

CRITICAL REQUIREMENTS:
1. You MUST respond with a SINGLE valid JSON object
2. The JSON MUST contain ALL of these fields:
   - "name": A concise, descriptive name for the agent
   - "system_message": A comprehensive system prompt defining behavior
   - "emoji": A single emoji that represents the agent's role
   - "description": A brief description of the agent's purpose
   - "tags": You do not need to provide this field, as tags will be added by the user later

FORMATTING RULES:
1. NO explanatory text before or after the JSON
2. The JSON must be properly formatted and valid
3. All field names must be exactly as specified
4. All fields must have non-empty string values

Here are some example agents for reference:

%s

FIELD GUIDELINES:

1. name:
   - Should be concise and memorable
   - Use proper capitalization
   - No special characters except spaces
   - Example: "Tech Guru" or "History Scholar"

2. system_message:
   - Must be comprehensive and detailed
   - Define the agent's identity, expertise, and behavior
   - Include communication style and boundaries
   - Minimum 100 characters

3. emoji:
   - Must be a single Unicode emoji
   - Should clearly represent the agent's role
   - No text, just the emoji character

4. description:
   - Brief but informative summary
   - One or two sentences maximum
   - Highlight key capabilities
   - No more than 100 characters

Remember: Your response must be a single JSON object with all required fields. Any missing or empty fields will cause an error.`, examplesStr.String()), nil
}

// BuildAgent generates a new agent configuration based on user input
func (b *Builder) BuildAgent(description string) (*AgentSchema, error) {
	systemPrompt, err := b.getSystemPrompt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate system prompt: %v", err)
	}

	// Generate the agent configuration using the LLM
	response, err := b.llm.Generate(systemPrompt, description)
	if err != nil {
		return nil, fmt.Errorf("failed to generate agent configuration: %v", err)
	}

	// Parse the JSON response
	var agent AgentSchema
	if err := json.Unmarshal([]byte(response), &agent); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response as JSON: %v", err)
	}

	// Only validate the fields we expect from the LLM
	if agent.Name == "" || agent.SystemMessage == "" || agent.Description == "" {
		return nil, fmt.Errorf("the AI needs more details to create your agent")
	}

	// If the emoji is not set, set it to a default value
	if agent.Emoji == "" {
		agent.Emoji = "🤖"
	}

	// Initialize the remaining fields with empty values
	// These will be set by the handler later
	agent.LabelColor = ""
	agent.TextColor = ""
	agent.IsDefault = false

	return &agent, nil
}

// SaveAgent saves the generated agent configuration to a file
func (b *Builder) SaveAgent(agent *AgentSchema, outputPath string) error {
	data, err := yaml.Marshal(agent)
	if err != nil {
		return fmt.Errorf("failed to marshal agent configuration: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write agent configuration: %v", err)
	}

	return nil
} 