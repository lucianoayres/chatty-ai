# Chatty

Chatty is a command-line interface for chatting with Ollama's LLMs (Large Language Models) with streaming responses and conversation history.

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
go run main.go How are you doing today?
```

2. With quotes (required for messages containing special characters):

```bash
go run main.go "You're an AI, right?"
go run main.go "What's the meaning of life, the universe & everything?"
```

The program will:

- Stream the response in real-time
- Save the conversation history automatically
- Maintain context between messages
- Format output with consistent spacing and colors

### Special Commands

Clear conversation history:

```bash
go run main.go --clear
```

### Features

- Real-time streaming responses
- Persistent conversation history
- System message context
- Clear and readable output format
- Colored output (configurable)
- Consistent text margins

### Configuration

All configuration is done through constants in the code:

```go
// Appearance
topMargin   = 1       // Blank lines before response
bottomMargin = 1      // Blank lines after response
useColors = true      // Enable/disable colored output
assistantLabel = "Assistant"  // Label shown before responses

// Colors (RGB format)
assistantLabelColor = "\033[38;2;79;195;247m"   // #4FC3F7 (light blue)
assistantTextColor  = "\033[38;2;255;255;255m"  // #FFFFFF (white)

// AI Behavior
systemMessage = "You are a helpful AI assistant. Be concise and clear in your responses."

// Ollama API
ollamaBaseURL = "http://localhost:11434"  // Base URL for Ollama API
ollamaURLPath = "/api/chat"               // API endpoint path
ollamaModel   = "llama3.2"                // Model to use for chat
```

### Chat History

The chat history is automatically saved in `~/.chatty/chat_history.json`. The history includes:

- A system message that sets the AI's behavior
- All user messages and AI responses in chronological order
- Context preservation between conversations

## Error Handling

The program handles various error cases:

- Invalid command usage
- Ollama server connection issues
- File system errors
- JSON parsing errors

## Building

To build a standalone executable:

```bash
go build -o chatty
```

Then you can run it directly:

```bash
./chatty "Your question here"
# or
./chatty How are you today?
```

## Notes

- The chat history is maintained between sessions
- The AI maintains context of previous conversations
- Responses are streamed in real-time as they're generated
- All configuration can be easily modified through constants
- Colors can be disabled by setting `useColors = false`
- The API endpoint can be changed to connect to remote Ollama instances
