package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dpshade/pocket-prompt/internal/clipboard"
	"github.com/dpshade/pocket-prompt/internal/models"
	"github.com/dpshade/pocket-prompt/internal/renderer"
	"github.com/dpshade/pocket-prompt/internal/service"
)

// URLServer provides HTTP endpoints for iOS Shortcuts integration
type URLServer struct {
	service *service.Service
	port    int
}

// NewURLServer creates a new URL server instance
func NewURLServer(svc *service.Service, port int) *URLServer {
	return &URLServer{
		service: svc,
		port:    port,
	}
}

// Start begins serving HTTP requests
func (s *URLServer) Start() error {
	http.HandleFunc("/pocket-prompt/", s.handlePocketPrompt)
	http.HandleFunc("/health", s.handleHealth)
	
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("URL server starting on http://localhost%s", addr)
	log.Printf("iOS Shortcuts can now call URLs like:")
	log.Printf("  http://localhost%s/pocket-prompt/render/my-prompt-id", addr)
	log.Printf("  http://localhost%s/pocket-prompt/search?q=AI", addr)
	log.Printf("  http://localhost%s/pocket-prompt/boolean?expr=ai+AND+analysis", addr)
	
	return http.ListenAndServe(addr, nil)
}

// handleHealth provides a simple health check endpoint
func (s *URLServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"service": "pocket-prompt-url-server",
	})
}

// handlePocketPrompt routes pocket-prompt URL requests
func (s *URLServer) handlePocketPrompt(w http.ResponseWriter, r *http.Request) {
	// Enable CORS for cross-origin requests
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	
	if r.Method == "OPTIONS" {
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/pocket-prompt/")
	parts := strings.Split(path, "/")
	
	if len(parts) == 0 || parts[0] == "" {
		s.writeError(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	operation := parts[0]
	
	switch operation {
	case "render":
		s.handleRender(w, r, parts[1:])
	case "get":
		s.handleGet(w, r, parts[1:])
	case "list":
		s.handleList(w, r)
	case "search":
		s.handleSearch(w, r)
	case "boolean":
		s.handleBooleanSearch(w, r)
	case "saved-search":
		s.handleSavedSearch(w, r, parts[1:])
	case "saved-searches":
		s.handleSavedSearches(w, r, parts[1:])
	case "tags":
		s.handleTags(w, r)
	case "tag":
		s.handleTag(w, r, parts[1:])
	case "templates":
		s.handleTemplates(w, r)
	case "template":
		s.handleTemplate(w, r, parts[1:])
	default:
		s.writeError(w, fmt.Sprintf("Unknown operation: %s", operation), http.StatusNotFound)
	}
}

// handleRender renders a prompt with variables
func (s *URLServer) handleRender(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 0 {
		s.writeError(w, "Render requires a prompt ID", http.StatusBadRequest)
		return
	}

	promptID := parts[0]
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "text"
	}

	// Get prompt
	prompt, err := s.service.GetPrompt(promptID)
	if err != nil {
		s.writeError(w, fmt.Sprintf("Failed to get prompt: %v", err), http.StatusNotFound)
		return
	}

	// Parse variables from query parameters
	variables := make(map[string]interface{})
	for key, values := range r.URL.Query() {
		if key != "format" && len(values) > 0 {
			// Try to parse as number, fallback to string
			if num, err := strconv.ParseFloat(values[0], 64); err == nil {
				variables[key] = num
			} else if values[0] == "true" || values[0] == "false" {
				variables[key] = values[0] == "true"
			} else {
				variables[key] = values[0]
			}
		}
	}

	// Get template if referenced
	var template *models.Template
	if prompt.TemplateRef != "" {
		template, _ = s.service.GetTemplate(prompt.TemplateRef)
	}

	// Render prompt
	renderer := renderer.NewRenderer(prompt, template)
	
	var content string
	switch format {
	case "json":
		content, err = renderer.RenderJSON(variables)
	default:
		content, err = renderer.RenderText(variables)
	}

	if err != nil {
		s.writeError(w, fmt.Sprintf("Failed to render prompt: %v", err), http.StatusInternalServerError)
		return
	}

	s.writeToClipboardAndRespond(w, content, fmt.Sprintf("Rendered prompt: %s", promptID))
}

// handleGet retrieves a specific prompt
func (s *URLServer) handleGet(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 0 {
		s.writeError(w, "Get requires a prompt ID", http.StatusBadRequest)
		return
	}

	promptID := parts[0]
	format := r.URL.Query().Get("format")

	prompt, err := s.service.GetPrompt(promptID)
	if err != nil {
		s.writeError(w, fmt.Sprintf("Failed to get prompt: %v", err), http.StatusNotFound)
		return
	}

	var content string
	switch format {
	case "json":
		data, _ := json.MarshalIndent(prompt, "", "  ")
		content = string(data)
	default:
		content = fmt.Sprintf("ID: %s\nTitle: %s\nVersion: %s\nDescription: %s\nTags: %s\n\nContent:\n%s",
			prompt.ID, prompt.Name, prompt.Version, prompt.Summary, 
			strings.Join(prompt.Tags, ", "), prompt.Content)
	}

	s.writeToClipboardAndRespond(w, content, fmt.Sprintf("Retrieved prompt: %s", promptID))
}

// handleList lists all prompts
func (s *URLServer) handleList(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	tag := r.URL.Query().Get("tag")
	limitStr := r.URL.Query().Get("limit")
	
	var prompts []*models.Prompt
	var err error

	if tag != "" {
		prompts, err = s.service.FilterPromptsByTag(tag)
	} else {
		prompts, err = s.service.ListPrompts()
	}

	if err != nil {
		s.writeError(w, fmt.Sprintf("Failed to list prompts: %v", err), http.StatusInternalServerError)
		return
	}

	// Apply limit if specified
	if limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit < len(prompts) {
			prompts = prompts[:limit]
		}
	}

	content := s.formatPrompts(prompts, format)
	s.writeToClipboardAndRespond(w, content, fmt.Sprintf("Listed %d prompts", len(prompts)))
}

// handleSearch performs fuzzy text search
func (s *URLServer) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		s.writeError(w, "Search requires a query parameter 'q'", http.StatusBadRequest)
		return
	}

	format := r.URL.Query().Get("format")
	limitStr := r.URL.Query().Get("limit")
	tag := r.URL.Query().Get("tag")

	prompts, err := s.service.SearchPrompts(query)
	if err != nil {
		s.writeError(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Filter by tag if specified
	if tag != "" {
		var filtered []*models.Prompt
		for _, p := range prompts {
			for _, t := range p.Tags {
				if strings.EqualFold(t, tag) {
					filtered = append(filtered, p)
					break
				}
			}
		}
		prompts = filtered
	}

	// Apply limit if specified
	if limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit < len(prompts) {
			prompts = prompts[:limit]
		}
	}

	content := s.formatPrompts(prompts, format)
	s.writeToClipboardAndRespond(w, content, fmt.Sprintf("Found %d prompts for '%s'", len(prompts), query))
}

