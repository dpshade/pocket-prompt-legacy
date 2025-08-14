# Pocket Prompt

**Your personal LLM command center.** A unified interface and storage pool for all your AI interactions - prompts, agents, commands, templates, and context - owned and controlled by you.

Think of it as your **AI toolkit in a single binary**: a fast, keyboard-driven interface for managing, organizing, and deploying your personal collection of LLM resources across all your devices and workflows.

## Why Pocket Prompt?

**Stop losing your best prompts.** Stop rewriting the same instructions. Stop switching between tools to manage your AI workflows.

Pocket Prompt is your **unified command center** for all LLM interactions:

### 🎯 **Unified Storage Pool**
- **Prompts**: Your tried-and-true AI instructions
- **Agents**: Reusable AI personas and roles  
- **Commands**: Automation scripts and workflows
- **Templates**: Consistent structures across projects
- **Context**: Project-specific knowledge and constraints

### 🚀 **Multi-Interface Access**
- **TUI**: Fast, keyboard-driven terminal interface
- **CLI**: Headless automation for scripts and CI/CD
- **HTTP API**: iOS Shortcuts, web apps, and integrations
- **Git Sync**: Multi-device access with version control

### 💾 **Own Your Data**
- **Plain Text**: Markdown files with YAML frontmatter
- **Local First**: Everything stored on your machine
- **Git-Friendly**: Perfect version control integration  
- **Portable**: Single binary, no dependencies
- **Future-Proof**: Standard formats that work everywhere

### 🔍 **Intelligent Organization**
- **Fuzzy Search**: Find anything instantly
- **Boolean Search**: Complex tag-based filtering (`(ai AND analysis) OR writing`)
- **Variable System**: Parameterized, reusable prompts
- **Template Engine**: Consistent prompt structures
- **Auto-sync**: Background Git synchronization

## Installation

### One-Line Installation (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/dpshade/pocket-prompt/master/install.sh | bash
```

### Alternative Installation Methods

#### Using Go (if you have Go installed)
```bash
go install github.com/dpshade/pocket-prompt@latest
```

#### Manual Download (Linux/macOS)
```bash
# Download and install latest release
curl -s https://api.github.com/repos/dpshade/pocket-prompt/releases/latest | \
  grep "browser_download_url.*$(uname -s)_$(uname -m)" | \
  cut -d '"' -f 4 | \
  xargs curl -L -o pocket-prompt && \
  chmod +x pocket-prompt && \
  sudo mv pocket-prompt /usr/local/bin/
```

#### Using Homebrew (coming soon)
```bash
# brew install dpshade/tap/pocket-prompt
```

#### Build from Source
```bash
git clone https://github.com/dpshade/pocket-prompt.git
cd pocket-prompt
go build -o pocket-prompt
```

## Quick Start

1. Initialize your prompt library:
```bash
pocket-prompt --init
```

2. Launch the TUI:
```bash
pocket-prompt
```

Or use CLI mode for automation:
```bash
pocket-prompt list              # List all prompts
pocket-prompt search "AI"       # Search for prompts
pocket-prompt show prompt-id    # Display a specific prompt
pocket-prompt copy prompt-id    # Copy to clipboard
pocket-prompt --url-server      # Start HTTP API for iOS Shortcuts
```

3. Navigate with keyboard shortcuts:

   **Library View:**
   - `↑/k` / `↓/j` - Navigate up/down
   - `Enter` - Open prompt detail page
   - `e` - Edit selected prompt
   - `n` - Create new prompt
   - `t` - Manage templates
   - `/` - Search prompts (fuzzy search)
   - `Ctrl+B` - Boolean tag search
   - `q` - Quit

   **Prompt Detail View:**
   - `↑/k` / `↓/j` - Scroll content
   - `c` - Copy rendered prompt as plain text
   - `y` - Copy rendered prompt as JSON messages
   - `e` - Edit this prompt
   - `←/esc/b` - Back to library
   - `?` - Show help

   **Template Management:**
   - `1-9` - View template details
   - `n` - Create new template
   - `←/esc/b` - Back to library

   **Edit Forms:**
   - `Tab/↓` - Next field
   - `Shift+Tab/↑` - Previous field
   - `Ctrl+S` - Save changes
   - `←/esc/b` - Cancel (back to library)

## Git Sync Quick Start

**Sync your prompts across devices** with automatic Git backup and multi-device access.

### 1. Create a Git Repository

**Option A: GitHub (Recommended)**
```bash
# Create a private repository on GitHub first, then:
pocket-prompt git setup https://github.com/yourusername/my-prompts.git
```

**Option B: Using GitHub CLI**
```bash
# Let the setup script create the repo for you
./setup-github-sync.sh
```

**Option C: SSH (More Secure)**
```bash
pocket-prompt git setup git@github.com:yourusername/my-prompts.git
```

### 2. Authentication Setup

**For HTTPS (GitHub):**
- Create a Personal Access Token at https://github.com/settings/tokens
- Select "repo" scope for private repositories
- Use token as password when prompted

**For SSH (Recommended):**
```bash
# Generate SSH key if you don't have one
ssh-keygen -t ed25519 -C "your-email@example.com"

