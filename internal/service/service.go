package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
	// Check for custom directory from environment
	rootPath := os.Getenv("POCKET_PROMPT_DIR")
	store, err := storage.NewStorage(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	svc := &Service{
		storage: store,
	}

	// NOTE: Removed eager loading for faster startup
	// Prompts will be loaded on-demand or asynchronously

	return svc, nil
}

// LoadPromptsAsync loads prompts asynchronously and returns a function to check completion
func (s *Service) LoadPromptsAsync() func() ([]*models.Prompt, bool, error) {
	resultChan := make(chan struct {
		prompts []*models.Prompt
		err     error
	}, 1)

	go func() {
		prompts, err := s.storage.ListPrompts()
		if err == nil {
			s.prompts = prompts
		}
		resultChan <- struct {
			prompts []*models.Prompt
			err     error
		}{prompts, err}
	}()

	return func() ([]*models.Prompt, bool, error) {
		select {
		case result := <-resultChan:
			return result.prompts, true, result.err // completed
		default:
			return nil, false, nil // still loading
		}
	}
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

// ListPrompts returns all non-archived prompts
func (s *Service) ListPrompts() ([]*models.Prompt, error) {
	if len(s.prompts) == 0 {
		if err := s.loadPrompts(); err != nil {
			return nil, err
		}
	}
	
	// Filter out archived prompts
	var activePrompts []*models.Prompt
	for _, prompt := range s.prompts {
		if !s.isArchived(prompt) {
			activePrompts = append(activePrompts, prompt)
		}
	}
	return activePrompts, nil
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

// UpdatePrompt updates an existing prompt with version management
func (s *Service) UpdatePrompt(prompt *models.Prompt) error {
	// Get the existing prompt to check current version
	existing, err := s.GetPrompt(prompt.ID)
	if err != nil {
		return fmt.Errorf("cannot update non-existent prompt: %w", err)
	}

	// Archive the old version by adding 'archive' tag and saving it
	if err := s.archivePromptByTag(existing); err != nil {
		return fmt.Errorf("failed to archive old version: %w", err)
	}

	// Increment version
	newVersion, err := s.incrementVersion(existing.Version)
	if err != nil {
		return fmt.Errorf("failed to increment version: %w", err)
	}
	prompt.Version = newVersion

	// Update timestamp but keep original creation time and file path
	prompt.CreatedAt = existing.CreatedAt
	prompt.UpdatedAt = time.Now()
	if prompt.FilePath == "" {
		prompt.FilePath = existing.FilePath // Keep original file path
	}

	// Save the new version (without archive tag)
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

	// Delete the file from storage
	if err := s.storage.DeletePrompt(prompt); err != nil {
		return fmt.Errorf("failed to delete prompt file: %w", err)
	}

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

// archivePromptByTag archives a prompt by adding the 'archive' tag and updating filename
func (s *Service) archivePromptByTag(prompt *models.Prompt) error {
	// Create a copy of the prompt for archiving
	archivedPrompt := *prompt
	
	// Add 'archive' tag if not already present
	hasArchiveTag := false
	for _, tag := range archivedPrompt.Tags {
		if tag == "archive" {
			hasArchiveTag = true
			break
		}
	}
	if !hasArchiveTag {
		archivedPrompt.Tags = append(archivedPrompt.Tags, "archive")
	}
	
	// Update filename to include version for archived copy
	archiveFilename := fmt.Sprintf("%s-v%s.md", prompt.ID, prompt.Version)
	archivedPrompt.FilePath = filepath.Join("prompts", archiveFilename)
	
	// Save the archived version
	return s.storage.SavePrompt(&archivedPrompt)
}

// incrementVersion increments a semantic version string
func (s *Service) incrementVersion(currentVersion string) (string, error) {
	if currentVersion == "" {
		return "1.0.0", nil
	}
	
	// Parse semantic version (e.g., "1.2.3")
	parts := strings.Split(currentVersion, ".")
	if len(parts) != 3 {
		// If not semantic version, treat as simple increment
		if version, err := strconv.Atoi(currentVersion); err == nil {
			return strconv.Itoa(version + 1), nil
		}
		return currentVersion + ".1", nil
	}
	
	// Increment patch version (third number)
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return currentVersion + ".1", nil
	}
	
	return fmt.Sprintf("%s.%s.%d", parts[0], parts[1], patch+1), nil
}

// isArchived checks if a prompt has the 'archive' tag
func (s *Service) isArchived(prompt *models.Prompt) bool {
	for _, tag := range prompt.Tags {
		if tag == "archive" {
			return true
		}
	}
	return false
}

// ListArchivedPrompts returns only archived prompts
func (s *Service) ListArchivedPrompts() ([]*models.Prompt, error) {
	if len(s.prompts) == 0 {
		if err := s.loadPrompts(); err != nil {
			return nil, err
		}
	}
	
	// Filter for archived prompts only
	var archivedPrompts []*models.Prompt
	for _, prompt := range s.prompts {
		if s.isArchived(prompt) {
			archivedPrompts = append(archivedPrompts, prompt)
		}
	}
	return archivedPrompts, nil
}