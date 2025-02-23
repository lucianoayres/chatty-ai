name: "Max"

# Chatty

Chatty is a command-line interface for chatting with Ollama's LLMs (Large Language Models) with streaming responses, conversation history, and customizable AI personalities.

## Prerequisites

1. Install Ollama if you haven't already
2. Start the Ollama server:
   ```bash
   ollama serve
   ```
3. Install Chatty:
   ```bash
   git clone <your-repo-url>
   cd chatty
   go build -o bin/chatty cmd/chatty/main.go
   ```

## Getting Started

1. Initialize Chatty (required on first run):

   ```bash
   chatty init
   ```

   This will:

   - Create the ~/.chatty directory
   - Set up default configuration
   - Create assistants directory
   - Copy sample assistant configurations

2. List available assistants:

   ```bash
   chatty --list
   ```

3. Select an assistant (optional, defaults to Rocket 🚀):

   ```bash
   chatty --select rocket           # Single-word name
   chatty --select "Data Scientist" # Multi-word name
   ```

4. Start chatting:

   ```bash
   # Simple messages (no quotes needed)
   chatty How are you doing today?

   # Messages with special characters (use quotes)
   chatty "What's the meaning of life?"
   ```

## Commands

```bash
chatty init                      # Initialize Chatty environment
chatty --list                    # List available assistants
chatty --select <name>          # Switch assistants
chatty --current                # Show current assistant
chatty --clear                  # Clear all histories
chatty --clear <assistant>      # Clear specific assistant's history
chatty <message>                # Chat with current assistant
```

## Configuration

All settings are stored in `~/.chatty/config.json`:

```json
{
  "current_assistant": "rocket",
  "language_code": "en-US",
  "model": "llama2",
  "common_directives": "Be professional and formal..."
}
```

### Available Settings

- `current_assistant`: The active AI personality
- `language_code`: Language for interactions (default: "en-US")

  - Supported: en-US, es-ES, fr-FR, de-DE, it-IT, pt-BR, ja-JP, ko-KR, zh-CN
  - **Important**: After changing the language code, it's recommended to clear the chat history:

    ```bash
    # Clear history for specific assistant
    chatty --clear <assistant_name>

    # Or clear all histories
    chatty --clear all
    ```

  - This ensures consistent language use, as previous conversations in different languages may influence the assistant's language choice

- `model`: Ollama model to use (default: "llama2")
  - Must be installed in Ollama (use `ollama list` to see available models)
- `common_directives`: Custom personality traits for all assistants

## Built-in Assistants

Chatty comes with pre-configured AI personalities:

- **Rocket** 🚀 - Your friendly coding companion for development guidance and best practices
- **Tux** 🐧 - Your Linux terminal expert for command-line operations and shell scripting
- **Focus** 🎯 - Your efficiency expert for productivity and organization

## Creating Custom Assistants

1. Navigate to sample configurations:

   ```bash
   ls ~/.chatty/assistants/*.sample
   ```

2. Create your assistant:

   ```bash
   cp ~/.chatty/assistants/focus.yaml.sample ~/.chatty/assistants/myassistant.yaml
   ```

3. Edit the configuration:
   ```yaml
   name: "Assistant Name" # Display name
   system_message: | # Core personality
     You are [description]...
   emoji: "🤖" # Visual indicator
   label_color: "\u001b[38;5;75m" # Name color (blue)
   text_color: "\u001b[38;5;252m" # Response color
   description: "Brief description"
   ```

### File Naming Conventions

When creating assistant files:

- Use lowercase: `data_scientist.yaml`
- Use underscores for spaces: `machine_learning_expert.yaml`
- Use `.yaml` extension
- Keep names descriptive:

  ```bash
  # Good examples:
  python_expert.yaml
  data_scientist.yaml
  ml_assistant.yaml

  # Avoid:
  my_assistant_1.yaml
  AI_Assistant.yaml
  machine.learning.yaml
  ```

### Color Customization

Supports 256-color ANSI codes:

- Format: `\u001b[38;5;XXXm` (XXX = 0-255)
- Common colors:
  - 82: Bright green
  - 220: Yellow
  - 75: Blue
  - 213: Magenta
  - 252: Light gray

Color chart: https://www.ditig.com/256-colors-cheat-sheet

## Chat History

- Each assistant maintains separate history
- Stored in `~/.chatty/chat_history_<assistant>.json`
- Includes:
  - System message (personality)
  - All messages and responses
  - Conversation context
- Created automatically on first use
- Clear with `chatty --clear <assistant>` or `chatty --clear all`

## Project Structure

```
chatty/
├── cmd/
│   └── chatty/
│       ├── main.go            # Main application
│       └── assistants/
│           ├── assistants.go  # Assistant management
│           ├── builtin/       # Built-in assistants
│           │   ├── rocket.yaml
│           │   ├── tux.yaml
│           │   └── focus.yaml
│           └── samples/       # Sample configurations
│               └── focus.yaml
└── ~/.chatty/                # User data (created on init)
    ├── config.json          # User settings
    ├── chat_history_*.json  # Conversations
    └── assistants/         # Custom assistants
        └── *.yaml.sample   # Templates
```

## Troubleshooting

1. If you see "Chatty is not initialized":

   ```bash
   chatty init
   ```

2. If you get "invalid model" errors:

   - Check available models: `ollama list`
   - Update model in `~/.chatty/config.json`

3. If the assistant responds in the wrong language after changing `language_code`:

   - Clear the chat history to ensure consistent language use:
     ```bash
     chatty --clear <assistant_name>  # For specific assistant
     chatty --clear all              # For all assistants
     ```
   - The assistant may mix languages if previous conversations exist in different languages

4. To start fresh:

   ```bash
   rm -rf ~/.chatty
   chatty init
   ```

5. If Ollama connection fails:
   - Ensure Ollama is running: `ollama serve`
   - Check if the model is installed: `ollama list`
   - Install missing model: `ollama pull <model>`

## Development

### Prerequisites

- Go 1.16 or later
- Ollama installed and running
- Required model installed in Ollama

### Building

```bash
# Development build
task build

# Install to system
task install

# Run tests
task test

# Clean build artifacts
task clean
```

### Adding Features

1. Create assistant YAML files in `builtin/` for core assistants
2. Add sample configurations in `samples/`
3. Implement features in `main.go` or `assistants.go`
4. Follow existing code patterns

## Notes

- Run `chatty init` before first use
- Chat history persists between sessions
- Context is maintained per assistant
- Responses stream in real-time
- Assistants can be switched anytime
- API endpoint configurable for remote Ollama
- Each assistant maintains unique personality
- Sample templates provided for customization
