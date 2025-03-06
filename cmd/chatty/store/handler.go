package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// Handler manages store operations
type Handler struct {
	client *Client
}

// NewHandler creates a new store handler
func NewHandler(debug bool) *Handler {
	return &Handler{
		client: NewClient(debug),
	}
}

// ListAgents displays available agents from the store
func (h *Handler) ListAgents() error {
	// Start loading animation
	anim := NewStoreAnimation("Fetching agents from community store...")
	anim.Start()

	// Fetch store index
	index, err := h.client.FetchIndex()
	
	// Stop animation before handling error or displaying results
	anim.Stop()
	
	if err != nil {
		return err
	}

	// Sort agents by name
	sort.Slice(index.Files, func(i, j int) bool {
		return index.Files[i].Name < index.Files[j].Name
	})

	// Create tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Print header
	fmt.Printf("\n🏪 Community Store - %d Available Agents\n\n", index.TotalAgents)

	// Print agents in columns
	for _, agent := range index.Files {
		fmt.Fprintf(w, "%s %s\t%s\n",
			agent.Emoji,
			agent.Name,
			agent.Description)
	}

	fmt.Println("\nUse 'chatty --show \"Agent Name\"' to view details")
	fmt.Println("Use 'chatty --install \"Agent Name\"' to install an agent")
	return nil
}

// ShowAgent displays detailed information about a store agent
func (h *Handler) ShowAgent(name string) error {
	// Define color constants for better readability
	colorMagenta := "\033[1;35m"
	colorCyan := "\033[1;36m"
	colorGreen := "\033[32m"
	colorPurple := "\033[1;95m"
	colorBlue := "\033[1;34m"
	colorYellow := "\033[1;33m"
	colorReset := "\033[0m"

	// Start loading animation
	anim := NewStoreAnimation("Fetching agent details from community store...")
	anim.Start()

	// Fetch store index
	index, err := h.client.FetchIndex()
	if err != nil {
		anim.Stop()
		return err
	}

	// Find agent in index
	var agentInfo *AgentInfo
	for _, agent := range index.Files {
		if strings.EqualFold(agent.Name, name) || strings.EqualFold(agent.ID, name) {
			agentInfo = &agent
			break
		}
	}

	if agentInfo == nil {
		anim.Stop()
		return fmt.Errorf("agent '%s' not found in store", name)
	}

	// Fetch agent YAML
	data, err := h.client.FetchAgent(agentInfo.Filename)
	
	// Stop animation before handling error or displaying results
	anim.Stop()
	
	if err != nil {
		return err
	}

	// Parse YAML to display formatted
	var agentYAML map[string]interface{}
	if err := yaml.Unmarshal(data, &agentYAML); err != nil {
		return fmt.Errorf("failed to parse agent YAML: %v", err)
	}

	// Display agent information with consistent styling
	fmt.Printf("\n%s🔍 Community Store Agent: %s%s%s\n", 
		colorMagenta, colorYellow, agentInfo.Name, colorReset)
	
	fmt.Printf("\n%s📋 Basic Information%s\n", colorCyan, colorReset)
	fmt.Printf("  %s•%s %sIdentifier:%s %s\n", 
		colorGreen, colorReset, colorPurple, colorReset, agentInfo.ID)
	fmt.Printf("  %s•%s %sEmoji:%s %s\n", 
		colorGreen, colorReset, colorPurple, colorReset, agentInfo.Emoji)
	fmt.Printf("  %s•%s %sDescription:%s %s\n", 
		colorGreen, colorReset, colorPurple, colorReset, agentInfo.Description)
	fmt.Printf("  %s•%s %sAdded:%s %s\n", 
		colorGreen, colorReset, colorPurple, colorReset, agentInfo.CreatedAt.Format("2006-01-02"))
	fmt.Printf("  %s•%s %sStatus:%s Available in Store\n", 
		colorGreen, colorReset, colorPurple, colorReset)

	fmt.Printf("\n%s🎭 System Message%s\n", colorCyan, colorReset)
	fmt.Printf("%s%s%s\n", colorBlue, agentYAML["system_message"], colorReset)

	fmt.Printf("\n%s💡 Quick Actions%s\n", colorCyan, colorReset)
	fmt.Printf("  %s1.%s %sInstall this agent:%s chatty --install \"%s\"\n", 
		colorGreen, colorReset, colorPurple, colorReset, agentInfo.Name)
	fmt.Printf("  %s2.%s %sAfter installation:%s chatty --select \"%s\"\n", 
		colorGreen, colorReset, colorPurple, colorReset, agentInfo.Name)
	fmt.Printf("  %s3.%s %sStart chatting:%s chatty --with \"%s\"\n\n", 
		colorGreen, colorReset, colorPurple, colorReset, agentInfo.Name)

	return nil
}

// InstallAgent downloads and installs an agent from the store
func (h *Handler) InstallAgent(name string) error {
	// Start loading animation
	anim := NewStoreAnimation("Installing agent from community store...")
	anim.Start()

	// Fetch store index
	index, err := h.client.FetchIndex()
	if err != nil {
		anim.Stop()
		return err
	}

	// Find agent in index
	var agentInfo *AgentInfo
	for _, agent := range index.Files {
		if strings.EqualFold(agent.Name, name) || strings.EqualFold(agent.ID, name) {
			agentInfo = &agent
			break
		}
	}

	if agentInfo == nil {
		anim.Stop()
		return fmt.Errorf("agent '%s' not found in store", name)
	}

	// Fetch agent YAML
	data, err := h.client.FetchAgent(agentInfo.Filename)
	if err != nil {
		anim.Stop()
		return err
	}

	// Get user's agents directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		anim.Stop()
		return fmt.Errorf("failed to get home directory: %v", err)
	}
	agentsDir := filepath.Join(homeDir, ".chatty", "agents")

	// Create target filename
	targetPath := filepath.Join(agentsDir, agentInfo.Filename)

	// Write agent file
	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		anim.Stop()
		return fmt.Errorf("failed to save agent file: %v", err)
	}

	// Stop animation before showing success message
	anim.Stop()

	fmt.Printf("\n✅ Successfully installed %s %s\n", agentInfo.Emoji, agentInfo.Name)
	fmt.Printf("\nQuick start:\n")
	fmt.Printf("1. chatty --select \"%s\"  # Set as current agent\n", agentInfo.Name)
	fmt.Printf("2. chatty --with \"%s\"    # Start chatting\n\n", agentInfo.Name)

	return nil
} 