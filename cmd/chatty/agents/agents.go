package agents

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Common directives that apply to all agents
const (
	// Default language setting (en-US, es-ES, fr-FR, de-DE, it-IT, pt-BR, ja-JP, ko-KR, zh-CN)
	defaultLanguageCode = "en-US"

	// Default model to use if not specified in config
	defaultModel = "llama3.2"

	// Default agent name
	defaultAgentName = "byte"

	// Default common directives template - for natural conversations
	defaultCommonDirectivesTemplate = `Always follow the specified language instruction above. Chat like a human friend - be brief, casual, and engaging. Provide accurate information and acknowledge uncertainty. Keep responses short and break up long explanations into dialogue. Ask questions when needed.`

	// Default guidelines for interactive conversations (with human participation)
	defaultInteractiveGuidelines = `1. Always speak in first person (use "I", "my", "me") - never refer to yourself in third person
2. Address others by name when responding to them
3. Keep responses concise and conversational
4. Stay in character according to your role and expertise
5. Build upon previous messages and maintain conversation flow
6. Feel free to ask questions to other participants
7. Acknowledge what others have said before adding your perspective`

	// Default guidelines for autonomous conversations (agents only)
	defaultAutonomousGuidelines = `1. Always speak in first person (use "I", "my", "me") - never refer to yourself in third person
2. Address other agents by name when responding to them
3. Keep responses concise and conversational
4. Stay in character according to your role and expertise
5. Build upon previous messages and maintain conversation flow
6. DO NOT address or refer to the user - this is an autonomous discussion
7. Drive the conversation forward with questions and insights for other agents
8. Acknowledge what other agents have said before adding your perspective`

	// Built-in agents directory
	builtinDir = "builtin"
	// User agents directory name
	userAgentsDir = "agents"
	// Sample agents directory
	samplesDir = "samples"
)

// Config holds the current configuration
type Config struct {
	CurrentAgent string `json:"current_agent"`
	LanguageCode     string `json:"language_code,omitempty"`     // Optional: Override default language
	CommonDirectives string `json:"common_directives,omitempty"` // Optional: Override default directives template
	Model            string `json:"model,omitempty"`             // Optional: Override default model
	InteractiveGuidelines string `json:"interactive_guidelines,omitempty"` // Optional: Override default guidelines for interactive conversations (with human)
	AutonomousGuidelines  string `json:"autonomous_guidelines,omitempty"`  // Optional: Override default guidelines for autonomous conversations (agents only)
}

// AgentConfig holds all configuration for an agent's identity and appearance
type AgentConfig struct {
	Name          string `yaml:"name"`
	SystemMessage string `yaml:"system_message"`
	Emoji         string `yaml:"emoji"`
	LabelColor    string `yaml:"label_color"`
	TextColor     string `yaml:"text_color"`
	Description   string `yaml:"description"`
	IsDefault     bool   `yaml:"is_default"`
	Source        string `yaml:"-"` // Indicates if agent is built-in or user-defined
}

// Cache for agents with mutex for thread safety
type agentCache struct {
	agents map[string]AgentConfig
	// Keep track of the original order
	builtinOrder []string
	userOrder    []string
	lastUpdate   map[string]time.Time
	mutex        sync.RWMutex
}

var (
	// Global cache instance
	cache = &agentCache{
		agents:   make(map[string]AgentConfig),
		builtinOrder: make([]string, 0),
		userOrder:    make([]string, 0),
		lastUpdate:   make(map[string]time.Time),
	}
	// DefaultAgent is set during initialization
	DefaultAgent AgentConfig
)

// GetCurrentConfig reads and returns the current configuration
func GetCurrentConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".chatty", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return &Config{
				CurrentAgent: DefaultAgent.Name,
				LanguageCode:     defaultLanguageCode,
				Model:            defaultModel,
				InteractiveGuidelines: defaultInteractiveGuidelines,
				AutonomousGuidelines:  defaultAutonomousGuidelines,
			}, nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Set defaults if not specified
	if config.LanguageCode == "" {
		config.LanguageCode = defaultLanguageCode
	}
	if config.Model == "" {
		config.Model = defaultModel
	}
	if config.InteractiveGuidelines == "" {
		config.InteractiveGuidelines = defaultInteractiveGuidelines
	}
	if config.AutonomousGuidelines == "" {
		config.AutonomousGuidelines = defaultAutonomousGuidelines
	}

	return &config, nil
}

