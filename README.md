# 💬 Chatty AI · Your Terminal Time Machine

![Chat AI Banner](images/chatty_ai_banner.png)

## Generate & Chat with Your Dream Team of Historical Figures, Scientists, Experts & More!

Create and converse with any historical figure, scientist, expert, or personality you can imagine. From ancient philosophers to modern innovators, literary giants to tech visionaries - if you can describe them, Chatty can bring them to life in your terminal!

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

## 🎨 AI Agent Builder

Create your own AI agents with natural language! Our revolutionary AI Agent Builder uses AI to automatically generate complete agent personalities from simple descriptions or even just names. Just tell it who or what you want, and watch the magic happen! ✨

The AI will generate everything:

- Perfect system message capturing the agent's essence
- Fitting emoji that represents their character
- Concise yet informative description

You'll then get to review and fine-tune everything through an intuitive interface!

### 🌟 Creating Agents from Names

These are just examples - let your imagination run wild! Any name you can think of, from any field or era, can become a rich AI personality:

```bash
# Scientific Pioneers
chatty --build "Carl Jung"            # Pioneering psychoanalyst and founder of analytical psychology
chatty --build "Friedrich Nietzsche"  # Influential philosopher known for existentialism and nihilism
chatty --build "Grace Hopper"         # Computer science pioneer and inventor of COBOL

# Fictional Characters
chatty --build "Tony Stark"           # Your favorite genius billionaire playboy philanthropist
chatty --build "John Wick"            # The man, the myth, the legend with a dog
chatty --build "Hermione Granger"     # The brightest witch of her age
```

### 🎯 Creating Specialized Agents

Or describe exactly what you want:

```bash
# Modern Mentors
chatty --build "A blockchain developer who explains Web3 concepts through analogies from popular video games and memes"

# Creative Innovators
chatty --build "A film director who combines Christopher Nolan's plot complexity, Wes Anderson's visual style, and Tarantino's dialogue writing"

# Tech Guides
chatty --build "A cybersecurity expert who explains hacking concepts through heist movie scenarios and spy thriller references"

# Unique Specialists
chatty --build "A futurist sociologist who analyzes current trends through the lens of Black Mirror episodes and modern sci-fi stories"
```

### 💡 Tips for Great Agents

1. **Using Names**:

   - Just provide the name of any historical figure, expert, or personality
   - The AI will research and capture their essence
   - You can review and adjust the generated profile

2. **Using Descriptions**:
   - Be specific about expertise areas and personality traits
   - Mention inspirations or role models for behavior
   - Include desired communication style (formal, casual, humorous)
   - Specify unique perspectives or approaches

### ✨ Builder Features

- **AI-Powered Generation**: Complete agent profiles created automatically
- **Interactive Review**: Step-by-step process to review and customize
- **Smart Defaults**: AI-chosen emojis and colors that match the personality
- **Instant Testing**: Start chatting with your agent immediately
- **Easy Editing**: Fine-tune any aspect through an intuitive interface

### 🎯 Example Workflow

1. **Provide Input**:

   ```bash
   # Using a name:
   chatty --build "Michelangelo"

   # Or a description:
   chatty --build "A culinary historian who explores ancient recipes and cooking techniques, bringing the flavors of history to life with modern adaptations"
   ```

2. **Review AI-Generated Profile**:

   - The AI creates a complete agent profile
   - Review name, emoji, and description
   - Check the system message that defines their personality
   - Adjust any aspects if desired

3. **Customize Appearance**:

   - Choose from beautiful color combinations
   - Preview how they'll look in chat
   - Fine-tune their visual identity

4. **Choose Your Next Step**:
   - Start chatting immediately with your new agent
   - Save and set as your default agent
   - Save and exit to use later

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
