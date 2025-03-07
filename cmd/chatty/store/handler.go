package store

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
	// Define color constants for better readability
	colorMagenta := "\033[1;35m"
	colorCyan := "\033[1;36m"
	colorGreen := "\033[32m"
	colorPurple := "\033[1;95m"
	colorBlue := "\033[1;34m"
	colorYellow := "\033[1;33m"
	colorGray := "\033[1;30m"
	colorReset := "\033[0m"

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

	// Print store header
	fmt.Printf("\n%s🏪 Community Store%s\n", colorMagenta, colorReset)
	fmt.Printf("%s%s%s\n", colorMagenta, strings.Repeat("═", 50), colorReset)
	fmt.Printf("%s%d Agents Available%s\n\n", colorYellow, index.TotalAgents, colorReset)

	// Create tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print agents in a clean, organized layout
	fmt.Printf("%sAvailable Agents%s\n", colorCyan, colorReset)
	fmt.Printf("%s%s%s\n", colorGray, strings.Repeat("─", 50), colorReset)

	for _, agent := range index.Files {
		// Print each agent with emoji, name, and description
		fmt.Fprintf(w, "  %s %s\t%s\n",
			agent.Emoji,
			agent.Name,
			agent.Description)
	}
	w.Flush()
	fmt.Println()

	// Print help section
	fmt.Printf("%s💡 Quick Actions%s\n", colorCyan, colorReset)
	fmt.Printf("%s%s%s\n", colorGray, strings.Repeat("─", 50), colorReset)
	fmt.Printf("  %s1.%s %sView agent details:%s chatty --show %s\"Agent Name\"%s\n",
		colorGreen, colorReset, colorPurple, colorReset, colorBlue, colorReset)
	fmt.Printf("  %s2.%s %sInstall an agent:%s chatty --install %s\"Agent Name\"%s\n",
		colorGreen, colorReset, colorPurple, colorReset, colorBlue, colorReset)

	// Print tips
	fmt.Printf("\n%s💭 Tips%s\n", colorCyan, colorReset)
	fmt.Printf("%s%s%s\n", colorGray, strings.Repeat("─", 50), colorReset)
	fmt.Printf("  • Use %s--show%s to view full agent capabilities and system messages\n",
		colorPurple, colorReset)
	fmt.Printf("  • Installed agents appear in %s--list%s under 'Custom & Community Agents'\n",
		colorPurple, colorReset)
	fmt.Printf("  • You can combine agents in group chats: %schatty --with \"Agent1,Agent2\"%s\n",
		colorBlue, colorReset)
	fmt.Printf("  • Store is updated regularly with new agents\n\n")

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
	
	// Display author if present
	if author, ok := agentYAML["author"].(string); ok && author != "" {
		fmt.Printf("  %s•%s %sAuthor:%s %s\n", 
			colorGreen, colorReset, colorPurple, colorReset, author)
	}
	
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
	// Define color constants for better readability
	colorMagenta := "\033[1;35m"
	colorCyan := "\033[1;36m"
	colorGreen := "\033[32m"
	colorPurple := "\033[1;95m"
	colorBlue := "\033[1;34m"
	colorYellow := "\033[1;33m"
	colorReset := "\033[0m"

	// Start loading animation
	anim := NewStoreAnimation("Checking agent status...")
	anim.Start()

	// First check if the agent is a built-in agent
	_, filename, _, _ := runtime.Caller(0)
	builtinPath := filepath.Join(filepath.Dir(filename), "builtin")
	
	files, err := os.ReadDir(builtinPath)
	if err == nil { // Only check if we can read the directory
		for _, file := range files {
			if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
				path := filepath.Join(builtinPath, file.Name())
				data, err := os.ReadFile(path)
				if err != nil {
					continue
				}

				var agentConfig struct {
					Name string `yaml:"name"`
				}
				if err := yaml.Unmarshal(data, &agentConfig); err != nil {
					continue
				}

				if strings.EqualFold(agentConfig.Name, name) {
					anim.Stop()
					fmt.Printf("\n%s📝 Note:%s %s%s%s is a built-in agent and is already available\n\n", 
						colorYellow, colorReset,
						colorMagenta, name, colorReset)
					fmt.Printf("%s💡 Quick Actions:%s\n", colorCyan, colorReset)
					fmt.Printf("  %s1.%s %sStart chatting:%s chatty --with %s\"%s\"%s\n",
						colorGreen, colorReset, colorPurple, colorReset, colorBlue, name, colorReset)
					fmt.Printf("  %s2.%s %sSet as default:%s chatty --select %s\"%s\"%s\n",
						colorGreen, colorReset, colorPurple, colorReset, colorBlue, name, colorReset)
					fmt.Printf("  %s3.%s %sView details:%s chatty --show %s\"%s\"%s\n\n",
						colorGreen, colorReset, colorPurple, colorReset, colorBlue, name, colorReset)
					return nil
				}
			}
		}
	}

	// Then check if it's already installed as a custom agent
	homeDir, err := os.UserHomeDir()
	if err != nil {
		anim.Stop()
		return fmt.Errorf("failed to get home directory: %v", err)
	}
	agentsDir := filepath.Join(homeDir, ".chatty", "agents")

	// Read all installed agent files
	files, err = os.ReadDir(agentsDir)
	if err == nil { // Only check if we can read the directory
		for _, file := range files {
			if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
				path := filepath.Join(agentsDir, file.Name())
				data, err := os.ReadFile(path)
				if err != nil {
					continue
				}

				var agentConfig struct {
					Name string `yaml:"name"`
				}
				if err := yaml.Unmarshal(data, &agentConfig); err != nil {
					continue
				}

				if strings.EqualFold(agentConfig.Name, name) {
					anim.Stop()
					fmt.Printf("\n%s📝 Note:%s Agent %s%s%s is already installed\n\n", 
						colorYellow, colorReset,
						colorMagenta, name, colorReset)
					fmt.Printf("%s💡 Quick Actions:%s\n", colorCyan, colorReset)
					fmt.Printf("  %s1.%s %sStart chatting:%s chatty --with %s\"%s\"%s\n",
						colorGreen, colorReset, colorPurple, colorReset, colorBlue, name, colorReset)
					fmt.Printf("  %s2.%s %sSet as default:%s chatty --select %s\"%s\"%s\n",
						colorGreen, colorReset, colorPurple, colorReset, colorBlue, name, colorReset)
					fmt.Printf("  %s3.%s %sView details:%s chatty --show %s\"%s\"%s\n\n",
						colorGreen, colorReset, colorPurple, colorReset, colorBlue, name, colorReset)
					return nil
				}
			}
		}
	}

	// Change animation message for store check
	anim.Stop()
	anim = NewStoreAnimation("Fetching agent from community store...")
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

	// Create target filename
	targetPath := filepath.Join(agentsDir, agentInfo.Filename)

	// Write agent file
	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		anim.Stop()
		return fmt.Errorf("failed to save agent file: %v", err)
	}

	// Stop animation before showing success message
	anim.Stop()

	fmt.Printf("\n%s✅ Successfully installed %s %s%s\n", 
		colorGreen, agentInfo.Emoji, agentInfo.Name, colorReset)
	fmt.Printf("\n%s💡 Quick Actions:%s\n", colorCyan, colorReset)
	fmt.Printf("  %s1.%s %sSet as current agent:%s chatty --select %s\"%s\"%s\n",
		colorGreen, colorReset, colorPurple, colorReset, colorBlue, agentInfo.Name, colorReset)
	fmt.Printf("  %s2.%s %sStart chatting:%s chatty --with %s\"%s\"%s\n\n",
		colorGreen, colorReset, colorPurple, colorReset, colorBlue, agentInfo.Name, colorReset)

	return nil
}

// GetIndex retrieves the store index
func (h *Handler) GetIndex() (*StoreIndex, error) {
	// Start loading animation
	anim := NewStoreAnimation("Fetching store index...")
	anim.Start()

	// Fetch store index
	index, err := h.client.FetchIndex()
	
	// Stop animation before handling error or returning
	anim.Stop()
	
	if err != nil {
		return nil, err
	}

	return index, nil
} 