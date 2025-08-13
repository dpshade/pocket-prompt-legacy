package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dpshade/pocket-prompt/internal/clipboard"
	"github.com/dpshade/pocket-prompt/internal/models"
	"github.com/dpshade/pocket-prompt/internal/renderer"
	"github.com/dpshade/pocket-prompt/internal/service"
)

// CLI provides headless command-line interface functionality
type CLI struct {
	service *service.Service
}

// NewCLI creates a new CLI instance
func NewCLI(svc *service.Service) *CLI {
	return &CLI{service: svc}
}

// ExecuteCommand processes a CLI command and returns the result
func (c *CLI) ExecuteCommand(args []string) error {
	if len(args) == 0 {
		return c.printUsage()
	}

	command := args[0]
	commandArgs := args[1:]

	switch command {
	case "list", "ls":
		return c.listPrompts(commandArgs)
	case "search":
		return c.searchPrompts(commandArgs)
	case "get", "show":
		return c.showPrompt(commandArgs)
	case "create", "new":
		return c.createPrompt(commandArgs)
	case "edit":
		return c.editPrompt(commandArgs)
	case "delete", "rm":
		return c.deletePrompt(commandArgs)
	case "copy":
		return c.copyPrompt(commandArgs)
	case "render":
		return c.renderPrompt(commandArgs)
	case "templates":
		return c.handleTemplates(commandArgs)
	case "tags":
		return c.handleTags(commandArgs)
	case "archive":
		return c.handleArchive(commandArgs)
	case "search-saved":
		return c.handleSavedSearches(commandArgs)
	case "git":
		return c.handleGit(commandArgs)
	case "help":
		return c.printHelp(commandArgs)
	default:
		return fmt.Errorf("unknown command: %s. Use 'help' for usage information", command)
	}
}

// listPrompts lists all prompts
func (c *CLI) listPrompts(args []string) error {
	var format string
	var tag string
	var showArchived bool

	// Parse flags
	for i, arg := range args {
		switch arg {
		case "--format", "-f":
			if i+1 < len(args) {
				format = args[i+1]
			}
		case "--tag", "-t":
			if i+1 < len(args) {
				tag = args[i+1]
			}
		case "--archived", "-a":
			showArchived = true
		}
	}

	var prompts []*models.Prompt
	var err error

	if showArchived {
		prompts, err = c.service.ListArchivedPrompts()
	} else if tag != "" {
		prompts, err = c.service.FilterPromptsByTag(tag)
	} else {
		prompts, err = c.service.ListPrompts()
	}

	if err != nil {
		return fmt.Errorf("failed to list prompts: %w", err)
	}

	return c.formatOutput(prompts, format)
}

// searchPrompts searches prompts using query or boolean expression
func (c *CLI) searchPrompts(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("search requires a query")
	}

	var format string
	var boolean bool
	query := strings.Join(args, " ")

	// Parse flags from query
	parts := strings.Fields(query)
	var cleanedParts []string
	for i, part := range parts {
		switch part {
		case "--format", "-f":
			if i+1 < len(parts) {
				format = parts[i+1]
			}
		case "--boolean", "-b":
			boolean = true
		default:
			if i == 0 || (parts[i-1] != "--format" && parts[i-1] != "-f") {
				cleanedParts = append(cleanedParts, part)
			}
		}
	}

	query = strings.Join(cleanedParts, " ")

	var prompts []*models.Prompt
	var err error

	if boolean {
		// For now, implement a simple boolean search parser
		// This is a simplified implementation - a full parser would be more complex
		if strings.Contains(query, " AND ") || strings.Contains(query, " OR ") {
			return fmt.Errorf("boolean search not fully implemented in CLI mode yet - use simple tag filtering instead")
		}
		// Treat as simple tag search for now
		prompts, err = c.service.FilterPromptsByTag(query)
	} else {
		prompts, err = c.service.SearchPrompts(query)
	}

	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	return c.formatOutput(prompts, format)
}

