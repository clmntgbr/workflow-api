package variable

import (
	"encoding/json"
	"testing"
)

func TestNormalizeJSONPath_hydraMember(t *testing.T) {
	got := NormalizeJSONPath(`$.hydra:member[0].id`)
	want := `$["hydra:member"][0].id`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestNormalizeJSONPath_atContext(t *testing.T) {
	got := NormalizeJSONPath(`$.hydra:member[0].@id`)
	want := `$["hydra:member"][0]["@id"]`
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

func TestExtractResponsePaths_keepsDotNotation(t *testing.T) {
	raw := `{"hydra:member":[{"id":"abc"}]}`
	var body any
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		t.Fatal(err)
	}

	paths := ExtractResponsePaths(body, "")
	want := `$.hydra:member[0].id`
	found := false
	for _, path := range paths {
		if path == want {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected %q in paths, got %v", want, paths)
	}

	value, err := ExtractByPath(body, want)
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}
	if value != "abc" {
		t.Fatalf("got %v", value)
	}
}
