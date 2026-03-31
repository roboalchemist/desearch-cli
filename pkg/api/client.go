package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// BaseURL is the Desearch API base URL.
const BaseURL = "https://api.desearch.ai"

// Client is a Desearch API client.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Desearch API client.
func NewClient(apiKey string) *Client {
	return &Client{
		baseURL:    BaseURL,
		apiKey:     apiKey,
		httpClient: http.DefaultClient,
	}
}

// SearchRequest represents a search request to the Desearch API.
type SearchRequest struct {
	Prompt                string   `json:"prompt"`
	Tools                []string `json:"tools"`
	StartDate            *string  `json:"start_date,omitempty"`
	EndDate              *string  `json:"end_date,omitempty"`
	DateFilter           *string  `json:"date_filter,omitempty"`
	Streaming            *bool    `json:"streaming,omitempty"`
	ResultType           *string  `json:"result_type,omitempty"`
	SystemMessage        *string  `json:"system_message,omitempty"`
	ScoringSystemMessage *string  `json:"scoring_system_message,omitempty"`
	Count                *int     `json:"count,omitempty"`
}

// SearchResponse represents a search response from the Desearch API.
type SearchResponse struct {
	HackerNewsSearch []map[string]interface{} `json:"hacker_news_search,omitempty"`
	RedditSearch     []map[string]interface{} `json:"reddit_search,omitempty"`
	Search           []map[string]interface{} `json:"search,omitempty"`
	YoutubeSearch    []map[string]interface{} `json:"youtube_search,omitempty"`
	Tweets           []map[string]interface{} `json:"tweets,omitempty"`
	Text             *string                 `json:"text,omitempty"`
	MinerLinkScores  map[string]string       `json:"miner_link_scores,omitempty"`
	Completion       *string                 `json:"completion,omitempty"`
}

// Search performs a non-streaming search request.
func (c *Client) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	// Disable streaming for Search
	streaming := false
	req.Streaming = &streaming

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/desearch/ai/search", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &searchResp, nil
}

// SearchStream performs a streaming search request and returns a bufio.Reader for the response body.
func (c *Client) SearchStream(ctx context.Context, req *SearchRequest) (*bufio.Reader, error) {
	// Ensure streaming is enabled
	streaming := true
	req.Streaming = &streaming

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/desearch/ai/search", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return bufio.NewReader(resp.Body), nil
}
