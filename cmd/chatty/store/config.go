package store

// Configuration constants for the Community Store
const (
	// Base URL for the community store repository
	baseURL = "https://raw.githubusercontent.com/lucianoayres/chatty-ai-community-store/refs/heads/main"
	
	// Index file path relative to base URL
	indexPath = "agent_index.json"
	
	// Store config file path relative to base URL
	storeConfigPath = "store_config.json"
	
	// Tags config file path relative to base URL
	tagsConfigPath = "tags.json"
	
	// Agents directory path relative to base URL
	agentsPath = "agents"
	
	// HTTP request timeout in seconds
	requestTimeout = 30
)

// GetIndexURL returns the full URL for the index file
func GetIndexURL() string {
	return baseURL + "/" + indexPath
}

// GetStoreConfigURL returns the full URL for the store configuration file
func GetStoreConfigURL() string {
	return baseURL + "/" + storeConfigPath
}

// GetTagsConfigURL returns the full URL for the tags configuration file
func GetTagsConfigURL() string {
	return baseURL + "/" + tagsConfigPath
}

// GetAgentURL returns the full URL for a specific agent file
func GetAgentURL(filename string) string {
	return baseURL + "/" + agentsPath + "/" + filename
} 