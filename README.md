# Chatty 🤖

Transform your terminal into a vibrant AI-powered workspace where historical figures, scientists, philosophers, and experts come alive! Chatty isn't just another CLI tool - it's your gateway to engaging conversations with some of history's most fascinating minds.

## ✨ What Makes Chatty Special?

- 🎭 **Rich Character Roster**: From Shakespeare to Einstein, Plato to Marie Curie - engage with personalities who shaped history
- 🗣️ **Multi-Agent Conversations**: Create unique discussions between up to 15 different characters
- 🤖 **Autonomous Mode**: Watch as historical figures and experts discuss topics on their own
- 🌈 **Beautiful Terminal UI**: Real-time streaming responses with custom colors and animations
- 🌍 **Multi-Language Support**: Chat in any language your Ollama model understands
- 📝 **Persistent Memory**: Each agent remembers your conversations

## 🎭 Available Agents

### 💻 Built-in Agents

These come pre-installed and ready to use:

- **Ada** 💻 - Software development expert with algorithmic mastery
- **Aristotle** 🏛️ - Ancient Greek philosopher and polymath
- **Byte** 🤖 - Default agent for general assistance
- **Cleopatra** 👑 - Last active ruler of the Ptolemaic Kingdom of Egypt
- **Curie** ⚛️ - Pioneering physicist and chemist
- **Einstein** 🧮 - Revolutionary physicist who redefined our understanding of the universe
- **Nimble** 🎯 - Task-focused productivity assistant
- **Tesla** ⚡ - Visionary inventor and electrical engineer
- **Tux** 🐧 - Linux and command-line specialist
- **Twain** ✍️ - American humorist and social critic

### 🎭 Sample Personality Agents

Chatty comes with over 50 fascinating personalities available as templates in `~/.chatty/agents/*.sample`. Below are some examples from each category. To use any of them, you'll need to create your own copies:

```bash
# List available samples
ls ~/.chatty/agents/*.sample

# Create your own agent from a sample
cp ~/.chatty/agents/einstein.yaml.sample ~/.chatty/agents/einstein.yaml
```

#### 🔬 Scientific Minds (Examples)

- **Richard Feynman**: Physics explained with clarity and enthusiasm
- **Marie Curie**: Methodical approach to discovery
- **Carl Sagan**: Bringing cosmic wonder to your terminal

#### 📚 Literary Giants (Examples)

- **William Shakespeare**: Poetic mastery and character insight
- **Jane Austen**: Sharp social observation and wit
- **Edgar Allan Poe**: Master of mystery and psychological depth

#### 🏛️ Philosophers & Thinkers (Examples)

- **Plato**: Ancient wisdom for modern questions
- **Nietzsche**: Challenging conventional thinking
- **Simone de Beauvoir**: Feminist philosophy and existentialism

#### 💻 Technical Experts (Examples)

- **Software Architect**: System design wisdom
- **Security Expert**: Cybersecurity insights

## 🚀 Quick Start

```bash
# Install Ollama first (https://ollama.ai)
ollama serve

# Initialize Chatty
chatty init

# Start a simple chat
chatty "Tell me about quantum physics"

# Create a fascinating discussion
chatty --converse "Richard Feynman","Carl Sagan","Albert Einstein" --starter "Explain quantum entanglement to a child" --auto

# Host a literary salon
chatty --converse "William Shakespeare","Jane Austen","Edgar Allan Poe" --starter "How would you approach writing in the digital age?"
```

## 🎯 Cool Use Cases

### 🔮 Time-Traveling Discussions

```bash
# First, set up your desired personality agents
cp ~/.chatty/agents/newton.yaml.sample ~/.chatty/agents/newton.yaml
cp ~/.chatty/agents/tesla.yaml.sample ~/.chatty/agents/tesla.yaml
cp ~/.chatty/agents/turing.yaml.sample ~/.chatty/agents/turing.yaml

# Then start your discussion
chatty --converse "Isaac Newton","Nikola Tesla","Alan Turing" --starter "What would you think of modern AI?"
```

