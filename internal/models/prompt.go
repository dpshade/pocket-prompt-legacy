package models

import (
	"time"
)

// Prompt represents a prompt artifact with YAML frontmatter and markdown content
type Prompt struct {
	// Frontmatter fields
	ID           string                 `yaml:"id"`
	Version      string                 `yaml:"version"`
	Name         string                 `yaml:"title"`
	Summary      string                 `yaml:"description"`
	Tags         []string               `yaml:"tags"`
	Variables    []Variable             `yaml:"variables"`
	TemplateRef  string                 `yaml:"template,omitempty"`
	Metadata     map[string]interface{} `yaml:"metadata,omitempty"`
	CreatedAt    time.Time              `yaml:"created_at"`
	UpdatedAt    time.Time              `yaml:"updated_at"`

	// Content fields
	Content     string `yaml:"-"` // The markdown content after frontmatter
	FilePath    string `yaml:"-"` // Path to the file
	ContentHash string `yaml:"-"` // SHA256 hash of the content
}

// Variable represents a template variable with type and default value
type Variable struct {
	Name        string      `yaml:"name"`
	Type        string      `yaml:"type"` // string, number, boolean, list
	Description string      `yaml:"description,omitempty"`
	Default     interface{} `yaml:"default,omitempty"`
	Required    bool        `yaml:"required"`
	Options     []string    `yaml:"options,omitempty"` // For enum-like variables
}

// Implement list.Item interface for bubbles list component

// FilterValue returns the value used for filtering in lists
func (p Prompt) FilterValue() string {
	return p.Name
}

// Title satisfies the list.Item interface
func (p Prompt) Title() string {
	if p.Name != "" {
		return p.Name
	}
	return p.ID
}

// Description satisfies the list.Item interface  
func (p Prompt) Description() string {
	if p.Summary != "" {
		return p.Summary
	}
	if len(p.Tags) > 0 {
		return "Tags: " + joinTags(p.Tags)
	}
	return ""
}

func joinTags(tags []string) string {
	result := ""
	for i, tag := range tags {
		if i > 0 {
			result += ", "
		}
		result += tag
	}
	return result
}