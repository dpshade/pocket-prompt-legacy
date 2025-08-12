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
	TemplateRef  string                 `yaml:"template,omitempty"`
	Metadata     map[string]interface{} `yaml:"metadata,omitempty"`
	CreatedAt    time.Time              `yaml:"created_at"`
	UpdatedAt    time.Time              `yaml:"updated_at"`

	// Content fields
	Content     string `yaml:"-"` // The markdown content after frontmatter
	FilePath    string `yaml:"-"` // Path to the file
	ContentHash string `yaml:"-"` // SHA256 hash of the content
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
	desc := ""
	if p.Summary != "" {
		desc = p.Summary
	}
	
	// Add last edited info
	if !p.UpdatedAt.IsZero() {
		lastEdited := " • Last edited: " + p.UpdatedAt.Format("2006-01-02 15:04")
		if desc == "" {
			desc = lastEdited[3:] // Remove the " • " prefix when it's the first item
		} else {
			desc += lastEdited
		}
	}
	
	// Add tags if available
	if len(p.Tags) > 0 {
		tagsStr := " • Tags: " + joinTags(p.Tags)
		if desc == "" {
			desc = tagsStr[3:] // Remove the " • " prefix when it's the first item
		} else {
			desc += tagsStr
		}
	}
	
	return desc
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