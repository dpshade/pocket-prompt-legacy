package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/dylanshade/pocket-prompt/internal/clipboard"
	"github.com/dylanshade/pocket-prompt/internal/models"
	"github.com/dylanshade/pocket-prompt/internal/renderer"
	"github.com/dylanshade/pocket-prompt/internal/service"
)

// ViewMode represents the current view in the TUI
type ViewMode int

const (
	ViewLibrary ViewMode = iota
	ViewPromptDetail
	ViewVariables
	ViewCreateMenu
	ViewCreateFromScratch
	ViewCreateFromTemplate
	ViewTemplateList
	ViewEditPrompt
	ViewEditTemplate
	ViewTemplateDetail
	ViewTemplateManagement
)

// Model represents the TUI application state
type Model struct {
	service  *service.Service
	viewMode ViewMode

	// UI components
	promptList list.Model
	viewport   viewport.Model
	help       help.Model
	keys       KeyMap

	// Data
	prompts        []*models.Prompt
	templates      []*models.Template
	selectedPrompt *models.Prompt
	selectedTemplate *models.Template
	variables      map[string]interface{}

	// Creation state
	newPrompt      *models.Prompt
	createForm     *CreateForm
	templateForm   *TemplateForm
	selectForm     *SelectForm
	editMode       bool

	// Rendered content
	renderedContent     string
	renderedContentJSON string
	glamourRenderer     *glamour.TermRenderer

	// Window dimensions
	width  int
	height int

	// Status messages
	statusMsg     string
	statusTimeout int

	// Error state
	err error
}

// KeyMap defines all key bindings
type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Enter  key.Binding
	Back   key.Binding
	Quit   key.Binding
	Help   key.Binding
	Search key.Binding
	Copy     key.Binding
	CopyJSON key.Binding
	Export   key.Binding
	New      key.Binding
	Edit     key.Binding
	Templates key.Binding
}

// ShortHelp returns keybindings to show in the mini help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings to show in the full help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Enter, k.Back, k.Search, k.New},
		{k.Edit, k.Templates, k.Copy, k.CopyJSON},
		{k.Export, k.Help, k.Quit},
	}
}

var keys = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "back"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "forward"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "b"),
		key.WithHelp("esc/b", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Copy: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "copy"),
	),
	CopyJSON: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "copy as JSON"),
	),
	Export: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "export"),
	),
	New: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new prompt"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	Templates: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "templates"),
	),
}

// NewModel creates a new TUI model
func NewModel(svc *service.Service) (*Model, error) {
	// Load prompts
	prompts, err := svc.ListPrompts()
	if err != nil {
		// Continue even if we can't load prompts initially
		prompts = []*models.Prompt{}
	}

	// Load templates
	templates, err := svc.ListTemplates()
	if err != nil {
		// It's okay if templates fail to load
		templates = []*models.Template{}
	}

	// Convert prompts to list items
	items := make([]list.Item, len(prompts))
	for i, p := range prompts {
		items[i] = p
	}

	// Create list
	l := list.New(items, list.NewDefaultDelegate(), 80, 20) // Default size, will be updated on first WindowSizeMsg
	l.Title = ""  // We'll handle title in the view
	l.SetShowStatusBar(false) // We'll handle status in our custom view
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false) // We'll handle help text ourselves

	// Create viewport for preview
	vp := viewport.New(80, 20) // Default size, will be updated on first WindowSizeMsg
	vp.Style = lipgloss.NewStyle().
		Padding(1, 2)

	// Create glamour renderer for markdown
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create glamour renderer: %w", err)
	}

	return &Model{
		service:         svc,
		viewMode:        ViewLibrary,
		promptList:      l,
		viewport:        vp,
		help:            help.New(),
		keys:            keys,
		prompts:         prompts,
		templates:       templates,
		variables:       make(map[string]interface{}),
		glamourRenderer: renderer,
	}, nil
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// tickMsg is sent to clear the status message
type tickMsg time.Time

