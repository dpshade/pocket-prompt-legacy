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

// PullChanges pulls changes from the remote repository with conflict resolution
func (g *GitSync) PullChanges() error {
	if !g.IsEnabled() {
		return nil // Silently skip if not enabled
	}

	// First, fetch the latest changes from remote
	if err := g.runGitCommand("fetch", "origin"); err != nil {
		return fmt.Errorf("failed to fetch from remote: %w", err)
	}

	// Check if we're behind the remote
	behind, err := g.isBehindRemote()
	if err != nil {
		return fmt.Errorf("failed to check remote status: %w", err)
	}

	if !behind {
		return nil // Already up to date
	}

	// Try to pull with merge strategy
	err = g.runGitCommand("pull", "origin", g.getCurrentBranch())
	if err != nil {
		// If pull failed, likely due to conflicts or divergent branches
		return g.handlePullConflict(err)
	}

	return nil
}

// SyncChanges commits and pushes changes to git with improved error handling
func (g *GitSync) SyncChanges(message string) error {
	if !g.IsEnabled() {
		return nil // Silently skip if not enabled
	}

	// First, try to pull any remote changes
	if err := g.PullChanges(); err != nil {
		// Don't fail immediately - log the error and try to proceed
		fmt.Printf("Warning: Failed to pull remote changes: %v\n", err)
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

	// Push changes with retry logic
	if err := g.pushWithRetry(); err != nil {
		return fmt.Errorf("committed locally but failed to push: %w", err)
	}

	return nil
}

// pushWithRetry attempts to push with automatic conflict resolution
func (g *GitSync) pushWithRetry() error {
	// First attempt
	err := g.runGitCommand("push", "origin", g.getCurrentBranch())
	if err == nil {
		return nil // Success
	}

	// If push failed, try to pull and push again
	if strings.Contains(err.Error(), "rejected") || strings.Contains(err.Error(), "non-fast-forward") {
		fmt.Printf("Push rejected, attempting to pull and retry...\n")
		
		// Try to pull latest changes
		if pullErr := g.PullChanges(); pullErr != nil {
			return fmt.Errorf("push failed and pull failed: push=%v, pull=%v", err, pullErr)
		}

		// Retry push
		if retryErr := g.runGitCommand("push", "origin", g.getCurrentBranch()); retryErr != nil {
			return fmt.Errorf("push failed after pull: original=%v, retry=%v", err, retryErr)
		}

		return nil // Success after retry
	}

	return err // Other type of error
}

// handlePullConflict handles pull conflicts by attempting automatic resolution
func (g *GitSync) handlePullConflict(pullErr error) error {
	errStr := pullErr.Error()
	
	// Handle divergent branches
	if strings.Contains(errStr, "divergent") || strings.Contains(errStr, "hint: You have divergent branches") {
		fmt.Printf("Detected divergent branches, attempting merge strategy...\n")
		
		// Try merge strategy
		err := g.runGitCommand("pull", "--strategy=recursive", "--strategy-option=ours", "origin", g.getCurrentBranch())
		if err == nil {
			return nil // Merge successful
		}
		
		// If merge failed, try rebase
		fmt.Printf("Merge failed, attempting rebase...\n")
		err = g.runGitCommand("pull", "--rebase", "origin", g.getCurrentBranch())
		if err == nil {
			return nil // Rebase successful
		}
		
		// If both failed, reset to remote state (nuclear option)
		fmt.Printf("Both merge and rebase failed, resetting to remote state...\n")
		return g.resetToRemote()
	}
	
	// Handle merge conflicts
	if strings.Contains(errStr, "conflict") || strings.Contains(errStr, "CONFLICT") {
		fmt.Printf("Detected merge conflicts, attempting automatic resolution...\n")
		return g.resolveConflictsAutomatically()
	}
	
	return pullErr // Unhandled error type
}

// resetToRemote resets local branch to match remote (nuclear option)
func (g *GitSync) resetToRemote() error {
	branch := g.getCurrentBranch()
	
	// Fetch latest
	if err := g.runGitCommand("fetch", "origin"); err != nil {
		return fmt.Errorf("failed to fetch before reset: %w", err)
	}
	
	// Hard reset to remote branch
	if err := g.runGitCommand("reset", "--hard", fmt.Sprintf("origin/%s", branch)); err != nil {
		return fmt.Errorf("failed to reset to remote: %w", err)
	}
	
	fmt.Printf("Successfully reset to remote state\n")
	return nil
}

// resolveConflictsAutomatically attempts to resolve merge conflicts automatically
func (g *GitSync) resolveConflictsAutomatically() error {
	// Get list of conflicted files
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	cmd.Dir = g.baseDir
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get conflicted files: %w", err)
	}
	
	conflictedFiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(conflictedFiles) == 0 || conflictedFiles[0] == "" {
		return fmt.Errorf("no conflicted files found")
	}
	
	// For each conflicted file, prefer remote version (safer for prompt files)
	for _, file := range conflictedFiles {
		if file == "" {
			continue
		}
		
		// Accept remote version
		if err := g.runGitCommand("checkout", "--theirs", file); err != nil {
			return fmt.Errorf("failed to resolve conflict in %s: %w", file, err)
		}
		
		// Stage the resolved file
		if err := g.runGitCommand("add", file); err != nil {
			return fmt.Errorf("failed to stage resolved file %s: %w", file, err)
		}
	}
	
	// Complete the merge
	if err := g.runGitCommand("commit", "--no-edit"); err != nil {
		return fmt.Errorf("failed to complete merge: %w", err)
	}
	
	fmt.Printf("Successfully resolved conflicts in %d files\n", len(conflictedFiles))
	return nil
}

