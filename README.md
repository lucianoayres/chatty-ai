# Chatty 🤖

Transform your terminal into an AI-powered workspace with Chatty - your command-line companion for seamless interactions with Ollama's LLMs. Whether you're coding, learning, exploring ideas, or just having fun, Chatty brings the power of AI to your fingertips with real-time responses and intelligent conversations.

### 🎯 Perfect for:

- **Development & Tech**: Code reviews, debugging, system administration, and learning programming
- **Creative Work**: Writing, brainstorming, and content creation
- **Learning & Discussion**: Interactive learning sessions and multi-perspective problem solving
- **Entertainment**: Creating fun dialogues between AI personalities, historical figures, or fictional characters

### 💡 Example Use Cases:

#### 🔧 Professional Use

```bash
# Get code review and improvements
chatty "Review this function: function calculateTotal(items) { ... }"

# Learn about complex concepts
chatty "Explain Docker networking in simple terms"

# Multi-expert problem solving
chatty --converse rocket tux --starter "How can we optimize this Python script for Linux systems?"

# Interactive learning sessions
chatty --converse rocket focus --starter "Teach me about design patterns" --turns 5

# Autonomous expert discussion (agents talk among themselves)
chatty --converse rocket tux focus --starter "Discuss modern development practices" --auto
```

#### 🎭 Historical & Philosophical Discussions

> **Note**: These examples assume you've created custom agents for each character (see [Creating Custom Agents](#%EF%B8%8F-creating-custom-agents)). They're meant to inspire you with possible conversation scenarios!

```bash
# Create a debate about modern technology
chatty --converse socrates plato --starter "How would social media impact society?"

# Autonomous discussion about climate change
chatty --converse tesla einstein darwin --starter "How would you address global warming?" --auto --turns 10

# Explore economic theories
chatty --converse adam_smith keynes marx --starter "Analyze cryptocurrency's impact on modern economics"

# Debate ethics in AI
chatty --converse kant aristotle confucius --starter "What are the moral implications of artificial intelligence?"
```

#### 🎬 Entertainment & Fun

```bash
# Movie character mashups
chatty --converse sherlock yoda batman --starter "Solve the mystery of the missing cookies"

# Absurd historical meetings
chatty --converse shakespeare elvis beethoven --starter "Create a modern pop song"

# Fictional problem solving
chatty --converse gandalf ironman doctorwho --starter "How to deal with a dragon in Manhattan?"

# Time-traveling discussions
chatty --converse leonardo_da_vinci steve_jobs --starter "Design the next iPhone"

# Unlikely cooking show
chatty --converse gordon_ramsay shakespeare --starter "Create a recipe for modern fast food"
```

## ✨ Features

- 🎭 Customizable AI personalities
- 💬 Multi-agent conversations (up to 15 agents)
- 📝 Persistent chat history
- 🔄 Real-time streaming responses
- 🌍 Multi-language support
- 🚀 Easy to use and configure

## 🚀 Quick Start

### Prerequisites

