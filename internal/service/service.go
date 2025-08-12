package service

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/dylanshade/pocket-prompt/internal/models"
	"github.com/dylanshade/pocket-prompt/internal/storage"
	"github.com/sahilm/fuzzy"
)

// Service provides business logic for prompt management
type Service struct {
	storage *storage.Storage
	prompts []*models.Prompt // Cached prompts for fast access
}

// NewService creates a new service instance
func NewService() (*Service, error) {
	store, err := storage.NewStorage("")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	svc := &Service{
		storage: store,
	}

	// Load prompts into cache
	if err := svc.loadPrompts(); err != nil {
		// It's okay if we can't load prompts initially (library might not be initialized)
		// Just log the error
		fmt.Printf("Note: Could not load prompts (library may not be initialized): %v\n", err)
	}

	return svc, nil
}

// InitLibrary initializes a new prompt library
func (s *Service) InitLibrary() error {
	return s.storage.InitLibrary()
}

// loadPrompts loads all prompts into memory for fast access
func (s *Service) loadPrompts() error {
	prompts, err := s.storage.ListPrompts()
	if err != nil {
		return err
	}
	s.prompts = prompts
	return nil
}

// ListPrompts returns all prompts
func (s *Service) ListPrompts() ([]*models.Prompt, error) {
	if len(s.prompts) == 0 {
		if err := s.loadPrompts(); err != nil {
			return nil, err
		}
	}
	return s.prompts, nil
}

// SearchPrompts searches prompts by query string
func (s *Service) SearchPrompts(query string) ([]*models.Prompt, error) {
	prompts, err := s.ListPrompts()
	if err != nil {
		return nil, err
	}

	if query == "" {
		return prompts, nil
	}

	// Create searchable strings for each prompt
	var searchStrings []string
	for _, p := range prompts {
		searchStr := fmt.Sprintf("%s %s %s %s", 
			p.Name, 
			p.Summary, 
			p.ID,
			strings.Join(p.Tags, " "))
		searchStrings = append(searchStrings, searchStr)
	}

	// Perform fuzzy search
	matches := fuzzy.Find(query, searchStrings)
	
	// Build result list
	var results []*models.Prompt
	for _, match := range matches {
		results = append(results, prompts[match.Index])
	}

	return results, nil
}

// GetPrompt returns a prompt by ID
func (s *Service) GetPrompt(id string) (*models.Prompt, error) {
	prompts, err := s.ListPrompts()
	if err != nil {
		return nil, err
	}

	for _, p := range prompts {
		if p.ID == id {
			return p, nil
		}
	}

	return nil, fmt.Errorf("prompt not found: %s", id)
}

// CreatePrompt creates a new prompt
func (s *Service) CreatePrompt(prompt *models.Prompt) error {
	// Set timestamps
	now := time.Now()
	prompt.CreatedAt = now
	prompt.UpdatedAt = now

	// Generate file path if not set
	if prompt.FilePath == "" {
		prompt.FilePath = filepath.Join("prompts", fmt.Sprintf("%s.md", prompt.ID))
	}

	// Save to storage
	if err := s.storage.SavePrompt(prompt); err != nil {
		return err
	}

	// Reload prompts cache
	return s.loadPrompts()
}

// UpdatePrompt updates an existing prompt
func (s *Service) UpdatePrompt(prompt *models.Prompt) error {
	// Update timestamp
	prompt.UpdatedAt = time.Now()

	// Save to storage
	if err := s.storage.SavePrompt(prompt); err != nil {
		return err
	}

	// Reload prompts cache
	return s.loadPrompts()
}

// DeletePrompt deletes a prompt by ID
func (s *Service) DeletePrompt(id string) error {
	prompt, err := s.GetPrompt(id)
	if err != nil {
		return err
	}

	// Delete file
	// TODO: Implement file deletion in storage
	_ = prompt

	// Reload prompts cache
	return s.loadPrompts()
}

// FilterPromptsByTag returns prompts that have the specified tag
func (s *Service) FilterPromptsByTag(tag string) ([]*models.Prompt, error) {
	prompts, err := s.ListPrompts()
	if err != nil {
		return nil, err
	}

	var filtered []*models.Prompt
	for _, p := range prompts {
		for _, t := range p.Tags {
			if t == tag {
				filtered = append(filtered, p)
				break
			}
		}
	}

	return filtered, nil
}

// GetAllTags returns all unique tags from all prompts
func (s *Service) GetAllTags() ([]string, error) {
	prompts, err := s.ListPrompts()
	if err != nil {
		return nil, err
	}

	tagMap := make(map[string]bool)
	for _, p := range prompts {
		for _, tag := range p.Tags {
			tagMap[tag] = true
		}
	}

	var tags []string
	for tag := range tagMap {
		tags = append(tags, tag)
	}

	return tags, nil
}

// ListTemplates returns all available templates
func (s *Service) ListTemplates() ([]*models.Template, error) {
	return s.storage.ListTemplates()
}

// GetTemplate returns a template by ID
func (s *Service) GetTemplate(id string) (*models.Template, error) {
	templates, err := s.ListTemplates()
	if err != nil {
		return nil, err
	}

	for _, t := range templates {
		if t.ID == id {
			return t, nil
		}
	}

	return nil, fmt.Errorf("template not found: %s", id)
}

// SavePrompt saves a prompt (create or update)
func (s *Service) SavePrompt(prompt *models.Prompt) error {
	// Check if this is an existing prompt
	existing, err := s.GetPrompt(prompt.ID)
	if err == nil {
		// Update existing prompt
		prompt.CreatedAt = existing.CreatedAt // Keep original creation time
		prompt.UpdatedAt = time.Now()
		return s.UpdatePrompt(prompt)
	} else {
		// Create new prompt
		return s.CreatePrompt(prompt)
	}
}

// SaveTemplate saves a template (create or update)
func (s *Service) SaveTemplate(template *models.Template) error {
	// Set file path if not set
	if template.FilePath == "" {
		template.FilePath = filepath.Join("templates", fmt.Sprintf("%s.md", template.ID))
	}

	// Check if this is an existing template
	existing, err := s.GetTemplate(template.ID)
	if err == nil {
		// Update existing template
		template.CreatedAt = existing.CreatedAt // Keep original creation time
		template.UpdatedAt = time.Now()
	} else {
		// Create new template
		now := time.Now()
		template.CreatedAt = now
		template.UpdatedAt = now
	}

	// Save to storage
	return s.storage.SaveTemplate(template)
}