// getCurrentBranch returns the current git branch name
func (g *GitSync) getCurrentBranch() string {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = g.baseDir
	output, err := cmd.Output()
	if err != nil {
		return "main" // Default fallback
	}
	
	branch := strings.TrimSpace(string(output))
	if branch == "" {
		return "main" // Fallback for detached HEAD
	}
	
	return branch
}

// isBehindRemote checks if local branch is behind remote
func (g *GitSync) isBehindRemote() (bool, error) {
	branch := g.getCurrentBranch()
	
	// Get remote hash
	remoteCmd := exec.Command("git", "rev-parse", fmt.Sprintf("origin/%s", branch))
	remoteCmd.Dir = g.baseDir
	remoteOutput, err := remoteCmd.Output()
	if err != nil {
		// Remote branch might not exist yet
		return false, nil
	}
	remoteHash := strings.TrimSpace(string(remoteOutput))
	
	// Get local hash
	localCmd := exec.Command("git", "rev-parse", "HEAD")
	localCmd.Dir = g.baseDir
	localOutput, err := localCmd.Output()
	if err != nil {
		return false, err
	}
	localHash := strings.TrimSpace(string(localOutput))
	
	// If hashes are different, check if we're behind
	if remoteHash != localHash {
		// Check if remote hash is reachable from local (i.e., we're behind)
		mergeBaseCmd := exec.Command("git", "merge-base", "--is-ancestor", localHash, remoteHash)
		mergeBaseCmd.Dir = g.baseDir
		err := mergeBaseCmd.Run()
		return err == nil, nil // If no error, we're behind
	}
	
	return false, nil // Up to date
}

// BackgroundSync runs continuous background synchronization
func (g *GitSync) BackgroundSync(ctx context.Context, interval time.Duration) {
	if !g.IsEnabled() {
		return
	}
	
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Silently pull changes in background
			if err := g.PullChanges(); err != nil {
				// Log but don't spam - only log once per error type
				if !strings.Contains(err.Error(), "timeout") {
					fmt.Printf("Background sync warning: %v\n", err)
				}
			}
		}
	}
}