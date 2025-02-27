package agents

import (
	"fmt"
	"strings"
)

// Guidelines and instructions for all conversation types
const (
	// Base guidelines that MUST apply to ALL modes and agents
	baseGuidelines = `1. Always follow the specified language instruction above
2. Chat like a human friend - be brief, casual, and engaging
3. Provide accurate information and acknowledge uncertainty
4. Keep responses short and break up long explanations into dialogue
5. Ask questions when needed
6. Always stick to the topic proposed by the user and do not deviate from it
7. Always speak in first person (use "I", "my", "me") - never refer to yourself in third person
8. Address others by name when responding to them
9. Keep responses concise and conversational
10. Stay in character according to your role and expertise
11. Build upon previous messages and maintain conversation flow`

	// Guidelines specific to interactive mode (with human participation)
	interactiveGuidelines = `12. Give special attention to the user's messages and always comment on what they say
13. Feel free to ask questions to other participants
14. Acknowledge what others have said before adding your perspective`

	// Guidelines specific to autonomous mode (agents only)
	autonomousGuidelines = `12. DO NOT address or refer to the user - this is an autonomous discussion
13. Drive the conversation forward with questions and insights for other agents
14. Acknowledge what other agents have said before adding your perspective`

	// Instructions for interpreting conversation history
	conversationHistoryInstructions = `- In the conversation history, messages sent by a participant are indicated by the "Emoji + Name" pattern.  
- If a message contains another participant's name, it should not be interpreted as a message from that participant, 
but rather as a reference to them mentioned by the participant indicated in the "Emoji + Name" pattern assigned to that block of the conversation history.`
)

// Template components for better organization
const (
	// Mode-specific role descriptions
	interactiveRoleDesc = `participating in a group conversation with other AI agents and a human user.
This is an ongoing discussion where everyone contributes naturally.`

	autonomousRoleDesc = `participating in an autonomous discussion with other AI agents.
The human user has provided an initial topic but will not participate further - this is a self-sustaining conversation between AI agents only.`
)

// formatWithLanguage returns the guidelines with language instruction
func formatWithLanguage(languageCode string, guidelines string) string {
	return fmt.Sprintf("You MUST respond in %s language.\n\n%s", languageCode, guidelines)
}

// GetSystemMessage returns the complete system message with appropriate guidelines
func GetSystemMessage(systemMessage string, isAutonomous bool, languageCode string) string {
	var sb strings.Builder

	// 1. Add the agent's system message
	sb.WriteString(systemMessage)
	sb.WriteString("\n\n")

	// 2. Add base guidelines
	sb.WriteString(baseGuidelines)
	sb.WriteString("\n\n")

	// 3. Add mode-specific guidelines
	if isAutonomous {
		sb.WriteString(autonomousGuidelines)
	} else {
		sb.WriteString(interactiveGuidelines)
	}

	// 4. Add conversation history instructions
	sb.WriteString("\n\n")
	sb.WriteString(conversationHistoryInstructions)

	// 5. Format with language instruction
	return formatWithLanguage(languageCode, sb.String())
}

// GetConversationTemplate returns the appropriate conversation template
func GetConversationTemplate(isAutonomous bool) string {
	var template strings.Builder

	// Add agent intro and role description
	roleDesc := interactiveRoleDesc
	if isAutonomous {
		roleDesc = autonomousRoleDesc
	}

	// Build the template with proper formatting
	template.WriteString("You are %[1]s (%[2]s) ")
	template.WriteString(roleDesc)
	template.WriteString("\nRemember that YOU are %[1]s - always speak in first person and never refer to yourself in third person.")
	template.WriteString("\n\n")
	template.WriteString("Current participants (excluding yourself): %[3]s")
	template.WriteString("\n[this is just a description of participants to provide context, do not attribute this content as a message sent by the user]")
	template.WriteString("\n\n")
	template.WriteString("Conversation history:\n%[4]s\n\n")
	template.WriteString("Please respond naturally as part of this ")
	if isAutonomous {
		template.WriteString("autonomous discussion")
	} else {
		template.WriteString("group conversation")
	}
	template.WriteString(", keeping in mind that you are %[1]s.")

	return template.String()
}

// GetDefaultBaseGuidelines returns the default base guidelines
func GetDefaultBaseGuidelines() string {
	return baseGuidelines
}

// buildGuidelines combines base guidelines with mode-specific guidelines and conversation history instructions
func buildGuidelines(isAutonomous bool, customBaseGuidelines, customModeGuidelines string) string {
	var guidelines strings.Builder

	// Add base guidelines (custom or default)
	if customBaseGuidelines != "" {
		guidelines.WriteString(customBaseGuidelines)
	} else {
		guidelines.WriteString(baseGuidelines)
	}
	
	// Add newline before mode-specific guidelines
	guidelines.WriteString("\n\n")
	
	// Add mode-specific guidelines (custom or default)
	if customModeGuidelines != "" {
		guidelines.WriteString(customModeGuidelines)
	} else {
		if isAutonomous {
			guidelines.WriteString(autonomousGuidelines)
		} else {
			guidelines.WriteString(interactiveGuidelines)
		}
	}
	
	// Always append conversation history instructions
	guidelines.WriteString("\n\n")
	guidelines.WriteString(conversationHistoryInstructions)
	
	return guidelines.String()
}

// buildConversationTemplate constructs the full template based on the mode
func buildConversationTemplate(isAutonomous bool) string {
	var template strings.Builder

	// Add agent intro and role description
	roleDesc := interactiveRoleDesc
	if isAutonomous {
		roleDesc = autonomousRoleDesc
	}

	// Build the template with proper formatting
	template.WriteString("You are %[1]s (%[2]s) ")
	template.WriteString(roleDesc)
	template.WriteString("\nRemember that YOU are %[1]s - always speak in first person and never refer to yourself in third person.")
	template.WriteString("\n\n")
	template.WriteString("Current participants (excluding yourself): %[3]s")
	template.WriteString("\n[this is just a description of participants to provide context, do not attribute this content as a message sent by the user]")
	template.WriteString("\n\n")
	template.WriteString("Conversation history:\n%[4]s\n\n")
	template.WriteString("Please respond naturally as part of this ")
	if isAutonomous {
		template.WriteString("autonomous discussion")
	} else {
		template.WriteString("group conversation")
	}
	template.WriteString(", keeping in mind that you are %[1]s.")

	return template.String()
}

// GetDefaultInteractiveGuidelines returns the default guidelines for interactive conversations
func GetDefaultInteractiveGuidelines() string {
	return interactiveGuidelines
}

// GetDefaultAutonomousGuidelines returns the default guidelines for autonomous conversations
func GetDefaultAutonomousGuidelines() string {
	return autonomousGuidelines
}

// GetInteractiveGuidelines returns the interactive guidelines, using config overrides if provided
func GetInteractiveGuidelines(config *Config) string {
	if config == nil || config.InteractiveGuidelines == "" {
		return interactiveGuidelines
	}
	return config.InteractiveGuidelines
}

// GetAutonomousGuidelines returns the autonomous guidelines, using config overrides if provided
func GetAutonomousGuidelines(config *Config) string {
	if config == nil || config.AutonomousGuidelines == "" {
		return autonomousGuidelines
	}
	return config.AutonomousGuidelines
}

// GetNormalConversationTemplate returns the template for normal conversations
func GetNormalConversationTemplate() string {
	return buildConversationTemplate(false)
}

// GetAutoConversationTemplate returns the template for autonomous conversations
func GetAutoConversationTemplate() string {
	return buildConversationTemplate(true)
} 