// handleBooleanSearch performs boolean expression search
func (s *URLServer) handleBooleanSearch(w http.ResponseWriter, r *http.Request) {
	expr := r.URL.Query().Get("expr")
	if expr == "" {
		s.writeError(w, "Boolean search requires an 'expr' parameter", http.StatusBadRequest)
		return
	}

	format := r.URL.Query().Get("format")
	
	// URL decode the expression
	decodedExpr, err := url.QueryUnescape(expr)
	if err != nil {
		s.writeError(w, fmt.Sprintf("Invalid expression encoding: %v", err), http.StatusBadRequest)
		return
	}

	// Parse boolean expression
	boolExpr, err := s.parseBooleanExpression(decodedExpr)
	if err != nil {
		s.writeError(w, fmt.Sprintf("Invalid boolean expression: %v", err), http.StatusBadRequest)
		return
	}

	// Execute search
	prompts, err := s.service.SearchPromptsByBooleanExpression(boolExpr)
	if err != nil {
		s.writeError(w, fmt.Sprintf("Boolean search failed: %v", err), http.StatusInternalServerError)
		return
	}

	content := s.formatPrompts(prompts, format)
	s.writeToClipboardAndRespond(w, content, fmt.Sprintf("Boolean search found %d prompts", len(prompts)))
}

