package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dylanshade/pocket-prompt/internal/models"
)

// CreateForm handles prompt creation
type CreateForm struct {
	inputs       []textinput.Model
	textarea     textarea.Model
	focused      int
	submitted    bool
	fromScratch  bool // True for simplified "from scratch" form
}

// Form field indices
const (
	idField = iota
	versionField
	titleField
	descriptionField
	tagsField
	variablesField
	templateRefField
	contentField
)

// NewCreateFormFromScratch creates a simplified empty form for starting from scratch
func NewCreateFormFromScratch() *CreateForm {
	inputs := make([]textinput.Model, 7) // Keep same size for compatibility

	// ID field - completely empty
	inputs[idField] = textinput.New()
	inputs[idField].Focus()
	inputs[idField].CharLimit = 50
	inputs[idField].Width = 40

	// Title field - completely empty
	inputs[titleField] = textinput.New()
	inputs[titleField].CharLimit = 100
	inputs[titleField].Width = 40

	// Initialize other fields as empty but unused
	for i := versionField; i <= templateRefField; i++ {
		if i != titleField { // Skip title field as it's already initialized
			inputs[i] = textinput.New()
		}
	}

	// Content textarea - completely empty
	ta := textarea.New()
	ta.SetWidth(80)
	ta.SetHeight(10)

	return &CreateForm{
		inputs:      inputs,
		textarea:    ta,
		focused:     0,
		fromScratch: true,
	}
}

// NewCreateForm creates a new prompt creation form with helpful placeholders
func NewCreateForm() *CreateForm {
	inputs := make([]textinput.Model, 7) // Increased from 4 to 7

	// ID field
	inputs[idField] = textinput.New()
	inputs[idField].Placeholder = "prompt-id"
	inputs[idField].Focus()
	inputs[idField].CharLimit = 50
	inputs[idField].Width = 40

	// Version field
	inputs[versionField] = textinput.New()
	inputs[versionField].Placeholder = "1.0.0"
	inputs[versionField].CharLimit = 20
	inputs[versionField].Width = 20

	// Title field
	inputs[titleField] = textinput.New()
	inputs[titleField].Placeholder = "Prompt Title"
	inputs[titleField].CharLimit = 100
	inputs[titleField].Width = 40

	// Description field
	inputs[descriptionField] = textinput.New()
	inputs[descriptionField].Placeholder = "Brief description of the prompt"
	inputs[descriptionField].CharLimit = 255
	inputs[descriptionField].Width = 60

	// Tags field
	inputs[tagsField] = textinput.New()
	inputs[tagsField].Placeholder = "tag1, tag2, tag3"
	inputs[tagsField].CharLimit = 200
	inputs[tagsField].Width = 40

	// Variables field
	inputs[variablesField] = textinput.New()
	inputs[variablesField].Placeholder = "name:type:required:default, ..."
	inputs[variablesField].CharLimit = 500
	inputs[variablesField].Width = 60

	// Template reference field
	inputs[templateRefField] = textinput.New()
	inputs[templateRefField].Placeholder = "template-id (optional)"
	inputs[templateRefField].CharLimit = 100
	inputs[templateRefField].Width = 40

	// Content textarea
	ta := textarea.New()
	ta.Placeholder = "Enter your prompt content here..."
	ta.SetWidth(80)
	ta.SetHeight(10)

	return &CreateForm{
		inputs:   inputs,
		textarea: ta,
		focused:  0,
	}
}

