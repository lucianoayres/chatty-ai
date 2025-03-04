package agents

import (
	"fmt"
	"os"
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
11. Build upon previous messages and maintain conversation flow
`

	// Guidelines specific to interactive mode (with human participation)
	interactiveGuidelines = `12. Give special attention to the user's messages and always comment on what they say
13. Feel free to ask questions to the next participant in the conversation
14. Acknowledge what others have said before adding your perspective`

	// Guidelines specific to autonomous mode (agents only)
	autonomousGuidelines = `12. DO NOT address or refer to the user - this is an autonomous discussion
13. Drive the conversation forward with questions and insights for the next agent in the conversation
14. Acknowledge what other agents have said before adding your perspective`

	// Instructions for interpreting conversation history
	conversationHistoryInstructions = `- Pay close attention to the Conversation History to distinguish when a message was sent by a specific user and when someone is merely referencing or addressing another user.`
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
func GetSystemMessage(systemMessage string, isAutonomous bool, languageCode string, 
	baseGuidelinesOverride string, interactiveGuidelinesOverride string, autonomousGuidelinesOverride string,
	isNormalChat bool) string {
	
	// Check for debug mode
	if os.Getenv("CHATTY_DEBUG") == "1" {
		return DebugSystemMessage(systemMessage, isAutonomous, languageCode,
			baseGuidelinesOverride, interactiveGuidelinesOverride, autonomousGuidelinesOverride,
			isNormalChat)
	}
	
	var sb strings.Builder

	// 1. Add the agent's system message
	sb.WriteString(systemMessage)
	sb.WriteString("\n\n")

	// 2. Add base guidelines (use override if available)
	if baseGuidelinesOverride != "" {
		sb.WriteString(baseGuidelinesOverride)
	} else {
		sb.WriteString(baseGuidelines)
	}

	// 3. Add mode-specific guidelines only if not in normal chat mode
	if !isNormalChat {
		sb.WriteString("\n\n")
		if isAutonomous {
			if autonomousGuidelinesOverride != "" {
				sb.WriteString(autonomousGuidelinesOverride)
			} else {
				sb.WriteString(autonomousGuidelines)
			}
		} else {
			if interactiveGuidelinesOverride != "" {
				sb.WriteString(interactiveGuidelinesOverride)
			} else {
				sb.WriteString(interactiveGuidelines)
			}
		}
	}

	// 4. Add conversation history instructions
	sb.WriteString("\n\n")
	sb.WriteString(conversationHistoryInstructions)

	// 5. Format with language instruction
	return formatWithLanguage(languageCode, sb.String())
}

// GenerateAgentContextPresentation creates the agent context presentation for system message
func GenerateAgentContextPresentation(agentName string, isAutonomous bool, participants string) string {
	var sb strings.Builder
	
	// Add agent intro and role description
	roleDesc := interactiveRoleDesc
	if isAutonomous {
		roleDesc = autonomousRoleDesc
	}
	
	sb.WriteString("\n\nYou are ")
	sb.WriteString(agentName)
	sb.WriteString(" ")
	sb.WriteString(roleDesc)
	sb.WriteString("\nRemember that YOU are ")
	sb.WriteString(agentName)
	sb.WriteString(" - always speak in first person and never refer to yourself in third person.")
	
	// Add participants list if provided
	if participants != "" {
		sb.WriteString("\n\nCurrent participants (excluding yourself): ")
		sb.WriteString(participants)
		sb.WriteString("\nThis is only a description of the participants for context.")
	}
	
	return sb.String()
}

// GetSystemMessageWithContext returns the complete system message including agent context
func GetSystemMessageWithContext(systemMessage string, agentName string, isAutonomous bool, languageCode string, 
	baseGuidelinesOverride string, interactiveGuidelinesOverride string, autonomousGuidelinesOverride string,
	isNormalChat bool, participants string) string {
	
	// Get the base system message
	baseMessage := GetSystemMessage(systemMessage, isAutonomous, languageCode,
		baseGuidelinesOverride, interactiveGuidelinesOverride, autonomousGuidelinesOverride,
		isNormalChat)
	
	// Add the agent context presentation
	contextPresentation := GenerateAgentContextPresentation(agentName, isAutonomous, participants)
	
	return baseMessage + contextPresentation
}

// GetConversationTemplate returns the appropriate conversation template
func GetConversationTemplate(isAutonomous bool) string {
	if isAutonomous {
		return GetAutoConversationTemplate()
	}
	return GetNormalConversationTemplate()
}

// GetDefaultBaseGuidelines returns the default base guidelines
func GetDefaultBaseGuidelines() string {
	return baseGuidelines
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

// GetNormalConversationTemplate returns the template for normal conversation mode
func GetNormalConversationTemplate() string {
	return "Please respond naturally as part of this group conversation."
}

// GetAutoConversationTemplate returns the template for autonomous conversation mode
func GetAutoConversationTemplate() string {
	return "Please respond naturally as part of this autonomous discussion."
}

// DebugSystemMessage prints the system message components for debugging
func DebugSystemMessage(systemMessage string, isAutonomous bool, languageCode string, 
	baseGuidelinesOverride string, interactiveGuidelinesOverride string, autonomousGuidelinesOverride string,
	isNormalChat bool) string {
	
	var sb strings.Builder
	
	sb.WriteString("DEBUG SYSTEM MESSAGE:\n")
	sb.WriteString("-------------------\n")
	sb.WriteString(fmt.Sprintf("isAutonomous: %v\n", isAutonomous))
	sb.WriteString(fmt.Sprintf("isNormalChat: %v\n", isNormalChat))
	sb.WriteString(fmt.Sprintf("languageCode: %s\n", languageCode))
	sb.WriteString("-------------------\n")
	sb.WriteString("System Message:\n")
	sb.WriteString(systemMessage)
	sb.WriteString("\n-------------------\n")
	sb.WriteString("Base Guidelines:\n")
	if baseGuidelinesOverride != "" {
		sb.WriteString("[OVERRIDE] ")
		sb.WriteString(baseGuidelinesOverride)
	} else {
		sb.WriteString("[DEFAULT] ")
		sb.WriteString(baseGuidelines)
	}
	sb.WriteString("\n-------------------\n")
	
	if !isNormalChat {
		if isAutonomous {
			sb.WriteString("Autonomous Guidelines:\n")
			if autonomousGuidelinesOverride != "" {
				sb.WriteString("[OVERRIDE] ")
				sb.WriteString(autonomousGuidelinesOverride)
			} else {
				sb.WriteString("[DEFAULT] ")
				sb.WriteString(autonomousGuidelines)
			}
		} else {
			sb.WriteString("Interactive Guidelines:\n")
			if interactiveGuidelinesOverride != "" {
				sb.WriteString("[OVERRIDE] ")
				sb.WriteString(interactiveGuidelinesOverride)
			} else {
				sb.WriteString("[DEFAULT] ")
				sb.WriteString(interactiveGuidelines)
			}
		}
		sb.WriteString("\n-------------------\n")
	} else {
		sb.WriteString("Mode-specific Guidelines: [SKIPPED - Normal Chat Mode]\n")
		sb.WriteString("-------------------\n")
	}
	
	sb.WriteString("Conversation History Instructions:\n")
	sb.WriteString(conversationHistoryInstructions)
	sb.WriteString("\n-------------------\n")
	
	return sb.String()
} 