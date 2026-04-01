package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"
)

// BaseURL is the Desearch API base URL.
const BaseURL = "https://api.desearch.ai"

// Client is a Desearch API client.
type Client struct {
	APIKey     string
	HTTPClient *http.Client
	BaseURL    string
}

// NewClient creates a new Desearch API client.
func NewClient(apiKey string) *Client {
	return &Client{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		BaseURL: BaseURL,
	}
}

// SearchRequest represents a search request to the Desearch API.
type SearchRequest struct {
	Prompt               string   `json:"prompt"`
	Tools                []string `json:"tools,omitempty"`
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
	Search           []WebResult        `json:"search,omitempty"`
	HackerNewsSearch []HackerNewsResult `json:"hacker_news_search,omitempty"`
	RedditSearch     []RedditResult     `json:"reddit_search,omitempty"`
	YoutubeSearch    []YoutubeResult    `json:"youtube_search,omitempty"`
	Tweets           []TweetResult      `json:"tweets,omitempty"`
	WikipediaSearch  []WikipediaResult  `json:"wikipedia_search,omitempty"`
	ArxivSearch      []ArxivResult      `json:"arxiv_search,omitempty"`
	Text             string             `json:"text,omitempty"`
	MinerLinkScores  map[string]string  `json:"miner_link_scores,omitempty"`
	Completion       string             `json:"completion,omitempty"`
}