// Update handles form updates
func (f *CreateForm) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			f.nextField()
		case "shift+tab", "up":
			f.prevField()
		case "enter":
			if f.focused == contentField {
				// Let textarea handle enter
				var cmd tea.Cmd
				f.textarea, cmd = f.textarea.Update(msg)
				return cmd
			} else {
				f.nextField()
			}
		case "ctrl+s":
			f.submitted = true
			return nil
		}
	}

	// Update the focused field
	if f.focused == contentField {
		var cmd tea.Cmd
		f.textarea, cmd = f.textarea.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		var cmd tea.Cmd
		f.inputs[f.focused], cmd = f.inputs[f.focused].Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// nextField moves to the next form field
func (f *CreateForm) nextField() {
	if f.focused == contentField {
		f.textarea.Blur()
	} else {
		f.inputs[f.focused].Blur()
	}
	
	if f.fromScratch {
		// Simplified navigation: ID -> Title -> Content
		switch f.focused {
		case idField:
			f.focused = titleField
		case titleField:
			f.focused = contentField
		case contentField:
			f.focused = idField
		default:
			f.focused = idField
		}
	} else {
		// Full form navigation
		f.focused++
		if f.focused >= len(f.inputs)+1 { // +1 for textarea
			f.focused = 0
		}
	}
	
	if f.focused == contentField {
		f.textarea.Focus()
	} else {
		f.inputs[f.focused].Focus()
	}
}

// prevField moves to the previous form field
func (f *CreateForm) prevField() {
	if f.focused == contentField {
		f.textarea.Blur()
	} else {
		f.inputs[f.focused].Blur()
	}
	
	if f.fromScratch {
		// Simplified navigation: Content -> Title -> ID
		switch f.focused {
		case idField:
			f.focused = contentField
		case titleField:
			f.focused = idField
		case contentField:
			f.focused = titleField
		default:
			f.focused = idField
		}
	} else {
		// Full form navigation
		f.focused--
		if f.focused < 0 {
			f.focused = len(f.inputs) // Points to textarea
		}
	}
	
	if f.focused == contentField {
		f.textarea.Focus()
	} else {
		f.inputs[f.focused].Focus()
	}
}

// ToPrompt converts form data to a Prompt model
func (f *CreateForm) ToPrompt() *models.Prompt {
	now := time.Now()
	
	if f.fromScratch {
		// Simplified form: only use ID, Title, and Content
		return &models.Prompt{
			ID:        f.inputs[idField].Value(),
			Version:   "1.0.0", // Default version for scratch forms
			Name:      f.inputs[titleField].Value(),
			Summary:   "", // Empty summary for scratch forms
			Tags:      []string{}, // No tags for scratch forms
			Variables: []models.Variable{}, // No variables for scratch forms
			TemplateRef: "", // No template reference for scratch forms
			CreatedAt: now,
			UpdatedAt: now,
			Content:   f.textarea.Value(),
		}
	}
	
	// Full form processing
	tags := []string{}
	if f.inputs[tagsField].Value() != "" {
		tagList := strings.Split(f.inputs[tagsField].Value(), ",")
		for _, tag := range tagList {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	// Parse variables from the variables field
	variables := []models.Variable{}
	if f.inputs[variablesField].Value() != "" {
		varList := strings.Split(f.inputs[variablesField].Value(), ",")
		for _, varStr := range varList {
			parts := strings.Split(strings.TrimSpace(varStr), ":")
			if len(parts) >= 2 {
				variable := models.Variable{
					Name: strings.TrimSpace(parts[0]),
					Type: strings.TrimSpace(parts[1]),
				}
				if len(parts) >= 3 {
					variable.Required = strings.TrimSpace(parts[2]) == "true"
				}
				if len(parts) >= 4 {
					variable.Default = strings.TrimSpace(parts[3])
				}
				variables = append(variables, variable)
			}
		}
	}

	// Get version as entered by user (no default)
	version := f.inputs[versionField].Value()

	return &models.Prompt{
		ID:          f.inputs[idField].Value(),
		Version:     version,
		Name:        f.inputs[titleField].Value(),
		Summary:     f.inputs[descriptionField].Value(),
		Tags:        tags,
		Variables:   variables,
		TemplateRef: f.inputs[templateRefField].Value(),
		CreatedAt:   now,
		UpdatedAt:   now,
		Content:     f.textarea.Value(),
	}
}

// IsSubmitted returns whether the form has been submitted
func (f *CreateForm) IsSubmitted() bool {
	return f.submitted
}

// Reset resets the form
func (f *CreateForm) Reset() {
	for i := range f.inputs {
		f.inputs[i].SetValue("")
	}
	f.textarea.SetValue("")
	f.focused = 0
	f.submitted = false
	f.inputs[0].Focus()
}

// LoadPrompt loads an existing prompt into the form for editing
func (f *CreateForm) LoadPrompt(prompt *models.Prompt) {
	f.inputs[idField].SetValue(prompt.ID)
	f.inputs[versionField].SetValue(prompt.Version)
	f.inputs[titleField].SetValue(prompt.Name)
	f.inputs[descriptionField].SetValue(prompt.Summary)
	
	// Convert tags slice to comma-separated string
	tags := ""
	for i, tag := range prompt.Tags {
		if i > 0 {
			tags += ", "
		}
		tags += tag
	}
	f.inputs[tagsField].SetValue(tags)
	
	// Convert variables to string format
	variables := ""
	for i, variable := range prompt.Variables {
		if i > 0 {
			variables += ", "
		}
		variables += variable.Name + ":" + variable.Type
		if variable.Required {
			variables += ":true"
		} else {
			variables += ":false"
		}
		if variable.Default != nil {
			variables += ":" + fmt.Sprintf("%v", variable.Default)
		}
	}
	f.inputs[variablesField].SetValue(variables)
	
	f.inputs[templateRefField].SetValue(prompt.TemplateRef)
	f.textarea.SetValue(prompt.Content)
}

// TemplateForm handles template creation and editing
type TemplateForm struct {
	inputs    []textinput.Model
	textarea  textarea.Model
	focused   int
	submitted bool
}

// Template form field indices
const (
	templateIdField = iota
	templateVersionField
	templateNameField
	templateDescField
	templateSlotsField
	templateContentField
)

// NewTemplateFormFromScratch creates a completely empty template form
func NewTemplateFormFromScratch() *TemplateForm {
	inputs := make([]textinput.Model, 5)

	// ID field - completely empty
	inputs[templateIdField] = textinput.New()
	inputs[templateIdField].Focus()
	inputs[templateIdField].CharLimit = 50
	inputs[templateIdField].Width = 40

	// Version field - completely empty
	inputs[templateVersionField] = textinput.New()
	inputs[templateVersionField].CharLimit = 20
	inputs[templateVersionField].Width = 20

	// Name field - completely empty
	inputs[templateNameField] = textinput.New()
	inputs[templateNameField].CharLimit = 100
	inputs[templateNameField].Width = 40

	// Description field - completely empty
	inputs[templateDescField] = textinput.New()
	inputs[templateDescField].CharLimit = 255
	inputs[templateDescField].Width = 60

	// Slots field - completely empty
	inputs[templateSlotsField] = textinput.New()
	inputs[templateSlotsField].CharLimit = 500
	inputs[templateSlotsField].Width = 60

	// Content textarea - completely empty
	ta := textarea.New()
	ta.SetWidth(80)
	ta.SetHeight(15)

	return &TemplateForm{
		inputs:   inputs,
		textarea: ta,
		focused:  0,
	}
}

// NewTemplateForm creates a new template form with helpful placeholders
func NewTemplateForm() *TemplateForm {
	inputs := make([]textinput.Model, 5) // Increased from 3 to 5

	// ID field
	inputs[templateIdField] = textinput.New()
	inputs[templateIdField].Placeholder = "template-id"
	inputs[templateIdField].Focus()
	inputs[templateIdField].CharLimit = 50
	inputs[templateIdField].Width = 40

	// Version field
	inputs[templateVersionField] = textinput.New()
	inputs[templateVersionField].Placeholder = "1.0.0"
	inputs[templateVersionField].CharLimit = 20
	inputs[templateVersionField].Width = 20

	// Name field
	inputs[templateNameField] = textinput.New()
	inputs[templateNameField].Placeholder = "Template Name"
	inputs[templateNameField].CharLimit = 100
	inputs[templateNameField].Width = 40

	// Description field
	inputs[templateDescField] = textinput.New()
	inputs[templateDescField].Placeholder = "Brief description of the template"
	inputs[templateDescField].CharLimit = 255
	inputs[templateDescField].Width = 60

	// Slots field
	inputs[templateSlotsField] = textinput.New()
	inputs[templateSlotsField].Placeholder = "name:description:required:default, ..."
	inputs[templateSlotsField].CharLimit = 500
	inputs[templateSlotsField].Width = 60

	// Content textarea
	ta := textarea.New()
	ta.Placeholder = "Enter template content with {{slots}}..."
	ta.SetWidth(80)
	ta.SetHeight(15)

	return &TemplateForm{
		inputs:   inputs,
		textarea: ta,
		focused:  0,
	}
}

// Update handles template form updates
func (f *TemplateForm) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			f.nextField()
		case "shift+tab", "up":
			f.prevField()
		case "enter":
			if f.focused == templateContentField {
				var cmd tea.Cmd
				f.textarea, cmd = f.textarea.Update(msg)
				return cmd
			} else {
				f.nextField()
			}
		case "ctrl+s":
			f.submitted = true
			return nil
		}
	}

	// Update the focused field
	if f.focused == templateContentField {
		var cmd tea.Cmd
		f.textarea, cmd = f.textarea.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		var cmd tea.Cmd
		f.inputs[f.focused], cmd = f.inputs[f.focused].Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// nextField moves to the next form field
func (f *TemplateForm) nextField() {
	if f.focused == templateContentField {
		f.textarea.Blur()
	} else {
		f.inputs[f.focused].Blur()
	}
	f.focused++
	if f.focused >= len(f.inputs)+1 { // +1 for textarea
		f.focused = 0
	}
	if f.focused == templateContentField {
		f.textarea.Focus()
	} else {
		f.inputs[f.focused].Focus()
	}
}

// prevField moves to the previous form field
func (f *TemplateForm) prevField() {
	if f.focused == templateContentField {
		f.textarea.Blur()
	} else {
		f.inputs[f.focused].Blur()
	}
	f.focused--
	if f.focused < 0 {
		f.focused = len(f.inputs) // Points to textarea
	}
	if f.focused == templateContentField {
		f.textarea.Focus()
	} else {
		f.inputs[f.focused].Focus()
	}
}

// ToTemplate converts form data to a Template model
func (f *TemplateForm) ToTemplate() *models.Template {
	// Parse slots from the slots field
	slots := []models.Slot{}
	if f.inputs[templateSlotsField].Value() != "" {
		slotList := strings.Split(f.inputs[templateSlotsField].Value(), ",")
		for _, slotStr := range slotList {
			parts := strings.Split(strings.TrimSpace(slotStr), ":")
			if len(parts) >= 1 {
				slot := models.Slot{
					Name: strings.TrimSpace(parts[0]),
				}
				if len(parts) >= 2 {
					slot.Description = strings.TrimSpace(parts[1])
				}
				if len(parts) >= 3 {
					slot.Required = strings.TrimSpace(parts[2]) == "true"
				}
				if len(parts) >= 4 {
					slot.Default = strings.TrimSpace(parts[3])
				}
				slots = append(slots, slot)
			}
		}
	}

	// Get version as entered by user (no default)
	version := f.inputs[templateVersionField].Value()

	now := time.Now()
	return &models.Template{
		ID:          f.inputs[templateIdField].Value(),
		Version:     version,
		Name:        f.inputs[templateNameField].Value(),
		Description: f.inputs[templateDescField].Value(),
		Slots:       slots,
		CreatedAt:   now,
		UpdatedAt:   now,
		Content:     f.textarea.Value(),
	}
}

// LoadTemplate loads an existing template into the form for editing
func (f *TemplateForm) LoadTemplate(template *models.Template) {
	f.inputs[templateIdField].SetValue(template.ID)
	f.inputs[templateVersionField].SetValue(template.Version)
	f.inputs[templateNameField].SetValue(template.Name)
	f.inputs[templateDescField].SetValue(template.Description)
	
	// Convert slots to string format
	slots := ""
	for i, slot := range template.Slots {
		if i > 0 {
			slots += ", "
		}
		slots += slot.Name
		if slot.Description != "" {
			slots += ":" + slot.Description
		} else {
			slots += ":"
		}
		if slot.Required {
			slots += ":true"
		} else {
			slots += ":false"
		}
		if slot.Default != "" {
			slots += ":" + slot.Default
		}
	}
	f.inputs[templateSlotsField].SetValue(slots)
	
	f.textarea.SetValue(template.Content)
}

// IsSubmitted returns whether the form has been submitted
func (f *TemplateForm) IsSubmitted() bool {
	return f.submitted
}

// Reset resets the template form
func (f *TemplateForm) Reset() {
	for i := range f.inputs {
		f.inputs[i].SetValue("")
	}
	f.textarea.SetValue("")
	f.focused = 0
	f.submitted = false
	f.inputs[0].Focus()
}

// SelectForm handles selection from a list of options
type SelectForm struct {
	options   []SelectOption
	selected  int
	submitted bool
}

// SelectOption represents an option in the select form
type SelectOption struct {
	Label       string
	Description string
	Value       interface{}
}

// NewSelectForm creates a new select form
func NewSelectForm(options []SelectOption) *SelectForm {
	return &SelectForm{
		options:  options,
		selected: 0,
	}
}

// Update handles select form updates
func (f *SelectForm) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if f.selected > 0 {
				f.selected--
			}
		case "down", "j":
			if f.selected < len(f.options)-1 {
				f.selected++
			}
		case "enter":
			f.submitted = true
			return nil
		}
	}
	return nil
}

// GetSelected returns the selected option
func (f *SelectForm) GetSelected() *SelectOption {
	if f.selected >= 0 && f.selected < len(f.options) {
		return &f.options[f.selected]
	}
	return nil
}

// IsSubmitted returns whether an option has been selected
func (f *SelectForm) IsSubmitted() bool {
	return f.submitted
}

// Reset resets the select form
func (f *SelectForm) Reset() {
	f.selected = 0
	f.submitted = false
}