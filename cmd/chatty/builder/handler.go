package builder

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	ollamaBaseURL = "http://localhost:11434/api/generate"  // Full URL for Ollama generate API
	ollamaModel   = "llama3.2"  // Model to use for generating agent configurations

	// Default ANSI color codes
	defaultLabelColor = "\u001b[38;5;39m"  // Light blue
	defaultTextColor  = "\u001b[38;5;251m" // Light gray

	// UI Colors
	colorTitle = "\u001b[38;5;39m"    // Light blue for titles
	colorSection = "\u001b[38;5;171m" // Purple for section headers
	colorHighlight = "\u001b[38;5;82m" // Green for highlights
	colorPrompt = "\u001b[38;5;251m"   // Light gray for prompts
	colorValue = "\u001b[38;5;255m"    // White for values
	colorAccent = "\u001b[38;5;208m"   // Orange for accents
	colorReset = "\u001b[0m"           // Reset color
)

// Animation represents a loading animation
type Animation struct {
	stopChan chan bool
}

// startAnimation starts a loading animation and returns a handle to stop it
func startAnimation() *Animation {
	anim := &Animation{
		stopChan: make(chan bool),
	}

	// Tips about using the program
	messages := []string{
		"💡 Tip: Use 'chatty --with \"Agent Name\"' to chat with a specific agent",
		"💡 Tip: Press Enter without typing to keep the current value when editing",
		"💡 Tip: You can set a default agent with 'chatty --select \"Agent Name\"'",
		"💡 Tip: View all available agents with 'chatty --list'",
		"💡 Tip: Use 'chatty --show \"Agent Name\"' to see agent details",
		"💡 Tip: Type ':wq' on a new line to finish editing multi-line text",
		"💡 Tip: You can install sample agents with 'chatty --install \"Agent Name\"'",
		"💡 Tip: Browse sample agents with 'chatty --list-more'",
		"💡 Tip: Agents are saved in ~/.chatty/agents as YAML files",
		"💡 Tip: Each agent can have its own unique color scheme",
		"💡 Tip: Use descriptive names for your agents to remember their roles",
		"💡 Tip: The system message defines your agent's personality",
		"💡 Tip: Choose emojis that represent your agent's role",
		"💡 Tip: You can have multiple agents for different tasks",
		"💡 Tip: Make your agent's description clear and specific",
		"💡 Tip: Test different colors to find the perfect combination",
		"💡 Tip: You can edit any field before saving your agent",
		"💡 Tip: Use 'chatty --uninstall \"Agent Name\"' to remove an agent",
		"💡 Tip: View conversation history in ~/.chatty/history",
		"💡 Tip: Debug mode can be enabled with --debug flag",
	}

	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		frameIndex := 0
		currentMessage := ""
		prefix := colorAccent + frames[0] + " " + colorReset
		
		// Create a shuffled list of indices for random selection
		messageIndices := make([]int, len(messages))
		for i := range messageIndices {
			messageIndices[i] = i
		}
		
		// Fisher-Yates shuffle
		for i := len(messageIndices) - 1; i > 0; i-- {
			j := time.Now().UnixNano() % int64(i+1)
			messageIndices[i], messageIndices[j] = messageIndices[j], messageIndices[i]
		}
		
		currentIndex := 0
		lastMessageChange := time.Now()
		isErasing := false
		
		for {
			select {
			case <-anim.stopChan:
				fmt.Printf("\r%s\r", strings.Repeat(" ", 80)) // Clear the line
				return
			default:
				// Update the prefix with current frame
				prefix = colorAccent + frames[frameIndex] + " " + colorReset
				
				// Change message every 15 seconds
				if time.Since(lastMessageChange) >= 15*time.Second {
					isErasing = true
					lastMessageChange = time.Now()
				}

				if isErasing {
					// Erase the current message character by character
					if len(currentMessage) > 0 {
						currentMessage = currentMessage[:len(currentMessage)-1]
						time.Sleep(15 * time.Millisecond) // Faster backspace speed
					} else {
						// When fully erased, prepare for the next message
						isErasing = false
						currentIndex = (currentIndex + 1) % len(messages)
						// If we've used all messages, reshuffle
						if currentIndex == 0 {
							for i := len(messageIndices) - 1; i > 0; i-- {
								j := time.Now().UnixNano() % int64(i+1)
								messageIndices[i], messageIndices[j] = messageIndices[j], messageIndices[i]
							}
						}
					}
				} else {
					// Type the new message character by character
					targetMessage := messages[messageIndices[currentIndex]]
					if len(currentMessage) < len(targetMessage) {
						currentMessage = targetMessage[:len(currentMessage) + 1]
						time.Sleep(30 * time.Millisecond) // Normal typing speed
					}
				}

				// Clear the line and print the current frame and message
				fmt.Printf("\r%s\r%s%s", 
					strings.Repeat(" ", 80),
					prefix,
					currentMessage)
				
				frameIndex = (frameIndex + 1) % len(frames)
				if !isErasing && len(currentMessage) == len(messages[messageIndices[currentIndex]]) {
					time.Sleep(80 * time.Millisecond) // Normal spinner speed
				}
			}
		}
	}()

	return anim
}