// handleSavedSearch executes a saved search
func (s *URLServer) handleSavedSearch(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 0 {
		s.writeError(w, "Saved search requires a search name", http.StatusBadRequest)
		return
	}

	searchName := parts[0]
	format := r.URL.Query().Get("format")

	prompts, err := s.service.ExecuteSavedSearch(searchName)
	if err != nil {
		s.writeError(w, fmt.Sprintf("Failed to execute saved search: %v", err), http.StatusNotFound)
		return
	}

	content := s.formatPrompts(prompts, format)
	s.writeToClipboardAndRespond(w, content, fmt.Sprintf("Saved search '%s' found %d prompts", searchName, len(prompts)))
}

// handleSavedSearches lists saved searches
func (s *URLServer) handleSavedSearches(w http.ResponseWriter, r *http.Request, parts []string) {
	operation := "list"
	if len(parts) > 0 {
		operation = parts[0]
	}

	switch operation {
	case "list":
		searches, err := s.service.ListSavedSearches()
		if err != nil {
			s.writeError(w, fmt.Sprintf("Failed to list saved searches: %v", err), http.StatusInternalServerError)
			return
		}

		var content strings.Builder
		for _, search := range searches {
			content.WriteString(fmt.Sprintf("%s: %s\n", search.Name, search.Expression.String()))
		}

		s.writeToClipboardAndRespond(w, content.String(), fmt.Sprintf("Listed %d saved searches", len(searches)))
	default:
		s.writeError(w, fmt.Sprintf("Unknown saved searches operation: %s", operation), http.StatusNotFound)
	}
}

// handleTags lists all tags
func (s *URLServer) handleTags(w http.ResponseWriter, r *http.Request) {
	tags, err := s.service.GetAllTags()
	if err != nil {
		s.writeError(w, fmt.Sprintf("Failed to get tags: %v", err), http.StatusInternalServerError)
		return
	}

	content := strings.Join(tags, "\n")
	s.writeToClipboardAndRespond(w, content, fmt.Sprintf("Listed %d tags", len(tags)))
}

// handleTag lists prompts with a specific tag
func (s *URLServer) handleTag(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 0 {
		s.writeError(w, "Tag operation requires a tag name", http.StatusBadRequest)
		return
	}

	tagName := parts[0]
	format := r.URL.Query().Get("format")

	prompts, err := s.service.FilterPromptsByTag(tagName)
	if err != nil {
		s.writeError(w, fmt.Sprintf("Failed to filter by tag: %v", err), http.StatusInternalServerError)
		return
	}

	content := s.formatPrompts(prompts, format)
	s.writeToClipboardAndRespond(w, content, fmt.Sprintf("Tag '%s' has %d prompts", tagName, len(prompts)))
}

// handleTemplates lists all templates
func (s *URLServer) handleTemplates(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")

	templates, err := s.service.ListTemplates()
	if err != nil {
		s.writeError(w, fmt.Sprintf("Failed to list templates: %v", err), http.StatusInternalServerError)
		return
	}

	var content string
	switch format {
	case "json":
		data, _ := json.MarshalIndent(templates, "", "  ")
		content = string(data)
	case "ids":
		var ids []string
		for _, t := range templates {
			ids = append(ids, t.ID)
		}
		content = strings.Join(ids, "\n")
	default:
		var lines []string
		for _, t := range templates {
			line := fmt.Sprintf("%s - %s", t.ID, t.Name)
			if t.Description != "" {
				line += fmt.Sprintf("\n  %s", t.Description)
			}
			lines = append(lines, line)
		}
		content = strings.Join(lines, "\n\n")
	}

	s.writeToClipboardAndRespond(w, content, fmt.Sprintf("Listed %d templates", len(templates)))
}

// handleTemplate gets a specific template
func (s *URLServer) handleTemplate(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 0 {
		s.writeError(w, "Template operation requires a template ID", http.StatusBadRequest)
		return
	}

	templateID := parts[0]
	format := r.URL.Query().Get("format")

	template, err := s.service.GetTemplate(templateID)
	if err != nil {
		s.writeError(w, fmt.Sprintf("Failed to get template: %v", err), http.StatusNotFound)
		return
	}

	var content string
	switch format {
	case "json":
		data, _ := json.MarshalIndent(template, "", "  ")
		content = string(data)
	default:
		content = fmt.Sprintf("ID: %s\nName: %s\nVersion: %s\nDescription: %s\n\nContent:\n%s",
			template.ID, template.Name, template.Version, template.Description, template.Content)
		
		if len(template.Slots) > 0 {
			content += "\n\nSlots:\n"
			for _, slot := range template.Slots {
				content += fmt.Sprintf("  %s", slot.Name)
				if slot.Required {
					content += " [required]"
				}
				if slot.Default != "" {
					content += fmt.Sprintf(" [default: %s]", slot.Default)
				}
				if slot.Description != "" {
					content += fmt.Sprintf(" - %s", slot.Description)
				}
				content += "\n"
			}
		}
	}

	s.writeToClipboardAndRespond(w, content, fmt.Sprintf("Retrieved template: %s", templateID))
}

