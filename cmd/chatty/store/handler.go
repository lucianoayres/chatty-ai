package store

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Handler implements store operations
type Handler struct {
	client       *Client
	storeConfig  *StoreConfig
	tagsConfig   *TagsConfig
	debug        bool
}

// NewHandler creates a new store handler
func NewHandler(debug bool) *Handler {
	return &Handler{
		client: NewClient(debug),
		debug:  debug,
	}
}

// LoadStoreConfigs loads store configurations if not already loaded
func (h *Handler) LoadStoreConfigs() {
	// Only load if not already loaded
	if h.storeConfig == nil {
		storeConfig, err := h.client.FetchStoreConfig()
		if err != nil && h.debug {
			fmt.Printf("Warning: Failed to load store config: %v\n", err)
		}
		h.storeConfig = storeConfig
	}

	if h.tagsConfig == nil {
		tagsConfig, err := h.client.FetchTagsConfig()
		if err != nil && h.debug {
			fmt.Printf("Warning: Failed to load tags config: %v\n", err)
		}
		h.tagsConfig = tagsConfig
	}
}

// categorizeAgents groups agents by their categories based on tags
func (h *Handler) categorizeAgents(agents []AgentInfo) map[string][]AgentInfo {
	categorized := make(map[string][]AgentInfo)
	
	// Always include "All Agents" category
	categorized["All Agents"] = agents
	
	// If no store config is available, return just the "All Agents" category
	if h.storeConfig == nil {
		return categorized
	}
	
	// Process "New Arrivals" category if enabled
	if h.storeConfig.StorefrontSettings.NewItemsConfig.Enabled {
		newAgents := h.filterNewAgents(agents)
		if len(newAgents) > 0 {
			categorized["New Arrivals"] = newAgents
		}
	}
	
	// Process "Featured" category if enabled
	if h.storeConfig.StorefrontSettings.FeaturedItemsConfig.Enabled {
		// Filter agents with "featured" tag
		var featuredAgents []AgentInfo
		for _, agent := range agents {
			for _, tag := range agent.Tags {
				if tag == "featured" {
					featuredAgents = append(featuredAgents, agent)
					break
				}
			}
			
			// Respect the max items limit
			if len(featuredAgents) >= h.storeConfig.StorefrontSettings.FeaturedItemsConfig.MaxItems {
				break
			}
		}
		
		if len(featuredAgents) > 0 {
			categorized["Featured"] = featuredAgents
		}
	}
	
	// Process configured categories from store settings
	for _, category := range h.storeConfig.StorefrontSettings.Categories {
		// Skip "New Arrivals" and "Featured" as they are handled separately
		if category.Name == "New Arrivals" || category.Name == "Featured" {
			continue
		}
		
		categoryAgents := h.filterAgentsByTags(agents, category.Tags)
		if len(categoryAgents) > 0 {
			categorized[category.Name] = categoryAgents
		}
	}
	
	// Add "Uncategorized" for agents without tags
	uncategorized := h.filterUncategorizedAgents(agents)
	if len(uncategorized) > 0 {
		categorized["Uncategorized"] = uncategorized
	}
	
	return categorized
}

// filterNewAgents returns agents added within the configured time window
func (h *Handler) filterNewAgents(agents []AgentInfo) []AgentInfo {
	if h.storeConfig == nil {
		return nil
	}
	
	config := h.storeConfig.StorefrontSettings.NewItemsConfig
	cutoffDate := time.Now().AddDate(0, 0, -config.TimeWindowDays)
	
	var newAgents []AgentInfo
	for _, agent := range agents {
		if agent.CreatedAt.After(cutoffDate) {
			newAgents = append(newAgents, agent)
		}
		
		// Respect the max items limit
		if len(newAgents) >= config.MaxItems {
			break
		}
	}
	
	return newAgents
}