// stopAnimation stops the loading animation
func (a *Animation) stopAnimation() {
	a.stopChan <- true
}

// Handler manages the agent building process
type Handler struct {
	builder *Builder
	debug   bool
}

// NewHandler creates a new builder handler
func NewHandler(debug bool) *Handler {
	// Create the LLM client with the model specifically for building agents
	llm := NewOllamaClient(ollamaBaseURL, ollamaModel)
	llm.SetDebug(debug)  // Set debug mode on the LLM client

	// Create the builder with default configuration
	builder := NewBuilder(DefaultBuilderConfig(), llm)

	return &Handler{
		builder: builder,
		debug:   debug,
	}
}

// readUserInput reads a line of user input with a prompt
func readUserInput(prompt string, defaultValue string) string {
	fmt.Printf("%s [%s]: ", prompt, defaultValue)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

// readColorInput reads a color code from user input
func readColorInput(prompt string, defaultValue string, agent *AgentSchema, isLabelColor bool, selectedLabelColor string) string {
	colors := map[string]string{
		"Blue":    "\u001b[38;5;39m",
		"Green":   "\u001b[38;5;82m",
		"Purple":  "\u001b[38;5;171m",
		"Orange":  "\u001b[38;5;208m",
		"Red":     "\u001b[38;5;196m",
		"Cyan":    "\u001b[38;5;51m",
		"Yellow":  "\u001b[38;5;226m",
		"Pink":    "\u001b[38;5;213m",
		"White":   "\u001b[38;5;255m",
		"Gray":    "\u001b[38;5;245m",
	}

	// Convert map to sorted slice for consistent ordering
	type colorOption struct {
		name string
		code string
	}
	colorOptions := []colorOption{
		{"Blue", colors["Blue"]},
		{"Green", colors["Green"]},
		{"Purple", colors["Purple"]},
		{"Orange", colors["Orange"]},
		{"Red", colors["Red"]},
		{"Cyan", colors["Cyan"]},
		{"Yellow", colors["Yellow"]},
		{"Pink", colors["Pink"]},
		{"White", colors["White"]},
		{"Gray", colors["Gray"]},
	}

	for {
		fmt.Print("\033[H\033[2J") // Clear screen
		
		if isLabelColor {
			fmt.Printf("\n🎨 Choose a color for the agent's name:\n\n")
			// Show preview of current name with each color
			for i, color := range colorOptions {
				fmt.Printf("%d. %s %s%s\u001b[0m\n", 
					i+1,
					agent.Emoji,
					color.code,
					agent.Name)
			}
		} else {
			fmt.Printf("\n🎨 Choose a color for the agent's messages:\n\n")
			// Show preview of current message with each color
			for i, color := range colorOptions {
				fmt.Printf("%d. %s %s%s\u001b[0m: %s%s\u001b[0m\n",
					i+1,
					agent.Emoji,
					selectedLabelColor,
					agent.Name,
					color.code,
					"Hello! I am your AI assistant.")
			}
		}

		// Show current selection if any
		if defaultValue != "" {
			fmt.Printf("\nCurrent selection: %s%s\u001b[0m\n", defaultValue, "■■■■")
		}

		// Prompt for selection
		fmt.Printf("\nSelect a color (1-%d, Enter for default): ", len(colorOptions))
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		// Handle default value
		if input == "" {
			return defaultValue
		}

		// Parse number
		if num, err := strconv.Atoi(input); err == nil && num > 0 && num <= len(colorOptions) {
			return colorOptions[num-1].code
		}

		fmt.Printf("Please enter a number between 1 and %d\n", len(colorOptions))
		time.Sleep(2 * time.Second) // Give time to read the error message
	}
}

// showColorExamples displays example ANSI colors
func showColorExamples() {
	// This function is no longer needed as color examples are shown in readColorInput
}

// showAgentFields displays the current agent configuration
func showAgentFields(agent *AgentSchema) {
	fmt.Printf("\n%s📝 Current Configuration%s\n", colorSection, colorReset)
	fmt.Printf("%s═══════════════════%s\n\n", colorSection, colorReset)
	
	fmt.Printf("%s1.%s Name:        %s%s%s\n", 
		colorHighlight, colorReset, 
		colorValue, agent.Name, colorReset)
	
	fmt.Printf("%s2.%s Emoji:       %s%s%s\n", 
		colorHighlight, colorReset, 
		colorValue, agent.Emoji, colorReset)
	
	fmt.Printf("%s3.%s Description: %s%s%s\n", 
		colorHighlight, colorReset, 
		colorValue, agent.Description, colorReset)
	
	fmt.Printf("\n%s4.%s System Message:\n%s%s%s\n", 
		colorHighlight, colorReset, 
		colorValue, agent.SystemMessage, colorReset)
}

// editAgentFields allows the user to edit agent fields through a menu
func editAgentFields(agent *AgentSchema) bool {
	for {
		fmt.Print("\033[H\033[2J") // Clear screen
		showAgentFields(agent)
		
		fmt.Printf("\n%s🛠️  Edit Options%s\n", colorSection, colorReset)
		fmt.Printf("%s═════════════%s\n", colorSection, colorReset)
		fmt.Printf("\n%sChoose a field to edit (1-4) or press Enter to continue with appearance:%s\n", colorPrompt, colorReset)
		fmt.Printf("%s1)%s Name\n", colorHighlight, colorReset)
		fmt.Printf("%s2)%s Emoji\n", colorHighlight, colorReset)
		fmt.Printf("%s3)%s Description\n", colorHighlight, colorReset)
		fmt.Printf("%s4)%s System Message\n", colorHighlight, colorReset)
		
		fmt.Printf("\nChoice: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		
		if input == "" {
			return true // Continue to appearance configuration
		}
		
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > 4 {
			fmt.Printf("%sInvalid choice. Please enter a number between 1 and 4.%s\n", colorAccent, colorReset)
			time.Sleep(2 * time.Second)
			continue
		}
		
		// Show current value and prompt for edit
		fmt.Print("\033[H\033[2J") // Clear screen
		fmt.Printf("%s📝 Edit Field%s\n", colorSection, colorReset)
		fmt.Printf("%s════════════%s\n\n", colorSection, colorReset)
		
		switch choice {
		case 1:
			fmt.Printf("Current name: %s%s%s\n", colorValue, agent.Name, colorReset)
			agent.Name = readUserInput(colorPrompt+"New name (Enter to keep current)"+colorReset, agent.Name)
		case 2:
			fmt.Printf("Current emoji: %s%s%s\n", colorValue, agent.Emoji, colorReset)
			agent.Emoji = readUserInput(colorPrompt+"New emoji (Enter to keep current)"+colorReset, agent.Emoji)
		case 3:
			fmt.Printf("Current description: %s%s%s\n", colorValue, agent.Description, colorReset)
			agent.Description = readUserInput(colorPrompt+"New description (Enter to keep current)"+colorReset, agent.Description)
		case 4:
			fmt.Printf("%sCurrent system message:%s\n", colorPrompt, colorReset)
			fmt.Printf("%s%s%s\n", colorValue, agent.SystemMessage, colorReset)
			agent.SystemMessage = readMultilineInput(colorPrompt+"New system message (Enter to keep current)"+colorReset, agent.SystemMessage)
		}
	}
}

// HandleBuildCommand processes the build command
func (h *Handler) HandleBuildCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing agent description. Usage: chatty --build \"<agent description>\"")
	}

	// Get the agent description
	description := strings.Join(args, " ")

	// Clear screen and show welcome message
	fmt.Print("\033[H\033[2J") // Clear screen
	fmt.Printf("%s🎨 Agent Builder%s\n", colorTitle, colorReset)
	fmt.Printf("%s══════════════%s\n", colorTitle, colorReset)
	fmt.Printf("\n%s📝 Creating a new agent based on your description:%s\n", colorSection, colorReset)
	fmt.Printf("   %s%s%s\n\n", colorValue, description, colorReset)

	if h.debug {
		fmt.Printf("%s🔍 Debug Mode: Generating agent configuration...%s\n", colorAccent, colorReset)
	}

	// Start the loading animation
	var anim *Animation
	if !h.debug {
		anim = startAnimation()
	}

	// Generate the initial agent configuration
	agent, err := h.builder.BuildAgent(description)
	
	// Stop the animation if it was started
	if anim != nil {
		anim.stopAnimation()
	}

	if err != nil {
		if h.debug {
			fmt.Printf("\n%s❌ Debug Mode: Generation failed:%s\n%v\n", colorAccent, colorReset, err)
		}
		return fmt.Errorf("failed to build agent: %v", err)
	}

	if h.debug {
		fmt.Printf("\n%s✅ Debug Mode: Initial configuration generated%s\n", colorAccent, colorReset)
		yamlData, _ := yaml.Marshal(agent)
		fmt.Printf("\n%s%s%s\n", colorValue, string(yamlData), colorReset)
	}

	// Show the generated configuration and allow editing
	editAgentFields(agent)

	// Configure appearance
	fmt.Printf("\n%s3️⃣ Appearance%s\n", colorSection, colorReset)
	fmt.Printf("%sChoose colors for your agent's appearance in the chat.%s\n", colorPrompt, colorReset)
	
	// First select label color
	agent.LabelColor = readColorInput("Select label color", defaultLabelColor, agent, true, "")
	
	// Then select text color, passing the selected label color
	agent.TextColor = readColorInput("Select text color", defaultTextColor, agent, false, agent.LabelColor)

	// Preview the agent's appearance
	fmt.Printf("\n%s👀 Preview%s\n", colorSection, colorReset)
	fmt.Printf("%s═════════%s\n", colorSection, colorReset)
	previewAgent(*agent)

	// Get the user's agents directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}
	agentsDir := filepath.Join(homeDir, ".chatty", "agents")

	// Create the output file path
	filename := strings.ToLower(strings.ReplaceAll(agent.Name, " ", "_")) + ".yaml"
	outputPath := filepath.Join(agentsDir, filename)

	// Confirm before saving
	fmt.Printf("\n%s💾 Save Agent%s\n", colorSection, colorReset)
	fmt.Printf("%s═══════════%s\n", colorSection, colorReset)
	fmt.Printf("%sThe agent will be saved to: %s%s%s\n", 
		colorPrompt, 
		colorValue, 
		outputPath,
		colorReset)
	if !confirmAction(colorPrompt + "Would you like to save this agent?" + colorReset) {
		fmt.Printf("\n%s❌ Agent creation cancelled.%s\n", colorAccent, colorReset)
		return nil
	}

	// Save the agent configuration
	if err := h.builder.SaveAgent(agent, outputPath); err != nil {
		return fmt.Errorf("failed to save agent: %v", err)
	}

	// Show success message with next steps
	fmt.Printf("\n%s✅ Success! Your new agent has been created.%s\n\n", colorHighlight, colorReset)
	fmt.Printf("%s🚀 Quick Actions:%s\n", colorSection, colorReset)
	fmt.Printf("  %s1.%s Start chatting:     %schatty --with \"%s\"%s\n", 
		colorHighlight, colorReset, colorValue, agent.Name, colorReset)
	fmt.Printf("  %s2.%s Set as default:     %schatty --select \"%s\"%s\n", 
		colorHighlight, colorReset, colorValue, agent.Name, colorReset)
	fmt.Printf("  %s3.%s View agent details: %schatty --show \"%s\"%s\n", 
		colorHighlight, colorReset, colorValue, agent.Name, colorReset)

	return nil
}

