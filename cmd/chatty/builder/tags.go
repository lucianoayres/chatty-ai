package builder

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

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

const (
	// Base URL for the community store repository
	tagsBaseURL = "https://raw.githubusercontent.com/lucianoayres/chatty-ai-community-store/refs/heads/main"
	tagsConfigPath = "tags.json"
	tagsRequestTimeout = 30 // seconds
)

// GetTagsConfigURL returns the full URL for the tags configuration file
func GetTagsConfigURL() string {
	return tagsBaseURL + "/" + tagsConfigPath
}

// FetchTagsConfig retrieves the tags configuration from the remote URL
func FetchTagsConfig(debug bool) (*TagsConfig, error) {
	if debug {
		fmt.Println("Fetching tags configuration from:", GetTagsConfigURL())
	}

	// Create an HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(tagsRequestTimeout) * time.Second,
	}

	// Make the HTTP request
	resp, err := client.Get(GetTagsConfigURL())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to tags source: %v", err)
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

// SelectTags prompts the user to select tags for the agent
func SelectTags(debug bool) ([]string, error) {
	// Clear screen before starting
	fmt.Print("\033[J")
	
	fmt.Printf("\n%s4️⃣ Tag Selection%s\n", colorSection, colorReset)
	fmt.Printf("%sChoose tags for your agent (1-5 tags required).%s\n", colorPrompt, colorReset)
	
	// Create a channel for stopping the spinner
	done := make(chan bool)
	
	// Define spinner characters
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	
	// Start spinner goroutine
	go func() {
		i := 0
		for {
			select {
			case <-done:
				return
			default:
				fmt.Printf("\r%sLoading tags... %s%s", colorValue, spinner[i], colorReset)
				time.Sleep(100 * time.Millisecond)
				i = (i + 1) % len(spinner)
			}
		}
	}()
	
	// Fetch tags from the remote URL
	tagsConfig, err := FetchTagsConfig(debug)
	
	// Stop spinner
	done <- true
	close(done)
	
	// Clear the loading message
	fmt.Print("\r\033[K") // Clear the current line
	
	if err != nil {
		fmt.Printf("%s⚠️ Warning: Failed to load tags. Using default tags.%s\n", 
			colorAccent, colorReset)
		// Return some default tags if we can't fetch them
		return []string{"general"}, nil
	}
	
	// Display available tags
	fmt.Printf("\n%sAvailable Tags:%s\n", colorSection, colorReset)
	
	// Create maps to store index <-> tag relationships
	tagIndexToName := make(map[int]string)     // Map index to tag name
	tagIndexToID := make(map[int]string)       // Map index to original tag ID
	tagNameToIndex := make(map[string]int)     // Map tag name to index
	
	// Sort tag IDs for consistent display
	var tagKeys []string
	for k := range tagsConfig.Tags {
		tagKeys = append(tagKeys, k)
	}
	
	// Print the tags with sequential numbers, excluding "Featured"
	fmt.Println()
	index := 1
	for _, id := range tagKeys {
		tag := tagsConfig.Tags[id]
		
		// Skip the "Featured" category
		if tag.Name == "Featured" {
			continue
		}
		
		tagIndexToName[index] = tag.Name
		tagIndexToID[index] = id
		tagNameToIndex[tag.Name] = index
		
		fmt.Printf("%s%d.%s %s%s%s - %s\n", 
			colorValue, index, colorReset,
			colorHighlight, tag.Name, colorReset, 
			tag.Description)
		
		index++
	}
	
	// Get user input
	var selectedTags []string
	
	for {
		fmt.Printf("\n%sEnter tag numbers (comma-separated, 1-5 tags):%s ", 
			colorPrompt, colorReset)
		
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		
		// Split by commas
		tagStrs := strings.Split(input, ",")
		
		// Validate and convert to tag IDs
		valid := true
		var selectedTagsTemp []string
		selectedIndexes := make(map[int]bool) // Track selected indexes to prevent duplicates
		
		for _, tagStr := range tagStrs {
			tagStr = strings.TrimSpace(tagStr)
			if tagStr == "" {
				continue
			}
			
			// Parse the tag index
			tagIndex, err := strconv.Atoi(tagStr)
			if err != nil || tagIndex < 1 || tagIndex >= index {
				fmt.Printf("%s❌ Invalid tag number: %s%s\n", 
					colorAccent, tagStr, colorReset)
				valid = false
				break
			}
			
			// Check for duplicates
			if selectedIndexes[tagIndex] {
				fmt.Printf("%s❌ Duplicate tag: '%s' already selected%s\n",
					colorAccent, tagIndexToName[tagIndex], colorReset)
				valid = false
				break
			}
			
			// Mark as selected and add to the list
			selectedIndexes[tagIndex] = true
			selectedTagsTemp = append(selectedTagsTemp, tagIndexToName[tagIndex])
		}
		
		// Check number of tags
		if valid && (len(selectedTagsTemp) < 1 || len(selectedTagsTemp) > 5) {
			fmt.Printf("%s❌ Please select between 1 and 5 tags%s\n", 
				colorAccent, colorReset)
			valid = false
		}
		
		// If valid, use these tags
		if valid {
			selectedTags = selectedTagsTemp
			break
		}
	}
	
	// Display selected tags
	fmt.Printf("\n%sSelected Tags:%s\n", colorSection, colorReset)
	for _, tagName := range selectedTags {
		fmt.Printf("%s✓%s %s%s%s\n", 
			colorGreen, colorReset,
			colorHighlight, tagName, colorReset)
	}
	
	// Clean up before returning
	// Wait a moment for the user to see the selected tags
	fmt.Printf("\n%s✓ Tags selected successfully%s\n", colorGreen, colorReset)
	time.Sleep(1 * time.Second)
	
	// Clear screen before returning
	fmt.Print("\033[J")
	
	return selectedTags, nil
} 