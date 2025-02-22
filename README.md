# Chatty

Chatty is a command-line interface for chatting with Ollama's LLMs (Large Language Models) with streaming responses, conversation history, and customizable AI personalities.

## Features

- Real-time streaming responses
- Persistent conversation history
- Multiple AI personalities to choose from
- Customizable appearance and behavior
- Clear and readable output format
- Colored output (configurable)
- Consistent text margins

## Available Assistants

Chatty comes with several pre-configured AI personalities:

1. **Ghostly** (Default) - A friendly and ethereal presence, helping with a gentle touch
2. **Sage** - A wise mentor focused on deep understanding and guidance
3. **Nova** - A tech-savvy assistant with a passion for innovation
4. **Terra** - An eco-conscious assistant promoting sustainability
5. **Atlas** - A structured assistant focusing on organization and planning
6. **Tux** - A Linux terminal expert specializing in command-line operations and shell scripting

Each assistant comes with:

- Unique personality and communication style
- Custom emoji and color scheme
- Specialized system message
- Distinct area of expertise

## Prerequisites

- Go 1.16 or later
- Ollama installed and running locally
- The `llama3.2` model installed in Ollama (or modify the model name in constants)

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
./bin/chatty --clear

# List available assistants
./bin/chatty --list

# Switch to a different assistant
./bin/chatty --select Sage
```

### Configuration

All configuration is done through constants and the assistant configuration system:

```go
// Core configuration
ollamaModel   = "llama3.2"               // Model to use for chat
ollamaBaseURL = "http://localhost:11434"  // Base URL for Ollama API
ollamaURLPath = "/api/chat"              // API endpoint path
historyFile   = "chat_history.json"      // File to store chat history

// Display settings
useEmoji     = true    // Enable/disable emoji display
useColors    = true    // Enable/disable colored output
topMargin    = 1      // Blank lines before response
bottomMargin = 1      // Blank lines after response
```

Each assistant's configuration includes:

- Name
- System message template
- Emoji
- Label color
- Text color
- Description

### Chat History

Each assistant maintains their own chat history file in the `~/.chatty/` directory:

- Files are named `chat_history_<assistant>.json` (e.g., `chat_history_ghostly.json`)
- Each history includes:
  - A system message that sets the AI's behavior
  - All user messages and AI responses in chronological order
  - Context preservation between conversations
- When switching assistants:
  - The program automatically loads the correct history file
  - If no history exists for an assistant, a new one is created
  - Each assistant maintains their own conversation context
- When adding new assistants to the code:
  - A new history file is automatically created on first use
  - No manual setup is required
  - Each new assistant starts with a fresh conversation

## Error Handling

The program handles various error cases:

- Invalid command usage
- Ollama server connection issues
- File system errors
- JSON parsing errors
- Invalid assistant selection

## Building

To build a standalone executable:

```bash
# Create the bin directory if it doesn't exist
mkdir -p bin

# Build the executable
go build -o bin/chatty cmd/chatty/main.go
```

Then you can run it directly:

```bash
./bin/chatty "Your question here"
# or
./bin/chatty How are you today?
```

## Notes

- The chat history is maintained between sessions
- The AI maintains context of previous conversations
- Responses are streamed in real-time as they're generated
- All configuration can be easily modified through constants
- Colors can be disabled by setting `useColors = false`
- Emoji can be disabled by setting `useEmoji = false`
- Assistant personalities can be switched at any time
- The API endpoint can be changed to connect to remote Ollama instances
- Each assistant maintains their unique personality and style
- System messages are automatically updated when switching assistants