// showPrompt displays a specific prompt
func (c *CLI) showPrompt(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("show requires a prompt ID")
	}

	id := args[0]
	var format string
	var render bool
	var variables map[string]interface{}

	// Parse flags
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--format", "-f":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		case "--render", "-r":
			render = true
		case "--var":
			if i+1 < len(args) {
				if variables == nil {
					variables = make(map[string]interface{})
				}
				parts := strings.SplitN(args[i+1], "=", 2)
				if len(parts) == 2 {
					variables[parts[0]] = parts[1]
				}
				i++
			}
		}
	}

	prompt, err := c.service.GetPrompt(id)
	if err != nil {
		return fmt.Errorf("failed to get prompt: %w", err)
	}

	if render {
		var template *models.Template
		if prompt.TemplateRef != "" {
			template, _ = c.service.GetTemplate(prompt.TemplateRef)
		}

		r := renderer.NewRenderer(prompt, template)
		
		switch format {
		case "json":
			content, err := r.RenderJSON(variables)
			if err != nil {
				return fmt.Errorf("failed to render JSON: %w", err)
			}
			fmt.Print(content)
		default:
			content, err := r.RenderText(variables)
			if err != nil {
				return fmt.Errorf("failed to render text: %w", err)
			}
			fmt.Print(content)
		}
		return nil
	}

	return c.formatSinglePrompt(prompt, format)
}

// createPrompt creates a new prompt
func (c *CLI) createPrompt(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("create requires a prompt ID")
	}

	id := args[0]
	var title, description, content, template string
	var tags []string

	// Parse flags
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--title":
			if i+1 < len(args) {
				title = args[i+1]
				i++
			}
		case "--description":
			if i+1 < len(args) {
				description = args[i+1]
				i++
			}
		case "--content":
			if i+1 < len(args) {
				content = args[i+1]
				i++
			}
		case "--template":
			if i+1 < len(args) {
				template = args[i+1]
				i++
			}
		case "--tags":
			if i+1 < len(args) {
				tags = strings.Split(args[i+1], ",")
				for j := range tags {
					tags[j] = strings.TrimSpace(tags[j])
				}
				i++
			}
		case "--stdin":
			// Read content from stdin
			var buf strings.Builder
			for {
				var line string
				n, err := fmt.Scanln(&line)
				if n == 0 || err != nil {
					break
				}
				buf.WriteString(line + "\n")
			}
			content = buf.String()
		}
	}

	prompt := &models.Prompt{
		ID:          id,
		Version:     "1.0.0",
		Name:        title,
		Summary:     description,
		Content:     content,
		Tags:        tags,
		TemplateRef: template,
	}

	if err := c.service.CreatePrompt(prompt); err != nil {
		return fmt.Errorf("failed to create prompt: %w", err)
	}

	fmt.Printf("Created prompt: %s\n", id)
	return nil
}

// editPrompt edits an existing prompt
func (c *CLI) editPrompt(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("edit requires a prompt ID")
	}

	id := args[0]
	prompt, err := c.service.GetPrompt(id)
	if err != nil {
		return fmt.Errorf("failed to get prompt: %w", err)
	}

	// Parse flags to update fields
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--title":
			if i+1 < len(args) {
				prompt.Name = args[i+1]
				i++
			}
		case "--description":
			if i+1 < len(args) {
				prompt.Summary = args[i+1]
				i++
			}
		case "--content":
			if i+1 < len(args) {
				prompt.Content = args[i+1]
				i++
			}
		case "--template":
			if i+1 < len(args) {
				prompt.TemplateRef = args[i+1]
				i++
			}
		case "--tags":
			if i+1 < len(args) {
				tags := strings.Split(args[i+1], ",")
				for j := range tags {
					tags[j] = strings.TrimSpace(tags[j])
				}
				prompt.Tags = tags
				i++
			}
		case "--add-tag":
			if i+1 < len(args) {
				tag := strings.TrimSpace(args[i+1])
				// Check if tag already exists
				found := false
				for _, t := range prompt.Tags {
					if t == tag {
						found = true
						break
					}
				}
				if !found {
					prompt.Tags = append(prompt.Tags, tag)
				}
				i++
			}
		case "--remove-tag":
			if i+1 < len(args) {
				tag := strings.TrimSpace(args[i+1])
				var newTags []string
				for _, t := range prompt.Tags {
					if t != tag {
						newTags = append(newTags, t)
					}
				}
				prompt.Tags = newTags
				i++
			}
		}
	}

	if err := c.service.UpdatePrompt(prompt); err != nil {
		return fmt.Errorf("failed to update prompt: %w", err)
	}

	fmt.Printf("Updated prompt: %s\n", id)
	return nil
}