# Add to GitHub: https://github.com/settings/ssh/new
cat ~/.ssh/id_ed25519.pub
```

### 3. Verify Setup

```bash
pocket-prompt git status        # Check sync status
```

### 4. Automatic Sync

That's it! Your prompts now automatically:
- ✅ **Backup every change** to your repository
- ✅ **Pull updates** from other devices every 5 minutes  
- ✅ **Resolve conflicts** automatically when possible
- ✅ **Sync in background** without interrupting your workflow

### Multi-Device Usage

**On a new device:**
```bash
# Clone your existing prompt library
git clone https://github.com/yourusername/my-prompts.git ~/.pocket-prompt
cd ~/.pocket-prompt
pocket-prompt git enable    # Enable sync on this device
```

**Quick Commands:**
```bash
pocket-prompt git sync      # Force sync now
pocket-prompt git pull      # Pull latest changes
pocket-prompt git status    # Check sync status
```

## Prompt Structure

Prompts are Markdown files with YAML frontmatter. All headers are fully editable:

```markdown
---
id: my-prompt
version: 1.0.0
title: My Awesome Prompt
description: A helpful prompt for X
tags:
  - category1
  - category2
variables:
  - name: topic
    type: string
    required: true
    description: The topic to analyze
    default: "AI ethics"
  - name: depth
    type: number
    required: false
    default: 3
template: analysis-template
---

# Prompt Content

Analyze the following topic: {{topic}} with depth level {{depth}}.

Please provide:
1. Key insights
2. Recommendations
3. Next steps
```

### Editable Headers for Prompts:
- **ID**: Unique identifier for the prompt
- **Version**: Semantic version (e.g., "1.0.0")
- **Title**: Display name for the prompt
- **Description**: Brief summary of what the prompt does
- **Tags**: Comma-separated categories for organization
- **Variables**: Template variables with types, defaults, and requirements
- **Template**: Reference to a template ID (optional)
- **Content**: The actual prompt text with variable placeholders

## Directory Structure

```
~/.pocket-prompt/
├── prompts/       # Your prompt files
├── templates/     # Reusable templates
├── packs/         # Curated collections
└── .pocket-prompt/
    ├── index.json # Search index
    └── cache/     # Rendered prompts cache
