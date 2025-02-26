# Chatty AI Agents 🤖

Transform your terminal into a vibrant AI-powered workspace where historical figures, scientists, philosophers, and experts come alive! Chatty isn't just another CLI tool - it's your gateway to engaging conversations with some of history's most fascinating minds.

## ✨ What Makes Chatty Special?

- 🎭 **Rich Character Roster**: From Shakespeare to Einstein, Plato to Marie Curie - engage with personalities who shaped history
- 🗣️ **Multi-Agent Conversations**: Create unique discussions between up to 15 different characters
- 🤖 **Autonomous Mode**: Watch as historical figures and experts discuss topics on their own
- 🌈 **Beautiful Terminal UI**: Real-time streaming responses with custom colors and animations
- 🌍 **Multi-Language Support**: Chat in any language your Ollama model understands
- 📝 **Persistent Memory**: Each agent remembers your conversations

## 🚀 Quick Start

```bash
# Install Ollama first (https://ollama.ai)
ollama serve

# First-time setup (required)
chatty init

# Start a simple chat with the default agent
chatty "What can you do?"

# Switch to a different agent
chatty --select "Einstein"

# Then engage in impossible conversations
chatty "Can you explain the quantum mechanics behind TikTok viral videos, or is it just pure sorcery?"

# Play games with your historical figures
chatty --converse "Shakespeare,Jane Austen,Gandalf" --starter "Let's play Two Truths and a Lie!"

# Create the movie plot of your dreams
chatty --converse "Cleopatra,Asimov,Beethoven" --starter "Encene a movie called Cybor Hamsters: Escape from the Moon" --auto

# Unlearn everything you know about any topic
chatty --converse Zeus,Turing,Tux --starter "Explain how the cloud works, wrong answers only" --auto

# Plan spectacular events
chatty --converse Shakespeare,Aristotle,Dracula --starter "Let's plan the ultimate vampire-themed dinner party"

# Brainstorm revolutionary business ideas
chatty --converse Marx,Tesla,"Sherlock Holmes" --starter "Pitch a revolutionary business idea that combines electricity, detective work, and communism"
```

## Available Agents

### 💻 Built-in Agents

Step into a world of extraordinary conversations with our diverse roster of pre-installed agents! Chat with brilliant minds like **Einstein** about the mysteries of the universe, explore the art of code with **Ada**, or debate philosophy with **Aristotle** and **Plato**. Need technical help? **Tux** and **Nimble** are ready to assist with Linux and productivity. Want something different? Share stories with **Shakespeare**, investigate mysteries with **Sherlock Holmes**, or discuss revolution with **Marx**.

Each agent brings their unique perspective and expertise to the conversation. Use `chatty --list` to see all available agents and their specialties.

### 🎭 Sample Personality Agents

Want even more fascinating conversations? Discover our collection of over 50 additional personalities ready to be brought to life! Explore the cosmos with **Carl Sagan**, unravel the mysteries of consciousness with **Sigmund Freud**, or dive into the depths of gothic literature with **Edgar Allan Poe**. Challenge your perspectives with **Nietzsche**'s philosophical provocations, or get cybersecurity insights from our **Security Expert**.

To start using these personalities:

```bash
# List all available samples
ls ~/.chatty/agents/*.sample

# Pick your favorite and create your own copy
cp ~/.chatty/agents/sagan.yaml.sample ~/.chatty/agents/sagan.yaml
```

The possibilities are endless - mix and match personalities to create unique conversations that span across time, disciplines, and perspectives!

## 🎯 Cool Use Cases

### 🔮 Time-Traveling Discussions

```bash
# First, set up your desired personality agents
cp ~/.chatty/agents/sagan.yaml.sample ~/.chatty/agents/sagan.yaml
cp ~/.chatty/agents/hawking.yaml.sample ~/.chatty/agents/hawking.yaml
cp ~/.chatty/agents/captain_nemo.yaml.sample ~/.chatty/agents/captain_nemo.yaml

# Then start your discussion
chatty --converse "Carl Sagan","Stephen Hawking","Captain Nemo" --starter "Let's explore the mysteries of space and sea - which frontier is more fascinating?"
```

### 🎓 Learning & Exploration

```bash
# Using built-in agents
chatty --converse "Ada","Nimble" --starter "How to organize a complex software project?"

# Using personality agents (after setting them up)
chatty --converse "Feynman","Marie Curie","Darwin" --starter "Explain how scientific discovery happens"
```

### 🎭 Entertainment & Creativity

```bash
# Unlikely Collaborations
chatty --converse "Shakespeare","Mark Twain","Jane Austen" --starter "Write a story about time travel"

# Cultural Conversations
chatty --converse "Mozart","Louis Armstrong","Elvis Presley" --starter "Create a new musical genre"

# Philosophical Debates
chatty --converse "Socrates","Kant","Albert Camus" --starter "Is social media making us happier?"
```

## 💭 Chat Modes

### 👤 Single Agent Conversations

Have a one-on-one chat with any agent:

```bash
# List available agents and their specialties
chatty --list

# Choose your conversation partner
chatty --select "Sherlock Holmes"

# Get creative with your questions
chatty "Analyze my coffee stains and deduce my morning routine"

# Check who you're currently talking to
chatty --current

# Clear chat history for a fresh start
chatty --clear "Shakespeare"
```

**Note**: When using `--select`, the agent name must exactly match the `name` field in the agent's YAML file. For example, if the YAML contains `name: "Marie Curie"`, you must use `--select "Marie Curie"` (including quotes if the name contains spaces).

### 🤝 Multi-Agent Conversations

Create interactive discussions between AI agents:

```bash
# Basic examples - single-word names don't need quotes
chatty --converse Ada,Tux --starter "Let's discuss Linux development"
chatty --converse Einstein,Newton,Darwin --starter "Let's discuss gravity"

# When an agent has multiple words in their name, use quotes
chatty --converse "Marie Curie",Einstein,Darwin --starter "Discuss scientific method"
chatty --converse "Edgar Allan Poe","Mark Twain",Shakespeare --starter "Write a story"

# You can also use quotes for all names (recommended for consistency)
chatty --converse "Einstein","Newton","Darwin" --starter "How would you explain gravity?"

# More examples
chatty --converse Ada,Tux,Nimble --starter "How can we improve code quality?"
chatty --converse "Marie Curie","Ada Lovelace",Einstein --starter "Women in science"
chatty --converse Mozart,"Louis Armstrong","Elvis Presley" --starter "Future of music"

# Using special characters (escape with \)
chatty --converse Ada,Tux --starter "How to make \$100 last a month?"

# Auto mode and turn limits work the same way
chatty --converse Einstein,Newton --starter "Discuss gravity" --auto
chatty --converse "Marie Curie",Einstein --starter "Future of physics" --turns 5
```

**Note**: When using `--converse`, you can either:

- Skip quotes for single-word names: `Ada,Tux,Einstein`
- Use quotes only for multi-word names: `"Marie Curie",Einstein,Newton`
- Use quotes for all names (recommended): `"Einstein","Newton","Darwin"`

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
