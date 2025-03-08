package share

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"chatty/cmd/chatty/agents"
)

// Handler manages agent sharing operations
type Handler struct {
	config    ShareConfig
	validator *Validator
	debug     bool
}

// NewHandler creates a new share handler
func NewHandler(debug bool) *Handler {
	return &Handler{
		config:    DefaultShareConfig(),
		validator: NewValidator(debug),
		debug:     debug,
	}
}

// orderedAgentFields represents the desired order of fields in the YAML file
type orderedAgentFields struct {
	Name          string   `yaml:"name"`
	Author        string   `yaml:"author,omitempty"`
	SystemMessage string   `yaml:"system_message"`
	Emoji         string   `yaml:"emoji"`
	LabelColor    string   `yaml:"label_color"`
	TextColor     string   `yaml:"text_color"`
	Description   string   `yaml:"description"`
	Tags          []string `yaml:"tags"`
	IsDefault     bool     `yaml:"is_default"`
}

// ShareAgent handles the agent sharing process
func (h *Handler) ShareAgent(agentName string) error {
	// Define colors for output
	colorMagenta := "\033[1;35m"
	colorCyan := "\033[1;36m"
	colorGreen := "\033[32m"
	colorYellow := "\033[1;33m"
	colorRed := "\033[1;31m"
	colorBlue := "\033[1;34m"
	colorReset := "\033[0m"

	fmt.Printf("\n%s📝 Preparing to share %s with the community...%s\n\n",
		colorMagenta, agentName, colorReset)

	// Start validation animation
	anim := NewShareAnimation("Validating agent configuration...")
	anim.Start()

	// Get agent configuration
	agent := agents.GetAgentConfig(agentName)
	if agent.Name == agents.DefaultAgent.Name {
		anim.Stop()
		return fmt.Errorf("agent '%s' not found", agentName)
	}

	// Validate agent
	result := h.validator.ValidateAgent(agent)
	anim.Stop()

	// Print validation results
	fmt.Printf("%s1. Validating agent configuration...%s\n", colorCyan, colorReset)
	
	// Check if we have a name conflict error
	nameConflict := false
	for _, err := range result.Errors {
		if strings.Contains(err, "already exists in the store") {
			nameConflict = true
			break
		}
	}
	
	// Handle name conflict by allowing user to rename
	if nameConflict {
		fmt.Printf("\n%s⚠️ Name Conflict Detected%s\n", colorYellow, colorReset)
		fmt.Printf("An agent with the name '%s%s%s' already exists in the community store.\n", 
			colorCyan, agent.Name, colorReset)
		fmt.Printf("You must rename your agent to proceed with sharing.\n\n")
		
		// Prompt for a new name
		reader := bufio.NewReader(os.Stdin)
		originalName := agent.Name
		
		for {
			fmt.Printf("%sEnter a new name for your agent:%s ", colorYellow, colorReset)
			newName, _ := reader.ReadString('\n')
			newName = strings.TrimSpace(newName)
			
			if newName == "" {
				fmt.Printf("%s❌ Name cannot be empty. Please try again.%s\n", colorRed, colorReset)
				continue
			}
			
			// Check if the new name is valid format
			isValid, errorMsg := h.validator.isValidAgentName(newName)
			if !isValid {
				fmt.Printf("%s❌ %s%s\n\n", colorRed, errorMsg, colorReset)
				continue
			}
			
			// Check if the new name also exists in the store
			exists, err := h.validator.CheckStoreForDuplicateName(newName)
			if err != nil {
				fmt.Printf("%s⚠️ Warning: Could not check for duplicate names: %v%s\n", 
					colorYellow, err, colorReset)
			} else if exists {
				fmt.Printf("%s❌ The name '%s' also exists in the store. Please choose another name.%s\n\n", 
					colorRed, newName, colorReset)
				continue
			}
			
			// Update the agent name
			agent.Name = newName
			fmt.Printf("\n%s✓ Agent renamed to '%s'%s\n\n", colorGreen, newName, colorReset)
			
			// Re-validate the agent with the new name
			result = h.validator.ValidateAgent(agent)
			
			// If it's now valid, proceed; otherwise, show other errors
			if !result.IsValid {
				fmt.Printf("%s❌ Validation failed with other issues:%s\n", colorRed, colorReset)
				for _, err := range result.Errors {
					if !strings.Contains(err, "already exists in the store") {
						fmt.Printf("   - %s\n", err)
					}
				}
				return fmt.Errorf("validation failed")
			}
			
			// Prompt to save the agent locally with the new name
			fmt.Printf("%sWould you like to save the agent locally with the new name? [Y/n]:%s ", 
				colorCyan, colorReset)
			saveResponse, _ := reader.ReadString('\n')
			saveResponse = strings.ToLower(strings.TrimSpace(saveResponse))
			
			// Save locally unless the user specifically declines
			if saveResponse != "n" && saveResponse != "no" {
				if err := h.saveRenamedAgent(agent, originalName); err != nil {
					fmt.Printf("%s⚠️ Warning: Could not save renamed agent locally: %v%s\n", 
						colorYellow, err, colorReset)
				} else {
					fmt.Printf("%s✓ Agent saved locally with the new name%s\n\n", colorGreen, colorReset)
				}
			}
			
			break
		}
	} else if !result.IsValid {
		fmt.Printf("\n%s❌ Validation failed:%s\n", colorRed, colorReset)
		for _, err := range result.Errors {
			fmt.Printf("   - %s\n", err)
		}
		
		// If tags are missing, show specific guidance
		tagError := false
		for _, err := range result.Errors {
			if strings.Contains(err, "tag") {
				tagError = true
				break
			}
		}
		
		if tagError {
			fmt.Printf("\n%s📌 Tag Requirements:%s\n", colorYellow, colorReset)
			fmt.Printf("   - Each agent must have 1-5 tags\n")
			fmt.Printf("   - Tags can be added through the 'Edit tags' option in the agent editor\n")
			fmt.Printf("   - Run '%schatty --build%s' to create or edit an agent with tags\n\n", colorBlue, colorReset)
		}
		
		fmt.Println("\nPlease fix these issues and try again.")
		return fmt.Errorf("validation failed")
	}

	// If there were warnings but validation passed
	if len(result.Warnings) > 0 && result.IsValid {
		fmt.Printf("\n%s⚠️  Warnings:%s\n", colorYellow, colorReset)
		for _, warning := range result.Warnings {
			fmt.Printf("   - %s\n", warning)
		}
	}

	// If validation is successful
	if result.IsValid {
		fmt.Printf("   %s✓%s Required fields present\n", colorGreen, colorReset)
		fmt.Printf("   %s✓%s Format valid\n", colorGreen, colorReset)
		fmt.Printf("   %s✓%s Security checks passed\n", colorGreen, colorReset)
		fmt.Printf("   %s✓%s Tags valid (%d tags)\n\n", colorGreen, colorReset, len(agent.Tags))

		// Show tags in validation success
		if len(agent.Tags) > 0 {
			fmt.Printf("   %sTags:%s ", colorCyan, colorReset)
			for i, tag := range agent.Tags {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("%s%s%s", colorBlue, tag, colorReset)
			}
			fmt.Println("\n")
		}
	}

	// Collect author information
	fmt.Printf("%s2. Author Information%s\n", colorCyan, colorReset)
	fmt.Printf("   Please provide your name as you want it to appear in the agent's metadata.\n")
	fmt.Printf("   This will help the community know who created this agent.\n\n")
	fmt.Printf("%s❓ Author name:%s ", colorCyan, colorReset)
	
	reader := bufio.NewReader(os.Stdin)
	authorName, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error reading author name: %v", err)
	}
	authorName = strings.TrimSpace(authorName)
	
	if authorName == "" {
		return fmt.Errorf("author name is required")
	}

	// Show forking instructions
	originalRepoURL := h.config.BaseURL
	fmt.Printf("\n%s3. Repository Setup Required%s\n", colorCyan, colorReset)
	fmt.Printf("   Please fork this repository: %s%s%s\n\n", colorBlue, originalRepoURL, colorReset)
	
	fmt.Printf("%s📌 Follow these steps:%s\n", colorYellow, colorReset)
	fmt.Printf("   1. Visit the repository URL above\n")
	fmt.Printf("   2. Click the 'Fork' button in the top-right\n")
	fmt.Printf("   3. Wait for GitHub to complete the fork\n\n")

	// Ask for confirmation
	fmt.Printf("%s❓ Have you forked the repository? [y/N]:%s ", colorCyan, colorReset)
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" {
		return fmt.Errorf("please fork the repository before continuing")
	}

	// Ask for forked repository URL
	fmt.Printf("\n%s4. Enter your forked repository URL:%s\n", colorCyan, colorReset)
	fmt.Printf("   (Example: https://github.com/your-username/chatty-ai-community-store)\n")
	fmt.Printf("   URL: ")
	
	forkedRepoURL, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error reading input: %v", err)
	}
	forkedRepoURL = strings.TrimSpace(forkedRepoURL)

	if forkedRepoURL == "" || !strings.Contains(forkedRepoURL, "github.com") {
		return fmt.Errorf("invalid repository URL provided")
	}

	// Create ordered struct with fields in desired order
	orderedAgent := orderedAgentFields{
		Name:          agent.Name,
		Author:        authorName,
		SystemMessage: agent.SystemMessage,
		Emoji:         agent.Emoji,
		LabelColor:    agent.LabelColor,
		TextColor:     agent.TextColor,
		Description:   agent.Description,
		Tags:          agent.Tags,
		IsDefault:     false, // Always false for shared agents
	}

	// Marshal with ordered fields
	agentYAML, err := yaml.Marshal(orderedAgent)
	if err != nil {
		return fmt.Errorf("failed to marshal agent YAML: %v", err)
	}

	// Generate branch name safely
	safeAgentName := strings.ToLower(strings.ReplaceAll(agent.Name, " ", "-"))
	
	// Create branch name with date
	timestamp := time.Now().Format("20060102")
	branchName := fmt.Sprintf("agent-submission/%s-%s",
		safeAgentName,
		timestamp)

	// Build the new file URL in the user's fork
	// Create filename without timestamp suffix
	filenameSafe := strings.ToLower(strings.ReplaceAll(agent.Name, " ", "_"))
	newFileURL := fmt.Sprintf("%s/new/%s?filename=%s&value=%s&message=%s&branch=%s",
		forkedRepoURL,
		"main",
		url.QueryEscape(fmt.Sprintf("agents/%s.yaml", filenameSafe)),
		url.QueryEscape(string(agentYAML)),
		url.QueryEscape(fmt.Sprintf("Add new agent: %s", agent.Name)),
		url.QueryEscape(branchName))

	// Open browser with the new file URL
	fmt.Printf("\n%s5. Creating agent file...%s\n", colorCyan, colorReset)
	if err := h.openBrowser(newFileURL); err != nil {
		fmt.Printf("\n%s⚠️  Could not open browser automatically. Please open this URL manually:%s\n%s\n",
			colorYellow, colorReset, newFileURL)
	}

	// Show final instructions
	fmt.Printf("\n%s6. Final Instructions%s\n", colorCyan, colorReset)
	fmt.Printf("   - Please wait for the repository maintainers to review your submission.\n")
	fmt.Printf("   - Once approved, your agent will be added to the community store.\n")
	fmt.Printf("   - Thank you for contributing to the community!\n\n")

	return nil
}