// deletePrompt deletes a prompt
func (c *CLI) deletePrompt(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("delete requires a prompt ID")
	}

	id := args[0]
	var force bool

	// Parse flags
	for _, arg := range args[1:] {
		if arg == "--force" || arg == "-f" {
			force = true
		}
	}

	if !force {
		fmt.Printf("Are you sure you want to delete prompt '%s'? (y/N): ", id)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	if err := c.service.DeletePrompt(id); err != nil {
		return fmt.Errorf("failed to delete prompt: %w", err)
	}

	fmt.Printf("Deleted prompt: %s\n", id)
	return nil
}

// copyPrompt copies a prompt to clipboard
func (c *CLI) copyPrompt(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("copy requires a prompt ID")
	}

	id := args[0]
	var format string
	var variables map[string]interface{}

	// Parse flags
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--format", "-f":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		case "--var":
			if i+1 < len(args) {
				if variables == nil {
					variables = make(map[string]interface{})
				}
				parts := strings.SplitN(args[i+1], "=", 2)
				if len(parts) == 2 {
					variables[parts[0]] = parts[1]
				}
				i++
			}
		}
	}

	prompt, err := c.service.GetPrompt(id)
	if err != nil {
		return fmt.Errorf("failed to get prompt: %w", err)
	}

	var template *models.Template
	if prompt.TemplateRef != "" {
		template, _ = c.service.GetTemplate(prompt.TemplateRef)
	}

	r := renderer.NewRenderer(prompt, template)
	
	var content string
	switch format {
	case "json":
		content, err = r.RenderJSON(variables)
	default:
		content, err = r.RenderText(variables)
	}

	if err != nil {
		return fmt.Errorf("failed to render prompt: %w", err)
	}

	if err := clipboard.Copy(content); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	fmt.Printf("Copied prompt '%s' to clipboard\n", id)
	return nil
}

// renderPrompt renders a prompt with variables
func (c *CLI) renderPrompt(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("render requires a prompt ID")
	}

	id := args[0]
	var format string
	var variables map[string]interface{}

	// Parse flags
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--format", "-f":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		case "--var":
			if i+1 < len(args) {
				if variables == nil {
					variables = make(map[string]interface{})
				}
				parts := strings.SplitN(args[i+1], "=", 2)
				if len(parts) == 2 {
					variables[parts[0]] = parts[1]
				}
				i++
			}
		}
	}

	prompt, err := c.service.GetPrompt(id)
	if err != nil {
		return fmt.Errorf("failed to get prompt: %w", err)
	}

	var template *models.Template
	if prompt.TemplateRef != "" {
		template, _ = c.service.GetTemplate(prompt.TemplateRef)
	}

	r := renderer.NewRenderer(prompt, template)
	
	switch format {
	case "json":
		content, err := r.RenderJSON(variables)
		if err != nil {
			return fmt.Errorf("failed to render JSON: %w", err)
		}
		fmt.Print(content)
	default:
		content, err := r.RenderText(variables)
		if err != nil {
			return fmt.Errorf("failed to render text: %w", err)
		}
		fmt.Print(content)
	}

	return nil
}