// clearStatusCmd returns a command that clears the status message after a delay
func clearStatusCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tickMsg:
		if m.statusTimeout > 0 {
			m.statusTimeout--
			if m.statusTimeout == 0 {
				m.statusMsg = ""
			} else {
				return m, clearStatusCmd()
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update component sizes based on current view
		switch m.viewMode {
		case ViewLibrary:
			// Library takes full width
			m.promptList.SetSize(msg.Width, msg.Height-6) // Reserve space for title and help
		case ViewPromptDetail:
			// Viewport takes full width for detail view
			m.viewport.Width = msg.Width - 4  // Padding
			m.viewport.Height = msg.Height - 8 // Reserve space for title, metadata, and help
		}

		// Update glamour renderer width for detail view
		glamourWidth := m.width - 8 // Account for padding
		if glamourWidth > 0 {
			renderer, err := glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(glamourWidth),
			)
			if err == nil {
				m.glamourRenderer = renderer
			}
		}

	case tea.KeyMsg:
		if m.promptList.FilterState() == list.Filtering {
			// Let the list handle filtering
			newListModel, cmd := m.promptList.Update(msg)
			m.promptList = newListModel
			return m, cmd
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Enter):
			if m.viewMode == ViewLibrary {
				if i, ok := m.promptList.SelectedItem().(*models.Prompt); ok {
					m.selectedPrompt = i
					m.viewMode = ViewPromptDetail
					// Render the prompt preview
					if err := m.renderPreview(); err != nil {
						m.err = err
					}
				}
			}

		default:
			// Handle Ctrl+S for saving forms
			if msg.String() == "ctrl+s" {
				switch m.viewMode {
				case ViewEditPrompt:
					if m.createForm != nil && m.createForm.IsSubmitted() {
						// Save the prompt
						prompt := m.createForm.ToPrompt()
						if m.editMode && m.selectedPrompt != nil {
							// Keep original creation date and increment version for edits
							prompt.CreatedAt = m.selectedPrompt.CreatedAt
						}
						if err := m.service.SavePrompt(prompt); err != nil {
							m.statusMsg = fmt.Sprintf("Save failed: %v", err)
							m.statusTimeout = 3
						} else {
							m.statusMsg = "Prompt saved successfully!"
							m.statusTimeout = 2
							// Refresh prompt list
							if prompts, err := m.service.ListPrompts(); err == nil {
								m.prompts = prompts
								// Update list items
								items := make([]list.Item, len(prompts))
								for i, p := range prompts {
									items[i] = p
								}
								m.promptList.SetItems(items)
							}
							// Go back to library
							m.viewMode = ViewLibrary
							m.createForm = nil
							m.editMode = false
						}
						return m, clearStatusCmd()
					}
				case ViewEditTemplate:
					if m.templateForm != nil && m.templateForm.IsSubmitted() {
						// Save the template
						template := m.templateForm.ToTemplate()
						if m.editMode && m.selectedTemplate != nil {
							// Keep original creation date for edits
							template.CreatedAt = m.selectedTemplate.CreatedAt
						}
						if err := m.service.SaveTemplate(template); err != nil {
							m.statusMsg = fmt.Sprintf("Save failed: %v", err)
							m.statusTimeout = 3
						} else {
							m.statusMsg = "Template saved successfully!"
							m.statusTimeout = 2
							// Refresh template list
							if templates, err := m.service.ListTemplates(); err == nil {
								m.templates = templates
							}
							// Go back to template management
							m.viewMode = ViewTemplateManagement
							m.templateForm = nil
							m.editMode = false
						}
						return m, clearStatusCmd()
					}
				}
			}
			

		case key.Matches(msg, m.keys.Back):
			switch m.viewMode {
			case ViewPromptDetail, ViewVariables:
				m.viewMode = ViewLibrary
				m.selectedPrompt = nil
				m.renderedContent = ""
				m.renderedContentJSON = ""
			case ViewCreateMenu, ViewCreateFromScratch, ViewCreateFromTemplate, ViewTemplateList:
				if m.viewMode == ViewTemplateList || m.viewMode == ViewCreateFromTemplate {
					m.viewMode = ViewCreateMenu
				} else {
					m.viewMode = ViewLibrary
				}
				m.newPrompt = nil
				m.createForm = nil
				m.selectForm = nil
			case ViewEditPrompt, ViewEditTemplate:
				m.viewMode = ViewLibrary
				m.createForm = nil
				m.templateForm = nil
				m.editMode = false
			case ViewTemplateManagement, ViewTemplateDetail:
				if m.viewMode == ViewTemplateDetail {
					m.viewMode = ViewTemplateManagement
				} else {
					m.viewMode = ViewLibrary
				}
				m.selectedTemplate = nil
				m.selectForm = nil
			}

		case key.Matches(msg, m.keys.Left):
			// Use left arrow for back navigation in detail views
			switch m.viewMode {
			case ViewPromptDetail, ViewVariables:
				m.viewMode = ViewLibrary
				m.selectedPrompt = nil
				m.renderedContent = ""
				m.renderedContentJSON = ""
				return m, nil
			case ViewCreateMenu, ViewCreateFromScratch, ViewCreateFromTemplate, ViewTemplateList:
				if m.viewMode == ViewTemplateList || m.viewMode == ViewCreateFromTemplate {
					m.viewMode = ViewCreateMenu
				} else {
					m.viewMode = ViewLibrary
				}
				m.newPrompt = nil
				m.createForm = nil
				m.selectForm = nil
				return m, nil
			case ViewEditPrompt, ViewEditTemplate:
				m.viewMode = ViewLibrary
				m.createForm = nil
				m.templateForm = nil
				m.editMode = false
				return m, nil
			case ViewTemplateManagement, ViewTemplateDetail:
				if m.viewMode == ViewTemplateDetail {
					m.viewMode = ViewTemplateManagement
				} else {
					m.viewMode = ViewLibrary
				}
				m.selectedTemplate = nil
				m.selectForm = nil
				return m, nil
			}

		case key.Matches(msg, m.keys.Search):
			if m.viewMode == ViewLibrary {
				m.promptList.SetFilteringEnabled(true)
				return m, nil
			}

		case key.Matches(msg, m.keys.New):
			if m.viewMode == ViewLibrary {
				// Initialize the create menu select form
				options := []SelectOption{
					{
						Label:       "Create from scratch",
						Description: "Start with a blank prompt",
						Value:       "scratch",
					},
					{
						Label:       "Use a template",
						Description: "Start from an existing template",
						Value:       "template",
					},
				}
				m.selectForm = NewSelectForm(options)
				m.viewMode = ViewCreateMenu
				return m, nil
			}

		case key.Matches(msg, m.keys.Edit):
			switch m.viewMode {
			case ViewLibrary:
				if i, ok := m.promptList.SelectedItem().(*models.Prompt); ok {
					m.selectedPrompt = i
					m.createForm = NewCreateForm()
					m.createForm.LoadPrompt(i)
					m.editMode = true
					m.viewMode = ViewEditPrompt
				}
			case ViewPromptDetail:
				if m.selectedPrompt != nil {
					m.createForm = NewCreateForm()
					m.createForm.LoadPrompt(m.selectedPrompt)
					m.editMode = true
					m.viewMode = ViewEditPrompt
				}
			case ViewTemplateDetail:
				if m.selectedTemplate != nil {
					m.templateForm = NewTemplateForm()
					m.templateForm.LoadTemplate(m.selectedTemplate)
					m.editMode = true
					m.viewMode = ViewEditTemplate
				}
			}

		case key.Matches(msg, m.keys.Templates):
			if m.viewMode == ViewLibrary {
				// Create template management select form
				options := []SelectOption{
					{
						Label:       "Create new template",
						Description: "Start with a blank template",
						Value:       "new",
					},
				}
				// Add existing templates as options
				for _, template := range m.templates {
					options = append(options, SelectOption{
						Label:       template.Name,
						Description: template.Description,
						Value:       template,
					})
				}
				m.selectForm = NewSelectForm(options)
				m.viewMode = ViewTemplateManagement
				return m, nil
			}

		case key.Matches(msg, m.keys.Copy):
			if m.viewMode == ViewPromptDetail && m.renderedContent != "" {
				if statusMsg, err := clipboard.CopyWithFallback(m.renderedContent); err != nil {
					m.statusMsg = fmt.Sprintf("Copy failed: %v", err)
					m.statusTimeout = 3
				} else {
					m.statusMsg = statusMsg
					m.statusTimeout = 2
				}
				return m, clearStatusCmd()
			}

		case key.Matches(msg, m.keys.CopyJSON):
			if m.viewMode == ViewPromptDetail && m.renderedContentJSON != "" {
				if _, err := clipboard.CopyWithFallback(m.renderedContentJSON); err != nil {
					m.statusMsg = fmt.Sprintf("JSON copy failed: %v", err)
					m.statusTimeout = 3
				} else {
					m.statusMsg = "Copied as JSON messages!"
					m.statusTimeout = 2
				}
				return m, clearStatusCmd()
			}

		}
	}

	// Update the appropriate component based on view mode
	switch m.viewMode {
	case ViewLibrary:
		newListModel, cmd := m.promptList.Update(msg)
		m.promptList = newListModel
		cmds = append(cmds, cmd)

	case ViewPromptDetail:
		newViewport, cmd := m.viewport.Update(msg)
		m.viewport = newViewport
		cmds = append(cmds, cmd)

	case ViewVariables:
		// TODO: Handle variable form updates

	case ViewCreateMenu:
		if m.selectForm != nil {
			cmd := m.selectForm.Update(msg)
			cmds = append(cmds, cmd)
			// Check if an option was selected
			if m.selectForm.IsSubmitted() {
				selected := m.selectForm.GetSelected()
				if selected != nil {
					switch selected.Value {
					case "scratch":
						m.viewMode = ViewCreateFromScratch
						m.createForm = NewCreateFormFromScratch()
					case "template":
						// Initialize template selection
						if len(m.templates) > 0 {
							templateOptions := make([]SelectOption, len(m.templates))
							for i, template := range m.templates {
								templateOptions[i] = SelectOption{
									Label:       template.Name,
									Description: template.Description,
									Value:       template,
								}
							}
							m.selectForm = NewSelectForm(templateOptions)
							m.viewMode = ViewTemplateList
						} else {
							m.statusMsg = "No templates available"
							m.statusTimeout = 2
							m.viewMode = ViewLibrary
							cmds = append(cmds, clearStatusCmd())
						}
					}
				}
			}
		}

	case ViewTemplateList:
		if m.selectForm != nil {
			cmd := m.selectForm.Update(msg)
			cmds = append(cmds, cmd)
			// Check if a template was selected
			if m.selectForm.IsSubmitted() {
				selected := m.selectForm.GetSelected()
				if selected != nil {
					if template, ok := selected.Value.(*models.Template); ok {
						m.selectedTemplate = template
						m.viewMode = ViewCreateFromTemplate
						// TODO: Initialize form with template
					}
				}
			}
		}

	case ViewEditPrompt:
		if m.createForm != nil {
			cmd := m.createForm.Update(msg)
			cmds = append(cmds, cmd)
		}

	case ViewEditTemplate:
		if m.templateForm != nil {
			cmd := m.templateForm.Update(msg)
			cmds = append(cmds, cmd)
		}

	case ViewCreateFromScratch:
		if m.createForm != nil {
			cmd := m.createForm.Update(msg)
			cmds = append(cmds, cmd)
			// Check if form was submitted
			if m.createForm.IsSubmitted() {
				prompt := m.createForm.ToPrompt()
				if err := m.service.SavePrompt(prompt); err != nil {
					m.statusMsg = fmt.Sprintf("Save failed: %v", err)
					m.statusTimeout = 3
				} else {
					m.statusMsg = "Prompt created successfully!"
					m.statusTimeout = 2
					// Refresh prompt list
					if prompts, err := m.service.ListPrompts(); err == nil {
						m.prompts = prompts
						// Update list items
						items := make([]list.Item, len(prompts))
						for i, p := range prompts {
							items[i] = p
						}
						m.promptList.SetItems(items)
					}
					// Go back to library
					m.viewMode = ViewLibrary
					m.createForm = nil
				}
				cmds = append(cmds, clearStatusCmd())
			}
		}

	case ViewTemplateManagement:
		if m.selectForm != nil {
			cmd := m.selectForm.Update(msg)
			cmds = append(cmds, cmd)
			// Check if an option was selected
			if m.selectForm.IsSubmitted() {
				selected := m.selectForm.GetSelected()
				if selected != nil {
					switch selected.Value {
					case "new":
						m.templateForm = NewTemplateFormFromScratch()
						m.editMode = false
						m.viewMode = ViewEditTemplate
						m.selectForm = nil
					default:
						// Selected an existing template
						if template, ok := selected.Value.(*models.Template); ok {
							m.selectedTemplate = template
							m.viewMode = ViewTemplateDetail
							m.selectForm = nil
						}
					}
				}
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n  Error: %v\n\n  Press 'q' to quit.\n", m.err)
	}

	var mainView string

	switch m.viewMode {
	case ViewLibrary:
		mainView = m.renderLibraryView()

	case ViewPromptDetail:
		mainView = m.renderPromptDetailView()

	case ViewVariables:
		mainView = m.renderVariablesView()

	case ViewCreateMenu:
		mainView = m.renderCreateMenuView()

	case ViewCreateFromScratch:
		mainView = m.renderCreateFromScratchView()

	case ViewCreateFromTemplate:
		mainView = m.renderCreateFromTemplateView()

	case ViewTemplateList:
		mainView = m.renderTemplateListView()

	case ViewEditPrompt:
		mainView = m.renderEditPromptView()

	case ViewEditTemplate:
		mainView = m.renderEditTemplateView()

	case ViewTemplateDetail:
		mainView = m.renderTemplateDetailView()

	case ViewTemplateManagement:
		mainView = m.renderTemplateManagementView()

	default:
		mainView = "Unknown view mode"
	}

	// Add status message at the bottom if present
	if m.statusMsg != "" {
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Bold(true).
			Padding(0, 1)
		
		statusBar := statusStyle.Render(m.statusMsg)
		return lipgloss.JoinVertical(lipgloss.Left, mainView, statusBar)
	}

	return mainView
}

// renderLibraryView renders the prompt library list
func (m Model) renderLibraryView() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Padding(0, 1)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(0, 1)

	title := titleStyle.Render("Pocket Prompt Library")
	help := helpStyle.Render("Press Enter to view • e to edit • n to create • t for templates • / to search • q to quit")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		m.promptList.View(),
		help,
	)
}

