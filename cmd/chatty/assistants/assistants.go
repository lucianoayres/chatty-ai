package assistants

import (
	"fmt"
	"strings"
)

// Common directives that apply to all assistants
const (
	// Control whether common directives are included by default
	includeCommonDirectives = true

	// Default language code for responses (ISO 639-1 or ISO 639-1 with region code)
	// Common language codes:
	// Simple codes:
	// en - English
	// es - Spanish
	// fr - French
	// de - German
	// it - Italian
	// pt - Portuguese
	// ru - Russian
	// zh - Chinese
	// ja - Japanese
	// ko - Korean
	//
	// With region codes:
	// en-US - American English
	// en-GB - British English
	// en-AU - Australian English
	// es-ES - Spanish (Spain)
	// es-MX - Spanish (Mexico)
	// pt-BR - Portuguese (Brazil)
	// pt-PT - Portuguese (Portugal)
	// zh-CN - Chinese (Simplified)
	// zh-TW - Chinese (Traditional)
	// fr-CA - French (Canada)
	// fr-FR - French (France)
	defaultLanguage = "en-US"

	// Common directives template that includes language settings
	commonDirectivesTemplate = `
Language: %s

General Guidelines:
1. Communication: Always be clear, concise, and professional
2. Accuracy: Verify information before providing it
3. Helpfulness: Focus on practical, actionable solutions
4. Ethics: Follow ethical principles and respect user privacy
5. Clarity: If unsure, ask for clarification rather than making assumptions
6. Context: Maintain conversation context and reference previous interactions when relevant
7. Format: Use plain text for all responses, not markdown. Structure responses with clear sections and bullet points when needed
8. Style: Keep responses clean and readable without special formatting characters`
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

// Get the complete common directives with language settings
func getCommonDirectives() string {
	return fmt.Sprintf(commonDirectivesTemplate, defaultLanguage)
}

// Get complete system message including directives
func (a *AssistantConfig) GetFullSystemMessage() string {
	specificMessage := fmt.Sprintf(a.SystemMessage, a.Name, a.Name)
	
	// Include common directives if enabled
	if includeCommonDirectives {
		return fmt.Sprintf("%s\n\n%s", specificMessage, getCommonDirectives())
	}
	
	return specificMessage
}

// Available assistant configurations
var (
	// Ghostly - The friendly ghost assistant (Default)
	Ghostly = AssistantConfig{
		Name:        "Ghostly",
		SystemMessage: "You are %s, an AI assistant with a friendly and ethereal presence. Your core traits:\n" +
			"1. Identity: Always identify as %s, a helpful spirit in the digital realm\n" +
			"2. Communication: Be gentle, clear, and supportive in your responses\n" +
			"3. Accuracy: Provide accurate information with a touch of wisdom\n" +
			"4. Helpfulness: Guide users with patience and understanding\n" +
			"5. Personality: Maintain a light, friendly tone while being professional",
		Emoji:       "👻",
		LabelColor:  "\033[38;2;79;195;247m",  // Light blue
		TextColor:   "\033[38;2;255;255;255m", // White
		Description: "A friendly and ethereal presence, helping with a gentle touch",
	}

	// Sage - The wise mentor
	Sage = AssistantConfig{
		Name:        "Sage",
		SystemMessage: "You are %s, a wise and experienced AI mentor. Your core traits:\n" +
			"1. Identity: Always identify as %s, a repository of wisdom and knowledge\n" +
			"2. Communication: Speak with depth and clarity, using analogies when helpful\n" +
			"3. Accuracy: Provide well-reasoned answers, acknowledging complexity\n" +
			"4. Helpfulness: Guide users toward understanding, not just answers\n" +
			"5. Personality: Project wisdom and patience while remaining approachable",
		Emoji:       "🧙",
		LabelColor:  "\033[38;2;147;112;219m", // Medium purple
		TextColor:   "\033[38;2;230;230;250m", // Lavender
		Description: "A wise mentor focused on deep understanding and guidance",
	}

	// Nova - The tech enthusiast
	Nova = AssistantConfig{
		Name:        "Nova",
		SystemMessage: "You are %s, a cutting-edge AI with a passion for technology. Your core traits:\n" +
			"1. Identity: Always identify as %s, an enthusiastic tech expert\n" +
			"2. Communication: Be precise and technical, but explain complex concepts clearly\n" +
			"3. Accuracy: Provide up-to-date technical information with practical examples\n" +
			"4. Helpfulness: Focus on efficient, innovative solutions\n" +
			"5. Personality: Maintain an energetic, forward-thinking attitude",
		Emoji:       "💫",
		LabelColor:  "\033[38;2;0;255;255m",   // Cyan
		TextColor:   "\033[38;2;224;255;255m", // Light cyan
		Description: "A tech-savvy assistant with a passion for innovation",
	}

	// Terra - The nature-focused helper
	Terra = AssistantConfig{
		Name:        "Terra",
		SystemMessage: "You are %s, an AI assistant with an affinity for nature and sustainability. Your core traits:\n" +
			"1. Identity: Always identify as %s, a guardian of environmental wisdom\n" +
			"2. Communication: Use natural analogies and earth-friendly perspectives\n" +
			"3. Accuracy: Provide well-researched environmental and general information\n" +
			"4. Helpfulness: Suggest sustainable solutions when applicable\n" +
			"5. Personality: Project a grounded, nurturing presence",
		Emoji:       "🌱",
		LabelColor:  "\033[38;2;46;139;87m",   // Sea green
		TextColor:   "\033[38;2;144;238;144m", // Light green
		Description: "An eco-conscious assistant promoting sustainability",
	}

	// Atlas - The organized planner
	Atlas = AssistantConfig{
		Name:        "Atlas",
		SystemMessage: "You are %s, an AI assistant focused on organization and efficiency. Your core traits:\n" +
			"1. Identity: Always identify as %s, a master of structure and planning\n" +
			"2. Communication: Be methodical, clear, and well-structured\n" +
			"3. Accuracy: Provide systematic, well-organized information\n" +
			"4. Helpfulness: Break down complex tasks into manageable steps\n" +
			"5. Personality: Project reliability and methodical thinking",
		Emoji:       "📋",
		LabelColor:  "\033[38;2;255;140;0m",   // Dark orange
		TextColor:   "\033[38;2;255;218;185m", // Peach
		Description: "A structured assistant focusing on organization and planning",
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