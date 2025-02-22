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

const historyFile = "chat_history.json"

func loadHistory() ([]Message, error) {
    // Create history directory if it doesn't exist
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return nil, err
    }
    historyDir := filepath.Join(homeDir, ".chatty")
    if err := os.MkdirAll(historyDir, 0755); err != nil {
        return nil, err
    }

    historyPath := filepath.Join(historyDir, historyFile)
    data, err := os.ReadFile(historyPath)
    if err != nil {
        if os.IsNotExist(err) {
            return []Message{}, nil
        }
        return nil, err
    }

    var history []Message
    if err := json.Unmarshal(data, &history); err != nil {
        return nil, err
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

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run main.go \"Your message here\"")
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
        Model:    "llama3.2", // Using llama3.2 model
        Messages: history,
        Stream:   true,
    }

    jsonData, err := json.Marshal(chatReq)
    if err != nil {
        fmt.Printf("Error marshaling request: %v\n", err)
        return
    }

    // Make the API request
    resp, err := http.Post("http://localhost:11434/api/chat", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        fmt.Printf("Error making request: %v\n", err)
        return
    }
    defer resp.Body.Close()

    // Create a decoder for the streaming response
    decoder := json.NewDecoder(resp.Body)
    
    // Process the streaming response
    var fullResponse strings.Builder
    for {
        var streamResp ChatResponse
        if err := decoder.Decode(&streamResp); err != nil {
            if err == io.EOF {
                break
            }
            fmt.Printf("Error decoding response: %v\n", err)
            return
        }

        // Print the response chunk immediately
        fmt.Print(streamResp.Message.Content)
        fullResponse.WriteString(streamResp.Message.Content)

        if streamResp.Done {
            break
        }
    }

    // Add assistant's response to history
    history = append(history, Message{
        Role:    "assistant",
        Content: fullResponse.String(),
    })

    // Save updated history
    if err := saveHistory(history); err != nil {
        fmt.Printf("\nError saving history: %v\n", err)
    }

    fmt.Println() // Add a newline at the end
} 