// renderPromptDetailView renders the selected prompt in full-page view
func (m Model) renderPromptDetailView() string {
	if m.selectedPrompt == nil {
		return "No prompt selected"
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Padding(0, 1)

	backButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		MarginRight(2)

	metadataStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(0, 1)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(0, 1)

	// Create back button and title
	backButton := backButtonStyle.Render("← Back")
	title := titleStyle.Render(m.selectedPrompt.Title())
	
	headerLine := lipgloss.JoinHorizontal(lipgloss.Left, backButton, title)

	// Create metadata line
	metadata := fmt.Sprintf("ID: %s • Version: %s", m.selectedPrompt.ID, m.selectedPrompt.Version)
	if len(m.selectedPrompt.Tags) > 0 {
		tags := ""
		for i, tag := range m.selectedPrompt.Tags {
			if i > 0 {
				tags += ", "
			}
			tags += tag
		}
		metadata += fmt.Sprintf(" • Tags: %s", tags)
	}
	metadataLine := metadataStyle.Render(metadata)

	// Help text
	help := helpStyle.Render("Press c to copy • y to copy as JSON • e to edit • ←/esc/b to go back")

	// Content viewport
	content := m.viewport.View()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerLine,
		metadataLine,
		"",
		content,
		"",
		help,
	)
}