// readMultilineInput reads multiline input with a default value
func readMultilineInput(prompt, defaultValue string) string {
	fmt.Printf("\n%s (Press Ctrl+D or type ':wq' on a new line to finish):\n", prompt)
	fmt.Printf("Current value:\n%s\n\n", defaultValue)
	fmt.Println("Enter new value (or press Enter to keep current):")

	// Check if user just pressed Enter
	reader := bufio.NewReader(os.Stdin)
	firstLine, _ := reader.ReadString('\n')
	if strings.TrimSpace(firstLine) == "" {
		return defaultValue
	}

	// Read multiple lines
	var lines []string
	lines = append(lines, firstLine)
	for {
		line, _ := reader.ReadString('\n')
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == ":wq" || trimmedLine == "" {
			break
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "")
}

// previewAgent shows how the agent will look in chat
func previewAgent(agent AgentSchema) {
	fmt.Printf("\nChat preview:\n")
	fmt.Printf("════════════\n")
	// Show the agent's name with label color
	fmt.Printf("%s %s%s%s: ", 
		agent.Emoji, 
		agent.LabelColor, 
		agent.Name,
		colorReset)
	
	// Show a sample message with text color
	fmt.Printf("%sHello! I am %s, ready to assist you.%s\n", 
		agent.TextColor, 
		agent.Name,
		colorReset)
}

// confirmAction asks for user confirmation
func confirmAction(prompt string) bool {
	fmt.Printf("\n%s (y/N): ", prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.ToLower(strings.TrimSpace(input)) == "y"
} 