// formatPrompts formats a list of prompts for output
func (s *URLServer) formatPrompts(prompts []*models.Prompt, format string) string {
	switch format {
	case "json":
		data, _ := json.MarshalIndent(prompts, "", "  ")
		return string(data)
	case "ids":
		var ids []string
		for _, p := range prompts {
			ids = append(ids, p.ID)
		}
		return strings.Join(ids, "\n")
	case "table":
		var lines []string
		lines = append(lines, fmt.Sprintf("%-20s %-30s %-15s %s", "ID", "Title", "Version", "Updated"))
		lines = append(lines, strings.Repeat("-", 80))
		for _, p := range prompts {
			title := p.Name
			if len(title) > 30 {
				title = title[:27] + "..."
			}
			lines = append(lines, fmt.Sprintf("%-20s %-30s %-15s %s", 
				p.ID, title, p.Version, p.UpdatedAt.Format("2006-01-02")))
		}
		return strings.Join(lines, "\n")
	default:
		var lines []string
		for _, p := range prompts {
			line := fmt.Sprintf("%s - %s", p.ID, p.Name)
			if p.Summary != "" {
				line += fmt.Sprintf("\n  %s", p.Summary)
			}
			if len(p.Tags) > 0 {
				line += fmt.Sprintf("\n  Tags: %s", strings.Join(p.Tags, ", "))
			}
			lines = append(lines, line)
		}
		return strings.Join(lines, "\n\n")
	}
}

// writeToClipboardAndRespond puts content in clipboard and sends success response
func (s *URLServer) writeToClipboardAndRespond(w http.ResponseWriter, content, message string) {
	// Copy to clipboard
	if statusMsg, err := clipboard.CopyWithFallback(content); err != nil {
		log.Printf("Warning: failed to copy to clipboard: %v", err)
		// Continue anyway - content might still be useful
	} else {
		log.Printf("Clipboard: %s", statusMsg)
	}

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success": true,
		"message": message,
		"clipboard": "Content copied to clipboard",
		"length": len(content),
	}
	json.NewEncoder(w).Encode(response)
}

// writeError sends an error response
func (s *URLServer) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error": message,
	})
}

// parseBooleanExpression parses a boolean search expression
// This is a simplified implementation - could be enhanced with a proper parser
func (s *URLServer) parseBooleanExpression(expr string) (*models.BooleanExpression, error) {
	expr = strings.TrimSpace(expr)
	
	// Handle NOT expressions
	if strings.HasPrefix(strings.ToUpper(expr), "NOT ") {
		inner := strings.TrimSpace(expr[4:])
		innerExpr, err := s.parseBooleanExpression(inner)
		if err != nil {
			return nil, err
		}
		return models.NewNotExpression(innerExpr), nil
	}
	
	// Handle OR expressions (lower precedence)
	if orParts := strings.Split(expr, " OR "); len(orParts) > 1 {
		var expressions []*models.BooleanExpression
		for _, part := range orParts {
			subExpr, err := s.parseBooleanExpression(strings.TrimSpace(part))
			if err != nil {
				return nil, err
			}
			expressions = append(expressions, subExpr)
		}
		return models.NewOrExpression(expressions...), nil
	}
	
	// Handle AND expressions (higher precedence)
	if andParts := strings.Split(expr, " AND "); len(andParts) > 1 {
		var expressions []*models.BooleanExpression
		for _, part := range andParts {
			subExpr, err := s.parseBooleanExpression(strings.TrimSpace(part))
			if err != nil {
				return nil, err
			}
			expressions = append(expressions, subExpr)
		}
		return models.NewAndExpression(expressions...), nil
	}
	
	// Remove parentheses if present
	if strings.HasPrefix(expr, "(") && strings.HasSuffix(expr, ")") {
		return s.parseBooleanExpression(expr[1 : len(expr)-1])
	}
	
	// Single tag expression
	return models.NewTagExpression(expr), nil
}