// renderVariablesView renders the variables form (placeholder)
func (m Model) renderVariablesView() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Padding(0, 1)

	backButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		MarginRight(2)

	// Create back button and title
	backButton := backButtonStyle.Render("← Back")
	title := titleStyle.Render("Prompt Variables")
	
	headerLine := lipgloss.JoinHorizontal(lipgloss.Left, backButton, title)
	
	// Placeholder for variables form
	content := "Variables form coming soon...\n\nPress esc/b to go back"

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerLine,
		"",
		content,
	)
}

// renderCreateMenuView renders the create menu using SelectForm
func (m Model) renderCreateMenuView() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Padding(0, 1)

	backButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		MarginRight(2)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("33")).
		Bold(true).
		Padding(0, 1)

	unselectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Padding(0, 1)

	descriptionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		Padding(0, 3)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(0, 1)

	// Create back button and title
	backButton := backButtonStyle.Render("← Back")
	title := titleStyle.Render("Create New Prompt")
	
	headerLine := lipgloss.JoinHorizontal(lipgloss.Left, backButton, title)

	if m.selectForm == nil {
		return lipgloss.JoinVertical(lipgloss.Left, headerLine, "", "No options available")
	}

	// Render options
	var optionLines []string
	for i, option := range m.selectForm.options {
		var style lipgloss.Style
		if i == m.selectForm.selected {
			style = selectedStyle
		} else {
			style = unselectedStyle
		}
		
		optionLine := style.Render("▶ " + option.Label)
		optionLines = append(optionLines, optionLine)
		
		if option.Description != "" {
			descLine := descriptionStyle.Render(option.Description)
			optionLines = append(optionLines, descLine)
		}
		
		optionLines = append(optionLines, "") // Add spacing
	}

	help := helpStyle.Render("↑/↓ or k/j to navigate • Enter to select • ←/esc/b to go back")

	// Join all elements
	allElements := []string{headerLine, ""}
	allElements = append(allElements, optionLines...)
	allElements = append(allElements, help)

	return lipgloss.JoinVertical(lipgloss.Left, allElements...)
}