// filterAgentsByTags returns agents that have any of the specified tags
func (h *Handler) filterAgentsByTags(agents []AgentInfo, filterTags []string) []AgentInfo {
	if len(filterTags) == 0 {
		return nil
	}
	
	var filtered []AgentInfo
	for _, agent := range agents {
		// Check if the agent has any of the filter tags
		for _, filterTag := range filterTags {
			hasTag := false
			for _, agentTag := range agent.Tags {
				if agentTag == filterTag {
					hasTag = true
					break
				}
			}
			
			if hasTag {
				filtered = append(filtered, agent)
				break  // Break after adding the agent once
			}
		}
	}
	
	return filtered
}

// filterUncategorizedAgents returns agents that don't have any tags
func (h *Handler) filterUncategorizedAgents(agents []AgentInfo) []AgentInfo {
	var uncategorized []AgentInfo
	for _, agent := range agents {
		if len(agent.Tags) == 0 {
			uncategorized = append(uncategorized, agent)
		}
	}
	return uncategorized
}

// getCategoryDescription returns the description for a category
func (h *Handler) getCategoryDescription(categoryName string) string {
	if h.storeConfig == nil {
		return ""
	}
	
	// Handle special categories
	if categoryName == "All Agents" {
		return "Complete list of all available agents"
	}
	
	if categoryName == "Uncategorized" {
		return "Agents without specific categorization"
	}
	
	// Special handling for "Featured" category
	if strings.EqualFold(categoryName, "Featured") {
		for _, category := range h.storeConfig.StorefrontSettings.Categories {
			if strings.EqualFold(category.Name, "Featured") {
				return category.Description
			}
		}
		return "Handpicked agents recommended for all users"
	}
	
	// Look up from configured categories with case-insensitive matching
	for _, category := range h.storeConfig.StorefrontSettings.Categories {
		if strings.EqualFold(category.Name, categoryName) {
			return category.Description
		}
	}
	
	return ""
}

