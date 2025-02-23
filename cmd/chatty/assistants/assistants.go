package assistants

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

// Common directives that apply to all assistants
const (
	// Control whether common directives are included by default
	includeCommonDirectives = true

	// Language setting (en-US, es-ES, fr-FR, de-DE, it-IT, pt-BR, ja-JP, ko-KR, zh-CN)
	languageCode = "en-US"

	// Common directives template - for natural conversations
	commonDirectivesTemplate = `Language: %s

Chat like a human friend - be brief, casual, and engaging. Provide accurate information and acknowledge uncertainty. Keep responses short and break up long explanations into dialogue. Ask questions when needed.`

	// Built-in assistants directory
	builtinDir = "builtin"
)

// AssistantConfig holds all configuration for an assistant's identity and appearance
type AssistantConfig struct {
	Name          string `yaml:"name"`
	SystemMessage string `yaml:"system_message"`
	Emoji         string `yaml:"emoji"`
	LabelColor    string `yaml:"label_color"`
	TextColor     string `yaml:"text_color"`
	Description   string `yaml:"description"`
	IsDefault     bool   `yaml:"is_default"`
}

// Get complete system message including directives
func (a *AssistantConfig) GetFullSystemMessage() string {
	// Include common directives if enabled
	if includeCommonDirectives {
		directives := fmt.Sprintf(commonDirectivesTemplate, languageCode)
		return fmt.Sprintf("%s\n%s", a.SystemMessage, directives)
	}
	return a.SystemMessage
}

var (
	// builtinAssistants holds all assistants loaded from YAML files
	builtinAssistants []AssistantConfig
	// DefaultAssistant is set during initialization
	DefaultAssistant AssistantConfig
)

// loadBuiltinAssistants loads all YAML files from the builtin directory
func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	
	// Load built-in assistants
	files, err := os.ReadDir(filepath.Join(dir, builtinDir))
	if err != nil {
		panic(fmt.Sprintf("Failed to read builtin assistants directory: %v", err))
	}

	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			path := filepath.Join(dir, builtinDir, file.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				panic(fmt.Sprintf("Failed to read assistant file %s: %v", file.Name(), err))
			}

			var assistant AssistantConfig
			if err := yaml.Unmarshal(data, &assistant); err != nil {
				panic(fmt.Sprintf("Failed to parse assistant file %s: %v", file.Name(), err))
			}

			builtinAssistants = append(builtinAssistants, assistant)
			if assistant.IsDefault {
				DefaultAssistant = assistant
			}
		}
	}

	// If no default assistant was specified, use the first one
	if DefaultAssistant.Name == "" && len(builtinAssistants) > 0 {
		DefaultAssistant = builtinAssistants[0]
	}
}

// GetAssistantConfig returns the specified assistant configuration or the default
func GetAssistantConfig(name string) AssistantConfig {
	for _, assistant := range builtinAssistants {
		if strings.EqualFold(assistant.Name, name) {
			return assistant
		}
	}
	return DefaultAssistant
}

// ListAssistants returns a formatted string of all available assistants
func ListAssistants() string {
	var sb strings.Builder
	sb.WriteString("Available assistants:\n")
	for _, assistant := range builtinAssistants {
		if assistant.Name == DefaultAssistant.Name {
			sb.WriteString(fmt.Sprintf("* %s (Default) - %s\n", assistant.Name, assistant.Description))
		} else {
			sb.WriteString(fmt.Sprintf("* %s - %s\n", assistant.Name, assistant.Description))
		}
	}
	return sb.String()
}

// IsValidAssistant checks if the given name is a valid assistant
func IsValidAssistant(name string) bool {
	for _, assistant := range builtinAssistants {
		if strings.EqualFold(assistant.Name, name) {
			return true
		}
	}
	return false
}

// GetHistoryFileName returns the history filename for a given assistant
func GetHistoryFileName(assistantName string) string {
	return fmt.Sprintf("chat_history_%s.json", strings.ToLower(assistantName))
}

// GetFormattedLabelColor returns the properly formatted ANSI color code
func (a *AssistantConfig) GetFormattedLabelColor() string {
	return a.LabelColor
}

// GetFormattedTextColor returns the properly formatted ANSI color code
func (a *AssistantConfig) GetFormattedTextColor() string {
	return a.TextColor
} 