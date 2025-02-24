package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

// Add these new types after the existing types
type ConversationConfig struct {
    Assistants []string
    Starter    string
    Turns      int  // 0 means infinite
    Current    int  // Current turn
}

// Add this new type for conversation history
type ConversationHistory struct {
    Messages []Message
}

// Add this new animation type that includes assistant info
type ConversationAnimation struct {
    stopChan chan bool
    assistant assistants.AssistantConfig
}

const (
    // Core configuration
    ollamaBaseURL = "http://localhost:11434"  // Base URL for Ollama API
    ollamaURLPath = "/api/chat"              // API endpoint path
    historyDir    = ".chatty"               // Directory to store chat histories
    configFile    = "config.json"           // File to store current assistant selection

    // Request timeouts and retry settings
    maxRetries = 5                          // Increased from 3 to 5
    initialRetryDelay = 2 * time.Second     // Initial delay before first retry
    maxRetryDelay = 30 * time.Second        // Maximum delay between retries
    requestTimeout = 30 * time.Second     // Initial connection timeout
    readTimeout = 60 * time.Second       // Timeout for reading each chunk
    writeTimeout = 30 * time.Second      // Timeout for writing requests
    keepAliveTimeout = 24 * time.Hour    // Keep-alive timeout (24 hours)
    
    // Connection pool settings
    maxIdleConns = 100
    maxConnsPerHost = 100
    idleConnTimeout = 90 * time.Second
    
    // Display configuration
    chatTopMargin     = 1           // Number of blank lines before response in chat mode
    chatBottomMargin   = 1          // Number of blank lines after response in chat mode
    converseMargin     = 1          // Number of blank lines between messages in converse mode
    useEmoji      = true        // Enable/disable emoji display
    useColors     = true        // Enable/disable colored output
    colorReset    = "\033[0m"   // Reset color code
    
    // Animation configuration
    frameDelay   = 200          // Milliseconds between animation frames

    // Conversation context template
    conversationContextTemplate = `You are %s (%s) participating in a group conversation with other AI assistants and a human user. This is an ongoing discussion where everyone contributes naturally. Remember that YOU are %s - always speak in first person and never refer to yourself in third person.

    Current participants (excluding yourself):
    %s

    Important guidelines:
    1. Always speak in first person (use "I", "my", "me") - never refer to yourself in third person
    2. Address others by name when responding to them
    3. Keep responses concise and conversational
    4. Stay in character according to your role and expertise
    5. Build upon previous messages and maintain conversation flow
    6. Feel free to ask questions to other participants
    7. Acknowledge what others have said before adding your perspective

    Conversation history:
    %s

    Previous message was from: %s
    Their message: "%s"

    Please respond naturally as part of this group conversation, keeping in mind that you are %s.`

    // Maximum number of previous messages to include in conversation context
    maxConversationHistory = 6  // This will include the last 3 exchanges (3 pairs of messages)

    // Visual formatting
    turnEmoji = "💭"  // Changed to speech bubble for conversation
    turnColor = "\033[1;35m" // Bright magenta
    turnNumberColor = "\033[1;36m" // Bright cyan
    turnSeparator = "•" // Bullet point separator
    turnSeparatorColor = "\033[38;5;240m" // Dark gray
    timeIndicatorColor = "\033[38;5;246m" // Medium gray
    timeEmojiColor = "\033[38;5;220m" // Yellow for time emoji
    timeEmoji = "⏱️"
    inputPromptColor = "\033[1;37m" // Bright white
    inputHintColor = "\033[2;37m" // Dim gray
    elapsedTimeColor = "\033[38;5;246m" // Gray

    // Visual formatting for time
    timeHeaderColor = "\033[38;5;75m"  // Light blue for headers
    timeValueColor = "\033[38;5;252m"  // Light gray for values
    startTimeEmoji = "🗓️"  // Calendar emoji

    // Conversation mode timeouts (much higher due to multiple assistants and longer responses)
    converseRequestTimeout = 300 * time.Second  // Initial connection timeout for converse mode
    converseReadTimeout = 300 * time.Second    // Timeout for reading each chunk in converse mode
    converseWriteTimeout = 300 * time.Second    // Timeout for writing requests in converse mode
)

// Animation control
type Animation struct {
    stopChan chan bool
}