// ListAgents displays all agents in the store, organized by categories
func (h *Handler) ListAgents() error {
	// Define color constants for better readability
	colorMagenta := "\033[1;35m"
	colorCyan := "\033[1;36m"
	colorYellow := "\033[1;33m" // Bright yellow for better visibility
	colorWhite := "\033[1;37m"  // Bright white for better visibility
	colorReset := "\033[0m"
	colorGray := "\033[1;37m"   // Changed from dark gray to light gray for better readability
	
	// Start loading animation
	anim := NewStoreAnimation("Fetching community store data...")
	anim.Start()
	
	// Fetch index
	index, err := h.client.FetchIndex()
	if err != nil {
		anim.Stop()
		return err
	}
	
	// Load store configurations
	h.LoadStoreConfigs()
	
	// Stop animation before displaying results
	anim.Stop()
	
	// Display header
	fmt.Printf("\n%s Community Store (%d agents available)%s\n", 
		colorMagenta, index.TotalAgents, colorReset)
	fmt.Printf("%s%s%s\n", 
		colorMagenta, strings.Repeat("━", 50), colorReset)
	
	// Categorize agents
	categorizedAgents := h.categorizeAgents(index.Files)
	
	// Get category names and sort them
	var categoryNames []string
	for category := range categorizedAgents {
		categoryNames = append(categoryNames, category)
	}
	
	// Apply custom sorting to categories
	sort.Slice(categoryNames, func(i, j int) bool {
		categoryPriorities := make(map[string]int)
		
		// Apply category priorities from configuration file order
		if h.storeConfig != nil {
			for i, category := range h.storeConfig.StorefrontSettings.Categories {
				categoryPriorities[category.Name] = i + 1
			}
		}
		
		// Special handling for built-in categories
		categoryPriorities["All Agents"] = 999
		categoryPriorities["Uncategorized"] = 998
		
		// Get priorities for the two categories being compared
		priorityI := categoryPriorities[categoryNames[i]]
		priorityJ := categoryPriorities[categoryNames[j]]
		
		// If both have defined priorities, use them
		if priorityI > 0 && priorityJ > 0 {
			return priorityI < priorityJ
		}
		
		// If only one has a defined priority, it comes first
		if priorityI > 0 {
			return true
		}
		if priorityJ > 0 {
			return false
		}
		
		// Default to alphabetical sorting
		return categoryNames[i] < categoryNames[j]
	})
	
	// Display agents by category
	for _, category := range categoryNames {
		agents := categorizedAgents[category]
		
		// Skip empty categories
		if len(agents) > 0 {
			// Get category description
			description := h.getCategoryDescription(category)
			
			// Display category header
			fmt.Printf("\n%s📂 %s%s%s (%d)%s\n", 
				colorCyan, colorCyan, category, colorReset, len(agents), colorReset)
			
			// Display category description if available
			if description != "" {
				fmt.Printf("%s%s%s\n", colorGray, description, colorReset)
			}
			
			// Print a separator line under the category header
			fmt.Printf("%s%s%s\n", colorCyan, strings.Repeat("━", 50), colorReset)
			
			// Display agents in this category
			for _, agent := range agents {
				// Print each agent with emoji, name, and description
				fmt.Printf("  %s %s%s%s%s %s%s%s\n",
					agent.Emoji,
					colorYellow,
					agent.Name,
					colorReset,
					h.formatAuthor(agent.Author),
					colorWhite,
					agent.Description,
					colorReset)
			}
		}
	}
	
	// Display help section
	fmt.Printf("\n%s💡 Quick Actions%s\n", colorCyan, colorReset)
	fmt.Printf("%s%s%s\n", colorCyan, strings.Repeat("━", 50), colorReset)
	fmt.Printf("   %s1.%s %sView agent details:%s chatty --show %s\"Agent Name\"%s\n", 
		colorYellow, colorReset, colorYellow, colorReset, colorYellow, colorReset)
	fmt.Printf("   %s2.%s %sInstall an agent:%s chatty --install %s\"Agent Name\"%s\n", 
		colorYellow, colorReset, colorYellow, colorReset, colorYellow, colorReset)
	fmt.Printf("   %s3.%s %sFilter by category:%s chatty --store --category %s\"Category Name\"%s\n", 
		colorYellow, colorReset, colorYellow, colorReset, colorYellow, colorReset)
	fmt.Printf("   %s4.%s %sSearch agents:%s chatty --store --search %s\"query\"%s\n\n", 
		colorYellow, colorReset, colorYellow, colorReset, colorYellow, colorReset)
	
	return nil
}