// renderCreateFromScratchView renders the create from scratch form
func (m Model) renderCreateFromScratchView() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Padding(0, 1)

	backButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		MarginRight(2)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Bold(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(0, 1)

	// Create back button and title
	backButton := backButtonStyle.Render("← Back")
	title := titleStyle.Render("Create from Scratch")
	
	headerLine := lipgloss.JoinHorizontal(lipgloss.Left, backButton, title)

	if m.createForm == nil {
		return lipgloss.JoinVertical(lipgloss.Left, headerLine, "", "No form available")
	}

	// Build simplified form fields - only essentials for "from scratch"
	var formFields []string

	// ID field
	idLabel := labelStyle.Render("ID:")
	formFields = append(formFields, idLabel, m.createForm.inputs[idField].View(), "")

	// Title field
	titleLabel := labelStyle.Render("Title:")
	formFields = append(formFields, titleLabel, m.createForm.inputs[titleField].View(), "")

	// Content field
	contentLabel := labelStyle.Render("Content:")
	formFields = append(formFields, contentLabel, m.createForm.textarea.View(), "")

	// Help text
	help := helpStyle.Render("Tab/↓ next field • Shift+Tab/↑ prev field • Ctrl+S save • ←/esc/b cancel")

	// Join all elements
	allElements := []string{headerLine, ""}
	allElements = append(allElements, formFields...)
	allElements = append(allElements, help)

	return lipgloss.JoinVertical(lipgloss.Left, allElements...)
}

