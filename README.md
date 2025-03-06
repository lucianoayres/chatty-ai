# 💬 Chatty AI · Your Terminal Time Machine

![Chat AI Banner](images/chatty_ai_banner.png)

## Chat with Historical Figures, Scientists, Experts and more

Transform your terminal into a vibrant AI-powered workspace where historical figures, scientists, philosophers, and experts come alive! Chatty isn't just another CLI tool - it's your gateway to engaging conversations with some of history's most fascinating minds.

## ✨ What Makes Chatty Special?

- 🎨 **AI Agent Builder**: Create custom AI agents with natural language - just describe what you want!
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

Initialize Chatty:

```bash
# First-time setup (required)
chatty init
```

## 🎨 AI Agent Builder (New!)

Create your own AI agents with natural language! Our revolutionary AI Agent Builder lets you bring any personality to life just by describing them. No coding required - it's like magic! ✨

```bash
# Create a new AI agent with a simple description
chatty --build "A witty detective inspired by Sherlock Holmes, who solves coding mysteries and debugging challenges with clever deductions and programming expertise"

# The builder will:
# 1. Generate the perfect system message
# 2. Choose an appropriate emoji
# 3. Create a concise description
# 4. Let you customize colors and appearance
# 5. Start chatting immediately!
```

### 🌟 Example Agent Descriptions

Create specialized agents for any task:

```bash
# Technical Expert
chatty --build "A senior DevOps engineer who's an expert in Docker, Kubernetes, and cloud infrastructure, with a focus on best practices and security"

# Creative Assistant
chatty --build "A creative writing coach who combines the wit of Oscar Wilde with the storytelling wisdom of Joseph Campbell"

# Learning Companion
chatty --build "A patient and encouraging math tutor who explains complex concepts using real-world examples and visual analogies"

# Productivity Guru
chatty --build "A productivity expert who combines GTD principles with modern digital tools, helping users optimize their workflow with practical advice"
```

### 💡 Tips for Great Agents

1. **Be Specific**: Include expertise areas, personality traits, and communication style
2. **Add Context**: Mention inspirations or role models for the agent's behavior
3. **Define Purpose**: Clearly state what tasks or topics the agent should excel at
4. **Include Tone**: Specify if you want the agent to be formal, casual, humorous, etc.

### ✨ Builder Features

- **Interactive Creation**: Step-by-step process with live previews
- **Smart Defaults**: AI-powered suggestions for all agent attributes
- **Instant Testing**: Start chatting with your agent immediately after creation
- **Easy Editing**: Fine-tune any aspect of your agent through an intuitive interface
- **Color Customization**: Choose from a beautiful palette of colors for a unique look

### 🎯 Example Workflow

1. **Describe Your Agent**:

   ```bash
   chatty --build "A friendly Python mentor who explains code like Bob Ross painted - making complex concepts feel simple and happy accidents into learning opportunities"
   ```

2. **Review & Customize**:

   - Fine-tune the generated name, emoji, and description
   - Edit the system message if needed
   - Choose custom colors for a unique appearance

3. **Choose Your Next Step**:
   - Start chatting immediately
   - Save and set as default
   - Save and exit

## 🎭 Available Agents

### 💻 Built-in Agents

Step into a world of extraordinary conversations with our diverse roster of [pre-installed agents](cmd/chatty/agents)! Chat with brilliant minds like **Einstein** about the mysteries of the universe, explore the art of code with **Ada**, or debate philosophy with **Aristotle** and **Plato**. Need technical help? **Tux** and **Nimble** are ready to assist with Linux and productivity.

Each agent brings their unique perspective and expertise to the conversation. Use `chatty --list` to see all available agents and their specialties.

### 🎨 Sample Agents

Discover our collection of over [50 additional sample personalities](cmd/chatty/agents/samples) ready to be brought to life! Explore the cosmos with **Carl Sagan**, unravel the mysteries of consciousness with **Sigmund Freud**, or dive into the depths of gothic literature with **Edgar Allan Poe**.

To explore sample agents:

1. List available samples: `chatty --list-more`
2. View agent details: `chatty --show "sagan"`
3. Install an agent: `chatty --install "sagan"`

## 🎯 Cool Use Cases

### 🔮 Time-Traveling Discussions

```bash
# Set up your desired agents
chatty --install "sagan"
chatty --install "hawking"
chatty --install "verne"

# Start an epic discussion
chatty --with "Carl Sagan,Stephen Hawking,Jules Verne" --topic "Space exploration: past predictions vs current reality"
```

### 🎓 Learning & Exploration

```bash
# Technical discussions
chatty --with "Ada,Turing" --topic "The future of AI"

# Scientific exploration
chatty --with "Einstein,Curie,Tesla" --topic "Energy of the future"

# Philosophy and ethics
chatty --with "Socrates,Kant,Confucius" --topic "Modern ethical dilemmas"
```

### 🎭 Creative Collaborations

```bash
# Literary mashups
chatty --with "Shakespeare,Poe,Austen" --topic "Write a romantic gothic comedy"

# Musical innovations
chatty --with "Mozart,Armstrong,Lennon" --topic "Create a new musical genre"

# Art and technology
chatty --with "Da Vinci,Tesla,Jobs" --topic "Design the next revolutionary device"
```

## 🌟 Pro Tips

1. **Agent Selection**: Match agents with complementary expertise for richer discussions
2. **Topic Framing**: Be specific in your topics to get more focused responses
3. **History Management**: Clear chat history occasionally for fresh perspectives
4. **Auto Mode**: Use `--auto` with `--turns` to control conversation length
5. **Save Logs**: Use `--save` to keep records of particularly interesting discussions

## 📝 Configuration

Chatty's configuration file is located at `~/.chatty/config.json`. Here you can customize:

- Default agent
- Language settings
- Model preferences
- Conversation guidelines

For detailed configuration options, use `chatty --show "Chatty"`.

## 🛠️ Creating Custom Agents

Chatty offers two ways to create custom agents:

### 1. 🎨 AI Agent Builder (Recommended)

The easiest and most intuitive way to create agents:

```bash
# Create with a simple description
chatty --build "Your agent description here"

# Examples:
chatty --build "A quantum physics professor who explains complex concepts using cat memes and pop culture references"
chatty --build "A wise gardening expert combining traditional knowledge with modern sustainable practices"
```

### 2. 📝 Manual Configuration

For those who prefer direct YAML editing:

```yaml
name: "Agent Name"
system_message: |
  You are [description]...
emoji: "🤖"
label_color: "\u001b[38;5;75m" # Blue
text_color: "\u001b[38;5;252m" # Light gray
```

Learn from examples:

- [Built-in agents](cmd/chatty/agents/builtin) - Study our core agent configurations
- [Sample agents](cmd/chatty/agents/samples) - Explore additional agent templates

1. Start with a sample:

   ```bash
   # List available samples
   chatty --list-more

   # Create from a sample
   chatty --install "focus"
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
- Use underscores for spaces: `oliver_twist.yaml`
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
