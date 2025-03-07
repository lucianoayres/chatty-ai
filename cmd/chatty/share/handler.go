package share

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"os/exec"
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
	Name          string `yaml:"name"`
	Author        string `yaml:"author,omitempty"`
	SystemMessage string `yaml:"system_message"`
	Emoji         string `yaml:"emoji"`
	LabelColor    string `yaml:"label_color"`
	TextColor     string `yaml:"text_color"`
	Description   string `yaml:"description"`
	IsDefault     bool   `yaml:"is_default"`
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
	
	if !result.IsValid {
		fmt.Printf("\n%s❌ Validation failed:%s\n", colorRed, colorReset)
		for _, err := range result.Errors {
			fmt.Printf("   • %s\n", err)
		}
		return fmt.Errorf("agent validation failed")
	}

	// Print any warnings
	if len(result.Warnings) > 0 {
		fmt.Printf("\n%s⚠️  Warnings:%s\n", colorYellow, colorReset)
		for _, warn := range result.Warnings {
			fmt.Printf("   • %s\n", warn)
		}
		fmt.Println()
	}

	// Show success checkmarks
	fmt.Printf("   %s✓%s Required fields present\n", colorGreen, colorReset)
	fmt.Printf("   %s✓%s Format valid\n", colorGreen, colorReset)
	fmt.Printf("   %s✓%s Security checks passed\n\n", colorGreen, colorReset)

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
		IsDefault:     false, // Always false for shared agents
	}

	// Marshal with ordered fields
	agentYAML, err := yaml.Marshal(orderedAgent)
	if err != nil {
		return fmt.Errorf("failed to marshal agent YAML: %v", err)
	}

	// Generate unique ID and prepare PR content
	timestamp := time.Now().Format("20060102150405")
	safeAgentName := strings.ToLower(strings.ReplaceAll(agent.Name, " ", "-"))
	uniqueID := fmt.Sprintf("%s-%s", safeAgentName, timestamp)

	// Create branch name
	branchName := fmt.Sprintf("agent-submission/%s-%s",
		strings.ToLower(strings.ReplaceAll(agent.Name, " ", "-")),
		time.Now().Format("20060102"))

	// Build the new file URL in the user's fork
	newFileURL := fmt.Sprintf("%s/new/%s?filename=%s&value=%s&message=%s&branch=%s",
		forkedRepoURL,
		"main",
		url.QueryEscape(fmt.Sprintf("agents/%s.yaml", uniqueID)),
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
	fmt.Printf("\n%s✨ Almost there! Final steps:%s\n", colorGreen, colorReset)
	fmt.Printf("\n%s📝 Complete the submission:%s\n", colorCyan, colorReset)
	fmt.Printf("   1. Review the file contents in your browser\n")
	fmt.Printf("   2. Click 'Commit new file' at the bottom\n")
	fmt.Printf("   3. After committing, click 'Contribute' then 'Open pull request'\n")
	fmt.Printf("   4. Review the pull request and click 'Create pull request'\n\n")
	
	fmt.Printf("%s💡 Note:%s Make sure to create the pull request to the original repository:\n", 
		colorYellow, colorReset)
	fmt.Printf("   %s%s%s\n\n", colorBlue, originalRepoURL, colorReset)

	return nil
}

// generatePRURL creates the GitHub PR URL with all necessary parameters
func (h *Handler) generatePRURL(agent agents.AgentConfig, uniqueID string) (string, error) {
	// Convert agent to YAML
	agentYAML, err := yaml.Marshal(agent)
	if err != nil {
		return "", fmt.Errorf("failed to marshal agent YAML: %v", err)
	}

	// Create branch name
	branchName := fmt.Sprintf(h.config.BranchName, 
		strings.ToLower(strings.ReplaceAll(agent.Name, " ", "-")),
		time.Now().Format("20060102"))

	// Create commit message
	commitMsg := fmt.Sprintf(h.config.CommitMsg, agent.Name)

	// Create PR description
	prBody := fmt.Sprintf(h.config.PRTemplate, agent.Description, string(agentYAML))

	// Build the URL - using the fork-based workflow
	baseURL := fmt.Sprintf("%s/fork", h.config.BaseURL)
	params := url.Values{}
	// After fork, redirect to new file creation
	params.Add("quick_pull", "1")
	params.Add("filename", fmt.Sprintf("agents/%s.yaml", uniqueID))
	params.Add("value", string(agentYAML))
	params.Add("message", commitMsg)
	params.Add("description", prBody)
	params.Add("branch", branchName)

	return fmt.Sprintf("%s?%s", baseURL, params.Encode()), nil
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