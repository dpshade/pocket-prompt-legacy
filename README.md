# Pocket Prompt

A unified place to keep LLM context - a single-binary, portable prompt interface with a first-class Terminal User Interface (TUI) for browsing, previewing, rendering, and managing prompts as versioned artifacts you own.

## Features

- **File-Based Storage**: All prompts are stored as plain Markdown files with YAML frontmatter
- **TUI Interface**: Fast, keyboard-driven interface with polished design system built with Charmbracelet
- **CLI Mode**: Comprehensive headless mode for automation, scripting, and CI/CD pipelines
- **Variable Substitution**: Define variables in prompts and fill them at runtime
- **Template System**: Reusable templates for consistent prompt structures
- **Fuzzy Search**: Quickly find prompts with fuzzy matching
- **Advanced Git Sync**: Automatic conflict resolution and background synchronization for multi-machine workflows
- **Portable**: Single binary with no external dependencies
- **Git-Friendly**: Plain text files that work perfectly with version control

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
```

3. Navigate with keyboard shortcuts:

   **Library View:**
   - `↑/k` / `↓/j` - Navigate up/down
   - `Enter` - Open prompt detail page
   - `e` - Edit selected prompt
   - `n` - Create new prompt
   - `t` - Manage templates
   - `/` - Search prompts
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
pocket-prompt search "keyword"              # Search prompts
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

## Roadmap

- [x] CLI commands (render, copy, lint)
- [x] Clipboard integration
- [x] Export formats (JSON, plain text)
- [x] Advanced Git synchronization
- [x] Comprehensive UI design system
- [ ] Linter for prompt validation
- [ ] Pack management
- [ ] DNS TXT publishing
- [ ] Signature verification

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

MIT License - See [LICENSE](LICENSE) for details.