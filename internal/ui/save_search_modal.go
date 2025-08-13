package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dpshade/pocket-prompt/internal/models"
)

// SaveSearchModal provides a modal for saving boolean searches
type SaveSearchModal struct {
	nameInput   textinput.Model
	expression  *models.BooleanExpression
	isActive       bool
	width          int
	height         int
	submitted      bool
	savedSearch    *models.SavedSearch
	editMode       bool
	originalSearch *models.SavedSearch
}

// NewSaveSearchModal creates a new save search modal
func NewSaveSearchModal() *SaveSearchModal {
	nameInput := textinput.New()
	nameInput.Placeholder = "Enter search name"
	nameInput.Focus()
	nameInput.CharLimit = 50
	nameInput.Width = 50

	return &SaveSearchModal{
		nameInput: nameInput,
		isActive:  false,
	}
}

// SetExpression sets the boolean expression to save
func (m *SaveSearchModal) SetExpression(expr *models.BooleanExpression) {
	m.expression = expr
}

// Update handles input for the modal
func (m *SaveSearchModal) Update(msg tea.Msg) tea.Cmd {
	if !m.isActive {
		return nil
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			m.isActive = false
			m.submitted = false
			m.savedSearch = nil
			m.nameInput.SetValue("")
			return nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			name := m.nameInput.Value()
			if name != "" && m.expression != nil {
				// Create saved search
				m.savedSearch = &models.SavedSearch{
					Name:       name,
					Expression: m.expression,
				}
				m.submitted = true
				return nil
			}
		}

		// Update the name input
		m.nameInput, cmd = m.nameInput.Update(msg)
	}

	return cmd
}

// View renders the modal
func (m *SaveSearchModal) View() string {
	if !m.isActive {
		return ""
	}

	// Modal styles - use terminal default colors
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Width(60)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Bold(true)

	helpStyle := lipgloss.NewStyle().
		Italic(true).
		MarginTop(1)

	expressionStyle := lipgloss.NewStyle().
		Reverse(true).
		Padding(0, 1)

	var content []string

	// Title
	title := "💾 Save Boolean Search"
	if m.editMode {
		title = "✏️ Edit Boolean Search"
	}
	content = append(content, titleStyle.Render(title))
	content = append(content, "")

	// Show the expression being saved
	if m.expression != nil {
		content = append(content, "Expression: "+expressionStyle.Render(m.expression.String()))
		content = append(content, "")
	}

	// Name field
	content = append(content, labelStyle.Render("Name:"))
	content = append(content, m.nameInput.View())
	content = append(content, "")

	// Help
	helpText := "Enter: save • Esc: cancel"
	if m.editMode {
		helpText = "Enter: update • Esc: cancel"
	}
	content = append(content, helpStyle.Render(helpText))

	// Join content and apply modal styling
	modalContent := lipgloss.JoinVertical(lipgloss.Left, content...)
	return modalStyle.Render(modalContent)
}

// SetActive sets the modal active state
func (m *SaveSearchModal) SetActive(active bool) {
	m.isActive = active
	if active {
		m.submitted = false
		m.savedSearch = nil
		m.nameInput.Focus()
		if !m.editMode {
			m.nameInput.SetValue("")
		}
	}
}

// SetEditMode configures the modal for editing an existing search
func (m *SaveSearchModal) SetEditMode(savedSearch *models.SavedSearch, newExpression *models.BooleanExpression) {
	m.editMode = true
	m.originalSearch = savedSearch
	m.expression = newExpression
	m.nameInput.SetValue(savedSearch.Name)
}

// ClearEditMode clears edit mode
func (m *SaveSearchModal) ClearEditMode() {
	m.editMode = false
	m.originalSearch = nil
}

// IsEditMode returns whether the modal is in edit mode
func (m *SaveSearchModal) IsEditMode() bool {
	return m.editMode
}

// GetOriginalSearch returns the original search being edited
func (m *SaveSearchModal) GetOriginalSearch() *models.SavedSearch {
	return m.originalSearch
}

// IsActive returns whether the modal is active
func (m *SaveSearchModal) IsActive() bool {
	return m.isActive
}

// IsSubmitted returns whether the form was submitted
func (m *SaveSearchModal) IsSubmitted() bool {
	return m.submitted
}

// GetSavedSearch returns the created saved search
func (m *SaveSearchModal) GetSavedSearch() *models.SavedSearch {
	return m.savedSearch
}

// Resize updates the modal dimensions
func (m *SaveSearchModal) Resize(width, height int) {
	m.width = width
	m.height = height
	
	// Adjust input width based on modal size
	inputWidth := min(50, width-12)
	m.nameInput.Width = inputWidth
}