// formatAuthor returns a formatted author string if present
func (h *Handler) formatAuthor(author string) string {
	if author == "" {
		return " - "
	}
	return " by " + author + " - "
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

// ListAgentsByCategory displays agents filtered by a specific category
func (h *Handler) ListAgentsByCategory(categoryName string) error {
	// Define color constants for better readability
	colorMagenta := "\033[1;35m"
	colorCyan := "\033[1;36m"
	colorYellow := "\033[1;33m" // Bright yellow for better visibility
	colorWhite := "\033[1;37m"  // Bright white for better visibility
	colorReset := "\033[0m"
	colorGray := "\033[1;37m"   // Changed from dark gray to light gray for better readability
	
	// Start loading animation
	anim := NewStoreAnimation("Fetching community store data...")
	anim.Start()
	
	// Fetch index
	index, err := h.client.FetchIndex()
	if err != nil {
		anim.Stop()
		return err
	}
	
	// Load store configurations
	h.LoadStoreConfigs()
	
	// Categorize agents
	categorizedAgents := h.categorizeAgents(index.Files)
	
	// Find the matching category (case-insensitive)
	var matchedCategory string
	var agents []AgentInfo
	
	for category, categoryAgents := range categorizedAgents {
		if strings.EqualFold(category, categoryName) {
			matchedCategory = category
			agents = categoryAgents
			break
		}
	}
	
	// Stop animation before displaying results
	anim.Stop()
	
	// Check if the category was found
	if matchedCategory == "" {
		fmt.Printf("\n%s❌ Category not found: %s%s\n\n", 
			colorYellow, categoryName, colorReset)
		
		// Show available categories
		fmt.Printf("%sAvailable categories:%s\n", colorCyan, colorReset)
		
		// Sort categories alphabetically
		var categories []string
		for category := range categorizedAgents {
			if category != "All Agents" && len(categorizedAgents[category]) > 0 {
				categories = append(categories, category)
			}
		}
		sort.Strings(categories)
		
		// Display available categories
		for _, category := range categories {
			fmt.Printf("  • %s (%d agents)\n", category, len(categorizedAgents[category]))
		}
		
		fmt.Println("\nTry one of these categories with:")
		fmt.Printf("  chatty --store --category \"Category Name\"\n\n")
		
		return fmt.Errorf("category not found: %s", categoryName)
	}
	
	// Get category description
	description := h.getCategoryDescription(matchedCategory)
	
	// Display header
	fmt.Printf("\n%s📂 Category: %s%s%s (%d agents)%s\n", 
		colorMagenta, colorYellow, matchedCategory, colorReset, len(agents), colorReset)
	
	// Display category description if available
	if description != "" {
		fmt.Printf("%s%s%s\n", colorGray, description, colorReset)
	}
	
	fmt.Printf("%s%s%s\n\n", 
		colorMagenta, strings.Repeat("━", 50), colorReset)
	
	// Display agents
	for _, agent := range agents {
		// Print each agent with emoji, name, and description
		fmt.Printf("  %s %s%s%s%s %s%s%s\n",
			agent.Emoji,
			colorYellow,
			agent.Name,
			colorReset,
			h.formatAuthor(agent.Author),
			colorWhite,
			agent.Description,
			colorReset)
		
		fmt.Println()
	}
	
	// Display help section
	fmt.Printf("%s💡 Quick Actions%s\n", colorCyan, colorReset)
	fmt.Printf("%s%s%s\n", colorCyan, strings.Repeat("━", 50), colorReset)
	fmt.Printf("   %s1.%s %sView agent details:%s chatty --show %s\"Agent Name\"%s\n", 
		colorYellow, colorReset, colorYellow, colorReset, colorYellow, colorReset)
	fmt.Printf("   %s2.%s %sInstall an agent:%s chatty --install %s\"Agent Name\"%s\n", 
		colorYellow, colorReset, colorYellow, colorReset, colorYellow, colorReset)
	fmt.Printf("   %s3.%s %sReturn to store:%s chatty --store\n\n", 
		colorYellow, colorReset, colorYellow, colorReset)
	
	return nil
}

// ListAgentsByTags displays agents filtered by specific tags
func (h *Handler) ListAgentsByTags(tags []string) error {
	// Define color constants for better readability
	colorMagenta := "\033[1;35m"
	colorCyan := "\033[1;36m"
	colorYellow := "\033[1;33m" // Bright yellow for better visibility
	colorWhite := "\033[1;37m"  // Bright white for better visibility
	colorReset := "\033[0m"
	colorBlue := "\033[1;34m"
	colorGray := "\033[1;37m"   // Changed from dark gray to light gray for better readability
	
	// Start loading animation
	anim := NewStoreAnimation("Fetching community store data...")
	anim.Start()
	
	// Fetch index
	index, err := h.client.FetchIndex()
	if err != nil {
		anim.Stop()
		return err
	}
	
	// Load store configurations
	h.LoadStoreConfigs()
	
	// Filter agents by tags
	filteredAgents := h.filterAgentsByTags(index.Files, tags)
	
	// Stop animation before displaying results
	anim.Stop()
	
	// Check if any agents match the tags
	if len(filteredAgents) == 0 {
		fmt.Printf("\n%s❌ No agents found with tags: %s%s\n\n", 
			colorYellow, strings.Join(tags, ", "), colorReset)
		
		// Show available tags if we have tag definitions
		if h.tagsConfig != nil && len(h.tagsConfig.Tags) > 0 {
			fmt.Printf("%sAvailable tags:%s\n", colorCyan, colorReset)
			
			// Get and sort tag names
			var tagNames []string
			for tagName := range h.tagsConfig.Tags {
				tagNames = append(tagNames, tagName)
			}
			sort.Strings(tagNames)
			
			// Display available tags
			for _, tagName := range tagNames {
				tagDef := h.tagsConfig.Tags[tagName]
				fmt.Printf("  • %s%s%s: %s%s%s\n", 
					colorBlue, tagDef.Name, colorReset, 
					colorWhite, tagDef.Description, colorReset)
			}
			fmt.Println()
		}
		
		return fmt.Errorf("no agents found with specified tags")
	}
	
	// Display header with tag filter
	fmt.Printf("\n%s🏷️  Agents with tags: %s%s%s (%d agents)%s\n", 
		colorMagenta, colorYellow, strings.Join(tags, ", "), colorReset, len(filteredAgents), colorReset)
	
	fmt.Printf("%s%s%s\n\n", 
		colorMagenta, strings.Repeat("━", 50), colorReset)
	
	// Display all filtered agents
	for _, agent := range filteredAgents {
		// Print each agent with emoji, name, and description
		fmt.Printf("  %s %s%s%s%s %s%s%s\n",
			agent.Emoji,
			colorYellow,
			agent.Name,
			colorReset,
			h.formatAuthor(agent.Author),
			colorWhite,
			agent.Description,
			colorReset)
		
		// Display tags with highlighting for matched tags
		if len(agent.Tags) > 0 {
			fmt.Printf("     %sTags:%s ", colorGray, colorReset)
			for i, tag := range agent.Tags {
				if i > 0 {
					fmt.Print(", ")
				}
				
				// Highlight matched tags
				isMatched := false
				for _, filterTag := range tags {
					if strings.EqualFold(tag, filterTag) {
						isMatched = true
						break
					}
				}
				
				if isMatched {
					fmt.Printf("%s%s%s", colorCyan, tag, colorReset)
				} else {
					fmt.Printf("%s%s%s", colorBlue, tag, colorReset)
				}
			}
			fmt.Println()
		}
		
		fmt.Println()
	}
	
	// Display help section
	fmt.Printf("%s💡 Quick Actions%s\n", colorCyan, colorReset)
	fmt.Printf("%s%s%s\n", colorCyan, strings.Repeat("━", 50), colorReset)
	fmt.Printf("   %s1.%s %sView agent details:%s chatty --show %s\"Agent Name\"%s\n", 
		colorYellow, colorReset, colorYellow, colorReset, colorYellow, colorReset)
	fmt.Printf("   %s2.%s %sInstall an agent:%s chatty --install %s\"Agent Name\"%s\n", 
		colorYellow, colorReset, colorYellow, colorReset, colorYellow, colorReset)
	fmt.Printf("   %s3.%s %sReturn to store:%s chatty --store\n\n", 
		colorYellow, colorReset, colorYellow, colorReset)
	
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
	anim := NewStoreAnimation("Fetching agent index from community store...")
	anim.Start()

	// Fetch the store index
	index, err := h.client.FetchIndex()
	
	// Stop animation before returning
	anim.Stop()
	
	if err != nil {
		return nil, err
	}

	return index, nil
}

// SearchAgents searches for agents matching the query in name, description, or tags
func (h *Handler) SearchAgents(query string) error {
	// Define color constants for better readability
	colorMagenta := "\033[1;35m"
	colorCyan := "\033[1;36m"
	colorYellow := "\033[1;33m" // Bright yellow for better visibility
	colorGreen := "\033[32m"
	colorReset := "\033[0m"
	colorBlue := "\033[1;34m"
	colorWhite := "\033[1;37m"  // Bright white for better visibility
	colorGray := "\033[1;37m"   // Changed from dark gray to light gray for better readability
	
	// Convert query to lowercase for case-insensitive search
	searchTerm := strings.ToLower(query)

	// Fetch index
	index, err := h.GetIndex()
	if err != nil {
		return err
	}

	// Filter agents by search term
	var matchedAgents []AgentInfo
	for _, agent := range index.Files {
		// Check if the search term appears in name, description, or tags
		nameMatch := strings.Contains(strings.ToLower(agent.Name), searchTerm)
		descMatch := strings.Contains(strings.ToLower(agent.Description), searchTerm)
		
		// Check tags
		tagMatch := false
		for _, tag := range agent.Tags {
			if strings.Contains(strings.ToLower(tag), searchTerm) {
				tagMatch = true
				break
			}
		}
		
		// Add agent if any field matches
		if nameMatch || descMatch || tagMatch {
			matchedAgents = append(matchedAgents, agent)
		}
	}
	
	// Check if any matches were found
	if len(matchedAgents) == 0 {
		fmt.Printf("\n%s❌ No agents found matching: '%s'%s\n\n", 
			colorYellow, query, colorReset)
		fmt.Printf("%sTry a different search term or browse all agents with:%s\n", 
			colorCyan, colorReset)
		fmt.Printf("  chatty --store\n\n")
		return fmt.Errorf("no agents found matching search term: %s", query)
	}
	
	// Sort agents alphabetically by name
	sort.Slice(matchedAgents, func(i, j int) bool {
		return matchedAgents[i].Name < matchedAgents[j].Name
	})
	
	// Display header with search term
	fmt.Printf("\n%s🔍 Search results for: '%s' (%d agents found)%s\n", 
		colorMagenta, query, len(matchedAgents), colorReset)
	fmt.Printf("%s%s%s\n\n", 
		colorMagenta, strings.Repeat("━", 50), colorReset)
	
	// Display matched agents
	for _, agent := range matchedAgents {
		// Print each agent with emoji, name, and description
		fmt.Printf("  %s %s%s%s%s %s%s%s\n",
			agent.Emoji,
			colorYellow,
			agent.Name,
			colorReset,
			h.formatAuthor(agent.Author),
			colorWhite,
			agent.Description,
			colorReset)
		
		// Display tags if available
		if len(agent.Tags) > 0 {
			fmt.Printf("     %sTags:%s ", colorGray, colorReset)
			for i, tag := range agent.Tags {
				if i > 0 {
					fmt.Print(", ")
				}
				
				// Highlight tag if it matches the search term
				if strings.Contains(strings.ToLower(tag), searchTerm) {
					fmt.Printf("%s%s%s", colorCyan, tag, colorReset)
				} else {
					fmt.Printf("%s%s%s", colorBlue, tag, colorReset)
				}
			}
			fmt.Println()
		}
		
		fmt.Println()
	}
	
	// Display help section
	fmt.Printf("%s💡 Quick Actions%s\n", colorCyan, colorReset)
	fmt.Printf("%s%s%s\n", colorCyan, strings.Repeat("─", 50), colorReset)
	fmt.Printf("   %s1.%s %sView agent details:%s chatty --show %s\"Agent Name\"%s\n", 
		colorGreen, colorReset, colorYellow, colorReset, colorBlue, colorReset)
	fmt.Printf("   %s2.%s %sInstall an agent:%s chatty --install %s\"Agent Name\"%s\n", 
		colorGreen, colorReset, colorYellow, colorReset, colorBlue, colorReset)
	fmt.Printf("   %s3.%s %sReturn to store:%s chatty --store\n", 
		colorGreen, colorReset, colorYellow, colorReset)
	
	return nil
} 