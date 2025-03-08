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

// FetchIndex retrieves the store index
func (c *Client) FetchIndex() (*StoreIndex, error) {
	if c.debug {
		fmt.Println("Fetching store index from:", GetIndexURL())
	}

	// Make the HTTP request
	resp, err := c.httpClient.Get(GetIndexURL())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to community store: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch store index: HTTP %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read store index: %v", err)
	}

	// Parse the JSON response
	var index StoreIndex
	if err := json.Unmarshal(body, &index); err != nil {
		return nil, fmt.Errorf("failed to parse store index: %v", err)
	}

	return &index, nil
}

// FetchStoreConfig retrieves the store configuration
func (c *Client) FetchStoreConfig() (*StoreConfig, error) {
	if c.debug {
		fmt.Println("Fetching store configuration from:", GetStoreConfigURL())
	}

	// Make the HTTP request
	resp, err := c.httpClient.Get(GetStoreConfigURL())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to community store: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch store configuration: HTTP %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read store configuration: %v", err)
	}

	// Parse the JSON response
	var config StoreConfig
	if err := json.Unmarshal(body, &config); err != nil {
		return nil, fmt.Errorf("failed to parse store configuration: %v", err)
	}

	return &config, nil
}

// FetchTagsConfig retrieves the tags configuration
func (c *Client) FetchTagsConfig() (*TagsConfig, error) {
	if c.debug {
		fmt.Println("Fetching tags configuration from:", GetTagsConfigURL())
	}

	// Make the HTTP request
	resp, err := c.httpClient.Get(GetTagsConfigURL())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to community store: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch tags configuration: HTTP %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read tags configuration: %v", err)
	}

	// Parse the JSON response
	var config TagsConfig
	if err := json.Unmarshal(body, &config); err != nil {
		return nil, fmt.Errorf("failed to parse tags configuration: %v", err)
	}

	return &config, nil
}

// FetchAgent retrieves an agent's YAML file from the store
func (c *Client) FetchAgent(filename string) ([]byte, error) {
	agentURL := GetAgentURL(filename)
	
	if c.debug {
		fmt.Println("Fetching agent from:", agentURL)
	}

	// Make the HTTP request
	resp, err := c.httpClient.Get(agentURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to community store: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch agent: HTTP %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read agent data: %v", err)
	}

	return body, nil
} 