// renderCreateFromTemplateView renders template-based creation
func (m Model) renderCreateFromTemplateView() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Padding(0, 1)

	backButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		MarginRight(2)

	// Create back button and title
	backButton := backButtonStyle.Render("← Back")
	title := titleStyle.Render("Create from Template")
	
	headerLine := lipgloss.JoinHorizontal(lipgloss.Left, backButton, title)

	content := "Template creation form will go here...\n\nPress ←/esc/b to go back"

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerLine,
		"",
		content,
	)
}

// renderTemplateListView renders the template selection list using SelectForm
func (m Model) renderTemplateListView() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Padding(0, 1)

	backButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		MarginRight(2)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("33")).
		Bold(true).
		Padding(0, 1)

	unselectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Padding(0, 1)

	descriptionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		Padding(0, 3)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(0, 1)

	// Create back button and title
	backButton := backButtonStyle.Render("← Back")
	title := titleStyle.Render("Select Template")
	
	headerLine := lipgloss.JoinHorizontal(lipgloss.Left, backButton, title)

	if m.selectForm == nil || len(m.selectForm.options) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, headerLine, "", "No templates available")
	}

	// Render template options
	var optionLines []string
	for i, option := range m.selectForm.options {
		var style lipgloss.Style
		if i == m.selectForm.selected {
			style = selectedStyle
		} else {
			style = unselectedStyle
		}
		
		optionLine := style.Render("▶ " + option.Label)
		optionLines = append(optionLines, optionLine)
		
		if option.Description != "" {
			descLine := descriptionStyle.Render(option.Description)
			optionLines = append(optionLines, descLine)
		}
		
		optionLines = append(optionLines, "") // Add spacing
	}

	help := helpStyle.Render("↑/↓ or k/j to navigate • Enter to select • ←/esc/b to go back")

	// Join all elements
	allElements := []string{headerLine, ""}
	allElements = append(allElements, optionLines...)
	allElements = append(allElements, help)

	return lipgloss.JoinVertical(lipgloss.Left, allElements...)
}

