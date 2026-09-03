package presenter

import (
	domainheader "go-api/internal/domain/header"
)

type HeaderSuggestion struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

type HeaderValueSuggestion struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Count int64  `json:"count"`
}

func NewHeaderSuggestionResponse(suggestion domainheader.HeaderSuggestion) HeaderSuggestion {
	return HeaderSuggestion{
		Key:   suggestion.Key,
		Count: suggestion.Count,
	}
}

func NewHeaderSuggestionsResponse(suggestions []domainheader.HeaderSuggestion) []HeaderSuggestion {
	items := make([]HeaderSuggestion, 0, len(suggestions))
	for _, suggestion := range suggestions {
		items = append(items, NewHeaderSuggestionResponse(suggestion))
	}
	return items
}

func NewHeaderValueSuggestionResponse(suggestion domainheader.HeaderValueSuggestion) HeaderValueSuggestion {
	return HeaderValueSuggestion{
		Key:   suggestion.Key,
		Value: suggestion.Value,
		Count: suggestion.Count,
	}
}

func NewHeaderValueSuggestionsResponse(suggestions []domainheader.HeaderValueSuggestion) []HeaderValueSuggestion {
	items := make([]HeaderValueSuggestion, 0, len(suggestions))
	for _, suggestion := range suggestions {
		items = append(items, NewHeaderValueSuggestionResponse(suggestion))
	}
	return items
}
