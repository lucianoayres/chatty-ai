name: "Max"

# Chatty

Chatty is a command-line interface for chatting with Ollama's LLMs (Large Language Models) with streaming responses, conversation history, and customizable AI personalities.

## Features

- Real-time streaming responses
- Persistent conversation history per assistant
- Multiple specialized AI personalities
- Customizable appearance and behavior
- Clear and readable output format
- Colored output with 256-color support
- Consistent text margins
- Sample assistant configurations for easy customization

## Built-in Assistants

These are the assistants that come pre-configured with Chatty. You can use them as-is, customize them, or create entirely new ones (see [Customizing Assistants](#customizing-assistants) section below).

- **Rocket** 🚀 - Your friendly coding companion for development guidance and best practices
- **Tux** 🐧 - Your Linux terminal expert for command-line operations and shell scripting
- **Focus** 🎯 - Your efficiency expert for productivity and organization

Want more? You can easily create your own assistants by:

1. Using the provided sample configurations in `~/.chatty/assistants/*.sample`
2. Copying and customizing existing assistants
3. Creating new ones from scratch with your own personality and specialization

Each assistant can have its own:

- Unique personality and expertise
- Custom emoji and color scheme
- Specialized knowledge domain
- Conversation style and approach

## Project Structure

```
chatty/
├── cmd/
│   └── chatty/
│       ├── main.go            # Main application code
│       └── assistants/
│           ├── assistants.go  # Assistant management
│           ├── builtin/       # Built-in assistant configurations
│           │   ├── rocket.yaml   # Coding assistant
│           │   ├── tux.yaml      # Linux terminal expert
│           │   └── focus.yaml    # Productivity expert
│           └── samples/       # Sample assistant configurations
│               └── focus.yaml    # Example assistant template
└── ~/.chatty/                # User data directory (created automatically)
    ├── config.json          # Current assistant selection
    ├── chat_history_*.json  # Conversation histories
    └── assistants/         # User-defined assistants
        └── *.yaml.sample   # Sample assistant configurations
```

## Prerequisites

- Go 1.16 or later
- Ollama installed and running locally
- The `llama3.2` model installed in Ollama

## Installation

1. Clone this repository:

```bash
git clone <your-repo-url>
cd chatty
```

2. Make sure Ollama is running:

```bash
ollama serve
```

3. Install the llama3.2 model if you haven't already:

```bash
ollama pull llama3.2
```

4. Build the application:

```bash
# Using go build
go build -o bin/chatty cmd/chatty/main.go

# Or using the provided task file
task build
```

## Usage

### Basic Chat

You can send messages in two ways:

1. Without quotes (for simple messages):

```bash
./bin/chatty How are you doing today?
```

2. With quotes (required for messages containing special characters):

```bash
./bin/chatty "What's the meaning of life?"
```

### Special Commands

```bash
# Clear conversation history
./bin/chatty --clear              # Clear all histories
./bin/chatty --clear rocket      # Clear specific assistant's history
./bin/chatty --clear "Data Scientist"  # Clear history for assistant with multi-word name

# List available assistants
./bin/chatty --list

# Switch to a different assistant
./bin/chatty --select rocket           # Single-word name
./bin/chatty --select "Data Scientist" # Multi-word name

# Show current assistant
./bin/chatty --current
```

## Chat History

Each assistant maintains their own chat history in `~/.chatty/`:

- Files are named `chat_history_<assistant>.json`
- Histories include:
  - System message that sets the AI's behavior
  - All user messages and AI responses
  - Context preservation between conversations
- Histories are automatically created on first use
- Each assistant maintains independent conversation context

## Customizing Assistants

### Using Sample Assistants

When you first run Chatty, it automatically creates sample assistant configurations in your home directory:

1. Look for `.sample` files in `~/.chatty/assistants/`:

   ```bash
   ls ~/.chatty/assistants/*.sample
   ```

2. Create your own assistant by copying a sample:

   ```bash
   cp ~/.chatty/assistants/focus.yaml.sample ~/.chatty/assistants/myassistant.yaml
   ```

3. Edit the new file to customize your assistant:
   ```bash
   nano ~/.chatty/assistants/myassistant.yaml
   ```

### Assistant Configuration Fields

Each assistant is defined by a YAML file with the following fields:

```yaml
# The name of your assistant (required)
# Can be a single word or multiple words (e.g., "Data Scientist")
# When using multi-word names in commands, remember to quote them:
# ./chatty --select "Data Scientist"
name: "Assistant Name"

# The system message defines your assistant's personality (required)
system_message: |
  You are [assistant description].

  1. Core Identity: [main role and characteristics]
  2. Personality: [personality traits]
  3. Communication: [communication style]
  4. Approach: [how to handle tasks]
  5. Knowledge: [areas of expertise]
  6. Special Focus: [specific strengths]
  7. Boundaries: [limitations and guidelines]

# Choose an emoji that represents your assistant (required)
emoji: "🤖"

# Set the color for your assistant's name (required)
# Use 256-color ANSI codes: \u001b[38;5;XXXm where XXX is 0-255
label_color: "\u001b[38;5;75m" # Blue

# Set the color for your assistant's responses (required)
text_color: "\u001b[38;5;252m" # Light gray

# A brief description of what your assistant does (required)
description: "A brief description of the assistant"
```

### Color Customization

The application supports 256-color ANSI codes for rich terminal colors:

- Format: `\u001b[38;5;XXXm` where XXX is the color code (0-255)
- Each assistant can have unique label and text colors
- Use online ANSI color pickers to find appropriate codes
- Common colors:
  - 82: Bright green
  - 220: Bright yellow
  - 75: Bright blue
  - 213: Bright magenta
  - 252: Light gray

You can find a complete color chart here: https://www.ditig.com/256-colors-cheat-sheet

### Assistant Loading Order

Assistants are loaded in the following order:

1. User-defined assistants (from `~/.chatty/assistants/`)
2. Built-in assistants (if not overridden by user assistants)

This means you can:

- Override built-in assistants by creating a file with the same name
- Create entirely new assistants with unique names
- Keep your custom assistants separate from the application

### Best Practices

When creating custom assistants:

1. Use descriptive names that reflect the assistant's purpose
2. Write clear and focused system messages
3. Choose appropriate emojis that represent the role
4. Use contrasting colors for readability
5. Keep descriptions concise but informative
6. Test the assistant with various queries
7. Maintain conversation context appropriately

### File Naming Conventions

When creating your YAML files in `~/.chatty/assistants/`:

1. Use lowercase letters for the file name: `data_scientist.yaml` (not `Data_Scientist.yaml`)
2. Replace spaces with underscores: `machine_learning_expert.yaml` (not `machine learning expert.yaml`)
3. Use `.yaml` extension (not `.yml`)
4. Keep the file name simple and related to the assistant's name:

   ```bash
   # Good examples:
   python_expert.yaml     # For an assistant named "Python Expert"
   data_scientist.yaml    # For "Data Scientist"
   ml_assistant.yaml      # For "Machine Learning Assistant"

   # Avoid:
   my_assistant_1.yaml    # Not descriptive
   AI_Assistant.yaml      # Don't use uppercase
   machine.learning.yaml  # Don't use dots
   ```

Note: The file name doesn't have to match the assistant's display name exactly. The display name is defined by the `name` field inside the YAML file and can contain spaces and proper capitalization.

## Development

### Building

```bash
# Create a development build
task build

# Install to system
task install

# Run tests
task test

# Clean build artifacts
task clean
```

### Adding Features

1. Create new assistant YAML files in `builtin/` for core assistants
2. Add sample configurations in `samples/` for user reference
3. Implement new features in `main.go` or `assistants.go`
4. Follow the existing code structure and patterns

## Error Handling

The program handles various error cases:

- Invalid command usage
- Ollama server connection issues
- File system errors
- JSON parsing errors
- Invalid assistant selection
- Sample file copying issues

## Notes

- The chat history is maintained between sessions
- The AI maintains context of previous conversations
- Responses are streamed in real-time as they're generated
- Assistant personalities can be switched at any time
- The API endpoint can be changed to connect to remote Ollama instances
- Each assistant maintains their unique personality and style
- Sample assistants are provided as templates for customization
