package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"chatty/cmd/chatty/agents"
	"chatty/cmd/chatty/builder"
	"chatty/cmd/chatty/store"
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
    Agents []string
    Starter    string
    Turns      int  // 0 means infinite
    Current    int  // Current turn
    AutoMode   bool // If true, agents converse among themselves without user input
    SaveFile   string // Path to save conversation log
}

// Add this new type for conversation history
type ConversationHistory struct {
    Messages []Message
}

// Add this new animation type that includes agent info
type ConversationAnimation struct {
    stopChan chan bool
    agent agents.AgentConfig
}

const (
    // Core configuration
    ollamaBaseURL = "http://localhost:11434"  // Base URL for Ollama API
    ollamaURLPath = "/api/chat"              // API endpoint path
    historyDir    = ".chatty"               // Directory to store chat histories
    configFile    = "config.json"           // File to store current agent selection

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
                fmt.Printf("\r%s%s", colorize(getAgentLabel(), currentAgent.LabelColor), frames[frameIndex])
                
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
    fmt.Printf("\r%s", colorize(getAgentLabel(), currentAgent.LabelColor))
}

type Config struct {
    CurrentAgent string `json:"current_agent"`
}

// Current agent configuration
var currentAgent = agents.DefaultAgent

// Get system message using agent name
func getSystemMessage() string {
    // Single agent chat is never in auto mode
    return currentAgent.GetFullSystemMessage(false, "")
}

// Format text with color if enabled
func colorize(text, color string) string {
    if useColors {
        return color + text + colorReset
    }
    return text
}

// Get formatted agent label with optional emoji
func getAgentLabel() string {
    label := currentAgent.Name
    if useEmoji {
        return currentAgent.Emoji + " " + label + ": "
    }
    return label + ": "
}

// Get the history file path for a specific agent
func getHistoryPathForAgent(agentName string) (string, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    
    // Get the history file path from the agents package
    historyFile := agents.GetHistoryFileName(agentName)
    return filepath.Join(homeDir, historyFile), nil
}

// Get the history file path for the current agent
func getHistoryPath() (string, error) {
    return getHistoryPathForAgent(currentAgent.Name)
}

// Cache for chat histories
var historyCache = make(map[string][]Message)

// loadHistory loads chat history with caching
func loadHistory() ([]Message, error) {
    // Check cache first
    if history, exists := historyCache[currentAgent.Name]; exists {
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
        historyCache[currentAgent.Name] = history
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
        historyCache[currentAgent.Name] = history
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
    } else if !strings.Contains(history[0].Content, currentAgent.Name) {
        // Update system message if it's for a different agent
        history[0] = Message{
            Role:    "system",
            Content: getSystemMessage(),
        }
    }

    // Cache the loaded history
    historyCache[currentAgent.Name] = history
    return history, nil
}

// saveHistory saves chat history and updates cache
func saveHistory(history []Message) error {
    // Update cache
    historyCache[currentAgent.Name] = history

    // Ensure directory exists before trying to save
    historyPath, err := getHistoryPath()
    if err != nil {
        return err
    }

    // Create directory if it doesn't exist
    dir := filepath.Dir(historyPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("failed to create history directory: %v", err)
    }

    // Save the history file
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
                fmt.Println("No chat histories found. Fresh conversations will be started for each agent.")
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
            fmt.Println("All chat histories have been cleared. Fresh conversations will be started for each agent.")
        } else {
            fmt.Println("No chat histories found. Fresh conversations will be started for each agent.")
        }
        return nil
    }

    // Clear specific agent's history
    if !agents.IsValidAgent(target) {
        return fmt.Errorf("invalid agent name: %s", target)
    }

    // Get proper case for agent name
    agentConfig := agents.GetAgentConfig(target)
    properName := agentConfig.Name

    // Clear cache for this agent
    delete(historyCache, properName)

    historyPath, err := getHistoryPathForAgent(target)
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

// Update the animation functions for conversation mode
func startConversationAnimation(agent agents.AgentConfig) *ConversationAnimation {
    anim := &ConversationAnimation{
        stopChan: make(chan bool),
        agent: agent,
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
                // Clear line and print current frame with correct agent label
                label := fmt.Sprintf("%s %s: ", agent.Emoji, agent.Name)
                fmt.Printf("\r%s%s", colorize(label, agent.LabelColor), frames[frameIndex])
                
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
    // Clear the animation and prepare for response with correct agent label
    label := fmt.Sprintf("%s %s: ", a.agent.Emoji, a.agent.Name)
    fmt.Printf("\r%s", colorize(label, a.agent.LabelColor))
}

// Add this new function at the top level
func checkOllamaReady() error {
    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Get(ollamaBaseURL + "/api/tags")
    if err != nil {
        if os.IsTimeout(err) || strings.Contains(err.Error(), "connection refused") {
            return fmt.Errorf("ollama is not ready. please ensure 'ollama serve' is running and the service is fully initialized")
        }
        return fmt.Errorf("error checking ollama: %v", err)
    }
    defer resp.Body.Close()
    return nil
}

// Add global signal channel
var (
    debugMode bool
    globalStopChan = make(chan os.Signal, 1)
)

// Update the makeAPIRequestWithRetry function
func makeAPIRequestWithRetry(jsonData []byte, agent string) (*http.Response, error) {
    // First, check if Ollama is ready
    if err := checkOllamaReady(); err != nil {
        return nil, err
    }

    var lastErr error
    retryDelay := initialRetryDelay

    for attempt := 1; attempt <= maxRetries; attempt++ {
        // Show retry attempt if not first try
        if attempt > 1 {
            fmt.Printf("\nRetrying request for %s (attempt %d/%d)...\n", agent, attempt, maxRetries)
        }

        resp, err := makeAPIRequest(jsonData)
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

            // Wait for either the delay to complete or an interrupt
            select {
            case <-time.After(retryDelay):
                continue
            case <-globalStopChan:
                return nil, fmt.Errorf("interrupted")
            }
        }
    }

    return nil, fmt.Errorf("after %d attempts: %v", maxRetries, lastErr)
}

// Update the makeAPIRequest function
func makeAPIRequest(jsonData []byte) (*http.Response, error) {
    // Print request JSON in debug mode
    if debugMode {
        // Pretty print the JSON with indentation
        var prettyJSON bytes.Buffer
        if err := json.Indent(&prettyJSON, jsonData, "", "    "); err != nil {
            fmt.Printf("Error formatting JSON: %v\n", err)
            return nil, err
        }
        
        // Print with colors and formatting
        fmt.Printf("\n%sDebug: Request JSON:%s\n", 
            "\033[38;5;208m", // Orange color for debug
            colorReset)
        fmt.Printf("%s%s%s\n\n",
            "\033[38;5;39m", // Light blue for JSON
            prettyJSON.String(),
            colorReset)
    }

    // Create the request
    req, err := http.NewRequest("POST", getOllamaAPI(), bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("error creating request: %v", err)
    }
    req.Header.Set("Content-Type", "application/json")

    // Create a custom transport with optimized settings for streaming
    transport := &http.Transport{
        MaxIdleConns:        maxIdleConns,
        MaxConnsPerHost:     maxConnsPerHost,
        DisableCompression:  false,
        DisableKeepAlives:   false,
        ForceAttemptHTTP2:   true,
        MaxIdleConnsPerHost: maxConnsPerHost,
        WriteBufferSize:     64 * 1024,
        ReadBufferSize:      64 * 1024,
    }

    // Create client with no timeout
    client := &http.Client{
        Transport: transport,
    }

    // Make the request
    resp, err := client.Do(req)
    if err != nil {
        if strings.Contains(err.Error(), "connection refused") {
            return nil, fmt.Errorf("could not connect to Ollama - make sure 'ollama serve' is running")
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
                return nil, fmt.Errorf("invalid model '%s' - please check your config.json file", agents.GetCurrentModel())
            }
            return nil, fmt.Errorf("API error: %s", errorResponse.Error)
        }
        
        return nil, fmt.Errorf("API error (status %d): failed to process request", resp.StatusCode)
    }

    return resp, nil
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

