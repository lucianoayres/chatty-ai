package share

import (
	"fmt"
	"time"
)

// ShareConfig holds configuration for the sharing feature
type ShareConfig struct {
	BaseURL     string // Base URL for the community store repository
	PRTemplate  string // Template for pull request description
	BranchName  string // Branch name format for submissions
	CommitMsg   string // Commit message format
}

// DefaultShareConfig returns the default configuration
func DefaultShareConfig() ShareConfig {
	return ShareConfig{
		BaseURL:     "https://github.com/lucianoayres/chatty-ai-community-store",
		PRTemplate:  "## Agent Submission\n\n### Description\n%s\n\n### Tags\n%s\n\n### Preview\n```yaml\n%s\n```",
		BranchName:  "agent-submission/%s-%s",
		CommitMsg:   "Add new agent: %s",
	}
}

// ValidationResult represents the result of agent validation
type ValidationResult struct {
	IsValid     bool
	Errors      []string
	Warnings    []string
}

// ShareAnimation handles loading animations for sharing operations
type ShareAnimation struct {
	stopChan chan bool
	message  string
}

// NewShareAnimation creates a new share animation
func NewShareAnimation(message string) *ShareAnimation {
	return &ShareAnimation{
		stopChan: make(chan bool),
		message:  message,
	}
}

// Start begins the loading animation
func (a *ShareAnimation) Start() {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	frameIndex := 0
	
	// Start animation in background
	go func() {
		for {
			select {
			case <-a.stopChan:
				return
			default:
				// Clear line and print current frame with message
				fmt.Printf("\r\033[K%s %s", frames[frameIndex], a.message)
				
				// Move to next frame
				frameIndex = (frameIndex + 1) % len(frames)
				
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

// Stop ends the loading animation
func (a *ShareAnimation) Stop() {
	a.stopChan <- true
	// Clear the animation line
	fmt.Print("\r\033[K")
} 