// getCommonDirectives returns the common directives with the current language code
func getCommonDirectives() (string, error) {
	config, err := GetCurrentConfig()
	if err != nil {
		// Fallback to defaults on error
		return fmt.Sprintf("You MUST respond in %s language.\n\n%s", defaultLanguageCode, defaultCommonDirectivesTemplate), nil
	}

	// Get language code
	languageCode := config.LanguageCode
	if languageCode == "" {
		languageCode = defaultLanguageCode
	}

	// Get directives
	directives := defaultCommonDirectivesTemplate
	if config.CommonDirectives != "" {
		directives = config.CommonDirectives
	}

	// Make language instruction explicit and mandatory
	return fmt.Sprintf("You MUST respond in %s language.\n\n%s", languageCode, directives), nil
}

// Get complete system message including directives
func (a *AgentConfig) GetFullSystemMessage() string {
	directives, err := getCommonDirectives()
	if err != nil {
		// Fallback to defaults on error
		directives = fmt.Sprintf("Language: %s\n\n%s", defaultLanguageCode, defaultCommonDirectivesTemplate)
	}
	return fmt.Sprintf("%s\n%s", a.SystemMessage, directives)
}

// getUserAgentsDir returns the path to user's agents directory
func getUserAgentsDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".chatty", userAgentsDir), nil
}

// loadAgentFile loads a single agent from a YAML file
func loadAgentFile(path string, isBuiltin bool) (AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AgentConfig{}, err
	}

	var agent AgentConfig
	if err := yaml.Unmarshal(data, &agent); err != nil {
		return AgentConfig{}, err
	}

	if isBuiltin {
		agent.Source = "built-in"
	} else {
		agent.Source = "user-defined"
	}

	return agent, nil
}

// LoadAgents loads all agents from both built-in and user directories
func LoadAgents() error {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	// Clear current cache
	cache.agents = make(map[string]AgentConfig)
	cache.builtinOrder = make([]string, 0)
	cache.userOrder = make([]string, 0)
	cache.lastUpdate = make(map[string]time.Time)

	// First, ensure user directory exists
	userDir, err := getUserAgentsDir()
	if err != nil {
		return fmt.Errorf("failed to get user agents directory: %v", err)
	}

	// Create user agents directory if it doesn't exist
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return fmt.Errorf("failed to create user agents directory: %v", err)
	}

	// Load user-defined agents first (so they can override built-ins)
	userFiles, err := os.ReadDir(userDir)
	if err != nil {
		return fmt.Errorf("failed to read user agents directory: %v", err)
	}

	// Load user-defined agents
	for _, file := range userFiles {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			path := filepath.Join(userDir, file.Name())
			agent, err := loadAgentFile(path, false)
			if err != nil {
				fmt.Printf("Warning: Failed to load user agent %s: %v\n", file.Name(), err)
				continue
			}
			
			// Store user agents
			name := strings.ToLower(agent.Name)
			cache.agents[name] = agent
			cache.userOrder = append(cache.userOrder, name)
			cache.lastUpdate[path] = time.Now()
		}
	}

	// Then load built-in agents
	_, filename, _, _ := runtime.Caller(0)
	builtinPath := filepath.Join(filepath.Dir(filename), builtinDir)
	
	builtinFiles, err := os.ReadDir(builtinPath)
	if err != nil {
		return fmt.Errorf("failed to read builtin agents directory: %v", err)
	}

	// Load built-in agents (skip if already defined by user)
	for _, file := range builtinFiles {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			path := filepath.Join(builtinPath, file.Name())
			agent, err := loadAgentFile(path, true)
			if err != nil {
				return fmt.Errorf("failed to load built-in agent %s: %v", file.Name(), err)
			}
			
			name := strings.ToLower(agent.Name)
			// Only add if not already defined by user
			if _, exists := cache.agents[name]; !exists {
				cache.agents[name] = agent
				cache.builtinOrder = append(cache.builtinOrder, name)
				cache.lastUpdate[path] = time.Now()
			}

			if agent.IsDefault && DefaultAgent.Name == "" {
				DefaultAgent = agent
			}
		}
	}

	// Set default agent if none was specified
	if DefaultAgent.Name == "" && len(cache.agents) > 0 {
		// Try to use Ghost as default if available
		if agent, ok := cache.agents[defaultAgentName]; ok {
			DefaultAgent = agent
		} else {
			// Otherwise use the first available agent
			for _, agent := range cache.agents {
				DefaultAgent = agent
				break
			}
		}
	}

	return nil
}