// formatOutput formats prompts for output
func (c *CLI) formatOutput(prompts []*models.Prompt, format string) error {
	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(prompts)
	case "ids":
		for _, p := range prompts {
			fmt.Println(p.ID)
		}
	case "table":
		fmt.Printf("%-20s %-30s %-15s %s\n", "ID", "Title", "Version", "Updated")
		fmt.Println(strings.Repeat("-", 80))
		for _, p := range prompts {
			title := p.Name
			if len(title) > 30 {
				title = title[:27] + "..."
			}
			fmt.Printf("%-20s %-30s %-15s %s\n", 
				p.ID, title, p.Version, p.UpdatedAt.Format("2006-01-02"))
		}
	default:
		for _, p := range prompts {
			fmt.Printf("%s - %s\n", p.ID, p.Name)
			if p.Summary != "" {
				fmt.Printf("  %s\n", p.Summary)
			}
			if len(p.Tags) > 0 {
				fmt.Printf("  Tags: %s\n", strings.Join(p.Tags, ", "))
			}
			fmt.Println()
		}
	}
	return nil
}

// formatSinglePrompt formats a single prompt for output
func (c *CLI) formatSinglePrompt(prompt *models.Prompt, format string) error {
	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(prompt)
	default:
		fmt.Printf("ID: %s\n", prompt.ID)
		fmt.Printf("Title: %s\n", prompt.Name)
		fmt.Printf("Version: %s\n", prompt.Version)
		if prompt.Summary != "" {
			fmt.Printf("Description: %s\n", prompt.Summary)
		}
		if len(prompt.Tags) > 0 {
			fmt.Printf("Tags: %s\n", strings.Join(prompt.Tags, ", "))
		}
		if prompt.TemplateRef != "" {
			fmt.Printf("Template: %s\n", prompt.TemplateRef)
		}
		fmt.Printf("Created: %s\n", prompt.CreatedAt.Format("2006-01-02 15:04"))
		fmt.Printf("Updated: %s\n", prompt.UpdatedAt.Format("2006-01-02 15:04"))
		fmt.Printf("\nContent:\n%s\n", prompt.Content)
	}
	return nil
}

// Additional command handlers would go here...
// This is a simplified implementation focusing on core functionality

