package assistants

import (
	"fmt"
	"strings"
)

// Common directives that apply to all assistants
const (
	// Control whether common directives are included by default
	includeCommonDirectives = true

	// Language setting (en-US, es-ES, fr-FR, de-DE, it-IT, pt-BR, ja-JP, ko-KR, zh-CN)
	languageCode = "en-US"

	// Common directives template - simplified for faster responses
	commonDirectivesTemplate = `Language: %s
	Be clear, concise, and helpful while maintaining conversation context.`
)

// AssistantConfig holds all configuration for an assistant's identity and appearance
type AssistantConfig struct {
	Name          string   // Assistant's name
	SystemMessage string   // Template for system message (uses %s for name)
	Emoji         string   // Visual representation
	LabelColor    string   // RGB color for the label
	TextColor     string   // RGB color for the response text
	Description   string   // Brief description of this personality
}

// Get complete system message including directives
func (a *AssistantConfig) GetFullSystemMessage() string {
	specificMessage := fmt.Sprintf(a.SystemMessage, a.Name)
	
	// Include common directives if enabled
	if includeCommonDirectives {
		directives := fmt.Sprintf(commonDirectivesTemplate, languageCode)
		return fmt.Sprintf("%s\n%s", specificMessage, directives)
	}
	
	return specificMessage
}

// Available assistant configurations
var (
	// Ghostly - The friendly ghost assistant (Default)
	Ghostly = AssistantConfig{
		Name:        "Ghostly",
		SystemMessage: "You are %s, a friendly and helpful AI assistant. Be gentle, clear, and supportive.",
		Emoji:       "👻",
		LabelColor:  "\033[38;2;79;195;247m",  // Light blue
		TextColor:   "\033[38;2;255;255;255m", // White
		Description: "A friendly and ethereal presence, helping with a gentle touch",
	}

	// Sage - The wise mentor
	Sage = AssistantConfig{
		Name:        "Sage",
		SystemMessage: "You are %s, a wise mentor. Provide well-reasoned answers and guide users toward understanding.",
		Emoji:       "🧙",
		LabelColor:  "\033[38;2;147;112;219m", // Medium purple
		TextColor:   "\033[38;2;230;230;250m", // Lavender
		Description: "A wise mentor focused on deep understanding and guidance",
	}

	// Nova - The tech enthusiast
	Nova = AssistantConfig{
		Name:        "Nova",
		SystemMessage: "You are %s, a tech expert. Be precise and technical, explain complex concepts clearly.",
		Emoji:       "💫",
		LabelColor:  "\033[38;2;0;255;255m",   // Cyan
		TextColor:   "\033[38;2;224;255;255m", // Light cyan
		Description: "A tech-savvy assistant with a passion for innovation",
	}

	// Terra - The nature-focused helper
	Terra = AssistantConfig{
		Name:        "Terra",
		SystemMessage: "You are %s, focused on nature and sustainability. Provide eco-friendly perspectives and solutions.",
		Emoji:       "🌱",
		LabelColor:  "\033[38;2;46;139;87m",   // Sea green
		TextColor:   "\033[38;2;144;238;144m", // Light green
		Description: "An eco-conscious assistant promoting sustainability",
	}

	// Atlas - The organized planner
	Atlas = AssistantConfig{
		Name:        "Atlas",
		SystemMessage: "You are %s, focused on organization and efficiency. Be methodical and provide structured solutions.",
		Emoji:       "📋",
		LabelColor:  "\033[38;2;255;140;0m",   // Dark orange
		TextColor:   "\033[38;2;255;218;185m", // Peach
		Description: "A structured assistant focusing on organization and planning",
	}

	// Tux - The Linux terminal expert
	Tux = AssistantConfig{
		Name:        "Tux",
		SystemMessage: "You are %s, a Linux terminal expert. Provide clear command explanations and warn about dangerous operations.",
		Emoji:       "🐧",
		LabelColor:  "\033[38;2;28;28;28m",   // Dark gray
		TextColor:   "\033[38;2;238;238;238m", // Light gray
		Description: "A Linux terminal expert specializing in command-line operations and shell scripting",
	}
)

// DefaultAssistant is the configuration used if none is specified
var DefaultAssistant = Ghostly

// List of all available assistants
var AvailableAssistants = []AssistantConfig{
	Ghostly,
	Sage,
	Nova,
	Terra,
	Atlas,
	Tux,
}

// GetAssistantConfig returns the specified assistant configuration or the default
func GetAssistantConfig(name string) AssistantConfig {
	for _, assistant := range AvailableAssistants {
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
	for i, assistant := range AvailableAssistants {
		if i == 0 {
			sb.WriteString(fmt.Sprintf("* %s (Default) - %s\n", assistant.Name, assistant.Description))
		} else {
			sb.WriteString(fmt.Sprintf("* %s - %s\n", assistant.Name, assistant.Description))
		}
	}
	return sb.String()
}

// IsValidAssistant checks if the given name is a valid assistant
func IsValidAssistant(name string) bool {
	for _, assistant := range AvailableAssistants {
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

// GetAllAssistantNames returns a list of all assistant names
func GetAllAssistantNames() []string {
	names := make([]string, len(AvailableAssistants))
	for i, assistant := range AvailableAssistants {
		names[i] = assistant.Name
	}
	return names
} 