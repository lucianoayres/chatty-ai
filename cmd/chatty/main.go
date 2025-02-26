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

	"chatty/cmd/chatty/agents"
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

    // Conversation context templates
    normalConversationTemplate = `You are %s (%s) participating in a group conversation with other AI agents and a human user. This is an ongoing discussion where everyone contributes naturally. Remember that YOU are %s - always speak in first person and never refer to yourself in third person.

    Current participants (excluding yourself):
    %s

    Important guidelines:
    %s

    Conversation history:
    %s

    Previous message was from: %s
    Their message: "%s"

    Please respond naturally as part of this group conversation, keeping in mind that you are %s.`

    // Add this new constant for auto conversation guidelines
    defaultAutonomousGuidelines = `1. Always speak in first person (use "I", "my", "me") - never refer to yourself in third person
2. Address other agents by name when responding to them
3. Keep responses concise and conversational
4. Stay in character according to your role and expertise
5. Build upon previous messages and maintain conversation flow
6. DO NOT address or refer to the user - this is an autonomous discussion
7. Drive the conversation forward with questions and insights for other agents
8. Acknowledge what other agents have said before adding your perspective`

    // Update the autoConversationTemplate to use the guidelines
    autoConversationTemplate = `You are %s (%s) participating in an autonomous discussion with other AI agents. The human user has provided an initial topic but will not participate further - this is a self-sustaining conversation between AI agents only. Remember that YOU are %s - always speak in first person and never refer to yourself in third person.

    Current participants (excluding yourself):
    %s

    Important guidelines:
    %s

    Conversation history:
    %s

    Previous message was from: %s
    Their message: "%s"

    Please respond naturally as part of this autonomous discussion, keeping in mind that you are %s.`

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

    // Conversation mode timeouts (much higher due to multiple agents and longer responses)
    converseRequestTimeout = 300 * time.Second  // Initial connection timeout for converse mode
    converseReadTimeout = 300 * time.Second    // Timeout for reading each chunk in converse mode
    converseWriteTimeout = 300 * time.Second    // Timeout for writing requests in converse mode

    // Add this new constant for normal conversation guidelines
    defaultInteractiveGuidelines = `1. Always speak in first person (use "I", "my", "me") - never refer to yourself in third person
2. Address others by name when responding to them
3. Keep responses concise and conversational
4. Stay in character according to your role and expertise
5. Build upon previous messages and maintain conversation flow
6. Feel free to ask questions to other participants
7. Acknowledge what others have said before adding your perspective`
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
    return currentAgent.GetFullSystemMessage()
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
    baseDir := filepath.Join(homeDir, historyDir)
    if err := os.MkdirAll(baseDir, 0755); err != nil {
        return "", err
    }
    
    historyFile := agents.GetHistoryFileName(agentName)
    return filepath.Join(baseDir, historyFile), nil
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

// HTTP client without timeout (we'll use context for initial timeout)
var httpClient = &http.Client{}

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
            return fmt.Errorf("Ollama is not ready. Please ensure 'ollama serve' is running and the service is fully initialized")
        }
        return fmt.Errorf("Error checking Ollama: %v", err)
    }
    defer resp.Body.Close()
    return nil
}

// Update the makeAPIRequestWithRetry function
func makeAPIRequestWithRetry(jsonData []byte, history []Message, agent string, isConversation bool) (*http.Response, error) {
    // First, check if Ollama is ready
    if err := checkOllamaReady(); err != nil {
        return nil, err
    }

    var lastErr error
    retryDelay := initialRetryDelay

    // Create a channel for interrupt signals
    stopChan := make(chan os.Signal, 1)
    signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
    defer signal.Stop(stopChan)

    for attempt := 1; attempt <= maxRetries; attempt++ {
        // Show retry attempt if not first try
        if attempt > 1 {
            fmt.Printf("\nRetrying request for %s (attempt %d/%d)...\n", agent, attempt, maxRetries)
        }

        resp, err := makeAPIRequest(jsonData, history, isConversation)
        if err == nil {
            return resp, nil
        }

        // If we were interrupted, stop retrying
        if err.Error() == "interrupted" {
            return nil, err
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
            case <-stopChan:
                return nil, fmt.Errorf("interrupted")
            }
        }
    }

    return nil, fmt.Errorf("after %d attempts: %v", maxRetries, lastErr)
}

// Add debug flag at the top with other vars
var debugMode bool