### 🎓 Learning & Exploration

```bash
# Using built-in agents
chatty --converse "Ada","Nimble" --starter "How to organize a complex software project?"

# Using personality agents (after setting them up)
chatty --converse "Richard Feynman","Marie Curie","Charles Darwin" --starter "Explain how scientific discovery happens"
```

### 🎭 Entertainment & Creativity

```bash
# Unlikely Collaborations
chatty --converse "William Shakespeare","Mark Twain","Jane Austen" --starter "Write a story about time travel"

# Cultural Conversations
chatty --converse "Wolfgang Amadeus Mozart","Louis Armstrong","Elvis Presley" --starter "Create a new musical genre"

# Philosophical Debates
chatty --converse "Socrates","Immanuel Kant","Albert Camus" --starter "Is social media making us happier?"
```

## 💭 Chat Modes

### 👤 Single Agent Conversations

Have a one-on-one chat with any agent:

```bash
# Chat with the default agent (Byte)
chatty "Help me understand Docker containers"

# Switch to a different agent
chatty --select "Richard Feynman"
chatty "Can you explain quantum mechanics to me?"

# Quick examples with different agents
chatty --select "William Shakespeare" "What makes a great story?"
chatty --select "Marie Curie" "Tell me about your research"
chatty --select "Plato" "What is justice?"

# See who you're talking to
chatty --current

# List all available agents (shows exact names to use with --select)
chatty --list
```

**Note**: When using `--select`, the agent name must exactly match the `name` field in the agent's YAML file. For example, if the YAML contains `name: "Marie Curie"`, you must use `--select "Marie Curie"` (including quotes if the name contains spaces).

### 🤝 Multi-Agent Conversations

Create interactive discussions between AI agents:

```bash
# Basic conversation (2-15 agents)
chatty --converse "Ada","Tux" --starter "Let's discuss Linux development"

# Three-way conversation
chatty --converse "Ada","Tux","Nimble" --starter "How can we improve code quality?"

# Limited turns (stop after N turns)
chatty --converse "Ada","Tux" --starter "Discuss AI trends" --turns 3

# Autonomous conversation (agents talk among themselves)
chatty --converse "Ada","Tux","Nimble" --starter "Debate software architecture" --auto

# Auto conversation with turn limit
chatty --converse "Ada","Tux" --starter "Explore cloud computing" --auto --turns 5

# Using special characters (escape with \)
chatty --converse "Ada","Tux" --starter "How to build a startup with \$100?"
```

**Note**: Agent names in the `--converse` argument must be comma-separated and match exactly with their YAML file names. Use quotes for names containing spaces.

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

## ⚙️ Configuration

Settings are stored in `~/.chatty/config.json`:

```json
{
  "current_agent": "byte",
  "language_code": "en-US",
  "model": "llama3.2",
  "common_directives": "Be professional and formal..."
}
```

### Available Settings

- `current_agent`: Active AI personality (defaults to "byte")
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

1. Start with a sample:

   ```bash
   # List available samples
   ls ~/.chatty/agents/*.sample

   # Create from a sample
   cp ~/.chatty/agents/focus.yaml.sample ~/.chatty/agents/myagent.yaml
   ```

2. Or create from scratch:
   ```yaml
   name: "Agent Name"
   system_message: |
     You are [description]...
   emoji: "🤖"
   label_color: "\u001b[38;5;75m" # Blue
   text_color: "\u001b[38;5;252m" # Light gray
   description: "Brief description"
   is_default: false
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

Found a bug? Have a cool idea? We'd love your help!

- Open an issue to report bugs or suggest features
- Submit pull requests to improve the code
- Star the repo if you find it useful

## 📄 License

MIT License - Feel free to use, modify, and share!
Copyright (c) 2024 [Your Name/Organization]
