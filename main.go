package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dylanshade/pocket-prompt/internal/service"
	"github.com/dylanshade/pocket-prompt/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

var version = "0.1.0"

func main() {
	var showVersion bool
	var initLib bool

	flag.BoolVar(&showVersion, "version", false, "Print version information")
	flag.BoolVar(&initLib, "init", false, "Initialize a new prompt library")
	flag.Parse()

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
