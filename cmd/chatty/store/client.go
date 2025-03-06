package store

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles communication with the community store
type Client struct {
	httpClient *http.Client
	debug      bool
}

// NewClient creates a new store client
func NewClient(debug bool) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: time.Duration(requestTimeout) * time.Second,
		},
		debug: debug,
	}
}

// FetchIndex retrieves and parses the store index
func (c *Client) FetchIndex() (*StoreIndex, error) {
	if c.debug {
		fmt.Printf("Fetching store index from: %s\n", GetIndexURL())
	}

	resp, err := c.httpClient.Get(GetIndexURL())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch store index: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch store index: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read store index: %v", err)
	}

	if c.debug {
		fmt.Printf("Received store index (%d bytes)\n", len(data))
	}

	var index StoreIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse store index: %v", err)
	}

	return &index, nil
}

// FetchAgent retrieves an agent's YAML file from the store
func (c *Client) FetchAgent(filename string) ([]byte, error) {
	if c.debug {
		fmt.Printf("Fetching agent YAML from: %s\n", GetAgentURL(filename))
	}

	resp, err := c.httpClient.Get(GetAgentURL(filename))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch agent YAML: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch agent YAML: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read agent YAML: %v", err)
	}

	if c.debug {
		fmt.Printf("Received agent YAML (%d bytes)\n", len(data))
	}

	return data, nil
} 