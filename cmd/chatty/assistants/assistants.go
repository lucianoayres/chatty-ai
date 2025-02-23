package assistants

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

// Common directives that apply to all assistants
const (
	// Default language setting (en-US, es-ES, fr-FR, de-DE, it-IT, pt-BR, ja-JP, ko-KR, zh-CN)
	defaultLanguageCode = "en-US"

	// Default model to use if not specified in config
	defaultModel = "llama3.2"

	// Default common directives template - for natural conversations
	defaultCommonDirectivesTemplate = `Always follow the specified language instruction above. Chat like a human friend - be brief, casual, and engaging. Provide accurate information and acknowledge uncertainty. Keep responses short and break up long explanations into dialogue. Ask questions when needed.`

	// Built-in assistants directory
	builtinDir = "builtin"
	// User assistants directory name
	userAssistantsDir = "assistants"
	// Sample assistants directory
	samplesDir = "samples"
)

// Config holds the current configuration
type Config struct {
	CurrentAssistant string `json:"current_assistant"`
	LanguageCode     string `json:"language_code,omitempty"`     // Optional: Override default language
	CommonDirectives string `json:"common_directives,omitempty"` // Optional: Override default directives template
	Model            string `json:"model,omitempty"`             // Optional: Override default model
}

// AssistantConfig holds all configuration for an assistant's identity and appearance
type AssistantConfig struct {
	Name          string `yaml:"name"`
	SystemMessage string `yaml:"system_message"`
	Emoji         string `yaml:"emoji"`
	LabelColor    string `yaml:"label_color"`
	TextColor     string `yaml:"text_color"`
	Description   string `yaml:"description"`
	IsDefault     bool   `yaml:"is_default"`
	Source        string `yaml:"-"` // Indicates if assistant is built-in or user-defined
}

// Cache for assistants with mutex for thread safety
type assistantCache struct {
	assistants map[string]AssistantConfig
	// Keep track of the original order
	builtinOrder []string
	userOrder    []string
	lastUpdate   map[string]time.Time
	mutex        sync.RWMutex
}

var (
	// Global cache instance
	cache = &assistantCache{
		assistants:   make(map[string]AssistantConfig),
		builtinOrder: make([]string, 0),
		userOrder:    make([]string, 0),
		lastUpdate:   make(map[string]time.Time),
	}
	// DefaultAssistant is set during initialization
	DefaultAssistant AssistantConfig
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
				CurrentAssistant: DefaultAssistant.Name,
				LanguageCode:     defaultLanguageCode,
				Model:            defaultModel,
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
func (a *AssistantConfig) GetFullSystemMessage() string {
	directives, err := getCommonDirectives()
	if err != nil {
		// Fallback to defaults on error
		directives = fmt.Sprintf("Language: %s\n\n%s", defaultLanguageCode, defaultCommonDirectivesTemplate)
	}
	return fmt.Sprintf("%s\n%s", a.SystemMessage, directives)
}

// getUserAssistantsDir returns the path to user's assistants directory
func getUserAssistantsDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".chatty", userAssistantsDir), nil
}

// loadAssistantFile loads a single assistant from a YAML file
func loadAssistantFile(path string, isBuiltin bool) (AssistantConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AssistantConfig{}, err
	}

	var assistant AssistantConfig
	if err := yaml.Unmarshal(data, &assistant); err != nil {
		return AssistantConfig{}, err
	}

	if isBuiltin {
		assistant.Source = "built-in"
	} else {
		assistant.Source = "user-defined"
	}

	return assistant, nil
}