```

## Creating New Prompts

### From Scratch
1. Press `n` in the library view
2. Navigate to "Create from scratch" using `↑/↓` or `k/j`
3. Press `Enter` to select
4. Fill in the form fields:
   - **ID**: Unique identifier (e.g., "my-prompt")
   - **Version**: Semantic version (defaults to "1.0.0")
   - **Title**: Display name
   - **Description**: Brief summary
   - **Tags**: Comma-separated categories
   - **Variables**: Template variables (format: `name:type:required:default`)
   - **Template Ref**: Reference to template ID (optional)
   - **Content**: The actual prompt text
5. Save with `Ctrl+S`

### From Templates
1. Press `n` in the library view
2. Navigate to "Use a template" using `↑/↓` or `k/j`
3. Press `Enter` to select
4. Choose from available templates using arrow keys:
   - Analysis Template
   - Creative Writing Template
   - Technical Documentation Template
5. Press `Enter` to select template
6. Fill in template-specific fields
7. Customize the generated prompt

## Editing Prompts and Templates

### Edit Existing Prompts
1. In library view, select a prompt and press `e`
2. Or open prompt detail view and press `e`
3. Modify any field (ID is read-only for existing prompts)
4. Press `Ctrl+S` to save changes
5. Press `←/esc/b` to cancel without saving

### Template Management
1. Press `t` in library view to access template management
2. Select a template by number to view details
3. Press `e` in template detail view to edit
4. Press `n` to create a new template

### Template Creation
Templates use special slot syntax for customizable sections:
- `{{slot_name}}` - Basic slot substitution
- `{{#if condition}}...{{/if}}` - Conditional sections
- YAML frontmatter defines slot properties and constraints

## Variables

Define variables in your prompts to make them reusable:

```yaml
variables:
  - name: language
    type: string
    required: true
    default: "Python"
  - name: focus_areas
    type: string
    required: false
    default: "performance, security"
```

Variable types:
- `string` - Text values
- `number` - Numeric values
- `boolean` - True/false values
- `list` - Arrays of values

## Boolean Search

Boolean search provides advanced tag-based filtering using logical operators. Access it by pressing `Ctrl+B` in the library view.

### Syntax

Use logical operators to create complex search expressions:

- **AND**: Find prompts that have all specified tags
  ```
  ai AND analysis
  ```

- **OR**: Find prompts that have any of the specified tags
  ```
  writing OR creative
  ```

- **NOT**: Exclude prompts with specific tags
  ```
  NOT deprecated
  ```

- **Parentheses**: Group expressions for complex logic
  ```
  (ai AND analysis) OR writing
  ```

### Examples

```bash
# Find prompts tagged with both "ai" and "analysis"
ai AND analysis

# Find prompts with either "writing" or "creative" tags
writing OR creative

# Find AI prompts but exclude deprecated ones
ai AND NOT deprecated

# Complex query: AI analysis or writing prompts, but not templates
(ai AND analysis) OR writing AND NOT template

# Find prompts with specific combinations
(python OR javascript) AND tutorial AND NOT beginner
```

### Features

- **Live Search**: Results update as you type
- **Tag Autocomplete**: Shows available tags for reference
- **Save Searches**: Save complex expressions with `Ctrl+S`
- **Edit Saved Searches**: Modify and reuse saved boolean expressions
- **Keyboard Navigation**: Use `Tab` to switch between search input and results

### Keyboard Shortcuts in Boolean Search Modal

- `Tab` - Toggle focus between search input and results
- `↑/↓` or `k/j` - Navigate through search results
- `Enter` - Apply search and return to list (when in search input) or select result (when in results)
- `Ctrl+S` - Save current search expression
- `Ctrl+H` - Toggle help text
- `Esc` - Close boolean search modal

## Templates

Templates provide consistent structure across prompts. All headers are fully editable:

```yaml
---
id: analysis-template
version: 1.0.0
name: Analysis Template
description: Template for analytical prompts
slots:
  - name: identity
    description: The role to play
    required: true
    default: "expert analyst"
  - name: output_format
    description: Format for the response
    required: false
    default: "bullet points"
---

You are an {{identity}}.

{{content}}

Format your response as {{output_format}}.
```

### Editable Headers for Templates:
- **ID**: Unique identifier for the template
- **Version**: Semantic version (e.g., "1.0.0")
- **Name**: Display name for the template
- **Description**: Brief summary of what the template provides
- **Slots**: Template slots with descriptions, requirements, and defaults
- **Content**: The template text with slot placeholders

### Variable and Slot Formats:
When editing in forms, use these formats:

**Variables**: `name:type:required:default, name2:type:required:default`
- Example: `topic:string:true:AI ethics, depth:number:false:3`

**Slots**: `name:description:required:default, name2:description:required:default`
- Example: `identity:The role to play:true:expert analyst, format:Output format:false:bullet points`

## Clipboard Support

Pocket Prompt supports copying to clipboard on:
- **macOS**: Uses `pbcopy`
- **Linux**: Uses `xclip`, `xsel`, or `wl-copy` (Wayland)
- **Windows**: Uses `clip`

Copy formats:
- **Plain text** (`c`): Raw rendered prompt text
- **JSON messages** (`y`): Formatted for LLM APIs like OpenAI

## CLI Mode

Pocket Prompt includes a comprehensive CLI mode for automation:

```bash
# Basic operations
pocket-prompt list                          # List all prompts
pocket-prompt list --format json            # JSON output
pocket-prompt search "keyword"              # Search prompts (fuzzy)
pocket-prompt search --boolean "ai AND analysis"  # Boolean tag search
pocket-prompt show prompt-id                # Display prompt
pocket-prompt copy prompt-id                # Copy to clipboard
pocket-prompt render prompt-id --var key=value  # Render with variables

# Create and edit
pocket-prompt create new-prompt-id          # Create new prompt
pocket-prompt edit prompt-id                # Edit existing prompt
pocket-prompt delete prompt-id              # Delete prompt

# Template management
pocket-prompt templates list                # List templates
pocket-prompt templates show template-id    # Show template details

# Git synchronization
pocket-prompt git status                    # Check sync status
pocket-prompt git sync                      # Manual sync
pocket-prompt git pull                      # Pull remote changes
```

Output formats: `--format table|json|ids` for scripting and integration.

## Git Synchronization

**One-command setup** - just provide your repository URL:

```bash
pocket-prompt git setup https://github.com/username/my-prompts.git
```

or with SSH:

```bash
pocket-prompt git setup git@github.com:username/my-prompts.git
```

That's it! The app automatically:

✅ **Initializes the Git repository**  
✅ **Creates initial commit and README**  
✅ **Configures the remote repository**  
✅ **Handles authentication guidance**  
✅ **Starts background synchronization**  

### Advanced Features

Once set up, enjoy automatic Git sync with:

- **Automatic Conflict Resolution**: Smart merging strategies for concurrent edits
- **Background Sync**: Continuous monitoring and pulling of remote changes every 5 minutes
- **Resilient Push**: Automatic retry with pull-and-merge on push failures
- **Recovery Options**: Force sync to recover from complex merge scenarios

### Authentication Support

The setup command provides helpful guidance for:
- **GitHub Personal Access Tokens** for HTTPS
- **SSH key setup** for secure authentication (recommended)
- **Multiple authentication methods** with clear error messages

### Manual Commands

```bash
pocket-prompt git status      # Check sync status
pocket-prompt git sync        # Manual sync
pocket-prompt git pull        # Pull remote changes
```

## HTTP API Server

Pocket Prompt includes a built-in HTTP API server perfect for **iOS Shortcuts integration** and automation workflows. The server provides URL-based access to all prompt operations with clipboard-based responses for seamless mobile integration.

### Starting the Server

```bash
pocket-prompt --url-server                    # Start on default port 8080
pocket-prompt --url-server --port 9000        # Start on custom port
```

The server provides helpful startup information:
```
URL server starting on http://localhost:8080
iOS Shortcuts can now call URLs like:
  http://localhost:8080/pocket-prompt/render/my-prompt-id
  http://localhost:8080/pocket-prompt/search?q=AI
  http://localhost:8080/pocket-prompt/boolean?expr=ai+AND+analysis
```

### API Endpoints

All endpoints return content directly in the response body with appropriate content types (text/plain or application/json).

#### Prompt Operations
```bash
# List all prompts
GET /pocket-prompt/list?format=json&limit=10&tag=ai

# Search prompts (fuzzy search)
GET /pocket-prompt/search?q=machine+learning&format=table&limit=5

# Get specific prompt
GET /pocket-prompt/get/my-prompt-id?format=json

# Render prompt with variables
GET /pocket-prompt/render/my-prompt-id?var1=value&var2=test&format=text
```

#### Search Operations
```bash
# Boolean expression search
GET /pocket-prompt/boolean?expr=ai+AND+analysis
GET /pocket-prompt/boolean?expr=(python+OR+javascript)+AND+tutorial

# Execute saved search
GET /pocket-prompt/saved-search/my-saved-search

# List all saved searches
GET /pocket-prompt/saved-searches/list
```

#### Tag Operations
```bash
# List all tags
GET /pocket-prompt/tags

# Get prompts with specific tag
GET /pocket-prompt/tag/python?format=ids
```

#### Template Operations
```bash
# List all templates
GET /pocket-prompt/templates?format=json

# Get specific template
GET /pocket-prompt/template/my-template-id
```

### Response Formats

Control output format with the `format` parameter:

- `format=text` (default) - Human-readable text
- `format=json` - JSON structure
- `format=ids` - Just prompt/template IDs
- `format=table` - Formatted table view

### iOS Shortcuts Integration

The API returns content directly in HTTP responses, perfect for iOS Shortcuts:

#### Basic Prompt Access
1. **Get Contents of URL**: `http://localhost:8080/pocket-prompt/render/my-prompt`
2. **Use response content** directly in ChatGPT, Claude, or other apps

#### Search and Select Workflow
1. **Get Contents of URL**: `http://localhost:8080/pocket-prompt/search?q=AI&format=ids`
2. **Split Text** by new lines to get prompt IDs
3. **Choose from Menu** - Select a prompt ID
4. **Get Contents of URL**: `http://localhost:8080/pocket-prompt/render/[chosen-id]`
5. **Use response content** as your prompt

#### Advanced Boolean Search
1. **Text Input**: Enter boolean expression like `(ai AND analysis) OR writing`
2. **Get Contents of URL**: `http://localhost:8080/pocket-prompt/boolean?expr=[encoded-expression]`
3. **Process response content** - matching prompts returned directly

#### Variable-Based Rendering
1. **Ask for Input**: "Topic"
2. **Ask for Input**: "Detail Level"  
3. **Get Contents of URL**: `http://localhost:8080/pocket-prompt/render/analysis?topic=[input1]&depth=[input2]`
4. **Use response content** - customized prompt ready for AI

### Example iOS Shortcuts

**Quick AI Prompt**: 
- Choose from predefined prompt list → Render → Copy to AI app

**Smart Search**:
- Voice input "Search for coding prompts" → API search → Select result → Render

**Dynamic Prompt Builder**:
- Input variables via Shortcuts → Render with variables → Ready for AI

### Security & Local Access

- **Localhost only** - No external network access required
- **No authentication** - Designed for local use
- **Direct HTTP responses** - Standard REST API pattern
- **Works offline** - No internet dependency
- **Automatic git sync** - Keeps prompts updated every 5 minutes (configurable)

### API Documentation

```bash
GET /help or GET /api
# Returns comprehensive API documentation in markdown format

GET /help?format=json  
# Returns structured JSON documentation with all endpoints
```

### Health Check

```bash
GET /health
# Returns: {"status": "ok", "service": "pocket-prompt-url-server"}
```

## Roadmap

- [x] CLI commands (render, copy, lint)
- [x] Clipboard integration
- [x] Export formats (JSON, plain text)
- [x] Advanced Git synchronization
- [x] Comprehensive UI design system
- [x] HTTP API server for iOS Shortcuts integration
- [ ] Linter for prompt validation
- [ ] Pack management
- [ ] DNS TXT publishing
- [ ] Signature verification

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

MIT License - See [LICENSE](LICENSE) for details.