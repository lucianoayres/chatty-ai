# 💬 Chatty AI · Your Terminal Time Machine

![Chat AI Banner](images/chatty_ai_banner.png)

## Chat with Historical Figures, Scientists, Experts and more

Transform your terminal into a vibrant AI-powered workspace where historical figures, scientists, philosophers, and experts come alive! Chatty isn't just another CLI tool - it's your gateway to engaging conversations with some of history's most fascinating minds.

## ✨ What Makes Chatty Special?

- 🎭 **Rich Character Roster**: From Shakespeare to Einstein, Plato to Marie Curie - engage with personalities who shaped history
- 🗣️ **Multi-Agent Conversations**: Create unique discussions between up to 15 different characters
- 🎲 **Random Conversations**: Let fate decide your conversation partners for unexpected and exciting discussions
- 🤖 **Autonomous Mode**: Watch as historical figures and experts discuss topics on their own
- 📝 **Persistent Memory**: Each agent remembers your conversations
- 🌍 **Multi-Language Support**: Chat in multiple languages with your AI models

## Screenshot

![Chatty Screenshot](images/chatty_ai_screenshot_01.png)

## 🔧 Prerequisites

Chatty requires [Ollama](https://ollama.ai) to run the AI models.

## 🚀 Quick Start

After installing Ollama, pull the llama3.2 model:

```bash
# Pull the llama3.2 model
ollama pull llama3.2

# Start the Ollama service
ollama serve
```

Clone the repository:

```bash
git clone https://github.com/chatty-ai/chatty.git
```

Copy the `chatty` binary to `/usr/local/bin`:

```bash
sudo cp bin/chatty /usr/local/bin
```

Start Chatty:

```bash
# First-time setup (required)
chatty init
```

Try one of the various modes:

```bash
# Send a one-off message to the default agent (not a chat session)
chatty "What can you do?"

# Switch to a different agent
chatty --select "Einstein"

# Then engage in impossible conversations
chatty "Can you explain the quantum mechanics behind TikTok viral videos, or is it just pure sorcery?"

# Start a direct chat with a specific agent
chatty --with "Freud"

# Group chat with multiple agents
chatty --with "Shakespeare,Jane Austen,Gandalf"

# Play games with your favorite historical figures
chatty --with "Cleopatra,Asimov,Beethoven" --topic "Let's play Two Truths and a Lie" --auto

# Let fate decide your conversation partners
chatty --with-random 4 --topic "Let's have a completely unexpected discussion!" --auto

# Use a text file as the conversation starter
chatty --with "Plato,Aristotle" --topic-file ~/my_philosophy_questions.txt

# Unlearn everything you know about any topic
chatty --with "Zeus,Turing,Tux" --topic "Explain how the cloud works, wrong answers only" --auto

# Plan spectacular events
chatty --with "Shakespeare,Aristotle,Dracula" --topic "Let's plan the ultimate vampire-themed dinner party"

# Brainstorm revolutionary business ideas
chatty --with "Marx,Tesla,Sherlock Holmes" --topic "Pitch a revolutionary business idea that combines electricity, detective work, and communism"
```

## Available Agents

![Chatty Banner](images/chatty_ai_small_banner_01.png)

### 💻 Built-in Agents

Step into a world of extraordinary conversations with our diverse roster of [pre-installed agents](cmd/chatty/agents)! Chat with brilliant minds like **Einstein** about the mysteries of the universe, explore the art of code with **Ada**, or debate philosophy with **Aristotle** and **Plato**. Need technical help? **Tux** and **Nimble** are ready to assist with Linux and productivity. Want something different? Share stories with **Shakespeare**, investigate mysteries with **Sherlock Holmes**, or discuss revolution with **Marx**.

Each agent brings their unique perspective and expertise to the conversation. Use `chatty --list` to see all available agents and their specialties.

### 🎭 Sample Personality Agents

Want even more fascinating conversations? Discover our collection of over [50 additional sample personalities](cmd/chatty/agents/samples) ready to be brought to life! Explore the cosmos with **Carl Sagan**, unravel the mysteries of consciousness with **Sigmund Freud**, or dive into the depths of gothic literature with **Edgar Allan Poe**. Challenge your perspectives with **Nietzsche**'s philosophical provocations, or get cybersecurity insights from our **Security Expert**.

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
chatty --with "Carl Sagan,Stephen Hawking,Captain Nemo" --topic "Let's explore the mysteries of space and sea - which frontier is more fascinating?"
```

### 🎓 Learning & Exploration

```bash
# Using built-in agents
chatty --with "Ada,Nimble" --topic "How to organize a complex software project?"

# Using personality agents (after setting them up)
chatty --with "Feynman,Marie Curie,Darwin" --topic "Explain how scientific discovery happens"
```

### 🎭 Entertainment & Creativity

```bash
# Unlikely Collaborations
chatty --with "Shakespeare,Mark Twain,Jane Austen" --topic "Write a story about time travel"

# Cultural Conversations
chatty --with "Mozart,Louis Armstrong,Elvis Presley" --topic "Create a new musical genre"

# Philosophical Debates
chatty --with "Socrates,Kant,Albert Camus" --topic "Is social media making us happier?"
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

# Start a direct chat with a specific agent
chatty --with "Einstein"

# Start a direct chat with an initial topic
chatty --with "Einstein" --topic "Explain relativity to a 5-year-old"

# Save chat logs to a file
chatty "What's the meaning of life?" --save meaning_of_life.txt

# Check who you're currently talking to
chatty --current

# Clear chat history for a fresh start
chatty --clear "Shakespeare"
```

**Note**: When using `--select` or `--with`, the agent name must exactly match the `name` field in the agent's YAML file. For example, if the YAML contains `name: "Marie Curie"`, you must use `--select "Marie Curie"` (including quotes if the name contains spaces).

### 🤝 Multi-Agent Conversations

Create interactive discussions between AI agents:

```bash
# Basic examples - comma-separated list of agents
chatty --with "Ada,Tux"
# You'll be prompted to enter a topic to start the conversation

# Provide a topic directly
chatty --with "Einstein,Newton,Darwin" --topic "Let's discuss gravity"

# When an agent has multiple words in their name, include them in the comma-separated list
chatty --with "Marie Curie,Einstein,Darwin" --topic "Discuss scientific method"
chatty --with "Edgar Allan Poe,Mark Twain,Shakespeare" --topic "Write a story"

# Random agent conversations - let fate decide!
chatty --with-random 3
# You'll be prompted to enter a topic to start the conversation

# Random agents with a specific topic and autonomous mode
chatty --with-random 5 --topic "Brainstorm crazy ideas" --auto

# Limit the number of conversation turns
chatty --with-random 4 --topic "What's the meaning of life?" --turns 10

# Save conversation logs to a file
# This is an experimental feature and performance may vary
chatty --with "Einstein,Newton" --topic "Discuss gravity" --save gravity_discussion.txt
chatty --with-random 3 --topic "Brainstorm ideas" --auto --save brainstorm.txt

# More examples
chatty --with "Ada,Tux,Nimble" --topic "How can we improve code quality?"
chatty --with "Marie Curie,Ada Lovelace,Einstein" --topic "Women in science"
chatty --with "Mozart,Louis Armstrong,Elvis Presley" --topic "Future of music"

# Using special characters (escape with \)
chatty --with "Ada,Tux" --topic "How to make \$100 last a month?"

# Use a text file as the topic message
chatty --with "Shakespeare,Tolkien" --topic-file story_prompt.txt
chatty --with-random 3 --topic-file research_topic.txt

# Auto mode and turn limits work the same way
chatty --with "Einstein,Newton" --topic "Discuss gravity" --auto
chatty --with "Marie Curie,Einstein" --topic "Future of physics" --turns 5
```

**Note**: When using `--with`, provide a comma-separated list of agent names. For multi-word agent names, include them in the list with the commas.

When using `--with-random`, just specify the number of agents (between 2 and 15) you want in the conversation. The agents will be randomly selected from both built-in and user-defined agents.

You can provide the topic message in two ways:

- Using `--topic "message"` for direct text input
- Using `--topic-file path` to read the message from a text file

If you don't provide a topic with the `--topic` or `--topic-file` flags in interactive mode, you'll be prompted to enter one when the conversation starts.

### How Conversations Work

1. First turn starts with your topic message
2. Each agent responds in sequence (no duplicates allowed)
3. After each turn:
   - In normal mode: you're prompted for a new message
   - In auto mode (--auto): agents continue the conversation automatically
4. Conversation ends when:
   - Specified number of turns is reached (if --turns used)
   - You press Ctrl+C
   - In normal mode: you enter an empty message
   - In auto mode: you press Ctrl+C to stop

## ⚙️ Configuration

Settings are stored in `~/.chatty/config.json`. The default configuration includes only required fields:

```json
{
  "current_agent": "chatty",
  "language_code": "en-US",
  "model": "llama3.2"
}
```

### Available Settings

Required fields:

- `current_agent`: Active AI personality (defaults to "chatty")
- `language_code`: Language for interactions (default: "en-US")
- `model`: Ollama model to use (default: "llama3.2")

Optional fields:

- `base_guidelines`: Override the default conversation style and behavior for all agents
- `interactive_guidelines`: Guidelines for conversations with human participation (default mode)
- `autonomous_guidelines`: Guidelines for autonomous agent conversations (--auto mode)

Example with optional fields:

```json
{
  "current_agent": "chatty",
  "language_code": "en-US",
  "model": "llama3.2",
  "base_guidelines": "Be professional and formal in your responses. Focus on accuracy and clarity.",
  "interactive_guidelines": "Always speak in first person and Acknowledge others before adding your view",
  "autonomous_guidelines": "Always speak in first person and Drive the conversation with questions"
}
```

The guidelines settings control how agents behave in different conversation modes:

- `interactive_guidelines`: Used in regular conversations where humans participate. This is the default mode when using `--with` without the `--auto` flag. These guidelines encourage agents to interact with both human users and other agents.

- `autonomous_guidelines`: Used only when the `--auto` flag is enabled. These guidelines are specifically designed for agent-to-agent conversations without human participation, encouraging more autonomous discussion between the agents.

You can modify these guidelines to create different conversation dynamics or enforce specific interaction patterns. For example, you might want to make autonomous conversations more focused on debate, or interactive conversations more educational.

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

GNU General Public License v3.0 - See [LICENSE](LICENSE) file for details.
Copyright (c) 2025 Chatty AI