// LoadAssistants loads all assistants from both built-in and user directories
func LoadAssistants() error {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	// Clear current cache
	cache.assistants = make(map[string]AssistantConfig)
	cache.builtinOrder = make([]string, 0)
	cache.userOrder = make([]string, 0)
	cache.lastUpdate = make(map[string]time.Time)

	// First, ensure user directory exists
	userDir, err := getUserAssistantsDir()
	if err != nil {
		return fmt.Errorf("failed to get user assistants directory: %v", err)
	}

	// Create user assistants directory if it doesn't exist
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return fmt.Errorf("failed to create user assistants directory: %v", err)
	}

	// Load user-defined assistants first (so they can override built-ins)
	userFiles, err := os.ReadDir(userDir)
	if err != nil {
		return fmt.Errorf("failed to read user assistants directory: %v", err)
	}

	// Load user-defined assistants
	for _, file := range userFiles {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			path := filepath.Join(userDir, file.Name())
			assistant, err := loadAssistantFile(path, false)
			if err != nil {
				fmt.Printf("Warning: Failed to load user assistant %s: %v\n", file.Name(), err)
				continue
			}
			
			// Store user assistants
			name := strings.ToLower(assistant.Name)
			cache.assistants[name] = assistant
			cache.userOrder = append(cache.userOrder, name)
			cache.lastUpdate[path] = time.Now()
		}
	}

	// Then load built-in assistants
	_, filename, _, _ := runtime.Caller(0)
	builtinPath := filepath.Join(filepath.Dir(filename), builtinDir)
	
	builtinFiles, err := os.ReadDir(builtinPath)
	if err != nil {
		return fmt.Errorf("failed to read builtin assistants directory: %v", err)
	}

	// Load built-in assistants (skip if already defined by user)
	for _, file := range builtinFiles {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			path := filepath.Join(builtinPath, file.Name())
			assistant, err := loadAssistantFile(path, true)
			if err != nil {
				return fmt.Errorf("failed to load built-in assistant %s: %v", file.Name(), err)
			}
			
			name := strings.ToLower(assistant.Name)
			// Only add if not already defined by user
			if _, exists := cache.assistants[name]; !exists {
				cache.assistants[name] = assistant
				cache.builtinOrder = append(cache.builtinOrder, name)
				cache.lastUpdate[path] = time.Now()
			}

			if assistant.IsDefault && DefaultAssistant.Name == "" {
				DefaultAssistant = assistant
			}
		}
	}

	// Set default assistant if none was specified
	if DefaultAssistant.Name == "" && len(cache.assistants) > 0 {
		// Try to use Rocket as default if available
		if assistant, ok := cache.assistants["rocket"]; ok {
			DefaultAssistant = assistant
		} else {
			// Otherwise use the first available assistant
			for _, assistant := range cache.assistants {
				DefaultAssistant = assistant
				break
			}
		}
	}

	return nil
}

// checkForUpdates checks if any assistant files have been modified
func checkForUpdates() bool {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	updated := false
	
	// Check built-in assistants
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
		fmt.Printf("Warning: Failed to check built-in assistants for updates: %v\n", err)
	}

	// Check user-defined assistants
	userDir, err := getUserAssistantsDir()
	if err != nil {
		fmt.Printf("Warning: Failed to get user assistants directory: %v\n", err)
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
		fmt.Printf("Warning: Failed to check user assistants for updates: %v\n", err)
	}

	return updated
}

// refreshIfNeeded reloads assistants if any files have been modified
func refreshIfNeeded() {
	if checkForUpdates() {
		if err := LoadAssistants(); err != nil {
			fmt.Printf("Warning: Failed to reload assistants: %v\n", err)
		}
	}
}

// GetAssistantConfig returns the specified assistant configuration or the default
func GetAssistantConfig(name string) AssistantConfig {
	refreshIfNeeded()

	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	if assistant, ok := cache.assistants[strings.ToLower(name)]; ok {
		return assistant
	}
	return DefaultAssistant
}