// Update makeAPIRequest function
func makeAPIRequest(jsonData []byte, history []Message, isConversation bool) (*http.Response, error) {
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

// Update the handleMultiAgentConversation function
func handleMultiAgentConversation(config ConversationConfig) error {
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
    agentConfigs := make([]agents.AgentConfig, len(config.Agents))
    for i, name := range config.Agents {
        if !agents.IsValidAgent(name) {
            return fmt.Errorf("invalid agent name: %s", name)
        }
        agentConfigs[i] = agents.GetAgentConfig(name)
    }

    fmt.Println() // Add top margin
    fmt.Printf("Starting conversation between %d agents:\n", len(config.Agents))
    for i, agent := range agentConfigs {
        fmt.Printf("%d. %s %s\n", i+1, agent.Emoji, agent.Name)
    }
    fmt.Println() // Single line margin at bottom

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
            Content: agent.GetFullSystemMessage(),
        }}
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

    // Create a channel for graceful shutdown in auto mode
    stopChan := make(chan os.Signal, 1)
    if config.AutoMode {
        signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
        fmt.Println("\n🤖 Auto-conversation mode enabled. Press Ctrl+C to stop.")
        // Ensure we clean up the signal handler when we're done
        defer signal.Stop(stopChan)
    }

    for {
        // Update last active time
        state.lastActive = time.Now()

        // Check for stop signal before starting a new turn
        if config.AutoMode {
            select {
            case <-stopChan:
                fmt.Printf("\n\nConversation ended after %s\n",
                    formatElapsedTime(state.startTime, time.Now()))
                return nil
            default:
            }
        }

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
            
        // Print bottom separator and add exactly one line break
        fmt.Printf("%s%s%s\n", turnSeparatorColor, strings.Repeat("─", 60), colorReset)

        // Print the user's message at the start of each turn
        if firstMessage {
            // Add single blank line before user message
            fmt.Println()
            fmt.Println(colorize(formatUserMessage(config.Starter), "\033[1;36m"))
            firstMessage = false
            // Add blank line after first message in both auto and normal modes
            fmt.Println()
        } else if !config.AutoMode {
            // Only show user messages in non-auto mode after first turn
            fmt.Println(colorize(formatUserMessage(currentMessage), "\033[1;36m"))
        } else {
            // In auto mode after first turn, add blank line before first agent response
            fmt.Println()
        }

        for i, agent := range agentConfigs {
            // Check for stop signal before each agent's response
            if config.AutoMode {
                select {
                case <-stopChan:
                    fmt.Printf("\n\nConversation ended after %s\n",
                        formatElapsedTime(state.startTime, time.Now()))
                    return nil
                default:
                }
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
                    participants.WriteString(fmt.Sprintf("%d. %s (%s) - %s\n", 
                        j+1, 
                        other.Name, 
                        other.Emoji,
                        other.Description))
                }
            }
            if !config.AutoMode {
                participants.WriteString(fmt.Sprintf("%d. User (👤) - Human participant guiding the conversation\n", 
                    len(agentConfigs)))
            }

            // Get recent conversation history
            recentHistory := getRecentConversationHistory(conversationLog.String())

            // Create conversation context with identity reinforcement
            var templateToUse string
            var guidelines string
            if config.AutoMode {
                templateToUse = autoConversationTemplate
                // Get guidelines from config
                agentConfig, err := agents.GetCurrentConfig()
                if err != nil {
                    guidelines = defaultAutonomousGuidelines
                } else {
                    if agentConfig.AutonomousGuidelines != "" {
                        guidelines = agentConfig.AutonomousGuidelines
                    } else {
                        guidelines = defaultAutonomousGuidelines
                    }
                }
            } else {
                templateToUse = normalConversationTemplate
                // Get guidelines from config
                agentConfig, err := agents.GetCurrentConfig()
                if err != nil {
                    guidelines = defaultInteractiveGuidelines
                } else {
                    if agentConfig.InteractiveGuidelines != "" {
                        guidelines = agentConfig.InteractiveGuidelines
                    } else {
                        guidelines = defaultInteractiveGuidelines
                    }
                }
            }
            
            // Get the previous message source
            var prevMessageSource string
            if firstMessage {
                prevMessageSource = "User"
            } else {
                prevMessageSource = currentMessage
            }

            context := fmt.Sprintf(templateToUse,
                agent.Name,
                agent.Emoji,
                agent.Name,
                participants.String(),
                guidelines,
                recentHistory,
                prevMessageSource,
                currentMessage,
                agent.Name)

            // Reset this agent's history to keep context minimal
            histories[i] = []Message{{
                Role:    "system",
                Content: agent.GetFullSystemMessage(),
            }}

            // Add only the current context
            histories[i] = append(histories[i], Message{
                Role:    "user",
                Content: context + "\n\nYour response:",
            })

            // Prepare the request with full conversation history
            chatReq := ChatRequest{
                Model:    agents.GetCurrentModel(),
                Messages: histories[i],
                Stream:   true,
            }

            jsonData, err := json.Marshal(chatReq)
            if err != nil {
                anim.stopAnimation()
                return fmt.Errorf("error marshaling request for %s: %v", agent.Name, err)
            }

            // Make the API request with retry
            resp, err := makeAPIRequestWithRetry(jsonData, histories[i], agent.Name, true)
            if err != nil {
                anim.stopAnimation()
                return fmt.Errorf("error making request for %s: %v", agent.Name, err)
            }

            // Process the response
            fullResponseText, err := processStreamResponse(resp, anim, true)
            if err != nil {
                // Check if this was due to a stop signal
                if config.AutoMode {
                    select {
                    case <-stopChan:
                        fmt.Printf("\n\nConversation ended after %s\n",
                            formatElapsedTime(state.startTime, time.Now()))
                        return nil
                    default:
                    }
                }
                return fmt.Errorf("error processing response from %s: %v", agent.Name, err)
            }

            // Add the agent's response to their history
            histories[i] = append(histories[i], Message{
                Role:    "agent",
                Content: fullResponseText,
            })

            // Update conversation log
            conversationLog.WriteString(fmt.Sprintf("%s %s: %s\n", 
                agent.Emoji, 
                agent.Name, 
                fullResponseText))

            // Update for next iteration
            currentMessage = fullResponseText

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

    for {
        var streamResp ChatResponse
        err := decoder.Decode(&streamResp)
        
        if err == io.EOF {
            return fullResponse.String(), nil
        }
        if err != nil {
            return fullResponse.String(), fmt.Errorf("error reading response: %v", err)
        }

        // In debug mode, show the response chunk
        if debugMode {
            // Pretty print the response JSON
            prettyJSON, err := json.MarshalIndent(streamResp, "", "    ")
            if err != nil {
                fmt.Printf("Error formatting debug JSON: %v\n", err)
            } else {
                fmt.Printf("\n%sDebug: Stream Response:%s\n%s%s%s",
                    "\033[38;5;208m", // Orange for debug header
                    colorReset,
                    "\033[38;5;39m", // Light blue for JSON
                    string(prettyJSON),
                    colorReset)
            }
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

    // Copy sample agents
    if err := agents.CopySampleAgents(); err != nil {
        fmt.Printf("Warning: Failed to copy sample agents: %v\n", err)
    } else {
        fmt.Printf("%s✓%s Copied sample agent configurations\n", 
            "\033[32m", colorReset)
    }

    // Get default agent info
    defaultAgent := agents.DefaultAgent

    // Print success message with enhanced formatting
    fmt.Printf("\n%s🎉 Chatty has been successfully initialized!%s\n\n", 
        "\033[1;32m", colorReset)

    fmt.Printf("%s📌 Default Agent:%s\n", "\033[1;36m", colorReset)
    fmt.Printf("   %s %s%s%s - %s\n\n",
        defaultAgent.Emoji,
        defaultAgent.LabelColor,
        defaultAgent.Name,
        colorReset,
        defaultAgent.Description)

    fmt.Printf("%s🎯 Quick Start Guide:%s\n", "\033[1;35m", colorReset)
    fmt.Printf("   1. %sStart chatting:%s chatty \"Hello, how can you help me?\"\n",
        "\033[1;33m", colorReset)
    fmt.Printf("   2. %sList agents:%s chatty --list\n",
        "\033[1;33m", colorReset)
    fmt.Printf("   3. %sSwitch agents:%s chatty --select <name>\n",
        "\033[1;33m", colorReset)
    fmt.Printf("   4. %sStart a group chat:%s chatty --converse rocket tux --starter \"Let's talk\"\n",
        "\033[1;33m", colorReset)

    fmt.Printf("\n%s💡 Pro Tips:%s\n", "\033[1;36m", colorReset)
    fmt.Printf("   • Use %s--auto%s flag in group chats for autonomous agent discussions\n",
        "\033[1;33m", colorReset)
    fmt.Printf("   • Press %sCtrl+C%s to stop auto-conversations gracefully\n",
        "\033[1;33m", colorReset)
    fmt.Printf("   • Use %s--turns N%s to limit conversation length\n",
        "\033[1;33m", colorReset)

    fmt.Printf("\n%s🌟 Ready to start your AI journey!%s\n\n",
        "\033[1;32m", colorReset)
    
    return nil
}

// Add this new function before handleMultiAgentConversation
func makeAgentRequest(agent agents.AgentConfig, message string) (*http.Response, error) {
    // Initialize history with just the system message for this conversation
    history := []Message{{
        Role:    "system",
        Content: agent.GetFullSystemMessage(),
    }}

    // Add the current message
    history = append(history, Message{
        Role:    "user",
        Content: message,
    })

    // Prepare the request
    chatReq := ChatRequest{
        Model:    agents.GetCurrentModel(),
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
        fmt.Println("   • Install sample AI agents")
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
        fmt.Println("Usage: chatty \"Your message here\"")
        fmt.Println("Special commands:")
        fmt.Println("  init                          Initialize Chatty environment")
        fmt.Println("  --clear [all|agent_name]  Clear chat history (all or specific agent)")
        fmt.Println("  --list                       List available agents")
        fmt.Println("  --select <agent_name>    Select an agent")
        fmt.Println("  --current                    Show current agent")
        fmt.Println("  --converse <agents...>   Start a conversation between agents")
        fmt.Println("      --starter \"message\"      Initial message to start the conversation")
        fmt.Println("      --turns N                Number of conversation turns (default: infinite)")
        fmt.Println("  --debug                      Show debug information including request JSON")
        return
    }

    // Handle special commands
    switch os.Args[1] {
    case "--converse":
        if len(os.Args) < 4 {
            fmt.Println("Usage: chatty --converse agent1,agent2[,agent3...] --starter \"message\" [--turns N] [--auto]")
            fmt.Println("\nOptions:")
            fmt.Println("  --starter \"message\"  Initial message to start the conversation (required)")
            fmt.Println("  --turns N           Number of conversation turns (default: infinite)")
            fmt.Println("  --auto              Enable autonomous conversation mode (agents talk among themselves)")
            return
        }

        // Parse arguments
        var config ConversationConfig
        var starter string
        var turns int
        var foundStarterArg bool
        var autoMode bool

        // Parse the comma-separated agent names
        agentList := strings.Split(os.Args[2], ",")
        config.Agents = make([]string, 0, len(agentList))
        for _, agent := range agentList {
            trimmedAgent := strings.TrimSpace(agent)
            if trimmedAgent != "" {
                config.Agents = append(config.Agents, trimmedAgent)
            }
        }

        // Find the --starter argument and check for --auto
        for i := 3; i < len(os.Args); i++ {
            switch os.Args[i] {
            case "--starter":
                foundStarterArg = true
                // Check if next argument exists
                if i+1 >= len(os.Args) {
                    fmt.Println("Error: --starter argument is missing")
                    fmt.Println("\nUsage: --starter \"your message here\"")
                    return
                }
                // Take the next argument as is
                starter = os.Args[i+1]
                i++ // Skip the next argument since we've used it
            case "--auto":
                autoMode = true
            case "--turns":
                if i+1 < len(os.Args) {
                    turns, err = strconv.Atoi(os.Args[i+1])
                    if err != nil {
                        fmt.Printf("Error: invalid turns value: %v\n", err)
                        return
                    }
                    i++ // Skip the next argument since we've used it
                }
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

        if len(config.Agents) < 2 {
            fmt.Println("Error: at least two agents must be specified, separated by commas")
            fmt.Println("\nExample: chatty --converse plato,aristotle,socrates --starter \"Let's discuss philosophy\"")
            return
        }

        config.Starter = starter
        config.Turns = turns
        config.AutoMode = autoMode

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
    case "--select":
        if len(os.Args) < 3 {
            fmt.Println("Please specify an agent name")
            return
        }
        
        // Validate agent name before making any changes
        if !agents.IsValidAgent(os.Args[2]) {
            fmt.Printf("Error: Invalid agent name '%s'\n", os.Args[2])
            fmt.Println("\nAvailable agents:")
            fmt.Print(agents.ListAgents())
            return
        }
        
        currentAgent = agents.GetAgentConfig(os.Args[2])
        if err := agents.UpdateCurrentAgent(os.Args[2]); err != nil {
            fmt.Printf("Error saving agent selection: %v\n", err)
            return
        }
        fmt.Printf("Switched to %s [%s%s%s] %s\n", 
            currentAgent.Emoji,
            currentAgent.LabelColor,
            currentAgent.Name,
            "\u001b[0m", // Reset color
            currentAgent.Description)
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
        Model:    agents.GetCurrentModel(),
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

    // Show agent label immediately
    fmt.Printf("%s", colorize(getAgentLabel(), currentAgent.LabelColor))

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

    // Add agent's response to history
    history = append(history, Message{
        Role:    "agent",
        Content: fullResponseText,
    })

    // Save updated history
    if err := saveHistory(history); err != nil {
        fmt.Printf("Error saving history: %v\n", err)
    }
} 