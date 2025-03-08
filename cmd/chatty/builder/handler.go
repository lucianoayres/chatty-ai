package builder

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	colorGreen = "\u001b[32m"          // Green for success marks
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
		"💡 Tip: You can install agents with 'chatty --install \"Agent Name\"'",
		"💡 Tip: Browse community agents with 'chatty --store'",
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
				fmt.Printf("\r%s\r", strings.Repeat(" ", 2)) // Clear the line with just 2 spaces
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
						// Clear the entire line and reprint
						fmt.Printf("\r%s\r%s%s", 
							strings.Repeat(" ", len(prefix) + len(currentMessage) + 2),
							prefix,
							currentMessage)
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
						// Clear the entire line and reprint
						fmt.Printf("\r%s\r%s%s", 
							strings.Repeat(" ", len(prefix) + len(currentMessage) + 2),
							prefix,
							currentMessage)
						time.Sleep(30 * time.Millisecond) // Normal typing speed
					}
				}

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

// menuOption represents a selectable menu item
type menuOption struct {
	label string
	value string
}

// readKey reads a single keystroke from stdin
func readKey() ([]byte, error) {
	// Put terminal in raw mode
	exec.Command("stty", "-F", "/dev/tty", "raw").Run()
	defer exec.Command("stty", "-F", "/dev/tty", "cooked").Run()

	buffer := make([]byte, 3)
	n, err := os.Stdin.Read(buffer)
	if err != nil {
		return nil, err
	}
	return buffer[:n], nil
}

// showLightBarMenu displays a menu with a light bar selector
func showLightBarMenu(title string, options []menuOption, defaultIndex int) (int, error) {
	currentIndex := defaultIndex
	
	// Clear screen from cursor down before starting the menu
	fmt.Print("\033[J")
	
	// Save initial cursor position - this will be our menu start position
	fmt.Print("\033[s")
	
	// Draw the menu
	drawMenu := func() {
		// Restore cursor to saved position
		fmt.Print("\033[u")
		
		// Clear from cursor to end of screen
		fmt.Print("\033[J")
		
		// Show title with a separator line
		fmt.Printf("%s%s%s\n", colorSection, title, colorReset)
		fmt.Printf("%s%s%s\n\n", colorSection, strings.Repeat("─", len(title)), colorReset)

		// Display options
		for i, opt := range options {
			if i == currentIndex {
				// Highlighted option
				fmt.Printf(" %s▶ %s%s\n", colorHighlight, opt.label, colorReset)
			} else {
				// Normal option
				fmt.Printf("   %s\n", opt.label)
			}
		}

		// Show navigation help
		fmt.Printf("\n%s↑/↓: Navigate • Enter: Select • Esc: Cancel%s\n", colorPrompt, colorReset)
	}
	
	// Initial menu draw
	drawMenu()
	
	for {
		// Read keystroke
		key, err := readKey()
		if err != nil {
			return -1, err
		}

		// Handle key press
		if len(key) == 1 {
			switch key[0] {
			case 13: // Enter
				// Clear just the menu area
				fmt.Print("\033[u\033[J")
				return currentIndex, nil
			case 27: // Escape
				// Clear just the menu area
				fmt.Print("\033[u\033[J")
				return -1, nil
			}
		} else if len(key) == 3 {
			switch key[2] {
			case 65: // Up arrow
				if currentIndex > 0 {
					currentIndex--
					drawMenu()
				}
			case 66: // Down arrow
				if currentIndex < len(options)-1 {
					currentIndex++
					drawMenu()
				}
			}
		}
	}
}