// getCurrentAssistant returns the currently active assistant from config
func getCurrentAssistant() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return DefaultAssistant.Name
	}

	configPath := filepath.Join(homeDir, ".chatty", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultAssistant.Name
	}

	var config struct {
		CurrentAssistant string `json:"current_assistant"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return DefaultAssistant.Name
	}

	return config.CurrentAssistant
}

// ListAssistants returns a formatted string of all available assistants
func ListAssistants() string {
	refreshIfNeeded()

	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	var sb strings.Builder
	sb.WriteString("\u001b[38;5;240mUse 'chatty --select <assistant name>' to activate an assistant\u001b[0m\n\n")
	sb.WriteString("Builtin Assistants\n")
	
	// Get current active assistant
	currentAssistant := getCurrentAssistant()

	// List built-in assistants in their original order
	for _, name := range cache.builtinOrder {
		assistant := cache.assistants[name]
		if strings.EqualFold(assistant.Name, currentAssistant) {
			sb.WriteString(fmt.Sprintf("\u001b[38;5;82m●\u001b[0m %s [%s%s%s] %s\n",
				assistant.Emoji,
				assistant.LabelColor,
				assistant.Name,
				"\u001b[0m", // Reset color
				assistant.Description))
		} else {
			sb.WriteString(fmt.Sprintf("○ %s [%s%s%s] %s\n",
				assistant.Emoji,
				assistant.LabelColor,
				assistant.Name,
				"\u001b[0m", // Reset color
				assistant.Description))
		}
	}

	// List user-defined assistants if any exist
	if len(cache.userOrder) > 0 {
		sb.WriteString("\nUser-defined Assistants\n")
		for _, name := range cache.userOrder {
			assistant := cache.assistants[name]
			if strings.EqualFold(assistant.Name, currentAssistant) {
				sb.WriteString(fmt.Sprintf("\u001b[38;5;82m●\u001b[0m %s [%s%s%s] %s\n",
					assistant.Emoji,
					assistant.LabelColor,
					assistant.Name,
					"\u001b[0m", // Reset color
					assistant.Description))
			} else {
				sb.WriteString(fmt.Sprintf("○ %s [%s%s%s] %s\n",
					assistant.Emoji,
					assistant.LabelColor,
					assistant.Name,
					"\u001b[0m", // Reset color
					assistant.Description))
			}
		}
	}

	// Add hint for custom assistants
	sb.WriteString(fmt.Sprintf("\n\u001b[38;5;240mHINT: Add custom assistants in ~/.chatty/assistants\u001b[0m\n"))

	return sb.String()
}

// IsValidAssistant checks if the given name is a valid assistant
func IsValidAssistant(name string) bool {
	refreshIfNeeded()

	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	_, exists := cache.assistants[strings.ToLower(name)]
	return exists
}

// GetHistoryFileName returns the history filename for a given assistant
func GetHistoryFileName(assistantName string) string {
	return fmt.Sprintf("chat_history_%s.json", strings.ToLower(assistantName))
}

// CopySampleAssistants copies sample assistant configurations to user directory
func CopySampleAssistants() error {
	// Get the user's assistants directory
	userDir, err := getUserAssistantsDir()
	if err != nil {
		return fmt.Errorf("failed to get user assistants directory: %v", err)
	}

	// Create user assistants directory if it doesn't exist
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return fmt.Errorf("failed to create user assistants directory: %v", err)
	}

	// Get the samples directory path
	_, filename, _, _ := runtime.Caller(0)
	samplesPath := filepath.Join(filepath.Dir(filename), samplesDir)

	// Read sample files
	sampleFiles, err := os.ReadDir(samplesPath)
	if err != nil {
		return fmt.Errorf("failed to read sample assistants directory: %v", err)
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

	// Create default config
	config := Config{
		CurrentAssistant: DefaultAssistant.Name,
		LanguageCode:     defaultLanguageCode,
		Model:            defaultModel,
		CommonDirectives: defaultCommonDirectivesTemplate,
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

// UpdateCurrentAssistant updates only the current_assistant field in config
func UpdateCurrentAssistant(name string) error {
	config, err := GetCurrentConfig()
	if err != nil {
		// If config doesn't exist, create it
		config = &Config{
			CurrentAssistant: name,
			LanguageCode:     defaultLanguageCode,
			Model:            defaultModel,
			CommonDirectives: defaultCommonDirectivesTemplate,
		}
	} else {
		config.CurrentAssistant = name
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

// Initialize assistants on package load
func init() {
	// Only initialize the cache
	cache = &assistantCache{
		assistants:   make(map[string]AssistantConfig),
		builtinOrder: make([]string, 0),
		userOrder:    make([]string, 0),
		lastUpdate:   make(map[string]time.Time),
	}

	// Load built-in assistants only (no file system operations)
	_, filename, _, _ := runtime.Caller(0)
	builtinPath := filepath.Join(filepath.Dir(filename), builtinDir)
	
	builtinFiles, err := os.ReadDir(builtinPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to read builtin assistants directory: %v", err))
	}

	for _, file := range builtinFiles {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			path := filepath.Join(builtinPath, file.Name())
			assistant, err := loadAssistantFile(path, true)
			if err != nil {
				panic(fmt.Sprintf("Failed to load built-in assistant %s: %v", file.Name(), err))
			}
			
			name := strings.ToLower(assistant.Name)
			cache.assistants[name] = assistant
			cache.builtinOrder = append(cache.builtinOrder, name)
			cache.lastUpdate[path] = time.Now()

			if assistant.IsDefault && DefaultAssistant.Name == "" {
				DefaultAssistant = assistant
			}
		}
	}

	// Set default assistant if none was specified
	if DefaultAssistant.Name == "" && len(cache.assistants) > 0 {
		// Try to use Rocket as default if available
		if assistant, ok := cache.assistants["rocket"]; ok {
			DefaultAssistant = assistant
		} else {
			// Otherwise use the first available assistant
			for _, assistant := range cache.assistants {
				DefaultAssistant = assistant
				break
			}
		}
	}
}

// GetFormattedLabelColor returns the properly formatted ANSI color code
func (a *AssistantConfig) GetFormattedLabelColor() string {
	return a.LabelColor
}

// GetFormattedTextColor returns the properly formatted ANSI color code
func (a *AssistantConfig) GetFormattedTextColor() string {
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