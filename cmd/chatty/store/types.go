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
	Tags        []string  `json:"tags,omitempty"`
	Author      string    `json:"author,omitempty"`
}

// StoreConfig represents the store configuration settings
type StoreConfig struct {
	StorefrontSettings StorefrontSettings `json:"storefrontSettings"`
}

// StorefrontSettings contains the configuration for the store display
type StorefrontSettings struct {
	Categories []CategoryConfig `json:"categories"`
}

// CategoryConfig defines a category in the store
type CategoryConfig struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Tags           []string `json:"tags"`
	Enabled        bool     `json:"enabled"`
	MaxItems       int      `json:"maxItems,omitempty"`
	TimeWindowDays int      `json:"timeWindowDays,omitempty"`
}

// TagsConfig represents the available tags configuration
type TagsConfig struct {
	Tags map[string]TagDefinition `json:"tags"`
}

// TagDefinition defines a single tag with its metadata
type TagDefinition struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Examples    []string `json:"examples"`
}

// StoreAnimation handles animated loading for store operations
type StoreAnimation struct {
	stopChan chan bool
	message  string
}

// NewStoreAnimation creates a new animation with the specified message
func NewStoreAnimation(message string) *StoreAnimation {
	return &StoreAnimation{
		stopChan: make(chan bool),
		message:  message,
	}
}

// Start begins the animation in a separate goroutine
func (a *StoreAnimation) Start() {
	// Animation frames
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	frameIndex := 0

	// Start animation in a goroutine
	go func() {
		for {
			select {
			case <-a.stopChan:
				fmt.Printf("\r%s\r", "                                                  ")
				return
			default:
				fmt.Printf("\r%s %s", frames[frameIndex], a.message)
				frameIndex = (frameIndex + 1) % len(frames)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

// Stop ends the animation
func (a *StoreAnimation) Stop() {
	a.stopChan <- true
} 