// editAgentFields allows the user to edit agent fields through a light bar menu
func editAgentFields(agent *AgentSchema) bool {
	for {
		// Clear the screen before showing configuration
		fmt.Print("\033[J")
		
		// Display current configuration
		showAgentFields(agent)

		// Prepare menu options
		options := []menuOption{
			{label: "Continue with these settings", value: "continue"},
			{label: "Edit name", value: "name"},
			{label: "Edit emoji", value: "emoji"},
			{label: "Edit description", value: "description"},
			{label: "Edit system message", value: "system"},
			{label: "Edit tags", value: "tags"},
		}

		// Show menu and get selection
		selected, err := showLightBarMenu("🛠️  Edit Options", options, 0)
		if err != nil || selected == -1 {
			return false
		}

		// Handle selection
		switch options[selected].value {
		case "continue":
			return true
		case "name":
			fmt.Printf("\n%s✏️  Edit Name%s\n", colorSection, colorReset)
			fmt.Printf("%s════════════%s\n\n", colorSection, colorReset)
			fmt.Printf("Current name: %s%s%s\n", colorValue, agent.Name, colorReset)
			agent.Name = readUserInput(colorPrompt+"New name (Enter to keep current)"+colorReset, agent.Name)
		case "emoji":
			fmt.Printf("\n%s✏️  Edit Emoji%s\n", colorSection, colorReset)
			fmt.Printf("%s═════════════%s\n\n", colorSection, colorReset)
			fmt.Printf("Current emoji: %s%s%s\n", colorValue, agent.Emoji, colorReset)
			agent.Emoji = readUserInput(colorPrompt+"New emoji (Enter to keep current)"+colorReset, agent.Emoji)
		case "description":
			fmt.Printf("\n%s✏️  Edit Description%s\n", colorSection, colorReset)
			fmt.Printf("%s══════════════════%s\n\n", colorSection, colorReset)
			fmt.Printf("Current description: %s%s%s\n", colorValue, agent.Description, colorReset)
			agent.Description = readUserInput(colorPrompt+"New description (Enter to keep current)"+colorReset, agent.Description)
		case "system":
			fmt.Printf("\n%s✏️  Edit System Message%s\n", colorSection, colorReset)
			fmt.Printf("%s═══════════════════%s\n\n", colorSection, colorReset)
			fmt.Printf("%sCurrent system message:%s\n", colorPrompt, colorReset)
			fmt.Printf("%s%s%s\n", colorValue, agent.SystemMessage, colorReset)
			agent.SystemMessage = readMultilineInput(colorPrompt+"New system message (Enter to keep current)"+colorReset, agent.SystemMessage)
		case "tags":
			// Clear screen before showing tag edit interface
			fmt.Print("\033[J")
			
			fmt.Printf("\n%s✏️  Edit Tags%s\n", colorSection, colorReset)
			fmt.Printf("%s═══════════%s\n\n", colorSection, colorReset)
			
			// Show current tags
			if len(agent.Tags) > 0 {
				fmt.Printf("Current tags: ")
				for i, tag := range agent.Tags {
					if i > 0 {
						fmt.Print(", ")
					}
					fmt.Printf("%s%s%s", colorValue, tag, colorReset)
				}
				fmt.Println()
			} else {
				fmt.Printf("No tags currently assigned.\n")
			}
			
			// Get new tags
			tags, err := SelectTags(false) // debug mode off for cleaner UI
			if err == nil {
				agent.Tags = tags
			}
			
			// Clear screen after tag selection to prevent duplication
			fmt.Print("\033[J")
		}
	}
}

// readColorInput reads a color code using a light bar menu
func readColorInput(prompt string, defaultValue string, agent *AgentSchema, isLabelColor bool, selectedLabelColor string) string {
	// Print separator for visual clarity without clearing screen
	fmt.Printf("\n%s%s%s\n", colorSection, strings.Repeat("─", 50), colorReset)
	
	type colorOption struct {
		name string
		code string
	}

	colorOptions := []colorOption{
		{"Blue", "\u001b[38;5;39m"},
		{"Green", "\u001b[38;5;82m"},
		{"Purple", "\u001b[38;5;171m"},
		{"Orange", "\u001b[38;5;208m"},
		{"Red", "\u001b[38;5;196m"},
		{"Cyan", "\u001b[38;5;51m"},
		{"Yellow", "\u001b[38;5;226m"},
		{"Pink", "\u001b[38;5;213m"},
		{"White", "\u001b[38;5;255m"},
		{"Gray", "\u001b[38;5;245m"},
	}

	// Convert color options to menu options with preview
	menuOptions := make([]menuOption, len(colorOptions))
	for i, color := range colorOptions {
		if isLabelColor {
			// Preview for label color
			menuOptions[i] = menuOption{
				label: fmt.Sprintf("%s %s%s%s",
					agent.Emoji,
					color.code,
					agent.Name,
					colorReset),
				value: color.code,
			}
		} else {
			// Preview for text color
			menuOptions[i] = menuOption{
				label: fmt.Sprintf("%s %s%s%s: %s%s%s",
					agent.Emoji,
					selectedLabelColor,
					agent.Name,
					colorReset,
					color.code,
					"Hello! I am your AI assistant.",
					colorReset),
				value: color.code,
			}
		}
	}

	// Find default index
	defaultIndex := 0
	for i, opt := range colorOptions {
		if opt.code == defaultValue {
			defaultIndex = i
			break
		}
	}

	title := "🎨 Choose a color for the agent's " + (map[bool]string{true: "name", false: "messages"})[isLabelColor]
	selected, err := showLightBarMenu(title, menuOptions, defaultIndex)
	if err != nil || selected == -1 {
		return defaultValue
	}

	return menuOptions[selected].value
}

