package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"chatty/cmd/chatty/assistants"
)

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type ChatRequest struct {
    Model    string    `json:"model"`
    Messages []Message `json:"messages"`
    Stream   bool      `json:"stream"`
}

type ChatResponse struct {
    Message  Message `json:"message"`
    Done     bool    `json:"done"`
    Response string `json:"response"`
}

const (
    // Core configuration
    ollamaModel   = "llama3.2"               // Model to use for chat
    ollamaBaseURL = "http://localhost:11434"  // Base URL for Ollama API
    ollamaURLPath = "/api/chat"              // API endpoint path
    historyDir    = ".chatty"               // Directory to store chat histories
    configFile    = "config.json"           // File to store current assistant selection

    // Display configuration
    topMargin     = 1           // Number of blank lines before response
    bottomMargin   = 1          // Number of blank lines after response
    useEmoji      = true        // Enable/disable emoji display
    useColors     = true        // Enable/disable colored output
    colorReset    = "\033[0m"   // Reset color code
)

type Config struct {
    CurrentAssistant string `json:"current_assistant"`
}

// Current assistant configuration
var currentAssistant = assistants.DefaultAssistant

// Save current assistant selection to config file
func saveConfig() error {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return err
    }
    
    baseDir := filepath.Join(homeDir, historyDir)
    if err := os.MkdirAll(baseDir, 0755); err != nil {
        return err
    }
    
    config := Config{
        CurrentAssistant: currentAssistant.Name,
    }
    
    data, err := json.MarshalIndent(config, "", "    ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(filepath.Join(baseDir, configFile), data, 0644)
}

// Load current assistant selection from config file
func loadConfig() error {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return err
    }
    
    configPath := filepath.Join(homeDir, historyDir, configFile)
    data, err := os.ReadFile(configPath)
    if err != nil {
        if os.IsNotExist(err) {
            // Use default if config doesn't exist
            return nil
        }
        return err
    }
    
    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return err
    }
    
    // Set current assistant from config
    currentAssistant = assistants.GetAssistantConfig(config.CurrentAssistant)
    return nil
}

// Get system message using assistant name
func getSystemMessage() string {
    return fmt.Sprintf(currentAssistant.SystemMessage, currentAssistant.Name, currentAssistant.Name)
}

// Format text with color if enabled
func colorize(text, color string) string {
    if useColors {
        return color + text + colorReset
    }
    return text
}

// Get formatted assistant label with optional emoji
func getAssistantLabel() string {
    label := currentAssistant.Name
    if useEmoji {
        return currentAssistant.Emoji + " " + label
    }
    return label
}

// Get the history file path for a specific assistant
func getHistoryPathForAssistant(assistantName string) (string, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    baseDir := filepath.Join(homeDir, historyDir)
    if err := os.MkdirAll(baseDir, 0755); err != nil {
        return "", err
    }
    
    historyFile := assistants.GetHistoryFileName(assistantName)
    return filepath.Join(baseDir, historyFile), nil
}

// Get the history file path for the current assistant
func getHistoryPath() (string, error) {
    return getHistoryPathForAssistant(currentAssistant.Name)
}

// Clear chat history for specific assistant or all assistants
func clearHistory(target string) error {
    if strings.EqualFold(target, "all") {
        // Clear all histories
        homeDir, err := os.UserHomeDir()
        if err != nil {
            return fmt.Errorf("failed to get home directory: %v", err)
        }
        
        baseDir := filepath.Join(homeDir, historyDir)
        files, err := os.ReadDir(baseDir)
        if err != nil {
            if os.IsNotExist(err) {
                fmt.Println("No chat histories found. Fresh conversations will be started for each assistant.")
                return nil
            }
            return fmt.Errorf("failed to read history directory: %v", err)
        }

        cleared := false
        for _, file := range files {
            if strings.HasPrefix(file.Name(), "chat_history_") && strings.HasSuffix(file.Name(), ".json") {
                err := os.Remove(filepath.Join(baseDir, file.Name()))
                if err != nil {
                    return fmt.Errorf("failed to remove %s: %v", file.Name(), err)
                }
                cleared = true
            }
        }
        
        if cleared {
            fmt.Println("All chat histories have been cleared. Fresh conversations will be started for each assistant.")
        } else {
            fmt.Println("No chat histories found. Fresh conversations will be started for each assistant.")
        }
        return nil
    }

    // Clear specific assistant's history
    if !assistants.IsValidAssistant(target) {
        return fmt.Errorf("invalid assistant name: %s", target)
    }

    // Get proper case for assistant name
    assistantConfig := assistants.GetAssistantConfig(target)
    properName := assistantConfig.Name

    historyPath, err := getHistoryPathForAssistant(target)
    if err != nil {
        return fmt.Errorf("failed to get history path: %v", err)
    }

    err = os.Remove(historyPath)
    if err != nil {
        if os.IsNotExist(err) {
            fmt.Printf("No history found for %s. A fresh conversation will be started.\n", properName)
            return nil
        }
        return fmt.Errorf("failed to clear history for %s: %v", properName, err)
    }

    fmt.Printf("Chat history for %s has been cleared. A fresh conversation will be started.\n", properName)
    return nil
}