// renderEditPromptView renders the prompt editing form
func (m Model) renderEditPromptView() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Padding(0, 1)

	backButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		MarginRight(2)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Bold(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(0, 1)

	// Create back button and title
	backButton := backButtonStyle.Render("← Back")
	title := titleStyle.Render("Edit Prompt")
	
	headerLine := lipgloss.JoinHorizontal(lipgloss.Left, backButton, title)

	if m.createForm == nil {
		return lipgloss.JoinVertical(lipgloss.Left, headerLine, "", "No form available")
	}

	// Build form fields
	var formFields []string

	// ID field (read-only in edit mode)
	idLabel := labelStyle.Render("ID (read-only):")
	idValue := ""
	if m.editMode && m.selectedPrompt != nil {
		idValue = m.selectedPrompt.ID
	} else {
		idValue = m.createForm.inputs[idField].View()
	}
	formFields = append(formFields, idLabel, idValue, "")

	// Version field
	versionLabel := labelStyle.Render("Version:")
	formFields = append(formFields, versionLabel, m.createForm.inputs[versionField].View(), "")

	// Title field
	titleLabel := labelStyle.Render("Title:")
	formFields = append(formFields, titleLabel, m.createForm.inputs[titleField].View(), "")

	// Description field
	descLabel := labelStyle.Render("Description:")
	formFields = append(formFields, descLabel, m.createForm.inputs[descriptionField].View(), "")

	// Tags field
	tagsLabel := labelStyle.Render("Tags:")
	formFields = append(formFields, tagsLabel, m.createForm.inputs[tagsField].View(), "")

	// Variables field
	variablesLabel := labelStyle.Render("Variables:")
	formFields = append(formFields, variablesLabel, m.createForm.inputs[variablesField].View(), "")

	// Template reference field
	templateRefLabel := labelStyle.Render("Template Ref:")
	formFields = append(formFields, templateRefLabel, m.createForm.inputs[templateRefField].View(), "")

	// Content field
	contentLabel := labelStyle.Render("Content:")
	formFields = append(formFields, contentLabel, m.createForm.textarea.View(), "")

	// Help text
	help := helpStyle.Render("Tab/↓ next field • Shift+Tab/↑ prev field • Ctrl+S save • ←/esc/b cancel")

	// Join all elements
	allElements := []string{headerLine, ""}
	allElements = append(allElements, formFields...)
	allElements = append(allElements, help)

	return lipgloss.JoinVertical(lipgloss.Left, allElements...)
}

// renderEditTemplateView renders the template editing form
func (m Model) renderEditTemplateView() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Padding(0, 1)

	backButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		MarginRight(2)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Bold(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(0, 1)

	// Create back button and title
	backButton := backButtonStyle.Render("← Back")
	title := titleStyle.Render("Edit Template")
	
	headerLine := lipgloss.JoinHorizontal(lipgloss.Left, backButton, title)

	if m.templateForm == nil {
		return lipgloss.JoinVertical(lipgloss.Left, headerLine, "", "No form available")
	}

	// Build form fields
	var formFields []string

	// ID field (read-only in edit mode)
	idLabel := labelStyle.Render("ID (read-only):")
	idValue := ""
	if m.editMode && m.selectedTemplate != nil {
		idValue = m.selectedTemplate.ID
	} else {
		idValue = m.templateForm.inputs[templateIdField].View()
	}
	formFields = append(formFields, idLabel, idValue, "")

	// Version field
	versionLabel := labelStyle.Render("Version:")
	formFields = append(formFields, versionLabel, m.templateForm.inputs[templateVersionField].View(), "")

	// Name field
	nameLabel := labelStyle.Render("Name:")
	formFields = append(formFields, nameLabel, m.templateForm.inputs[templateNameField].View(), "")

	// Description field
	descLabel := labelStyle.Render("Description:")
	formFields = append(formFields, descLabel, m.templateForm.inputs[templateDescField].View(), "")

	// Slots field
	slotsLabel := labelStyle.Render("Slots:")
	formFields = append(formFields, slotsLabel, m.templateForm.inputs[templateSlotsField].View(), "")

	// Content field
	contentLabel := labelStyle.Render("Content:")
	formFields = append(formFields, contentLabel, m.templateForm.textarea.View(), "")

	// Help text
	help := helpStyle.Render("Tab/↓ next field • Shift+Tab/↑ prev field • Ctrl+S save • ←/esc/b cancel")

	// Join all elements
	allElements := []string{headerLine, ""}
	allElements = append(allElements, formFields...)
	allElements = append(allElements, help)

	return lipgloss.JoinVertical(lipgloss.Left, allElements...)
}