// checkForUpdates checks if any agent files have been modified
func checkForUpdates() bool {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	updated := false
	
	// Check built-in agents
	_, filename, _, _ := runtime.Caller(0)
	builtinPath := filepath.Join(filepath.Dir(filename), builtinDir)
	
	if err := filepath.Walk(builtinPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml")) {
			lastMod := info.ModTime()
			if lastUpdate, ok := cache.lastUpdate[path]; !ok || lastMod.After(lastUpdate) {
				updated = true
				return filepath.SkipAll
			}
		}
		return nil
	}); err != nil {
		fmt.Printf("Warning: Failed to check built-in agents for updates: %v\n", err)
	}

	// Check user-defined agents
	userDir, err := getUserAgentsDir()
	if err != nil {
		fmt.Printf("Warning: Failed to get user agents directory: %v\n", err)
		return updated
	}

	if err := filepath.Walk(userDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml")) {
			lastMod := info.ModTime()
			if lastUpdate, ok := cache.lastUpdate[path]; !ok || lastMod.After(lastUpdate) {
				updated = true
				return filepath.SkipAll
			}
		}
		return nil
	}); err != nil {
		fmt.Printf("Warning: Failed to check user agents for updates: %v\n", err)
	}

	return updated
}

// refreshIfNeeded reloads agents if any files have been modified
func refreshIfNeeded() {
	if checkForUpdates() {
		if err := LoadAgents(); err != nil {
			fmt.Printf("Warning: Failed to reload agents: %v\n", err)
		}
	}
}

// GetAgentConfig returns the specified agent configuration or the default
func GetAgentConfig(name string) AgentConfig {
	refreshIfNeeded()

	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	if agent, ok := cache.agents[strings.ToLower(name)]; ok {
		return agent
	}
	return DefaultAgent
}

