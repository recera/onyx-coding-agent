# Onyx AI Coding TUI

A sophisticated Terminal User Interface (TUI) for an AI-powered coding assistant that can read, write, edit, search, and run commands in your projects. Built with Go (Bubbletea + Lipgloss) for the frontend and TypeScript (AI SDK) for the AI agent backend.

## Features

- 🎨 Beautiful terminal interface with Bubbletea and Lipgloss
- 🤖 AI-powered coding assistant using OpenAI's GPT models
- 🧠 **Code Intelligence**: Builds a knowledge graph of your code for deep analysis.
- 📁 **File Operations**: List, read, write, and edit files.
- 🔍 **Search**: Find text or patterns across your codebase.
- 🏃 **Run Commands**: Execute shell commands directly.
- 💬 Real-time chat interface with streaming responses.
- 🔐 Secure API key input.
- 📊 Tool call visualization.
- 🌍 **Works from any directory** - operates on your current working directory.
- 🚀 Single command launch from anywhere.

## Architecture

```
┌─────────────────────────────────┐
│     Go TUI (Bubbletea)          │
│   - User Interface              │
│   - Input Handling              │
│   - Display Management          │
└────────────┬────────────────────┘
             │ stdio (JSON)
             ↓
┌─────────────────────────────────┐
│  TypeScript Agent (AI SDK)      │
│   - AI Logic                    │
│   - Tool Execution              │
│   - OpenAI Integration          │
└─────────────────────────────────┘
```

## Code Intelligence Engine

Onyx TUI includes a powerful `graph_service` that acts as its code intelligence engine. When you start the TUI, this service analyzes your project's source code (supporting Go, Python, and TypeScript) and builds a detailed knowledge graph using a KuzuDB graph database.

This graph captures:
- **Code Entities**: Functions, classes, methods, structs, interfaces, etc.
- **Relationships**: Calls, inheritance, implementations, imports, and more.

This deep understanding of your code allows the AI agent to answer complex questions, analyze dependencies, and understand the architecture of your project. The `run_cypher` tool provides direct access to query this knowledge graph.

## Prerequisites

- **Node.js** (v18 or higher)
- **Go** (v1.20 or higher)
- **OpenAI API Key**

## Installation

### Quick Install (Recommended)

1. **Clone or download this project**

2. **Run the installation script:**
   ```bash
   cd onyx-tui
   ./install.sh
   ```

3. **Add to PATH (if not already):**
   ```bash
   export PATH="$HOME/.local/bin:$PATH"
   ```
   Add this line to your `~/.bashrc` or `~/.zshrc` to make it permanent.

4. **Use from any directory:**
   ```bash
   cd /your/project/directory
   onyx
   ```

### Local Development

If you want to run the TUI for development without installing it globally, you can use the `launch.sh` script. This script compiles both the Go TUI and the TypeScript agent and then runs the application.

```bash
cd onyx-tui
./launch.sh
```
This is the recommended way to run the TUI while making changes to the code.

## Usage

### After Installation

1. **Navigate to your project:**
   ```bash
   cd /path/to/your/project
   ```

2. **Start Onyx:**
   ```bash
   onyx
   ```

3. **Enter your OpenAI API key** when prompted (hidden input)

4. **Start coding with AI assistance!**

### Interacting with the AI

- **Send Messages**: Type and press `Ctrl+S` (or `Enter` for single line)
- **Exit**: Press `Ctrl+C` or `Esc`

### Available Tools

The AI agent has access to powerful file system and command execution tools:

#### 📁 **list_files**
List files and directories in your project.
- Example: "Show me all TypeScript files in the src directory"

#### 📖 **read_file**
Read the contents of any file.
- Example: "What does the package.json file contain?"

#### ✏️ **write_file**
Create new files with specified content.
- Example: "Create a new README.md file with installation instructions"

#### 🔧 **edit_file**
Edit existing files by replacing text.
- Example: "Fix the typo in config.js where it says 'teh' instead of 'the'"

#### 🔍 **search_files**
Search for text or patterns across your codebase.
- Example: "Find all TODO comments in the project"

#### 🏃 **run_command**
Execute shell commands in your project directory.
- Example: "Run npm test and show me the results"

#### ⚡ **run_cypher**
Query the codebase's knowledge graph using Cypher.
- Example: "Find all functions that call the 'calculateSum' function"

#### 🌐 **web_search**
Perform a web search to find information, documentation, or solutions.
- Example: "Search for the latest version of the 'react' npm package"

#### 🔗 **url_extract**
Extract content from a URL.
- Example: "Summarize the content of https://example.com/some-article"

### Example Prompts

- "What files are in this project?"
- "Read the main.go file and explain what it does"
- "Create a simple Python script that prints hello world"
- "Search for all functions that start with 'handle'"
- "Run 'npm install' and then 'npm test'"
- "Help me fix the syntax error in app.js"
- "Add a new function to utils.js that validates email addresses"

## Project Structure

```
onyx-tui/
├── agent/                  # TypeScript AI agent
│   ├── src/
│   │   └── agent.ts       # Main agent logic with AI SDK
│   ├── package.json       # Node dependencies
│   └── tsconfig.json      # TypeScript configuration
├── main.go                # Go TUI application
├── go.mod                 # Go module definition
├── launch.sh              # Launch script
└── README.md              # This file
```

## Communication Protocol

The TUI and agent communicate via JSON messages over stdio:

### Message Types
- `init`: Initialize agent with API key
- `chat`: Send user message
- `response`: Agent response
- `tool_call`: Tool invocation notification
- `stream_chunk`: Streaming response chunk
- `error`: Error message

### Example Flow
```json
→ {"type":"init","data":{"apiKey":"sk-..."}}
← {"type":"response","data":{"status":"initialized"}}
→ {"type":"chat","data":{"message":"Hello"}}
← {"type":"tool_call","data":{"toolName":"code_analyzer"}}
← {"type":"response","data":{"content":"..."}}
```

## Troubleshooting

### Issue: "Node.js is not installed"
**Solution**: Install Node.js from https://nodejs.org/

### Issue: "Go is not installed"
**Solution**: Install Go from https://golang.org/

### Issue: "API key error"
**Solution**: Ensure your OpenAI API key is valid and has sufficient credits

### Issue: TUI doesn't start
**Solution**: Check the log files:
- `onyx-tui.log`: TUI logs
- `agent-error.log`: Agent stderr output

### Issue: Agent not responding
**Solution**: 
1. Check if Node.js dependencies are installed: `cd agent && npm install`
2. Verify the agent can run standalone: `cd agent && npm start`
3. Check `agent-error.log` for errors

## Development

The `launch.sh` script handles the development workflow. It automatically rebuilds both the Go TUI and the TypeScript agent whenever you run it.

### Modifying the Agent (TypeScript)
1.  Edit files in the `agent/src/` directory.
2.  Run `./launch.sh` to see your changes. The script will automatically recompile the TypeScript code.

### Modifying the TUI (Go)
1.  Edit `main.go` or other Go source files.
2.  Run `./launch.sh`. The script will automatically rebuild the Go binary.

### Adding Real Tools
Replace the placeholder tool implementations in `agent/src/agent.ts` with actual logic:
```typescript
execute: async ({ path, analysisType }) => {
  // Add your real implementation here
  const result = await analyzeCode(path, analysisType);
  return result;
}
```

## Environment Variables

- `OPENAI_API_KEY`: Set this to avoid entering the API key each time
  ```bash
  export OPENAI_API_KEY="sk-..."
  ./launch.sh
  ```

## License

MIT

## Contributing

Feel free to submit issues and enhancement requests!