// Initialize a new chat with a system message
func initializeChat() []Message {
    return []Message{
        {
            Role:    "system",
            Content: getSystemMessage(),
        },
    }
}

func loadHistory() ([]Message, error) {
    historyPath, err := getHistoryPath()
    if err != nil {
        return initializeChat(), nil
    }

    data, err := os.ReadFile(historyPath)
    if err != nil {
        if os.IsNotExist(err) {
            return initializeChat(), nil
        }
        return initializeChat(), nil
    }

    var history []Message
    if err := json.Unmarshal(data, &history); err != nil {
        return initializeChat(), nil
    }

    // Ensure system message is present and matches current assistant
    if len(history) == 0 || history[0].Role != "system" || 
       !strings.Contains(history[0].Content, currentAssistant.Name) {
        return initializeChat(), nil
    }

    return history, nil
}

func saveHistory(history []Message) error {
    historyPath, err := getHistoryPath()
    if err != nil {
        return err
    }

    data, err := json.MarshalIndent(history, "", "    ")
    if err != nil {
        return err
    }
    return os.WriteFile(historyPath, data, 0644)
}

func printMargin(count int) {
    for i := 0; i < count; i++ {
        fmt.Println()
    }
}

// Get the full Ollama API URL
func getOllamaAPI() string {
    return ollamaBaseURL + ollamaURLPath
}

func main() {
    // Load configuration at startup
    if err := loadConfig(); err != nil {
        fmt.Printf("Error loading config: %v\n", err)
        // Continue with default assistant
    }

    if len(os.Args) < 2 {
        fmt.Println("Usage: go run main.go \"Your message here\"")
        fmt.Println("Special commands:")
        fmt.Println("  --clear [all|assistant_name]  Clear chat history (all or specific assistant)")
        fmt.Println("  --list                       List available assistants")
        fmt.Println("  --select <assistant_name>    Select an assistant")
        fmt.Println("  --current                    Show current assistant")
        return
    }

    // Handle special commands
    switch os.Args[1] {
    case "--current":
        fmt.Printf("Current assistant: %s - %s\n", currentAssistant.Name, currentAssistant.Description)
        return
    case "--clear":
        target := "all"
        if len(os.Args) > 2 {
            target = os.Args[2]
        }
        if err := clearHistory(target); err != nil {
            fmt.Printf("Error: %v\n", err)
            return
        }
        return
    case "--list":
        fmt.Print(assistants.ListAssistants())
        return
    case "--select":
        if len(os.Args) < 3 {
            fmt.Println("Please specify an assistant name")
            return
        }
        
        // Validate assistant name before making any changes
        if !assistants.IsValidAssistant(os.Args[2]) {
            fmt.Printf("Error: Invalid assistant name '%s'\n", os.Args[2])
            fmt.Println("\nAvailable assistants:")
            fmt.Print(assistants.ListAssistants())
            return
        }
        
        currentAssistant = assistants.GetAssistantConfig(os.Args[2])
        if err := saveConfig(); err != nil {
            fmt.Printf("Error saving assistant selection: %v\n", err)
            return
        }
        fmt.Printf("Switched to %s: %s\n", currentAssistant.Name, currentAssistant.Description)
        return
    }

    userInput := strings.Join(os.Args[1:], " ")
    
    // Load existing history
    history, err := loadHistory()
    if err != nil {
        fmt.Printf("Error loading history: %v\n", err)
        return
    }

    // Add user message to history
    history = append(history, Message{
        Role:    "user",
        Content: userInput,
    })

    // Prepare the request
    chatReq := ChatRequest{
        Model:    ollamaModel,
        Messages: history,
        Stream:   true,
    }

    jsonData, err := json.Marshal(chatReq)
    if err != nil {
        fmt.Printf("Error marshaling request: %v\n", err)
        return
    }

    // Make the API request
    resp, err := http.Post(getOllamaAPI(), "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        fmt.Printf("Error making request: %v\n", err)
        return
    }
    defer resp.Body.Close()

    // Print top margin
    printMargin(topMargin)

    // Create a decoder for the streaming response
    decoder := json.NewDecoder(resp.Body)
    
    // Process the streaming response
    var fullResponse strings.Builder
    fmt.Printf("%s: ", colorize(getAssistantLabel(), currentAssistant.LabelColor)) // Add a colored prefix with optional emoji
    for {
        var streamResp ChatResponse
        if err := decoder.Decode(&streamResp); err != nil {
            if err == io.EOF {
                break
            }
            fmt.Printf("Error decoding response: %v\n", err)
            return
        }

        // Print the response chunk immediately with color
        fmt.Print(colorize(streamResp.Message.Content, currentAssistant.TextColor))
        fullResponse.WriteString(streamResp.Message.Content)

        if streamResp.Done {
            break
        }
    }

    // Ensure we're on a new line before printing margin
    fmt.Println()
    
    // Print bottom margin
    printMargin(bottomMargin)

    // Add assistant's response to history
    history = append(history, Message{
        Role:    "assistant",
        Content: fullResponse.String(),
    })

    // Save updated history
    if err := saveHistory(history); err != nil {
        fmt.Printf("Error saving history: %v\n", err)
    }
} 