// Update the handleMultiAgentConversation function to format participants list without newlines
func handleMultiAgentConversation(config ConversationConfig) error {
    // Validate configuration
    if len(config.Agents) < 2 {
        return fmt.Errorf("at least two agents are required for a conversation")
    }

    // Check maximum number of agents
    const maxAgents = 15
    if len(config.Agents) > maxAgents {
        return fmt.Errorf("too many agents: maximum allowed is %d, but got %d", maxAgents, len(config.Agents))
    }

    // Check for duplicate agents
    seen := make(map[string]bool)
    for _, name := range config.Agents {
        // Convert to proper case using GetAgentConfig to ensure consistent comparison
        properName := agents.GetAgentConfig(name).Name
        if seen[properName] {
            return fmt.Errorf("duplicate agent detected: %s (each agent can only be included once)", properName)
        }
        seen[properName] = true
    }

    // Validate all agents exist
    for _, name := range config.Agents {
        if !agents.IsValidAgent(name) {
            return fmt.Errorf("invalid agent name: %s", name)
        }
    }

    // Load agent configurations
    agentConfigs := make([]agents.AgentConfig, 0, len(config.Agents))
    for _, agentName := range config.Agents {
        agentConfigs = append(agentConfigs, agents.GetAgentConfig(agentName))
    }

    // Set up colors and emojis for the UI
    turnSeparatorColor := "\033[1;36m" // Black
    turnColor := "\033[1;36m"          // Yellow
    turnNumberColor := "\033[1;37m"    // Cyan
    timeHeaderColor := "\033[1;32m"    // Green
    timeValueColor := "\033[1;37m"     // White
    elapsedTimeColor := "\033[1;35m"   // Magenta
    inputPromptColor := "\033[1;36m"   // Cyan
    inputHintColor := "\033[1;30m"     // Gray
    colorReset := "\033[0m"
    turnEmoji := "🔄"
    converseMargin := 1

    // We don't need to print the welcome message and participants list here
    // as it's already printed in the main function before calling this function

    currentMessage := config.Starter
    currentTurn := 1
    firstMessage := true

    // Initialize conversation histories
    var conversationLog strings.Builder
    conversationLog.WriteString(fmt.Sprintf("👤 User: %s\n", config.Starter))

    // Initialize conversation histories for each agent
    histories := make([][]Message, len(agentConfigs))
    for i, agent := range agentConfigs {
        // Initialize with system message and conversation context
        histories[i] = []Message{{
            Role:    "system",
            Content: agent.GetFullSystemMessage(config.AutoMode, ""),
        }}
    }

    // Add the initial user message to all agent histories
    for i := range histories {
        histories[i] = append(histories[i], Message{
            Role:    "user",
            Content: config.Starter,
        })
    }

    // Create a reader for user input (only used in non-auto mode)
    var reader *bufio.Reader
    if !config.AutoMode {
        reader = bufio.NewReader(os.Stdin)
    }

    // Initialize conversation state
    state := ConversationState{
        startTime: time.Now(),
        lastActive: time.Now(),
    }

    // Maximum number of messages to keep in history per agent
    const maxMessagesPerAgent = 20

    if config.AutoMode {
        fmt.Println("\n🤖 Auto-conversation mode enabled. Press Ctrl+C to stop.")
    }

    // Create a shared conversation history that will be used to build each agent's history
    sharedHistory := []Message{
        {
            Role:    "user",
            Content: config.Starter,
        },
    }

    for {
        // Update last active time
        state.lastActive = time.Now()

        // Check for stop signal before starting a new turn
        select {
        case <-globalStopChan:
            fmt.Printf("\n\nConversation ended after %s\n",
                formatElapsedTime(state.startTime, time.Now()))
            os.Exit(0)
        default:
        }

        // Print turn header with improved structure
        elapsed := formatElapsedTime(state.startTime, state.lastActive)
        
        // Add single blank line before the top separator
        fmt.Println()
        
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
            
        // Print bottom separator
        fmt.Printf("%s%s%s\n", turnSeparatorColor, strings.Repeat("─", 60), colorReset)

        // In auto mode, print the user's message
        if !config.AutoMode {
            fmt.Println(colorize(formatUserMessage(currentMessage), "\033[1;36m"))
            fmt.Println()  // Single blank line after user message
        } else {
            fmt.Println()  // Single blank line after separator for auto mode
        }

        // Update conversation log with user message first (except for the first turn)
        if firstMessage {
            // First message is already in the log from initialization
            firstMessage = false
            
            // Print the first message in auto mode
            if config.AutoMode {
                fmt.Println(colorize(formatUserMessage(currentMessage), "\033[1;36m"))
                fmt.Println()  // Single blank line after user message
            }
        } else {
            // For non-auto mode, add the user's new message to the conversation log and shared history
            if !config.AutoMode {
                // Add user message to conversation log
                conversationLog.WriteString(fmt.Sprintf("👤 User: %s\n", currentMessage))
                
                // Add the user message to the shared history
                sharedHistory = append(sharedHistory, Message{
                    Role:    "user",
                    Content: currentMessage,
                })
            }
            // For auto mode, we don't add a new user message after the first turn
        }

        // Process each agent's response in this turn
        for i, agent := range agentConfigs {
            // Check for stop signal before each agent's response
            select {
            case <-globalStopChan:
                fmt.Printf("\n\nConversation ended after %s\n",
                    formatElapsedTime(state.startTime, time.Now()))
                os.Exit(0)
            default:
            }

            // In auto mode, only add margin between agent responses
            if i > 0 {
                for i := 0; i < converseMargin; i++ {
                    fmt.Println()
                }
            }
            
            // Start animation with correct agent
            anim := startConversationAnimation(agent)

            // Build participants list excluding current agent
            var participants strings.Builder
            for j, other := range agentConfigs {
                if j != i {  // Skip current agent
                    if j > 0 {
                        participants.WriteString(" ")
                    }
                    participants.WriteString(fmt.Sprintf("%d. %s (%s) - %s", 
                        j+1, 
                        other.Name, 
                        other.Emoji,
                        other.Description))
                }
            }
            if !config.AutoMode {
                if participants.Len() > 0 {
                    participants.WriteString(" ")
                }
                participants.WriteString(fmt.Sprintf("%d. User (👤) - Human participant guiding the conversation", 
                    len(agentConfigs)))
            }

            // Update the system message with the participants list
            histories[i][0] = Message{
                Role:    "system",
                Content: agent.GetFullSystemMessage(config.AutoMode, participants.String()),
            }

            // Build this agent's history from the shared history
            agentHistory := []Message{histories[i][0]} // Start with the system message
            
            // Add all messages from the shared history
            agentHistory = append(agentHistory, sharedHistory...)
            
            // Add a final user message instructing the agent to respond naturally
            // Check if the last message is already an instruction
            lastMsgIndex := len(agentHistory) - 1
            
            // Always add an instruction message to ensure consistent behavior
            if lastMsgIndex >= 0 && agentHistory[lastMsgIndex].Role == "user" && 
               (agentHistory[lastMsgIndex].Content == "Please respond naturally as part of this conversation." ||
                agentHistory[lastMsgIndex].Content == "Respond naturally as part of this conversation and do not add prefixes like '/<Your name/> said:' to your messages.") {
                // Replace the last message with our instruction
                agentHistory[lastMsgIndex] = Message{
                    Role:    "user",
                    Content: "Respond naturally as part of this conversation and do not add prefixes like '/<Your name/> said:' to your messages.",
                }
            } else {
                // Add a new instruction message
                agentHistory = append(agentHistory, Message{
                    Role:    "user",
                    Content: "Respond naturally as part of this conversation and do not add prefixes like '</Your name/> said:' to your messages.",
                })
            }

            // Ensure we don't exceed the token limit
            if len(agentHistory) > maxMessagesPerAgent {
                // Keep the system message and the most recent messages
                agentHistory = append([]Message{agentHistory[0]}, agentHistory[len(agentHistory)-maxMessagesPerAgent+1:]...)
            }

            // Prepare the request with full conversation history
            chatReq := ChatRequest{
                Model:    agents.GetCurrentModel(),
                Messages: agentHistory,
                Stream:   true,
            }

            jsonData, err := json.Marshal(chatReq)
            if err != nil {
                anim.stopAnimation()
                return fmt.Errorf("error marshaling request for %s: %v", agent.Name, err)
            }

            // Make the API request with retry
            resp, err := makeAPIRequestWithRetry(jsonData, agent.Name)
            if err != nil {
                anim.stopAnimation()
                return fmt.Errorf("error making request for %s: %v", agent.Name, err)
            }

            // Process the response
            fullResponseText, err := processStreamResponse(resp, anim)
            if err != nil {
                // Check if this was due to a stop signal
                if config.AutoMode {
                    select {
                    case <-globalStopChan:
                        fmt.Printf("\n\nConversation ended after %s\n",
                            formatElapsedTime(state.startTime, time.Now()))
                        os.Exit(0)
                    default:
                    }
                }
                return fmt.Errorf("error processing response from %s: %v", agent.Name, err)
            }

            // Update conversation log
            conversationLog.WriteString(fmt.Sprintf("%s %s: %s\n", 
                agent.Emoji, 
                agent.Name, 
                fullResponseText))

            // Add the agent's response to the shared history
            sharedHistory = append(sharedHistory, Message{
                Role:    "assistant",
                Content: fmt.Sprintf("%s said: %s", agent.Name, fullResponseText),
            })

            // In auto mode, the last agent's response becomes the prompt for the next turn
            if config.AutoMode && i == len(agentConfigs)-1 {
                currentMessage = fullResponseText
            }

            // Print margins
            for i := 0; i < converseMargin; i++ {
                fmt.Println()
            }

            // If this is the last agent in the turn
            if i == len(agentConfigs)-1 {
                // Add extra blank line before next turn separator
                fmt.Println()
                
                // Check if we should continue
                if config.Turns > 0 && currentTurn >= config.Turns {
                    fmt.Printf("\nConversation completed after %d turns.\n", config.Turns)
                    if config.SaveFile != "" {
                        if err := saveConversationLog(config.SaveFile, conversationLog.String()); err != nil {
                            fmt.Printf("Warning: Failed to save conversation log: %v\n", err)
                        } else {
                            fmt.Printf("Conversation log saved to: %s\n", config.SaveFile)
                        }
                    }
                    return nil
                }

                if config.AutoMode {
                    // Add a small delay between turns in auto mode
                    time.Sleep(2 * time.Second)
                    currentTurn++
                    continue
                }

                // Print margins and input prompt for non-auto mode
                fmt.Println()  // Single blank line before input prompt
                fmt.Printf("%sType your message:%s\n", inputPromptColor, colorReset)
                fmt.Printf("%s[Press Enter with empty message to end the conversation]%s\n", inputHintColor, colorReset)
                fmt.Print(colorize("👤 User: ", "\033[1;36m"))
                
                // Read user input
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
                    if config.SaveFile != "" {
                        if err := saveConversationLog(config.SaveFile, conversationLog.String()); err != nil {
                            fmt.Printf("Warning: Failed to save conversation log: %v\n", err)
                        } else {
                            fmt.Printf("Conversation log saved to: %s\n", config.SaveFile)
                        }
                    }
                    return nil
                }

                // Update conversation log
                conversationLog.WriteString(fmt.Sprintf("👤 User: %s\n", currentMessage))
                
                // Add user message to history
                history := append(histories[i], Message{
                    Role:    "user",
                    Content: currentMessage,
                })
                
                // Update history for this agent
                histories[i] = history
                
                fmt.Println()  // Single blank line after user input
                
                currentTurn++
            }
        }
    }
}