func (c *CLI) handleTemplates(args []string) error {
	if len(args) == 0 {
		// List templates
		templates, err := c.service.ListTemplates()
		if err != nil {
			return fmt.Errorf("failed to list templates: %w", err)
		}

		for _, t := range templates {
			fmt.Printf("%s - %s\n", t.ID, t.Name)
			if t.Description != "" {
				fmt.Printf("  %s\n", t.Description)
			}
			fmt.Println()
		}
		return nil
	}

	subcommand := args[0]
	switch subcommand {
	case "show":
		if len(args) < 2 {
			return fmt.Errorf("templates show requires a template ID")
		}
		template, err := c.service.GetTemplate(args[1])
		if err != nil {
			return fmt.Errorf("failed to get template: %w", err)
		}
		
		fmt.Printf("ID: %s\n", template.ID)
		fmt.Printf("Name: %s\n", template.Name)
		if template.Description != "" {
			fmt.Printf("Description: %s\n", template.Description)
		}
		fmt.Printf("Created: %s\n", template.CreatedAt.Format("2006-01-02 15:04"))
		fmt.Printf("Updated: %s\n", template.UpdatedAt.Format("2006-01-02 15:04"))
		fmt.Printf("\nContent:\n%s\n", template.Content)
		
		if len(template.Slots) > 0 {
			fmt.Println("\nSlots:")
			for _, slot := range template.Slots {
				fmt.Printf("  %s", slot.Name)
				if slot.Required {
					fmt.Print(" [required]")
				}
				if slot.Default != "" {
					fmt.Printf(" [default: %s]", slot.Default)
				}
				if slot.Description != "" {
					fmt.Printf(" - %s", slot.Description)
				}
				fmt.Println()
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown templates subcommand: %s", subcommand)
	}
}

func (c *CLI) handleTags(args []string) error {
	tags, err := c.service.GetAllTags()
	if err != nil {
		return fmt.Errorf("failed to get tags: %w", err)
	}

	for _, tag := range tags {
		fmt.Println(tag)
	}
	return nil
}

func (c *CLI) handleArchive(args []string) error {
	if len(args) == 0 {
		// List archived prompts
		prompts, err := c.service.ListArchivedPrompts()
		if err != nil {
			return fmt.Errorf("failed to list archived prompts: %w", err)
		}
		return c.formatOutput(prompts, "")
	}
	return fmt.Errorf("archive subcommands not implemented")
}

func (c *CLI) handleSavedSearches(args []string) error {
	if len(args) == 0 {
		// List saved searches
		searches, err := c.service.ListSavedSearches()
		if err != nil {
			return fmt.Errorf("failed to list saved searches: %w", err)
		}

		for _, search := range searches {
			fmt.Printf("%s: %s\n", search.Name, search.Expression.String())
		}
		return nil
	}

	subcommand := args[0]
	switch subcommand {
	case "run":
		if len(args) < 2 {
			return fmt.Errorf("search-saved run requires a search name")
		}
		prompts, err := c.service.ExecuteSavedSearch(args[1])
		if err != nil {
			return fmt.Errorf("failed to execute saved search: %w", err)
		}
		return c.formatOutput(prompts, "")
	default:
		return fmt.Errorf("unknown search-saved subcommand: %s", subcommand)
	}
}

func (c *CLI) handleGit(args []string) error {
	if len(args) == 0 {
		// Show git status
		status, err := c.service.GetGitSyncStatus()
		if err != nil {
			return fmt.Errorf("failed to get git status: %w", err)
		}
		fmt.Println("Git sync status:", status)
		return nil
	}

	subcommand := args[0]
	switch subcommand {
	case "enable":
		c.service.EnableGitSync()
		fmt.Println("Git sync enabled")
		return nil
	case "disable":
		c.service.DisableGitSync()
		fmt.Println("Git sync disabled")
		return nil
	case "status":
		status, err := c.service.GetGitSyncStatus()
		if err != nil {
			return fmt.Errorf("failed to get git status: %w", err)
		}
		fmt.Println(status)
		return nil
	default:
		return fmt.Errorf("unknown git subcommand: %s", subcommand)
	}
}

func (c *CLI) printUsage() error {
	fmt.Println(`pocket-prompt - Headless CLI mode

Usage: pocket-prompt <command> [options]

Commands:
  list, ls              List all prompts
  search <query>        Search prompts
  get, show <id>        Show a specific prompt
  create, new <id>      Create a new prompt
  edit <id>             Edit an existing prompt
  delete, rm <id>       Delete a prompt
  copy <id>             Copy prompt to clipboard
  render <id>           Render prompt with variables
  templates             Manage templates
  tags                  List all tags
  archive               Manage archived prompts
  search-saved          Manage saved searches
  git                   Git synchronization
  help                  Show help

Use 'pocket-prompt help <command>' for detailed help on a specific command.`)
	return nil
}

func (c *CLI) printHelp(args []string) error {
	if len(args) == 0 {
		return c.printUsage()
	}

	command := args[0]
	switch command {
	case "list", "ls":
		fmt.Println(`list - List all prompts

Usage: pocket-prompt list [options]

Options:
  --format, -f <format>  Output format (table, json, ids, default)
  --tag, -t <tag>        Filter by tag
  --archived, -a         Show archived prompts`)

	case "search":
		fmt.Println(`search - Search prompts

Usage: pocket-prompt search <query> [options]

Options:
  --format, -f <format>  Output format (table, json, ids, default)
  --boolean, -b          Use boolean expression search

Examples:
  pocket-prompt search "machine learning"
  pocket-prompt search --boolean "(ai AND analysis) OR writing"`)

	case "create", "new":
		fmt.Println(`create - Create a new prompt

Usage: pocket-prompt create <id> [options]

Options:
  --title <title>        Prompt title
  --description <desc>   Prompt description
  --content <content>    Prompt content
  --template <id>        Template to use
  --tags <tag1,tag2>     Comma-separated tags
  --stdin                Read content from stdin

Example:
  pocket-prompt create my-prompt --title "My Prompt" --content "Hello world"`)

	case "render":
		fmt.Println(`render - Render prompt with variables

Usage: pocket-prompt render <id> [options]

Options:
  --format, -f <format>  Output format (text, json)
  --var <name=value>     Set variable value (can be used multiple times)

Example:
  pocket-prompt render my-prompt --var name=John --var age=30`)

	default:
		fmt.Printf("No help available for command: %s\n", command)
	}

	return nil
}