// MarshalJSON sorts MinerLinkScores map keys for deterministic JSON output.
func (r SearchResponse) MarshalJSON() ([]byte, error) {
	// Sort MinerLinkScores keys
	var sortedScores []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	for k, v := range r.MinerLinkScores {
		sortedScores = append(sortedScores, struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}{k, v})
	}
	sort.Slice(sortedScores, func(i, j int) bool {
		return sortedScores[i].Key < sortedScores[j].Key
	})

	alias := struct {
		Search           []WebResult        `json:"search,omitempty"`
		HackerNewsSearch []HackerNewsResult `json:"hacker_news_search,omitempty"`
		RedditSearch     []RedditResult     `json:"reddit_search,omitempty"`
		YoutubeSearch    []YoutubeResult    `json:"youtube_search,omitempty"`
		Tweets           []TweetResult      `json:"tweets,omitempty"`
		WikipediaSearch  []WikipediaResult  `json:"wikipedia_search,omitempty"`
		ArxivSearch      []ArxivResult      `json:"arxiv_search,omitempty"`
		Text             string             `json:"text,omitempty"`
		MinerLinkScores  []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"miner_link_scores,omitempty"`
		Completion string `json:"completion,omitempty"`
	}{
		Search:           r.Search,
		HackerNewsSearch: r.HackerNewsSearch,
		RedditSearch:     r.RedditSearch,
		YoutubeSearch:    r.YoutubeSearch,
		Tweets:           r.Tweets,
		WikipediaSearch:  r.WikipediaSearch,
		ArxivSearch:      r.ArxivSearch,
		Text:             r.Text,
		MinerLinkScores:  sortedScores,
		Completion:       r.Completion,
	}
	return json.MarshalIndent(alias, "", "  ")
}

// WebResult represents a web search result.
type WebResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

// HackerNewsResult represents a Hacker News search result.
type HackerNewsResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

// RedditResult represents a Reddit search result.
type RedditResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

// YoutubeResult represents a YouTube search result.
type YoutubeResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

// TweetUser represents a Twitter user.
type TweetUser struct {
	Username       string `json:"username"`
	Name           string `json:"name"`
	ID             string `json:"id_str"`
	Description    string `json:"description"`
	FollowersCount int    `json:"followers_count,omitempty"`
	FollowingCount int    `json:"following_count,omitempty"`
}

// TweetResult represents a tweet search result.
type TweetResult struct {
	ID               string         `json:"id"`
	Text             string         `json:"text"`
	URL              string         `json:"url,omitempty"`
	User             TweetUser      `json:"user"`
	LikeCount        int            `json:"like_count"`
	RetweetCount     int            `json:"retweet_count"`
	ReplyCount       int            `json:"reply_count"`
	QuoteCount       int            `json:"quote_count"`
	BookmarkCount    int            `json:"bookmark_count"`
	CreatedAt        string         `json:"created_at,omitempty"`
	Lang             string         `json:"lang,omitempty"`
	IsRetweet        bool           `json:"is_retweet,omitempty"`
	IsQuoteTweet     bool           `json:"is_quote_tweet,omitempty"`
	ConversationID   string         `json:"conversation_id,omitempty"`
	DisplayTextRange []int          `json:"display_text_range,omitempty"`
	Entities         *TweetEntities `json:"entities,omitempty"`
	ExtendedEntities *TweetEntities `json:"extended_entities,omitempty"`
	Media            []TweetMedia   `json:"media,omitempty"`
}

// TweetEntities represents tweet entities (hashtags, mentions, etc.).
type TweetEntities struct {
	Hashtags     []TweetHashtag `json:"hashtags"`
	Symbols      []TweetSymbol  `json:"symbols"`
	URLs         []TweetURL     `json:"urls"`
	UserMentions []TweetMention `json:"user_mentions"`
	Media        []TweetMedia   `json:"media"`
	Timestamps   []string       `json:"timestamps"`
}

// TweetHashtag represents a hashtag in a tweet.
type TweetHashtag struct {
	Indices [2]int `json:"indices"`
	Text    string `json:"text"`
}

// TweetSymbol represents a symbol/cashtag in a tweet.
type TweetSymbol struct {
	Indices [2]int `json:"indices"`
	Text    string `json:"text"`
}

// TweetURL represents a URL in a tweet.
type TweetURL struct {
	Indices     [2]int `json:"indices"`
	URL         string `json:"url"`
	DisplayURL  string `json:"display_url"`
	ExpandedURL string `json:"expanded_url"`
}

// TweetMention represents a user mention in a tweet.
type TweetMention struct {
	Indices    [2]int `json:"indices"`
	IDStr      string `json:"id_str"`
	Name       string `json:"name"`
	ScreenName string `json:"screen_name"`
}

// TweetMedia represents media attached to a tweet.
type TweetMedia struct {
	ID                   string                 `json:"id_str"`
	MediaURL             string                 `json:"media_url"`
	Type                 string                 `json:"type"`
	DisplayURL           string                 `json:"display_url"`
	ExpandedURL          string                 `json:"expanded_url"`
	URL                  string                 `json:"url"`
	Indices              [2]int                 `json:"indices"`
	Sizes                TweetMediaSizes        `json:"sizes"`
	OriginalInfo         TweetOriginalInfo      `json:"original_info"`
	Features             TweetMediaFeatures     `json:"features"`
	ExtMediaAvailability TweetMediaAvailability `json:"ext_media_availability"`
	MediaKey             string                 `json:"media_key,omitempty"`
}

// TweetMediaSizes represents available sizes for tweet media.
type TweetMediaSizes struct {
	Large  TweetMediaSize `json:"large"`
	Medium TweetMediaSize `json:"medium"`
	Small  TweetMediaSize `json:"small"`
	Thumb  TweetMediaSize `json:"thumb"`
}

// TweetMediaSize represents dimensions for tweet media.
type TweetMediaSize struct {
	W      int    `json:"w"`
	H      int    `json:"h"`
	Resize string `json:"resize"`
}

// TweetOriginalInfo represents original dimensions of tweet media.
type TweetOriginalInfo struct {
	Width     int              `json:"width"`
	Height    int              `json:"height"`
	FocusRect []TweetFocusRect `json:"focus_rects,omitempty"`
}

// TweetFocusRect represents a focus rectangle for crop.
type TweetFocusRect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// TweetMediaFeatures represents detected features for tweet media.
type TweetMediaFeatures struct {
	Large  TweetFaceRect `json:"large"`
	Medium TweetFaceRect `json:"medium"`
	Small  TweetFaceRect `json:"small"`
	Orig   TweetFaceRect `json:"orig"`
}

// TweetFaceRect represents detected faces in tweet media.
type TweetFaceRect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// TweetMediaAvailability represents availability status of tweet media.
type TweetMediaAvailability struct {
	Status string `json:"status"`
}

// WikipediaResult represents a Wikipedia search result.
type WikipediaResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

// ArxivResult represents an arXiv search result.
type ArxivResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
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

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/desearch/ai/search", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
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

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/desearch/ai/search", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
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