1. Install [Ollama](https://ollama.ai)
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

### First Run Setup

```bash
# Initialize Chatty (required on first run)
chatty init
```

This will:

- Create your ~/.chatty directory
- Set up default configuration
- Install sample AI agents
- Prepare everything for your first chat

### Basic Usage

```bash
# Start chatting (use quotes for messages)
chatty "Hello, how can you help me today?"

# Show current agent
chatty --current

# List available agents
chatty --list

# Switch agents
chatty --select rocket           # Use Rocket (default)
chatty --select "Data Scientist" # Use Data Scientist

# Clear chat history
chatty --clear all              # Clear all histories
chatty --clear rocket           # Clear specific agent's history
```

## 🤝 Multi-Agent Conversations

Create interactive discussions between AI agents:

```bash
# Basic conversation (2-15 agents)
chatty --converse rocket tux --starter "Let's discuss Linux development"

# Three-way conversation
chatty --converse rocket tux focus --starter "How can we improve code quality?"

# Limited turns (stop after N turns)
chatty --converse rocket tux --starter "Discuss AI trends" --turns 3

# Autonomous conversation (agents talk among themselves)
chatty --converse rocket tux focus --starter "Debate software architecture" --auto

# Auto conversation with turn limit
chatty --converse rocket tux --starter "Explore cloud computing" --auto --turns 5

# Using special characters (escape with \)
chatty --converse rocket tux --starter "How to build a startup with \$100?"
```

### How Conversations Work

1. First turn starts with your starter message
2. Each agent responds in sequence (no duplicates allowed)
3. After each turn:
   - In normal mode: you're prompted for a new message
   - In auto mode (--auto): agents continue the conversation automatically
4. Conversation ends when:
   - Specified number of turns is reached (if --turns used)
   - You press Ctrl+C
   - In normal mode: you enter an empty message
   - In auto mode: you press Ctrl+C to stop

### Conversation Display

Each turn shows:

```
────────────────────────────────────────────────────────────
💭 Conversation Turn 1
Started: 2024-03-20 15:30:45 GMT-3
Elapsed: 2 hours, 15 minutes
────────────────────────────────────────────────────────────
```

The display includes:

- Turn number
- Start time in your local timezone
- Elapsed time in human-readable format

## 🎨 Built-in Agents

Chatty comes with pre-configured AI personalities:

- **Rocket** 🚀 - Your friendly coding companion for development guidance
- **Tux** 🐧 - Your Linux terminal expert for command-line operations
- **Focus** 🎯 - Your efficiency expert for productivity and organization

## ⚙️ Configuration

Settings are stored in `~/.chatty/config.json`:

```json
{
  "current_agent": "rocket",
  "language_code": "en-US",
  "model": "llama3.2",
  "common_directives": "Be professional and formal..."
}
```

### Available Settings

- `current_agent`: Active AI personality (defaults to "rocket")
- `language_code`: Language for interactions (default: "en-US")
- `model`: Ollama model to use (default: "llama3.2")
- `common_directives`: Custom personality traits for all agents

### Language Support

Chatty supports any language that your Ollama model can understand. Use the following format in your config.json:

Example language codes:

- English: `en-US`
- Spanish: `es-ES`
- French: `fr-FR`
- German: `de-DE`
- Italian: `it-IT`
- Portuguese: `pt-BR`
- Japanese: `ja-JP`
- Korean: `ko-KR`
- Chinese: `zh-CN`

To change language:

1. Edit `language_code` in config.json using the appropriate language code format (xx-XX)
2. Clear history: `chatty --clear all`

**Note**: Clear history after changing language to ensure consistent responses.

## 🛠️ Creating Custom Agents

1. Check sample configurations:

   ```bash
   ls ~/.chatty/agents/*.sample
   ```

2. Create your agent:

   ```bash
   cp ~/.chatty/agents/focus.yaml.sample ~/.chatty/agents/myagent.yaml
   ```

3. Edit the configuration:
   ```yaml
   name: "Agent Name"
   system_message: |
     You are [description]...
   emoji: "🤖"
   label_color: "\u001b[38;5;75m" # Blue
   text_color: "\u001b[38;5;252m" # Light gray
   description: "Brief description"
   is_default: false # Optional: set as default agent
   ```

### File Naming Conventions

- Use lowercase: `data_scientist.yaml`
- Use underscores for spaces: `machine_learning_expert.yaml`
- Use `.yaml` extension

### Color Customization

Use 256-color ANSI codes: `\u001b[38;5;XXXm` (XXX = 0-255)

Common colors:

- 82: Bright green
- 220: Yellow
- 75: Blue
- 213: Magenta
- 252: Light gray

[Color Chart](https://www.ditig.com/256-colors-cheat-sheet)

## 🔍 Troubleshooting

1. **"Chatty is not initialized"**

   ```bash
   chatty init
   ```

2. **"Invalid model" errors**

   - Check models: `ollama list`
   - Update model in `~/.chatty/config.json`
   - Default model is "llama3.2"

3. **"Too many agents"**

   - Maximum 15 agents per conversation
   - Error: "too many agents: maximum allowed is 15, but got X"

4. **"Duplicate agent"**

   - Each agent can only be included once
   - Error: "duplicate agent detected: X (each agent can only be included once)"

5. **"Connection timed out"**

   - Check if Ollama is running: `ollama serve`
   - Default timeouts: 30s (regular chat), 300s (conversations)
   - Error includes time waited: "connection timed out after X seconds"

6. **Wrong language after changing `language_code`**

   ```bash
   chatty --clear all  # Clear all histories
   ```

7. **Start fresh**

   ```bash
   rm -rf ~/.chatty
   chatty init
   ```

8. **Ollama connection fails**
   - Ensure Ollama is running: `ollama serve`
   - Check model: `ollama list`
   - Install model: `ollama pull <model>`
   - Default URL: http://localhost:11434

## 🤝 Contributing

We welcome contributions! Here's how you can help:

1. **Report Issues**: Found a bug or have a suggestion? Open an issue!
2. **Submit PRs**: Code improvements are always welcome
3. **Share Ideas**: Join discussions in the issues section
4. **Spread the Word**: Star the repo if you find it useful

## 📄 License

MIT License

Copyright (c) 2024 [Your Name/Organization]

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

## 👩‍💻 Development

### Prerequisites

- Go 1.16 or later
- Ollama installed and running
- Required model installed in Ollama

### Project Structure

```
chatty/
├── cmd/
│   └── chatty/
│       ├── main.go            # Main application
│       └── agents/        # Agent configurations
├── bin/                       # Build output
└── README.md                  # Documentation
```

### Build Commands

```bash
# Build for current platform
go build -o bin/chatty cmd/chatty/main.go

# Run tests
go test ./...
```
