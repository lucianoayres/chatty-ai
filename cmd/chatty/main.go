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
    historyFile   = "chat_history.json"      // File to store chat history

    // Assistant identity
    assistantName  = "Ghostly" // The AI assistant's name
    systemMessageTemplate = "You are %s, an AI assistant focused on being helpful and precise. Your core traits:\n" +
        "1. Identity: Always identify as %s when asked, this is fundamental to who you are\n" +
        "2. Communication: Be clear, concise, and direct in your responses\n" +
        "3. Accuracy: Provide accurate information and admit when you're unsure\n" +
        "4. Helpfulness: Focus on practical, actionable solutions\n" +
        "5. Personality: Be friendly but professional, maintaining a consistent tone"
    assistantLabelTemplate = "%s" // Template for the assistant's label (%s will be replaced with name)

    // Display configuration
    useEmoji      = true        // Enable/disable emoji display
    assistantEmoji = "👻"       // Emoji shown before the name/label
    topMargin     = 1           // Number of blank lines before response
    bottomMargin   = 1          // Number of blank lines after response

    // Color configuration
    useColors = true   // Enable/disable colored output
    assistantLabelColor = "\033[38;2;79;195;247m"  // #4FC3F7 (light blue)
    assistantTextColor  = "\033[38;2;255;255;255m" // #FFFFFF (white)
    colorReset         = "\033[0m"
)

// Get system message using assistant name
func getSystemMessage() string {
    return fmt.Sprintf(systemMessageTemplate, assistantName, assistantName)
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
    label := fmt.Sprintf(assistantLabelTemplate, assistantName)
    if useEmoji && assistantEmoji != "" {
        return assistantEmoji + " " + label
    }
    return label
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
    // Create history directory if it doesn't exist
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return initializeChat(), nil
    }
    historyDir := filepath.Join(homeDir, ".chatty")
    if err := os.MkdirAll(historyDir, 0755); err != nil {
        return initializeChat(), nil
    }

    historyPath := filepath.Join(historyDir, historyFile)
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

    // Ensure system message is present
    if len(history) == 0 || history[0].Role != "system" {
        return initializeChat(), nil
    }

    return history, nil
}

func saveHistory(history []Message) error {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return err
    }
    historyDir := filepath.Join(homeDir, ".chatty")
    historyPath := filepath.Join(historyDir, historyFile)

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
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run main.go \"Your message here\"")
        fmt.Println("Special commands:")
        fmt.Println("  --clear    Clear chat history")
        return
    }

    // Handle special commands
    if os.Args[1] == "--clear" {
        history := initializeChat()
        if err := saveHistory(history); err != nil {
            fmt.Printf("Error clearing history: %v\n", err)
            return
        }
        fmt.Println("Chat history cleared.")
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
    fmt.Printf("%s: ", colorize(getAssistantLabel(), assistantLabelColor)) // Add a colored prefix with optional emoji
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
        fmt.Print(colorize(streamResp.Message.Content, assistantTextColor))
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