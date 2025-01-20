package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
    "strings"
)

// single bookmark structure
type Bookmark struct {
	ID          int       `json:"id"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	TagNames    []string  `json:"tag_names"`
	DateAdded   time.Time `json:"date_added"`
}

// API response structure
type APIResponse struct {
	Count    int        `json:"count"`
	Next     string     `json:"next"`
	Previous string     `json:"previous"`
	Results  []Bookmark `json:"results"`
}

// fetch bookmarks from the API using the token
func fetchBookmarks(url string, token string) (*APIResponse, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Token "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse APIResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, err
	}

	return &apiResponse, nil
}

// update bookmark
func updateBookmarkTags(bookmarkID int, tags []string, token string, apiBaseURL string) error {
	client := &http.Client{}
	url := fmt.Sprintf("%s/%d/", apiBaseURL, bookmarkID)

	// Create the JSON payload
	payload := map[string]interface{}{
		"tag_names": tags,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %v", err)
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Authorization", "Token "+token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check for a non-200 status code and handle accordingly
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update bookmark with status: %d. Response body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// checkURLValidity checks if a URL is valid with HEAD request and returns the validity and HTTP status code.
func checkURLValidity(url string) (bool, string) {
	client := &http.Client{}
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		// Check if the error is related to DNS resolution
		if _, ok := err.(*net.DNSError); ok {
			// Return DNS error
			return false, "DNS"
		}
		// Otherwise, return a generic other error
		return false, "other"
	}

	// pretend we're a regular browser (helps reducing the amount of 403s)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.5481.100 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		// Return a generic other error for network or connection issues
		return false, "other"
	}
	defer resp.Body.Close()

	// For valid HTTP status codes (2xx, 3xx), consider the link valid
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return true, ""
	}

	// For invalid status codes (4xx, 5xx), return the status code
	return false, fmt.Sprintf("HTTP_%d", resp.StatusCode)
}

func main() {
	// Fetch the API token and base URL from environment variables
	token := os.Getenv("API_TOKEN")
	if token == "" {
		fmt.Println("Error: API_TOKEN environment variable not set")
		return
	}

	apiBaseURL := os.Getenv("API_URL")
	if apiBaseURL == "" {
		fmt.Println("Error: API_URL environment variable not set")
		return
	}

	// Append the query parameters for pagination
	baseURL := fmt.Sprintf("%s/?limit=100", apiBaseURL)

	allBookmarks := []Bookmark{}
	nextURL := baseURL

	// Fetch all bookmarks with pagination
	for nextURL != "" {
		apiResponse, err := fetchBookmarks(nextURL, token)
		if err != nil {
			fmt.Printf("Error fetching bookmarks: %v\n", err)
			return
		}

		allBookmarks = append(allBookmarks, apiResponse.Results...)
		nextURL = apiResponse.Next
		fmt.Printf("Fetched %d bookmarks so far...\n", len(allBookmarks))
	}

	fmt.Printf("Total bookmarks fetched: %d\n", len(allBookmarks))

	// Validate and update all bookmarks
	for _, bookmark := range allBookmarks {
		valid, errorType := checkURLValidity(bookmark.URL)

		if valid {
			// link is OK
			fmt.Printf("Bookmark %d ✅ %s\n", bookmark.ID, bookmark.URL)

            // Remove any tags starting with @HEALTH_
            tags := bookmark.TagNames
            updatedTags := []string{}
            for _, t := range tags {
                if !strings.HasPrefix(t, "@HEALTH_") {
                    updatedTags = append(updatedTags, t)
                }
            }

            // Update the bookmark's tags if any were removed
            if len(updatedTags) != len(tags) {
                err := updateBookmarkTags(bookmark.ID, updatedTags, token, apiBaseURL)
                if err != nil {
                    fmt.Printf("Error updating bookmark %d: %v\n", bookmark.ID, err)
                }
            }
		} else {
			// link is not OK
			fmt.Printf("Bookmark %d ❌ %s %s\n", bookmark.ID, errorType, bookmark.URL)

			// Determine the appropriate tag based on the error type
			var tag string
			if errorType == "DNS" {
				tag = "@HEALTH_DNS"
			} else if errorType == "other" {
				tag = "@HEALTH_other"
			} else {
				tag = "@HEALTH_" + errorType // For HTTP errors (e.g., @HEALTH_HTTP_404)
			}

			// Add the error tag if it's not already in the tags
			// We are not removing previous error tags if the URL goes from 500 to 404. We only clean up the error tags if the link becomes valid again
			tags := bookmark.TagNames
			tagExists := false
			for _, t := range tags {
				if t == tag {
					tagExists = true
					break
				}
			}
			if !tagExists {
				tags = append(tags, tag)
			}

			// Update the bookmark's tags
			err := updateBookmarkTags(bookmark.ID, tags, token, apiBaseURL)
			if err != nil {
				fmt.Printf("Error updating bookmark %d: %v\n", bookmark.ID, err)
			}
		}
	}
}
