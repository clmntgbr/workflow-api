package variable

import (
	"encoding/json"
	"testing"

	"github.com/PaesslerAG/jsonpath"
)

func TestNormalizeJSONPath_hydraMember(t *testing.T) {
	got := NormalizeJSONPath(`$.hydra:member[0].id`)
	want := `$["hydra:member"][0].id`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestNormalizeJSONPath_atContext(t *testing.T) {
	got := NormalizeJSONPath(`$.@context`)
	want := `$["@context"]`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestNormalizeJSONPath_alreadyBracketed(t *testing.T) {
	path := `$["hydra:member"][0].id`
	if NormalizeJSONPath(path) != path {
		t.Fatalf("expected path to remain unchanged")
	}
}

func TestNormalizeJSONPath_simpleKey(t *testing.T) {
	got := NormalizeJSONPath(`$.id`)
	want := `$.id`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestExtractByPath_hydraCollection(t *testing.T) {
	raw := `{"hydra:member":[{"id":"01a00e71-6ee5-7c3b-8096-88be9c9e1ea3"}]}`
	var body any
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		t.Fatal(err)
	}

	value, err := ExtractByPath(body, `$.hydra:member[0].id`)
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}
	if value != "01a00e71-6ee5-7c3b-8096-88be9c9e1ea3" {
		t.Fatalf("got %v", value)
	}
}

func TestExtractResponsePaths_hydraCollection(t *testing.T) {
	raw := `{"hydra:member":[{"id":"abc"}],"@context":"/contexts/Order"}`
	var body any
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		t.Fatal(err)
	}

	paths := ExtractResponsePaths(body, "id")
	found := false
	for _, path := range paths {
		normalized := NormalizeJSONPath(path)
		value, err := jsonpath.Get(normalized, body)
		if err == nil && value == "abc" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected extractable id path, got %v", paths)
	}
}

func TestAppendPathSegment(t *testing.T) {
	tests := []struct {
		current string
		key     string
		want    string
	}{
		{"$", "id", "$.id"},
		{"$", "hydra:member", `$["hydra:member"]`},
		{`$["hydra:member"]`, "id", `$["hydra:member"].id`},
		{"$.foo", "@type", `$.foo["@type"]`},
	}
	for _, tt := range tests {
		got := appendPathSegment(tt.current, tt.key)
		if got != tt.want {
			t.Fatalf("appendPathSegment(%q, %q) = %q, want %q", tt.current, tt.key, got, tt.want)
		}
	}
}
