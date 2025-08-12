package storage

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dylanshade/pocket-prompt/internal/models"
	"gopkg.in/yaml.v3"
)

// Storage handles all file system operations for prompts, templates, and packs
type Storage struct {
	rootPath string
}

// NewStorage creates a new storage instance
func NewStorage(rootPath string) (*Storage, error) {
	if rootPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		rootPath = filepath.Join(homeDir, ".pocket-prompt")
	}

	return &Storage{rootPath: rootPath}, nil
}

// InitLibrary creates the directory structure for a prompt library
func (s *Storage) InitLibrary() error {
	dirs := []string{
		s.rootPath,
		filepath.Join(s.rootPath, "prompts"),
		filepath.Join(s.rootPath, "templates"),
		filepath.Join(s.rootPath, "packs"),
		filepath.Join(s.rootPath, ".pocket-prompt"),
		filepath.Join(s.rootPath, ".pocket-prompt", "cache"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// LoadPrompt loads a prompt from a markdown file with YAML frontmatter
func (s *Storage) LoadPrompt(path string) (*models.Prompt, error) {
	fullPath := filepath.Join(s.rootPath, path)
	
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open prompt file: %w", err)
	}
	defer file.Close()

	// Read the entire file
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read prompt file: %w", err)
	}

	// Parse frontmatter and content
	prompt, err := parsePromptFile(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse prompt: %w", err)
	}

	prompt.FilePath = path
	prompt.ContentHash = calculateHash(content)

	return prompt, nil
}

// SavePrompt saves a prompt to a markdown file with YAML frontmatter
func (s *Storage) SavePrompt(prompt *models.Prompt) error {
	fullPath := filepath.Join(s.rootPath, prompt.FilePath)
	
	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Serialize prompt to YAML frontmatter + markdown
	content, err := serializePrompt(prompt)
	if err != nil {
		return fmt.Errorf("failed to serialize prompt: %w", err)
	}

	// Write to file
	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write prompt file: %w", err)
	}

	return nil
}

// DeletePrompt deletes a prompt file from the file system
func (s *Storage) DeletePrompt(prompt *models.Prompt) error {
	fullPath := filepath.Join(s.rootPath, prompt.FilePath)
	
	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("prompt file does not exist: %s", fullPath)
	}
	
	// Delete the file
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete prompt file: %w", err)
	}
	
	return nil
}

// SaveTemplate saves a template to the file system
func (s *Storage) SaveTemplate(template *models.Template) error {
	fullPath := filepath.Join(s.rootPath, template.FilePath)
	
	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Serialize template to YAML frontmatter + markdown
	content, err := serializeTemplate(template)
	if err != nil {
		return fmt.Errorf("failed to serialize template: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}
	
	return nil
}

// ListPrompts returns all prompts in the library
func (s *Storage) ListPrompts() ([]*models.Prompt, error) {
	promptsDir := filepath.Join(s.rootPath, "prompts")
	
	var prompts []*models.Prompt
	err := filepath.Walk(promptsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			relPath, _ := filepath.Rel(s.rootPath, path)
			prompt, err := s.LoadPrompt(relPath)
			if err != nil {
				// Log error but continue walking
				fmt.Fprintf(os.Stderr, "Warning: failed to load prompt %s: %v\n", relPath, err)
				return nil
			}
			prompts = append(prompts, prompt)
		}

		return nil
	})

	return prompts, err
}

// LoadTemplate loads a template from a markdown file
func (s *Storage) LoadTemplate(path string) (*models.Template, error) {
	fullPath := filepath.Join(s.rootPath, path)
	
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open template file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	template, err := parseTemplateFile(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	template.FilePath = path
	return template, nil
}

// ListTemplates returns all templates in the library
func (s *Storage) ListTemplates() ([]*models.Template, error) {
	templatesDir := filepath.Join(s.rootPath, "templates")
	
	var templates []*models.Template
	err := filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			relPath, _ := filepath.Rel(s.rootPath, path)
			template, err := s.LoadTemplate(relPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to load template %s: %v\n", relPath, err)
				return nil
			}
			templates = append(templates, template)
		}

		return nil
	})

	return templates, err
}

// Helper functions

func parsePromptFile(content []byte) (*models.Prompt, error) {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	
	// Check for frontmatter delimiter
	if !scanner.Scan() || scanner.Text() != "---" {
		return nil, fmt.Errorf("missing frontmatter delimiter")
	}

	// Read frontmatter
	var frontmatterLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			break
		}
		frontmatterLines = append(frontmatterLines, line)
	}

	// Parse YAML frontmatter
	frontmatter := strings.Join(frontmatterLines, "\n")
	var prompt models.Prompt
	if err := yaml.Unmarshal([]byte(frontmatter), &prompt); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Read remaining content
	var contentLines []string
	for scanner.Scan() {
		contentLines = append(contentLines, scanner.Text())
	}
	// Join content and trim leading/trailing whitespace
	prompt.Content = strings.TrimSpace(strings.Join(contentLines, "\n"))

	return &prompt, nil
}

func parseTemplateFile(content []byte) (*models.Template, error) {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	
	// Check for frontmatter delimiter
	if !scanner.Scan() || scanner.Text() != "---" {
		return nil, fmt.Errorf("missing frontmatter delimiter")
	}

	// Read frontmatter
	var frontmatterLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			break
		}
		frontmatterLines = append(frontmatterLines, line)
	}

	// Parse YAML frontmatter
	frontmatter := strings.Join(frontmatterLines, "\n")
	var template models.Template
	if err := yaml.Unmarshal([]byte(frontmatter), &template); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Read remaining content
	var contentLines []string
	for scanner.Scan() {
		contentLines = append(contentLines, scanner.Text())
	}
	// Join content and trim leading/trailing whitespace
	template.Content = strings.TrimSpace(strings.Join(contentLines, "\n"))

	return &template, nil
}

func serializePrompt(prompt *models.Prompt) ([]byte, error) {
	var buf bytes.Buffer

	// Write frontmatter delimiter
	buf.WriteString("---\n")

	// Serialize prompt metadata to YAML
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(prompt); err != nil {
		return nil, fmt.Errorf("failed to encode frontmatter: %w", err)
	}

	// Write closing delimiter
	buf.WriteString("---\n")

	// Write content with proper spacing
	if prompt.Content != "" {
		buf.WriteString("\n")
		buf.WriteString(strings.TrimSpace(prompt.Content))
	}

	return buf.Bytes(), nil
}

// serializeTemplate converts a template to YAML frontmatter + markdown content
func serializeTemplate(template *models.Template) ([]byte, error) {
	var buf bytes.Buffer

	// Write frontmatter delimiter
	buf.WriteString("---\n")

	// Serialize template metadata to YAML
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(template); err != nil {
		return nil, fmt.Errorf("failed to encode frontmatter: %w", err)
	}

	// Write closing delimiter
	buf.WriteString("---\n")

	// Write content with proper spacing
	if template.Content != "" {
		buf.WriteString("\n")
		buf.WriteString(strings.TrimSpace(template.Content))
	}

	return buf.Bytes(), nil
}

func calculateHash(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}