// openBrowser opens the default browser with the given URL
func (h *Handler) openBrowser(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	return err
}

// saveRenamedAgent saves the agent with its new name to the user's agents directory
func (h *Handler) saveRenamedAgent(agent agents.AgentConfig, originalName string) error {
	// Get the home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}
	
	// Build the path to the agents directory
	agentsDir := filepath.Join(homeDir, ".chatty", "agents")
	
	// Create the new filename
	newFilename := fmt.Sprintf("%s.yaml", strings.ToLower(strings.ReplaceAll(agent.Name, " ", "_")))
	newFilePath := filepath.Join(agentsDir, newFilename)
	
	// Marshal the agent to YAML
	agentYAML, err := yaml.Marshal(agent)
	if err != nil {
		return fmt.Errorf("failed to marshal agent YAML: %v", err)
	}
	
	// Write to the new file
	if err := os.WriteFile(newFilePath, agentYAML, 0644); err != nil {
		return fmt.Errorf("failed to write renamed agent file: %v", err)
	}
	
	// Find and delete the original file
	originalFilename := fmt.Sprintf("%s.yaml", strings.ToLower(strings.ReplaceAll(originalName, " ", "_")))
	originalFilePath := filepath.Join(agentsDir, originalFilename)
	
	// Delete the original file (ignore errors if it doesn't exist)
	_ = os.Remove(originalFilePath)
	
	// Make sure the agents are reloaded
	agents.LoadAgents()
	
	return nil
}