// Start the jumping dots animation
func startAnimation() *Animation {
    anim := &Animation{
        stopChan: make(chan bool),
    }
    
    // Start animation in background
    go func() {
        frames := []string{"   ", ".  ", ".. ", "..."}
        frameIndex := 0
        
        for {
            select {
            case <-anim.stopChan:
                return
            default:
                // Clear line and print current frame
                fmt.Printf("\r%s%s", colorize(getAssistantLabel(), currentAssistant.LabelColor), frames[frameIndex])
                
                // Move to next frame
                frameIndex = (frameIndex + 1) % len(frames)
                
                time.Sleep(time.Millisecond * frameDelay)
            }
        }
    }()
    
    return anim
}

// Stop the animation
func (a *Animation) stopAnimation() {
    a.stopChan <- true
    // Clear the animation and prepare for response
    fmt.Printf("\r%s", colorize(getAssistantLabel(), currentAssistant.LabelColor))
}

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
    return currentAssistant.GetFullSystemMessage()
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
        return currentAssistant.Emoji + " " + label + ": "
    }
    return label + ": "
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

// Cache for chat histories
var historyCache = make(map[string][]Message)

// loadHistory loads chat history with caching
func loadHistory() ([]Message, error) {
    // Check cache first
    if history, exists := historyCache[currentAssistant.Name]; exists {
        return history, nil
    }

    // For first use, just return initial chat without file I/O
    historyPath, err := getHistoryPath()
    if err != nil {
        return initializeChat(), nil
    }

    // Quick check if file exists without reading it
    if _, err := os.Stat(historyPath); os.IsNotExist(err) {
        history := initializeChat()
        historyCache[currentAssistant.Name] = history
        return history, nil
    }

    // Only read file if it actually exists
    data, err := os.ReadFile(historyPath)
    if err != nil {
        return initializeChat(), nil
    }

    var history []Message
    if err := json.Unmarshal(data, &history); err != nil {
        history = initializeChat()
        historyCache[currentAssistant.Name] = history
        return history, nil
    }

    // If history is empty or doesn't start with system message, initialize it
    if len(history) == 0 {
        history = initializeChat()
    } else if history[0].Role != "system" {
        // Prepend system message if it's missing
        newHistory := make([]Message, len(history)+1)
        newHistory[0] = Message{
            Role:    "system",
            Content: getSystemMessage(),
        }
        copy(newHistory[1:], history)
        history = newHistory
    } else if !strings.Contains(history[0].Content, currentAssistant.Name) {
        // Update system message if it's for a different assistant
        history[0] = Message{
            Role:    "system",
            Content: getSystemMessage(),
        }
    }

    // Cache the loaded history
    historyCache[currentAssistant.Name] = history
    return history, nil
}

