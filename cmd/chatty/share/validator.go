package share

import (
	"fmt"
	"regexp"
	"strings"

	"chatty/cmd/chatty/agents"
	"chatty/cmd/chatty/store"
)

// Validator handles agent validation for sharing
type Validator struct {
	storeHandler *store.Handler
}

// NewValidator creates a new validator instance
func NewValidator(debug bool) *Validator {
	return &Validator{
		storeHandler: store.NewHandler(debug),
	}
}

// ValidateAgent performs all necessary validation checks on an agent
func (v *Validator) ValidateAgent(agent agents.AgentConfig) ValidationResult {
	result := ValidationResult{
		IsValid: true,
		Errors:  make([]string, 0),
		Warnings: make([]string, 0),
	}

	// Check if agent is built-in
	if agent.Source == "built-in" {
		result.IsValid = false
		result.Errors = append(result.Errors, "Cannot share built-in agents")
		return result
	}

	// Validate required fields
	if err := v.validateRequiredFields(agent, &result); err != nil {
		result.IsValid = false
		return result
	}

	// Check for name conflicts with store
	if err := v.checkNameConflicts(agent, &result); err != nil {
		result.Warnings = append(result.Warnings, 
			fmt.Sprintf("Could not check for name conflicts: %v", err))
	}

	// Validate content
	v.validateContent(agent, &result)

	return result
}

// validateRequiredFields checks that all required fields are present and valid
func (v *Validator) validateRequiredFields(agent agents.AgentConfig, result *ValidationResult) error {
	if strings.TrimSpace(agent.Name) == "" {
		result.Errors = append(result.Errors, "Agent name is required")
	}

	if strings.TrimSpace(agent.Description) == "" {
		result.Errors = append(result.Errors, "Agent description is required")
	}

	if strings.TrimSpace(agent.SystemMessage) == "" {
		result.Errors = append(result.Errors, "System message is required")
	}

	if strings.TrimSpace(agent.Emoji) == "" {
		result.Errors = append(result.Errors, "Emoji is required")
	}

	// Validate emoji is a single character
	if len([]rune(agent.Emoji)) != 1 {
		result.Errors = append(result.Errors, "Emoji must be a single character")
	}

	return nil
}

// checkNameConflicts checks if the agent name conflicts with existing store agents
func (v *Validator) checkNameConflicts(agent agents.AgentConfig, result *ValidationResult) error {
	// Get store index
	index, err := v.storeHandler.GetIndex()
	if err != nil {
		return err
	}

	// Check for exact matches
	for _, storeAgent := range index.Files {
		if strings.EqualFold(storeAgent.Name, agent.Name) {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("An agent with name '%s' already exists in the store", agent.Name))
			break
		}
	}

	return nil
}

// validateContent checks agent content for security and appropriateness
func (v *Validator) validateContent(agent agents.AgentConfig, result *ValidationResult) {
	// Check for potentially malicious content in system message
	if v.containsMaliciousContent(agent.SystemMessage) {
		result.Errors = append(result.Errors, 
			"System message contains potentially malicious content")
	}

	// Validate name format
	if !v.isValidName(agent.Name) {
		result.Errors = append(result.Errors,
			"Agent name can only contain letters, numbers, spaces, and basic punctuation")
	}

	// Check description length
	if len(agent.Description) > 100 {
		result.Warnings = append(result.Warnings,
			"Description is too long (max 100 characters)")
	}

	// Check system message length
	if len(agent.SystemMessage) > 2000 {
		result.Warnings = append(result.Warnings,
			"System message is very long (recommended max 2000 characters)")
	}
}

// containsMaliciousContent checks for potentially harmful content
func (v *Validator) containsMaliciousContent(content string) bool {
	// List of patterns to check for
	patterns := []string{
		`(?i)(rm|remove|del|delete)\s+(-rf?|/s)\s+.*`,  // File deletion commands
		`(?i)system\s*\(.*\)`,                          // System calls
		`(?i)exec\s*\(.*\)`,                           // Code execution
		`(?i)<script.*>.*</script>`,                   // Script tags
		`(?i)eval\s*\(.*\)`,                          // Eval functions
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, content)
		if matched {
			return true
		}
	}

	return false
}

// isValidName checks if the agent name contains only allowed characters
func (v *Validator) isValidName(name string) bool {
	// Allow letters, numbers, spaces, and basic punctuation
	pattern := `^[a-zA-Z0-9\s\-_.,!?'"\(\)]+$`
	matched, _ := regexp.MatchString(pattern, name)
	return matched
} 