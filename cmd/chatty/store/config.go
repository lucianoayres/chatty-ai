package store

// Configuration constants for the Community Store
const (
	// Base URL for the community store repository
	baseURL = "https://raw.githubusercontent.com/lucianoayres/chatty-ai-community-store/refs/heads/main"
	
	// Index file path relative to base URL
	indexPath = "index.json"
	
	// Agents directory path relative to base URL
	agentsPath = "agents"
	
	// HTTP request timeout in seconds
	requestTimeout = 30
)

// GetIndexURL returns the full URL for the index file
func GetIndexURL() string {
	return baseURL + "/" + indexPath
}

// GetAgentURL returns the full URL for an agent's YAML file
func GetAgentURL(filename string) string {
	return baseURL + "/" + agentsPath + "/" + filename
} 