// showColorExamples displays example ANSI colors
func showColorExamples() {
	// This function is no longer needed as color examples are shown in readColorInput
}

// showAgentFields displays the current agent configuration
func showAgentFields(agent *AgentSchema) {
	// Add some spacing at the top
	fmt.Printf("\n\n%s📝 Current Configuration%s\n", colorSection, colorReset)
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

	// Show tags if any are present
	if len(agent.Tags) > 0 {
		fmt.Printf("\n%s5.%s Tags:        ", 
			colorHighlight, colorReset)
		
		for i, tag := range agent.Tags {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%s%s%s", colorValue, tag, colorReset)
		}
		fmt.Println()
	}
}

// HandleBuildCommand processes the build command
func (h *Handler) HandleBuildCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing agent description. Usage: chatty --build \"<agent description>\"")
	}

	// Get the agent description
	description := strings.Join(args, " ")

	// Clear screen when entering builder mode
	fmt.Print("\033[H\033[2J")
	
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
	var agent *AgentSchema
	var err error
	maxRetries := 3
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if h.debug {
			fmt.Printf("\n%s🔄 Debug Mode: Attempt %d of %d...%s\n", colorAccent, attempt, maxRetries, colorReset)
		}
		
		agent, err = h.builder.BuildAgent(description)
		if err == nil {
			break // Success, exit retry loop
		}
		
		if attempt < maxRetries {
			if h.debug {
				fmt.Printf("\n%s❌ Debug Mode: Attempt %d failed: %v%s\n", colorAccent, attempt, err, colorReset)
				fmt.Printf("%s⏳ Debug Mode: Retrying in 2 seconds...%s\n", colorAccent, colorReset)
			}
			time.Sleep(2 * time.Second) // Wait before retrying
		}
	}
	
	// Stop the animation if it was started
	if anim != nil {
		anim.stopAnimation()
	}

	if err != nil {
		if h.debug {
			fmt.Printf("\n%s❌ Debug Mode: All attempts failed:%s\n%v\n", colorAccent, colorReset, err)
		}
		return fmt.Errorf("failed to build agent after %d attempts: %v", maxRetries, err)
	}

	if h.debug {
		fmt.Printf("\n%s✅ Debug Mode: Agent configuration generated successfully%s\n", colorAccent, colorReset)
		yamlData, _ := yaml.Marshal(agent)
		fmt.Printf("\n%s%s%s\n", colorValue, string(yamlData), colorReset)
	}

	// Clear screen before showing editor
	fmt.Print("\033[J")

	// Show the generated configuration and allow editing
	editAgentFields(agent)

	// Clear screen before showing appearance options
	fmt.Print("\033[J")

	// Configure appearance
	fmt.Printf("\n%s3️⃣ Appearance%s\n", colorSection, colorReset)
	fmt.Printf("%sChoose colors for your agent's appearance in the chat.%s\n", colorPrompt, colorReset)
	
	// First select label color
	agent.LabelColor = readColorInput("Select label color", defaultLabelColor, agent, true, "")
	
	// Then select text color, passing the selected label color
	agent.TextColor = readColorInput("Select text color", defaultTextColor, agent, false, agent.LabelColor)

	// Check if the agent already has valid tags from the editor
	if len(agent.Tags) >= 1 && len(agent.Tags) <= 5 {
		// Agent already has valid tags, show them but don't require re-selection
		fmt.Printf("\n%s4️⃣ Tags%s\n", colorSection, colorReset)
		fmt.Printf("%sAgent already has the following tags:%s\n", colorPrompt, colorReset)
		fmt.Println()
		for _, tag := range agent.Tags {
			fmt.Printf("%s✓%s %s%s%s\n", 
				colorGreen, colorReset,
				colorHighlight, tag, colorReset)
		}
	} else {
		// No valid tags yet, prompt for selection
		selectedTags, err := SelectTags(h.debug)
		if err != nil {
			fmt.Printf("%s⚠️ Warning: Error selecting tags: %v%s\n", 
				colorAccent, err, colorReset)
			// Use default tag if there's an error
			agent.Tags = []string{"general"}
		} else {
			agent.Tags = selectedTags
		}

		// Ensure tags are selected - if empty, don't allow proceeding
		for len(agent.Tags) == 0 || len(agent.Tags) > 5 {
			fmt.Printf("\n%s❌ Error: You must select 1-5 tags before continuing.%s\n", 
				colorAccent, colorReset)
			selectedTags, err := SelectTags(h.debug)
			if err != nil {
				agent.Tags = []string{"general"} // Fallback to default
				break
			}
			agent.Tags = selectedTags
		}
	}

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

	// Show save options menu
	fmt.Printf("\n%s💾 Next Steps%s\n", colorSection, colorReset)
	fmt.Printf("%s═══════════%s\n", colorSection, colorReset)
	
	// Validate that tags are set before saving
	if len(agent.Tags) < 1 || len(agent.Tags) > 5 {
		fmt.Printf("\n%s⚠️ Warning: You must have 1-5 tags assigned before saving.%s\n", 
			colorAccent, colorReset)
		selectedTags, err := SelectTags(h.debug)
		if err == nil {
			agent.Tags = selectedTags
		} else {
			agent.Tags = []string{"general"} // Fallback
		}
	}
	
	fmt.Printf("%sAgent will be saved to: %s%s%s\n\n", 
		colorPrompt, 
		colorValue, 
		outputPath,
		colorReset)
		
	// Display the tags as part of save info
	fmt.Printf("%sTags: %s", colorPrompt, colorReset)
	for i, tag := range agent.Tags {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%s%s%s", colorHighlight, tag, colorReset)
	}
	fmt.Println("\n")

	options := []menuOption{
		{label: fmt.Sprintf("💬 Save & Chat with %s%s%s", colorValue, agent.Name, colorReset), value: "chat"},
		{label: "💾 Save & Exit", value: "save"},
		{label: "❌ Do not save & Exit", value: "cancel"},
	}

	selected, err := showLightBarMenu("Choose what to do next", options, 0)
	if err != nil || selected == -1 {
		fmt.Printf("\n%s❌ Agent creation cancelled.%s\n", colorAccent, colorReset)
		return nil
	}

	switch options[selected].value {
	case "cancel":
		fmt.Printf("\n%s❌ Agent creation cancelled.%s\n", colorAccent, colorReset)
		return nil
	case "chat", "save":
		// Save the agent configuration
		if err := h.builder.SaveAgent(agent, outputPath); err != nil {
			return fmt.Errorf("failed to save agent: %v", err)
		}

		if options[selected].value == "chat" {
			// Start chat with the new agent
			fmt.Printf("\n%s✨ Starting chat with %s%s%s...\n", 
				colorHighlight, 
				colorValue, 
				agent.Name, 
				colorReset)
			
			// Execute the chat command
			cmd := exec.Command("chatty", "--with", agent.Name)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		} else {
			// Show success message with next steps
			fmt.Printf("\n%s✅ Success! Your new agent has been created.%s\n\n", colorHighlight, colorReset)
			fmt.Printf("%s🚀 Quick Actions:%s\n", colorSection, colorReset)
			fmt.Printf("  %s1.%s Start chatting:     %schatty --with \"%s\"%s\n", 
				colorHighlight, colorReset, 
				colorValue, agent.Name, colorReset)
			fmt.Printf("  %s2.%s Set as default:     %schatty --select \"%s\"%s\n", 
				colorHighlight, colorReset, 
				colorValue, agent.Name, colorReset)
			fmt.Printf("  %s3.%s View agent details: %schatty --show \"%s\"%s\n", 
				colorHighlight, colorReset, 
				colorValue, agent.Name, colorReset)
		}
	}

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
		
	// Display tags if available
	if len(agent.Tags) > 0 {
		fmt.Print("\nTags: ")
		for i, tag := range agent.Tags {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%s%s%s", colorHighlight, tag, colorReset)
		}
		fmt.Println()
	}
} 