// renderTemplateDetailView renders template details
func (m Model) renderTemplateDetailView() string {
	if m.selectedTemplate == nil {
		return "No template selected"
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Padding(0, 1)

	backButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		MarginRight(2)

	metadataStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(0, 1)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(0, 1)

	// Create back button and title
	backButton := backButtonStyle.Render("← Back")
	title := titleStyle.Render(m.selectedTemplate.Name)
	
	headerLine := lipgloss.JoinHorizontal(lipgloss.Left, backButton, title)

	// Create metadata line
	metadata := fmt.Sprintf("ID: %s • Version: %s", m.selectedTemplate.ID, m.selectedTemplate.Version)
	metadataLine := metadataStyle.Render(metadata)

	// Help text
	help := helpStyle.Render("Press e to edit • ←/esc/b to go back")

	// Content (template preview)
	content := m.selectedTemplate.Content

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerLine,
		metadataLine,
		"",
		content,
		"",
		help,
	)
}

// renderTemplateManagementView renders template management menu using SelectForm
func (m Model) renderTemplateManagementView() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Padding(0, 1)

	backButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		MarginRight(2)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("33")).
		Bold(true).
		Padding(0, 1)

	unselectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Padding(0, 1)

	descriptionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		Padding(0, 3)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(0, 1)

	// Create back button and title
	backButton := backButtonStyle.Render("← Back")
	title := titleStyle.Render("Template Management")
	
	headerLine := lipgloss.JoinHorizontal(lipgloss.Left, backButton, title)

	if m.selectForm == nil {
		return lipgloss.JoinVertical(lipgloss.Left, headerLine, "", "No options available")
	}

	// Render options
	var optionLines []string
	for i, option := range m.selectForm.options {
		var style lipgloss.Style
		if i == m.selectForm.selected {
			style = selectedStyle
		} else {
			style = unselectedStyle
		}
		
		optionLine := style.Render("▶ " + option.Label)
		optionLines = append(optionLines, optionLine)
		
		if option.Description != "" {
			descLine := descriptionStyle.Render(option.Description)
			optionLines = append(optionLines, descLine)
		}
		
		optionLines = append(optionLines, "") // Add spacing
	}

	help := helpStyle.Render("↑/↓ or k/j to navigate • Enter to select • ←/esc/b to go back")

	// Join all elements
	allElements := []string{headerLine, ""}
	allElements = append(allElements, optionLines...)
	allElements = append(allElements, help)

	return lipgloss.JoinVertical(lipgloss.Left, allElements...)
}

// renderPreview renders the selected prompt for preview
func (m *Model) renderPreview() error {
	if m.selectedPrompt == nil {
		return fmt.Errorf("no prompt selected")
	}

	// Create a renderer for the prompt
	r := renderer.NewRenderer(m.selectedPrompt, nil)

	// Render with current variables
	rendered, err := r.RenderText(m.variables)
	if err != nil {
		// Show the raw content if rendering fails
		rendered = m.selectedPrompt.Content
	}

	// Also render as JSON for the 'y' copy option
	renderedJSON, err := r.RenderJSON(m.variables)
	if err != nil {
		renderedJSON = ""
	}

	// Format with glamour for display
	formatted, err := m.glamourRenderer.Render(rendered)
	if err != nil {
		formatted = rendered
	}

	m.renderedContent = rendered
	m.renderedContentJSON = renderedJSON
	m.viewport.SetContent(formatted)
	return nil
}