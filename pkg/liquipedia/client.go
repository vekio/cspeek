// Package liquipedia provides a client for the Liquipedia v3 matches API.
// It preserves API records; match classification and display belong to the caller.
package liquipedia

import (
	"compress/gzip"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const DefaultBaseURL = "https://api.liquipedia.net/api/v3/"

// Config configures authentication and transport. Only APIKey is required.
type Config struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// Client can be reused across concurrent requests.
type Client struct {
	apiKey   string
	matchURL url.URL
	http     *http.Client
}

// NewClient uses a 30-second timeout unless an HTTPClient is supplied.
func NewClient(config Config) (*Client, error) {
	key := strings.TrimSpace(config.APIKey)
	if key == "" || strings.ContainsAny(key, "\r\n") {
		return nil, errors.New("liquipedia: a valid API key is required")
	}
	base := config.BaseURL
	if base == "" {
		base = DefaultBaseURL
	}
	u, err := url.Parse(base)
	if err != nil || u == nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("liquipedia: base URL must be an HTTP(S) URL without credentials, query, or fragment")
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/match"
	u.RawPath = ""
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{apiKey: key, matchURL: *u, http: httpClient}, nil
}

// MatchesOptions describes one API page. No game, tier, or time filter is implicit.
type MatchesOptions struct {
	Wiki string
	// Conditions uses the API's [[field::value]] grammar, without URL encoding.
	// Build it from trusted values; URL encoding does not escape that grammar.
	Conditions string
	Fields     []string // Empty requests all fields.
	Order      string   // Empty uses date ASC,objectname ASC.
	Limit      int      // Zero uses 20; maximum 1000.
	Offset     int      // Number of API records skipped, not unique matches.
	RawStreams bool
	StreamURLs bool
}

// MatchesPage preserves duplicates and warnings from a single API response.
// Advance Offset by the requested Limit after a full page. Keep filters and time
// cutoffs fixed while paging; upstream changes can still affect offset pagination.
type MatchesPage struct {
	Matches  []Match
	Warnings []string
}

// APIError reports an HTTP failure or an API error, including those sent with 200.
type APIError struct {
	StatusCode int
	Messages   []string
	Warnings   []string
}

func (e *APIError) Error() string {
	message := fmt.Sprintf("liquipedia: HTTP %d", e.StatusCode)
	if len(e.Messages) > 0 {
		message += ": " + strings.Join(e.Messages, "; ")
	}
	return message
}

// Matches fetches one page. It does not retry, classify, deduplicate, or print.
func (c *Client) Matches(ctx context.Context, options MatchesOptions) (MatchesPage, error) {
	var page MatchesPage
	if strings.TrimSpace(options.Wiki) == "" {
		return page, errors.New("liquipedia: wiki is required")
	}
	if options.Limit < 0 || options.Limit > 1000 || options.Offset < 0 {
		return page, errors.New("liquipedia: limit must be 0..1000 and offset must be nonnegative")
	}
	limit := options.Limit
	if limit == 0 {
		limit = 20
	}
	order := options.Order
	if order == "" {
		order = "date ASC,objectname ASC"
	}
	u := c.matchURL
	q := url.Values{
		"wiki":       {options.Wiki},
		"limit":      {strconv.Itoa(limit)},
		"offset":     {strconv.Itoa(options.Offset)},
		"order":      {order},
		"rawstreams": {strconv.FormatBool(options.RawStreams)},
		"streamurls": {strconv.FormatBool(options.StreamURLs)},
	}
	if options.Conditions != "" {
		q.Set("conditions", options.Conditions)
	}
	if len(options.Fields) > 0 {
		q.Set("query", strings.Join(options.Fields, ","))
	}
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return page, fmt.Errorf("liquipedia: create request: %w", err)
	}
	req.Header.Set("Authorization", "Apikey "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	// Explicit gzip support also works with a supplied transport that disables
	// Go's automatic compression handling.
	req.Header.Set("Accept-Encoding", "gzip")
	res, err := c.http.Do(req)
	if err != nil {
		return page, fmt.Errorf("liquipedia: request matches: %w", err)
	}
	defer res.Body.Close()
	var body io.Reader = res.Body
	if strings.EqualFold(res.Header.Get("Content-Encoding"), "gzip") {
		reader, err := gzip.NewReader(res.Body)
		if err != nil {
			return page, fmt.Errorf("liquipedia: read gzip response: %w", err)
		}
		defer reader.Close()
		body = reader
	}
	data, err := io.ReadAll(body)
	if err != nil {
		return page, fmt.Errorf("liquipedia: read response: %w", err)
	}
	// Read API errors before decoding records: result may be empty on failure.
	var envelope struct {
		Result   jsontext.Value `json:"result"`
		Errors   []string       `json:"error"`
		Warnings []string       `json:"warning"`
	}
	decodeErr := json.Unmarshal(data, &envelope)
	if res.StatusCode != http.StatusOK || len(envelope.Errors) > 0 {
		return page, &APIError{StatusCode: res.StatusCode, Messages: envelope.Errors, Warnings: envelope.Warnings}
	}
	if decodeErr != nil {
		return page, fmt.Errorf("liquipedia: decode response: %w", decodeErr)
	}
	if len(envelope.Result) == 0 || string(envelope.Result) == "null" {
		return page, errors.New("liquipedia: response is missing a result array")
	}
	if err := json.Unmarshal(envelope.Result, &page.Matches); err != nil {
		return MatchesPage{}, fmt.Errorf("liquipedia: decode matches: %w", err)
	}
	page.Warnings = envelope.Warnings
	return page, nil
}