// Update the processStreamResponse function to use the conversation animation
func processStreamResponse(resp *http.Response, anim any) (string, error) {
    var fullResponse strings.Builder
    var firstChunk bool = true
    
    // Create a buffered reader for better performance
    reader := bufio.NewReaderSize(resp.Body, 64*1024)
    
    // Create a decoder that reads from our buffered reader
    decoder := json.NewDecoder(reader)

    for {
        // Check for interrupt
        select {
        case <-globalStopChan:
            return fullResponse.String(), fmt.Errorf("interrupted")
        default:
        }

        var streamResp ChatResponse
        err := decoder.Decode(&streamResp)
        
        if err == io.EOF {
            return fullResponse.String(), nil
        }
        if err != nil {
            return fullResponse.String(), fmt.Errorf("error reading response: %v", err)
        }

        // Handle first chunk animation
        if firstChunk {
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
            textColor = currentAgent.TextColor
        case *ConversationAnimation:
            textColor = a.agent.TextColor
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
    fmt.Println("\n🚀 Initializing Chatty environment...")
    
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
    fmt.Printf("%s✓%s Created %s~/.chatty%s directory\n", 
        "\033[32m", colorReset,  // Green checkmark
        "\033[1;34m", colorReset) // Blue path

    // Initialize agents
    if err := agents.CreateDefaultConfig(); err != nil {
        return fmt.Errorf("failed to create default config: %v", err)
    }
    fmt.Printf("%s✓%s Created default configuration\n", 
        "\033[32m", colorReset)

    // Create agents directory
    agentsDir := filepath.Join(chattyDir, "agents")
    if err := os.MkdirAll(agentsDir, 0755); err != nil {
        return fmt.Errorf("failed to create agents directory: %v", err)
    }
    fmt.Printf("%s✓%s Created agents directory\n", 
        "\033[32m", colorReset)

    // Get default agent info
    defaultAgent := agents.DefaultAgent

    // Define color constants for better readability
    colorMagenta := "\033[1;35m"
    colorCyan := "\033[1;36m"
    colorGreen := "\033[32m"
    colorPurple := "\033[1;95m"  // Nice purple color for command labels
    colorBlue := "\033[1;34m"

    // Print success message with enhanced formatting
    fmt.Printf("\n%s🎉 Chatty has been successfully initialized!%s\n\n", 
        colorGreen, colorReset)

    fmt.Printf("%s📌 Default Agent:%s\n", colorCyan, colorReset)
    fmt.Printf("   %s %s%s%s - %s\n\n",
        defaultAgent.Emoji,
        defaultAgent.LabelColor,
        defaultAgent.Name,
        colorReset,
        defaultAgent.Description)

    fmt.Printf("%s🎯 Quick Start Guide%s\n", colorMagenta, colorReset)
    
    // Basic Commands Section
    fmt.Printf("\n   %s💬 Basic Commands:%s\n", colorCyan, colorReset)
    fmt.Printf("   • %sOne-off message:%s chatty %s\"Your message here\"%s\n",
        colorPurple, colorReset, colorBlue, colorReset)
    fmt.Printf("   • %sSimple chat mode:%s chatty --with %s<agent_name>%s\n",
        colorPurple, colorReset, colorBlue, colorReset)
    
    // Agent Management Section
    fmt.Printf("\n   %s🧠 Agent Management:%s\n", colorCyan, colorReset)
    fmt.Printf("   • %sList installed agents:%s chatty --list\n",
        colorPurple, colorReset)
    fmt.Printf("   • %sBrowse store agents:%s chatty --store\n",
        colorPurple, colorReset)
    fmt.Printf("   • %sView agent details:%s chatty --show %s<agent_name>%s\n",
        colorPurple, colorReset, colorBlue, colorReset)
    fmt.Printf("   • %sInstall an agent:%s chatty --install %s<agent_name>%s\n",
        colorPurple, colorReset, colorBlue, colorReset)
    fmt.Printf("   • %sSwitch active agent:%s chatty --select %s<agent_name>%s\n",
        colorPurple, colorReset, colorBlue, colorReset)
    
    // Advanced Features Section
    fmt.Printf("\n   %s🌟 Advanced Features:%s\n", colorCyan, colorReset)
    fmt.Printf("   • %sGroup chat mode:%s chatty --with %s\"Einstein,Ada,Tux\"%s --topic \"Let's talk\"\n",
        colorPurple, colorReset, colorBlue, colorReset)
    fmt.Printf("   • %sRandom agents chat:%s chatty --with-random %s3%s --topic \"Hello\"\n",
        colorPurple, colorReset, colorBlue, colorReset)
    fmt.Printf("   • %sAutonomous mode:%s Add %s--auto%s flag to group chats\n",
        colorPurple, colorReset, colorBlue, colorReset)
    fmt.Printf("   • %sSave conversations:%s Add %s--save filename.txt%s to any command\n",
        colorPurple, colorReset, colorBlue, colorReset)

    fmt.Printf("\n%s💡 Pro Tips:%s\n", colorCyan, colorReset)
    fmt.Printf("   • Use %s--turns N%s to limit conversation length\n",
        colorPurple, colorReset)
    fmt.Printf("   • Press %sCtrl+C%s to stop auto-conversations gracefully\n",
        colorPurple, colorReset)
    fmt.Printf("   • Use %s--topic-file path.txt%s to read topics from a file\n",
        colorPurple, colorReset)
    fmt.Printf("   • Clear history with %s--clear all%s or %s--clear \"Agent Name\"%s\n",
        colorPurple, colorReset, colorPurple, colorReset)

    fmt.Printf("\n%s🌟 Ready to start your AI journey!%s\n\n",
        colorGreen, colorReset)
    
    return nil
}

// Add this new function after the makeAgentRequest function
func getRandomAgents(count int) ([]string, error) {
    if count < 2 || count > 15 {
        return nil, fmt.Errorf("number of agents must be between 2 and 15, got %d", count)
    }

    // Get all available agents
    allAgents := agents.GetAllAgentNames()
    if len(allAgents) < count {
        return nil, fmt.Errorf("not enough agents available: requested %d but only have %d", count, len(allAgents))
    }

    // Create a copy of the slice to shuffle
    shuffled := make([]string, len(allAgents))
    copy(shuffled, allAgents)

    // Fisher-Yates shuffle
    for i := len(shuffled) - 1; i > 0; i-- {
        j := time.Now().UnixNano() % int64(i+1)
        shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
    }

    // Return the first 'count' agents
    return shuffled[:count], nil
}

// readStarterFile reads the contents of a text file to use as a starter message
func readStarterFile(filePath string) (string, error) {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return "", fmt.Errorf("failed to read starter file: %v", err)
    }
    return strings.TrimSpace(string(content)), nil
}

// Add this new function to save conversation logs
func saveConversationLog(logPath string, content string) error {
    // Create directory if it doesn't exist
    dir := filepath.Dir(logPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("failed to create directory: %v", err)
    }

    // Write the content to the file
    if err := os.WriteFile(logPath, []byte(content), 0644); err != nil {
        return fmt.Errorf("failed to save conversation log: %v", err)
    }

    return nil
}

// Add a new function to handle single-agent chat
func handleSingleAgentChat(agentName string, starter string, saveFile string) error {
    // Validate agent exists
    if !agents.IsValidAgent(agentName) {
        return fmt.Errorf("invalid agent name: %s", agentName)
    }
    
    // Get agent configuration
    agent := agents.GetAgentConfig(agentName)
    
    // Print welcome message
    fmt.Printf("\n💬 Chat with %s %s\n", agent.Emoji, agent.Name)
    fmt.Printf("%s\n", agent.Description)

    // Show exit message at the beginning of the chat
    fmt.Println()
    fmt.Printf("Press Enter with empty message to end the conversation")
    fmt.Println("\n")
    
    // Initialize conversation log
    var conversationLog strings.Builder
    if starter != "" {
        conversationLog.WriteString(fmt.Sprintf("👤 User: %s\n", starter))
    }
    
    // Try to load existing history for this agent
    historyPath, err := getHistoryPathForAgent(agentName)
    var history []Message
    
    if err == nil {
        // Check if history file exists
        if _, err := os.Stat(historyPath); err == nil {
            // Read and parse the history file
            data, err := os.ReadFile(historyPath)
            if err == nil {
                var existingHistory []Message
                if err := json.Unmarshal(data, &existingHistory); err == nil && len(existingHistory) > 0 {
                    // Successfully loaded history
                    history = existingHistory
                }
            }
        }
    }
    
    // If no history was loaded, initialize with system message
    if len(history) == 0 {
        history = []Message{{
            Role:    "system",
            Content: agent.GetFullSystemMessage(false, ""),
        }}
    } else {
        // Prepend system message to existing history
        history = append([]Message{{
            Role:    "system",
            Content: agent.GetFullSystemMessage(false, ""),
        }}, history...)
    }
    
    // Add starter message if provided
    currentMessage := starter
    if currentMessage != "" {
        history = append(history, Message{
            Role:    "user",
            Content: currentMessage,
        })
        
        // Print the starter message
        fmt.Println(colorize(formatUserMessage(currentMessage), "\033[1;36m"))
        fmt.Println()  // Single blank line after user message
    }
    
    // Create a reader for user input
    reader := bufio.NewReader(os.Stdin)
    
    // Maximum number of messages to keep in history
    const maxMessagesInHistory = 50
    
    for {
        // If we have a current message, get agent's response
        if currentMessage != "" {
            // Start animation
            anim := startConversationAnimation(agent)
            
            // Add instruction message
            history = append(history, Message{
                Role:    "user",
                Content: "Respond naturally as part of this conversation and do not add prefixes like '<Your name> said:' to your messages.",
            })
            
            // Ensure we don't exceed the token limit
            if len(history) > maxMessagesInHistory {
                // Keep the system message and the most recent messages
                history = append([]Message{history[0]}, history[len(history)-maxMessagesInHistory+1:]...)
            }
            
            // Prepare the request
            chatReq := ChatRequest{
                Model:    agents.GetCurrentModel(),
                Messages: history,
                Stream:   true,
            }
            
            jsonData, err := json.Marshal(chatReq)
            if err != nil {
                anim.stopAnimation()
                return fmt.Errorf("error marshaling request: %v", err)
            }
            
            // Make the API request with retry
            resp, err := makeAPIRequestWithRetry(jsonData, agent.Name)
            if err != nil {
                anim.stopAnimation()
                return fmt.Errorf("error making request: %v", err)
            }
            
            // Process the response
            fullResponseText, err := processStreamResponse(resp, anim)
            if err != nil {
                return fmt.Errorf("error processing response: %v", err)
            }
            
            // Update conversation log
            conversationLog.WriteString(fmt.Sprintf("%s %s: %s\n", 
                agent.Emoji, 
                agent.Name, 
                fullResponseText))
            
            // Add the response to history
            history = append(history, Message{
                Role:    "assistant",
                Content: fullResponseText,
            })
            
            // Remove the instruction message
            if len(history) >= 2 && history[len(history)-2].Role == "user" && 
               history[len(history)-2].Content == "Respond naturally as part of this conversation and do not add prefixes like '<Your name> said:' to your messages." {
                // Remove the instruction message
                history = append(history[:len(history)-2], history[len(history)-1])
            }
            
            fmt.Println()  // Single blank line after response
        }
        
        // Add a blank line before prompting for user input
        fmt.Println()
        
        // Prompt for user input
        fmt.Print(colorize("👤 User: ", "\033[1;36m"))
        
        // Read user input
        newMessage, err := reader.ReadString('\n')
        if err != nil {
            return fmt.Errorf("error reading input: %v", err)
        }
        
        // Trim whitespace
        currentMessage = strings.TrimSpace(newMessage)
        
        // Check for exit
        if currentMessage == "" {
            fmt.Println("\nConversation ended.")
            
            // Save conversation history for the agent
            historyPath, err := getHistoryPathForAgent(agentName)
            if err == nil {
                // Filter out system messages and special instruction messages before saving
                var filteredHistory []Message
                for _, msg := range history {
                    if msg.Role == "system" || 
                       (msg.Role == "user" && msg.Content == "Respond naturally as part of this conversation and do not add prefixes like '<Your name> said:' to your messages.") {
                        continue
                    }
                    filteredHistory = append(filteredHistory, msg)
                }
                
                // Save the filtered history
                data, err := json.MarshalIndent(filteredHistory, "", "    ")
                if err == nil {
                    if err := os.WriteFile(historyPath, data, 0644); err != nil {
                        fmt.Printf("Warning: Failed to save conversation history: %v\n", err)
                    }
                }
            }
            
            // Save conversation log to file if requested
            if saveFile != "" {
                if err := saveConversationLog(saveFile, conversationLog.String()); err != nil {
                    fmt.Printf("Warning: Failed to save conversation log: %v\n", err)
                } else {
                    fmt.Printf("Conversation log saved to: %s\n", saveFile)
                }
            }
            return nil
        }
        
        // Update conversation log
        conversationLog.WriteString(fmt.Sprintf("👤 User: %s\n", currentMessage))
        
        // Add user message to history
        history = append(history, Message{
            Role:    "user",
            Content: currentMessage,
        })
        
        fmt.Println()  // Single blank line after user input
    }
}

// Update the main function to handle the new command
func main() {
    // Set up global signal handler at program start
    signal.Notify(globalStopChan, os.Interrupt, syscall.SIGTERM)
    defer func() {
        signal.Stop(globalStopChan)
        // Force immediate exit
        os.Exit(0)
    }()

    // Add signal handler goroutine
    go func() {
        <-globalStopChan
        fmt.Println("\nInterrupted by user. Exiting...")
        os.Exit(0)
    }()

    // Add debug flag check at the start
    for i, arg := range os.Args {
        if arg == "--debug" {
            debugMode = true
            // Remove debug flag from args
            os.Args = append(os.Args[:i], os.Args[i+1:]...)
            break
        }
    }

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
        fmt.Println("   • Install built-in AI agents")
        fmt.Println("   • Prepare everything for your first chat")
        os.Exit(1)
    }

    // Now that we know chatty is initialized, load agents
    if err := agents.LoadAgents(); err != nil {
        fmt.Printf("Error loading agents: %v\n", err)
        os.Exit(1)
    }

    // Load configuration at startup
    config, err := agents.GetCurrentConfig()
    if err != nil {
        fmt.Printf("Error loading config: %v\n", err)
        // Continue with default agent
    } else {
        // Set current agent from config
        currentAgent = agents.GetAgentConfig(config.CurrentAgent)
    }

    if len(os.Args) < 2 {
        fmt.Println("Usage: chatty \"Your message here\" [--save <filename>]")
        fmt.Println("Special commands:")
        fmt.Println("  init                          Initialize Chatty environment")
        fmt.Println("  --clear [all|agent_name]      Clear chat history (all or specific agent)")
        fmt.Println("  --list                        List available agents")
        fmt.Println("  --select <agent_name>         Select an agent")
        fmt.Println("  --current                     Show current agent")
        fmt.Println("  --with <agent_name>           Start a direct chat with a single agent")
        fmt.Println("  --with <agent1>,<agent2>,...  Start a conversation between agents (interactive mode)")
        fmt.Println("      --topic \"message\"         Initial message for the conversation (required for --auto)")
        fmt.Println("      --topic-file <path>       Read initial message from a text file (required for --auto)")
        fmt.Println("      --turns N                 Number of conversation turns (default: infinite)")
        fmt.Println("      --auto                    Enable autonomous conversation mode")
        fmt.Println("      --save <filename>         Save conversation log to a file")
        fmt.Println("  --with-random <N>             Start a conversation with N random agents")
        fmt.Println("  --install <agent_name>        Install a new agent from the store")
        fmt.Println("  --uninstall <agent_name>      Uninstall a user-defined agent")
        fmt.Println("  --show <agent_name>           Show detailed information about an agent")
        fmt.Println("  --store                       List available agents in store")
        fmt.Println("\nOptions for simple chat mode:")
        fmt.Println("  --save <filename>             Save conversation log to a file")
        fmt.Println("\nNote: The --debug flag can be used with any command to show debug information.")
        return
    }

    // Handle special commands
    switch os.Args[1] {
    case "--build":
        handler := builder.NewHandler(debugMode)
        if err := handler.HandleBuildCommand(os.Args[2:]); err != nil {
            fmt.Printf("Error: %v\n", err)
            os.Exit(1)
        }
        return
    case "--with":
        if len(os.Args) < 3 {
            fmt.Println("Usage: chatty --with <agent_name> or <agent1>,<agent2>,... [options]")
            fmt.Println("\nOptions:")
            fmt.Println("  --topic \"message\"         Initial message for the conversation (required for --auto)")
            fmt.Println("  --topic-file <path>       Read initial message from a text file (required for --auto)")
            fmt.Println("  --turns N                 Number of conversation turns (default: infinite)")
            fmt.Println("  --auto                    Enable autonomous conversation mode (requires --topic)")
            fmt.Println("  --save <filename>         Save conversation log to a file")
            return
        }

        // Parse the comma-separated agent names
        agentList := strings.Split(os.Args[2], ",")
        var agentNames []string
        for _, agent := range agentList {
            trimmedAgent := strings.TrimSpace(agent)
            if trimmedAgent != "" {
                agentNames = append(agentNames, trimmedAgent)
            }
        }

        if len(agentNames) == 0 {
            fmt.Println("Error: No valid agent names provided")
            return
        }

        // Single agent mode
        if len(agentNames) == 1 {
            // Get agent name
            agentName := agentNames[0]
            
            // Check if agent exists
            if !agents.IsValidAgent(agentName) {
                fmt.Printf("Error: Invalid agent name '%s'\n", agentName)
                fmt.Println("\nAvailable agents:")
                fmt.Print(agents.ListAgents())
                return
            }

            // Parse other arguments
            var saveFile string
            var topicMessage string
            var foundTopicArg bool
            var autoMode bool

            // Find the --topic, --topic-file, and --save arguments
            for i := 3; i < len(os.Args); i++ {
                switch os.Args[i] {
                case "--topic":
                    if foundTopicArg {
                        fmt.Println("Error: cannot use both --topic and --topic-file")
                        return
                    }
                    foundTopicArg = true
                    if i+1 >= len(os.Args) {
                        fmt.Println("Error: --topic argument is missing")
                        fmt.Println("\nUsage: --topic \"your message here\"")
                        return
                    }
                    topicMessage = os.Args[i+1]
                    i++
                case "--topic-file":
                    if foundTopicArg {
                        fmt.Println("Error: cannot use both --topic and --topic-file")
                        return
                    }
                    foundTopicArg = true
                    if i+1 >= len(os.Args) {
                        fmt.Println("Error: --topic-file argument is missing")
                        fmt.Println("\nUsage: --topic-file <path>")
                        return
                    }
                    var err error
                    topicMessage, err = readStarterFile(os.Args[i+1])
                    if err != nil {
                        fmt.Printf("Error: %v\n", err)
                        return
                    }
                    i++
                case "--auto":
                    autoMode = true
                case "--save":
                    if i+1 >= len(os.Args) {
                        fmt.Println("Error: --save argument is missing")
                        fmt.Println("\nUsage: --save <filename>")
                        return
                    }
                    saveFile = os.Args[i+1]
                    i++
                }
            }

            // Auto mode is not valid for single agent chat
            if autoMode {
                fmt.Println("Error: --auto flag is only valid for multi-agent conversations")
                return
            }

            // Start the single-agent chat with the provided topic message or empty string
            if err := handleSingleAgentChat(agentName, topicMessage, saveFile); err != nil {
                fmt.Printf("Error: %v\n", err)
                return
            }
            return
        } else {
            // Multi-agent mode
            // Parse other arguments
            var config ConversationConfig
            var topicMessage string
            var turns int
            var foundTopicArg bool
            var autoMode bool
            var saveFile string

            // Set the agent names
            config.Agents = agentNames

            // Find the --topic, --topic-file, --auto, --turns, and --save arguments
            for i := 3; i < len(os.Args); i++ {
                switch os.Args[i] {
                case "--topic":
                    if foundTopicArg {
                        fmt.Println("Error: cannot use both --topic and --topic-file")
                        return
                    }
                    foundTopicArg = true
                    if i+1 >= len(os.Args) {
                        fmt.Println("Error: --topic argument is missing")
                        fmt.Println("\nUsage: --topic \"your message here\"")
                        return
                    }
                    topicMessage = os.Args[i+1]
                    i++
                case "--topic-file":
                    if foundTopicArg {
                        fmt.Println("Error: cannot use both --topic and --topic-file")
                        return
                    }
                    foundTopicArg = true
                    if i+1 >= len(os.Args) {
                        fmt.Println("Error: --topic-file argument is missing")
                        fmt.Println("\nUsage: --topic-file <path>")
                        return
                    }
                    var err error
                    topicMessage, err = readStarterFile(os.Args[i+1])
                    if err != nil {
                        fmt.Printf("Error: %v\n", err)
                        return
                    }
                    i++
                case "--auto":
                    autoMode = true
                case "--turns":
                    if i+1 < len(os.Args) {
                        turns, err = strconv.Atoi(os.Args[i+1])
                        if err != nil {
                            fmt.Printf("Error: invalid turns value: %v\n", err)
                            return
                        }
                        i++
                    }
                case "--save":
                    if i+1 >= len(os.Args) {
                        fmt.Println("Error: --save argument is missing")
                        fmt.Println("\nUsage: --save <filename>")
                        return
                    }
                    saveFile = os.Args[i+1]
                    i++
                }
            }

            // For autonomous mode, topic is required
            if autoMode && !foundTopicArg {
                fmt.Println("Error: --topic or --topic-file is required when using --auto")
                return
            }

            // Display participants list for auto mode
            if autoMode {
                fmt.Println("\n🎭 Multi-agent conversation started")
                fmt.Println("Participants:")
                for i, agentName := range agentNames {
                    agent := agents.GetAgentConfig(agentName)
                    fmt.Printf("%d. %s %s - %s\n", i+1, agent.Emoji, agent.Name, agent.Description)
                }
                fmt.Println()
            }

            // For interactive mode, if no topic is provided, prompt the user
            if !autoMode && !foundTopicArg {
                fmt.Println("\n🎭 Multi-agent conversation started")
                fmt.Println("Participants:")
                for i, agentName := range agentNames {
                    agent := agents.GetAgentConfig(agentName)
                    fmt.Printf("%d. %s %s - %s\n", i+1, agent.Emoji, agent.Name, agent.Description)
                }
                
                fmt.Println("\nEnter your message to start the conversation:")
                fmt.Println()
                fmt.Println("Press Enter with empty message to end the conversation")
                fmt.Println()
                fmt.Print(colorize("👤 User: ", "\033[1;36m"))
                
                reader := bufio.NewReader(os.Stdin)
                input, err := reader.ReadString('\n')
                if err != nil {
                    fmt.Printf("Error reading input: %v\n", err)
                    return
                }
                
                topicMessage = strings.TrimSpace(input)
                if topicMessage == "" {
                    fmt.Println("Conversation cancelled.")
                    return
                }
            }

            config.Starter = topicMessage
            config.Turns = turns
            config.AutoMode = autoMode
            config.SaveFile = saveFile

            if err := handleMultiAgentConversation(config); err != nil {
                fmt.Printf("Error: %v\n", err)
                return
            }
            return
        }

    case "--with-random":
        if len(os.Args) < 3 {
            fmt.Println("Usage: chatty --with-random <number_of_agents> [options]")
            fmt.Println("\nOptions:")
            fmt.Println("  --topic \"message\"         Initial message for the conversation (required for --auto)")
            fmt.Println("  --topic-file <path>       Read initial message from a text file (required for --auto)")
            fmt.Println("  --turns N                 Number of conversation turns (default: infinite)")
            fmt.Println("  --auto                    Enable autonomous conversation mode")
            fmt.Println("  --save <filename>         Save conversation log to a file")
            return
        }

        // Parse number of agents
        numAgents, err := strconv.Atoi(os.Args[2])
        if err != nil {
            fmt.Printf("Error: invalid number of agents: %v\n", err)
            return
        }

        // Get random agents
        selectedAgents, err := getRandomAgents(numAgents)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            return
        }

        // Parse other arguments
        var config ConversationConfig
        var topicMessage string
        var turns int
        var foundTopicArg bool
        var autoMode bool
        var saveFile string

        // Set the randomly selected agents
        config.Agents = selectedAgents

        // Find the --topic, --topic-file, --auto, --turns, and --save arguments
        for i := 3; i < len(os.Args); i++ {
            switch os.Args[i] {
            case "--topic":
                if foundTopicArg {
                    fmt.Println("Error: cannot use both --topic and --topic-file")
                    return
                }
                foundTopicArg = true
                if i+1 >= len(os.Args) {
                    fmt.Println("Error: --topic argument is missing")
                    fmt.Println("\nUsage: --topic \"your message here\"")
                    return
                }
                topicMessage = os.Args[i+1]
                i++
            case "--topic-file":
                if foundTopicArg {
                    fmt.Println("Error: cannot use both --topic and --topic-file")
                    return
                }
                foundTopicArg = true
                if i+1 >= len(os.Args) {
                    fmt.Println("Error: --topic-file argument is missing")
                    fmt.Println("\nUsage: --topic-file <path>")
                    return
                }
                var err error
                topicMessage, err = readStarterFile(os.Args[i+1])
                if err != nil {
                    fmt.Printf("Error: %v\n", err)
                    return
                }
                i++
            case "--auto":
                autoMode = true
            case "--turns":
                if i+1 < len(os.Args) {
                    turns, err = strconv.Atoi(os.Args[i+1])
                    if err != nil {
                        fmt.Printf("Error: invalid turns value: %v\n", err)
                        return
                    }
                    i++
                }
            case "--save":
                if i+1 >= len(os.Args) {
                    fmt.Println("Error: --save argument is missing")
                    fmt.Println("\nUsage: --save <filename>")
                    return
                }
                saveFile = os.Args[i+1]
                i++
            }
        }

        // For autonomous mode, topic is required
        if autoMode && !foundTopicArg {
            fmt.Println("Error: --topic or --topic-file is required when using --auto")
            return
        }

        // Display participants list for auto mode
        if autoMode {
            fmt.Println("\n🎭 Multi-agent conversation started")
            fmt.Println("Participants:")
            for i, agentName := range selectedAgents {
                agent := agents.GetAgentConfig(agentName)
                fmt.Printf("%d. %s %s - %s\n", i+1, agent.Emoji, agent.Name, agent.Description)
            }
            fmt.Println()
        }

        // For interactive mode, if no topic is provided, prompt the user
        if !autoMode && !foundTopicArg {
            fmt.Println("\n🎭 Multi-agent conversation started")
            fmt.Println("Participants:")
            for i, agentName := range selectedAgents {
                agent := agents.GetAgentConfig(agentName)
                fmt.Printf("%d. %s %s - %s\n", i+1, agent.Emoji, agent.Name, agent.Description)
            }
            
            fmt.Println("\nEnter your message to start the conversation:")
            fmt.Println()
            fmt.Println("Press Enter with empty message to end the conversation")
            fmt.Println()
            fmt.Print(colorize("👤 User: ", "\033[1;36m"))
            
            reader := bufio.NewReader(os.Stdin)
            input, err := reader.ReadString('\n')
            if err != nil {
                fmt.Printf("Error reading input: %v\n", err)
                return
            }
            
            topicMessage = strings.TrimSpace(input)
            if topicMessage == "" {
                fmt.Println("Conversation cancelled.")
                return
            }
        }

        config.Starter = topicMessage
        config.Turns = turns
        config.AutoMode = autoMode
        config.SaveFile = saveFile

        if err := handleMultiAgentConversation(config); err != nil {
            fmt.Printf("Error: %v\n", err)
            return
        }
        return

    case "--current":
        fmt.Printf("Current agent: %s - %s\n", currentAgent.Name, currentAgent.Description)
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
        fmt.Print(agents.ListAgents())
        return
    case "--store":
        handler := store.NewHandler(debugMode)
        if err := handler.ListAgents(); err != nil {
            fmt.Printf("Error: %v\n", err)
            os.Exit(1)
        }
        return
    case "--install":
        if len(os.Args) < 3 {
            fmt.Println("Error: Missing agent name. Usage: chatty --install <agent_name>")
            fmt.Println("\nUse 'chatty --store' to see available agents.")
            os.Exit(1)
        }

        handler := store.NewHandler(debugMode)
        if err := handler.InstallAgent(os.Args[2]); err != nil {
            fmt.Printf("Error: %v\n", err)
            os.Exit(1)
        }
        return
    case "--uninstall":
        if len(os.Args) < 3 {
            fmt.Println("Error: Please specify an agent name to uninstall")
            fmt.Println("\nUsage: chatty --uninstall \"Agent Name\"")
            fmt.Println("\nNote: Only user-defined agents can be uninstalled.")
            fmt.Println("To see available user-defined agents, use: chatty --list")
            os.Exit(1)
        }

        agentName := os.Args[2]
        
        // Define colors
        colorMagenta := "\u001b[1;35m"
        colorRed := "\u001b[1;31m"
        colorGreen := "\u001b[1;32m"
        colorPurple := "\u001b[1;95m"
        colorReset := "\u001b[0m"

        // Try to uninstall the agent
        if err := agents.UninstallAgent(agentName); err != nil {
            if strings.Contains(err.Error(), "cannot uninstall built-in agent") {
                fmt.Printf("\n%s🚫 Error:%s Cannot uninstall %s%s%s - it is a built-in agent\n", 
                    colorRed, colorReset, colorMagenta, agentName, colorReset)
                fmt.Println("\nOnly user-defined agents can be uninstalled.")
                fmt.Printf("To see available user-defined agents, use: %schatty --list%s\n",
                    colorPurple, colorReset)
            } else if strings.Contains(err.Error(), "not found") {
                fmt.Printf("\n%s🚫 Error:%s Agent %s%s%s not found\n", 
                    colorRed, colorReset, colorMagenta, agentName, colorReset)
                fmt.Printf("\nTo see available agents, use: %schatty --list%s\n",
                    colorPurple, colorReset)
            } else {
                fmt.Printf("\n%s🚫 Error:%s %v\n", colorRed, colorReset, err)
            }
            os.Exit(1)
        }

        // Success message
        fmt.Printf("\n%s✅ Success:%s Agent %s%s%s has been uninstalled\n", 
            colorGreen, colorReset, colorMagenta, agentName, colorReset)
        
        // Show available actions
        fmt.Println("\nQuick Actions:")
        fmt.Printf("  • %sView available agents:%s chatty --list\n", 
            colorPurple, colorReset)
        fmt.Printf("  • %sView store agents:%s chatty --store\n", 
            colorPurple, colorReset)
        os.Exit(0)
    case "--show":
        if len(os.Args) < 3 {
            fmt.Println("Error: Missing agent name. Usage: chatty --show <agent_name>")
            os.Exit(1)
        }

        // First try local agents
        if agents.IsValidAgent(os.Args[2]) {
            // Define color constants for better readability
            colorMagenta := "\033[1;35m"
            colorCyan := "\033[1;36m"
            colorGreen := "\033[32m"
            colorPurple := "\033[1;95m"
            colorBlue := "\033[1;34m"
            colorYellow := "\033[1;33m"
            colorReset := "\033[0m"
            
            // Force a refresh of the agents cache
            if err := agents.LoadAgents(); err != nil {
                fmt.Printf("Error refreshing agents: %v\n", err)
                os.Exit(1)
            }
            
            // Check if the agent exists
            if !agents.IsValidAgent(os.Args[2]) {
                // If not a valid agent, check if it's a sample agent
                homeDir, err := os.UserHomeDir()
                if err != nil {
                    fmt.Printf("Error getting home directory: %v\n", err)
                    os.Exit(1)
                }
                
                sampleAgentPath := filepath.Join(homeDir, ".chatty", "agents", os.Args[2]+".yaml.sample")
                
                // First check if it's installed under a different name
                if _, err := os.Stat(sampleAgentPath); err == nil {
                    // Read the sample file to get the actual agent name
                    data, err := os.ReadFile(sampleAgentPath)
                    if err == nil {
                        var sampleAgent agents.AgentConfig
                        if err := yaml.Unmarshal(data, &sampleAgent); err == nil {
                            // Check if this agent is already installed under its proper name
                            if agents.IsValidAgent(sampleAgent.Name) {
                                fmt.Printf("\n%s⚠️  Note:%s The agent '%s' is already installed as '%s'\n", 
                                    colorYellow, colorReset, os.Args[2], sampleAgent.Name)
                                fmt.Printf("Please use: chatty --show \"%s\"\n\n", sampleAgent.Name)
                                os.Exit(1)
                            }
                        }
                    }
                    
                    // If we get here, it's a genuine sample agent that's not installed
                    // Read the file again for display
                    data, err = os.ReadFile(sampleAgentPath)
                    if err != nil {
                        fmt.Printf("Error reading sample agent file: %v\n", err)
                        os.Exit(1)
                    }
                    
                    // Parse the YAML to get agent details
                    var agent agents.AgentConfig
                    if err := yaml.Unmarshal(data, &agent); err != nil {
                        fmt.Printf("Error parsing sample agent file: %v\n", err)
                        os.Exit(1)
                    }

                    fmt.Printf("\n%s🔍 Sample Agent Profile: %s%s%s\n", 
                        colorMagenta, colorYellow, agent.Name, colorReset)
                    
                    fmt.Printf("\n%s📋 Basic Information%s\n", colorCyan, colorReset)
                    fmt.Printf("  %s•%s %sIdentifier:%s %s\n", 
                        colorGreen, colorReset, colorPurple, colorReset, os.Args[2])
                    fmt.Printf("  %s•%s %sEmoji:%s %s\n", 
                        colorGreen, colorReset, colorPurple, colorReset, agent.Emoji)
                    fmt.Printf("  %s•%s %sDescription:%s %s\n", 
                        colorGreen, colorReset, colorPurple, colorReset, agent.Description)
                    fmt.Printf("  %s•%s %sStatus:%s Sample (Not Installed)\n", 
                        colorGreen, colorReset, colorPurple, colorReset)

                    fmt.Printf("\n%s🎭 System Message%s\n", colorCyan, colorReset)
                    fmt.Printf("%s%s%s\n", colorBlue, agent.SystemMessage, colorReset)

                    fmt.Printf("\n%s💡 Quick Actions%s\n", colorCyan, colorReset)
                    fmt.Printf("  %s1.%s %sInstall this agent:%s chatty --install %s\n", 
                        colorGreen, colorReset, colorPurple, colorReset, os.Args[2])
                    fmt.Printf("  %s2.%s %sAfter installation:%s chatty --select \"%s\"\n", 
                        colorGreen, colorReset, colorPurple, colorReset, agent.Name)
                    fmt.Printf("  %s3.%s %sStart chatting:%s chatty --with \"%s\"\n\n", 
                        colorGreen, colorReset, colorPurple, colorReset, agent.Name)
                } else {
                    fmt.Printf("Error: Agent '%s' not found\n", os.Args[2])
                    fmt.Println("\nTry these commands:")
                    fmt.Printf("  • %sView available agents:%s chatty --list\n", 
                        colorPurple, colorReset)
                    fmt.Printf("  • %sView sample agents:%s chatty --list-more\n", 
                        colorPurple, colorReset)
                    os.Exit(1)
                }
            } else {
                // Get the agent configuration using the agents package
                agent := agents.GetAgentConfig(os.Args[2])
                
                // Determine agent type and status
                agentType := "User Agent"
                if agent.Source == "built-in" {
                    agentType = "Built-in Agent"
                }
                
                // Get current agent for status
                currentAgentConfig, err := agents.GetCurrentConfig()
                isActive := false
                if err == nil && currentAgentConfig != nil && strings.EqualFold(currentAgentConfig.CurrentAgent, agent.Name) {
                    isActive = true
                }
                
                fmt.Printf("\n%s🔍 %s Profile: %s%s%s\n", 
                    colorMagenta, agentType, colorYellow, agent.Name, colorReset)
                
                fmt.Printf("\n%s📋 Basic Information%s\n", colorCyan, colorReset)
                fmt.Printf("  %s•%s %sIdentifier:%s %s\n", 
                    colorGreen, colorReset, colorPurple, colorReset, strings.ToLower(agent.Name))
                fmt.Printf("  %s•%s %sEmoji:%s %s\n", 
                    colorGreen, colorReset, colorPurple, colorReset, agent.Emoji)
                fmt.Printf("  %s•%s %sDescription:%s %s\n", 
                    colorGreen, colorReset, colorPurple, colorReset, agent.Description)
                fmt.Printf("  %s•%s %sType:%s %s\n", 
                    colorGreen, colorReset, colorPurple, colorReset, agentType)
                
                // Determine status text
                var statusText string
                if isActive {
                    statusText = "Active (Current Agent)"
                } else {
                    statusText = "Installed"
                }
                
                fmt.Printf("  %s•%s %sStatus:%s %s%s%s\n", 
                    colorGreen, colorReset, colorPurple, colorReset,
                    colorGreen, statusText, colorReset)

                fmt.Printf("\n%s🎭 System Message%s\n", colorCyan, colorReset)
                fmt.Printf("%s%s%s\n", colorBlue, agent.SystemMessage, colorReset)

                fmt.Printf("\n%s💡 Quick Actions%s\n", colorCyan, colorReset)
                
                // Track action number
                actionNum := 1

                // Show "Set as current agent" only if not active
                if !isActive {
                    fmt.Printf("  %s%d.%s %sSet as current agent:%s chatty --select \"%s\"\n", 
                        colorGreen, actionNum, colorReset, colorPurple, colorReset, agent.Name)
                    actionNum++
                }

                // Show chat actions with proper numbering
                fmt.Printf("  %s%d.%s %sStart a chat:%s chatty --with \"%s\"\n", 
                    colorGreen, actionNum, colorReset, colorPurple, colorReset, agent.Name)
                actionNum++

                fmt.Printf("  %s%d.%s %sStart group chat:%s chatty --with \"%s,<other_agent>\"\n", 
                    colorGreen, actionNum, colorReset, colorPurple, colorReset, agent.Name)
                actionNum++

                fmt.Printf("  %s%d.%s %sAuto conversation:%s chatty --with \"%s,<other_agent>\" --auto --topic \"<topic or message>\"\n", 
                    colorGreen, actionNum, colorReset, colorPurple, colorReset, agent.Name)
                actionNum++

                fmt.Printf("  %s%d.%s %sClear chat history:%s chatty --clear \"%s\"\n\n", 
                    colorGreen, actionNum, colorReset, colorPurple, colorReset, agent.Name)
            }
        } else {
            // Try store agents
            handler := store.NewHandler(debugMode)
            if err := handler.ShowAgent(os.Args[2]); err != nil {
                fmt.Printf("Error: Agent '%s' not found locally or in store\n", os.Args[2])
                fmt.Println("\nTry these commands:")
                fmt.Printf("  • View local agents:  chatty --list\n")
                fmt.Printf("  • View store agents:  chatty --store\n")
                os.Exit(1)
            }
        }
        return
    }

    // Parse arguments for --save
    var saveFile string
    var messageArgs []string
    for i := 1; i < len(os.Args); i++ {
        if os.Args[i] == "--save" {
            if i+1 >= len(os.Args) {
                fmt.Println("Error: --save argument is missing")
                fmt.Println("\nUsage: --save <filename>")
                return
            }
            saveFile = os.Args[i+1]
            i++ // Skip the filename in next iteration
        } else {
            messageArgs = append(messageArgs, os.Args[i])
        }
    }

    userInput := strings.Join(messageArgs, " ")
    if userInput == "" {
        fmt.Println("Error: message cannot be empty")
        return
    }
    
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

    // Convert existing history to use proper roles for the request
    var newHistory []Message
    
    // Always start with the system message
    for i, msg := range history {
        if i == 0 && msg.Role == "system" {
            // Keep the system message as is
            newHistory = append(newHistory, msg)
        } else if msg.Role == "user" {
            // Keep user messages as is
            newHistory = append(newHistory, msg)
        } else if msg.Role == "assistant" {
            // Keep assistant messages as is
            newHistory = append(newHistory, msg)
        }
    }
    
    // Prepare the request
    chatReq := ChatRequest{
        Model:    agents.GetCurrentModel(),
        Messages: newHistory,
        Stream:   true,
    }

    jsonData, err := json.Marshal(chatReq)
    if err != nil {
        fmt.Printf("Error marshaling request: %v\n", err)
        return
    }

    // Print top margin
    printChatMargin(chatTopMargin)

    // Start the animation before making the API request
    fmt.Printf("%s", colorize(getAgentLabel(), currentAgent.LabelColor))
    anim := startAnimation()

    // Make the API request with timeout (passing false for regular chat)
    resp, err := makeAPIRequest(jsonData)
    if err != nil {
        anim.stopAnimation() // Stop animation on error
        fmt.Printf("\nError: %v\n", err)
        if strings.Contains(err.Error(), "invalid model") {
            fmt.Printf("\nHint: Edit ~/.chatty/config.json to set a valid model name\n")
            fmt.Printf("Available models can be listed with: ollama list\n")
        }
        return
    }
    defer resp.Body.Close()

    // Process the streaming response (passing false for regular chat)
    fullResponseText, err := processStreamResponse(resp, anim)
    if err != nil {
        fmt.Printf("\nError: %v\n", err)
        return
    }

    // Ensure we're on a new line before printing margin
    fmt.Println()
    
    // Print bottom margin
    printChatMargin(chatBottomMargin)

    // Save the response to history
    history = append(history, Message{
        Role:    "assistant",
        Content: fullResponseText,
    })

    // Save updated history
    if err := saveHistory(history); err != nil {
        fmt.Printf("\nWarning: Failed to save chat history: %v\n", err)
    }

    // Save conversation log if requested
    if saveFile != "" {
        var conversationLog strings.Builder
        conversationLog.WriteString(fmt.Sprintf("👤 User: %s\n", userInput))
        conversationLog.WriteString(fmt.Sprintf("%s %s: %s\n", 
            currentAgent.Emoji, 
            currentAgent.Name, 
            fullResponseText))
        
        if err := saveConversationLog(saveFile, conversationLog.String()); err != nil {
            fmt.Printf("Warning: Failed to save conversation log: %v\n", err)
        } else {
            fmt.Printf("Conversation log saved to: %s\n", saveFile)
        }
    }
} 