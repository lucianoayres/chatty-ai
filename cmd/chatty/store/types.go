package store

import (
	"fmt"
	"time"
)

// StoreIndex represents the community store index file structure
type StoreIndex struct {
	Version     string       `json:"version"`
	TotalAgents int         `json:"total_agents"`
	Files       []AgentInfo `json:"files"`
}

// AgentInfo represents an agent entry in the store index
type AgentInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Filename    string    `json:"filename"`
	Description string    `json:"description"`
	Emoji       string    `json:"emoji"`
	CreatedAt   time.Time `json:"created_at"`
}

// StoreAnimation handles loading animations for store operations
type StoreAnimation struct {
	stopChan chan bool
	message  string
}

// NewStoreAnimation creates a new store animation
func NewStoreAnimation(message string) *StoreAnimation {
	return &StoreAnimation{
		stopChan: make(chan bool),
		message:  message,
	}
}

// Start begins the loading animation
func (a *StoreAnimation) Start() {
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
func (a *StoreAnimation) Stop() {
	a.stopChan <- true
	// Clear the animation line
	fmt.Print("\r\033[K")
} 