// generatePRURL creates the GitHub PR URL with all necessary parameters
func (h *Handler) generatePRURL(agent agents.AgentConfig) (string, error) {
	// Convert agent to YAML
	agentYAML, err := yaml.Marshal(agent)
	if err != nil {
		return "", fmt.Errorf("failed to marshal agent YAML: %v", err)
	}

	// Create branch name with date (but not for filename)
	safeAgentName := strings.ToLower(strings.ReplaceAll(agent.Name, " ", "-"))
	timestamp := time.Now().Format("20060102")
	branchName := fmt.Sprintf(h.config.BranchName, safeAgentName, timestamp)

	// Create commit message
	commitMsg := fmt.Sprintf(h.config.CommitMsg, agent.Name)

	// Format tags for PR description
	tagsText := "- " + strings.Join(agent.Tags, "\n- ")

	// Create PR description
	prBody := fmt.Sprintf(h.config.PRTemplate, agent.Description, tagsText, string(agentYAML))

	// Build the URL - using the fork-based workflow
	baseURL := fmt.Sprintf("%s/fork", h.config.BaseURL)
	params := url.Values{}
	
	// Create filename without timestamp suffix
	safeFilename := strings.ToLower(strings.ReplaceAll(agent.Name, " ", "_"))
	filename := fmt.Sprintf("agents/%s.yaml", safeFilename)
	
	// After fork, redirect to new file creation
	params.Add("quick_pull", "1")
	params.Add("filename", filename)
	params.Add("value", string(agentYAML))
	params.Add("message", commitMsg)
	params.Add("description", prBody)
	params.Add("branch", branchName)

	return fmt.Sprintf("%s?%s", baseURL, params.Encode()), nil
} 