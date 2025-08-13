package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dpshade/pocket-prompt/internal/cli"
	"github.com/dpshade/pocket-prompt/internal/service"
	"github.com/dpshade/pocket-prompt/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

var version = "0.1.0"

func printHelp() {
	fmt.Printf(`pocket-prompt - Terminal-based AI prompt management

USAGE:
    pocket-prompt [OPTIONS] [COMMAND]

OPTIONS:
    --help        Show this help information
    --version     Print version information  
    --init        Initialize a new prompt library

COMMANDS:
    (no command)       Start interactive TUI mode
    list, ls           List all prompts
    search <query>     Search prompts
    get, show <id>     Show a specific prompt
    create, new <id>   Create a new prompt
    edit <id>          Edit an existing prompt
    delete, rm <id>    Delete a prompt
    copy <id>          Copy prompt to clipboard
    render <id>        Render prompt with variables
    templates          List templates
    template           Template management (create, edit, delete, show)
    tags               List all tags
    archive            Manage archived prompts
    search-saved       Manage saved searches
    boolean-search     Boolean search operations (create, edit, delete, list, run)
    export             Export prompts and templates
    import             Import prompts and templates
    git                Git synchronization commands
    help               Show CLI command help

EXAMPLES:
    pocket-prompt                                    # Start interactive mode
    pocket-prompt --init                             # Initialize new library
    pocket-prompt list --format table               # List prompts in table format
    pocket-prompt search "machine learning"         # Search prompts
    pocket-prompt create my-prompt --title "Test"   # Create new prompt
    pocket-prompt render my-prompt --var name=John  # Render with variables
    pocket-prompt template create my-template        # Create template
    pocket-prompt boolean-search run "(ai OR ml)"   # Boolean search
    pocket-prompt export all --output backup.json   # Export everything
    pocket-prompt git setup <repo-url>              # Setup git sync
    pocket-prompt help <command>                     # Get detailed help

STORAGE:
    Default directory: ~/.pocket-prompt
    Override with: POCKET_PROMPT_DIR=<path>

For more information, visit: https://github.com/dpshade/pocket-prompt
`)
}

func main() {
	var showVersion bool
	var initLib bool
	var showHelp bool

	flag.BoolVar(&showVersion, "version", false, "Print version information")
	flag.BoolVar(&initLib, "init", false, "Initialize a new prompt library")
	flag.BoolVar(&showHelp, "help", false, "Show help information")
	flag.Parse()

	if showHelp {
		printHelp()
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("pocket-prompt version %s\n", version)
		os.Exit(0)
	}

	// Initialize service with file storage
	svc, err := service.NewService()
	if err != nil {
		fmt.Println(err)
		return
	}

	if initLib {
		if err := svc.InitLibrary(); err != nil {
			fmt.Println("Error initializing library:", err)
			return
		}
		fmt.Println("Initialized Pocket Prompt library")
		return
	}

	// Check if we have command line arguments for CLI mode
	args := flag.Args()
	if len(args) > 0 {
		// CLI mode - execute command and exit
		cliHandler := cli.NewCLI(svc)
		if err := cliHandler.ExecuteCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// No arguments provided - start TUI mode
	// Initialize TUI
	model, err := ui.NewModel(svc)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Start TUI program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		return
	}
}