// saveHistory saves chat history and updates cache
func saveHistory(history []Message) error {
    // Update cache
    historyCache[currentAssistant.Name] = history

    // Ensure directory exists before trying to save
    historyPath, err := getHistoryPath()
    if err != nil {
        return err
    }

    // Create directory if it doesn't exist
    dir := filepath.Dir(historyPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    // Append mode for the file
    data, err := json.MarshalIndent(history, "", "    ")
    if err != nil {
        return err
    }
    return os.WriteFile(historyPath, data, 0644)
}

// clearHistory clears chat history and cache
func clearHistory(target string) error {
    if strings.EqualFold(target, "all") {
        // Clear all histories and cache
        homeDir, err := os.UserHomeDir()
        if err != nil {
            return fmt.Errorf("failed to get home directory: %v", err)
        }
        
        baseDir := filepath.Join(homeDir, historyDir)
        files, err := os.ReadDir(baseDir)
        if err != nil {
            if os.IsNotExist(err) {
                // Clear cache
                historyCache = make(map[string][]Message)
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
        
        // Clear cache
        historyCache = make(map[string][]Message)
        
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

    // Clear cache for this assistant
    delete(historyCache, properName)

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

// Update printMargin function name to be more specific
func printChatMargin(count int) {
    for i := 0; i < count; i++ {
        fmt.Println()
    }
}

// Get the full Ollama API URL
func getOllamaAPI() string {
    return ollamaBaseURL + ollamaURLPath
}

// HTTP client without timeout (we'll use context for initial timeout)
var httpClient = &http.Client{}

// Update the animation functions for conversation mode
func startConversationAnimation(assistant assistants.AssistantConfig) *ConversationAnimation {
    anim := &ConversationAnimation{
        stopChan: make(chan bool),
        assistant: assistant,
    }
    
    // Start animation in background
    go func() {
        frames := []string{"   ", ".  ", ".. ", "..."}
        frameIndex := 0
        
        for {
            select {
            case <-anim.stopChan:
                return
            default:
                // Clear line and print current frame with correct assistant label
                label := fmt.Sprintf("%s %s: ", assistant.Emoji, assistant.Name)
                fmt.Printf("\r%s%s", colorize(label, assistant.LabelColor), frames[frameIndex])
                
                // Move to next frame
                frameIndex = (frameIndex + 1) % len(frames)
                
                time.Sleep(time.Millisecond * frameDelay)
            }
        }
    }()
    
    return anim
}

// Stop the conversation animation
func (a *ConversationAnimation) stopAnimation() {
    a.stopChan <- true
    // Clear the animation and prepare for response with correct assistant label
    label := fmt.Sprintf("%s %s: ", a.assistant.Emoji, a.assistant.Name)
    fmt.Printf("\r%s", colorize(label, a.assistant.LabelColor))
}

// Update makeAPIRequestWithRetry function
func makeAPIRequestWithRetry(jsonData []byte, history []Message, assistant string, isConversation bool) (*http.Response, error) {
    var lastErr error
    retryDelay := initialRetryDelay

    for attempt := 1; attempt <= maxRetries; attempt++ {
        // Show retry attempt if not first try
        if attempt > 1 {
            fmt.Printf("\nRetrying request for %s (attempt %d/%d)...\n", assistant, attempt, maxRetries)
        }

        resp, err := makeAPIRequest(jsonData, history, isConversation)
        if err == nil {
            return resp, nil
        }

        lastErr = err

        // Don't wait after the last attempt
        if attempt < maxRetries {
            // Calculate next delay with exponential backoff
            if attempt > 1 {
                retryDelay = time.Duration(float64(retryDelay) * 1.5)
                if retryDelay > maxRetryDelay {
                    retryDelay = maxRetryDelay
                }
            }

            fmt.Printf("Request failed: %v\nWaiting %.0f seconds before retry...\n", 
                err, retryDelay.Seconds())
            time.Sleep(retryDelay)
        }
    }

    return nil, fmt.Errorf("after %d attempts: %v", maxRetries, lastErr)
}

// Update makeAPIRequest function
func makeAPIRequest(jsonData []byte, history []Message, isConversation bool) (*http.Response, error) {
    // Create the request
    req, err := http.NewRequest("POST", getOllamaAPI(), bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("error creating request: %v", err)
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Connection", "keep-alive")
    req.Header.Set("Keep-Alive", "timeout=86400")  // 24 hours in seconds

    // Use appropriate timeouts based on mode
    currentRequestTimeout := requestTimeout
    currentWriteTimeout := writeTimeout
    if isConversation {
        currentRequestTimeout = converseRequestTimeout
        currentWriteTimeout = converseWriteTimeout
    }

    // Create a custom transport with optimized settings for streaming
    transport := &http.Transport{
        MaxIdleConns:        maxIdleConns,
        MaxConnsPerHost:     maxConnsPerHost,
        IdleConnTimeout:     keepAliveTimeout,
        DisableCompression:  false,
        DisableKeepAlives:   false,
        ForceAttemptHTTP2:   true,
        MaxIdleConnsPerHost: maxConnsPerHost,
        ResponseHeaderTimeout: currentWriteTimeout,
        ExpectContinueTimeout: 1 * time.Second,
        WriteBufferSize:     64 * 1024,
        ReadBufferSize:      64 * 1024,
    }

    // Create client with initial connection timeout only
    client := &http.Client{
        Transport: transport,
        Timeout:   currentRequestTimeout,
    }

    // Make the request
    resp, err := client.Do(req)
    if err != nil {
        if os.IsTimeout(err) {
            return nil, fmt.Errorf("connection timed out after %.0f seconds - the model might be busy", currentRequestTimeout.Seconds())
        }
        if strings.Contains(err.Error(), "connection refused") {
            return nil, fmt.Errorf("could not connect to Ollama server - make sure 'ollama serve' is running")
        }
        return nil, fmt.Errorf("error connecting to Ollama: %v", err)
    }

    // Check for error responses
    if resp.StatusCode != http.StatusOK {
        defer resp.Body.Close()
        
        // Try to read error details
        var errorResponse struct {
            Error string `json:"error"`
        }
        body, _ := io.ReadAll(resp.Body)
        if err := json.Unmarshal(body, &errorResponse); err == nil && errorResponse.Error != "" {
            if strings.Contains(errorResponse.Error, "model") {
                return nil, fmt.Errorf("invalid model '%s' - please check your config.json file", assistants.GetCurrentModel())
            }
            return nil, fmt.Errorf("API error: %s", errorResponse.Error)
        }
        
        return nil, fmt.Errorf("API error (status %d): failed to process request", resp.StatusCode)
    }

    return resp, nil
}

// Add this helper function to get recent conversation history
func getRecentConversationHistory(fullHistory string) string {
    lines := strings.Split(fullHistory, "\n")
    if len(lines) <= maxConversationHistory {
        return fullHistory
    }
    
    // Keep only the most recent messages
    recentLines := lines[len(lines)-maxConversationHistory:]
    return strings.Join(recentLines, "\n")
}

// Add this new function to format user messages consistently
func formatUserMessage(message string) string {
    return fmt.Sprintf("👤 User: %s", message)
}

// Add this new type to track conversation time
type ConversationState struct {
    startTime time.Time
    lastActive time.Time
}

// Update the formatElapsedTime function to be more precise
func formatElapsedTime(start, current time.Time) string {
    duration := current.Sub(start)
    
    parts := []string{}
    
    // Calculate each time unit
    years := int(duration.Hours() / (24 * 365))
    months := int(duration.Hours() / (24 * 30)) % 12
    weeks := int(duration.Hours() / (24 * 7)) % 4
    days := int(duration.Hours() / 24) % 7
    hours := int(duration.Hours()) % 24
    minutes := int(duration.Minutes()) % 60
    
    // Add non-zero units to parts
    if years > 0 {
        if years == 1 {
            parts = append(parts, "1 year")
        } else {
            parts = append(parts, fmt.Sprintf("%d years", years))
        }
    }
    if months > 0 {
        if months == 1 {
            parts = append(parts, "1 month")
        } else {
            parts = append(parts, fmt.Sprintf("%d months", months))
        }
    }
    if weeks > 0 {
        if weeks == 1 {
            parts = append(parts, "1 week")
        } else {
            parts = append(parts, fmt.Sprintf("%d weeks", weeks))
        }
    }
    if days > 0 {
        if days == 1 {
            parts = append(parts, "1 day")
        } else {
            parts = append(parts, fmt.Sprintf("%d days", days))
        }
    }
    if hours > 0 {
        if hours == 1 {
            parts = append(parts, "1 hour")
        } else {
            parts = append(parts, fmt.Sprintf("%d hours", hours))
        }
    }
    if minutes > 0 || len(parts) == 0 {
        if minutes == 1 {
            parts = append(parts, "1 minute")
        } else {
            parts = append(parts, fmt.Sprintf("%d minutes", minutes))
        }
    }
    
    return strings.Join(parts, ", ")
}

// Update the handleMultiAssistantConversation function
func handleMultiAssistantConversation(config ConversationConfig) error {
    if len(config.Assistants) < 2 {
        return fmt.Errorf("at least two assistants are required for a conversation")
    }

    // Check maximum number of assistants
    const maxAssistants = 15
    if len(config.Assistants) > maxAssistants {
        return fmt.Errorf("too many assistants: maximum allowed is %d, but got %d", maxAssistants, len(config.Assistants))
    }

    // Check for duplicate assistants
    seen := make(map[string]bool)
    for _, name := range config.Assistants {
        // Convert to proper case using GetAssistantConfig to ensure consistent comparison
        properName := assistants.GetAssistantConfig(name).Name
        if seen[properName] {
            return fmt.Errorf("duplicate assistant detected: %s (each assistant can only be included once)", properName)
        }
        seen[properName] = true
    }

    // Validate all assistants exist
    assistantConfigs := make([]assistants.AssistantConfig, len(config.Assistants))
    for i, name := range config.Assistants {
        if !assistants.IsValidAssistant(name) {
            return fmt.Errorf("invalid assistant name: %s", name)
        }
        assistantConfigs[i] = assistants.GetAssistantConfig(name)
    }

    fmt.Println() // Add top margin
    fmt.Printf("Starting conversation between %d assistants:\n", len(config.Assistants))
    for i, assistant := range assistantConfigs {
        fmt.Printf("%d. %s %s\n", i+1, assistant.Emoji, assistant.Name)
    }
    fmt.Println() // Single line margin at bottom

    currentMessage := config.Starter
    currentTurn := 1
    lastSpeaker := "User"
    firstMessage := true

    // Initialize conversation histories
    var conversationLog strings.Builder
    conversationLog.WriteString(fmt.Sprintf("👤 User: %s\n", config.Starter))

    // Initialize conversation histories for each assistant
    histories := make([][]Message, len(assistantConfigs))
    for i, assistant := range assistantConfigs {
        // Initialize with system message and conversation context
        histories[i] = []Message{{
            Role:    "system",
            Content: assistant.GetFullSystemMessage(),
        }}
    }

    // Create a reader for user input
    reader := bufio.NewReader(os.Stdin)

    // Initialize conversation state
    state := ConversationState{
        startTime: time.Now(),
        lastActive: time.Now(),
    }

    for {
        // Update last active time
        state.lastActive = time.Now()

        // Print turn header with improved structure
        elapsed := formatElapsedTime(state.startTime, state.lastActive)
        
        // Print top separator
        fmt.Printf("%s%s%s\n", turnSeparatorColor, strings.Repeat("─", 60), colorReset)
        
        // Print turn information
        fmt.Printf("%s%s Conversation Turn%s %s%d%s\n",
            turnColor, turnEmoji, colorReset,
            turnNumberColor, currentTurn, colorReset)
        
        // Print time information with better structure and local timezone
        localStartTime := state.startTime.Local()
        
        // Format timezone as GMT±X
        _, offset := localStartTime.Zone()
        var gmtOffset string
        if offset >= 0 {
            gmtOffset = fmt.Sprintf("GMT+%d", offset/3600)
        } else {
            gmtOffset = fmt.Sprintf("GMT%d", offset/3600)
        }
        
        // Fix the formatting strings to match the number of arguments
        fmt.Printf("%s%sStarted:%s %s%s %s\n",
            timeHeaderColor,
            timeHeaderColor, colorReset,
            timeValueColor,
            localStartTime.Format("2006-01-02 15:04:05 ")+gmtOffset,
            colorReset)
            
        fmt.Printf("%s%sElapsed:%s %s%s%s\n",
            timeHeaderColor,
            timeHeaderColor, colorReset,
            timeValueColor, elapsed,
            colorReset)
            
        // Print bottom separator and add extra line break
        fmt.Printf("%s%s%s\n\n", turnSeparatorColor, strings.Repeat("─", 60), colorReset)

        // Print the user's message at the start of each turn (after the turn title)
        if firstMessage {
            fmt.Println(colorize(formatUserMessage(config.Starter), "\033[1;36m"))
            firstMessage = false
        } else {
            fmt.Println(colorize(formatUserMessage(currentMessage), "\033[1;36m"))
        }

        for i, assistant := range assistantConfigs {
            // Print margins
            for i := 0; i < converseMargin; i++ {
                fmt.Println()
            }
            
            // Start animation with correct assistant
            anim := startConversationAnimation(assistant)

            // Build participants list excluding current assistant
            var participants strings.Builder
            for j, other := range assistantConfigs {
                if j != i {  // Skip current assistant
                    participants.WriteString(fmt.Sprintf("%d. %s (%s) - %s\n", 
                        j+1, 
                        other.Name, 
                        other.Emoji,
                        other.Description))
                }
            }
            participants.WriteString(fmt.Sprintf("%d. User (👤) - Human participant guiding the conversation\n", 
                len(assistantConfigs)))

            // Get recent conversation history
            recentHistory := getRecentConversationHistory(conversationLog.String())

            // Create conversation context with identity reinforcement
            context := fmt.Sprintf(conversationContextTemplate,
                assistant.Name,
                assistant.Emoji,
                assistant.Name,
                participants.String(),
                recentHistory,  // Use recent history instead of full history
                lastSpeaker,
                currentMessage,
                assistant.Name)

            // Reset this assistant's history to keep context minimal
            histories[i] = []Message{{
                Role:    "system",
                Content: assistant.GetFullSystemMessage(),
            }}

            // Add only the current context
            histories[i] = append(histories[i], Message{
                Role:    "user",
                Content: context + "\n\nYour response:",
            })

            // Prepare the request with full conversation history
            chatReq := ChatRequest{
                Model:    assistants.GetCurrentModel(),
                Messages: histories[i],
                Stream:   true,
            }

            jsonData, err := json.Marshal(chatReq)
            if err != nil {
                anim.stopAnimation()
                return fmt.Errorf("error marshaling request for %s: %v", assistant.Name, err)
            }

            // Make the API request with retry
            resp, err := makeAPIRequestWithRetry(jsonData, histories[i], assistant.Name, true)
            if err != nil {
                anim.stopAnimation()
                return fmt.Errorf("error making request for %s: %v", assistant.Name, err)
            }

            // Process the response
            fullResponseText, err := processStreamResponse(resp, anim, true)
            if err != nil {
                return fmt.Errorf("error processing response from %s: %v", assistant.Name, err)
            }

            // Add the assistant's response to their history
            histories[i] = append(histories[i], Message{
                Role:    "assistant",
                Content: fullResponseText,
            })

            // Update conversation log
            conversationLog.WriteString(fmt.Sprintf("%s %s: %s\n", 
                assistant.Emoji, 
                assistant.Name, 
                fullResponseText))

            // Update for next iteration
            currentMessage = fullResponseText
            lastSpeaker = assistant.Name

            // Print margins
            for i := 0; i < converseMargin; i++ {
                fmt.Println()
            }

            // If this is the last assistant in the turn
            if i == len(assistantConfigs)-1 {
                // Check if we should continue
                if config.Turns > 0 && currentTurn >= config.Turns {
                    fmt.Printf("\nConversation completed after %d turns.\n", config.Turns)
                    return nil
                }

                // Print margins and input prompt
                fmt.Println()  // Single blank line before input prompt
                fmt.Printf("%sType your message:%s\n", inputPromptColor, colorReset)
                fmt.Printf("%s[Press Enter with empty message to end the conversation]%s\n", inputHintColor, colorReset)
                fmt.Print(colorize("👤 User: ", "\033[1;36m"))

                newMessage, err := reader.ReadString('\n')
                if err != nil {
                    return fmt.Errorf("error reading input: %v", err)
                }

                // Trim whitespace and update current message
                currentMessage = strings.TrimSpace(newMessage)
                if currentMessage == "" {
                    fmt.Printf("\n%sConversation ended after %s%s\n",
                        elapsedTimeColor,
                        formatElapsedTime(state.startTime, time.Now()),
                        colorReset)
                    return nil
                }

                // Update conversation log with user's message
                conversationLog.WriteString(fmt.Sprintf("👤 User: %s\n", currentMessage))

                lastSpeaker = "User"
                currentTurn++
            }
        }
    }
}

// Update the processStreamResponse function to use the conversation animation
func processStreamResponse(resp *http.Response, anim interface{}, isConversation bool) (string, error) {
    var fullResponse strings.Builder
    var firstChunk bool = true
    
    // Create a buffered reader for better performance
    reader := bufio.NewReaderSize(resp.Body, 64*1024)
    
    // Create a decoder that reads from our buffered reader
    decoder := json.NewDecoder(reader)
    
    // Use appropriate timeout based on mode
    currentReadTimeout := readTimeout
    if isConversation {
        currentReadTimeout = converseReadTimeout
    }

    // Create a timer for chunk timeout
    chunkTimer := time.NewTimer(currentReadTimeout)
    defer chunkTimer.Stop()

    for {
        // Reset timer for next chunk
        chunkTimer.Reset(currentReadTimeout)

        // Create a channel for the decoding operation
        done := make(chan error, 1)
        var streamResp ChatResponse

        // Start decoding in a goroutine
        go func() {
            done <- decoder.Decode(&streamResp)
        }()

        // Wait for either timeout or successful decode
        select {
        case err := <-done:
            if err == io.EOF {
                return fullResponse.String(), nil
            }
            if err != nil {
                return fullResponse.String(), fmt.Errorf("error reading response: %v", err)
            }
        case <-chunkTimer.C:
            return fullResponse.String(), fmt.Errorf("timeout waiting for model response")
        }
        
        // Process the chunk
        if firstChunk {
            // Handle both animation types
            switch a := anim.(type) {
            case *Animation:
                a.stopAnimation()
            case *ConversationAnimation:
                a.stopAnimation()
            }
            firstChunk = false
        }
        
        // Print the response chunk with the appropriate color
        var textColor string
        switch a := anim.(type) {
        case *Animation:
            textColor = currentAssistant.TextColor
        case *ConversationAnimation:
            textColor = a.assistant.TextColor
        }
        
        fmt.Print(colorize(streamResp.Message.Content, textColor))
        fullResponse.WriteString(streamResp.Message.Content)
        
        if streamResp.Done {
            return fullResponse.String(), nil
        }
    }
}

// Check if chatty is initialized
func isChattyInitialized() bool {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return false
    }
    
    chattyDir := filepath.Join(homeDir, historyDir)
    if _, err := os.Stat(chattyDir); os.IsNotExist(err) {
        return false
    }
    
    return true
}

// Initialize chatty environment
func initializeChatty() error {
    fmt.Println("Initializing Chatty environment...")
    
    // Create necessary directories and files
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return fmt.Errorf("failed to get home directory: %v", err)
    }

    // Create .chatty directory
    chattyDir := filepath.Join(homeDir, historyDir)
    if err := os.MkdirAll(chattyDir, 0755); err != nil {
        return fmt.Errorf("failed to create chatty directory: %v", err)
    }
    fmt.Println("✓ Created ~/.chatty directory")

    // Initialize assistants
    if err := assistants.CreateDefaultConfig(); err != nil {
        return fmt.Errorf("failed to create default config: %v", err)
    }
    fmt.Println("✓ Created default configuration")

    // Create assistants directory
    assistantsDir := filepath.Join(chattyDir, "assistants")
    if err := os.MkdirAll(assistantsDir, 0755); err != nil {
        return fmt.Errorf("failed to create assistants directory: %v", err)
    }
    fmt.Println("✓ Created assistants directory")

    // Copy sample assistants
    if err := assistants.CopySampleAssistants(); err != nil {
        fmt.Printf("Warning: Failed to copy sample assistants: %v\n", err)
    } else {
        fmt.Println("✓ Copied sample assistant configurations")
    }

    fmt.Println("\nChatty has been successfully initialized!")
    fmt.Println("\nYou can now:")
    fmt.Println("1. List available assistants:   chatty --list")
    fmt.Println("2. Select an assistant:         chatty --select <name>")
    fmt.Println("3. Start chatting:              chatty \"Your message here\"")
    fmt.Println("\nEnjoy your conversations! 🚀")
    
    return nil
}

// Add this new function before handleMultiAssistantConversation
func makeAssistantRequest(assistant assistants.AssistantConfig, message string) (*http.Response, error) {
    // Initialize history with just the system message for this conversation
    history := []Message{{
        Role:    "system",
        Content: assistant.GetFullSystemMessage(),
    }}

    // Add the current message
    history = append(history, Message{
        Role:    "user",
        Content: message,
    })

    // Prepare the request
    chatReq := ChatRequest{
        Model:    assistants.GetCurrentModel(),
        Messages: history,
        Stream:   true,
    }

    jsonData, err := json.Marshal(chatReq)
    if err != nil {
        return nil, fmt.Errorf("error marshaling request: %v", err)
    }

    // Make the API request
    return makeAPIRequest(jsonData, history, false)
}

// Update the main() function to handle the new command
func main() {
    // Check if this is the init command
    if len(os.Args) > 1 && os.Args[1] == "init" {
        if isChattyInitialized() {
            fmt.Println("Chatty is already initialized.")
            fmt.Println("To start over, remove the ~/.chatty directory and run 'chatty init' again.")
            return
        }
        if err := initializeChatty(); err != nil {
            fmt.Printf("Error initializing Chatty: %v\n", err)
            os.Exit(1)
        }
        return
    }

    // For all other commands, check if chatty is initialized
    if !isChattyInitialized() {
        fmt.Println("\n🚫 Chatty needs to be initialized before first use!")
        fmt.Println("\nWhat's happening?")
        fmt.Println("   Chatty requires some initial setup to create your personal chat environment.")
        fmt.Println("\n🔧 How to fix this:")
        fmt.Println("   Simply run the following command:")
        fmt.Printf("   %s%s chatty init%s\n", "\033[1;36m", "\033[1m", "\033[0m")
        fmt.Println("\n💡 This will:")
        fmt.Println("   • Create your personal chat directory (~/.chatty)")
        fmt.Println("   • Set up default configurations")
        fmt.Println("   • Install sample AI assistants")
        fmt.Println("   • Prepare everything for your first chat")
        os.Exit(1)
    }

    // Now that we know chatty is initialized, load assistants
    if err := assistants.LoadAssistants(); err != nil {
        fmt.Printf("Error loading assistants: %v\n", err)
        os.Exit(1)
    }

    // Load configuration at startup
    config, err := assistants.GetCurrentConfig()
    if err != nil {
        fmt.Printf("Error loading config: %v\n", err)
        // Continue with default assistant
    } else {
        // Set current assistant from config
        currentAssistant = assistants.GetAssistantConfig(config.CurrentAssistant)
    }

    if len(os.Args) < 2 {
        fmt.Println("Usage: chatty \"Your message here\"")
        fmt.Println("Special commands:")
        fmt.Println("  init                          Initialize Chatty environment")
        fmt.Println("  --clear [all|assistant_name]  Clear chat history (all or specific assistant)")
        fmt.Println("  --list                       List available assistants")
        fmt.Println("  --select <assistant_name>    Select an assistant")
        fmt.Println("  --current                    Show current assistant")
        fmt.Println("  --converse <assistants...>   Start a conversation between assistants")
        fmt.Println("      --starter \"message\"      Initial message to start the conversation")
        fmt.Println("      --turns N                Number of conversation turns (default: infinite)")
        return
    }

    // Handle special commands
    switch os.Args[1] {
    case "--converse":
        if len(os.Args) < 4 {
            fmt.Println("Usage: chatty --converse <assistant1> <assistant2> [assistant3...] --starter \"message\" [--turns N]")
            fmt.Println("\nNote: The --starter argument must be enclosed in double quotes to preserve special characters.")
            return
        }

        // Parse arguments
        var config ConversationConfig
        var starter string
        var turns int
        var starterIndex int
        var foundStarterArg bool

        // Find the --starter argument
        for i := 2; i < len(os.Args); i++ {
            if os.Args[i] == "--starter" {
                starterIndex = i
                foundStarterArg = true
                // Check if next argument exists
                if i+1 >= len(os.Args) {
                    fmt.Println("Error: --starter argument is missing")
                    fmt.Println("\nUsage: --starter \"your message here\"")
                    return
                }
                // Take the next argument as is, without checking for quotes
                starter = os.Args[i+1]
                break
            }
        }

        if !foundStarterArg {
            fmt.Println("Error: --starter argument is required")
            fmt.Println("\nUsage: --starter \"your message here\"")
            return
        }

        if starter == "" {
            fmt.Println("Error: --starter argument cannot be empty")
            fmt.Println("\nUsage: --starter \"your message here\"")
            return
        }

        // Find the --turns argument
        for i := starterIndex + 1; i < len(os.Args); i++ {
            if os.Args[i] == "--turns" {
                if i+1 < len(os.Args) {
                    turns, err = strconv.Atoi(os.Args[i+1])
                    if err != nil {
                        fmt.Printf("Error: invalid turns value: %v\n", err)
                        return
                    }
                    break
                }
            }
        }

        // Collect assistant names (all arguments between --converse and --starter)
        for i := 2; i < len(os.Args); i++ {
            if os.Args[i] == "--starter" {
                break
            }
            config.Assistants = append(config.Assistants, os.Args[i])
        }

        config.Starter = starter
        config.Turns = turns

        if err := handleMultiAssistantConversation(config); err != nil {
            fmt.Printf("Error: %v\n", err)
            return
        }
        return
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
        if err := assistants.UpdateCurrentAssistant(os.Args[2]); err != nil {
            fmt.Printf("Error saving assistant selection: %v\n", err)
            return
        }
        fmt.Printf("Switched to %s [%s%s%s] %s\n", 
            currentAssistant.Emoji,
            currentAssistant.LabelColor,
            currentAssistant.Name,
            "\u001b[0m", // Reset color
            currentAssistant.Description)
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
        Model:    assistants.GetCurrentModel(),
        Messages: history,
        Stream:   true,
    }

    jsonData, err := json.Marshal(chatReq)
    if err != nil {
        fmt.Printf("Error marshaling request: %v\n", err)
        return
    }

    // Print top margin
    printChatMargin(chatTopMargin)

    // Show assistant label immediately
    fmt.Printf("%s", colorize(getAssistantLabel(), currentAssistant.LabelColor))

    // Start the animation before making the request
    anim := startAnimation()

    // Make the API request with timeout (passing false for regular chat)
    resp, err := makeAPIRequest(jsonData, history, false)
    if err != nil {
        anim.stopAnimation()
        fmt.Printf("\nError: %v\n", err)
        if strings.Contains(err.Error(), "invalid model") {
            fmt.Printf("\nHint: Edit ~/.chatty/config.json to set a valid model name\n")
            fmt.Printf("Available models can be listed with: ollama list\n")
        }
        return
    }
    defer resp.Body.Close()

    // Process the streaming response (passing false for regular chat)
    fullResponseText, err := processStreamResponse(resp, anim, false)
    if err != nil {
        fmt.Printf("\nError: %v\n", err)
        return
    }

    // Ensure we're on a new line before printing margin
    fmt.Println()
    
    // Print bottom margin
    printChatMargin(chatBottomMargin)

    // Add assistant's response to history
    history = append(history, Message{
        Role:    "assistant",
        Content: fullResponseText,
    })

    // Save updated history
    if err := saveHistory(history); err != nil {
        fmt.Printf("Error saving history: %v\n", err)
    }
} 