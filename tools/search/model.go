package search

import (
	"context"
	"net/http"
)

// Common constants
var (
	searchHTMLURL = "https://html.duckduckgo.com/html/"

	defaultTextSearchToolName = "duckduckgo_text_search"
	defaultTextSearchToolDesc = `DuckDuckGo plain-text web search.

Use when the answer may depend on current or external information, or when you need candidate URLs before fetching source pages.

Guidelines:
- For recent information, use time_range when appropriate and include concrete dates in the query.
- Prefer official, primary, authoritative, or original sources when available.
- Search results are candidates, not final evidence; use fetch for pages that matter or when snippets conflict.
- Avoid repeated identical searches; vary language, entity names, or source type when broadening coverage.
- When answering from web results, preserve the URLs you actually used so the final answer can cite sources.`
)

type Search interface {
	TextSearch(ctx context.Context, req *TextSearchRequest) (*TextSearchResponse, error)
}

// client represents the DuckDuckGo search client.
// It handles all search-related operations including request configuration,
// caching, and result parsing.
type client struct {
	httpCli    *http.Client
	maxResults int
	region     Region
}

// Region represents a geographical region for search results.
// Different regions may return different search results based on local relevance.
// others can be found at: https://duckduckgo.com/duckduckgo-help-pages/settings/params/
type Region string

// Available regions for DuckDuckGo search
const (
	// RegionWT represents World region (No specific region, default)
	RegionWT Region = "wt-wt"
	// RegionUS represents United States region
	RegionUS Region = "us-en"
	// RegionUK represents United Kingdom region
	RegionUK Region = "uk-en"
	// RegionDE represents Germany region
	RegionDE Region = "de-de"
	// RegionFR represents France region
	RegionFR Region = "fr-fr"
	// RegionJP represents Japan region
	RegionJP Region = "jp-jp"
	// RegionCN represents China region
	RegionCN Region = "cn-zh"
	// RegionRU represents Russia region
	RegionRU Region = "ru-ru"
)

// TimeRange represents the time range for search results.
type TimeRange string

const (
	// TimeRangeDay limits results to the past day
	TimeRangeDay TimeRange = "d"
	// TimeRangeWeek limits results to the past week
	TimeRangeWeek TimeRange = "w"
	// TimeRangeMonth limits results to the past month
	TimeRangeMonth TimeRange = "m"
	// TimeRangeYear limits results to the past year
	TimeRangeYear TimeRange = "y"
	// TimeRangeAny results at any time
	TimeRangeAny TimeRange = ""
)

type TextSearchRequest struct {
	// Query is the user's search query
	Query string `json:"query"`
	// TimeRange is the search time range
	// Default: TimeRangeAny
	TimeRange TimeRange `json:"time_range"`
}

// TextSearchResult represents a single search result.
// Contains the title, URL, and summary of the result.
type TextSearchResult struct {
	// Title is the title of the search result
	Title string `json:"title"`
	// URL is the web address of the result
	URL string `json:"url"`
	// Summary is the summary of the result content
	Summary string `json:"summary"`
}

// TextSearchResponse represents the complete response from a search request.
type TextSearchResponse struct {
	// Message is a brief status message for the model
	Message string `json:"message,omitempty"`
	// Results contains the list of search results
	Results []*TextSearchResult `json:"results,omitempty"`
	// ErrorMessage contains error information to guide the model
	ErrorMessage string `json:"error_message,omitempty"`
}
