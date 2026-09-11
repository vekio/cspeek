package liquipedia

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestMatchesRequestAndGzipResponse(t *testing.T) {
	conditions := "[[game::cs2]] AND ([[date::<2026-09-11 14:30:00]] OR [[date::2026-09-11 14:30:00]])"
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	fmt.Fprint(writer, `{"result":[{"objectname":"same","date":"2026-09-11 08:00:00","finished":0,"dateexact":1},{"objectname":"same","finished":false,"dateexact":true}],"warning":["partial data"]}`)
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	transport := transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != "GET" || r.URL.Path != "/api/v3/match" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		for name, want := range map[string]string{
			"Authorization": "Apikey test-key", "Accept": "application/json", "Accept-Encoding": "gzip",
		} {
			if got := r.Header.Get(name); got != want {
				t.Errorf("%s = %q, want %q", name, got, want)
			}
		}
		for name, want := range map[string]string{
			"wiki": "counterstrike", "conditions": conditions, "limit": "2", "offset": "4",
			"order": "date ASC,objectname ASC", "rawstreams": "false", "streamurls": "false",
			"query": "objectname,date,finished,dateexact",
		} {
			if got := r.URL.Query().Get(name); got != want {
				t.Errorf("%s = %q, want %q", name, got, want)
			}
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Encoding": {"gzip"}},
			Body:       io.NopCloser(bytes.NewReader(compressed.Bytes())),
			Request:    r,
		}, nil
	})
	client, err := NewClient(Config{
		APIKey:     "test-key",
		BaseURL:    "https://example.test/api/v3",
		HTTPClient: &http.Client{Transport: transport},
	})
	if err != nil {
		t.Fatal(err)
	}
	page, err := client.Matches(context.Background(), MatchesOptions{
		Wiki: "counterstrike", Conditions: conditions, Limit: 2, Offset: 4,
		Fields: []string{"objectname", "date", "finished", "dateexact"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Matches) != 2 || page.Matches[0].ObjectName != page.Matches[1].ObjectName {
		t.Fatalf("records must retain API order and duplicates: %+v", page.Matches)
	}
	if len(page.Warnings) != 1 || page.Warnings[0] != "partial data" {
		t.Fatalf("warnings lost: %v", page.Warnings)
	}
	if want := time.Date(2026, 9, 11, 8, 0, 0, 0, time.UTC); !page.Matches[0].Date.Equal(want) {
		t.Errorf("date = %s, want %s", page.Matches[0].Date, want)
	}
	for _, match := range page.Matches {
		if match.Finished || !match.DateExact {
			t.Errorf("incorrect boolean decoding: %+v", match)
		}
	}
}

func TestMatchesErrorsAndEmptyResults(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		apiErr bool
		fail   bool
	}{
		{"empty", 200, `{"result":[]}`, false, false},
		{"unauthorized", 401, `{"error":["invalid key"],"result":[]}`, true, true},
		{"API error with 200", 200, `{"error":["invalid query"],"result":[]}`, true, true},
		{"HTML gateway error", 502, `<html>Bad gateway</html>`, true, true},
		{"malformed JSON", 200, `{`, false, true},
		{"missing result", 200, `{}`, false, true},
		{"null result", 200, `{"result":null}`, false, true},
		{"invalid record", 200, `{"result":[{"finished":2}]}`, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := transportFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Query().Has("conditions") || r.URL.Query().Has("query") || r.URL.Query().Get("limit") != "20" {
					t.Error("unexpected implicit filtering or defaults")
				}
				return &http.Response{
					StatusCode: tt.status,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(tt.body)),
					Request:    r,
				}, nil
			})
			client, err := NewClient(Config{
				APIKey:     "test-key",
				BaseURL:    "https://example.test/api/v3",
				HTTPClient: &http.Client{Transport: transport},
			})
			if err != nil {
				t.Fatal(err)
			}
			page, err := client.Matches(context.Background(), MatchesOptions{Wiki: "counterstrike"})
			if (err != nil) != tt.fail {
				t.Fatalf("error = %v, want failure %v", err, tt.fail)
			}
			var apiErr *APIError
			if errors.As(err, &apiErr) != tt.apiErr {
				t.Fatalf("unexpected error type: %T", err)
			}
			if apiErr != nil && apiErr.StatusCode != tt.status {
				t.Errorf("HTTP status lost: %d", apiErr.StatusCode)
			}
			if !tt.fail && page.Matches == nil {
				t.Error("successful empty result should be an empty slice")
			}
		})
	}
}

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestContextCancellation(t *testing.T) {
	client, err := NewClient(Config{
		APIKey: "test-key",
		HTTPClient: &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
			return nil, r.Context().Err()
		})},
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = client.Matches(ctx, MatchesOptions{Wiki: "counterstrike"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation not preserved: %v", err)
	}
}

func TestValidationBeforeRequest(t *testing.T) {
	client, err := NewClient(Config{APIKey: "test-key", HTTPClient: &http.Client{
		Transport: transportFunc(func(*http.Request) (*http.Response, error) {
			t.Error("invalid options should not send a request")
			return nil, errors.New("unexpected request")
		}),
	}})
	if err != nil {
		t.Fatal(err)
	}
	for _, opts := range []MatchesOptions{
		{}, {Wiki: "counterstrike", Limit: -1}, {Wiki: "counterstrike", Limit: 1001}, {Wiki: "counterstrike", Offset: -1},
	} {
		if _, err := client.Matches(context.Background(), opts); err == nil {
			t.Errorf("accepted invalid options: %+v", opts)
		}
	}
	for _, config := range []Config{
		{}, {APIKey: "bad\nkey"}, {APIKey: "key", BaseURL: "://"}, {APIKey: "key", BaseURL: "ftp://host"},
	} {
		if _, err := NewClient(config); err == nil {
			t.Error("accepted invalid config")
		}
	}
}

func TestAPIErrorMessage(t *testing.T) {
	err := &APIError{StatusCode: 200, Messages: []string{"invalid condition", "unknown field"}}
	if !strings.Contains(err.Error(), "invalid condition; unknown field") {
		t.Fatal(err.Error())
	}
}
