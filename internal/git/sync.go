package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GitSync handles automatic git synchronization
type GitSync struct {
	baseDir string
	enabled bool
}

// NewGitSync creates a new GitSync instance
func NewGitSync(baseDir string) *GitSync {
	return &GitSync{
		baseDir: baseDir,
		enabled: false, // Will be set by checking if git is initialized
	}
}

// IsEnabled returns true if git sync is available and enabled
func (g *GitSync) IsEnabled() bool {
	return g.enabled && g.isGitInitialized()
}

// Initialize checks if git is set up and enables sync if available
func (g *GitSync) Initialize() error {
	if !g.isGitInitialized() {
		g.enabled = false
		return nil // Not an error, just not available
	}
	
	// Check if we have a remote configured
	if !g.hasRemote() {
		g.enabled = false
		return nil // Not an error, but can't sync without remote
	}
	
	g.enabled = true
	return nil
}

// Enable sets git sync to enabled (user preference)
func (g *GitSync) Enable() {
	g.enabled = true
}

// Disable sets git sync to disabled (user preference)
func (g *GitSync) Disable() {
	g.enabled = false
}

// isGitInitialized checks if the directory has git initialized
func (g *GitSync) isGitInitialized() bool {
	gitDir := filepath.Join(g.baseDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return false
	}
	return true
}

// hasRemote checks if git has a remote configured
func (g *GitSync) hasRemote() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "git", "remote", "-v")
	cmd.Dir = g.baseDir
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

// SyncChanges commits and pushes changes to git
func (g *GitSync) SyncChanges(message string) error {
	if !g.IsEnabled() {
		return nil // Silently skip if not enabled
	}

	// Stage all changes
	if err := g.runGitCommand("add", "-A"); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	// Check if there are changes to commit
	hasChanges, err := g.hasChangesToCommit()
	if err != nil {
		return fmt.Errorf("failed to check for changes: %w", err)
	}
	
	if !hasChanges {
		return nil // No changes to sync
	}

	// Commit changes
	commitMessage := fmt.Sprintf("%s - %s", message, time.Now().Format("2006-01-02 15:04:05"))
	if err := g.runGitCommand("commit", "-m", commitMessage); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Push changes (best effort - don't fail if push fails)
	if err := g.runGitCommand("push"); err != nil {
		// Log the error but don't fail the operation
		// The user can manually push later if needed
		return fmt.Errorf("committed locally but failed to push: %w", err)
	}

	return nil
}

// hasChangesToCommit checks if there are staged changes ready to commit
func (g *GitSync) hasChangesToCommit() (bool, error) {
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	cmd.Dir = g.baseDir
	err := cmd.Run()
	if err != nil {
		// diff --quiet returns non-zero exit code if there are differences
		if exitError, ok := err.(*exec.ExitError); ok {
			// Exit code 1 means there are differences (changes to commit)
			if exitError.ExitCode() == 1 {
				return true, nil
			}
		}
		return false, err
	}
	// Exit code 0 means no differences (no changes to commit)
	return false, nil
}

// runGitCommand executes a git command in the base directory with timeout
func (g *GitSync) runGitCommand(args ...string) error {
	return g.runGitCommandWithTimeout(10*time.Second, args...)
}

// runGitCommandWithTimeout executes a git command with custom timeout
func (g *GitSync) runGitCommandWithTimeout(timeout time.Duration, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = g.baseDir
	
	// Capture both stdout and stderr for better error messages
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("git %s timed out after %v", strings.Join(args, " "), timeout)
		}
		return fmt.Errorf("git %s failed: %s", strings.Join(args, " "), string(output))
	}
	
	return nil
}

// GetStatus returns the current git status information
func (g *GitSync) GetStatus() (string, error) {
	if !g.isGitInitialized() {
		return "Git not initialized", nil
	}
	
	if !g.hasRemote() {
		return "No remote configured", nil
	}
	
	if !g.enabled {
		return "Git sync disabled", nil
	}
	
	// Check if we're ahead/behind remote with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain", "--branch")
	cmd.Dir = g.baseDir
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "Git status timeout", nil
		}
		return "Git status unknown", err
	}
	
	statusLines := strings.Split(string(output), "\n")
	if len(statusLines) > 0 {
		branchLine := statusLines[0]
		if strings.Contains(branchLine, "[ahead") {
			return "Changes need to be pushed", nil
		}
		if strings.Contains(branchLine, "[behind") {
			return "Remote has new changes", nil
		}
	}
	
	// Check for uncommitted changes
	if len(statusLines) > 1 && statusLines[1] != "" {
		return "Uncommitted changes", nil
	}
	
	return "In sync", nil
}