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

## Available Assistants

- **Rocket** 🚀 - Your friendly coding companion for development guidance and best practices
- **Fin** 📈 - Your thoughtful guide for investment and financial planning
- **Flex** 💪 - Your motivating companion for fitness and exercise guidance
- **Zen** 🧘 - Your peaceful guide for mindfulness and mental wellness
- **Max** 🎯 - Your efficiency expert for productivity and organization
- **Sage** 📚 - Your dedicated companion for learning and academic growth
- **Nova** 🌟 - Your friendly guide through the world of technology
- **Vita** 🥗 - Your friendly guide for nutrition and healthy eating habits

## Project Structure

```
chatty/
├── cmd/
│   └── chatty/
│       ├── main.go            # Main application code
│       └── assistants/
│           ├── assistants.go  # Assistant management
│           └── builtin/       # Built-in assistant configurations
│               ├── rocket.yaml   # Coding assistant
│               ├── fin.yaml      # Financial advisor
│               ├── flex.yaml     # Fitness trainer
│               ├── zen.yaml      # Mindfulness guide
│               ├── max.yaml      # Productivity expert
│               ├── sage.yaml     # Educational tutor
│               ├── nova.yaml     # Tech guide
│               └── vita.yaml     # Nutrition expert
└── ~/.chatty/                # User data directory (created automatically)
    ├── config.json          # Current assistant selection
    └── chat_history_*.json  # Conversation histories
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

# List available assistants
./bin/chatty --list

# Switch to a different assistant
./bin/chatty --select nova

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

1. Create new assistant YAML files in `builtin/` if needed
2. Implement new features in `main.go` or `assistants.go`
3. Follow the existing code structure and patterns

## Error Handling

The program handles various error cases:

- Invalid command usage
- Ollama server connection issues
- File system errors
- JSON parsing errors
- Invalid assistant selection

## Notes

- The chat history is maintained between sessions
- The AI maintains context of previous conversations
- Responses are streamed in real-time as they're generated
- Assistant personalities can be switched at any time
- The API endpoint can be changed to connect to remote Ollama instances
- Each assistant maintains their unique personality and style