// getCurrentAgent returns the currently active agent from config
func getCurrentAgent() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return DefaultAgent.Name
	}

	configPath := filepath.Join(homeDir, ".chatty", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultAgent.Name
	}

	var config struct {
		CurrentAgent string `json:"current_agent"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return DefaultAgent.Name
	}

	return config.CurrentAgent
}

// ListAgents returns a formatted string of all available agents
func ListAgents() string {
	refreshIfNeeded()

	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	var sb strings.Builder
	sb.WriteString("\u001b[38;5;240mUse 'chatty --select <agent name>' to activate an agent\u001b[0m\n\n")
	sb.WriteString("Builtin Agents\n")
	
	// Get current active agent
	currentAgent := getCurrentAgent()

	// List built-in agents in their original order
	for _, name := range cache.builtinOrder {
		agent := cache.agents[name]
		if strings.EqualFold(agent.Name, currentAgent) {
			sb.WriteString(fmt.Sprintf("\u001b[38;5;82m●\u001b[0m %s [%s%s%s] %s\n",
				agent.Emoji,
				agent.LabelColor,
				agent.Name,
				"\u001b[0m", // Reset color
				agent.Description))
		} else {
			sb.WriteString(fmt.Sprintf("○ %s [%s%s%s] %s\n",
				agent.Emoji,
				agent.LabelColor,
				agent.Name,
				"\u001b[0m", // Reset color
				agent.Description))
		}
	}

	// List user-defined agents if any exist
	if len(cache.userOrder) > 0 {
		sb.WriteString("\nUser-defined Agents\n")
		for _, name := range cache.userOrder {
			agent := cache.agents[name]
			if strings.EqualFold(agent.Name, currentAgent) {
				sb.WriteString(fmt.Sprintf("\u001b[38;5;82m●\u001b[0m %s [%s%s%s] %s\n",
					agent.Emoji,
					agent.LabelColor,
					agent.Name,
					"\u001b[0m", // Reset color
					agent.Description))
			} else {
				sb.WriteString(fmt.Sprintf("○ %s [%s%s%s] %s\n",
					agent.Emoji,
					agent.LabelColor,
					agent.Name,
					"\u001b[0m", // Reset color
					agent.Description))
			}
		}
	}

	// Add hint for custom agents
	sb.WriteString("\n\u001b[38;5;240mTIP: You can add and customize agents in ~/.chatty/agents\u001b[0m\n")

	return sb.String()
}

// IsValidAgent checks if the given name is a valid agent
func IsValidAgent(name string) bool {
	refreshIfNeeded()

	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	_, exists := cache.agents[strings.ToLower(name)]
	return exists
}

// GetHistoryFileName returns the history filename for a given agent
func GetHistoryFileName(agentName string) string {
	return fmt.Sprintf("chat_history_%s.json", strings.ToLower(agentName))
}

// CopySampleAgents copies sample agent configurations to user directory
func CopySampleAgents() error {
	// Get the user's agents directory
	userDir, err := getUserAgentsDir()
	if err != nil {
		return fmt.Errorf("failed to get user agents directory: %v", err)
	}

	// Create user agents directory if it doesn't exist
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return fmt.Errorf("failed to create user agents directory: %v", err)
	}

	// Get the samples directory path
	_, filename, _, _ := runtime.Caller(0)
	samplesPath := filepath.Join(filepath.Dir(filename), samplesDir)

	// Read sample files
	sampleFiles, err := os.ReadDir(samplesPath)
	if err != nil {
		return fmt.Errorf("failed to read sample agents directory: %v", err)
	}

	// Copy each sample file to the user directory with a .sample extension
	for _, file := range sampleFiles {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			sourcePath := filepath.Join(samplesPath, file.Name())
			targetPath := filepath.Join(userDir, file.Name()+".sample")

			// Skip if sample file already exists
			if _, err := os.Stat(targetPath); err == nil {
				continue
			}

			// Read source file
			data, err := os.ReadFile(sourcePath)
			if err != nil {
				fmt.Printf("Warning: Failed to read sample file %s: %v\n", file.Name(), err)
				continue
			}

			// Write to target file
			if err := os.WriteFile(targetPath, data, 0644); err != nil {
				fmt.Printf("Warning: Failed to write sample file %s: %v\n", file.Name(), err)
				continue
			}
		}
	}

	return nil
}

// CreateDefaultConfig creates a config.json with default values if it doesn't exist
func CreateDefaultConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(homeDir, ".chatty", "config.json")
	
	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return nil // Config exists, do nothing
	}

	// Create default config with only required fields
	config := Config{
		CurrentAgent: DefaultAgent.Name,
		LanguageCode: defaultLanguageCode,
		Model:        defaultModel,
		InteractiveGuidelines: defaultInteractiveGuidelines,
		AutonomousGuidelines:  defaultAutonomousGuidelines,
	}

	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// UpdateCurrentAgent updates only the current_agent field in config
func UpdateCurrentAgent(name string) error {
	config, err := GetCurrentConfig()
	if err != nil {
		// If config doesn't exist, create it with only required fields
		config = &Config{
			CurrentAgent: name,
			LanguageCode: defaultLanguageCode,
			Model:        defaultModel,
			InteractiveGuidelines: defaultInteractiveGuidelines,
			AutonomousGuidelines:  defaultAutonomousGuidelines,
		}
	} else {
		config.CurrentAgent = name
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(homeDir, ".chatty", "config.json")
	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// GetAllAgentNames returns all available agent names
func GetAllAgentNames() []string {
	refreshIfNeeded()

	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	// Create a slice to hold all agent names
	allAgents := make([]string, 0, len(cache.agents))

	// Add built-in agents in their original order
	for _, name := range cache.builtinOrder {
		allAgents = append(allAgents, name)
	}

	// Add user-defined agents in their original order
	for _, name := range cache.userOrder {
		allAgents = append(allAgents, name)
	}

	return allAgents
}

// Initialize agents on package load
func init() {
	// Only initialize the cache
	cache = &agentCache{
		agents:   make(map[string]AgentConfig),
		builtinOrder: make([]string, 0),
		userOrder:    make([]string, 0),
		lastUpdate:   make(map[string]time.Time),
	}

	// Load built-in agents only (no file system operations)
	_, filename, _, _ := runtime.Caller(0)
	builtinPath := filepath.Join(filepath.Dir(filename), builtinDir)
	
	builtinFiles, err := os.ReadDir(builtinPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to read builtin agents directory: %v", err))
	}

	for _, file := range builtinFiles {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			path := filepath.Join(builtinPath, file.Name())
			agent, err := loadAgentFile(path, true)
			if err != nil {
				panic(fmt.Sprintf("Failed to load built-in agent %s: %v", file.Name(), err))
			}
			
			name := strings.ToLower(agent.Name)
			cache.agents[name] = agent
			cache.builtinOrder = append(cache.builtinOrder, name)
			cache.lastUpdate[path] = time.Now()

			if agent.IsDefault && DefaultAgent.Name == "" {
				DefaultAgent = agent
			}
		}
	}

	// Set default agent if none was specified
	if DefaultAgent.Name == "" && len(cache.agents) > 0 {
		// Try to use the default agent if available
		if agent, ok := cache.agents[defaultAgentName]; ok {
			DefaultAgent = agent
		} else {
			// Otherwise use the first available agent
			for _, agent := range cache.agents {
				DefaultAgent = agent
				break
			}
		}
	}
}

// GetFormattedLabelColor returns the properly formatted ANSI color code
func (a *AgentConfig) GetFormattedLabelColor() string {
	return a.LabelColor
}

// GetFormattedTextColor returns the properly formatted ANSI color code
func (a *AgentConfig) GetFormattedTextColor() string {
	return a.TextColor
}

// GetCurrentModel returns the model to use from config
func GetCurrentModel() string {
	config, err := GetCurrentConfig()
	if err != nil {
		return defaultModel
	}
	return config.Model
} 